import { screen, act } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { SpectatorPlayerPanel } from "./SpectatorPlayerPanel";
import { renderWithProviders } from "../../test/renderWithProviders";
import { makePlayerState, mockActiveSecondary, mockStratagems } from "../../test/fixtures";
import { worker } from "../../mocks/browser";
import type { SecondaryCard, PlayerState } from "../../types/game";

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

  it("shows the detachment names when present", () => {
    renderPanel({
      detachments: [
        { id: "det-gladius", name: "Gladius Task Force", points: 3 },
        { id: "det-allied", name: "Allied Cohort", points: 0 },
      ],
    });
    expect(screen.getByText("Gladius Task Force, Allied Cohort")).toBeTruthy();
  });

  it("renders CP and the per-category VP breakdown", () => {
    renderPanel({ cp: 4, vpPrimary: 6, vpSecondary: 3, vpPaint: 1 });
    expect(screen.getByText("CP").nextSibling?.textContent).toBe("4");
    expect(screen.getByText("Primary").nextSibling?.textContent).toBe("6");
    expect(screen.getByText("Secondary").nextSibling?.textContent).toBe("3");
    expect(screen.getByText("Paint").nextSibling?.textContent).toBe("1");
  });

  it("computes total VP as primary + secondary + paint", () => {
    renderPanel({ vpPrimary: 6, vpSecondary: 3, vpPaint: 1 });
    expect(screen.getByText("Total VP").nextSibling?.textContent).toBe("10");
  });

  it("shows tactical mode and the remaining deck count", () => {
    const deck: SecondaryCard[] = [
      { ...mockActiveSecondary, id: "d1" },
      { ...mockActiveSecondary, id: "d2" },
      { ...mockActiveSecondary, id: "d3" },
    ];
    renderPanel({ secondaryMode: "tactical", secondaryDeck: deck });
    expect(screen.getByText(/Secondaries \(Tactical\)/)).toBeTruthy();
    expect(screen.getByText(/Deck: 3/)).toBeTruthy();
  });

  it("shows fixed mode without a deck count", () => {
    renderPanel({ secondaryMode: "fixed", secondaryDeck: [] });
    expect(screen.getByText(/Secondaries \(Fixed\)/)).toBeTruthy();
    expect(screen.queryByText(/Deck:/)).toBeNull();
  });

  it("shows the empty placeholder when there are no active secondaries", () => {
    renderPanel({ secondaryHand: [] });
    expect(screen.getByText("No active secondaries")).toBeTruthy();
  });

  it("renders active secondaries with name and text", () => {
    renderPanel({ secondaryHand: [mockActiveSecondary] });
    expect(screen.getByText("Behind Enemy Lines")).toBeTruthy();
    expect(screen.getByText("Score VP for units in enemy deployment zone")).toBeTruthy();
  });

  it("renders scored secondaries with VP scored", () => {
    renderPanel({
      secondaryScored: [{ ...mockActiveSecondary, id: "a1", vpScored: 4 }],
    });
    expect(screen.getByText("Scored (1)")).toBeTruthy();
    expect(screen.getByText("Behind Enemy Lines")).toBeTruthy();
    expect(screen.getByText("+4")).toBeTruthy();
  });

  it("does not render a +VP marker for scored secondaries scored at zero", () => {
    renderPanel({
      secondaryScored: [{ ...mockActiveSecondary, id: "a1", vpScored: 0 }],
    });
    expect(screen.getByText("Scored (1)")).toBeTruthy();
    expect(screen.queryByText(/^\+/)).toBeNull();
  });

  it("hides the scored section when empty", () => {
    renderPanel({ secondaryScored: [] });
    expect(screen.queryByText(/^Scored \(/)).toBeNull();
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
});
