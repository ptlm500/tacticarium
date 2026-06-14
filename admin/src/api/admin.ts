import type { components } from "../../../shared/api.generated";
import { api, uploadFile } from "./client";

type Schemas = components["schemas"];

// Types derived from OpenAPI schema, with id guaranteed present when read from API
export type Faction = Schemas["Faction"] & { id: string };
export type Detachment = Schemas["Detachment"] & { id: string };
export type Stratagem = Schemas["Stratagem"] & { id: string };

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
  // Stratagems are read-only in 11th edition: only the list endpoint exists.
  stratagems: {
    list: (params?: Record<string, string>) => {
      const qs = params ? "?" + new URLSearchParams(params).toString() : "";
      return api.get<Stratagem[]>(`${base}/stratagems${qs}`);
    },
  },

  import: {
    factions: (file: File) => uploadFile<ImportResult>(`${base}/import/factions`, file),
    detachments: (file: File) => uploadFile<ImportResult>(`${base}/import/detachments`, file),
    stratagems: (file: File) => uploadFile<ImportResult>(`${base}/import/stratagems`, file),
    missions: (file: File) => uploadFile<ImportResult>(`${base}/import/missions`, file),
  },
};
