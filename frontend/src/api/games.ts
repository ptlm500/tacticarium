import { api } from "./client";
import { GameState, GameSummary } from "../types/game";

export const gamesApi = {
  create: () => api.post<{ id: string; inviteCode: string }>("/api/games"),
  join: (code: string) => api.post<{ id: string; inviteCode: string }>(`/api/games/join/${code}`),
  get: (id: string) => api.get<GameState>(`/api/games/${id}`),
  list: () => api.get<GameSummary[]>("/api/games"),
  getHistory: () => api.get<GameSummary[]>("/api/users/me/history"),
  getEvents: (id: string) => api.get<unknown[]>(`/api/games/${id}/events`),
};
