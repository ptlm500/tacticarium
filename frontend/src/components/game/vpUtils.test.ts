import { describe, it, expect } from "vite-plus/test";
import { buildPlayerStats } from "./vpUtils";
import type { NormalizedEvent } from "./eventFormatting";

function ev(partial: Partial<NormalizedEvent>): NormalizedEvent {
  return {
    playerNumber: 1,
    eventType: "",
    round: 1,
    phase: "command",
    ...partial,
  };
}

describe("buildPlayerStats", () => {
  it("uses appliedDelta when present (handles clamp divergence)", () => {
    const events: NormalizedEvent[] = [
      ev({
        eventType: "vp_primary_score",
        round: 1,
        data: { category: "primary", delta: 10, appliedDelta: 2 },
      }),
    ];
    const stats = buildPlayerStats(events, 1, 0);
    expect(stats.vpByRound[1].primary).toBe(2);
    expect(stats.totalVP).toBe(2);
  });

  it("subtracts reverted primary scores", () => {
    const events: NormalizedEvent[] = [
      ev({
        eventType: "vp_primary_score",
        round: 1,
        data: { category: "primary", delta: 5, appliedDelta: 5 },
      }),
      ev({
        eventType: "vp_primary_score_reverted",
        round: 2,
        data: { revertedRound: 1, scoringSlot: "end_of_command_phase", revertedDelta: 5 },
      }),
    ];
    const stats = buildPlayerStats(events, 1, 0);
    expect(stats.vpByRound[1].primary).toBe(0);
    expect(stats.totalVP).toBe(0);
  });

  it("includes vp_manual_adjust deltas for primary and secondary", () => {
    const events: NormalizedEvent[] = [
      ev({
        eventType: "vp_manual_adjust",
        round: 1,
        data: { category: "primary", delta: 3, appliedDelta: 3 },
      }),
      ev({
        eventType: "vp_manual_adjust",
        round: 1,
        data: { category: "secondary", delta: 2, appliedDelta: 2 },
      }),
    ];
    const stats = buildPlayerStats(events, 1, 0);
    expect(stats.vpByRound[1].primary).toBe(3);
    expect(stats.vpByRound[1].secondary).toBe(2);
    expect(stats.totalVP).toBe(5);
  });

  it("ignores gambit events (gambit is hidden in the frontend)", () => {
    const events: NormalizedEvent[] = [
      ev({
        eventType: "vp_gambit_score",
        round: 1,
        data: { category: "gambit", delta: 4, appliedDelta: 4 },
      }),
      ev({
        eventType: "challenger_scored",
        round: 1,
        data: { vpScored: 6 },
      }),
      ev({
        eventType: "vp_manual_adjust",
        round: 1,
        data: { category: "gambit", delta: 2, appliedDelta: 2 },
      }),
    ];
    const stats = buildPlayerStats(events, 1, 0);
    expect(stats.vpByRound[1]).toBeUndefined();
    expect(stats.totalVP).toBe(0);
  });

  it("ignores events for other players", () => {
    const events: NormalizedEvent[] = [
      ev({
        eventType: "vp_primary_score",
        playerNumber: 2,
        round: 1,
        data: { category: "primary", delta: 5, appliedDelta: 5 },
      }),
    ];
    const stats = buildPlayerStats(events, 1, 0);
    expect(stats.totalVP).toBe(0);
  });

  it("falls back to delta when appliedDelta is absent (legacy events)", () => {
    const events: NormalizedEvent[] = [
      ev({
        eventType: "vp_secondary_score",
        round: 1,
        data: { category: "secondary", delta: 4 },
      }),
    ];
    const stats = buildPlayerStats(events, 1, 0);
    expect(stats.vpByRound[1].secondary).toBe(4);
  });
});
