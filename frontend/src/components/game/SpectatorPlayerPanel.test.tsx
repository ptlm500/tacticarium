import { screen, act } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { SpectatorPlayerPanel } from "./SpectatorPlayerPanel";
import { renderWithProviders } from "../../test/renderWithProviders";
import { makePlayerState, mockActiveSecondary, mockStratagems } from "../../test/fixtures";
import { worker } from "../../mocks/browser";
import type { ActiveSecondary, PlayerState } from "../../types/game";

function renderPanel(
  player: Partial<PlayerState> = {},
  isActive = false,
  options: { stratagemsByFaction?: Record<string, typeof mockStratagems> } = {},
) {
  const stratagemsByFaction = options.stratagemsByFaction;
  if (stratagemsByFaction) {
    worker.use(
      http.get("http://localhost:8080/api/factions/:factionId/stratagems", ({ params }) => {
        const factionId = params.factionId as string;
        return HttpResponse.json(stratagemsByFaction[factionId] ?? []);
      }),
    );
  }
  return renderWithProviders(
    <SpectatorPlayerPanel player={makePlayerState(player)} isActive={isActive} />,
  );
}

describe("SpectatorPlayerPanel", () => {
  it("shows the player name and faction in the panel label", () => {
    renderPanel({ username: "Alice", factionName: "Space Marines" });
    expect(screen.getByText("Alice — Space Marines")).toBeTruthy();
  });

  it("falls back to 'Unknown faction' when factionName is missing", () => {
    renderPanel({ username: "Alice", factionName: "" });
    expect(screen.getByText("Alice — Unknown faction")).toBeTruthy();
  });

  it("shows the Active Turn badge when isActive is true", () => {
    renderPanel({}, true);
    expect(screen.getByText("Active Turn")).toBeTruthy();
  });

  it("hides the Active Turn badge when isActive is false", () => {
    renderPanel({}, false);
    expect(screen.queryByText("Active Turn")).toBeNull();
  });

  it("shows the detachment name when present", () => {
    renderPanel({ detachmentName: "Gladius Task Force" });
    expect(screen.getByText("Gladius Task Force")).toBeTruthy();
  });

  it("shows the Challenger badge when the player is the challenger", () => {
    renderPanel({ isChallenger: true });
    expect(screen.getByText("Challenger")).toBeTruthy();
  });

  it("hides the Challenger badge when the player is not the challenger", () => {
    renderPanel({ isChallenger: false });
    expect(screen.queryByText("Challenger")).toBeNull();
  });

  it("renders CP and the per-category VP breakdown", () => {
    renderPanel({ cp: 4, vpPrimary: 6, vpSecondary: 3, vpGambit: 2, vpPaint: 1 });
    expect(screen.getByText("CP").nextSibling?.textContent).toBe("4");
    expect(screen.getByText("Primary").nextSibling?.textContent).toBe("6");
    expect(screen.getByText("Secondary").nextSibling?.textContent).toBe("3");
    expect(screen.getByText("Gambit").nextSibling?.textContent).toBe("2");
    expect(screen.getByText("Paint").nextSibling?.textContent).toBe("1");
  });

  it("computes total VP as the sum of all VP categories", () => {
    renderPanel({ vpPrimary: 6, vpSecondary: 3, vpGambit: 2, vpPaint: 1 });
    expect(screen.getByText("Total VP").nextSibling?.textContent).toBe("12");
  });

  it("shows tactical mode and the remaining deck count", () => {
    const deck: ActiveSecondary[] = [
      { ...mockActiveSecondary, id: "d1" },
      { ...mockActiveSecondary, id: "d2" },
      { ...mockActiveSecondary, id: "d3" },
    ];
    renderPanel({ secondaryMode: "tactical", tacticalDeck: deck });
    expect(screen.getByText(/Secondaries \(Tactical\)/)).toBeTruthy();
    expect(screen.getByText(/Deck: 3/)).toBeTruthy();
  });

  it("shows fixed mode without a deck count", () => {
    renderPanel({ secondaryMode: "fixed", tacticalDeck: [] });
    expect(screen.getByText(/Secondaries \(Fixed\)/)).toBeTruthy();
    expect(screen.queryByText(/Deck:/)).toBeNull();
  });

  it("shows the empty placeholder when there are no active secondaries", () => {
    renderPanel({ activeSecondaries: [] });
    expect(screen.getByText("No active secondaries")).toBeTruthy();
  });

  it("renders active secondaries with name, description, and max VP", () => {
    renderPanel({ activeSecondaries: [mockActiveSecondary] });
    expect(screen.getByText("Behind Enemy Lines")).toBeTruthy();
    expect(screen.getByText("Score VP for units in enemy deployment zone")).toBeTruthy();
    expect(screen.getByText("5 VP max")).toBeTruthy();
  });

  it("renders achieved secondaries with VP scored", () => {
    renderPanel({
      achievedSecondaries: [{ ...mockActiveSecondary, id: "a1", vpScored: 4 }],
    });
    expect(screen.getByText("Achieved (1)")).toBeTruthy();
    expect(screen.getByText("Behind Enemy Lines")).toBeTruthy();
    expect(screen.getByText("+4")).toBeTruthy();
  });

  it("does not render a +VP marker for achieved secondaries scored at zero", () => {
    renderPanel({
      achievedSecondaries: [{ ...mockActiveSecondary, id: "a1", vpScored: 0 }],
    });
    expect(screen.getByText("Achieved (1)")).toBeTruthy();
    expect(screen.queryByText(/^\+/)).toBeNull();
  });

  it("renders discarded secondaries", () => {
    renderPanel({
      discardedSecondaries: [{ ...mockActiveSecondary, id: "d1", name: "Skirmish" }],
    });
    expect(screen.getByText("Discarded (1)")).toBeTruthy();
    expect(screen.getByText("Skirmish")).toBeTruthy();
  });

  it("hides the achieved and discarded sections when both are empty", () => {
    renderPanel({ achievedSecondaries: [], discardedSecondaries: [] });
    expect(screen.queryByText(/^Achieved \(/)).toBeNull();
    expect(screen.queryByText(/^Discarded \(/)).toBeNull();
  });

  it("hides the stratagems section when none have been used this phase", () => {
    renderPanel({ stratagemsUsedThisPhase: [] });
    expect(screen.queryByText("Stratagems This Phase")).toBeNull();
  });

  it("renders stratagem names (not IDs) for stratagems used this phase", async () => {
    await act(async () => {
      renderPanel({ stratagemsUsedThisPhase: ["strat-1", "strat-3"] }, false, {
        stratagemsByFaction: { "faction-sm": mockStratagems },
      });
    });

    await vi.waitFor(() => {
      expect(screen.getByText("Stratagems This Phase")).toBeTruthy();
      expect(screen.getByText("Command Re-roll")).toBeTruthy();
      expect(screen.getByText("Heroic Intervention")).toBeTruthy();
    });
    // Raw IDs must not leak through to the UI.
    expect(screen.queryByText("strat-1")).toBeNull();
    expect(screen.queryByText("strat-3")).toBeNull();
  });

  it("falls back to the stratagem ID when the catalog has no matching name", async () => {
    await act(async () => {
      renderPanel({ stratagemsUsedThisPhase: ["strat-1", "strat-unknown"] }, false, {
        stratagemsByFaction: { "faction-sm": mockStratagems },
      });
    });

    await vi.waitFor(() => {
      expect(screen.getByText("Command Re-roll")).toBeTruthy();
    });
    expect(screen.getByText("strat-unknown")).toBeTruthy();
  });

  it("uses the ID while the stratagem catalog is still loading", () => {
    // No worker.use() override — default REST handler returns [] synchronously,
    // but the data still arrives asynchronously. Without awaiting, the lookup
    // map is empty, so the panel must render the raw ID rather than crash.
    renderPanel({ stratagemsUsedThisPhase: ["strat-1"] });
    expect(screen.getByText("Stratagems This Phase")).toBeTruthy();
    expect(screen.getByText("strat-1")).toBeTruthy();
  });

  it("shows the Adapt or Die counter when uses are greater than zero", () => {
    renderPanel({ adaptOrDieUses: 2 });
    expect(screen.getByText(/Adapt or Die uses:/)).toBeTruthy();
    expect(screen.getByText("2")).toBeTruthy();
  });

  it("hides the Adapt or Die counter when uses are zero", () => {
    renderPanel({ adaptOrDieUses: 0 });
    expect(screen.queryByText(/Adapt or Die uses:/)).toBeNull();
  });
});
