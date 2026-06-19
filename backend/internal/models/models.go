package models

import "time"

type TeamName string

const (
	NexGen   TeamName = "NexGen"
	OmniSoft TeamName = "OmniSoft"
)

const (
	RoleOrganizer      = "organizer"
	RoleParticipant    = "participant"
	ModeQualifier      = "qualifier"
	ModeFinal          = "final"
	QualifierTeamCount = 8
)

type Player struct {
	ID              string   `json:"id"`
	Nickname        string   `json:"nickname"`
	Grade           int      `json:"grade"`
	Role            string   `json:"role"`
	Team            TeamName `json:"team"`
	QualifierTeamID string   `json:"qualifierTeamId,omitempty"`
	Level           int      `json:"level"`
	XP              int      `json:"xp"`
	HP              int      `json:"hp"`
	MaxHP           int      `json:"maxHp"`
	Attack          int      `json:"attack"`
	Defense         int      `json:"defense"`
	Speed           float64  `json:"speed"`
	Score           int      `json:"score"`
	CorrectAnswers  int      `json:"correctAnswers"`
	WrongAnswers    int      `json:"wrongAnswers"`
	WrongStreak     int      `json:"wrongStreak"`
	LockedUntil     int64    `json:"lockedUntil"`
	IsHost          bool     `json:"isHost"`
	IsBot           bool     `json:"isBot"`
	QuestionID      string   `json:"questionId"`
	SolvedTasks     []string `json:"solvedTasks"`
	LatestBuff      string   `json:"latestBuff,omitempty"`
	QualifierStatus string   `json:"qualifierStatus,omitempty"`
	ZoneSteps       int      `json:"zoneSteps,omitempty"`
	CaptureProgress int      `json:"captureProgress,omitempty"`
}

type BattleUnit struct {
	OwnerPlayerID string   `json:"ownerPlayerId"`
	Nickname      string   `json:"nickname"`
	Team          TeamName `json:"team"`
	HP            int      `json:"hp"`
	MaxHP         int      `json:"maxHp"`
	Attack        int      `json:"attack"`
	Defense       int      `json:"defense"`
	Speed         float64  `json:"speed"`
	Level         int      `json:"level"`
	Lane          int      `json:"lane"`
	Position      float64  `json:"position"`
	Target        string   `json:"target"`
	RespawnAt     int64    `json:"respawnAt,omitempty"`
	IsBoss        bool     `json:"isBoss,omitempty"`
}

type Projectile struct {
	ID       string   `json:"id"`
	Team     TeamName `json:"team"`
	From     float64  `json:"from"`
	To       float64  `json:"to"`
	FromLane int      `json:"fromLane"`
	ToLane   int      `json:"toLane"`
	Damage   int      `json:"damage"`
	Target   string   `json:"target"`
	HitTower bool     `json:"hitTower"`
}

type Team struct {
	Name       TeamName `json:"name"`
	TowerHP    int      `json:"towerHp"`
	MaxTowerHP int      `json:"maxTowerHp"`
	Score      int      `json:"score"`
}

type QualifierTeam struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Hue             int    `json:"hue"`
	Lane            int    `json:"lane"`
	Score           int    `json:"score"`
	ZoneSteps       int    `json:"zoneSteps"`
	CaptureProgress int    `json:"captureProgress"`
	Status          string `json:"status"`
}

type Settings struct {
	RoundDurationSeconds int `json:"roundDurationSeconds"`
	TowerHP              int `json:"towerHp"`
	TeamPlayerLimit      int `json:"teamPlayerLimit"`
	ZoneStepsToCenter    int `json:"zoneStepsToCenter"`
	ZonePushbackSteps    int `json:"zonePushbackSteps"`
	ZoneHoldSeconds      int `json:"zoneHoldSeconds"`
}

type Room struct {
	UniqueServerID   string                    `json:"uniqueServerId"`
	ServerName       string                    `json:"serverName"`
	MaxPlayers       int                       `json:"maxPlayers"`
	GradeMode        string                    `json:"gradeMode"`
	GameMode         string                    `json:"gameMode"`
	Status           string                    `json:"status"`
	OrganizerID      string                    `json:"organizerId"`
	Players          map[string]*Player        `json:"players"`
	Teams            map[TeamName]*Team        `json:"teams"`
	QualifierTeams   map[string]*QualifierTeam `json:"qualifierTeams"`
	Units            map[string]*BattleUnit    `json:"units"`
	Projectiles      []Projectile              `json:"projectiles"`
	CreatedAt        time.Time                 `json:"createdAt"`
	Settings         Settings                  `json:"settings"`
	EndsAt           int64                     `json:"endsAt,omitempty"`
	Winner           TeamName                  `json:"winner,omitempty"`
	StoryMessage     string                    `json:"storyMessage,omitempty"`
	LastEvent        string                    `json:"lastEvent,omitempty"`
	ZoneHolderID     string                    `json:"zoneHolderId,omitempty"`
	ZoneHolderTeamID string                    `json:"zoneHolderTeamId,omitempty"`
	QualifierSlots   int                       `json:"qualifierSlots,omitempty"`
	QualifiedTeamIDs []string                  `json:"qualifiedTeamIds,omitempty"`
}

type Question struct {
	ID               string   `json:"id"`
	Grade            int      `json:"grade"`
	Topic            string   `json:"topic"`
	Text             string   `json:"text"`
	Options          []string `json:"options"`
	CorrectAnswer    int      `json:"-"`
	Explanation      string   `json:"explanation,omitempty"`
	TimeLimitSeconds int      `json:"timeLimitSeconds"`
	Difficulty       int      `json:"difficulty"`
}

type TerminalTask struct {
    ID              string   `json:"id"`
    Title           string   `json:"title"`
    Description     string   `json:"description"`
    Language        string   `json:"language"`
    StarterCode     string   `json:"starterCode"`
    AcceptedAnswers []string `json:"-"`
    Reward          int      `json:"reward"`
    Difficulty      int      `json:"difficulty"`
    Grade           int      `json:"grade"` // 9, 10 или 11
}

type Event struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}
