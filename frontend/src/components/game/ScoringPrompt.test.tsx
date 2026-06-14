import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ScoringPrompt, ScoringPromptItem } from "./ScoringPrompt";
import { mockActiveSecondary } from "../../test/fixtures";

function renderPrompt(overrides: Partial<Parameters<typeof ScoringPrompt>[0]> = {}) {
  const defaultProps = {
    items: [] as ScoringPromptItem[],
    onScore: vi.fn(),
    secondaryHand: [],
    onConfirm: vi.fn(),
    onCancel: vi.fn(),
    ...overrides,
  };
  return { ...render(<ScoringPrompt {...defaultProps} />), props: defaultProps };
}

describe("ScoringPrompt", () => {
  it("renders Scoring Reminder title and scoring-specific buttons", () => {
    renderPrompt();
    expect(screen.getByText("Scoring Reminder")).toBeTruthy();
    expect(screen.getByText("I've scored, continue")).toBeTruthy();
    expect(screen.getByText("Let me score first")).toBeTruthy();
  });

  it("fires confirm and cancel callbacks", async () => {
    const user = userEvent.setup();
    const items: ScoringPromptItem[] = [{ kind: "secondary" }];
    const { props } = renderPrompt({ items });

    await user.click(screen.getByText("I've scored, continue"));
    expect(props.onConfirm).toHaveBeenCalledOnce();

    await user.click(screen.getByText("Let me score first"));
    expect(props.onCancel).toHaveBeenCalledOnce();
  });

  describe("PrimaryReminder", () => {
    it("renders primary scoring buttons", async () => {
      const user = userEvent.setup();
      const items: ScoringPromptItem[] = [
        {
          kind: "primary",
          missionName: "Supply Drop",
          scoringRules: [
            { label: "2 objectives", vp: 5, minRound: 2 },
            { label: "3+ objectives", vp: 10 },
          ],
          currentRound: 3,
          scoringSlot: "end_of_command_phase",
        },
      ];
      const { props } = renderPrompt({ items });

      expect(screen.getByText("Score Primary — Supply Drop")).toBeTruthy();

      await user.click(screen.getByText("3+ objectives (+10)"));
      expect(props.onScore).toHaveBeenCalledWith(
        "primary",
        10,
        "end_of_command_phase",
        "3+ objectives",
      );
    });

    it("disables scoring buttons locked by minRound", () => {
      const items: ScoringPromptItem[] = [
        {
          kind: "primary",
          missionName: "Supply Drop",
          scoringRules: [{ label: "Late bonus", vp: 5, minRound: 3 }],
          currentRound: 1,
          scoringSlot: "end_of_command_phase",
        },
      ];
      renderPrompt({ items });

      const btn = screen.getByText(/Late bonus/);
      expect(btn.closest("button")!.hasAttribute("disabled")).toBe(true);
    });
  });

  describe("SecondaryReminder", () => {
    it("renders active secondaries with name and prose text", () => {
      const items: ScoringPromptItem[] = [{ kind: "secondary" }];
      renderPrompt({
        items,
        secondaryHand: [mockActiveSecondary],
      });

      expect(screen.getByText("Behind Enemy Lines")).toBeTruthy();
      expect(screen.getByText(mockActiveSecondary.text)).toBeTruthy();
    });

    it("shows empty message when no active secondaries", () => {
      const items: ScoringPromptItem[] = [{ kind: "secondary" }];
      renderPrompt({ items, secondaryHand: [] });
      expect(screen.getByText("No active secondary missions.")).toBeTruthy();
    });
  });

  describe("custom labels", () => {
    it("renders custom title and labels when overridden", () => {
      const items: ScoringPromptItem[] = [{ kind: "secondary" }];
      renderPrompt({
        items,
        title: "Opponent's Turn Ended",
        description: "Score now.",
        confirmLabel: "Done",
        cancelLabel: "Dismiss",
      });
      expect(screen.getByText("Opponent's Turn Ended")).toBeTruthy();
      expect(screen.getByText("Done")).toBeTruthy();
      expect(screen.getByText("Dismiss")).toBeTruthy();
    });
  });

  describe("opponent_pending_secondary", () => {
    it("renders a read-only reminder listing the opponent's pending secondaries", () => {
      const items: ScoringPromptItem[] = [
        {
          kind: "opponent_pending_secondary",
          secondaries: [
            { ...mockActiveSecondary, id: "sabotage", name: "Sabotage" },
            { ...mockActiveSecondary, id: "defend", name: "Defend Stronghold" },
          ],
          opponentName: "Bob",
        },
      ];
      renderPrompt({ items });

      expect(screen.getByTestId("opponent-pending-secondary")).toBeTruthy();
      expect(screen.getByText(/Wait for Bob to score/)).toBeTruthy();
      expect(screen.getByText(/• Sabotage/)).toBeTruthy();
      expect(screen.getByText(/• Defend Stronghold/)).toBeTruthy();
    });
  });
});
