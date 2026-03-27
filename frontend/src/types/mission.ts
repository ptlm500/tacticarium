export interface MissionPack {
  id: string;
  name: string;
  description?: string;
}

export interface Mission {
  id: string;
  missionPackId: string;
  name: string;
  description?: string;
  deploymentMap?: string;
  rulesText?: string;
}

export interface Secondary {
  id: string;
  missionPackId: string;
  name: string;
  category: string;
  description: string;
  maxVp: number;
}

export interface Gambit {
  id: string;
  missionPackId: string;
  name: string;
  description: string;
  vpValue: number;
}
