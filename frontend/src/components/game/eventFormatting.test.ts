import { describe, expect, it } from "vite-plus/test";
import { formatEvent, type NormalizedEvent } from "./eventFormatting";

function primaryScore(data: Record<string, unknown>): NormalizedEvent {
  return { eventType: "vp_primary_score", playerNumber: 1, data };
}

describe("formatEvent — primary scoring", () => {
  it("renders the scoring rule label and slot when both are present", () => {
    const out = formatEvent(
      primaryScore({
        category: "primary",
        appliedDelta: 5,
        scoringSlot: "end_of_command_phase",
        scoringRuleLabel: "Hold the most",
      }),
    );
    expect(out).toBe("P1 scored 5 primary VP — Hold the most (end of command phase)");
  });

  it("falls back to the slot suffix when no rule label is present", () => {
    const out = formatEvent(
      primaryScore({
        category: "primary",
        appliedDelta: 5,
        scoringSlot: "end_of_command_phase",
      }),
    );
    expect(out).toBe("P1 scored 5 primary VP — end of command phase");
  });

  it("renders just the rule label when no slot is present", () => {
    const out = formatEvent(
      primaryScore({
        category: "primary",
        appliedDelta: 5,
        scoringRuleLabel: "Hold the most",
      }),
    );
    expect(out).toBe("P1 scored 5 primary VP — Hold the most");
  });
});

describe("formatEvent — secondary_moved", () => {
  it("formats a move with positive VP", () => {
    const out = formatEvent({
      eventType: "secondary_moved",
      playerNumber: 1,
      data: {
        secondaryName: "Behind Enemy Lines",
        fromPile: "active",
        toPile: "achieved",
        vpDelta: 4,
      },
    } as NormalizedEvent);
    expect(out).toBe("📝 P1 moved Behind Enemy Lines: active → achieved (+4 VP)");
  });

  it("formats a move with no VP change", () => {
    const out = formatEvent({
      eventType: "secondary_moved",
      playerNumber: 2,
      data: {
        secondaryName: "Engage on All Fronts",
        fromPile: "deck",
        toPile: "active",
        vpDelta: 0,
      },
    } as NormalizedEvent);
    expect(out).toBe("📝 P2 moved Engage on All Fronts: deck → active");
  });

  it("formats a move that revokes VP", () => {
    const out = formatEvent({
      eventType: "secondary_moved",
      playerNumber: 1,
      data: {
        secondaryName: "Sabotage",
        fromPile: "achieved",
        toPile: "active",
        vpDelta: -3,
      },
    } as NormalizedEvent);
    expect(out).toBe("📝 P1 moved Sabotage: achieved → active (-3 VP)");
  });
});
