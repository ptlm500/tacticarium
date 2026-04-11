import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { TacticalDrawReminder } from "./TacticalDrawReminder";

function renderReminder(overrides: Partial<Parameters<typeof TacticalDrawReminder>[0]> = {}) {
  const defaultProps = {
    deckSize: 5,
    activeCount: 1,
    onDraw: vi.fn(),
    ...overrides,
  };
  return { ...render(<TacticalDrawReminder {...defaultProps} />), props: defaultProps };
}

describe("TacticalDrawReminder", () => {
  it("shows draw button when eligible", async () => {
    const user = userEvent.setup();
    const { props } = renderReminder({ activeCount: 1, deckSize: 5 });

    const drawBtn = screen.getByText(/Draw Secondaries/);
    await user.click(drawBtn);
    expect(props.onDraw).toHaveBeenCalledOnce();
  });

  it("shows active count and draw prompt with 0 active", () => {
    renderReminder({ activeCount: 0, deckSize: 5 });
    expect(
      screen.getByText("You have 0/2 active secondaries. Draw to fill your active slots."),
    ).toBeTruthy();
  });

  it("shows active count and draw prompt with 1 active", () => {
    renderReminder({ activeCount: 1, deckSize: 5 });
    expect(
      screen.getByText("You have 1/2 active secondaries. Draw to fill your active slots."),
    ).toBeTruthy();
  });

  it('shows "already have 2" when at max', () => {
    renderReminder({ activeCount: 2, deckSize: 5 });
    expect(screen.getByText("You already have 2 active secondaries.")).toBeTruthy();
  });

  it('shows "deck empty" when no cards left', () => {
    renderReminder({ activeCount: 0, deckSize: 0 });
    expect(screen.getByText("Deck is empty.")).toBeTruthy();
  });

  it("shows remaining deck size on draw button", () => {
    renderReminder({ activeCount: 0, deckSize: 3 });
    expect(screen.getByText("Draw Secondaries (3 remaining)")).toBeTruthy();
  });

  it("hides draw button when deck is empty", () => {
    renderReminder({ activeCount: 0, deckSize: 0 });
    expect(screen.queryByText(/Draw Secondaries/)).toBeNull();
  });
});
