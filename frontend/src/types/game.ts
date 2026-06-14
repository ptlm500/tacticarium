import type { components } from "../../../shared/api.generated";

type Schemas = components["schemas"];

// --- Narrower union types (not expressed in OpenAPI spec) ---

// 11th edition wraps the five phases in Start-of-Turn / End-of-Turn steps.
export type Phase =
  | "setup"
  | "start_of_turn"
  | "command"
  | "movement"
  | "shooting"
  | "charge"
  | "fight"
  | "end_of_turn";
export type GameStatus = "setup" | "active" | "completed" | "abandoned";

export const PHASE_ORDER: Phase[] = [
  "start_of_turn",
  "command",
  "movement",
  "shooting",
  "charge",
  "fight",
  "end_of_turn",
];

export const PHASE_LABELS: Record<Phase, string> = {
  setup: "Setup",
  start_of_turn: "Start of Turn",
  command: "Command",
  movement: "Movement",
  shooting: "Shooting",
  charge: "Charge",
  fight: "Fight",
  end_of_turn: "End of Turn",
};

// --- Types derived from OpenAPI schema ---

/** A mission/secondary card in a player's deck, hand, or scored pile. */
export type SecondaryCard = Schemas["SecondaryCard"];

/** A mission/secondary card definition (awards DSL + prose). */
export type Card = Schemas["Card"];

/** An outstanding Layer-2 scoring prompt awaiting player confirmation. */
export type ScorePrompt = Schemas["ScorePrompt"];

/** The battlefield board: objectives + per-player sides. */
export type Board = Schemas["Board"];

/** A single board objective (position, role, control, tags). */
export type Objective = Schemas["Objective"];

/** A detachment a player has taken (id, name, and its DP cost). */
export type SelectedDetachment = Schemas["SelectedDetachment"];

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
  board: Board;
  vpPerGameCap: number;
  vpPerRoundCap: number;
  startOfTurnControl?: Record<string, number>;
  players: [PlayerState | null, PlayerState | null];
  createdAt: string;
  completedAt?: string;
  winnerId?: string;
  abandonRequestedBy?: number;
}

/**
 * Game event as received via WebSocket. Note: the HTTP endpoint
 * (/api/games/:id/events) returns a different shape with `eventData`
 * and a numeric `id` — see components["schemas"]["GameEvent"].
 */
export interface GameEvent {
  /** Persisted game_events.id assigned by the backend; used to dedupe between
   * REST history and live WebSocket events. */
  id?: number;
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

export type FactionStat = Schemas["FactionStat"];

export type UserStats = Schemas["UserStats"];
