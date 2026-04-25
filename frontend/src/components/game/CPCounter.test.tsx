import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { CPCounter } from "./CPCounter";

describe("CPCounter", () => {
  it("calls onAdjust(+1) when increase is clicked", async () => {
    const user = userEvent.setup();
    const onAdjust = vi.fn();
    render(<CPCounter cp={3} canGainCP={true} onAdjust={onAdjust} />);

    await user.click(screen.getByLabelText("Increase CP"));
    expect(onAdjust).toHaveBeenCalledWith(1);
  });

  it("calls onAdjust(+1) even when cap is reached (parent handles confirm)", async () => {
    const user = userEvent.setup();
    const onAdjust = vi.fn();
    render(<CPCounter cp={3} canGainCP={false} onAdjust={onAdjust} />);

    const increaseBtn = screen.getByLabelText("Increase CP") as HTMLButtonElement;
    expect(increaseBtn.disabled).toBe(false);

    await user.click(increaseBtn);
    expect(onAdjust).toHaveBeenCalledWith(1);
  });

  it("shows the cap-reached warning when canGainCP is false", () => {
    render(<CPCounter cp={3} canGainCP={false} onAdjust={vi.fn()} />);
    expect(screen.getByText("CP gain cap reached")).toBeTruthy();
  });

  it("hides the cap-reached warning when canGainCP is true", () => {
    render(<CPCounter cp={3} canGainCP={true} onAdjust={vi.fn()} />);
    expect(screen.queryByText("CP gain cap reached")).toBeNull();
  });

  it("disables the decrease button when CP is 0", () => {
    render(<CPCounter cp={0} canGainCP={true} onAdjust={vi.fn()} />);
    const decreaseBtn = screen.getByLabelText("Decrease CP") as HTMLButtonElement;
    expect(decreaseBtn.disabled).toBe(true);
  });
});
