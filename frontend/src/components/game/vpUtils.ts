import type { RestGameEvent } from "./eventFormatting";

export type { RestGameEvent };

export interface RoundVP {
  primary: number;
  secondary: number;
  gambit: number;
}

export interface PlayerSummaryStats {
  totalVP: number;
  vpByRound: Record<number, RoundVP>;
  stratagemsUsed: number;
  paint: number;
}

export function buildPlayerStats(
  events: RestGameEvent[],
  playerNumber: number,
  paint: number,
): PlayerSummaryStats {
  const vpByRound: Record<number, RoundVP> = {};
  let stratagemsUsed = 0;

  for (const e of events) {
    if (e.playerNumber !== playerNumber) continue;
    const round = e.round ?? 0;

    if (
      e.eventType === "vp_primary_score" ||
      e.eventType === "vp_secondary_score" ||
      e.eventType === "vp_gambit_score" ||
      e.eventType === "secondary_achieved" ||
      e.eventType === "challenger_scored"
    ) {
      if (!vpByRound[round]) vpByRound[round] = { primary: 0, secondary: 0, gambit: 0 };

      const delta = (e.eventData?.delta as number) ?? 0;
      const vpScored = (e.eventData?.vpScored as number) ?? 0;
      const amount = delta || vpScored;

      if (e.eventType === "vp_primary_score") {
        vpByRound[round].primary += amount;
      } else if (e.eventType === "vp_secondary_score" || e.eventType === "secondary_achieved") {
        vpByRound[round].secondary += amount;
      } else if (e.eventType === "vp_gambit_score" || e.eventType === "challenger_scored") {
        vpByRound[round].gambit += amount;
      }
    }

    if (e.eventType === "stratagem_used") {
      stratagemsUsed++;
    }
  }

  let totalVP = paint;
  for (const rv of Object.values(vpByRound)) {
    totalVP += rv.primary + rv.secondary + rv.gambit;
  }

  return { totalVP, vpByRound, stratagemsUsed, paint };
}

export function getEndReason(events: RestGameEvent[]): string | null {
  const endEvent = events.find((e) => e.eventType === "game_end");
  return (endEvent?.eventData?.reason as string) ?? null;
}

export function getRoundsPlayed(events: RestGameEvent[]): number {
  let max = 0;
  for (const e of events) {
    if (e.round != null && e.round > max) max = e.round;
  }
  return max;
}
