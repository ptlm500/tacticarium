import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ScoringPrompt, ScoringPromptItem } from "./ScoringPrompt";
import { mockActiveSecondary, mockFixedSecondary } from "../../test/fixtures";

function renderPrompt(overrides: Partial<Parameters<typeof ScoringPrompt>[0]> = {}) {
  const defaultProps = {
    items: [] as ScoringPromptItem[],
    onScore: vi.fn(),
    activeSecondaries: [],
    onAchieveSecondary: vi.fn(),
    onDiscardSecondary: vi.fn(),
    canGainCP: true,
    onScoreFixedVP: vi.fn(),
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
    it("renders active secondaries with achieve/discard buttons", async () => {
      const user = userEvent.setup();
      const items: ScoringPromptItem[] = [{ kind: "secondary" }];
      const { props } = renderPrompt({
        items,
        activeSecondaries: [mockActiveSecondary],
      });

      expect(screen.getByText("Behind Enemy Lines")).toBeTruthy();

      const scoreBtn = screen.getByText(/1 unit \+2/);
      await user.click(scoreBtn);
      expect(props.onAchieveSecondary).toHaveBeenCalledWith("sec-1", 2);

      await user.click(screen.getByText("Discard"));
      expect(props.onDiscardSecondary).toHaveBeenCalledWith("sec-1", true);
    });

    it("shows +1CP button when canGainCP is true", () => {
      const items: ScoringPromptItem[] = [{ kind: "secondary" }];
      renderPrompt({
        items,
        activeSecondaries: [mockActiveSecondary],
        canGainCP: true,
      });
      expect(screen.getByText("+1CP")).toBeTruthy();
    });

    it("hides +1CP button when canGainCP is false", () => {
      const items: ScoringPromptItem[] = [{ kind: "secondary" }];
      renderPrompt({
        items,
        activeSecondaries: [mockActiveSecondary],
        canGainCP: false,
      });
      expect(screen.queryByText("+1CP")).toBeNull();
    });

    it("shows empty message when no active secondaries", () => {
      const items: ScoringPromptItem[] = [{ kind: "secondary" }];
      renderPrompt({ items, activeSecondaries: [] });
      expect(screen.getByText("No active secondary missions.")).toBeTruthy();
    });
  });

  describe("FixedSecondaryReminder", () => {
    it("renders fixed secondaries with scoring options", async () => {
      const user = userEvent.setup();
      const items: ScoringPromptItem[] = [
        { kind: "fixed_secondary", secondaries: [mockFixedSecondary] },
      ];
      const { props } = renderPrompt({ items });

      expect(screen.getByText("Assassination")).toBeTruthy();
      expect(screen.getByText(/max 8 VP/)).toBeTruthy();

      await user.click(screen.getByText(/Character.*\+3VP/));
      expect(props.onScoreFixedVP).toHaveBeenCalledWith(3);
    });
  });

  describe("timing filtering", () => {
    const ownTurnSecondary = {
      ...mockActiveSecondary,
      id: "own-1",
      name: "Own Turn Card",
    };
    const opponentTurnSecondary = {
      ...mockActiveSecondary,
      id: "opp-1",
      name: "Sabotage",
      scoringTiming: "end_of_opponent_turn" as const,
    };

    it("default own-turn item hides end_of_opponent_turn secondaries", () => {
      const items: ScoringPromptItem[] = [{ kind: "secondary", timing: "end_of_own_turn" }];
      renderPrompt({
        items,
        activeSecondaries: [ownTurnSecondary, opponentTurnSecondary],
      });

      expect(screen.getByText("Own Turn Card")).toBeTruthy();
      expect(screen.queryByText("Sabotage")).toBeNull();
    });

    it("opponent-turn item shows only end_of_opponent_turn secondaries", () => {
      const items: ScoringPromptItem[] = [{ kind: "secondary", timing: "end_of_opponent_turn" }];
      renderPrompt({
        items,
        activeSecondaries: [ownTurnSecondary, opponentTurnSecondary],
      });

      expect(screen.getByText("Sabotage")).toBeTruthy();
      expect(screen.queryByText("Own Turn Card")).toBeNull();
    });

    it("untagged secondaries default to end_of_own_turn", () => {
      const items: ScoringPromptItem[] = [{ kind: "secondary", timing: "end_of_own_turn" }];
      // mockActiveSecondary has no scoringTiming field — should appear in own-turn prompt.
      renderPrompt({ items, activeSecondaries: [mockActiveSecondary] });
      expect(screen.getByText(mockActiveSecondary.name!)).toBeTruthy();
    });

    it("renders custom title and labels when overridden", () => {
      const items: ScoringPromptItem[] = [{ kind: "secondary", timing: "end_of_opponent_turn" }];
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
