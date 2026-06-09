package rooms

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"ggame/backend/internal/models"
	"ggame/backend/internal/questions"

	"github.com/gorilla/websocket"
)

type Manager struct {
	mu      sync.RWMutex
	writeMu sync.Mutex
	rooms   map[string]*models.Room
	clients map[string]map[*websocket.Conn]bool
}

type CreateInput struct {
	ServerName string          `json:"serverName"`
	MaxPlayers int             `json:"maxPlayers"`
	GradeMode  string          `json:"gradeMode"`
	GameMode   string          `json:"gameMode"`
	Settings   models.Settings `json:"settings"`
	Nickname   string          `json:"nickname"`
	Grade      int             `json:"grade"`
}

func NewManager() *Manager {
	return &Manager{rooms: map[string]*models.Room{}, clients: map[string]map[*websocket.Conn]bool{}}
}

func basePlayer(nickname string, grade int, host bool) *models.Player {
	return &models.Player{ID: randomID("P", 6), Nickname: strings.TrimSpace(nickname), Grade: grade, Level: 1, HP: 30, MaxHP: 30, Attack: 6, Defense: 2, Speed: 1, IsHost: host}
}

func (m *Manager) Create(in CreateInput) (*models.Room, *models.Player, error) {
	if in.MaxPlayers < 2 || in.MaxPlayers > 12 {
		return nil, nil, errors.New("maxPlayers должен быть от 2 до 12")
	}
	if strings.TrimSpace(in.Nickname) == "" {
		return nil, nil, errors.New("укажите nickname")
	}
	if in.Settings.TowerHP <= 0 {
		in.Settings.TowerHP = 160
	}
	if in.Settings.RoundDurationSeconds <= 0 {
		in.Settings.RoundDurationSeconds = 600
	}
	if in.Settings.TeamPlayerLimit <= 0 {
		in.Settings.TeamPlayerLimit = in.MaxPlayers / 2
	}
	roomID := randomID("CYB-", 5)
	player := basePlayer(in.Nickname, in.Grade, true)
	room := &models.Room{
		UniqueServerID: roomID, ServerName: in.ServerName, MaxPlayers: in.MaxPlayers, GradeMode: in.GradeMode, GameMode: in.GameMode,
		Status: "waiting", Players: map[string]*models.Player{player.ID: player}, Units: map[string]*models.BattleUnit{}, Projectiles: []models.Projectile{}, CreatedAt: time.Now(), Settings: in.Settings,
		Teams: map[models.TeamName]*models.Team{
			models.NexGen:   {Name: models.NexGen, TowerHP: in.Settings.TowerHP, MaxTowerHP: in.Settings.TowerHP},
			models.OmniSoft: {Name: models.OmniSoft, TowerHP: in.Settings.TowerHP, MaxTowerHP: in.Settings.TowerHP},
		},
	}
	m.mu.Lock()
	m.rooms[roomID] = room
	m.clients[roomID] = map[*websocket.Conn]bool{}
	m.mu.Unlock()
	return room, player, nil
}

func (m *Manager) Join(roomID, nickname string, grade int) (*models.Room, *models.Player, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	room, ok := m.rooms[strings.ToUpper(roomID)]
	if !ok {
		return nil, nil, errors.New("сервер не найден")
	}
	if room.Status != "waiting" {
		return nil, nil, errors.New("игра уже запущена")
	}
	if len(room.Players) >= room.MaxPlayers {
		return nil, nil, errors.New("сервер заполнен")
	}
	player := basePlayer(nickname, grade, false)
	room.Players[player.ID] = player
	go m.Broadcast(roomID, "room_state", room)
	return room, player, nil
}

func (m *Manager) AddBot(roomID, playerID string) (*models.Room, error) {
	m.mu.Lock()
	room, ok := m.rooms[strings.ToUpper(roomID)]
	if !ok {
		m.mu.Unlock()
		return nil, errors.New("сервер не найден")
	}
	host := room.Players[playerID]
	if host == nil || !host.IsHost {
		m.mu.Unlock()
		return nil, errors.New("только host может добавить бота")
	}
	if room.Status != "waiting" {
		m.mu.Unlock()
		return nil, errors.New("игра уже запущена")
	}
	if len(room.Players) >= room.MaxPlayers {
		m.mu.Unlock()
		return nil, errors.New("сервер заполнен")
	}
	for _, player := range room.Players {
		if player.IsBot {
			m.mu.Unlock()
			return nil, errors.New("тестовый бот уже добавлен")
		}
	}
	team := models.OmniSoft
	if host.Team == models.OmniSoft {
		team = models.NexGen
	} else if host.Team == "" && teamCount(room, models.NexGen) < teamCount(room, models.OmniSoft) {
		team = models.NexGen
	}
	if teamCount(room, team) >= room.Settings.TeamPlayerLimit {
		m.mu.Unlock()
		return nil, errors.New("для бота нет свободного места в команде")
	}
	bot := basePlayer("BOT-7", host.Grade, false)
	bot.ID = randomID("BOT-", 5)
	bot.IsBot = true
	bot.Team = team
	room.Players[bot.ID] = bot
	m.mu.Unlock()
	m.Broadcast(roomID, "room_state", room)
	return room, nil
}

