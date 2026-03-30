import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { StratagemPanel } from "./StratagemPanel";
import { mockStratagems } from "../../test/fixtures";

describe("StratagemPanel", () => {
  it("shows empty message when no stratagems available", () => {
    render(<StratagemPanel stratagems={[]} currentCP={5} onUse={vi.fn()} />);
    expect(screen.getByText("No stratagems available for this phase.")).toBeTruthy();
  });

  it("renders stratagem names and CP costs", () => {
    render(<StratagemPanel stratagems={mockStratagems} currentCP={5} onUse={vi.fn()} />);
    expect(screen.getByText("Command Re-roll")).toBeTruthy();
    expect(screen.getByText("Storm of Fire")).toBeTruthy();
    expect(screen.getByText("Heroic Intervention")).toBeTruthy();
  });

  it("renders stratagem type and phase/turn info", () => {
    render(<StratagemPanel stratagems={[mockStratagems[0]]} currentCP={5} onUse={vi.fn()} />);
    expect(screen.getByText("Core")).toBeTruthy();
    expect(screen.getByText("Either player's turn | Any phase")).toBeTruthy();
  });

  it("disables Use button when CP is insufficient", () => {
    render(<StratagemPanel stratagems={[mockStratagems[2]]} currentCP={1} onUse={vi.fn()} />);
    const useBtn = screen.getByRole("button", { name: "Use" });
    expect(useBtn.hasAttribute("disabled")).toBe(true);
  });

  it("enables Use button when CP is sufficient", () => {
    render(<StratagemPanel stratagems={[mockStratagems[0]]} currentCP={5} onUse={vi.fn()} />);
    const useBtn = screen.getByRole("button", { name: "Use" });
    expect(useBtn.hasAttribute("disabled")).toBe(false);
  });

  it("calls onUse with the stratagem when confirmed", async () => {
    const user = userEvent.setup();
    const onUse = vi.fn();
    vi.spyOn(window, "confirm").mockReturnValue(true);

    render(<StratagemPanel stratagems={[mockStratagems[0]]} currentCP={5} onUse={onUse} />);

    await user.click(screen.getByRole("button", { name: "Use" }));
    expect(onUse).toHaveBeenCalledWith(mockStratagems[0]);

    vi.restoreAllMocks();
  });

  it("does not call onUse when confirm is cancelled", async () => {
    const user = userEvent.setup();
    const onUse = vi.fn();
    vi.spyOn(window, "confirm").mockReturnValue(false);

    render(<StratagemPanel stratagems={[mockStratagems[0]]} currentCP={5} onUse={onUse} />);

    await user.click(screen.getByRole("button", { name: "Use" }));
    expect(onUse).not.toHaveBeenCalled();

    vi.restoreAllMocks();
  });

  it("renders legend text when provided", () => {
    const stratagemWithLegend = {
      ...mockStratagems[0],
      legend: "A powerful re-roll ability",
    };
    render(<StratagemPanel stratagems={[stratagemWithLegend]} currentCP={5} onUse={vi.fn()} />);
    expect(screen.getByText("A powerful re-roll ability")).toBeTruthy();
  });
});
