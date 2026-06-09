export type TeamName = "NexGen" | "OmniSoft" | "";
export interface Player {
  id: string; nickname: string; grade: number; team: TeamName; level: number; xp: number;
  hp: number; maxHp: number; attack: number; defense: number; speed: number; score: number;
  correctAnswers: number; wrongAnswers: number; lockedUntil: number; isHost: boolean; isBot: boolean; questionId: string; solvedTasks: string[];
}
export interface Team { name: TeamName; towerHp: number; maxTowerHp: number; score: number }
export interface Unit {
  ownerPlayerId: string; nickname: string; team: TeamName; hp: number; maxHp: number;
  attack: number; defense: number; speed: number; level: number; lane: number; position: number; target: string;
}
export interface Projectile { id: string; team: TeamName; from: number; to: number; fromLane: number; toLane: number; damage: number; target: string; hitTower: boolean }
export interface Room {
  uniqueServerId: string; serverName: string; maxPlayers: number; gradeMode: string; gameMode: string;
  status: "waiting" | "running" | "finished"; players: Record<string, Player>; teams: Record<string, Team>;
  units: Record<string, Unit>; projectiles: Projectile[]; settings: { roundDurationSeconds: number; towerHp: number; teamPlayerLimit: number };
  endsAt?: number; winner?: TeamName; storyMessage?: string;
}
export interface Question {
  id: string; grade: number; topic: string; text: string; options: string[];
  explanation?: string; timeLimitSeconds: number; difficulty: number;
}
export interface TerminalTask { id: string; title: string; description: string; reward: number }
export interface Session { playerId: string; nickname: string; grade: number; roomId?: string }