func (m *Manager) Get(roomID string) (*models.Room, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	room, ok := m.rooms[strings.ToUpper(roomID)]
	return room, ok
}

func (m *Manager) SelectTeam(roomID, playerID string, team models.TeamName) (*models.Room, error) {
	if team != models.NexGen && team != models.OmniSoft {
		return nil, errors.New("неизвестная команда")
	}
	m.mu.Lock()
	room, ok := m.rooms[strings.ToUpper(roomID)]
	if !ok {
		m.mu.Unlock()
		return nil, errors.New("сервер не найден")
	}
	player, ok := room.Players[playerID]
	if !ok {
		m.mu.Unlock()
		return nil, errors.New("игрок не найден")
	}
	count := 0
	for _, p := range room.Players {
		if p.Team == team && p.ID != playerID {
			count++
		}
	}
	if count >= room.Settings.TeamPlayerLimit {
		m.mu.Unlock()
		return nil, errors.New("команда заполнена")
	}
	player.Team = team
	m.mu.Unlock()
	m.Broadcast(roomID, "room_state", room)
	return room, nil
}

func (m *Manager) Start(roomID, playerID string) (*models.Room, error) {
	m.mu.Lock()
	room, ok := m.rooms[strings.ToUpper(roomID)]
	if !ok {
		m.mu.Unlock()
		return nil, errors.New("сервер не найден")
	}
	player := room.Players[playerID]
	if player == nil || !player.IsHost {
		m.mu.Unlock()
		return nil, errors.New("только host может запустить игру")
	}
	if room.Status != "waiting" {
		m.mu.Unlock()
		return nil, errors.New("игра уже запущена")
	}
	hasNex, hasOmni := false, false
	players := make([]*models.Player, 0, len(room.Players))
	for _, p := range room.Players {
		players = append(players, p)
	}
	sort.Slice(players, func(i, j int) bool { return players[i].ID < players[j].ID })
	teamLanes := map[models.TeamName]int{models.NexGen: 0, models.OmniSoft: 0}
	for _, p := range players {
		hasNex = hasNex || p.Team == models.NexGen
		hasOmni = hasOmni || p.Team == models.OmniSoft
		if p.Team == "" {
			m.mu.Unlock()
			return nil, errors.New("все игроки должны выбрать команду")
		}
		p.QuestionID = m.nextQuestionID(p.Grade, "")
		start := 8.0
		if p.Team == models.OmniSoft {
			start = 92
		}
		lane := teamLanes[p.Team]
		teamLanes[p.Team]++
		room.Units[p.ID] = &models.BattleUnit{OwnerPlayerID: p.ID, Nickname: p.Nickname, Team: p.Team, HP: p.HP, MaxHP: p.MaxHP, Attack: p.Attack, Defense: p.Defense, Speed: p.Speed, Level: p.Level, Lane: lane, Position: start}
	}
	if !hasNex || !hasOmni {
		m.mu.Unlock()
		return nil, errors.New("нужен хотя бы один игрок в каждой команде")
	}
	room.Status = "running"
	room.EndsAt = time.Now().Add(time.Duration(room.Settings.RoundDurationSeconds) * time.Second).Unix()
	if room.GameMode == "final_pvp" {
		room.StoryMessage = "Башни активированы. Кодеры запускают боевые алгоритмы. Победит команда, чей код окажется сильнее."
	} else {
		room.StoryMessage = "Вы подключились к периметру вражеской сети. Соберите ключи-модули, чтобы обойти защиту."
	}
	m.mu.Unlock()
	m.Broadcast(roomID, "game_started", room)
	go m.battleLoop(roomID)
	for _, p := range room.Players {
		if p.IsBot {
			go m.botLoop(roomID, p.ID)
		}
	}
	return room, nil
}

