import { screen, act } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { renderWithProviders } from "../test/renderWithProviders";
import { GameSetupPage } from "./GameSetupPage";
import { useGameStore } from "../stores/gameStore";
import { makeGameState, makePlayerState, mockUser } from "../test/fixtures";
import { ws } from "msw";
import { worker } from "../mocks/browser";
import { Route, Routes } from "react-router-dom";

function renderSetup() {
  const gs = makeGameState({
    status: "setup",
    players: [
      makePlayerState({
        factionId: "",
        factionName: "",
        detachments: [],
        secondaryMode: "",
        ready: false,
      }),
      null,
    ],
  });
  useGameStore.getState().setGameState(gs);
  localStorage.setItem("token", "test-token");

  const testLink = ws.link("ws://localhost:8080/ws/game/*");
  worker.use(
    testLink.addEventListener("connection", ({ client }) => {
      client.send(JSON.stringify({ type: "state_update", data: gs }));
    }),
  );

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
    await act(async () => {
      renderSetup();
    });

    await vi.waitFor(() => {
      expect(screen.getByText("Game Setup")).toBeTruthy();
    });
  });

  it("shows the invite code", async () => {
    await act(async () => {
      renderSetup();
    });

    await vi.waitFor(() => {
      expect(screen.getByText(/Invite: ABC123/)).toBeTruthy();
    });
  });

  it("renders faction picker from API", async () => {
    await act(async () => {
      renderSetup();
    });

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
    await act(async () => {
      renderSetup();
    });

    await vi.waitFor(() => {
      const readyBtn = screen.getByText("Ready Up");
      expect(readyBtn.closest("button")!.hasAttribute("disabled")).toBe(true);
    });
  });

  it("shows detachment section when faction is selected", async () => {
    const gs = makeGameState({
      status: "setup",
      players: [
        makePlayerState({
          factionId: "faction-sm",
          factionName: "Space Marines",
          detachments: [],
          secondaryMode: "",
          ready: false,
        }),
        null,
      ],
    });
    useGameStore.getState().setGameState(gs);
    localStorage.setItem("token", "test-token");

    const testLink = ws.link("ws://localhost:8080/ws/game/*");
    worker.use(
      testLink.addEventListener("connection", ({ client }) => {
        client.send(JSON.stringify({ type: "state_update", data: gs }));
      }),
    );

    await act(async () => {
      renderWithProviders(
        <Routes>
          <Route path="/game/:id/setup" element={<GameSetupPage />} />
        </Routes>,
        { user: mockUser, route: "/game/game-1/setup" },
      );
    });

    await vi.waitFor(() => {
      expect(screen.getByText("Detachments")).toBeTruthy();
    });

    // MSW returns mockDetachments for faction-sm
    await vi.waitFor(() => {
      expect(screen.getByText("Gladius Task Force")).toBeTruthy();
      expect(screen.getByText("Ironstorm Spearhead")).toBeTruthy();
    });
  });

  it("sends select_detachments with the chosen detachment and its points", async () => {
    const gs = makeGameState({
      status: "setup",
      players: [
        makePlayerState({
          detachments: [],
          side: "",
          forceDisposition: "",
          missionId: "",
          missionName: "",
          secondaryMode: "",
          ready: false,
        }),
        null,
      ],
    });
    useGameStore.getState().setGameState(gs);
    localStorage.setItem("token", "test-token");

    const sentActions: Array<Record<string, unknown>> = [];
    const testLink = ws.link("ws://localhost:8080/ws/game/*");
    worker.use(
      testLink.addEventListener("connection", ({ client }) => {
        client.send(JSON.stringify({ type: "state_update", data: gs }));
        client.addEventListener("message", (ev) => {
          if (typeof ev.data !== "string") return;
          try {
            const parsed = JSON.parse(ev.data);
            if (parsed?.type === "action" && parsed.data) {
              sentActions.push(parsed.data as Record<string, unknown>);
            }
          } catch {
            // ignore non-JSON frames
          }
        });
      }),
    );

    await act(async () => {
      renderWithProviders(
        <Routes>
          <Route path="/game/:id/setup" element={<GameSetupPage />} />
        </Routes>,
        { user: mockUser, route: "/game/game-1/setup" },
      );
    });

    const user = userEvent.setup();

    await vi.waitFor(() => {
      expect(screen.getByText("Gladius Task Force")).toBeTruthy();
    });
    await user.click(screen.getByRole("button", { name: /Gladius Task Force/ }));

    await vi.waitFor(() => {
      const action = sentActions.find((a) => a.type === "select_detachments");
      expect(action).toBeTruthy();
      expect(action!.detachments).toEqual([
        { id: "det-gladius", name: "Gladius Task Force", points: 3 },
      ]);
    });
  });

  it("shows first-player picker once a disposition is chosen and hides secondary section until chosen", async () => {
    const gs = makeGameState({
      status: "setup",
      firstTurnPlayer: 0, // not yet chosen
      players: [
        makePlayerState({
          factionId: "faction-sm",
          factionName: "Space Marines",
          detachments: [{ id: "det-gladius", name: "Gladius Task Force", points: 3 }],
          side: "attacker",
          forceDisposition: "take-and-hold",
          forceDispositionName: "Take and Hold",
          missionId: "",
          missionName: "",
          secondaryMode: "",
          ready: false,
        }),
        makePlayerState({
          userId: "user-2",
          username: "Opponent",
          playerNumber: 2,
          factionId: "faction-csm",
          factionName: "Chaos Space Marines",
          detachments: [{ id: "det-black-legion", name: "Black Legion", points: 2 }],
          secondaryMode: "",
          ready: false,
        }),
      ],
    });
    useGameStore.getState().setGameState(gs);
    localStorage.setItem("token", "test-token");

    const testLink = ws.link("ws://localhost:8080/ws/game/*");
    worker.use(
      testLink.addEventListener("connection", ({ client }) => {
        client.send(JSON.stringify({ type: "state_update", data: gs }));
      }),
    );

    await act(async () => {
      renderWithProviders(
        <Routes>
          <Route path="/game/:id/setup" element={<GameSetupPage />} />
        </Routes>,
        { user: mockUser, route: "/game/game-1/setup" },
      );
    });

    // First player picker is visible
    await vi.waitFor(() => {
      expect(screen.getByText("Who Goes First?")).toBeTruthy();
    });

    // Prompt to pick first is shown while not yet chosen
    expect(screen.getByText(/Pick who goes first before readying up/)).toBeTruthy();

    // Secondary missions section is gated until first player is chosen
    expect(screen.queryByText("Secondary Missions")).toBeNull();

    // Ready Up button is disabled
    const readyBtn = screen.getByText("Ready Up").closest("button")!;
    expect(readyBtn.hasAttribute("disabled")).toBe(true);
  });

  it("resets gameStore when unmounted", async () => {
    let result!: ReturnType<typeof renderSetup>;
    await act(async () => {
      result = renderSetup();
    });

    await vi.waitFor(() => {
      expect(useGameStore.getState().gameState).not.toBeNull();
    });

    await act(async () => {
      result.unmount();
    });

    expect(useGameStore.getState().gameState).toBeNull();
  });

  it("does not redirect to /game/:id when stored state is from a different game", async () => {
    // Stale state from a previously-open active game.
    useGameStore.getState().setGameState(
      makeGameState({
        gameId: "game-OLD",
        status: "active",
      }),
    );
    localStorage.setItem("token", "test-token");

    // The new game we're navigating to is in setup.
    const newGameState = makeGameState({
      gameId: "game-NEW",
      status: "setup",
      players: [
        makePlayerState({
          factionId: "",
          factionName: "",
          detachments: [],
          secondaryMode: "",
          ready: false,
        }),
        null,
      ],
    });

    const testLink = ws.link("ws://localhost:8080/ws/game/*");
    worker.use(
      testLink.addEventListener("connection", ({ client }) => {
        client.send(JSON.stringify({ type: "state_update", data: newGameState }));
      }),
    );

    await act(async () => {
      renderWithProviders(
        <Routes>
          <Route path="/game/:id/setup" element={<GameSetupPage />} />
          <Route path="/game/:id" element={<div>WRONG_GAME_PAGE</div>} />
        </Routes>,
        { user: mockUser, route: "/game/game-NEW/setup" },
      );
    });

    await vi.waitFor(() => {
      expect(screen.getByText("Game Setup")).toBeTruthy();
    });
    expect(screen.queryByText("WRONG_GAME_PAGE")).toBeNull();
  });

  it("shows the army-painted toggle once detachment is selected and reflects vpPaint state", async () => {
    const gs = makeGameState({
      status: "setup",
      players: [
        makePlayerState({
          detachments: [{ id: "det-gladius", name: "Gladius Task Force", points: 3 }],
          secondaryMode: "",
          ready: false,
          vpPaint: 10,
        }),
        null,
      ],
    });
    useGameStore.getState().setGameState(gs);
    localStorage.setItem("token", "test-token");

    const testLink = ws.link("ws://localhost:8080/ws/game/*");
    worker.use(
      testLink.addEventListener("connection", ({ client }) => {
        client.send(JSON.stringify({ type: "state_update", data: gs }));
      }),
    );

    await act(async () => {
      renderWithProviders(
        <Routes>
          <Route path="/game/:id/setup" element={<GameSetupPage />} />
        </Routes>,
        { user: mockUser, route: "/game/game-1/setup" },
      );
    });

    await vi.waitFor(() => {
      expect(screen.getByText("Army Painted")).toBeTruthy();
    });

    const toggle = screen.getByRole("switch");
    expect(toggle.getAttribute("aria-checked")).toBe("true");
    expect(screen.getByText("Painted")).toBeTruthy();
    expect(screen.getByText("+10 VP")).toBeTruthy();
  });

  it("sends set_paint_score over the websocket when the toggle is clicked", async () => {
    const gs = makeGameState({
      status: "setup",
      players: [
        makePlayerState({
          detachments: [{ id: "det-gladius", name: "Gladius Task Force", points: 3 }],
          secondaryMode: "",
          ready: false,
          vpPaint: 10,
        }),
        null,
      ],
    });
    useGameStore.getState().setGameState(gs);
    localStorage.setItem("token", "test-token");

    const sentActions: Array<Record<string, unknown>> = [];
    const testLink = ws.link("ws://localhost:8080/ws/game/*");
    worker.use(
      testLink.addEventListener("connection", ({ client }) => {
        client.send(JSON.stringify({ type: "state_update", data: gs }));
        client.addEventListener("message", (ev) => {
          if (typeof ev.data !== "string") return;
          try {
            const parsed = JSON.parse(ev.data);
            if (parsed?.type === "action" && parsed.data) {
              sentActions.push(parsed.data as Record<string, unknown>);
            }
          } catch {
            // ignore non-JSON frames (e.g. the client's pings)
          }
        });
      }),
    );

    await act(async () => {
      renderWithProviders(
        <Routes>
          <Route path="/game/:id/setup" element={<GameSetupPage />} />
        </Routes>,
        { user: mockUser, route: "/game/game-1/setup" },
      );
    });

    await vi.waitFor(() => {
      expect(screen.getByRole("switch")).toBeTruthy();
    });

    const user = userEvent.setup();
    await user.click(screen.getByRole("switch"));

    await vi.waitFor(() => {
      const paintAction = sentActions.find((a) => a.type === "set_paint_score");
      expect(paintAction).toBeTruthy();
      expect(paintAction!.score).toBe(0);
    });
  });

  it("shows the side picker once a detachment is selected", async () => {
    const gs = makeGameState({
      status: "setup",
      players: [
        makePlayerState({
          detachments: [{ id: "det-gladius", name: "Gladius Task Force", points: 3 }],
          side: "",
          forceDisposition: "",
          missionId: "",
          missionName: "",
          secondaryMode: "",
          ready: false,
        }),
        null,
      ],
    });
    useGameStore.getState().setGameState(gs);
    localStorage.setItem("token", "test-token");

    const testLink = ws.link("ws://localhost:8080/ws/game/*");
    worker.use(
      testLink.addEventListener("connection", ({ client }) => {
        client.send(JSON.stringify({ type: "state_update", data: gs }));
      }),
    );

    await act(async () => {
      renderWithProviders(
        <Routes>
          <Route path="/game/:id/setup" element={<GameSetupPage />} />
        </Routes>,
        { user: mockUser, route: "/game/game-1/setup" },
      );
    });

    await vi.waitFor(() => {
      expect(screen.getByText("Your Side")).toBeTruthy();
      expect(screen.getByText("Attacker")).toBeTruthy();
      expect(screen.getByText("Defender")).toBeTruthy();
    });

    // Force-disposition section is gated until a side is chosen.
    expect(screen.queryByText("Force Disposition")).toBeNull();
  });

  it("sends select_side and select_force_disposition over the websocket", async () => {
    const gs = makeGameState({
      status: "setup",
      players: [
        makePlayerState({
          detachments: [{ id: "det-gladius", name: "Gladius Task Force", points: 3 }],
          side: "attacker",
          forceDisposition: "",
          missionId: "",
          missionName: "",
          secondaryMode: "",
          ready: false,
        }),
        null,
      ],
    });
    useGameStore.getState().setGameState(gs);
    localStorage.setItem("token", "test-token");

    const sentActions: Array<Record<string, unknown>> = [];
    const testLink = ws.link("ws://localhost:8080/ws/game/*");
    worker.use(
      testLink.addEventListener("connection", ({ client }) => {
        client.send(JSON.stringify({ type: "state_update", data: gs }));
        client.addEventListener("message", (ev) => {
          if (typeof ev.data !== "string") return;
          try {
            const parsed = JSON.parse(ev.data);
            if (parsed?.type === "action" && parsed.data) {
              sentActions.push(parsed.data as Record<string, unknown>);
            }
          } catch {
            // ignore non-JSON frames (e.g. the client's pings)
          }
        });
      }),
    );

    await act(async () => {
      renderWithProviders(
        <Routes>
          <Route path="/game/:id/setup" element={<GameSetupPage />} />
        </Routes>,
        { user: mockUser, route: "/game/game-1/setup" },
      );
    });

    const user = userEvent.setup();

    await vi.waitFor(() => {
      expect(screen.getByText("Defender")).toBeTruthy();
    });
    await user.click(screen.getByText("Defender"));

    await vi.waitFor(() => {
      const sideAction = sentActions.find((a) => a.type === "select_side");
      expect(sideAction).toBeTruthy();
      expect(sideAction!.side).toBe("defender");
    });

    // The disposition picker is shown (side already set to attacker in fixture)
    // and is limited to the dispositions granted by the chosen detachment
    // (det-gladius → take-and-hold only).
    await vi.waitFor(() => {
      expect(screen.getByText("Take and Hold")).toBeTruthy();
    });
    expect(screen.queryByText("Hold the Line")).toBeNull();
    await user.click(screen.getByText("Take and Hold"));

    await vi.waitFor(() => {
      const dispAction = sentActions.find((a) => a.type === "select_force_disposition");
      expect(dispAction).toBeTruthy();
      expect(dispAction!.disposition).toBe("take-and-hold");
      expect(dispAction!.dispositionName).toBe("Take and Hold");
    });
  });

  it("shows the fixed-secondary picker in fixed mode and sends select_fixed_secondaries", async () => {
    const gs = makeGameState({
      status: "setup",
      firstTurnPlayer: 1,
      players: [
        makePlayerState({
          detachments: [{ id: "det-gladius", name: "Gladius Task Force", points: 3 }],
          side: "attacker",
          forceDisposition: "take-and-hold",
          forceDispositionName: "Take and Hold",
          secondaryMode: "fixed",
          fixedSecondaryIds: [],
          ready: false,
        }),
        null,
      ],
    });
    useGameStore.getState().setGameState(gs);
    localStorage.setItem("token", "test-token");

    const sentActions: Array<Record<string, unknown>> = [];
    const testLink = ws.link("ws://localhost:8080/ws/game/*");
    worker.use(
      testLink.addEventListener("connection", ({ client }) => {
        client.send(JSON.stringify({ type: "state_update", data: gs }));
        client.addEventListener("message", (ev) => {
          if (typeof ev.data !== "string") return;
          try {
            const parsed = JSON.parse(ev.data);
            if (parsed?.type === "action" && parsed.data) {
              sentActions.push(parsed.data as Record<string, unknown>);
            }
          } catch {
            // ignore non-JSON frames
          }
        });
      }),
    );

    await act(async () => {
      renderWithProviders(
        <Routes>
          <Route path="/game/:id/setup" element={<GameSetupPage />} />
        </Routes>,
        { user: mockUser, route: "/game/game-1/setup" },
      );
    });

    const user = userEvent.setup();

    await vi.waitFor(() => {
      expect(screen.getByText("Choose 2 fixed secondaries (0/2)")).toBeTruthy();
      expect(screen.getByText("Behind Enemy Lines")).toBeTruthy();
      expect(screen.getByText("Assassination")).toBeTruthy();
    });

    // Fixed players can't ready until they've chosen their set.
    expect(screen.getByText("Ready Up").closest("button")!.hasAttribute("disabled")).toBe(true);

    await user.click(screen.getByText("Behind Enemy Lines"));

    await vi.waitFor(() => {
      const action = sentActions.find((a) => a.type === "select_fixed_secondaries");
      expect(action).toBeTruthy();
      expect(action!.secondaryIds).toEqual(["sec-behind-lines"]);
    });
  });

  it("shows the resolved primary mission and caps once it is assigned", async () => {
    const gs = makeGameState({
      status: "setup",
      vpPerGameCap: 45,
      vpPerRoundCap: 15,
      players: [
        makePlayerState({
          detachments: [{ id: "det-gladius", name: "Gladius Task Force", points: 3 }],
          side: "attacker",
          forceDisposition: "take-and-hold",
          forceDispositionName: "Take and Hold",
          missionId: "battlefield-dominance",
          missionName: "Battlefield Dominance",
          secondaryMode: "",
          ready: false,
        }),
        null,
      ],
    });
    useGameStore.getState().setGameState(gs);
    localStorage.setItem("token", "test-token");

    const testLink = ws.link("ws://localhost:8080/ws/game/*");
    worker.use(
      testLink.addEventListener("connection", ({ client }) => {
        client.send(JSON.stringify({ type: "state_update", data: gs }));
      }),
    );

    await act(async () => {
      renderWithProviders(
        <Routes>
          <Route path="/game/:id/setup" element={<GameSetupPage />} />
        </Routes>,
        { user: mockUser, route: "/game/game-1/setup" },
      );
    });

    await vi.waitFor(() => {
      expect(screen.getByText("Your Primary Mission")).toBeTruthy();
      expect(screen.getByText("Battlefield Dominance")).toBeTruthy();
      expect(screen.getByText("45 VP / Game")).toBeTruthy();
      expect(screen.getByText("15 VP / Round")).toBeTruthy();
    });
  });
});
