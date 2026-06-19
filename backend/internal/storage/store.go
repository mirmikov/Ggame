package storage

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"ggame/backend/internal/models"

	"github.com/lib/pq"
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	DisplayName  string    `json:"displayName"`
	Grade        int       `json:"grade"`
	PasswordHash string    `json:"passwordHash"`
	CreatedAt    time.Time `json:"createdAt"`
}

type PublicUser struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	Grade       int    `json:"grade"`
}

type Store struct {
	db *sql.DB
}

func New(databaseURL string) (*Store, error) {
	if strings.TrimSpace(databaseURL) == "" {
		return nil, errors.New("DATABASE_URL is required")
	}
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(12)
	db.SetMaxIdleConns(4)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	store := &Store{db: db}
	if err := store.migrate(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) PublicUser(user *User) PublicUser {
	if user == nil {
		return PublicUser{}
	}
	return PublicUser{ID: user.ID, Email: user.Email, DisplayName: user.DisplayName, Grade: user.Grade}
}

func (s *Store) CreateUser(email, displayName, password string, grade int) (*User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	displayName = strings.TrimSpace(displayName)
	if email == "" || !strings.Contains(email, "@") {
		return nil, errors.New("укажите корректный email")
	}
	if len([]rune(displayName)) < 2 {
		return nil, errors.New("имя должно быть не короче 2 символов")
	}
	if len(password) < 6 {
		return nil, errors.New("пароль должен быть не короче 6 символов")
	}
	if grade < 9 || grade > 11 {
		grade = 9
	}
	hash, err := hashPassword(password)
	if err != nil {
		return nil, err
	}
	user := &User{
		ID:           randomToken(18),
		Email:        email,
		DisplayName:  displayName,
		Grade:        grade,
		PasswordHash: hash,
		CreatedAt:    time.Now(),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO users (id, email, display_name, grade, password_hash, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, user.ID, user.Email, user.DisplayName, user.Grade, user.PasswordHash, user.CreatedAt)
	if isUniqueViolation(err) {
		return nil, errors.New("пользователь с таким email уже зарегистрирован")
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Store) Authenticate(email, password string) (*User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	user, err := s.userByEmail(ctx, email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("неверный email или пароль")
	}
	if err != nil {
		return nil, err
	}
	if user == nil || !verifyPassword(user.PasswordHash, password) {
		return nil, errors.New("неверный email или пароль")
	}
	return user, nil
}

func (s *Store) CreateSession(userID string) (string, error) {
	token := randomToken(32)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO sessions (token, user_id, created_at, expires_at)
		VALUES ($1, $2, now(), now() + interval '30 days')
	`, token, userID)
	return token, err
}

func (s *Store) DeleteSession(token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE token = $1`, token)
	return err
}

func (s *Store) UserBySession(token string) (*User, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	row := s.db.QueryRowContext(ctx, `
		SELECT u.id, u.email, u.display_name, u.grade, u.password_hash, u.created_at
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token = $1 AND s.expires_at > now()
	`, token)
	user, err := scanUser(row)
	return user, err == nil
}

func (s *Store) LoadRooms() ([]*models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	rows, err := s.db.QueryContext(ctx, `SELECT state FROM rooms ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rooms := make([]*models.Room, 0)
	for rows.Next() {
		var data []byte
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}
		var room models.Room
		if err := json.Unmarshal(data, &room); err != nil {
			return nil, err
		}
		rooms = append(rooms, &room)
	}
	sort.SliceStable(rooms, func(i, j int) bool {
		return rooms[i].CreatedAt.Before(rooms[j].CreatedAt)
	})
	return rooms, rows.Err()
}

func (s *Store) SaveRoom(room *models.Room) error {
	if room == nil || room.UniqueServerID == "" {
		return nil
	}
	data, err := json.Marshal(room)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO rooms (id, server_name, game_mode, status, organizer_id, state, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, now())
		ON CONFLICT (id) DO UPDATE SET
			server_name = EXCLUDED.server_name,
			game_mode = EXCLUDED.game_mode,
			status = EXCLUDED.status,
			organizer_id = EXCLUDED.organizer_id,
			state = EXCLUDED.state,
			updated_at = now()
	`, strings.ToUpper(room.UniqueServerID), room.ServerName, room.GameMode, room.Status, room.OrganizerID, string(data), room.CreatedAt)
	return err
}

func (s *Store) AppendEvent(roomID, eventType string, payload map[string]any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO tournament_events (id, room_id, type, payload, created_at)
		VALUES ($1, $2, $3, $4::jsonb, now())
	`, randomToken(14), strings.ToUpper(roomID), eventType, string(data))
	return err
}

func (s *Store) migrate(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id text PRIMARY KEY,
			email text NOT NULL UNIQUE,
			display_name text NOT NULL,
			grade integer NOT NULL,
			password_hash text NOT NULL,
			created_at timestamptz NOT NULL DEFAULT now()
		);

		CREATE TABLE IF NOT EXISTS sessions (
			token text PRIMARY KEY,
			user_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			created_at timestamptz NOT NULL DEFAULT now(),
			expires_at timestamptz NOT NULL
		);

		CREATE TABLE IF NOT EXISTS rooms (
			id text PRIMARY KEY,
			server_name text NOT NULL,
			game_mode text NOT NULL,
			status text NOT NULL,
			organizer_id text NOT NULL,
			state jsonb NOT NULL,
			created_at timestamptz NOT NULL DEFAULT now(),
			updated_at timestamptz NOT NULL DEFAULT now()
		);

		CREATE TABLE IF NOT EXISTS tournament_events (
			id text PRIMARY KEY,
			room_id text NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
			type text NOT NULL,
			payload jsonb NOT NULL DEFAULT '{}'::jsonb,
			created_at timestamptz NOT NULL DEFAULT now()
		);

		CREATE INDEX IF NOT EXISTS sessions_user_id_idx ON sessions(user_id);
		CREATE INDEX IF NOT EXISTS rooms_status_idx ON rooms(status);
		CREATE INDEX IF NOT EXISTS tournament_events_room_id_created_at_idx ON tournament_events(room_id, created_at);
	`)
	return err
}

func (s *Store) userByEmail(ctx context.Context, email string) (*User, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, email, display_name, grade, password_hash, created_at
		FROM users
		WHERE email = $1
	`, email)
	return scanUser(row)
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanUser(row rowScanner) (*User, error) {
	user := &User{}
	err := row.Scan(&user.ID, &user.Email, &user.DisplayName, &user.Grade, &user.PasswordHash, &user.CreatedAt)
	return user, err
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}

func hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	sum := passwordDigest(salt, password)
	return fmt.Sprintf("%s:%s", base64.RawStdEncoding.EncodeToString(salt), hex.EncodeToString(sum)), nil
}

func verifyPassword(encoded, password string) bool {
	parts := strings.Split(encoded, ":")
	if len(parts) != 2 {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	expected, err := hex.DecodeString(parts[1])
	if err != nil {
		return false
	}
	actual := passwordDigest(salt, password)
	return subtle.ConstantTimeCompare(actual, expected) == 1
}

func passwordDigest(salt []byte, password string) []byte {
	data := append(append([]byte(nil), salt...), []byte(password)...)
	sum := sha256.Sum256(data)
	for i := 0; i < 120_000; i++ {
		next := sha256.Sum256(append(sum[:], data...))
		sum = next
	}
	return sum[:]
}

func randomToken(size int) string {
	data := make([]byte, size)
	if _, err := rand.Read(data); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return base64.RawURLEncoding.EncodeToString(data)
}
