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

  it("Use button is always enabled (user can always open the prompt)", () => {
    render(<StratagemPanel stratagems={[mockStratagems[2]]} currentCP={0} onUse={vi.fn()} />);
    const useBtn = screen.getByRole("button", { name: "Use" });
    expect(useBtn.hasAttribute("disabled")).toBe(false);
  });

  it("opens confirmation modal with default CP cost prefilled", async () => {
    const user = userEvent.setup();
    render(<StratagemPanel stratagems={[mockStratagems[0]]} currentCP={5} onUse={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: "Use" }));

    const input = screen.getByLabelText(/CP to spend/i) as HTMLInputElement;
    expect(input.value).toBe("1");
  });

  it("calls onUse with stratagem and default cost when confirmed unchanged", async () => {
    const user = userEvent.setup();
    const onUse = vi.fn();
    render(<StratagemPanel stratagems={[mockStratagems[0]]} currentCP={5} onUse={onUse} />);

    await user.click(screen.getByRole("button", { name: "Use" }));
    await user.click(screen.getByRole("button", { name: "Confirm" }));

    expect(onUse).toHaveBeenCalledWith(mockStratagems[0], 1);
  });

  it("allows overriding the CP cost (e.g. free stratagem)", async () => {
    const user = userEvent.setup();
    const onUse = vi.fn();
    render(<StratagemPanel stratagems={[mockStratagems[0]]} currentCP={5} onUse={onUse} />);

    await user.click(screen.getByRole("button", { name: "Use" }));
    const input = screen.getByLabelText(/CP to spend/i);
    await user.clear(input);
    await user.type(input, "0");
    await user.click(screen.getByRole("button", { name: "Confirm" }));

    expect(onUse).toHaveBeenCalledWith(mockStratagems[0], 0);
  });

  it("allows overriding the CP cost higher than default", async () => {
    const user = userEvent.setup();
    const onUse = vi.fn();
    render(<StratagemPanel stratagems={[mockStratagems[0]]} currentCP={5} onUse={onUse} />);

    await user.click(screen.getByRole("button", { name: "Use" }));
    const input = screen.getByLabelText(/CP to spend/i);
    await user.clear(input);
    await user.type(input, "3");
    await user.click(screen.getByRole("button", { name: "Confirm" }));

    expect(onUse).toHaveBeenCalledWith(mockStratagems[0], 3);
  });

  it("blocks confirm when the entered cost exceeds available CP", async () => {
    const user = userEvent.setup();
    const onUse = vi.fn();
    render(<StratagemPanel stratagems={[mockStratagems[0]]} currentCP={2} onUse={onUse} />);

    await user.click(screen.getByRole("button", { name: "Use" }));
    const input = screen.getByLabelText(/CP to spend/i);
    await user.clear(input);
    await user.type(input, "5");

    const confirmBtn = screen.getByRole("button", { name: "Confirm" });
    expect(confirmBtn.hasAttribute("disabled")).toBe(true);
    expect(screen.getByText(/only have 2 CP/i)).toBeTruthy();

    await user.click(confirmBtn);
    expect(onUse).not.toHaveBeenCalled();
  });

  it("does not call onUse when cancelled", async () => {
    const user = userEvent.setup();
    const onUse = vi.fn();
    render(<StratagemPanel stratagems={[mockStratagems[0]]} currentCP={5} onUse={onUse} />);

    await user.click(screen.getByRole("button", { name: "Use" }));
    await user.click(screen.getByRole("button", { name: "Cancel" }));

    expect(onUse).not.toHaveBeenCalled();
    expect(screen.queryByLabelText(/CP to spend/i)).toBeNull();
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