func (m *Manager) Answer(roomID, playerID, questionID string, answer int) (bool, string, *models.Room, error) {
	m.mu.Lock()
	room, ok := m.rooms[strings.ToUpper(roomID)]
	if !ok || room.Status != "running" {
		m.mu.Unlock()
		return false, "", nil, errors.New("игра не запущена")
	}
	p := room.Players[playerID]
	if p == nil {
		m.mu.Unlock()
		return false, "", nil, errors.New("игрок не найден")
	}
	if time.Now().Unix() < p.LockedUntil {
		m.mu.Unlock()
		return false, "", nil, errors.New("Theory временно заблокирована")
	}
	if p.QuestionID != questionID {
		m.mu.Unlock()
		return false, "", nil, errors.New("вопрос уже изменился")
	}
	q, found := questionByID(questionID)
	if !found {
		m.mu.Unlock()
		return false, "", nil, errors.New("вопрос не найден")
	}
	correct := answer == q.CorrectAnswer
	if correct {
		p.CorrectAnswers++
		p.WrongStreak = 0
		p.Score += 100 * q.Difficulty
		p.XP += 50 * q.Difficulty
		room.Teams[p.Team].Score += 100 * q.Difficulty
		if p.XP >= p.Level*100 {
			p.XP -= p.Level * 100
			p.Level++
			p.Attack += 2
			p.Defense++
			p.MaxHP += 5
			p.HP = p.MaxHP
			p.Speed += .1
			if u := room.Units[p.ID]; u != nil {
				u.Level, u.Attack, u.Defense, u.MaxHP, u.HP, u.Speed = p.Level, p.Attack, p.Defense, p.MaxHP, p.MaxHP, p.Speed
			}
		}
	} else {
		p.WrongAnswers++
		p.WrongStreak++
		if p.WrongStreak >= 3 {
			p.LockedUntil = time.Now().Add(60 * time.Second).Unix()
			p.WrongStreak = 0
		}
	}
	p.QuestionID = m.nextQuestionID(p.Grade, questionID)
	explanation := q.Explanation
	m.mu.Unlock()
	m.Broadcast(roomID, "player_answered", map[string]any{"playerId": playerID, "correct": correct, "explanation": explanation, "room": room})
	return correct, explanation, room, nil
}

func (m *Manager) Questions(grade int) []models.Question { return questions.ForGrade(grade) }
func (m *Manager) Tasks() []models.TerminalTask          { return questions.Tasks() }

func (m *Manager) SubmitTask(roomID, playerID, taskID, answer string) (bool, *models.Room, error) {
	m.mu.Lock()
	room, ok := m.rooms[strings.ToUpper(roomID)]
	if !ok || room.Status != "running" {
		m.mu.Unlock()
		return false, nil, errors.New("игра не запущена")
	}
	p := room.Players[playerID]
	if p == nil {
		m.mu.Unlock()
		return false, nil, errors.New("игрок не найден")
	}
	for _, solved := range p.SolvedTasks {
		if solved == taskID {
			m.mu.Unlock()
			return false, nil, errors.New("задача уже решена")
		}
	}
	var task *models.TerminalTask
	for _, candidate := range questions.Tasks() {
		if candidate.ID == taskID {
			copy := candidate
			task = &copy
			break
		}
	}
	if task == nil {
		m.mu.Unlock()
		return false, nil, errors.New("задача не найдена")
	}
	correct := strings.EqualFold(strings.Join(strings.Fields(answer), " "), task.ExpectedAnswer)
	if correct {
		p.SolvedTasks = append(p.SolvedTasks, taskID)
		p.Score += task.Reward
		room.Teams[p.Team].Score += task.Reward
	}
	m.mu.Unlock()
	m.Broadcast(roomID, "player_answered", map[string]any{"playerId": playerID, "correct": correct, "room": room})
	return correct, room, nil
}

func (m *Manager) nextQuestionID(grade int, previous string) string {
	q := questions.ForGrade(grade)
	if len(q) == 0 {
		q = questions.All()
	}
	for i := 0; i < 4; i++ {
		next := q[rand.Intn(len(q))].ID
		if next != previous {
			return next
		}
	}
	return q[0].ID
}

func (m *Manager) battleLoop(roomID string) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for range ticker.C {
		m.mu.Lock()
		room, ok := m.rooms[roomID]
		if !ok || room.Status != "running" {
			m.mu.Unlock()
			return
		}
		m.tick(room)
		finished := room.Status == "finished"
		m.mu.Unlock()
		event := "battle_tick"
		if finished {
			event = "game_finished"
		}
		m.Broadcast(roomID, event, room)
		if finished {
			return
		}
	}
}

func (m *Manager) botLoop(roomID, playerID string) {
	timer := time.NewTimer(3 * time.Second)
	defer timer.Stop()
	for range timer.C {
		m.mu.RLock()
		room, ok := m.rooms[roomID]
		if !ok || room.Status != "running" {
			m.mu.RUnlock()
			return
		}
		player := room.Players[playerID]
		if player == nil || player.QuestionID == "" {
			m.mu.RUnlock()
			return
		}
		questionID := player.QuestionID
		m.mu.RUnlock()

		q, found := questionByID(questionID)
		if found {
			answer := q.CorrectAnswer
			if rand.Intn(100) < 20 {
				answer = (q.CorrectAnswer + 1) % len(q.Options)
			}
			_, _, _, _ = m.Answer(roomID, playerID, questionID, answer)
		}
		timer.Reset(time.Duration(4+rand.Intn(4)) * time.Second)
	}
}

