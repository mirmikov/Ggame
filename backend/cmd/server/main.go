package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"ggame/backend/internal/models"
	"ggame/backend/internal/rooms"
	gamews "ggame/backend/internal/ws"
)

type api struct{ rooms *rooms.Manager }

func main() {
	manager := rooms.NewManager()
	a := &api{rooms: manager}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, map[string]string{"status": "ok"}) })
	mux.HandleFunc("/api/questions", a.questions)
	mux.HandleFunc("/api/tasks", a.tasks)
	mux.HandleFunc("/api/rooms", a.createRoom)
	mux.HandleFunc("/api/rooms/", a.roomAction)
	mux.Handle("/ws/rooms/", gamews.New(manager))
	mux.Handle("/", spaHandler(staticDir()))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Prometheus Battle listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, cors(mux)))
}

func (a *api) createRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	var in rooms.CreateInput
	if !decode(w, r, &in) {
		return
	}
	room, player, err := a.rooms.Create(in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, 201, map[string]any{"room": room, "player": player})
}

func (a *api) roomAction(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/rooms/"), "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	id := strings.ToUpper(parts[0])
	if len(parts) == 1 && r.Method == http.MethodGet {
		room, ok := a.rooms.Get(id)
		if !ok {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, 200, room)
		return
	}
	if len(parts) != 2 || r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	switch parts[1] {
	case "join":
		var in struct {
			Nickname string `json:"nickname"`
			Grade    int    `json:"grade"`
		}
		if !decode(w, r, &in) {
			return
		}
		room, player, err := a.rooms.Join(id, in.Nickname, in.Grade)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, 200, map[string]any{"room": room, "player": player})
	case "team":
		var in struct {
			PlayerID string          `json:"playerId"`
			Team     models.TeamName `json:"team"`
		}
		if !decode(w, r, &in) {
			return
		}
		room, err := a.rooms.SelectTeam(id, in.PlayerID, in.Team)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, 200, room)
	case "qualifier-team":
		var in struct {
			PlayerID string `json:"playerId"`
			TeamID   string `json:"teamId"`
		}
		if !decode(w, r, &in) {
			return
		}
		room, err := a.rooms.SelectQualifierTeam(id, in.PlayerID, in.TeamID)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, 200, room)
	case "bot":
		var in struct {
			PlayerID string `json:"playerId"`
		}
		if !decode(w, r, &in) {
			return
		}
		room, err := a.rooms.AddBot(id, in.PlayerID)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, 200, room)
	case "start":
		var in struct {
			PlayerID string `json:"playerId"`
		}
		if !decode(w, r, &in) {
			return
		}
		room, err := a.rooms.Start(id, in.PlayerID)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, 200, room)
	case "answer":
		var in struct {
			PlayerID   string `json:"playerId"`
			QuestionID string `json:"questionId"`
			Answer     int    `json:"answer"`
		}
		if !decode(w, r, &in) {
			return
		}
		correct, explanation, room, err := a.rooms.Answer(id, in.PlayerID, in.QuestionID, in.Answer)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, 200, map[string]any{"correct": correct, "explanation": explanation, "room": room})
	case "task":
		var in struct {
			PlayerID string `json:"playerId"`
			TaskID   string `json:"taskId"`
			Answer   string `json:"answer"`
		}
		if !decode(w, r, &in) {
			return
		}
		correct, room, err := a.rooms.SubmitTask(id, in.PlayerID, in.TaskID, in.Answer)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, 200, map[string]any{"correct": correct, "room": room})
	default:
		http.NotFound(w, r)
	}
}

func (a *api) questions(w http.ResponseWriter, r *http.Request) {
	grade, _ := strconv.Atoi(r.URL.Query().Get("grade"))
	writeJSON(w, 200, a.rooms.Questions(grade))
}
func (a *api) tasks(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, a.rooms.Tasks()) }

func decode(w http.ResponseWriter, r *http.Request, target any) bool {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid JSON"})
		return false
	}
	return true
}
func writeError(w http.ResponseWriter, err error) {
	writeJSON(w, 400, map[string]string{"error": err.Error()})
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func staticDir() string {
	if dir := os.Getenv("STATIC_DIR"); dir != "" {
		return dir
	}
	return "./public"
}

func spaHandler(dir string) http.Handler {
	files := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.NotFound(w, r)
			return
		}
		path := filepath.Join(dir, filepath.Clean(r.URL.Path))
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			files.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(dir, "index.html"))
	})
}
