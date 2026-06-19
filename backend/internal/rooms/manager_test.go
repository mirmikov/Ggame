package rooms

import (
	"fmt"
	"testing"

	"ggame/backend/internal/models"
)

func qualifierInput() CreateInput {
	return CreateInput{
		ServerName: "Qualifier test", MaxPlayers: 16, GradeMode: "9", GameMode: models.ModeQualifier,
		Nickname: "Organizer", Grade: 11,
		Settings: models.Settings{
			RoundDurationSeconds: 60, TowerHP: 200, TeamPlayerLimit: 2,
			ZoneStepsToCenter: 4, ZonePushbackSteps: 2, ZoneHoldSeconds: 5,
		},
	}
}

func joinEightTeams(t *testing.T, manager *Manager, roomID string) []*models.Player {
	t.Helper()
	players := make([]*models.Player, 0, models.QualifierTeamCount)
	for i := 1; i <= models.QualifierTeamCount; i++ {
		_, player, err := manager.Join(roomID, fmt.Sprintf("Player %d", i), 9+(i%3))
		if err != nil {
			t.Fatal(err)
		}
		if _, err = manager.SelectQualifierTeam(roomID, player.ID, fmt.Sprintf("T%d", i)); err != nil {
			t.Fatal(err)
		}
		players = append(players, player)
	}
	return players
}

func stopRoom(manager *Manager, roomID string) {
	manager.mu.Lock()
	if room := manager.rooms[roomID]; room != nil {
		room.Status = "finished"
	}
	manager.mu.Unlock()
}

func TestQualifierCreatesDefaultEightTeams(t *testing.T) {
	manager := NewManager()
	room, _, err := manager.Create(qualifierInput())
	if err != nil {
		t.Fatal(err)
	}
	if len(room.QualifierTeams) != models.QualifierTeamCount {
		t.Fatalf("expected 8 teams, got %d", len(room.QualifierTeams))
	}
	if room.MaxPlayers != models.QualifierTeamCount*room.Settings.TeamPlayerLimit {
		t.Fatalf("max players must be calculated from team capacity, got %d", room.MaxPlayers)
	}
}

func TestQualifierCanUseFewerTeams(t *testing.T) {
	manager := NewManager()
	in := qualifierInput()
	in.Settings.QualifierTeamCount = 4
	in.Settings.TeamPlayerLimit = 3
	room, _, err := manager.Create(in)
	if err != nil {
		t.Fatal(err)
	}
	if len(room.QualifierTeams) != 4 {
		t.Fatalf("expected 4 teams, got %d", len(room.QualifierTeams))
	}
	if room.MaxPlayers != 12 {
		t.Fatalf("expected 12 max players, got %d", room.MaxPlayers)
	}
	if room.QualifierTeams["T5"] != nil {
		t.Fatal("unexpected fifth qualifier team")
	}
}

func TestParticipantCanChooseQualifierTeam(t *testing.T) {
	manager := NewManager()
	room, _, err := manager.Create(qualifierInput())
	if err != nil {
		t.Fatal(err)
	}
	_, player, err := manager.Join(room.UniqueServerID, "Alice", 10)
	if err != nil {
		t.Fatal(err)
	}
	selected, err := manager.SelectQualifierTeam(room.UniqueServerID, player.ID, "T3")
	if err != nil {
		t.Fatal(err)
	}
	if selected.Players[player.ID].QualifierTeamID != "T3" {
		t.Fatal("participant did not join the selected qualifier team")
	}
}

func TestQualifierTeamCapacity(t *testing.T) {
	manager := NewManager()
	in := qualifierInput()
	in.Settings.TeamPlayerLimit = 1
	room, _, err := manager.Create(in)
	if err != nil {
		t.Fatal(err)
	}
	_, first, _ := manager.Join(room.UniqueServerID, "One", 10)
	_, second, _ := manager.Join(room.UniqueServerID, "Two", 10)
	if _, err = manager.SelectQualifierTeam(room.UniqueServerID, first.ID, "T1"); err != nil {
		t.Fatal(err)
	}
	if _, err = manager.SelectQualifierTeam(room.UniqueServerID, second.ID, "T1"); err == nil {
		t.Fatal("expected full team error")
	}
}

