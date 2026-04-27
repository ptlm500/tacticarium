import { screen, act } from "@testing-library/react";
import { ws, http, HttpResponse } from "msw";
import { Route, Routes } from "react-router-dom";
import { renderWithProviders } from "../test/renderWithProviders";
import { SpectatorPage } from "./SpectatorPage";
import { useGameStore } from "../stores/gameStore";
import { makeGameState, makePlayerState, mockActiveSecondary } from "../test/fixtures";
import { worker } from "../mocks/browser";

const API_URL = "http://localhost:8080";

function mockSpectatorWS(gs: ReturnType<typeof makeGameState>) {
  const link = ws.link("ws://localhost:8080/ws/game/*/spectate");
  worker.use(
    link.addEventListener("connection", ({ client }) => {
      client.send(JSON.stringify({ type: "state_update", data: gs }));
    }),
    http.get(`${API_URL}/api/games/:id/events`, () => HttpResponse.json([])),
  );
}

function renderSpectator(route = "/game/game-1/spectate") {
  return renderWithProviders(
    <Routes>
      <Route path="/game/:id/spectate" element={<SpectatorPage />} />
    </Routes>,
    { user: null, route },
  );
}

describe("SpectatorPage", () => {
  beforeEach(() => {
    useGameStore.getState().reset();
    localStorage.clear();
  });

  it("renders both player panels for an active game", async () => {
    const gs = makeGameState({
      players: [
        makePlayerState({
          username: "Alpha",
          factionName: "Space Marines",
          cp: 4,
          vpPrimary: 7,
        }),
        makePlayerState({
          userId: "user-2",
          playerNumber: 2,
          username: "Bravo",
          factionName: "Necrons",
          cp: 2,
          vpPrimary: 3,
          activeSecondaries: [mockActiveSecondary],
        }),
      ],
    });
    mockSpectatorWS(gs);

    await act(async () => {
      renderSpectator();
    });

    await vi.waitFor(() => {
      expect(screen.getByText(/Alpha — Space Marines/)).toBeTruthy();
    });
    expect(screen.getByText(/Bravo — Necrons/)).toBeTruthy();
    expect(screen.getByText(mockActiveSecondary.name)).toBeTruthy();
  });

  it("marks the active player's panel", async () => {
    const gs = makeGameState({
      activePlayer: 2,
      players: [
        makePlayerState({ username: "P1" }),
        makePlayerState({
          userId: "user-2",
          playerNumber: 2,
          username: "P2",
        }),
      ],
    });
    mockSpectatorWS(gs);

    await act(async () => {
      renderSpectator();
    });

    await vi.waitFor(() => {
      expect(screen.getByText(/P2 — Space Marines/)).toBeTruthy();
    });
    const activeBadges = screen.getAllByText(/Active Turn/i);
    expect(activeBadges).toHaveLength(1);
  });

  it("shows a placeholder when the game is in setup", async () => {
    const gs = makeGameState({ status: "setup" });
    mockSpectatorWS(gs);

    await act(async () => {
      renderSpectator();
    });

    await vi.waitFor(() => {
      expect(screen.getByText(/Awaiting Battle/i)).toBeTruthy();
    });
    expect(screen.queryByText(/Battle Round/)).toBeNull();
  });

  it("shows an ended placeholder for a completed game", async () => {
    const gs = makeGameState({ status: "completed" });
    mockSpectatorWS(gs);

    await act(async () => {
      renderSpectator();
    });

    await vi.waitFor(() => {
      expect(screen.getByText(/Battle Complete/i)).toBeTruthy();
    });
  });

  it("renders without an authenticated user", async () => {
    const gs = makeGameState();
    mockSpectatorWS(gs);

    await act(async () => {
      renderSpectator();
    });

    await vi.waitFor(() => {
      expect(screen.getByText(/Spectator/i)).toBeTruthy();
    });
  });

  it("updates rendered values from a live state_update", async () => {
    const initial = makeGameState({
      currentRound: 1,
      players: [
        makePlayerState({ username: "Alpha" }),
        makePlayerState({ userId: "user-2", playerNumber: 2, username: "Bravo" }),
      ],
    });
    const updated = makeGameState({
      currentRound: 4,
      players: [
        makePlayerState({ username: "Alpha" }),
        makePlayerState({ userId: "user-2", playerNumber: 2, username: "Bravo" }),
      ],
    });

    const link = ws.link("ws://localhost:8080/ws/game/*/spectate");
    worker.use(
      link.addEventListener("connection", ({ client }) => {
        client.send(JSON.stringify({ type: "state_update", data: initial }));
        setTimeout(() => {
          client.send(JSON.stringify({ type: "state_update", data: updated }));
        }, 10);
      }),
      http.get(`${API_URL}/api/games/:id/events`, () => HttpResponse.json([])),
    );

    await act(async () => {
      renderSpectator();
    });

    await vi.waitFor(() => {
      expect(screen.getByText(/Battle Round 4/)).toBeTruthy();
    });
  });

  it("resets the game store on unmount", async () => {
    const gs = makeGameState();
    mockSpectatorWS(gs);

    let result!: ReturnType<typeof renderSpectator>;
    await act(async () => {
      result = renderSpectator();
    });
    await vi.waitFor(() => {
      expect(useGameStore.getState().gameState).not.toBeNull();
    });

    await act(async () => {
      result.unmount();
    });

    expect(useGameStore.getState().gameState).toBeNull();
  });
});
