import { describe, it, expect } from "vite-plus/test";
import { buildPlayerStats, type RestGameEvent } from "./vpUtils";

function ev(partial: Partial<RestGameEvent>): RestGameEvent {
  return {
    id: 0,
    playerNumber: 1,
    eventType: "",
    eventData: null,
    round: 1,
    phase: "command",
    createdAt: "",
    ...partial,
  };
}

describe("buildPlayerStats", () => {
  it("uses appliedDelta when present (handles clamp divergence)", () => {
    const events: RestGameEvent[] = [
      ev({
        eventType: "vp_primary_score",
        round: 1,
        eventData: { category: "primary", delta: 10, appliedDelta: 2 },
      }),
    ];
    const stats = buildPlayerStats(events, 1, 0);
    expect(stats.vpByRound[1].primary).toBe(2);
    expect(stats.totalVP).toBe(2);
  });

  it("subtracts reverted primary scores", () => {
    const events: RestGameEvent[] = [
      ev({
        eventType: "vp_primary_score",
        round: 1,
        eventData: { category: "primary", delta: 5, appliedDelta: 5 },
      }),
      ev({
        eventType: "vp_primary_score_reverted",
        round: 2,
        eventData: { revertedRound: 1, scoringSlot: "end_of_command_phase", revertedDelta: 5 },
      }),
    ];
    const stats = buildPlayerStats(events, 1, 0);
    expect(stats.vpByRound[1].primary).toBe(0);
    expect(stats.totalVP).toBe(0);
  });

  it("includes vp_manual_adjust deltas", () => {
    const events: RestGameEvent[] = [
      ev({
        eventType: "vp_manual_adjust",
        round: 1,
        eventData: { category: "primary", delta: 3, appliedDelta: 3 },
      }),
      ev({
        eventType: "vp_manual_adjust",
        round: 1,
        eventData: { category: "gambit", delta: -2, appliedDelta: -2 },
      }),
    ];
    const stats = buildPlayerStats(events, 1, 0);
    expect(stats.vpByRound[1].primary).toBe(3);
    expect(stats.vpByRound[1].gambit).toBe(-2);
    expect(stats.totalVP).toBe(1);
  });

  it("ignores events for other players", () => {
    const events: RestGameEvent[] = [
      ev({
        eventType: "vp_primary_score",
        playerNumber: 2,
        round: 1,
        eventData: { category: "primary", delta: 5, appliedDelta: 5 },
      }),
    ];
    const stats = buildPlayerStats(events, 1, 0);
    expect(stats.totalVP).toBe(0);
  });

  it("falls back to delta when appliedDelta is absent (legacy events)", () => {
    const events: RestGameEvent[] = [
      ev({
        eventType: "vp_secondary_score",
        round: 1,
        eventData: { category: "secondary", delta: 4 },
      }),
    ];
    const stats = buildPlayerStats(events, 1, 0);
    expect(stats.vpByRound[1].secondary).toBe(4);
  });
});
