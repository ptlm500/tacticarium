import type { components } from "../../../shared/api.generated";

type Schemas = components["schemas"];

export type MissionPack = Schemas["MissionPack"] & { id: string };

export type ScoringAction = Schemas["ScoringAction"];

export type Mission = Schemas["Mission"] & { id: string };

export type ScoringOption = Schemas["ScoringOption"];

export type Secondary = Schemas["Secondary"] & { id: string };

export type MissionRule = Schemas["MissionRule"] & { id: string };

export type ChallengerCard = Schemas["ChallengerCard"] & { id: string };

export type Gambit = Schemas["Gambit"] & { id: string };
