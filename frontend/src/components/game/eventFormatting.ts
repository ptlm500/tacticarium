import type { GameEvent } from "../../types/game";

/** Normalized event shape used by shared components. */
export interface NormalizedEvent {
  eventType: string;
  playerNumber?: number;
  round?: number;
  phase?: string;
  data?: Record<string, unknown>;
  createdAt?: string;
}

/** REST API event shape from /api/games/:id/events. */
export interface RestGameEvent {
  id: number;
  playerNumber: number | null;
  eventType: string;
  eventData: Record<string, unknown> | null;
  round: number | null;
  phase: string | null;
  createdAt: string;
}

export function normalizeWsEvent(e: GameEvent): NormalizedEvent {
  return {
    eventType: e.eventType,
    playerNumber: e.playerNumber,
    round: e.round,
    phase: e.phase,
    data: e.data,
    createdAt: e.createdAt,
  };
}

export function normalizeRestEvent(e: RestGameEvent): NormalizedEvent {
  return {
    eventType: e.eventType,
    playerNumber: e.playerNumber ?? undefined,
    round: e.round ?? undefined,
    phase: e.phase ?? undefined,
    data: e.eventData ?? undefined,
    createdAt: e.createdAt,
  };
}

export function formatEvent(event: NormalizedEvent): string {
  const player = event.playerNumber ? `P${event.playerNumber}` : "";

  switch (event.eventType) {
    case "phase_advance":
      return `${player} advanced to ${event.data?.to || "next"} phase`;
    case "cp_gain":
      return `${player} gained ${event.data?.amount || 1} CP`;
    case "cp_adjust":
      return `${player} adjusted CP by ${event.data?.delta}`;
    case "stratagem_used": {
      const spent = event.data?.cpSpent;
      const original = event.data?.originalCpCost;
      const suffix =
        typeof original === "number" && typeof spent === "number" && spent !== original
          ? `${spent} CP, was ${original}`
          : `${spent} CP`;
      return `${player} used ${event.data?.stratagemName} (${suffix})`;
    }
    case "vp_primary_score":
    case "vp_secondary_score":
    case "vp_gambit_score":
      return `${player} scored ${event.data?.delta} ${event.data?.category} VP`;
    case "secondary_achieved":
      return `${player} achieved ${event.data?.secondaryName} (+${event.data?.vpScored} VP)`;
    case "challenger_scored":
      return `${player} completed challenger mission (+${event.data?.vpScored} VP)`;
    case "game_start":
      return "Game started!";
    case "game_end":
      return `Game ended (${event.data?.reason})`;
    case "player_concede":
      return `${player} conceded`;
    case "abandon_requested":
      return `${player} requested to abandon the game`;
    case "abandon_rejected":
      return `${player} declined the abandon request`;
    default:
      return `${player} ${event.eventType}`;
  }
}

export const HIGHLIGHT_EVENT_TYPES = new Set([
  "vp_primary_score",
  "vp_secondary_score",
  "vp_gambit_score",
  "secondary_achieved",
  "challenger_scored",
  "stratagem_used",
  "player_concede",
  "game_end",
  "game_start",
  "abandon_requested",
]);

export function isHighlightEvent(e: NormalizedEvent): boolean {
  return HIGHLIGHT_EVENT_TYPES.has(e.eventType);
}
