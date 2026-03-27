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
