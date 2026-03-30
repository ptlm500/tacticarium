import { screen } from "@testing-library/react";
import { renderWithProviders } from "../test/renderWithProviders";
import { GameSetupPage } from "./GameSetupPage";
import { useGameStore } from "../stores/gameStore";
import { makeGameState, makePlayerState, mockUser } from "../test/fixtures";
import { gameWs } from "../mocks/handlers/ws";
import { Route, Routes } from "react-router-dom";

function renderSetup() {
  const gs = makeGameState({
    status: "setup",
    missionId: "",
    missionName: "",
    twistId: "",
    twistName: "",
    players: [
      makePlayerState({
        factionId: "",
        factionName: "",
        detachmentId: "",
        detachmentName: "",
        secondaryMode: "",
        ready: false,
      }),
      null,
    ],
  });
  useGameStore.getState().setGameState(gs);
  localStorage.setItem("token", "test-token");

  gameWs.addEventListener("connection", ({ client }) => {
    client.send(JSON.stringify({ type: "state_update", data: gs }));
  });

  return renderWithProviders(
    <Routes>
      <Route path="/game/:id/setup" element={<GameSetupPage />} />
    </Routes>,
    { user: mockUser, route: "/game/game-1/setup" },
  );
}

describe("GameSetupPage", () => {
  beforeEach(() => {
    useGameStore.getState().reset();
    localStorage.clear();
  });

  it("renders the setup page header", async () => {
    renderSetup();

    await vi.waitFor(() => {
      expect(screen.getByText("Game Setup")).toBeTruthy();
    });
  });

  it("shows the invite code", async () => {
    renderSetup();

    await vi.waitFor(() => {
      expect(screen.getByText(/Invite: ABC123/)).toBeTruthy();
    });
  });

  it("renders faction picker from API", async () => {
    renderSetup();

    await vi.waitFor(() => {
      expect(screen.getByText("Your Faction")).toBeTruthy();
    });

    // MSW returns mockFactions: Space Marines, Chaos Space Marines, Orks
    await vi.waitFor(() => {
      expect(screen.getByText("Space Marines")).toBeTruthy();
      expect(screen.getByText("Chaos Space Marines")).toBeTruthy();
      expect(screen.getByText("Orks")).toBeTruthy();
    });
  });

  it("shows Ready Up button disabled when setup is incomplete", async () => {
    renderSetup();

    await vi.waitFor(() => {
      const readyBtn = screen.getByText("Ready Up");
      expect(readyBtn.closest("button")!.hasAttribute("disabled")).toBe(true);
    });
  });

  it("shows detachment section when faction is selected", async () => {
    const gs = makeGameState({
      status: "setup",
      missionId: "",
      missionName: "",
      twistId: "",
      twistName: "",
      players: [
        makePlayerState({
          factionId: "faction-sm",
          factionName: "Space Marines",
          detachmentId: "",
          detachmentName: "",
          secondaryMode: "",
          ready: false,
        }),
        null,
      ],
    });
    useGameStore.getState().setGameState(gs);
    localStorage.setItem("token", "test-token");

    gameWs.addEventListener("connection", ({ client }) => {
      client.send(JSON.stringify({ type: "state_update", data: gs }));
    });

    renderWithProviders(
      <Routes>
        <Route path="/game/:id/setup" element={<GameSetupPage />} />
      </Routes>,
      { user: mockUser, route: "/game/game-1/setup" },
    );

    await vi.waitFor(() => {
      expect(screen.getByText("Detachment")).toBeTruthy();
    });

    // MSW returns mockDetachments for faction-sm
    await vi.waitFor(() => {
      expect(screen.getByText("Gladius Task Force")).toBeTruthy();
      expect(screen.getByText("Ironstorm Spearhead")).toBeTruthy();
    });
  });

  it("shows mission section when detachment is selected", async () => {
    const gs = makeGameState({
      status: "setup",
      missionId: "",
      missionName: "",
      twistId: "",
      twistName: "",
      players: [
        makePlayerState({
          detachmentId: "det-gladius",
          detachmentName: "Gladius Task Force",
          secondaryMode: "",
          ready: false,
        }),
        null,
      ],
    });
    useGameStore.getState().setGameState(gs);
    localStorage.setItem("token", "test-token");

    gameWs.addEventListener("connection", ({ client }) => {
      client.send(JSON.stringify({ type: "state_update", data: gs }));
    });

    renderWithProviders(
      <Routes>
        <Route path="/game/:id/setup" element={<GameSetupPage />} />
      </Routes>,
      { user: mockUser, route: "/game/game-1/setup" },
    );

    await vi.waitFor(() => {
      expect(screen.getByText("Primary Mission")).toBeTruthy();
    });

    await vi.waitFor(() => {
      expect(screen.getByText("Supply Drop")).toBeTruthy();
      expect(screen.getByText("Scorched Earth")).toBeTruthy();
    });
  });
});
