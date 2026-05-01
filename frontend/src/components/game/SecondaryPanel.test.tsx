import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { SecondaryPanel } from "./SecondaryPanel";
import { mockActiveSecondary } from "../../test/fixtures";
import { ActiveSecondary } from "../../types/game";

const noop = () => {};

function makeDeck(count: number): ActiveSecondary[] {
  return Array.from({ length: count }, (_, i) => ({
    ...mockActiveSecondary,
    id: `deck-${i + 1}`,
    name: `Deck card ${i + 1}`,
  }));
}

function renderPanel(overrides?: {
  currentPhase?: string;
  isMyTurn?: boolean;
  currentRound?: number;
  canGainCP?: boolean;
  currentCP?: number;
  activeSecondaries?: ActiveSecondary[];
  achievedSecondaries?: ActiveSecondary[];
  discardedSecondaries?: ActiveSecondary[];
  tacticalDeck?: ActiveSecondary[];
  onMove?: (
    secondaryId: string,
    fromPile: "deck" | "active" | "achieved" | "discarded",
    toPile: "deck" | "active" | "achieved" | "discarded",
    vpScored?: number,
  ) => void;
}) {
  render(
    <SecondaryPanel
      mode="tactical"
      activeSecondaries={overrides?.activeSecondaries ?? [mockActiveSecondary]}
      achievedSecondaries={overrides?.achievedSecondaries ?? []}
      discardedSecondaries={overrides?.discardedSecondaries ?? []}
      tacticalDeck={overrides?.tacticalDeck ?? makeDeck(5)}
      currentRound={overrides?.currentRound ?? 2}
      currentPhase={overrides?.currentPhase ?? "command"}
      isMyTurn={overrides?.isMyTurn ?? true}
      currentCP={overrides?.currentCP ?? 3}
      canGainCP={overrides?.canGainCP ?? true}
      onAchieve={noop}
      onDiscard={noop}
      onNewOrders={noop}
      onReshuffle={noop}
      onDraw={noop}
      onMove={overrides?.onMove ?? noop}
      onScoreFixedVP={noop}
    />,
  );
}

