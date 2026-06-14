package ws

import (
	"net/http/httptest"
	"strings"
	"testing"

	"ggame/backend/internal/models"
	"ggame/backend/internal/rooms"

	"github.com/gorilla/websocket"
)

func TestInitialRoomState(t *testing.T) {
	manager := rooms.NewManager()
	room, _, err := manager.Create(rooms.CreateInput{
		ServerName: "WS test", MaxPlayers: 4, GradeMode: "mixed", GameMode: models.ModeQualifier,
		Nickname: "Organizer", Grade: 11,
		Settings: models.Settings{RoundDurationSeconds: 60, TowerHP: 200, TeamPlayerLimit: 2, ZoneStepsToCenter: 8, ZonePushbackSteps: 2, ZoneHoldSeconds: 15},
	})
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(New(manager))
	defer server.Close()
	conn, _, err := websocket.DefaultDialer.Dial(strings.Replace(server.URL, "http", "ws", 1)+"/ws/rooms/"+room.UniqueServerID, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	var event models.Event
	if err := conn.ReadJSON(&event); err != nil {
		t.Fatal(err)
	}
	if event.Type != "room_state" {
		t.Fatalf("unexpected event: %s", event.Type)
	}
}
