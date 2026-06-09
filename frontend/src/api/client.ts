import type { Question, Room, Session, TeamName, TerminalTask } from "../types";

const API = import.meta.env.VITE_API_URL || (import.meta.env.DEV ? "http://localhost:8080" : window.location.origin);

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const response = await fetch(`${API}${path}`, {
    ...options,
    headers: { "Content-Type": "application/json", ...options?.headers },
  });
  const body = await response.json();
  if (!response.ok) throw new Error(body.error || "Ошибка сети");
  return body;
}

export const api = {
  createRoom: (data: unknown) => request<{room: Room; player: Session & {id: string}}>("/api/rooms", { method: "POST", body: JSON.stringify(data) }),
  joinRoom: (id: string, nickname: string, grade: number) => request<{room: Room; player: {id: string}}>(`/api/rooms/${id}/join`, { method: "POST", body: JSON.stringify({nickname, grade}) }),
  chooseTeam: (id: string, playerId: string, team: TeamName) => request<Room>(`/api/rooms/${id}/team`, { method: "POST", body: JSON.stringify({playerId, team}) }),
  addBot: (id: string, playerId: string) => request<Room>(`/api/rooms/${id}/bot`, { method: "POST", body: JSON.stringify({playerId}) }),
  start: (id: string, playerId: string) => request<Room>(`/api/rooms/${id}/start`, { method: "POST", body: JSON.stringify({playerId}) }),
  answer: (id: string, playerId: string, questionId: string, answer: number) => request<{correct: boolean; explanation: string; room: Room}>(`/api/rooms/${id}/answer`, { method: "POST", body: JSON.stringify({playerId, questionId, answer}) }),
  questions: (grade: number) => request<Question[]>(`/api/questions?grade=${grade}`),
  tasks: () => request<TerminalTask[]>("/api/tasks"),
  submitTask: (id: string, playerId: string, taskId: string, answer: string) => request<{correct: boolean; room: Room}>(`/api/rooms/${id}/task`, { method: "POST", body: JSON.stringify({playerId, taskId, answer}) }),
};

export function roomSocket(roomId: string) {
  const url = API.replace(/^http/, "ws");
  return new WebSocket(`${url}/ws/rooms/${roomId}`);
}