describe("SecondaryPanel", () => {
  describe("New Orders button", () => {
    it("is shown during the command phase on the user's turn", () => {
      renderPanel({ currentPhase: "command", isMyTurn: true });
      expect(screen.getByText("New Orders")).toBeTruthy();
    });

    it("is hidden outside the command phase", () => {
      renderPanel({ currentPhase: "movement", isMyTurn: true });
      expect(screen.queryByText("New Orders")).toBeNull();
    });

    it("is hidden during the fight phase", () => {
      renderPanel({ currentPhase: "fight", isMyTurn: true });
      expect(screen.queryByText("New Orders")).toBeNull();
    });

    it("is hidden on the opponent's turn", () => {
      renderPanel({ currentPhase: "command", isMyTurn: false });
      expect(screen.queryByText("New Orders")).toBeNull();
    });
  });

  describe("Discard +1CP button", () => {
    it("is shown during the fight phase on the user's turn", () => {
      renderPanel({ currentPhase: "fight", isMyTurn: true });
      expect(screen.getByText("Discard +1CP")).toBeTruthy();
    });

    it("is hidden during the command phase", () => {
      renderPanel({ currentPhase: "command", isMyTurn: true });
      expect(screen.queryByText("Discard +1CP")).toBeNull();
      expect(screen.queryByText("Discard (CP capped)")).toBeNull();
    });

    it("is hidden during the movement phase", () => {
      renderPanel({ currentPhase: "movement", isMyTurn: true });
      expect(screen.queryByText("Discard +1CP")).toBeNull();
    });

    it("is hidden on the opponent's turn", () => {
      renderPanel({ currentPhase: "fight", isMyTurn: false });
      expect(screen.queryByText("Discard +1CP")).toBeNull();
    });

    it("is hidden in round 5 even during the fight phase", () => {
      renderPanel({ currentPhase: "fight", isMyTurn: true, currentRound: 5 });
      expect(screen.queryByText("Discard +1CP")).toBeNull();
    });
  });

  describe("Draw Secondaries button", () => {
    function renderDraw(isMyTurn: boolean) {
      render(
        <SecondaryPanel
          mode="tactical"
          activeSecondaries={[]}
          achievedSecondaries={[]}
          discardedSecondaries={[]}
          tacticalDeck={makeDeck(5)}
          currentRound={1}
          currentPhase="command"
          isMyTurn={isMyTurn}
          currentCP={3}
          canGainCP={true}
          onAchieve={noop}
          onDiscard={noop}
          onNewOrders={noop}
          onReshuffle={noop}
          onDraw={noop}
          onMove={noop}
          onScoreFixedVP={noop}
        />,
      );
      return screen.getByRole("button", { name: /Draw Secondaries/ }) as HTMLButtonElement;
    }

    it("is enabled on the active player's turn", () => {
      expect(renderDraw(true).disabled).toBe(false);
    });

    it("is disabled on the non-active player's turn", () => {
      expect(renderDraw(false).disabled).toBe(true);
    });
  });

  describe("free Discard button", () => {
    it("remains visible in all phases as an escape hatch", () => {
      for (const phase of ["command", "movement", "shooting", "charge", "fight"]) {
        const { unmount } = render(
          <SecondaryPanel
            mode="tactical"
            activeSecondaries={[mockActiveSecondary]}
            achievedSecondaries={[]}
            discardedSecondaries={[]}
            tacticalDeck={makeDeck(5)}
            currentRound={2}
            currentPhase={phase}
            isMyTurn={true}
            currentCP={3}
            canGainCP={true}
            onAchieve={noop}
            onDiscard={noop}
            onNewOrders={noop}
            onReshuffle={noop}
            onDraw={noop}
            onMove={noop}
            onScoreFixedVP={noop}
          />,
        );
        expect(screen.getByText("Discard")).toBeTruthy();
        unmount();
      }
    });
  });

  describe("Manage manually toggle", () => {
    it("hides normal controls when toggled on", async () => {
      const user = userEvent.setup();
      renderPanel({ currentPhase: "command", isMyTurn: true });
      expect(screen.queryByText("New Orders")).toBeTruthy();
      expect(screen.queryByText("Discard")).toBeTruthy();

      await user.click(screen.getByRole("checkbox", { name: /Manage manually/i }));

      expect(screen.queryByText("New Orders")).toBeNull();
      expect(screen.queryByText("Discard")).toBeNull();
      expect(screen.queryByRole("button", { name: /Draw Secondaries/ })).toBeNull();
    });

    it("exposes manual move buttons for active cards", async () => {
      const user = userEvent.setup();
      renderPanel();
      await user.click(screen.getByRole("checkbox", { name: /Manage manually/i }));
      expect(screen.getByText("Send to Deck")).toBeTruthy();
      expect(screen.getByText("Send to Discard")).toBeTruthy();
    });

    it("renders deck and discarded piles individually with move-to-active buttons", async () => {
      const user = userEvent.setup();
      const discarded = [{ ...mockActiveSecondary, id: "disc-1", name: "Discarded card" }];
      renderPanel({
        activeSecondaries: [],
        discardedSecondaries: discarded,
        tacticalDeck: makeDeck(2),
      });
      await user.click(screen.getByRole("checkbox", { name: /Manage manually/i }));

      expect(screen.getByText("Deck card 1")).toBeTruthy();
      expect(screen.getByText("Deck card 2")).toBeTruthy();
      expect(screen.getByText("Discarded card")).toBeTruthy();
      // 3 cards (2 deck + 1 discarded) each get a Move to Active button.
      expect(screen.getAllByText("Move to Active").length).toBe(3);
    });

    it("calls onMove with the correct pile names when buttons are clicked", async () => {
      const user = userEvent.setup();
      const calls: Array<[string, string, string, number | undefined]> = [];
      renderPanel({
        activeSecondaries: [{ ...mockActiveSecondary, id: "a-1", name: "Active card" }],
        onMove: (id, from, to, vp) => calls.push([id, from, to, vp]),
      });
      await user.click(screen.getByRole("checkbox", { name: /Manage manually/i }));
      await user.click(screen.getByText("Send to Deck"));
      await user.click(screen.getByText("Send to Discard"));

      expect(calls).toEqual([
        ["a-1", "active", "deck", undefined],
        ["a-1", "active", "discarded", undefined],
      ]);
    });
  });
});
