package ws

import (
	"net/http"
	"strings"

	"ggame/backend/internal/models"
	"ggame/backend/internal/rooms"

	"github.com/gorilla/websocket"
)

type Handler struct {
	Rooms    *rooms.Manager
	upgrader websocket.Upgrader
}

func New(manager *rooms.Manager) *Handler {
	return &Handler{Rooms: manager, upgrader: websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	roomID := strings.ToUpper(strings.TrimPrefix(r.URL.Path, "/ws/rooms/"))
	if _, ok := h.Rooms.Get(roomID); !ok {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	h.Rooms.AddClient(roomID, conn)
	if room, ok := h.Rooms.Get(roomID); ok {
		_ = h.Rooms.SendClient(conn, models.Event{Type: "room_state", Payload: room})
	}
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			h.Rooms.RemoveClient(roomID, conn)
			return
		}
	}
}
