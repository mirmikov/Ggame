package rooms

import (
	"testing"

	"ggame/backend/internal/models"
)

func qualifierInput() CreateInput {
	return CreateInput{
		ServerName: "Qualifier test", MaxPlayers: 4, GradeMode: "mixed", GameMode: models.ModeQualifier,
		Nickname: "Organizer", Grade: 11,
		Settings: models.Settings{
			RoundDurationSeconds: 60, TowerHP: 200, TeamPlayerLimit: 2,
			ZoneStepsToCenter: 4, ZonePushbackSteps: 2, ZoneHoldSeconds: 5,
		},
	}
}

func TestOrganizerDoesNotBecomeQualifierTeam(t *testing.T) {
	manager := NewManager()
	room, organizer, err := manager.Create(qualifierInput())
	if err != nil {
		t.Fatal(err)
	}
	_, first, err := manager.Join(room.UniqueServerID, "Alpha", 10)
	if err != nil {
		t.Fatal(err)
	}
	_, second, err := manager.Join(room.UniqueServerID, "Beta", 10)
	if err != nil {
		t.Fatal(err)
	}
	started, err := manager.Start(room.UniqueServerID, organizer.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := started.Units[organizer.ID]; ok {
		t.Fatal("organizer must not receive an arena unit")
	}
	if _, ok := started.Units[first.ID]; !ok {
		t.Fatal("first qualifier team must receive an arena unit")
	}
	if _, ok := started.Units[second.ID]; !ok {
		t.Fatal("second qualifier team must receive an arena unit")
	}
	if _, ok := started.Units["ROBOT-BOSS"]; ok {
		t.Fatal("zone qualifier must not create a robot")
	}
	if started.QualifierSlots != 1 {
		t.Fatalf("expected one final slot, got %d", started.QualifierSlots)
	}
}

func TestQualifierRequiresEvenTeamCount(t *testing.T) {
	manager := NewManager()
	room, organizer, err := manager.Create(qualifierInput())
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = manager.Join(room.UniqueServerID, "Only one", 10); err != nil {
		t.Fatal(err)
	}
	if _, err = manager.Start(room.UniqueServerID, organizer.ID); err == nil {
		t.Fatal("expected an error for an odd number of qualifier teams")
	}
}

func TestZoneTakeoverPushesPreviousHolderBack(t *testing.T) {
	manager := NewManager()
	room, organizer, err := manager.Create(qualifierInput())
	if err != nil {
		t.Fatal(err)
	}
	_, first, _ := manager.Join(room.UniqueServerID, "Alpha", 10)
	_, second, _ := manager.Join(room.UniqueServerID, "Beta", 10)
	if _, err = manager.Start(room.UniqueServerID, organizer.ID); err != nil {
		t.Fatal(err)
	}

	manager.mu.Lock()
	live := manager.rooms[room.UniqueServerID]
	manager.advanceToZoneLocked(live, live.Players[first.ID], 4)
	if live.ZoneHolderID != first.ID {
		manager.mu.Unlock()
		t.Fatal("first team must capture the zone")
	}
	manager.advanceToZoneLocked(live, live.Players[second.ID], 4)
	firstLive := live.Players[first.ID]
	secondLive := live.Players[second.ID]
	manager.mu.Unlock()

	if secondLive.QualifierStatus != "holding" {
		t.Fatal("challenger must become the new holder")
	}
	if firstLive.ZoneSteps != 2 {
		t.Fatalf("previous holder must be pushed two steps back, got %d", firstLive.ZoneSteps)
	}
}

func TestQualifierFinishesWhenHalfTeamsQualify(t *testing.T) {
	manager := NewManager()
	room, organizer, err := manager.Create(qualifierInput())
	if err != nil {
		t.Fatal(err)
	}
	_, first, _ := manager.Join(room.UniqueServerID, "Alpha", 10)
	_, second, _ := manager.Join(room.UniqueServerID, "Beta", 10)
	if _, err = manager.Start(room.UniqueServerID, organizer.ID); err != nil {
		t.Fatal(err)
	}

	manager.mu.Lock()
	live := manager.rooms[room.UniqueServerID]
	manager.advanceToZoneLocked(live, live.Players[first.ID], 4)
	for i := 0; i < live.Settings.ZoneHoldSeconds; i++ {
		manager.tickQualifierLocked(live)
	}
	status := live.Status
	qualified := append([]string(nil), live.QualifiedTeamIDs...)
	firstStatus := live.Players[first.ID].QualifierStatus
	secondStatus := live.Players[second.ID].QualifierStatus
	manager.mu.Unlock()

	if status != "finished" {
		t.Fatalf("qualifier must finish after half the teams qualify, got %s", status)
	}
	if len(qualified) != 1 || qualified[0] != first.ID {
		t.Fatalf("unexpected qualified teams: %#v", qualified)
	}
	if firstStatus != "qualified" || secondStatus != "eliminated" {
		t.Fatalf("unexpected statuses: first=%s second=%s", firstStatus, secondStatus)
	}
}

func TestFinalRequiresBothTeams(t *testing.T) {
	manager := NewManager()
	in := qualifierInput()
	in.GameMode = models.ModeFinal
	room, organizer, err := manager.Create(in)
	if err != nil {
		t.Fatal(err)
	}
	_, first, err := manager.Join(room.UniqueServerID, "Nex", 10)
	if err != nil {
		t.Fatal(err)
	}
	_, second, err := manager.Join(room.UniqueServerID, "Omni", 11)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = manager.SelectTeam(room.UniqueServerID, first.ID, models.NexGen); err != nil {
		t.Fatal(err)
	}
	if _, err = manager.Start(room.UniqueServerID, organizer.ID); err == nil {
		t.Fatal("expected error while second team is empty")
	}
	if _, err = manager.SelectTeam(room.UniqueServerID, second.ID, models.OmniSoft); err != nil {
		t.Fatal(err)
	}
	started, err := manager.Start(room.UniqueServerID, organizer.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := started.Units["ROBOT-BOSS"]; ok {
		t.Fatal("final must not create robot")
	}
}

func TestQuestionsDoNotExposeExplanation(t *testing.T) {
	manager := NewManager()
	items := manager.Questions(10)
	if len(items) == 0 {
		t.Fatal("expected questions")
	}
	for _, q := range items {
		if q.Explanation != "" {
			t.Fatalf("explanation leaked for %s", q.ID)
		}
	}
}

func TestCodeTaskMovesQualifierTeam(t *testing.T) {
	manager := NewManager()
	room, organizer, err := manager.Create(qualifierInput())
	if err != nil {
		t.Fatal(err)
	}
	_, participant, err := manager.Join(room.UniqueServerID, "Coder", 10)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = manager.Join(room.UniqueServerID, "Other", 10); err != nil {
		t.Fatal(err)
	}
	if _, err = manager.Start(room.UniqueServerID, organizer.ID); err != nil {
		t.Fatal(err)
	}
	correct, current, err := manager.SubmitTask(room.UniqueServerID, participant.ID, "py-2", "s == s[::-1]")
	if err != nil {
		t.Fatal(err)
	}
	if !correct {
		t.Fatal("expected accepted code answer")
	}
	p := current.Players[participant.ID]
	if len(p.SolvedTasks) != 1 || p.Score == 0 || p.ZoneSteps != 2 {
		t.Fatalf("task movement not applied: %#v", p)
	}
}
