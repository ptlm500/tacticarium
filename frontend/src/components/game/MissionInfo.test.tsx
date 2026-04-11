import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MissionInfo } from "./MissionInfo";
import { mockMissions, mockRules } from "../../test/fixtures";

const mission = mockMissions[0]; // Supply Drop
const twist = mockRules[0]; // Hidden Supplies

describe("MissionInfo", () => {
  it("is collapsed by default", () => {
    render(<MissionInfo mission={mission} twist={twist} />);

    expect(screen.getByText("Mission Info")).toBeTruthy();
    expect(screen.queryByText("Primary Mission")).toBeNull();
  });

  it("expands to show mission and twist details", async () => {
    const user = userEvent.setup();
    render(<MissionInfo mission={mission} twist={twist} />);

    await user.click(screen.getByText("Mission Info"));

    expect(screen.getByText("Primary Mission")).toBeTruthy();
    expect(screen.getByText("Supply Drop")).toBeTruthy();
    expect(screen.getByText("Control objectives to score VP.")).toBeTruthy();

    expect(screen.getByText("Twist")).toBeTruthy();
    expect(screen.getByText("Hidden Supplies")).toBeTruthy();
    expect(screen.getByText("Additional objectives appear.")).toBeTruthy();
  });

  it("shows scoring rules with VP values and round requirements", async () => {
    const user = userEvent.setup();
    render(<MissionInfo mission={mission} twist={twist} />);

    await user.click(screen.getByText("Mission Info"));

    expect(screen.getByText("Scoring:")).toBeTruthy();
    expect(screen.getByText("+5 VP")).toBeTruthy();
    expect(screen.getByText("2 objectives")).toBeTruthy();
    expect(screen.getByText("+10 VP")).toBeTruthy();
    expect(screen.getByText("3+ objectives")).toBeTruthy();
    // Both rules have minRound: 2
    const roundBadges = screen.getAllByText("(R2+)");
    expect(roundBadges).toHaveLength(2);
  });

  it("does not show round badge for minRound 1 or missing", async () => {
    const user = userEvent.setup();
    const missionNoMinRound = {
      ...mockMissions[1], // Scorched Earth — no minRound on rules
    };
    render(<MissionInfo mission={missionNoMinRound} twist={twist} />);

    await user.click(screen.getByText("Mission Info"));

    expect(screen.getByText("Burned 1")).toBeTruthy();
    expect(screen.queryByText(/\(R\d+\+\)/)).toBeNull();
  });

  it("shows 'None' when mission or twist is null", async () => {
    const user = userEvent.setup();
    render(<MissionInfo mission={null} twist={null} />);

    await user.click(screen.getByText("Mission Info"));

    const noneElements = screen.getAllByText("None");
    expect(noneElements).toHaveLength(2);
  });

  it("collapses when clicked again", async () => {
    const user = userEvent.setup();
    render(<MissionInfo mission={mission} twist={twist} />);

    await user.click(screen.getByText("Mission Info"));
    expect(screen.getByText("Supply Drop")).toBeTruthy();

    await user.click(screen.getByText("Mission Info"));
    expect(screen.queryByText("Supply Drop")).toBeNull();
  });
});
