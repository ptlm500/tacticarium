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
