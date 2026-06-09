package rooms

import (
	"testing"
	"time"

	"ggame/backend/internal/models"
)

func TestBotJoinsAndAnswers(t *testing.T) {
	manager := NewManager()
	room, host, err := manager.Create(CreateInput{
		ServerName: "Bot test",
		MaxPlayers: 2,
		GradeMode:  "9",
		GameMode:   "final_pvp",
		Nickname:   "Host",
		Grade:      9,
		Settings: models.Settings{
			RoundDurationSeconds: 20,
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
	if _, err := manager.Start(room.UniqueServerID, host.ID); err != nil {
		t.Fatal(err)
	}

	var bot *models.Player
	for _, player := range room.Players {
		if player.IsBot {
			bot = player
			break
		}
	}
	if bot == nil || bot.Team != models.OmniSoft {
		t.Fatalf("expected OmniSoft bot, got %#v", bot)
	}

	time.Sleep(3500 * time.Millisecond)
	if bot.CorrectAnswers+bot.WrongAnswers == 0 {
		t.Fatal("bot did not answer a question")
	}
}

func TestPlayersReceiveDistinctLanes(t *testing.T) {
	manager := NewManager()
	room, host, err := manager.Create(CreateInput{
		ServerName: "Lane test", MaxPlayers: 6, GradeMode: "9", GameMode: "final_pvp", Nickname: "Nex-1", Grade: 9,
		Settings: models.Settings{RoundDurationSeconds: 20, TowerHP: 100, TeamPlayerLimit: 3},
	})
	if err != nil {
		t.Fatal(err)
	}
	players := []*models.Player{host}
	for _, nickname := range []string{"Nex-2", "Nex-3", "Omni-1", "Omni-2", "Omni-3"} {
		_, player, joinErr := manager.Join(room.UniqueServerID, nickname, 9)
		if joinErr != nil {
			t.Fatal(joinErr)
		}
		players = append(players, player)
	}
	for i, player := range players {
		team := models.NexGen
		if i >= 3 {
			team = models.OmniSoft
		}
		if _, err := manager.SelectTeam(room.UniqueServerID, player.ID, team); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := manager.Start(room.UniqueServerID, host.ID); err != nil {
		t.Fatal(err)
	}
	for _, team := range []models.TeamName{models.NexGen, models.OmniSoft} {
		lanes := map[int]bool{}
		for _, unit := range room.Units {
			if unit.Team == team {
				lanes[unit.Lane] = true
			}
		}
		if len(lanes) != 3 {
			t.Fatalf("expected three distinct lanes for %s, got %#v", team, lanes)
		}
	}
}
