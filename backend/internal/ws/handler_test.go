package ws

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ggame/backend/internal/models"
	"ggame/backend/internal/rooms"

	"github.com/gorilla/websocket"
)

func TestBotAndBattleCanBroadcastTogether(t *testing.T) {
	manager := rooms.NewManager()
	room, host, err := manager.Create(rooms.CreateInput{
		ServerName: "WebSocket bot test",
		MaxPlayers: 2,
		GradeMode:  "9",
		GameMode:   "final_pvp",
		Nickname:   "Host",
		Grade:      9,
		Settings: models.Settings{
			RoundDurationSeconds: 12,
			TowerHP:              100,
			TeamPlayerLimit:      1,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.SelectTeam(room.UniqueServerID, host.ID, models.NexGen); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.AddBot(room.UniqueServerID, host.ID); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(New(manager))
	defer server.Close()
	conn, _, err := websocket.DefaultDialer.Dial(strings.Replace(server.URL, "http", "ws", 1)+"/ws/rooms/"+room.UniqueServerID, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	readDone := make(chan struct{})
	go func() {
		defer close(readDone)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	if _, err := manager.Start(room.UniqueServerID, host.ID); err != nil {
		t.Fatal(err)
	}
	time.Sleep(8 * time.Second)

	current, ok := manager.Get(room.UniqueServerID)
	if !ok || current.Status != "running" {
		t.Fatalf("expected running room, got %#v", current)
	}
}
