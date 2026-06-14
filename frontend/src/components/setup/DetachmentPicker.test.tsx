import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { DetachmentPicker } from "./DetachmentPicker";
import { mockDetachments } from "../../test/fixtures";

describe("DetachmentPicker", () => {
  it("shows the detachment-points budget", () => {
    render(
      <DetachmentPicker
        detachments={mockDetachments}
        selectedIds={[]}
        maxPoints={3}
        onChange={() => {}}
      />,
    );
    expect(screen.getByText("Detachment points: 0/3")).toBeTruthy();
  });

  it("emits the selected detachment with its points when toggled on", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(
      <DetachmentPicker
        detachments={mockDetachments}
        selectedIds={[]}
        maxPoints={3}
        onChange={onChange}
      />,
    );
    await user.click(screen.getByRole("button", { name: /Gladius Task Force/ }));
    expect(onChange).toHaveBeenCalledWith([
      { id: "det-gladius", name: "Gladius Task Force", points: 3 },
    ]);
  });

  it("disables a detachment that would exceed the budget", () => {
    // Gladius (3 pts) already chosen → no room for Ironstorm (2 pts).
    render(
      <DetachmentPicker
        detachments={mockDetachments}
        selectedIds={["det-gladius"]}
        maxPoints={3}
        onChange={() => {}}
      />,
    );
    expect(screen.getByText("Detachment points: 3/3")).toBeTruthy();
    expect(
      (screen.getByRole("button", { name: /Ironstorm Spearhead/ }) as HTMLButtonElement).disabled,
    ).toBe(true);
  });

  it("deselects an already-selected detachment", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(
      <DetachmentPicker
        detachments={mockDetachments}
        selectedIds={["det-gladius"]}
        maxPoints={3}
        onChange={onChange}
      />,
    );
    await user.click(screen.getByRole("button", { name: /Gladius Task Force/ }));
    expect(onChange).toHaveBeenCalledWith([]);
  });
});
