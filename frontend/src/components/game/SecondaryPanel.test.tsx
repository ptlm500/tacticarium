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
  currentPhase?: string;
  isMyTurn?: boolean;
  secondaryHand?: SecondaryCard[];
  secondaryScored?: SecondaryCard[];
  secondaryDeck?: SecondaryCard[];
  onMove?: (
    secondaryId: string,
    fromPile: "deck" | "active" | "achieved" | "discarded",
    toPile: "deck" | "active" | "achieved" | "discarded",
    vpScored?: number,
  ) => void;
}) {
  return render(
    <SecondaryPanel
      mode="tactical"
      secondaryHand={overrides?.secondaryHand ?? [mockActiveSecondary]}
      secondaryScored={overrides?.secondaryScored ?? []}
      secondaryDeck={overrides?.secondaryDeck ?? makeDeck(5)}
      currentPhase={overrides?.currentPhase ?? "command"}
      isMyTurn={overrides?.isMyTurn ?? true}
      onDiscard={noop}
      onDraw={noop}
      onMove={overrides?.onMove ?? noop}
    />,
  );
}

describe("SecondaryPanel", () => {
  describe("free Discard button", () => {
    it("remains visible across phases as an escape hatch", () => {
      for (const phase of ["command", "movement", "shooting", "charge", "fight"]) {
        const { unmount } = renderPanel({ currentPhase: phase });
        expect(screen.getByText("Discard")).toBeTruthy();
        unmount();
      }
    });
  });

  describe("Draw Secondaries button", () => {
    function renderDraw(isMyTurn: boolean, currentPhase = "command") {
      render(
        <SecondaryPanel
          mode="tactical"
          secondaryHand={[]}
          secondaryScored={[]}
          secondaryDeck={makeDeck(5)}
          currentPhase={currentPhase}
          isMyTurn={isMyTurn}
          onDiscard={noop}
          onDraw={noop}
          onMove={noop}
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

    it("is disabled outside the command phase", () => {
      expect(renderDraw(true, "fight").disabled).toBe(true);
    });
  });

  describe("details modal", () => {
    it("opens when an active secondary is clicked", async () => {
      const user = userEvent.setup();
      const card: SecondaryCard = {
        ...mockActiveSecondary,
        id: "active-detail",
        name: "Behind Enemy Lines",
        text: "Score VP for units in enemy deployment zone.",
      };
      renderPanel({ secondaryHand: [card] });

      // Card text initially appears only inside the card.
      expect(screen.getAllByText("Score VP for units in enemy deployment zone.")).toHaveLength(1);

      await user.click(screen.getByRole("button", { name: /Behind Enemy Lines/ }));

      // Once the modal opens, the text renders in both the card and the dialog.
      expect(screen.getByRole("dialog")).toBeTruthy();
      expect(
        screen.getAllByText("Score VP for units in enemy deployment zone.").length,
      ).toBeGreaterThan(1);
    });

    it("opens when a scored secondary is clicked", async () => {
      const user = userEvent.setup();
      const scored: SecondaryCard = {
        ...mockActiveSecondary,
        id: "ach-1",
        name: "Cleared the field",
        text: "All enemies destroyed.",
      };
      renderPanel({
        secondaryHand: [],
        secondaryScored: [scored],
      });

      await user.click(screen.getByRole("button", { name: /Cleared the field/ }));

      expect(screen.getByRole("dialog")).toBeTruthy();
      expect(screen.getByText("All enemies destroyed.")).toBeTruthy();
    });
  });

  describe("Manage manually toggle", () => {
    it("hides normal controls when toggled on", async () => {
      const user = userEvent.setup();
      renderPanel({ currentPhase: "command", isMyTurn: true });
      expect(screen.queryByText("Discard")).toBeTruthy();

      await user.click(screen.getByRole("checkbox", { name: /Manage manually/i }));

      expect(screen.queryByText("Discard")).toBeNull();
      expect(screen.queryByRole("button", { name: /Draw Secondaries/ })).toBeNull();
    });

    it("exposes kanban move buttons for active cards", async () => {
      const user = userEvent.setup();
      renderPanel({ secondaryDeck: [] });
      await user.click(screen.getByRole("checkbox", { name: /Manage manually/i }));
      expect(screen.getByRole("button", { name: "→ Deck" })).toBeTruthy();
      expect(screen.getByRole("button", { name: "→ Discard" })).toBeTruthy();
    });

    it("renders deck and discarded piles individually with move-to-active buttons", async () => {
      const user = userEvent.setup();
      renderPanel({
        secondaryHand: [],
        secondaryDeck: makeDeck(2),
      });
      await user.click(screen.getByRole("checkbox", { name: /Manage manually/i }));

      expect(screen.getByText("Deck card 1")).toBeTruthy();
      expect(screen.getByText("Deck card 2")).toBeTruthy();
      // 2 deck cards each get a → Active button.
      expect(screen.getAllByRole("button", { name: "→ Active" }).length).toBe(2);
    });

    it("calls onMove with the correct pile names when buttons are clicked", async () => {
      const user = userEvent.setup();
      const calls: Array<[string, string, string, number | undefined]> = [];
      renderPanel({
        secondaryHand: [{ ...mockActiveSecondary, id: "a-1", name: "Active card" }],
        secondaryDeck: [],
        onMove: (id, from, to, vp) => calls.push([id, from, to, vp]),
      });
      await user.click(screen.getByRole("checkbox", { name: /Manage manually/i }));
      await user.click(screen.getByRole("button", { name: "→ Deck" }));
      await user.click(screen.getByRole("button", { name: "→ Discard" }));

      expect(calls).toEqual([
        ["a-1", "active", "deck", undefined],
        ["a-1", "active", "discarded", undefined],
      ]);
    });

    it("disables the → Active button when active pile is at capacity", async () => {
      const user = userEvent.setup();
      renderPanel({
        secondaryHand: [
          { ...mockActiveSecondary, id: "a-1", name: "A1" },
          { ...mockActiveSecondary, id: "a-2", name: "A2" },
        ],
        secondaryDeck: makeDeck(1),
      });
      await user.click(screen.getByRole("checkbox", { name: /Manage manually/i }));

      const moveButtons = screen.getAllByRole("button", {
        name: "→ Active",
      }) as HTMLButtonElement[];
      expect(moveButtons.length).toBeGreaterThan(0);
      for (const btn of moveButtons) {
        expect(btn.disabled).toBe(true);
      }
    });
  });
});