func (m *Manager) tick(room *models.Room) {
	now := time.Now().Unix()
	room.Projectiles = room.Projectiles[:0]
	for _, u := range room.Units {
		if u.HP <= 0 {
			if u.RespawnAt <= now {
				p := room.Players[u.OwnerPlayerID]
				u.HP, u.Position, u.RespawnAt = p.MaxHP, 8, 0
				if u.Team == models.OmniSoft {
					u.Position = 92
				}
			}
			continue
		}
		var closest *models.BattleUnit
		distance := 999.0
		for _, enemy := range room.Units {
			d := math.Abs(u.Position - enemy.Position)
			if enemy.Team != u.Team && enemy.Lane == u.Lane && enemy.HP > 0 && d < distance {
				closest, distance = enemy, d
			}
		}
		direction := 1.0
		enemyTower := room.Teams[models.OmniSoft]
		towerPosition := 96.0
		if u.Team == models.OmniSoft {
			direction, enemyTower, towerPosition = -1, room.Teams[models.NexGen], 4
		}
		if closest != nil && distance <= 8 {
			u.Target = closest.OwnerPlayerID
			damage := max(1, u.Attack-closest.Defense/2)
			room.Projectiles = append(room.Projectiles, models.Projectile{ID: randomID("S-", 5), Team: u.Team, From: u.Position, To: closest.Position, FromLane: u.Lane, ToLane: closest.Lane, Damage: damage, Target: closest.OwnerPlayerID})
			closest.HP -= damage
			if closest.HP <= 0 {
				closest.HP, closest.RespawnAt = 0, now+5
			}
		} else if math.Abs(u.Position-towerPosition) <= 7 {
			u.Target = string(enemyTower.Name) + "-tower"
			damage := max(1, u.Attack)
			room.Projectiles = append(room.Projectiles, models.Projectile{ID: randomID("S-", 5), Team: u.Team, From: u.Position, To: towerPosition, FromLane: u.Lane, ToLane: -1, Damage: damage, Target: string(enemyTower.Name) + "-tower", HitTower: true})
			enemyTower.TowerHP -= damage
		} else {
			u.Target = ""
			u.Position += direction * u.Speed * 2.2
		}
	}
	if room.Teams[models.NexGen].TowerHP <= 0 || room.Teams[models.OmniSoft].TowerHP <= 0 || now >= room.EndsAt {
		room.Status = "finished"
		nex, omni := room.Teams[models.NexGen].TowerHP, room.Teams[models.OmniSoft].TowerHP
		if nex >= omni {
			room.Winner = models.NexGen
			room.StoryMessage = "NexGen получила доступ к ядру Прометея."
		} else {
			room.Winner = models.OmniSoft
			room.StoryMessage = "OmniSoft перехватила контроль над глобальной нейросетью."
		}
	}
}

func teamCount(room *models.Room, team models.TeamName) int {
	count := 0
	for _, player := range room.Players {
		if player.Team == team {
			count++
		}
	}
	return count
}

func (m *Manager) AddClient(roomID string, conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.clients[roomID] == nil {
		m.clients[roomID] = map[*websocket.Conn]bool{}
	}
	m.clients[roomID][conn] = true
}

func (m *Manager) RemoveClient(roomID string, conn *websocket.Conn) {
	m.writeMu.Lock()
	defer m.writeMu.Unlock()
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.clients[roomID], conn)
	conn.Close()
}

func (m *Manager) SendClient(conn *websocket.Conn, event models.Event) error {
	m.writeMu.Lock()
	defer m.writeMu.Unlock()
	return conn.WriteJSON(event)
}

func (m *Manager) Broadcast(roomID, eventType string, payload any) {
	m.writeMu.Lock()
	defer m.writeMu.Unlock()
	m.mu.RLock()
	clients := make([]*websocket.Conn, 0, len(m.clients[roomID]))
	for c := range m.clients[roomID] {
		clients = append(clients, c)
	}
	m.mu.RUnlock()
	event := models.Event{Type: eventType, Payload: payload}
	for _, c := range clients {
		_ = c.WriteJSON(event)
	}
}

func questionByID(id string) (models.Question, bool) {
	for _, q := range questions.All() {
		if q.ID == id {
			return q, true
		}
	}
	return models.Question{}, false
}

const letters = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func randomID(prefix string, length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return fmt.Sprintf("%s%s", prefix, b)
}
