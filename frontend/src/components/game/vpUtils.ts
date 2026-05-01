import type { NormalizedEvent, RestGameEvent } from "./eventFormatting";
import type { PlayerState } from "../../types/game";

export type { RestGameEvent };

export type VPCategory = "primary" | "secondary";

export interface RoundVP {
  primary: number;
  secondary: number;
}

export interface PlayerSummaryStats {
  totalVP: number;
  vpByRound: Record<number, RoundVP>;
  stratagemsUsed: number;
  paint: number;
}

export function buildPlayerStats(
  events: NormalizedEvent[],
  playerNumber: number,
  paint: number,
): PlayerSummaryStats {
  const vpByRound: Record<number, RoundVP> = {};
  let stratagemsUsed = 0;

  const bucket = (round: number) => {
    if (!vpByRound[round]) vpByRound[round] = { primary: 0, secondary: 0 };
    return vpByRound[round];
  };

  for (const e of events) {
    if (e.playerNumber !== playerNumber) continue;
    const round = e.round ?? 0;

    if (
      e.eventType === "vp_primary_score" ||
      e.eventType === "vp_secondary_score" ||
      e.eventType === "secondary_achieved"
    ) {
      const delta = (e.data?.delta as number) ?? 0;
      const appliedDelta = (e.data?.appliedDelta as number | undefined) ?? null;
      const vpScored = (e.data?.vpScored as number) ?? 0;
      const amount = appliedDelta ?? (delta || vpScored);

      const rv = bucket(round);
      if (e.eventType === "vp_primary_score") {
        rv.primary += amount;
      } else if (e.eventType === "vp_secondary_score" || e.eventType === "secondary_achieved") {
        rv.secondary += amount;
      }
    }

    if (e.eventType === "vp_primary_score_reverted") {
      const revertedRound = (e.data?.revertedRound as number | undefined) ?? round;
      const revertedDelta = (e.data?.revertedDelta as number | undefined) ?? 0;
      bucket(revertedRound).primary -= revertedDelta;
    }

    if (e.eventType === "vp_manual_adjust") {
      const category = e.data?.category as string | undefined;
      const appliedDelta = (e.data?.appliedDelta as number | undefined) ?? 0;
      if (category === "primary") bucket(round).primary += appliedDelta;
      else if (category === "secondary") bucket(round).secondary += appliedDelta;
    }

    if (e.eventType === "stratagem_used") {
      stratagemsUsed++;
    }
  }

  let totalVP = paint;
  for (const rv of Object.values(vpByRound)) {
    totalVP += rv.primary + rv.secondary;
  }

  return { totalVP, vpByRound, stratagemsUsed, paint };
}

export function getEndReason(events: NormalizedEvent[]): string | null {
  const endEvent = events.find((e) => e.eventType === "game_end");
  return (endEvent?.data?.reason as string) ?? null;
}

export function getRoundsPlayed(events: NormalizedEvent[]): number {
  let max = 0;
  for (const e of events) {
    if (e.round != null && e.round > max) max = e.round;
  }
  return max;
}

export function computeIntensityMax(...stats: (PlayerSummaryStats | null)[]): number {
  let max = 0;
  for (const s of stats) {
    if (!s) continue;
    for (const rv of Object.values(s.vpByRound)) {
      if (rv.primary > max) max = rv.primary;
      if (rv.secondary > max) max = rv.secondary;
    }
  }
  return max || 1;
}

export interface ScoringHeatmapData {
  normalizedEvents: NormalizedEvent[];
  statsByPlayerNumber: Record<number, PlayerSummaryStats>;
  rounds: number[];
  intensityMax: number;
}

export function buildScoringHeatmapData(
  normalizedEvents: NormalizedEvent[],
  players: (PlayerState | null)[],
  options: { roundCount?: number } = {},
): ScoringHeatmapData {
  const statsByPlayerNumber: Record<number, PlayerSummaryStats> = {};
  const allStats: PlayerSummaryStats[] = [];
  for (const p of players) {
    if (!p) continue;
    const stats = buildPlayerStats(normalizedEvents, p.playerNumber, p.vpPaint);
    statsByPlayerNumber[p.playerNumber] = stats;
    allStats.push(stats);
  }

  const roundCount = options.roundCount ?? getRoundsPlayed(normalizedEvents);
  const rounds = Array.from({ length: Math.max(roundCount, 0) }, (_, i) => i + 1);

  return {
    normalizedEvents,
    statsByPlayerNumber,
    rounds,
    intensityMax: computeIntensityMax(...allStats),
  };
}
