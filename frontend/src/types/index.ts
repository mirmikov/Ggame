export type TeamName = "NexGen" | "OmniSoft" | "";
export type Role = "organizer" | "participant";
export type GameMode = "qualifier" | "final";

export interface Player {
  id: string;
  nickname: string;
  grade: number;
  role: Role;
  team: TeamName;
  level: number;
  xp: number;
  hp: number;
  maxHp: number;
  attack: number;
  defense: number;
  speed: number;
  score: number;
  correctAnswers: number;
  wrongAnswers: number;
  wrongStreak: number;
  lockedUntil: number;
  isHost: boolean;
  isBot: boolean;
  questionId: string;
  solvedTasks: string[];
  latestBuff?: string;
  qualifierStatus?: "active" | "holding" | "qualified" | "eliminated" | "";
  zoneSteps?: number;
  captureProgress?: number;
}

export interface Team {
  name: TeamName;
  towerHp: number;
  maxTowerHp: number;
  score: number;
}

export interface Unit {
  ownerPlayerId: string;
  nickname: string;
  team: TeamName;
  hp: number;
  maxHp: number;
  attack: number;
  defense: number;
  speed: number;
  level: number;
  lane: number;
  position: number;
  target: string;
  respawnAt?: number;
  isBoss?: boolean;
}

export interface Projectile {
  id: string;
  team: TeamName;
  from: number;
  to: number;
  fromLane: number;
  toLane: number;
  damage: number;
  target: string;
  hitTower: boolean;
}

export interface Room {
  uniqueServerId: string;
  serverName: string;
  maxPlayers: number;
  gradeMode: string;
  gameMode: GameMode;
  status: "waiting" | "running" | "finished";
  organizerId: string;
  players: Record<string, Player>;
  teams: Record<string, Team>;
  units: Record<string, Unit>;
  projectiles: Projectile[];
  settings: {
    roundDurationSeconds: number;
    towerHp: number;
    teamPlayerLimit: number;
    zoneStepsToCenter: number;
    zonePushbackSteps: number;
    zoneHoldSeconds: number;
  };
  endsAt?: number;
  winner?: TeamName;
  storyMessage?: string;
  lastEvent?: string;
  zoneHolderId?: string;
  qualifierSlots?: number;
  qualifiedTeamIds?: string[];
}

export interface Question {
  id: string;
  grade: number;
  topic: string;
  text: string;
  options: string[];
  explanation?: string;
  timeLimitSeconds: number;
  difficulty: number;
}

export interface TerminalTask {
  id: string;
  title: string;
  description: string;
  language: string;
  starterCode: string;
  reward: number;
  difficulty: number;
}

export interface Session {
  playerId: string;
  nickname: string;
  grade: number;
  roomId?: string;
  role?: Role;
}