func TestQualifierStartsWithPartiallyFilledTeams(t *testing.T) {
	manager := NewManager()
	room, organizer, err := manager.Create(qualifierInput())
	if err != nil {
		t.Fatal(err)
	}
	_, player, _ := manager.Join(room.UniqueServerID, "Only one", 10)
	_, _ = manager.SelectQualifierTeam(room.UniqueServerID, player.ID, "T1")
	started, err := manager.Start(room.UniqueServerID, organizer.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer stopRoom(manager, room.UniqueServerID)
	if started.QualifierSlots != 1 {
		t.Fatalf("expected one final slot for one active team, got %d", started.QualifierSlots)
	}
	if started.QualifierTeams["T1"].Status != "active" {
		t.Fatal("filled team must be active")
	}
	if started.QualifierTeams["T2"].Status != "eliminated" {
		t.Fatal("empty team must be ignored")
	}
}

func TestQualifierStartsWithoutParticipants(t *testing.T) {
	manager := NewManager()
	room, organizer, err := manager.Create(qualifierInput())
	if err != nil {
		t.Fatal(err)
	}
	started, err := manager.Start(room.UniqueServerID, organizer.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer stopRoom(manager, room.UniqueServerID)
	if started.QualifierSlots != 0 {
		t.Fatalf("expected zero final slots without active teams, got %d", started.QualifierSlots)
	}
}

func TestQualifierStartsWithEightPopulatedTeams(t *testing.T) {
	manager := NewManager()
	room, organizer, err := manager.Create(qualifierInput())
	if err != nil {
		t.Fatal(err)
	}
	joinEightTeams(t, manager, room.UniqueServerID)
	started, err := manager.Start(room.UniqueServerID, organizer.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer stopRoom(manager, room.UniqueServerID)
	if started.QualifierSlots != 4 {
		t.Fatalf("expected four final slots, got %d", started.QualifierSlots)
	}
	if len(started.Units) != 0 {
		t.Fatal("qualifier arena must use team markers, not player battle units")
	}
}

func TestZoneTakeoverPushesPreviousTeamBack(t *testing.T) {
	manager := NewManager()
	room, organizer, err := manager.Create(qualifierInput())
	if err != nil {
		t.Fatal(err)
	}
	players := joinEightTeams(t, manager, room.UniqueServerID)
	if _, err = manager.Start(room.UniqueServerID, organizer.ID); err != nil {
		t.Fatal(err)
	}
	defer stopRoom(manager, room.UniqueServerID)

	manager.mu.Lock()
	live := manager.rooms[room.UniqueServerID]
	firstTeam := live.QualifierTeams["T1"]
	secondTeam := live.QualifierTeams["T2"]
	manager.advanceTeamToZoneLocked(live, firstTeam, live.Players[players[0].ID], 4)
	manager.advanceTeamToZoneLocked(live, secondTeam, live.Players[players[1].ID], 4)
	firstSteps := firstTeam.ZoneSteps
	secondStatus := secondTeam.Status
	holder := live.ZoneHolderTeamID
	manager.mu.Unlock()

	if holder != "T2" || secondStatus != "holding" {
		t.Fatal("challenger team must become the new zone holder")
	}
	if firstSteps != 2 {
		t.Fatalf("previous team must be pushed two steps back, got %d", firstSteps)
	}
}

func TestQualifierFinishesWhenFourTeamsQualify(t *testing.T) {
	manager := NewManager()
	room, organizer, err := manager.Create(qualifierInput())
	if err != nil {
		t.Fatal(err)
	}
	joinEightTeams(t, manager, room.UniqueServerID)
	if _, err = manager.Start(room.UniqueServerID, organizer.ID); err != nil {
		t.Fatal(err)
	}

	manager.mu.Lock()
	live := manager.rooms[room.UniqueServerID]
	for i := 1; i <= 4; i++ {
		manager.qualifyTeamLocked(live, live.QualifierTeams[fmt.Sprintf("T%d", i)])
	}
	manager.tickQualifierLocked(live)
	status := live.Status
	qualified := append([]string(nil), live.QualifiedTeamIDs...)
	manager.mu.Unlock()

	if status != "finished" {
		t.Fatalf("qualifier must finish after four teams qualify, got %s", status)
	}
	if len(qualified) != 4 {
		t.Fatalf("expected four qualified teams, got %#v", qualified)
	}
}

func TestCodeTaskMovesWholeQualifierTeam(t *testing.T) {
	manager := NewManager()
	room, organizer, err := manager.Create(qualifierInput())
	if err != nil {
		t.Fatal(err)
	}
	players := joinEightTeams(t, manager, room.UniqueServerID)
	_, teammate, err := manager.Join(room.UniqueServerID, "Teammate", 10)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = manager.SelectQualifierTeam(room.UniqueServerID, teammate.ID, "T1"); err != nil {
		t.Fatal(err)
	}
	if _, err = manager.Start(room.UniqueServerID, organizer.ID); err != nil {
		t.Fatal(err)
	}
	defer stopRoom(manager, room.UniqueServerID)
	correct, current, err := manager.SubmitTask(room.UniqueServerID, players[0].ID, "py-2", "s == s[::-1]")
	if err != nil {
		t.Fatal(err)
	}
	if !correct {
		t.Fatal("expected accepted code answer")
	}
	team := current.QualifierTeams["T1"]
	if team.ZoneSteps != 2 || current.Players[teammate.ID].ZoneSteps != 2 {
		t.Fatalf("team movement was not synchronized: team=%#v teammate=%#v", team, current.Players[teammate.ID])
	}
}

func TestFinalStartsWithOneSide(t *testing.T) {
	manager := NewManager()
	in := qualifierInput()
	in.GameMode = models.ModeFinal
	in.MaxPlayers = 4
	room, organizer, err := manager.Create(in)
	if err != nil {
		t.Fatal(err)
	}
	_, first, _ := manager.Join(room.UniqueServerID, "Nex", 10)
	_, second, _ := manager.Join(room.UniqueServerID, "Omni", 11)
	if _, err = manager.SelectTeam(room.UniqueServerID, first.ID, models.NexGen); err != nil {
		t.Fatal(err)
	}
	started, err := manager.Start(room.UniqueServerID, organizer.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer stopRoom(manager, room.UniqueServerID)
	if started.Players[second.ID].Team == "" {
		t.Fatal("unassigned final participant must be assigned automatically")
	}
	if len(started.Units) != 2 {
		t.Fatalf("expected units for both participants, got %d", len(started.Units))
	}
}

func TestQuestionsAndTasksAreExpanded(t *testing.T) {
	manager := NewManager()
	for grade := 9; grade <= 11; grade++ {
		items := manager.Questions(grade)
		if len(items) < 12 {
			t.Fatalf("expected at least 12 questions for grade %d, got %d", grade, len(items))
		}
		for _, q := range items {
			if q.Explanation != "" {
				t.Fatalf("explanation leaked for %s", q.ID)
			}
		}
	}
	if len(manager.Tasks()) < 18 {
		t.Fatalf("expected at least 18 code tasks, got %d", len(manager.Tasks()))
	}
}
