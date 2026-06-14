import type { components } from "../../../shared/api.generated";

type Schemas = components["schemas"];

/** A force disposition (attacker/defender deployment posture). */
export type ForceDisposition = Schemas["ForceDisposition"] & { id: string };

/** A primary mission (11th edition). */
export type Mission = Schemas["Mission"] & { id: string };

/** A mission/disposition matchup (which mission a disposition pairing plays). */
export type MissionMatchup = Schemas["MissionMatchup"] & { id: string };

/** A mission card (e.g. secondary mission card). */
export type MissionCard = Schemas["MissionCard"] & { id: string };

/** A deployment pattern (board layout: objectives, territories, zones). */
export type DeploymentPattern = Schemas["DeploymentPattern"] & { id: string };

/**
 * A single primary-scoring action shown in the scoring UI.
 *
 * This is a UI-only construct used by the scoring prompt / quick-score
 * components. It is no longer sourced from the reference-data API (the old
 * 10e Mission.scoringRules shape was removed in the 11e migration).
 */
export interface ScoringAction {
  label: string;
  vp: number;
  minRound?: number;
  scoringTiming?: string;
  description?: string;
}
