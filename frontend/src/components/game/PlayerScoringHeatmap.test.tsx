import { describe, it, expect } from "vite-plus/test";
import { screen, fireEvent, act } from "@testing-library/react";
import { renderWithProviders } from "../../test/renderWithProviders";
import { PlayerScoringHeatmap } from "./PlayerScoringHeatmap";
import { buildPlayerStats, computeIntensityMax } from "./vpUtils";
import type { NormalizedEvent } from "./eventFormatting";
import type { VPCategory } from "./vpUtils";

function ev(partial: Partial<NormalizedEvent>): NormalizedEvent {
  return {
    playerNumber: 1,
    eventType: "",
    round: 1,
    phase: "command",
    ...partial,
  };
}

const baseEvents: NormalizedEvent[] = [
  ev({
    eventType: "vp_primary_score",
    round: 1,
    data: {
      category: "primary",
      delta: 5,
      appliedDelta: 5,
      scoringSlot: "end_of_battle_round",
      scoringRuleLabel: "Hold the most",
    },
  }),
  ev({
    eventType: "vp_secondary_score",
    round: 1,
    data: { category: "secondary", delta: 3, appliedDelta: 3, secondaryName: "Behind Lines" },
  }),
  ev({
    eventType: "vp_primary_score",
    round: 2,
    data: {
      category: "primary",
      delta: 8,
      appliedDelta: 8,
      scoringSlot: "end_of_command_phase",
      scoringRuleLabel: "Hold 2+",
    },
  }),
  ev({
    eventType: "vp_gambit_score",
    round: 1,
    data: { category: "gambit", delta: 12, appliedDelta: 12 },
  }),
];

function renderHeatmap(
  events: NormalizedEvent[] = baseEvents,
  rounds: number[] = [1, 2],
  onCellClick: (round: number, category: VPCategory) => void = () => {},
) {
  const stats = buildPlayerStats(events, 1, 0);
  return renderWithProviders(
    <PlayerScoringHeatmap
      username="Alice"
      stats={stats}
      rounds={rounds}
      intensityMax={computeIntensityMax(stats)}
      onCellClick={onCellClick}
    />,
  );
}

describe("PlayerScoringHeatmap", () => {
  it("renders the player's primary and secondary VP for each round", () => {
    renderHeatmap();
    expect(screen.getByLabelText("Pri round 1: 5 VP, click for details")).toBeTruthy();
    expect(screen.getByLabelText("Sec round 1: 3 VP, click for details")).toBeTruthy();
    expect(screen.getByLabelText("Pri round 2: 8 VP, click for details")).toBeTruthy();
    expect(screen.getByLabelText("Sec round 2: 0 VP")).toBeTruthy();
  });

  it("excludes gambit events from the heatmap totals", () => {
    renderHeatmap();
    expect(screen.queryByLabelText(/12 VP/)).toBeNull();
  });

  it("disables empty cells (no scoring detail to open)", () => {
    renderHeatmap();
    const emptyCell = screen.getByLabelText("Sec round 2: 0 VP");
    expect((emptyCell as HTMLButtonElement).disabled).toBe(true);
  });

  it("invokes onCellClick with the round and category when a non-zero cell is clicked", async () => {
    const calls: Array<[number, VPCategory]> = [];
    renderHeatmap(baseEvents, [1, 2], (round, category) => calls.push([round, category]));
    await act(async () => {
      fireEvent.click(screen.getByLabelText("Pri round 1: 5 VP, click for details"));
    });
    expect(calls).toEqual([[1, "primary"]]);
  });
});
