export const queryKeys = {
  games: {
    all: ["games"] as const,
    list: () => ["games", "list"] as const,
    detail: (id: string) => ["games", "detail", id] as const,
    events: (id: string) => ["games", "events", id] as const,
  },
  history: {
    all: ["history"] as const,
    list: (filters?: { myFaction?: string; opponentFaction?: string }) =>
      ["history", "list", filters] as const,
    stats: () => ["history", "stats"] as const,
  },
  factions: {
    all: ["factions"] as const,
    list: () => ["factions", "list"] as const,
    detachments: (factionId: string) => ["factions", "detachments", factionId] as const,
    stratagems: (factionId: string) => ["factions", "stratagems", factionId] as const,
  },
  missions: {
    all: ["missions"] as const,
    list: (packId: string) => ["missions", "list", packId] as const,
    rules: (packId: string) => ["missions", "rules", packId] as const,
    secondaries: (packId: string) => ["missions", "secondaries", packId] as const,
  },
} as const;
