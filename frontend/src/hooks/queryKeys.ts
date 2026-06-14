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
    list: () => ["missions", "list"] as const,
    forceDispositions: () => ["missions", "force-dispositions"] as const,
    matchups: () => ["missions", "matchups"] as const,
    secondaryCards: () => ["missions", "secondary-cards"] as const,
    deploymentPatterns: () => ["missions", "deployment-patterns"] as const,
  },
} as const;
