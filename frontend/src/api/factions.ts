import { api } from "./client";
import { Faction, Detachment, Stratagem } from "../types/faction";

export const factionsApi = {
  list: () => api.get<Faction[]>("/api/factions"),
  getDetachments: (factionId: string) =>
    api.get<Detachment[]>(`/api/factions/${factionId}/detachments`),
  getStratagems: (factionId: string) =>
    api.get<Stratagem[]>(`/api/factions/${factionId}/stratagems`),
  getDetachmentStratagems: (detachmentId: string) =>
    api.get<Stratagem[]>(`/api/detachments/${detachmentId}/stratagems`),
};
