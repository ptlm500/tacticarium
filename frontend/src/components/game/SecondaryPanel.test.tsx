import { render, screen } from "@testing-library/react";
import { SecondaryPanel } from "./SecondaryPanel";
import { mockActiveSecondary } from "../../test/fixtures";

const noop = () => {};

function renderPanel(overrides?: {
  currentPhase?: string;
  isMyTurn?: boolean;
  currentRound?: number;
  canGainCP?: boolean;
  currentCP?: number;
}) {
  render(
    <SecondaryPanel
      mode="tactical"
      activeSecondaries={[mockActiveSecondary]}
      achievedSecondaries={[]}
      discardedSecondaries={[]}
      deckSize={5}
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
          deckSize={5}
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
            deckSize={5}
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
            onScoreFixedVP={noop}
          />,
        );
        expect(screen.getByText("Discard")).toBeTruthy();
        unmount();
      }
    });
  });
});
