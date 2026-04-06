import { api, uploadFile } from "./client";

// Types matching backend models
export interface Faction {
  id: string;
  name: string;
  wahapediaLink?: string;
}

export interface Detachment {
  id: string;
  factionId: string;
  name: string;
}

export interface Stratagem {
  id: string;
  factionId: string;
  detachmentId?: string;
  name: string;
  type: string;
  cpCost: number;
  legend?: string;
  turn: string;
  phase: string;
  description: string;
}

export interface MissionPack {
  id: string;
  name: string;
  description?: string;
}

export interface ScoringAction {
  label: string;
  vp: number;
  minRound?: number;
  description?: string;
  scoringTiming?: string;
}

export interface Mission {
  id: string;
  missionPackId: string;
  name: string;
  lore?: string;
  description: string;
  scoringRules: ScoringAction[];
  scoringTiming: string;
}

export interface ScoringOption {
  label: string;
  vp: number;
  mode?: string;
}

export interface Secondary {
  id: string;
  missionPackId: string;
  name: string;
  lore?: string;
  description: string;
  maxVp: number;
  isFixed: boolean;
  scoringOptions: ScoringOption[];
}

export interface Gambit {
  id: string;
  missionPackId: string;
  name: string;
  description: string;
  vpValue: number;
}

export interface ChallengerCard {
  id: string;
  missionPackId: string;
  name: string;
  lore?: string;
  description: string;
}

export interface MissionRule {
  id: string;
  missionPackId: string;
  name: string;
  lore?: string;
  description: string;
}

export interface ImportResult {
  entity: string;
  [key: string]: unknown;
}

const base = "/api/admin";

function crud<T>(path: string) {
  return {
    list: (params?: Record<string, string>) => {
      const qs = params ? "?" + new URLSearchParams(params).toString() : "";
      return api.get<T[]>(`${base}${path}${qs}`);
    },
    get: (id: string) => api.get<T>(`${base}${path}/${id}`),
    create: (data: T) => api.post<T>(`${base}${path}`, data),
    update: (id: string, data: Partial<T>) => api.put<T>(`${base}${path}/${id}`, data),
    delete: (id: string) => api.del(`${base}${path}/${id}`),
  };
}

export const adminApi = {
  factions: crud<Faction>("/factions"),
  detachments: crud<Detachment>("/detachments"),
  stratagems: crud<Stratagem>("/stratagems"),
  missionPacks: crud<MissionPack>("/mission-packs"),
  missions: crud<Mission>("/missions"),
  secondaries: crud<Secondary>("/secondaries"),
  gambits: crud<Gambit>("/gambits"),
  challengerCards: crud<ChallengerCard>("/challenger-cards"),
  missionRules: crud<MissionRule>("/mission-rules"),

  import: {
    factions: (file: File) => uploadFile<ImportResult>(`${base}/import/factions`, file),
    stratagems: (file: File) => uploadFile<ImportResult>(`${base}/import/stratagems`, file),
    missions: (file: File) => uploadFile<ImportResult>(`${base}/import/missions`, file),
  },
};
