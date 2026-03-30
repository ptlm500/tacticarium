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
    onDrawSecondary: vi.fn(),
    canGainCP: true,
    deckSize: 5,
    activeSecondaryCount: 1,
    onScoreFixedVP: vi.fn(),
    onConfirm: vi.fn(),
    onCancel: vi.fn(),
    ...overrides,
  };
  return { ...render(<ScoringPrompt {...defaultProps} />), props: defaultProps };
}

describe("ScoringPrompt", () => {
  it("renders scoring reminder header", () => {
    renderPrompt();
    expect(screen.getByText("Scoring Reminder")).toBeTruthy();
  });

  it("renders confirm and cancel buttons", async () => {
    const user = userEvent.setup();
    const { props } = renderPrompt();

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
        },
      ];
      const { props } = renderPrompt({ items });

      expect(screen.getByText("Score Primary — Supply Drop")).toBeTruthy();

      await user.click(screen.getByText("3+ objectives (+10)"));
      expect(props.onScore).toHaveBeenCalledWith("primary", 10);
    });

    it("disables scoring buttons locked by minRound", () => {
      const items: ScoringPromptItem[] = [
        {
          kind: "primary",
          missionName: "Supply Drop",
          scoringRules: [{ label: "Late bonus", vp: 5, minRound: 3 }],
          currentRound: 1,
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

  describe("TacticalDrawReminder", () => {
    it("shows draw button when eligible", async () => {
      const user = userEvent.setup();
      const items: ScoringPromptItem[] = [{ kind: "tactical_draw" }];
      const { props } = renderPrompt({
        items,
        activeSecondaryCount: 1,
        deckSize: 5,
      });

      const drawBtn = screen.getByText(/Draw Secondaries/);
      await user.click(drawBtn);
      expect(props.onDrawSecondary).toHaveBeenCalledOnce();
    });

    it('shows "already have 2" when at max', () => {
      const items: ScoringPromptItem[] = [{ kind: "tactical_draw" }];
      renderPrompt({ items, activeSecondaryCount: 2, deckSize: 5 });
      expect(screen.getByText("You already have 2 active secondaries.")).toBeTruthy();
    });

    it('shows "deck empty" when no cards left', () => {
      const items: ScoringPromptItem[] = [{ kind: "tactical_draw" }];
      renderPrompt({ items, activeSecondaryCount: 0, deckSize: 0 });
      expect(screen.getByText("Deck is empty.")).toBeTruthy();
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
});
