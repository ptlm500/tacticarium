import { api } from "./client";
import { GameState, GameSummary, UserStats } from "../types/game";

export const gamesApi = {
  create: () => api.post<{ id: string; inviteCode: string }>("/api/games"),
  join: (code: string) => api.post<{ id: string; inviteCode: string }>(`/api/games/join/${code}`),
  get: (id: string) => api.get<GameState>(`/api/games/${id}`),
  list: () => api.get<GameSummary[]>("/api/games"),
  getHistory: (filters?: { myFaction?: string; opponentFaction?: string }) => {
    const params = new URLSearchParams();
    if (filters?.myFaction) params.set("myFaction", filters.myFaction);
    if (filters?.opponentFaction) params.set("opponentFaction", filters.opponentFaction);
    const qs = params.toString();
    return api.get<GameSummary[]>(`/api/users/me/history${qs ? `?${qs}` : ""}`);
  },
  getEvents: (id: string) => api.get<unknown[]>(`/api/games/${id}/events`),
  getStats: () => api.get<UserStats>("/api/users/me/stats"),
  hide: (id: string) => api.post<void>(`/api/games/${id}/hide`),
};
