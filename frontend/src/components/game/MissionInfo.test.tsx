import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MissionInfo } from "./MissionInfo";
import { mockMissions } from "../../test/fixtures";

const mission = mockMissions[0]; // Supply Drop

describe("MissionInfo", () => {
  it("is collapsed by default", () => {
    render(<MissionInfo mission={mission} />);

    expect(screen.getByText("Mission Info")).toBeTruthy();
    expect(screen.queryByText("Primary Mission")).toBeNull();
  });

  it("expands to show the mission name and VP caps", async () => {
    const user = userEvent.setup();
    render(<MissionInfo mission={mission} />);

    await user.click(screen.getByText("Mission Info"));

    expect(screen.getByText("Primary Mission")).toBeTruthy();
    expect(screen.getByText("Supply Drop")).toBeTruthy();
    expect(screen.getByText(`${mission.vpPerRoundCap} VP / Round`)).toBeTruthy();
    expect(screen.getByText(`${mission.vpPerGameCap} VP / Game`)).toBeTruthy();
  });

  it("shows 'None' when the mission is null", async () => {
    const user = userEvent.setup();
    render(<MissionInfo mission={null} />);

    await user.click(screen.getByText("Mission Info"));

    const noneElements = screen.getAllByText("None");
    expect(noneElements).toHaveLength(1);
  });

  it("collapses when clicked again", async () => {
    const user = userEvent.setup();
    render(<MissionInfo mission={mission} />);

    await user.click(screen.getByText("Mission Info"));
    expect(screen.getByText("Supply Drop")).toBeTruthy();

    await user.click(screen.getByText("Mission Info"));
    expect(screen.queryByText("Supply Drop")).toBeNull();
  });
});
