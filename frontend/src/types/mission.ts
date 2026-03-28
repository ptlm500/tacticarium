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
}

export interface Mission {
  id: string;
  missionPackId: string;
  name: string;
  lore: string;
  description: string;
  scoringRules: ScoringAction[];
}

export interface Secondary {
  id: string;
  missionPackId: string;
  name: string;
  lore: string;
  description: string;
  maxVp: number;
  isFixed: boolean;
}

export interface MissionRule {
  id: string;
  missionPackId: string;
  name: string;
  lore: string;
  description: string;
}

export interface ChallengerCard {
  id: string;
  missionPackId: string;
  name: string;
  lore: string;
  description: string;
}

export interface Gambit {
  id: string;
  missionPackId: string;
  name: string;
  description: string;
  vpValue: number;
}
