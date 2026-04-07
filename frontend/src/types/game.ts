import type { components } from "./api.generated";

type Schemas = components["schemas"];

// --- Narrower union types (not expressed in OpenAPI spec) ---

export type Phase = "setup" | "command" | "movement" | "shooting" | "charge" | "fight";
export type GameStatus = "setup" | "active" | "completed" | "abandoned";

export const PHASE_ORDER: Phase[] = ["command", "movement", "shooting", "charge", "fight"];

export const PHASE_LABELS: Record<Phase, string> = {
  setup: "Setup",
  command: "Command",
  movement: "Movement",
  shooting: "Shooting",
  charge: "Charge",
  fight: "Fight",
};

// --- Types derived from OpenAPI schema ---

export type SecondaryObjective = Schemas["SecondaryObjective"];

export type ScoringOption = Schemas["ScoringOption"];

export type ActiveSecondary = Schemas["ActiveSecondary"];

export type PlayerState = Schemas["PlayerState"];

/** Game state with narrower Phase/GameStatus types and a 2-player tuple. */
export interface GameState {
  gameId: string;
  inviteCode: string;
  status: GameStatus;
  currentRound: number;
  currentTurn: number;
  currentPhase: Phase;
  activePlayer: number;
  firstTurnPlayer: number;
  missionPackId: string;
  missionId: string;
  missionName: string;
  twistId: string;
  twistName: string;
  players: [PlayerState | null, PlayerState | null];
  createdAt: string;
  completedAt?: string;
  winnerId?: string;
}

/**
 * Game event as received via WebSocket. Note: the HTTP endpoint
 * (/api/games/:id/events) returns a different shape with `eventData`
 * and a numeric `id` — see components["schemas"]["GameEvent"].
 */
export interface GameEvent {
  eventType: string;
  playerNumber?: number;
  round?: number;
  phase?: Phase;
  data?: Record<string, unknown>;
  createdAt?: string;
}

/** Game summary for list views, derived from OpenAPI schema. */
export type GameSummary = Omit<Schemas["GameSummary"], "status"> & {
  status: GameStatus;
};
