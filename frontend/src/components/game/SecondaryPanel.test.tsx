import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { SecondaryPanel } from "./SecondaryPanel";
import { mockActiveSecondary } from "../../test/fixtures";
import { SecondaryCard } from "../../types/game";

const noop = () => {};

function makeDeck(count: number): SecondaryCard[] {
  return Array.from({ length: count }, (_, i) => ({
    ...mockActiveSecondary,
    id: `deck-${i + 1}`,
    name: `Deck card ${i + 1}`,
  }));
}

function renderPanel(overrides?: {
  mode?: string;
  currentPhase?: string;
  isMyTurn?: boolean;
  secondaryHand?: SecondaryCard[];
  secondaryScored?: SecondaryCard[];
  secondaryDeck?: SecondaryCard[];
  onDraw?: () => void;
}) {
  return render(
    <SecondaryPanel
      mode={overrides?.mode ?? "tactical"}
      secondaryHand={overrides?.secondaryHand ?? [mockActiveSecondary]}
      secondaryScored={overrides?.secondaryScored ?? []}
      secondaryDeck={overrides?.secondaryDeck ?? makeDeck(5)}
      currentPhase={overrides?.currentPhase ?? "command"}
      isMyTurn={overrides?.isMyTurn ?? true}
      onDraw={overrides?.onDraw ?? noop}
    />,
  );
}

describe("SecondaryPanel", () => {
  it("renders nothing when no mode is set", () => {
    const { container } = render(
      <SecondaryPanel
        mode=""
        secondaryHand={[]}
        secondaryScored={[]}
        secondaryDeck={[]}
        currentPhase="command"
        isMyTurn
        onDraw={noop}
      />,
    );
    expect(container.firstChild).toBeNull();
  });

  it("renders hand cards", () => {
    renderPanel({ secondaryHand: [{ ...mockActiveSecondary, name: "Behind Enemy Lines" }] });
    expect(screen.getByText("Behind Enemy Lines")).toBeTruthy();
  });

  describe("Draw Secondaries button", () => {
    function renderDraw(isMyTurn: boolean, currentPhase = "command", onDraw = noop) {
      renderPanel({ secondaryHand: [], isMyTurn, currentPhase, onDraw });
      return screen.getByRole("button", { name: /Draw Secondaries/ }) as HTMLButtonElement;
    }

    it("is enabled on the active player's turn in the command phase", () => {
      expect(renderDraw(true).disabled).toBe(false);
    });

    it("is disabled on the non-active player's turn", () => {
      expect(renderDraw(false).disabled).toBe(true);
    });

    it("is disabled outside the command phase", () => {
      expect(renderDraw(true, "fight").disabled).toBe(true);
    });

    it("calls onDraw when clicked", async () => {
      const user = userEvent.setup();
      const onDraw = vi.fn();
      const btn = renderDraw(true, "command", onDraw);
      await user.click(btn);
      expect(onDraw).toHaveBeenCalledOnce();
    });

    it("is not shown in fixed mode", () => {
      renderPanel({ mode: "fixed", secondaryHand: [] });
      expect(screen.queryByRole("button", { name: /Draw Secondaries/ })).toBeNull();
    });
  });

  describe("scored pile", () => {
    it("renders scored cards with their VP", () => {
      renderPanel({
        secondaryHand: [],
        secondaryScored: [
          { ...mockActiveSecondary, id: "ach-1", name: "Cleared the field", vpScored: 4 },
        ],
      });
      expect(screen.getByText("Cleared the field")).toBeTruthy();
      expect(screen.getByText("+4")).toBeTruthy();
    });
  });

  describe("details modal", () => {
    it("opens when a hand card is clicked", async () => {
      const user = userEvent.setup();
      const card: SecondaryCard = {
        ...mockActiveSecondary,
        id: "active-detail",
        name: "Behind Enemy Lines",
        text: "Score VP for units in enemy deployment zone.",
      };
      renderPanel({ secondaryHand: [card] });

      await user.click(screen.getByRole("button", { name: /Behind Enemy Lines/ }));

      expect(screen.getByRole("dialog")).toBeTruthy();
      expect(
        screen.getAllByText("Score VP for units in enemy deployment zone.").length,
      ).toBeGreaterThan(1);
    });
  });
});
