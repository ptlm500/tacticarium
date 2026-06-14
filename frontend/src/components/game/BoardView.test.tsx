import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { BoardView } from "./BoardView";
import { mockBoardWithObjectives, mockDeploymentPatterns } from "../../test/fixtures";
import type { Board } from "../../types/game";

const emptyBoard: Board = { deploymentPatternId: "", objectives: [], playerSides: [] };

describe("BoardView", () => {
  it("renders the empty state when no objectives are placed", () => {
    render(<BoardView board={emptyBoard} pattern={null} onSetControl={() => {}} />);
    expect(screen.getByText("No deployment objectives placed")).toBeTruthy();
  });

  it("renders one control button per objective with role + controller in the label", () => {
    render(
      <BoardView
        board={mockBoardWithObjectives}
        pattern={mockDeploymentPatterns[0]}
        onSetControl={() => {}}
      />,
    );
    expect(
      screen.getByRole("button", { name: /Objective 1 \(central\), controlled by no one/ }),
    ).toBeTruthy();
    expect(
      screen.getByRole("button", { name: /Objective 2 \(home\), controlled by Player 1/ }),
    ).toBeTruthy();
    expect(
      screen.getByRole("button", { name: /Objective 3 \(expansion\), controlled by Player 2/ }),
    ).toBeTruthy();
  });

  it("cycles control none → P1 when an uncontrolled objective is clicked", async () => {
    const user = userEvent.setup();
    const onSetControl = vi.fn();
    render(
      <BoardView
        board={mockBoardWithObjectives}
        pattern={mockDeploymentPatterns[0]}
        onSetControl={onSetControl}
      />,
    );
    await user.click(screen.getByRole("button", { name: /Objective 1 \(central\)/ }));
    // objective index 0, controlledBy 0 → next is 1
    expect(onSetControl).toHaveBeenCalledWith(0, 1);
  });

  it("cycles control P2 → none when a P2-controlled objective is clicked", async () => {
    const user = userEvent.setup();
    const onSetControl = vi.fn();
    render(
      <BoardView
        board={mockBoardWithObjectives}
        pattern={mockDeploymentPatterns[0]}
        onSetControl={onSetControl}
      />,
    );
    await user.click(screen.getByRole("button", { name: /Objective 3 \(expansion\)/ }));
    // objective index 2, controlledBy 2 → next is 0
    expect(onSetControl).toHaveBeenCalledWith(2, 0);
  });
});
