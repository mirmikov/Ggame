package rooms

import (
	"encoding/json"
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

func basePlayer(nickname string, grade int, role string, host bool) *models.Player {
	if grade < 9 || grade > 11 {
		grade = 9
	}
	return &models.Player{
		ID: randomID("P", 6), Nickname: strings.TrimSpace(nickname), Grade: grade,
		Role: role, Level: 1, HP: 34, MaxHP: 34, Attack: 7, Defense: 2, Speed: 1,
		IsHost: host, SolvedTasks: []string{},
	}
}

func (m *Manager) Create(in CreateInput) (*models.Room, *models.Player, error) {
	if in.GameMode != models.ModeQualifier && in.GameMode != models.ModeFinal {
		in.GameMode = models.ModeQualifier
	}
	if in.GameMode == models.ModeQualifier {
		if in.Settings.TeamPlayerLimit < 1 || in.Settings.TeamPlayerLimit > 8 {
			in.Settings.TeamPlayerLimit = 4
		}
		in.MaxPlayers = models.QualifierTeamCount * in.Settings.TeamPlayerLimit
	}
	if in.MaxPlayers < 1 || in.MaxPlayers > 64 {
		return nil, nil, errors.New("число участников должно быть от 1 до 64")
	}
	if in.GameMode == models.ModeFinal && (in.Settings.TeamPlayerLimit < 1 || in.Settings.TeamPlayerLimit > in.MaxPlayers) {
		in.Settings.TeamPlayerLimit = max(1, (in.MaxPlayers+1)/2)
	}
	if strings.TrimSpace(in.Nickname) == "" {
		return nil, nil, errors.New("укажите имя организатора")
	}
	if in.Settings.TowerHP < 50 || in.Settings.TowerHP > 3000 {
		in.Settings.TowerHP = 260
	}
	if in.Settings.RoundDurationSeconds < 30 || in.Settings.RoundDurationSeconds > 3600 {
		in.Settings.RoundDurationSeconds = 600
	}
	if in.Settings.ZoneStepsToCenter < 4 || in.Settings.ZoneStepsToCenter > 20 {
		in.Settings.ZoneStepsToCenter = 8
	}
	if in.Settings.ZonePushbackSteps < 1 || in.Settings.ZonePushbackSteps >= in.Settings.ZoneStepsToCenter {
		in.Settings.ZonePushbackSteps = 2
	}
	if in.Settings.ZoneHoldSeconds < 5 || in.Settings.ZoneHoldSeconds > 120 {
		in.Settings.ZoneHoldSeconds = 15
	}

	roomID := randomID("CYB-", 5)
	organizer := basePlayer(in.Nickname, in.Grade, models.RoleOrganizer, true)
	room := &models.Room{
		UniqueServerID: roomID, ServerName: strings.TrimSpace(in.ServerName), MaxPlayers: in.MaxPlayers,
		GradeMode: in.GradeMode, GameMode: in.GameMode, Status: "waiting", OrganizerID: organizer.ID,
		Players: map[string]*models.Player{organizer.ID: organizer}, Units: map[string]*models.BattleUnit{},
		Projectiles: []models.Projectile{}, CreatedAt: time.Now(), Settings: in.Settings,
		QualifierTeams: defaultQualifierTeams(),
		Teams: map[models.TeamName]*models.Team{
			models.NexGen:   {Name: models.NexGen, TowerHP: in.Settings.TowerHP, MaxTowerHP: in.Settings.TowerHP},
			models.OmniSoft: {Name: models.OmniSoft, TowerHP: in.Settings.TowerHP, MaxTowerHP: in.Settings.TowerHP},
		},
		StoryMessage: "Организатор создал сессию. Участники подключаются по коду.",
	}
	if room.ServerName == "" {
		room.ServerName = "Информатический турнир"
	}

	m.mu.Lock()
	m.rooms[roomID] = room
	m.clients[roomID] = map[*websocket.Conn]bool{}
	snapshot := cloneRoom(room)
	player := clonePlayer(organizer)
	m.mu.Unlock()
	return snapshot, player, nil
}

func (m *Manager) Join(roomID, nickname string, grade int) (*models.Room, *models.Player, error) {
	if strings.TrimSpace(nickname) == "" {
		return nil, nil, errors.New("укажите имя участника")
	}
	m.mu.Lock()
	room, ok := m.rooms[strings.ToUpper(roomID)]
	if !ok {
		m.mu.Unlock()
		return nil, nil, errors.New("лобби не найдено")
	}
	if room.Status != "waiting" {
		m.mu.Unlock()
		return nil, nil, errors.New("тур уже запущен")
	}
	if participantCount(room) >= room.MaxPlayers {
		m.mu.Unlock()
		return nil, nil, errors.New("в лобби нет свободных мест")
	}
	player := basePlayer(nickname, grade, models.RoleParticipant, false)
	room.Players[player.ID] = player
	room.LastEvent = fmt.Sprintf("%s подключился к турниру", player.Nickname)
	snapshot := cloneRoom(room)
	playerCopy := clonePlayer(player)
	m.mu.Unlock()
	m.Broadcast(roomID, "room_state", snapshot)
	return snapshot, playerCopy, nil
}

// AddBot оставлен для совместимости старого API.
func (m *Manager) AddBot(roomID, playerID string) (*models.Room, error) {
	return nil, errors.New("в режиме захвата зоны боты не используются")
}

func (m *Manager) Get(roomID string) (*models.Room, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	room, ok := m.rooms[strings.ToUpper(roomID)]
	if !ok {
		return nil, false
	}
	return cloneRoom(room), true
}

func (m *Manager) SelectQualifierTeam(roomID, playerID, teamID string) (*models.Room, error) {
	m.mu.Lock()
	room, ok := m.rooms[strings.ToUpper(roomID)]
	if !ok {
		m.mu.Unlock()
		return nil, errors.New("лобби не найдено")
	}
	if room.Status != "waiting" {
		m.mu.Unlock()
		return nil, errors.New("тур уже запущен")
	}
	if room.GameMode != models.ModeQualifier {
		m.mu.Unlock()
		return nil, errors.New("выбор одной из восьми команд доступен только в отборочном туре")
	}
	player := room.Players[playerID]
	if player == nil || player.Role != models.RoleParticipant {
		m.mu.Unlock()
		return nil, errors.New("команду может выбрать только участник")
	}
	team := room.QualifierTeams[teamID]
	if team == nil {
		m.mu.Unlock()
		return nil, errors.New("неизвестная команда")
	}
	if qualifierTeamMemberCount(room, teamID) >= room.Settings.TeamPlayerLimit && player.QualifierTeamID != teamID {
		m.mu.Unlock()
		return nil, errors.New("в этой команде нет свободных мест")
	}
	player.QualifierTeamID = teamID
	room.LastEvent = fmt.Sprintf("%s вступил в команду %s", player.Nickname, team.Name)
	snapshot := cloneRoom(room)
	m.mu.Unlock()
	m.Broadcast(roomID, "room_state", snapshot)
	return snapshot, nil
}

func (m *Manager) SelectTeam(roomID, playerID string, team models.TeamName) (*models.Room, error) {
	if team != models.NexGen && team != models.OmniSoft {
		return nil, errors.New("неизвестная команда")
	}
	m.mu.Lock()
	room, ok := m.rooms[strings.ToUpper(roomID)]
	if !ok {
		m.mu.Unlock()
		return nil, errors.New("лобби не найдено")
	}
	if room.Status != "waiting" {
		m.mu.Unlock()
		return nil, errors.New("тур уже запущен")
	}
	if room.GameMode != models.ModeFinal {
		m.mu.Unlock()
		return nil, errors.New("в отборочном туре каждая подключённая команда играет сама за себя")
	}
	player := room.Players[playerID]
	if player == nil || player.Role != models.RoleParticipant {
		m.mu.Unlock()
		return nil, errors.New("команду может выбрать только участник")
	}
	if teamCount(room, team) >= room.Settings.TeamPlayerLimit && player.Team != team {
		m.mu.Unlock()
		return nil, errors.New("команда заполнена")
	}
	player.Team = team
	room.LastEvent = fmt.Sprintf("%s выбрал команду %s", player.Nickname, team)
	snapshot := cloneRoom(room)
	m.mu.Unlock()
	m.Broadcast(roomID, "room_state", snapshot)
	return snapshot, nil
}

func (m *Manager) Start(roomID, playerID string) (*models.Room, error) {
	m.mu.Lock()
	room, ok := m.rooms[strings.ToUpper(roomID)]
	if !ok {
		m.mu.Unlock()
		return nil, errors.New("лобби не найдено")
	}
	organizer := room.Players[playerID]
	if organizer == nil || organizer.Role != models.RoleOrganizer || !organizer.IsHost {
		m.mu.Unlock()
		return nil, errors.New("только организатор может запустить тур")
	}
	if room.Status != "waiting" {
		m.mu.Unlock()
		return nil, errors.New("тур уже запущен")
	}
	participants := sortedParticipants(room)
	if len(participants) == 0 {
		m.mu.Unlock()
		return nil, errors.New("нужен хотя бы один участник")
	}

	// Для отборочного режима подсчитываем команды, в которых есть участники
	activeTeams := 0
	for _, team := range room.QualifierTeams {
		if qualifierTeamMemberCount(room, team.ID) > 0 {
			activeTeams++
		}
	}
	if room.GameMode == models.ModeQualifier {
		if activeTeams < 2 {
			m.mu.Unlock()
			return nil, fmt.Errorf("для запуска нужно минимум 2 команды с участниками, сейчас только %d", activeTeams)
		}
	}

	room.Units = map[string]*models.BattleUnit{}
	room.Projectiles = []models.Projectile{}
	room.Winner = ""
	room.ZoneHolderID = ""
	room.ZoneHolderTeamID = ""
	room.QualifiedTeamIDs = []string{}
	room.QualifierSlots = 0
	for _, team := range room.Teams {
		team.TowerHP = team.MaxTowerHP
		team.Score = 0
	}

	if room.GameMode == models.ModeFinal {
		// финальная логика без изменений
		hasNex, hasOmni := false, false
		teamLanes := map[models.TeamName]int{models.NexGen: 0, models.OmniSoft: 0}
		for _, p := range participants {
			if p.Team == "" {
				m.mu.Unlock()
				return nil, errors.New("в финале каждый участник должен выбрать сторону")
			}
			hasNex = hasNex || p.Team == models.NexGen
			hasOmni = hasOmni || p.Team == models.OmniSoft
			p.QuestionID = m.nextQuestionID(p.Grade, "")
			p.QualifierStatus, p.ZoneSteps, p.CaptureProgress = "", 0, 0
			lane := teamLanes[p.Team]
			teamLanes[p.Team]++
			room.Units[p.ID] = unitFromPlayer(p, lane)
		}
		if !hasNex || !hasOmni {
			m.mu.Unlock()
			return nil, errors.New("в финале нужен минимум один участник с каждой стороны")
		}
		room.StoryMessage = "ФИНАЛ: две стороны усиливают бойцов теорией и кодом. Организатор транслирует арену."
	} else {
		// Отборочный режим: настройка команд
		room.QualifierSlots = 4
		if activeTeams < 4 {
			room.QualifierSlots = activeTeams // все проходят
		}
		for _, team := range room.QualifierTeams {
			team.Score = 0
			team.ZoneSteps = 0
			team.CaptureProgress = 0
			// Команды без участников помечаем как eliminated, чтобы не участвовали
			if qualifierTeamMemberCount(room, team.ID) == 0 {
				team.Status = "eliminated"
			} else {
				team.Status = "active"
			}
		}
		for _, p := range participants {
			if room.QualifierTeams[p.QualifierTeamID] == nil || room.QualifierTeams[p.QualifierTeamID].Status == "eliminated" {
				m.mu.Unlock()
				return nil, fmt.Errorf("участник %s не выбрал активную команду", p.Nickname)
			}
			p.Team = ""
			p.QuestionID = m.nextQuestionID(p.Grade, "")
			p.QualifierStatus = "active"
			p.ZoneSteps = 0
			p.CaptureProgress = 0
			p.LatestBuff = "СТАРТ // ОТБОРОЧНЫЙ ТУР"
		}
		room.StoryMessage = "ОТБОРОЧНЫЙ ТУР: команды движутся к центральной зоне. В финал проходят лучшие."
	}

	room.Status = "running"
	room.EndsAt = time.Now().Add(time.Duration(room.Settings.RoundDurationSeconds) * time.Second).Unix()
	room.LastEvent = "Тур запущен"
	snapshot := cloneRoom(room)
	m.mu.Unlock()
	m.Broadcast(roomID, "game_started", snapshot)
	go m.battleLoop(strings.ToUpper(roomID))
	return snapshot, nil
}

func unitFromPlayer(p *models.Player, lane int) *models.BattleUnit {
	start := 8.0
	if p.Team == models.OmniSoft {
		start = 92
	}
	return &models.BattleUnit{OwnerPlayerID: p.ID, Nickname: p.Nickname, Team: p.Team, HP: p.HP, MaxHP: p.MaxHP, Attack: p.Attack, Defense: p.Defense, Speed: p.Speed, Level: p.Level, Lane: lane, Position: start}
}

func (m *Manager) Answer(roomID, playerID, questionID string, answer int) (bool, string, *models.Room, error) {
	m.mu.Lock()
	room, ok := m.rooms[strings.ToUpper(roomID)]
	if !ok || room.Status != "running" {
		m.mu.Unlock()
		return false, "", nil, errors.New("тур не запущен")
	}
	p := room.Players[playerID]
	if p == nil || p.Role != models.RoleParticipant || p.IsBot {
		m.mu.Unlock()
		return false, "", nil, errors.New("задания доступны только участникам")
	}
	var qualifierTeam *models.QualifierTeam
	if room.GameMode == models.ModeQualifier {
		qualifierTeam = room.QualifierTeams[p.QualifierTeamID]
		if qualifierTeam == nil {
			m.mu.Unlock()
			return false, "", nil, errors.New("участник не состоит в команде")
		}
		if qualifierTeam.Status == "qualified" || qualifierTeam.Status == "eliminated" {
			m.mu.Unlock()
			return false, "", nil, errors.New("для этой команды отбор уже завершён")
		}
	}
	if time.Now().Unix() < p.LockedUntil {
		m.mu.Unlock()
		return false, "", nil, errors.New("теоретический модуль временно заблокирован")
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
		p.Score += 120 * q.Difficulty
		p.XP += 45 * q.Difficulty
		if room.GameMode == models.ModeQualifier {
			qualifierTeam.Score += 120 * q.Difficulty
			steps := 1
			if q.Difficulty >= 3 {
				steps = 2
			}
			m.advanceTeamToZoneLocked(room, qualifierTeam, p, steps)
		} else {
			room.Teams[p.Team].Score += 120 * q.Difficulty
			buff := m.applyRandomBuffLocked(room, p, q.Difficulty)
			p.LatestBuff = buff
			room.LastEvent = fmt.Sprintf("%s решил задачу и получил %s", p.Nickname, buff)
			m.applyLevelLocked(room, p)
		}
	} else {
		p.WrongAnswers++
		p.WrongStreak++
		room.LastEvent = fmt.Sprintf("%s ошибся в теоретическом модуле", p.Nickname)
		if p.WrongStreak >= 3 {
			p.LockedUntil = time.Now().Add(45 * time.Second).Unix()
			p.WrongStreak = 0
		}
	}
	p.QuestionID = m.nextQuestionID(p.Grade, questionID)
	snapshot := cloneRoom(room)
	m.mu.Unlock()
	m.Broadcast(roomID, "player_answered", map[string]any{"playerId": playerID, "correct": correct, "room": snapshot})
	return correct, q.Explanation, snapshot, nil
}

func (m *Manager) Questions(grade int) []models.Question {
	items := questions.ForGrade(grade)
	for i := range items {
		items[i].Explanation = ""
	}
	return items
}

func (m *Manager) Tasks() []models.TerminalTask { return questions.Tasks() }

func (m *Manager) SubmitTask(roomID, playerID, taskID, answer string) (bool, *models.Room, error) {
	m.mu.Lock()
	room, ok := m.rooms[strings.ToUpper(roomID)]
	if !ok || room.Status != "running" {
		m.mu.Unlock()
		return false, nil, errors.New("тур не запущен")
	}
	p := room.Players[playerID]
	if p == nil || p.Role != models.RoleParticipant || p.IsBot {
		m.mu.Unlock()
		return false, nil, errors.New("кодовые задания доступны только участникам")
	}
	var qualifierTeam *models.QualifierTeam
	if room.GameMode == models.ModeQualifier {
		qualifierTeam = room.QualifierTeams[p.QualifierTeamID]
		if qualifierTeam == nil {
			m.mu.Unlock()
			return false, nil, errors.New("участник не состоит в команде")
		}
		if qualifierTeam.Status == "qualified" || qualifierTeam.Status == "eliminated" {
			m.mu.Unlock()
			return false, nil, errors.New("для этой команды отбор уже завершён")
		}
	}
	for _, solved := range p.SolvedTasks {
		if solved == taskID {
			m.mu.Unlock()
			return false, nil, errors.New("эта задача уже решена")
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
		return false, nil, errors.New("кодовая задача не найдена")
	}
	actual := normalizeCode(answer)
	correct := false
	for _, accepted := range task.AcceptedAnswers {
		if actual == normalizeCode(accepted) {
			correct = true
			break
		}
	}
	if correct {
		p.SolvedTasks = append(p.SolvedTasks, taskID)
		p.Score += task.Reward
		p.XP += task.Reward / 5
		if room.GameMode == models.ModeQualifier {
			qualifierTeam.Score += task.Reward
			steps := 2
			if task.Difficulty >= 3 {
				steps = 3
			}
			m.advanceTeamToZoneLocked(room, qualifierTeam, p, steps)
		} else {
			room.Teams[p.Team].Score += task.Reward
			buff := m.applyRandomBuffLocked(room, p, max(2, task.Difficulty))
			p.LatestBuff = buff
			room.LastEvent = fmt.Sprintf("%s прошёл кодовый тест и получил %s", p.Nickname, buff)
			m.applyLevelLocked(room, p)
		}
	} else {
		p.WrongAnswers++
		room.LastEvent = fmt.Sprintf("Код %s не прошёл проверку", p.Nickname)
	}
	snapshot := cloneRoom(room)
	m.mu.Unlock()
	m.Broadcast(roomID, "player_answered", map[string]any{"playerId": playerID, "correct": correct, "room": snapshot})
	return correct, snapshot, nil
}

func normalizeCode(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.TrimSuffix(value, ";")
	return strings.Join(strings.Fields(value), "")
}

func (m *Manager) applyRandomBuffLocked(room *models.Room, p *models.Player, power int) string {
	u := room.Units[p.ID]
	switch rand.Intn(4) {
	case 0:
		amount := 1 + power
		p.Attack += amount
		if u != nil {
			u.Attack = p.Attack
		}
		return fmt.Sprintf("АТАКА +%d", amount)
	case 1:
		amount := max(1, power)
		p.Defense += amount
		if u != nil {
			u.Defense = p.Defense
		}
		return fmt.Sprintf("ЗАЩИТА +%d", amount)
	case 2:
		amount := 4 + power*2
		p.MaxHP += amount
		p.HP += amount
		if u != nil {
			u.MaxHP = p.MaxHP
			u.HP = min(u.MaxHP, u.HP+amount)
		}
		return fmt.Sprintf("HP +%d", amount)
	default:
		amount := .05 + float64(power)*.03
		p.Speed += amount
		if u != nil {
			u.Speed = p.Speed
		}
		return fmt.Sprintf("СКОРОСТЬ +%.2f", amount)
	}
}

func (m *Manager) applyLevelLocked(room *models.Room, p *models.Player) {
	for p.XP >= p.Level*160 {
		p.XP -= p.Level * 160
		p.Level++
		p.MaxHP += 4
		p.HP = p.MaxHP
		if u := room.Units[p.ID]; u != nil {
			u.Level = p.Level
			u.MaxHP = p.MaxHP
			u.HP = p.MaxHP
		}
	}
}

func (m *Manager) nextQuestionID(grade int, previous string) string {
	items := questions.ForGrade(grade)
	if len(items) == 0 {
		items = questions.All()
	}
	for i := 0; i < 6; i++ {
		next := items[rand.Intn(len(items))].ID
		if next != previous {
			return next
		}
	}
	return items[0].ID
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
		if room.GameMode == models.ModeQualifier {
			m.tickQualifierLocked(room)
		} else {
			m.tickFinalLocked(room)
		}
		finished := room.Status == "finished"
		snapshot := cloneRoom(room)
		m.mu.Unlock()
		event := "battle_tick"
		if finished {
			event = "game_finished"
		}
		m.Broadcast(roomID, event, snapshot)
		if finished {
			return
		}
	}
}

func (m *Manager) tickQualifierLocked(room *models.Room) {
	now := time.Now().Unix()
	if room.ZoneHolderTeamID != "" {
		holder := room.QualifierTeams[room.ZoneHolderTeamID]
		if holder == nil || holder.Status != "holding" {
			room.ZoneHolderTeamID = ""
		} else {
			holder.CaptureProgress++
			m.syncQualifierTeamMembersLocked(room, holder, fmt.Sprintf("ЗАХВАТ ЗОНЫ // %d/%d", holder.CaptureProgress, room.Settings.ZoneHoldSeconds))
			if holder.CaptureProgress >= room.Settings.ZoneHoldSeconds {
				m.qualifyTeamLocked(room, holder)
			}
		}
	}
	if len(room.QualifiedTeamIDs) >= room.QualifierSlots {
		m.finishQualifierLocked(room, false)
		return
	}
	if now >= room.EndsAt {
		m.finishQualifierLocked(room, true)
	}
}

func (m *Manager) advanceTeamToZoneLocked(room *models.Room, team *models.QualifierTeam, contributor *models.Player, steps int) {
	if team == nil || team.Status == "qualified" || team.Status == "eliminated" {
		return
	}
	total := room.Settings.ZoneStepsToCenter
	if room.ZoneHolderTeamID == team.ID {
		bonus := max(1, steps)
		team.CaptureProgress = min(room.Settings.ZoneHoldSeconds, team.CaptureProgress+bonus)
		m.syncQualifierTeamMembersLocked(room, team, fmt.Sprintf("УДЕРЖАНИЕ +%d // %d/%d", bonus, team.CaptureProgress, room.Settings.ZoneHoldSeconds))
		room.LastEvent = fmt.Sprintf("%s укрепляет контроль зоны для команды %s", contributor.Nickname, team.Name)
		return
	}
	team.ZoneSteps = min(total, team.ZoneSteps+steps)
	m.syncQualifierTeamMembersLocked(room, team, fmt.Sprintf("ПРОДВИЖЕНИЕ +%d // %d/%d", steps, team.ZoneSteps, total))
	if team.ZoneSteps < total {
		room.LastEvent = fmt.Sprintf("%s продвинул команду %s: %d/%d шагов", contributor.Nickname, team.Name, team.ZoneSteps, total)
		return
	}

	if room.ZoneHolderTeamID != "" && room.ZoneHolderTeamID != team.ID {
		old := room.QualifierTeams[room.ZoneHolderTeamID]
		if old != nil && old.Status == "holding" {
			old.Status = "active"
			old.CaptureProgress = 0
			old.ZoneSteps = max(0, total-room.Settings.ZonePushbackSteps)
			m.syncQualifierTeamMembersLocked(room, old, fmt.Sprintf("ВЫТЕСНЕНЫ ИЗ ЗОНЫ // −%d шага", room.Settings.ZonePushbackSteps))
			room.LastEvent = fmt.Sprintf("Команда %s вытеснила %s из центральной зоны", team.Name, old.Name)
		}
	} else {
		room.LastEvent = fmt.Sprintf("Команда %s первой вошла в центральную зону", team.Name)
	}
	room.ZoneHolderTeamID = team.ID
	team.Status = "holding"
	team.CaptureProgress = 0
	m.syncQualifierTeamMembersLocked(room, team, fmt.Sprintf("ЗОНА ЗАХВАЧЕНА // удерживайте %d сек", room.Settings.ZoneHoldSeconds))
}

func (m *Manager) syncQualifierTeamMembersLocked(room *models.Room, team *models.QualifierTeam, message string) {
	for _, p := range room.Players {
		if p.Role != models.RoleParticipant || p.QualifierTeamID != team.ID {
			continue
		}
		p.QualifierStatus = team.Status
		p.ZoneSteps = team.ZoneSteps
		p.CaptureProgress = team.CaptureProgress
		if message != "" {
			p.LatestBuff = message
		}
	}
}

func (m *Manager) qualifyTeamLocked(room *models.Room, team *models.QualifierTeam) {
	team.Status = "qualified"
	team.CaptureProgress = room.Settings.ZoneHoldSeconds
	room.QualifiedTeamIDs = append(room.QualifiedTeamIDs, team.ID)
	room.ZoneHolderTeamID = ""
	m.syncQualifierTeamMembersLocked(room, team, "ФИНАЛ // КВАЛИФИКАЦИЯ ПОЛУЧЕНА")
	room.LastEvent = fmt.Sprintf("Команда %s удержала зону и проходит в финал", team.Name)
}
func (m *Manager) finishQualifierLocked(room *models.Room, byTimeout bool) {
	if byTimeout && len(room.QualifiedTeamIDs) < room.QualifierSlots {
		qualified := map[string]bool{}
		for _, id := range room.QualifiedTeamIDs {
			qualified[id] = true
		}
		// Собираем только команды с участниками и ещё не квалифицированные
		candidates := make([]*models.QualifierTeam, 0)
		for _, team := range sortedQualifierTeams(room) {
			if !qualified[team.ID] && qualifierTeamMemberCount(room, team.ID) > 0 {
				candidates = append(candidates, team)
			}
		}
		sort.SliceStable(candidates, func(i, j int) bool {
			if candidates[i].CaptureProgress != candidates[j].CaptureProgress {
				return candidates[i].CaptureProgress > candidates[j].CaptureProgress
			}
			if candidates[i].ZoneSteps != candidates[j].ZoneSteps {
				return candidates[i].ZoneSteps > candidates[j].ZoneSteps
			}
			if candidates[i].Score != candidates[j].Score {
				return candidates[i].Score > candidates[j].Score
			}
			return candidates[i].ID < candidates[j].ID
		})
		for _, team := range candidates {
			if len(room.QualifiedTeamIDs) >= room.QualifierSlots {
				break
			}
			team.Status = "qualified"
			room.QualifiedTeamIDs = append(room.QualifiedTeamIDs, team.ID)
			m.syncQualifierTeamMembersLocked(room, team, "ФИНАЛ // ПРОХОД ПО РЕЙТИНГУ")
		}
	}
	qualified := map[string]bool{}
	names := make([]string, 0, len(room.QualifiedTeamIDs))
	for _, id := range room.QualifiedTeamIDs {
		qualified[id] = true
		if team := room.QualifierTeams[id]; team != nil {
			team.Status = "qualified"
			names = append(names, team.Name)
			m.syncQualifierTeamMembersLocked(room, team, "ФИНАЛ // КОМАНДА ПРОШЛА")
		}
	}
	// Остальные команды, которые не прошли, помечаем как eliminated
	for _, team := range room.QualifierTeams {
		if !qualified[team.ID] {
			team.Status = "eliminated"
			m.syncQualifierTeamMembersLocked(room, team, "ОТБОР ЗАВЕРШЁН")
		}
	}
	room.ZoneHolderID = ""
	room.ZoneHolderTeamID = ""
	room.Status = "finished"
	room.Winner = models.NexGen // победитель в отборе не определяется
	room.StoryMessage = fmt.Sprintf("В финал проходят %d команды: %s.", room.QualifierSlots, strings.Join(names, ", "))
	if byTimeout {
		room.LastEvent = "QUALIFIER // TIMEOUT RANKING"
	} else {
		room.LastEvent = "QUALIFIER // FOUR TEAMS REMAIN" // можно оставить для совместимости
	}
}

func (m *Manager) tickFinalLocked(room *models.Room) {
	now := time.Now().Unix()
	room.Projectiles = room.Projectiles[:0]
	ids := sortedUnitIDs(room)
	for _, id := range ids {
		u := room.Units[id]
		if u == nil {
			continue
		}
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
		for _, enemyID := range ids {
			enemy := room.Units[enemyID]
			if enemy == nil {
				continue
			}
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
			damage := max(1, u.Attack-closest.Defense/2)
			u.Target = closest.OwnerPlayerID
			closest.HP -= damage
			room.Projectiles = append(room.Projectiles, models.Projectile{ID: randomID("S-", 5), Team: u.Team, From: u.Position, To: closest.Position, FromLane: u.Lane, ToLane: closest.Lane, Damage: damage, Target: closest.OwnerPlayerID})
			if closest.HP <= 0 {
				closest.HP, closest.RespawnAt = 0, now+5
			}
		} else if math.Abs(u.Position-towerPosition) <= 7 {
			damage := max(1, u.Attack)
			u.Target = string(enemyTower.Name) + "-tower"
			enemyTower.TowerHP -= damage
			room.Projectiles = append(room.Projectiles, models.Projectile{ID: randomID("S-", 5), Team: u.Team, From: u.Position, To: towerPosition, FromLane: u.Lane, ToLane: -1, Damage: damage, Target: string(enemyTower.Name) + "-tower", HitTower: true})
		} else {
			u.Target = ""
			u.Position += direction * u.Speed * 2.2
		}
	}
	if room.Teams[models.NexGen].TowerHP <= 0 || room.Teams[models.OmniSoft].TowerHP <= 0 || now >= room.EndsAt {
		room.Status = "finished"
		nexHP, omniHP := room.Teams[models.NexGen].TowerHP, room.Teams[models.OmniSoft].TowerHP
		if nexHP == omniHP {
			if room.Teams[models.NexGen].Score >= room.Teams[models.OmniSoft].Score {
				room.Winner = models.NexGen
			} else {
				room.Winner = models.OmniSoft
			}
		} else if nexHP > omniHP {
			room.Winner = models.NexGen
		} else {
			room.Winner = models.OmniSoft
		}
		room.StoryMessage = fmt.Sprintf("Финал завершён. Победитель — %s.", room.Winner)
		room.LastEvent = "FINAL // COMPLETE"
	}
}

func defaultQualifierTeams() map[string]*models.QualifierTeam {
	names := []string{
		"Байт Форс",
		"Квант",
		"Синтаксис",
		"Алгоритм",
		"Нейрон",
		"Пиксель",
		"Вектор",
		"Кибер Волки",
	}
	teams := make(map[string]*models.QualifierTeam, len(names))
	for i, name := range names {
		id := fmt.Sprintf("T%d", i+1)
		teams[id] = &models.QualifierTeam{ID: id, Name: name, Hue: (i * 43) % 360, Lane: i, Status: "waiting"}
	}
	return teams
}

func sortedQualifierTeams(room *models.Room) []*models.QualifierTeam {
	teams := make([]*models.QualifierTeam, 0, len(room.QualifierTeams))
	for _, team := range room.QualifierTeams {
		teams = append(teams, team)
	}
	sort.Slice(teams, func(i, j int) bool {
		if teams[i].Lane != teams[j].Lane {
			return teams[i].Lane < teams[j].Lane
		}
		return teams[i].ID < teams[j].ID
	})
	return teams
}

func qualifierTeamMemberCount(room *models.Room, teamID string) int {
	count := 0
	for _, p := range room.Players {
		if p.Role == models.RoleParticipant && !p.IsBot && p.QualifierTeamID == teamID {
			count++
		}
	}
	return count
}

func participantCount(room *models.Room) int {
	count := 0
	for _, p := range room.Players {
		if p.Role == models.RoleParticipant && !p.IsBot {
			count++
		}
	}
	return count
}
func sortedParticipants(room *models.Room) []*models.Player {
	players := make([]*models.Player, 0)
	for _, p := range room.Players {
		if p.Role == models.RoleParticipant && !p.IsBot {
			players = append(players, p)
		}
	}
	sort.Slice(players, func(i, j int) bool { return players[i].ID < players[j].ID })
	return players
}
func sortedUnitIDs(room *models.Room) []string {
	ids := make([]string, 0, len(room.Units))
	for id := range room.Units {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
func teamCount(room *models.Room, team models.TeamName) int {
	count := 0
	for _, p := range room.Players {
		if p.Role == models.RoleParticipant && !p.IsBot && p.Team == team {
			count++
		}
	}
	return count
}

func (m *Manager) AddClient(roomID string, conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	roomID = strings.ToUpper(roomID)
	if m.clients[roomID] == nil {
		m.clients[roomID] = map[*websocket.Conn]bool{}
	}
	m.clients[roomID][conn] = true
}
func (m *Manager) RemoveClient(roomID string, conn *websocket.Conn) {
	m.mu.Lock()
	delete(m.clients[strings.ToUpper(roomID)], conn)
	m.mu.Unlock()
	m.writeMu.Lock()
	defer m.writeMu.Unlock()
	_ = conn.Close()
}
func (m *Manager) SendClient(conn *websocket.Conn, event models.Event) error {
	m.writeMu.Lock()
	defer m.writeMu.Unlock()
	_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	return conn.WriteJSON(event)
}
func (m *Manager) Broadcast(roomID, eventType string, payload any) {
	m.mu.RLock()
	clients := make([]*websocket.Conn, 0, len(m.clients[strings.ToUpper(roomID)]))
	for c := range m.clients[strings.ToUpper(roomID)] {
		clients = append(clients, c)
	}
	m.mu.RUnlock()
	event := models.Event{Type: eventType, Payload: payload}
	m.writeMu.Lock()
	defer m.writeMu.Unlock()
	for _, c := range clients {
		_ = c.SetWriteDeadline(time.Now().Add(5 * time.Second))
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
func cloneRoom(room *models.Room) *models.Room {
	if room == nil {
		return nil
	}
	data, _ := json.Marshal(room)
	var copy models.Room
	_ = json.Unmarshal(data, &copy)
	return &copy
}
func clonePlayer(player *models.Player) *models.Player {
	if player == nil {
		return nil
	}
	copy := *player
	copy.SolvedTasks = append([]string(nil), player.SolvedTasks...)
	return &copy
}

const letters = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func randomID(prefix string, length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return fmt.Sprintf("%s%s", prefix, b)
}
