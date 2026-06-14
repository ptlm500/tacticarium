import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ScorePromptsPanel } from "./ScorePromptsPanel";
import type { ScorePrompt } from "../../types/game";

function makePrompt(overrides?: Partial<ScorePrompt>): ScorePrompt {
  return {
    id: "prompt-1",
    category: "secondary",
    cardId: "bring-it-down",
    cardName: "Bring It Down",
    awardIndex: 0,
    round: 2,
    text: "Score 2 VP per destroyed enemy unit with Starting Strength 13+.",
    ...overrides,
  };
}

describe("ScorePromptsPanel", () => {
  it("renders nothing when there are no prompts", () => {
    const { container } = render(<ScorePromptsPanel prompts={[]} onConfirm={() => {}} />);
    expect(container.firstChild).toBeNull();
  });

  it("renders each prompt with its card name and prose", () => {
    render(<ScorePromptsPanel prompts={[makePrompt()]} onConfirm={() => {}} />);
    expect(screen.getByText("Bring It Down")).toBeTruthy();
    expect(
      screen.getByText("Score 2 VP per destroyed enemy unit with Starting Strength 13+."),
    ).toBeTruthy();
  });

  it("confirms with the chosen count after incrementing", async () => {
    const user = userEvent.setup();
    const onConfirm = vi.fn();
    render(<ScorePromptsPanel prompts={[makePrompt()]} onConfirm={onConfirm} />);

    await user.click(screen.getByLabelText("Increase count"));
    await user.click(screen.getByLabelText("Increase count"));
    await user.click(screen.getByText("Confirm"));

    expect(onConfirm).toHaveBeenCalledWith("prompt-1", 2);
  });

  it("does not let the count go below zero", async () => {
    const user = userEvent.setup();
    const onConfirm = vi.fn();
    render(<ScorePromptsPanel prompts={[makePrompt()]} onConfirm={onConfirm} />);

    expect((screen.getByLabelText("Decrease count") as HTMLButtonElement).disabled).toBe(true);
    await user.click(screen.getByText("Confirm"));
    expect(onConfirm).toHaveBeenCalledWith("prompt-1", 0);
  });
});
