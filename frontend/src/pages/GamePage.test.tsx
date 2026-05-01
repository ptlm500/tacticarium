import { screen, act } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { renderWithProviders } from "../test/renderWithProviders";
import { GamePage } from "./GamePage";
import { useGameStore } from "../stores/gameStore";
import {
  makeGameState,
  makePlayerState,
  mockUser,
  mockStratagems,
  mockActiveSecondary,
  mockFixedSecondary,
} from "../test/fixtures";
import { ws, http, HttpResponse } from "msw";
import { worker } from "../mocks/browser";
import { Route, Routes } from "react-router-dom";

const API_URL = "http://localhost:8080";

function renderGame(gameStateOverrides?: Parameters<typeof makeGameState>[0]) {
  const gs = makeGameState(gameStateOverrides);
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
      <Route path="/game/:id" element={<GamePage />} />
    </Routes>,
    { user: mockUser, route: "/game/game-1" },
  );
}

describe("GamePage", () => {
  beforeEach(() => {
    useGameStore.getState().reset();
    localStorage.clear();
  });

  it("resets gameStore when unmounted", async () => {
    let result!: ReturnType<typeof renderGame>;
    await act(async () => {
      result = renderGame();
    });

    await vi.waitFor(() => {
      expect(useGameStore.getState().gameState).not.toBeNull();
    });

    await act(async () => {
      result.unmount();
    });

    expect(useGameStore.getState().gameState).toBeNull();
  });

  it("redirects to /game/:id/setup when state is in setup phase for the current game", async () => {
    const gs = makeGameState({ gameId: "game-1", status: "setup" });
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
          <Route path="/game/:id" element={<GamePage />} />
          <Route path="/game/:id/setup" element={<div>SETUP_PAGE_REACHED</div>} />
        </Routes>,
        { user: mockUser, route: "/game/game-1" },
      );
    });

    await vi.waitFor(() => {
      expect(screen.getByText("SETUP_PAGE_REACHED")).toBeTruthy();
    });
  });

  it("does not redirect to /setup when stored state is from a different game", async () => {
    // Stale state from a previously-open setup game.
    useGameStore.getState().setGameState(makeGameState({ gameId: "game-OLD", status: "setup" }));
    localStorage.setItem("token", "test-token");

    // The new game we're navigating to is active.
    const newGameState = makeGameState({ gameId: "game-NEW", status: "active" });
    const testLink = ws.link("ws://localhost:8080/ws/game/*");
    worker.use(
      testLink.addEventListener("connection", ({ client }) => {
        client.send(JSON.stringify({ type: "state_update", data: newGameState }));
      }),
    );

    await act(async () => {
      renderWithProviders(
        <Routes>
          <Route path="/game/:id" element={<GamePage />} />
          <Route path="/game/:id/setup" element={<div>WRONG_SETUP_PAGE</div>} />
        </Routes>,
        { user: mockUser, route: "/game/game-NEW" },
      );
    });

    await vi.waitFor(() => {
      // Active-game UI renders (turn banner appears once new state arrives).
      expect(screen.getByText(/Battle Round/)).toBeTruthy();
    });
    expect(screen.queryByText("WRONG_SETUP_PAGE")).toBeNull();
  });

  it("shows Victory when game is completed and player won", async () => {
    await act(async () => {
      renderGame({
        status: "completed",
        winnerId: "user-1",
      });
    });

    await vi.waitFor(() => {
      expect(screen.getByText("Victory!")).toBeTruthy();
    });
  });

  it("shows Defeat when game is completed and player lost", async () => {
    await act(async () => {
      renderGame({
        status: "completed",
        winnerId: "user-2",
      });
    });

    await vi.waitFor(() => {
      expect(screen.getByText("Defeat")).toBeTruthy();
    });
  });

  it("shows Draw when game is completed with no winner", async () => {
    await act(async () => {
      renderGame({
        status: "completed",
        winnerId: undefined,
      });
    });

    await vi.waitFor(() => {
      expect(screen.getByText("Draw")).toBeTruthy();
    });
  });

  it("shows turn banner indicating player's turn", async () => {
    await act(async () => {
      renderGame({
        activePlayer: 1,
        currentRound: 2,
        currentPhase: "shooting",
      });
    });

    await vi.waitFor(() => {
      expect(screen.getByText(/Your Turn/)).toBeTruthy();
      expect(screen.getByText(/Battle Round 2/)).toBeTruthy();
      expect(screen.getByText(/Shooting Phase/)).toBeTruthy();
    });
  });

  it("shows opponent's turn in banner", async () => {
    await act(async () => {
      renderGame({
        activePlayer: 2,
        currentPhase: "movement",
      });
    });

    await vi.waitFor(() => {
      expect(screen.getByText(/Opponent's Turn/)).toBeTruthy();
    });
  });

  it("shows Advance Phase button only on player's turn", async () => {
    await act(async () => {
      renderGame({ activePlayer: 1 });
    });

    await vi.waitFor(() => {
      expect(screen.getByText("Advance Phase")).toBeTruthy();
    });
  });

  it("hides Advance Phase button on opponent's turn", async () => {
    await act(async () => {
      renderGame({ activePlayer: 2 });
    });

    await vi.waitFor(() => {
      expect(screen.getByText("Concede")).toBeTruthy();
      expect(screen.queryByText("Advance Phase")).toBeNull();
    });
  });

  describe("Revert Phase button", () => {
    it("is shown on the active player's turn", async () => {
      await act(async () => {
        renderGame({ activePlayer: 1, currentPhase: "movement" });
      });

      await vi.waitFor(() => {
        expect(screen.getByText("← Revert")).toBeTruthy();
      });
    });

    it("is hidden on the opponent's turn", async () => {
      await act(async () => {
        renderGame({ activePlayer: 2, currentPhase: "movement" });
      });

      await vi.waitFor(() => {
        expect(screen.getByText("Concede")).toBeTruthy();
      });
      expect(screen.queryByText("← Revert")).toBeNull();
    });

    it("opens the confirmation modal on click", async () => {
      await act(async () => {
        renderGame({ activePlayer: 1, currentPhase: "movement" });
      });

      await vi.waitFor(() => {
        expect(screen.getByText("← Revert")).toBeTruthy();
      });

      const user = userEvent.setup();
      await user.click(screen.getByText("← Revert"));

      await vi.waitFor(() => {
        // Modal title ("Revert Phase") appears; confirm button label is "Revert"
        expect(screen.getByText("Revert Phase")).toBeTruthy();
        expect(screen.getByText(/both players lose the 1 CP/i)).toBeTruthy();
      });
    });

    it("sends revert_phase action when confirmed", async () => {
      const wsMessages: string[] = [];
      const testLink = ws.link("ws://localhost:8080/ws/game/*");
      const gs = makeGameState({ activePlayer: 1, currentPhase: "shooting" });

      worker.use(
        testLink.addEventListener("connection", ({ client }) => {
          client.addEventListener("message", (event) => {
            wsMessages.push(typeof event.data === "string" ? event.data : "");
          });
          client.send(JSON.stringify({ type: "state_update", data: gs }));
        }),
      );

      useGameStore.getState().setGameState(gs);
      localStorage.setItem("token", "test-token");

      await act(async () => {
        renderWithProviders(
          <Routes>
            <Route path="/game/:id" element={<GamePage />} />
          </Routes>,
          { user: mockUser, route: "/game/game-1" },
        );
      });

      const user = userEvent.setup();

      await vi.waitFor(() => {
        expect(screen.getByText("← Revert")).toBeTruthy();
      });

      await user.click(screen.getByText("← Revert"));

      // The confirm button inside the modal is labelled "Revert"
      await vi.waitFor(() => {
        const confirmBtn = screen.getAllByRole("button").find((b) => b.textContent === "Revert");
        expect(confirmBtn).toBeTruthy();
      });
      const confirmBtn = screen.getAllByRole("button").find((b) => b.textContent === "Revert");
      await user.click(confirmBtn!);

      await vi.waitFor(() => {
        const msg = wsMessages.find((m) => m.includes("revert_phase"));
        expect(msg).toBeTruthy();
        const parsed = JSON.parse(msg!);
        expect(parsed.type).toBe("action");
        expect(parsed.data.type).toBe("revert_phase");
      });
    });

    it("does not send revert_phase when cancelled", async () => {
      const wsMessages: string[] = [];
      const testLink = ws.link("ws://localhost:8080/ws/game/*");
      const gs = makeGameState({ activePlayer: 1, currentPhase: "shooting" });

      worker.use(
        testLink.addEventListener("connection", ({ client }) => {
          client.addEventListener("message", (event) => {
            wsMessages.push(typeof event.data === "string" ? event.data : "");
          });
          client.send(JSON.stringify({ type: "state_update", data: gs }));
        }),
      );

      useGameStore.getState().setGameState(gs);
      localStorage.setItem("token", "test-token");

      await act(async () => {
        renderWithProviders(
          <Routes>
            <Route path="/game/:id" element={<GamePage />} />
          </Routes>,
          { user: mockUser, route: "/game/game-1" },
        );
      });

      const user = userEvent.setup();

      await vi.waitFor(() => {
        expect(screen.getByText("← Revert")).toBeTruthy();
      });

      await user.click(screen.getByText("← Revert"));
      await vi.waitFor(() => {
        expect(screen.getByText("Cancel")).toBeTruthy();
      });
      await user.click(screen.getByText("Cancel"));

      // Modal should close; no revert_phase message sent
      expect(wsMessages.find((m) => m.includes("revert_phase"))).toBeUndefined();
    });
  });

  describe("opponent active secondaries", () => {
    it("renders opponent's active tactical secondaries with name, description, and max VP", async () => {
      await act(async () => {
        renderGame({
          players: [
            makePlayerState(),
            makePlayerState({
              userId: "user-2",
              username: "Opponent",
              playerNumber: 2,
              secondaryMode: "tactical",
              activeSecondaries: [mockActiveSecondary],
            }),
          ],
        });
      });

      await vi.waitFor(() => {
        expect(screen.getByText("Opponent's Active Secondaries (Tactical)")).toBeTruthy();
        expect(screen.getByText("Score VP for units in enemy deployment zone")).toBeTruthy();
        // Opponent name appears in the header; the secondary name appears in the opponent block.
        expect(screen.getByText("Behind Enemy Lines")).toBeTruthy();
        expect(screen.getByText("5 VP max")).toBeTruthy();
      });
    });

    it("renders opponent's active fixed secondaries", async () => {
      await act(async () => {
        renderGame({
          players: [
            makePlayerState(),
            makePlayerState({
              userId: "user-2",
              username: "Opponent",
              playerNumber: 2,
              secondaryMode: "fixed",
              activeSecondaries: [mockFixedSecondary],
            }),
          ],
        });
      });

      await vi.waitFor(() => {
        expect(screen.getByText("Opponent's Active Secondaries (Fixed)")).toBeTruthy();
        expect(screen.getByText("Assassination")).toBeTruthy();
      });
    });

    it("does not render an opponent secondaries section when they have none", async () => {
      await act(async () => {
        renderGame({
          players: [
            makePlayerState(),
            makePlayerState({
              userId: "user-2",
              username: "Opponent",
              playerNumber: 2,
              factionName: "Chaos Space Marines",
              secondaryMode: "tactical",
              activeSecondaries: [],
            }),
          ],
        });
      });

      await vi.waitFor(() => {
        // Opponent block (with faction) is rendered, but no Active Secondaries header.
        expect(screen.getByText(/Chaos Space Marines/)).toBeTruthy();
      });
      expect(screen.queryByText(/Active Secondaries \(/)).toBeNull();
    });

    it("does not render any action buttons on opponent's secondaries", async () => {
      await act(async () => {
        renderGame({
          activePlayer: 1,
          currentPhase: "command",
          players: [
            makePlayerState(),
            makePlayerState({
              userId: "user-2",
              username: "Opponent",
              playerNumber: 2,
              secondaryMode: "tactical",
              activeSecondaries: [mockActiveSecondary],
            }),
          ],
        });
      });

      await vi.waitFor(() => {
        expect(screen.getByText("Opponent's Active Secondaries (Tactical)")).toBeTruthy();
      });

      // The opponent's secondary card should expose no Achieve / Discard / New Orders buttons.
      // The user has no active secondaries of their own, so these labels should not appear at all.
      expect(screen.queryByText("New Orders")).toBeNull();
      expect(screen.queryByText("Discard")).toBeNull();
      expect(screen.queryByText(/Discard \+1CP/)).toBeNull();
    });
  });

  it("filters stratagems by phase, turn, and detachment", async () => {
    // Override the stratagems API to return our mock data
    worker.use(
      http.get(`${API_URL}/api/factions/:factionId/stratagems`, () => {
        return HttpResponse.json(mockStratagems);
      }),
    );

    await act(async () => {
      renderGame({
        activePlayer: 1,
        currentPhase: "shooting",
      });
    });

    // Expand stratagems panel
    await vi.waitFor(() => {
      const stratagemBtn = screen.getByText(/Stratagems/);
      expect(stratagemBtn).toBeTruthy();
    });

    const user = userEvent.setup();
    await user.click(screen.getByText(/Stratagems/));

    // "Command Re-roll" has phase "Any phase" and turn "Either player's turn" — should show
    // "Storm of Fire" has phase "Shooting phase" and turn "Your turn" — should show (it's our turn)
    // "Heroic Intervention" has phase "Charge phase" and turn "Opponent's turn" — should NOT show
    // "Banner of Defiance" is a Challenger stratagem — should NOT show (filtered by type)
    await vi.waitFor(() => {
      expect(screen.getByText("Command Re-roll")).toBeTruthy();
      expect(screen.getByText("Storm of Fire")).toBeTruthy();
    });
    expect(screen.queryByText("Heroic Intervention")).toBeNull();
    expect(screen.queryByText("Banner of Defiance")).toBeNull();
  });

  it("updates UI when WebSocket sends state_update", async () => {
    await act(async () => {
      renderGame({ currentRound: 1, currentPhase: "command" });
    });

    await vi.waitFor(() => {
      expect(screen.getByText(/Battle Round 1/)).toBeTruthy();
    });

    // Simulate a WebSocket state update
    const updatedGs = makeGameState({ currentRound: 2, currentPhase: "movement" });
    useGameStore.getState().setGameState(updatedGs);

    await vi.waitFor(() => {
      expect(screen.getByText(/Battle Round 2/)).toBeTruthy();
      expect(screen.getByText(/Movement Phase/)).toBeTruthy();
    });
  });

  it("shows scoring prompt when advancing from fight phase", async () => {
    // Need missions API to return scoring rules
    worker.use(
      http.get(`${API_URL}/api/mission-packs/:packId/missions`, () => {
        return HttpResponse.json([
          {
            id: "mission-1",
            missionPackId: "chapter-approved-2025-26",
            name: "Supply Drop",
            lore: "",
            description: "",
            scoringRules: [{ label: "2 objectives", vp: 5 }],
            scoringTiming: "end_of_command_phase",
          },
        ]);
      }),
    );

    await act(async () => {
      renderGame({
        activePlayer: 1,
        currentPhase: "fight",
        currentRound: 3,
        currentTurn: 1,
        players: [
          makePlayerState({
            secondaryMode: "tactical",
            activeSecondaries: [mockActiveSecondary],
            tacticalDeck: [],
          }),
          makePlayerState({
            userId: "user-2",
            username: "Opponent",
            playerNumber: 2,
          }),
        ],
      });
    });

    const user = userEvent.setup();

    await vi.waitFor(() => {
      expect(screen.getByText("Advance Phase")).toBeTruthy();
    });

    await user.click(screen.getByText("Advance Phase"));

    // Fight phase should show secondary scoring prompt
    await vi.waitFor(() => {
      expect(screen.getByText("Scoring Reminder")).toBeTruthy();
    });
  });

  describe("primary scoring prompts", () => {
    function useMissionMock(
      overrides: Partial<{
        scoringRules: Array<{
          label: string;
          vp: number;
          minRound?: number;
          scoringTiming?: string;
        }>;
        scoringTiming: string;
        name: string;
      }> = {},
    ) {
      worker.use(
        http.get(`${API_URL}/api/mission-packs/:packId/missions`, () => {
          return HttpResponse.json([
            {
              id: "mission-1",
              missionPackId: "chapter-approved-2025-26",
              name: overrides.name ?? "Supply Drop",
              lore: "",
              description: "",
              scoringRules: overrides.scoringRules ?? [
                { label: "2 objectives", vp: 5, minRound: 2 },
                { label: "3+ objectives", vp: 10, minRound: 2 },
              ],
              scoringTiming: overrides.scoringTiming ?? "end_of_command_phase",
            },
          ]);
        }),
      );
    }

    it("shows primary scoring prompt when advancing out of command phase in BR2+", async () => {
      useMissionMock();

      await act(async () => {
        renderGame({
          activePlayer: 1,
          currentPhase: "command",
          currentRound: 2,
          currentTurn: 1,
        });
      });

      const user = userEvent.setup();

      // Wait for mission data to load before clicking Advance Phase
      await vi.waitFor(() => {
        expect(screen.getByText("Quick Score")).toBeTruthy();
      });

      await user.click(screen.getByText("Advance Phase"));

      await vi.waitFor(() => {
        expect(screen.getByText("Scoring Reminder")).toBeTruthy();
        expect(screen.getByText("Score Primary — Supply Drop")).toBeTruthy();
        // Scoring buttons appear in both Quick Score and the modal — check both exist
        expect(screen.getAllByText("2 objectives (+5)")).toHaveLength(2);
        expect(screen.getAllByText("3+ objectives (+10)")).toHaveLength(2);
      });
    });

    it("does not show primary scoring prompt in BR1", async () => {
      useMissionMock();

      await act(async () => {
        renderGame({
          activePlayer: 1,
          currentPhase: "command",
          currentRound: 1,
          currentTurn: 1,
        });
      });

      const user = userEvent.setup();

      // Wait for mission data to load
      await vi.waitFor(() => {
        expect(screen.getByText("Quick Score")).toBeTruthy();
      });

      await user.click(screen.getByText("Advance Phase"));

      // BR1 command phase should not trigger primary scoring — no prompt shown
      // (tactical draw might show if applicable, but not primary)
      await vi.waitFor(() => {
        expect(screen.queryByText("Score Primary — Supply Drop")).toBeNull();
      });
    });

    it("shows primary and secondary prompts together for end_of_turn timing", async () => {
      useMissionMock({
        name: "Terraform",
        scoringRules: [{ label: "Terraformed marker", vp: 1, scoringTiming: "end_of_turn" }],
        scoringTiming: "end_of_turn",
      });

      await act(async () => {
        renderGame({
          activePlayer: 1,
          currentPhase: "fight",
          currentRound: 3,
          currentTurn: 1,
          players: [
            makePlayerState({
              secondaryMode: "tactical",
              activeSecondaries: [mockActiveSecondary],
              tacticalDeck: [],
            }),
            makePlayerState({
              userId: "user-2",
              username: "Opponent",
              playerNumber: 2,
            }),
          ],
        });
      });

      const user = userEvent.setup();

      // Wait for mission data to load
      await vi.waitFor(() => {
        expect(screen.getByText("Quick Score")).toBeTruthy();
      });

      await user.click(screen.getByText("Advance Phase"));

      // Both primary and secondary prompts should appear together
      await vi.waitFor(() => {
        expect(screen.getByText("Scoring Reminder")).toBeTruthy();
        expect(screen.getByText(/Score Primary — Terraform/)).toBeTruthy();
        // Scoring button appears in both Quick Score and the modal
        expect(screen.getAllByText("Terraformed marker (+1)").length).toBeGreaterThanOrEqual(1);
        expect(screen.getByText("Score / Discard Secondaries")).toBeTruthy();
        // "Behind Enemy Lines" appears in both SecondaryPanel and the modal
        expect(screen.getAllByText("Behind Enemy Lines").length).toBeGreaterThanOrEqual(2);
      });
    });

    it("scores primary VP when clicking a scoring button in the prompt", async () => {
      useMissionMock();

      const wsMessages: string[] = [];
      const testLink = ws.link("ws://localhost:8080/ws/game/*");
      const gs = makeGameState({
        activePlayer: 1,
        currentPhase: "command",
        currentRound: 3,
        currentTurn: 1,
      });

      worker.use(
        testLink.addEventListener("connection", ({ client }) => {
          client.addEventListener("message", (event) => {
            wsMessages.push(typeof event.data === "string" ? event.data : "");
          });
          client.send(JSON.stringify({ type: "state_update", data: gs }));
        }),
      );

      useGameStore.getState().setGameState(gs);
      localStorage.setItem("token", "test-token");

      await act(async () => {
        renderWithProviders(
          <Routes>
            <Route path="/game/:id" element={<GamePage />} />
          </Routes>,
          { user: mockUser, route: "/game/game-1" },
        );
      });

      const user = userEvent.setup();

      // Wait for mission data to load
      await vi.waitFor(() => {
        expect(screen.getByText("Quick Score")).toBeTruthy();
      });

      await user.click(screen.getByText("Advance Phase"));

      await vi.waitFor(() => {
        expect(screen.getByText("Score Primary — Supply Drop")).toBeTruthy();
      });

      // Click the scoring button inside the modal (data-testid set), not the Quick Score one
      const modalButtons = screen.getAllByTestId("scoring-prompt-primary-btn");
      const modalButton = modalButtons.find((btn) => btn.textContent?.includes("3+ objectives"));
      await user.click(modalButton!);

      // Verify score_vp action was sent via WebSocket with scoringSlot
      await vi.waitFor(() => {
        const scoreMsg = wsMessages.find((m) => m.includes("score_vp"));
        expect(scoreMsg).toBeTruthy();
        const parsed = JSON.parse(scoreMsg!);
        expect(parsed.data.category).toBe("primary");
        expect(parsed.data.delta).toBe(10);
        expect(parsed.data.scoringSlot).toBe("end_of_command_phase");
      });
    });
  });

  describe("end_of_opponent_turn scoring", () => {
    const sabotage = {
      ...mockActiveSecondary,
      id: "sabotage-1",
      name: "Sabotage",
      scoringTiming: "end_of_opponent_turn" as const,
    };

    it("blocks active player's advance with opponent-pending reminder", async () => {
      worker.use(
        http.get(`${API_URL}/api/mission-packs/:packId/missions`, () =>
          HttpResponse.json([
            {
              id: "mission-1",
              missionPackId: "chapter-approved-2025-26",
              name: "Supply Drop",
              lore: "",
              description: "",
              scoringRules: [],
              scoringTiming: "end_of_command_phase",
            },
          ]),
        ),
      );

      await act(async () => {
        renderGame({
          activePlayer: 1,
          currentPhase: "fight",
          currentRound: 3,
          currentTurn: 1,
          players: [
            makePlayerState({
              secondaryMode: "tactical",
              activeSecondaries: [mockActiveSecondary],
              tacticalDeck: [],
            }),
            makePlayerState({
              userId: "user-2",
              username: "Opponent",
              playerNumber: 2,
              secondaryMode: "tactical",
              activeSecondaries: [sabotage],
            }),
          ],
        });
      });

      const user = userEvent.setup();
      await vi.waitFor(() => {
        expect(screen.getByText("Advance Phase")).toBeTruthy();
      });
      await user.click(screen.getByText("Advance Phase"));

      // The active player's modal includes a reminder block listing the
      // opponent's end_of_opponent_turn secondaries.
      await vi.waitFor(() => {
        expect(screen.getByText(/Wait for Opponent to score/)).toBeTruthy();
        expect(screen.getByTestId("opponent-pending-secondary")).toBeTruthy();
      });
    });

    it("excludes end_of_opponent_turn secondaries from the own-turn fight-phase prompt", async () => {
      worker.use(
        http.get(`${API_URL}/api/mission-packs/:packId/missions`, () => {
          return HttpResponse.json([
            {
              id: "mission-1",
              missionPackId: "chapter-approved-2025-26",
              name: "Supply Drop",
              lore: "",
              description: "",
              scoringRules: [],
              scoringTiming: "end_of_command_phase",
            },
          ]);
        }),
      );

      await act(async () => {
        renderGame({
          activePlayer: 1,
          currentPhase: "fight",
          currentRound: 3,
          currentTurn: 1,
          players: [
            makePlayerState({
              secondaryMode: "tactical",
              activeSecondaries: [mockActiveSecondary, sabotage],
              tacticalDeck: [],
            }),
            makePlayerState({
              userId: "user-2",
              username: "Opponent",
              playerNumber: 2,
            }),
          ],
        });
      });

      const user = userEvent.setup();
      await vi.waitFor(() => {
        expect(screen.getByText("Advance Phase")).toBeTruthy();
      });

      await user.click(screen.getByText("Advance Phase"));

      // The standard scoring prompt opens; Behind Enemy Lines (own-turn) shows in
      // the modal section, but Sabotage (opponent-turn) does NOT appear inside it.
      await vi.waitFor(() => {
        expect(screen.getByText("Score / Discard Secondaries")).toBeTruthy();
      });
      // Behind Enemy Lines also appears in the SecondaryPanel — the modal version
      // is the second occurrence.
      const behindAll = screen.getAllByText("Behind Enemy Lines");
      expect(behindAll.length).toBeGreaterThanOrEqual(2);
      // Sabotage appears in the SecondaryPanel (1 occurrence) but NOT inside the
      // modal — so total occurrences should be exactly 1.
      expect(screen.getAllByText("Sabotage")).toHaveLength(1);
    });

    it("fires reactive prompt when opponent's Fight phase ends", async () => {
      // Start with opponent (player 2) in fight phase.
      const initial = makeGameState({
        activePlayer: 2,
        currentPhase: "fight",
        currentRound: 2,
        currentTurn: 2,
        players: [
          makePlayerState({
            secondaryMode: "tactical",
            activeSecondaries: [sabotage],
          }),
          makePlayerState({
            userId: "user-2",
            username: "Opponent",
            playerNumber: 2,
          }),
        ],
      });
      useGameStore.getState().setGameState(initial);
      localStorage.setItem("token", "test-token");

      const testLink = ws.link("ws://localhost:8080/ws/game/*");
      worker.use(
        testLink.addEventListener("connection", ({ client }) => {
          client.send(JSON.stringify({ type: "state_update", data: initial }));
        }),
      );

      await act(async () => {
        renderWithProviders(
          <Routes>
            <Route path="/game/:id" element={<GamePage />} />
          </Routes>,
          { user: mockUser, route: "/game/game-1" },
        );
      });

      await vi.waitFor(() => {
        expect(screen.getByText(/Opponent's Turn/)).toBeTruthy();
      });
      // Initial state observed; reactive prompt should NOT have fired yet.
      expect(screen.queryByText("Opponent's Turn Ended")).toBeNull();

      // Opponent advances out of fight — turn rolls into round 3, player 1's command.
      const next = makeGameState({
        activePlayer: 1,
        currentPhase: "command",
        currentRound: 3,
        currentTurn: 1,
        players: initial.players,
      });
      await act(async () => {
        useGameStore.getState().setGameState(next);
      });

      await vi.waitFor(() => {
        expect(screen.getByText("Opponent's Turn Ended")).toBeTruthy();
        // Sabotage shows in the modal (plus once in the SecondaryPanel).
        expect(screen.getAllByText("Sabotage").length).toBeGreaterThanOrEqual(2);
      });
    });

    it("does not fire reactive prompt when player has no end_of_opponent_turn secondaries", async () => {
      const initial = makeGameState({
        activePlayer: 2,
        currentPhase: "fight",
        currentRound: 2,
        currentTurn: 2,
        players: [
          makePlayerState({
            secondaryMode: "tactical",
            activeSecondaries: [mockActiveSecondary],
          }),
          makePlayerState({
            userId: "user-2",
            username: "Opponent",
            playerNumber: 2,
          }),
        ],
      });
      useGameStore.getState().setGameState(initial);
      localStorage.setItem("token", "test-token");

      const testLink = ws.link("ws://localhost:8080/ws/game/*");
      worker.use(
        testLink.addEventListener("connection", ({ client }) => {
          client.send(JSON.stringify({ type: "state_update", data: initial }));
        }),
      );

      await act(async () => {
        renderWithProviders(
          <Routes>
            <Route path="/game/:id" element={<GamePage />} />
          </Routes>,
          { user: mockUser, route: "/game/game-1" },
        );
      });

      await vi.waitFor(() => {
        expect(screen.getByText(/Opponent's Turn/)).toBeTruthy();
      });

      const next = makeGameState({
        activePlayer: 1,
        currentPhase: "command",
        currentRound: 3,
        currentTurn: 1,
        players: initial.players,
      });
      await act(async () => {
        useGameStore.getState().setGameState(next);
      });

      // Wait for the new turn banner so we know the transition was processed.
      await vi.waitFor(() => {
        expect(screen.getByText(/Your Turn/)).toBeTruthy();
      });
      expect(screen.queryByText("Opponent's Turn Ended")).toBeNull();
    });
  });

  describe("connection status", () => {
    it("shows the opponent-disconnected indicator when opponentConnected is false", async () => {
      await act(async () => {
        renderGame({
          players: [
            makePlayerState(),
            makePlayerState({
              userId: "user-2",
              username: "Opponent",
              playerNumber: 2,
            }),
          ],
        });
      });

      // After connect, the WS sends state_update; player_disconnected isn't sent,
      // so opponentConnected stays false (its default).
      await vi.waitFor(() => {
        expect(screen.getByLabelText("Opponent disconnected")).toBeTruthy();
      });
    });

    it("hides the opponent-disconnected indicator after player_connected arrives", async () => {
      const gs = makeGameState();
      useGameStore.getState().setGameState(gs);
      localStorage.setItem("token", "test-token");

      const testLink = ws.link("ws://localhost:8080/ws/game/*");
      worker.use(
        testLink.addEventListener("connection", ({ client }) => {
          client.send(JSON.stringify({ type: "state_update", data: gs }));
          client.send(
            JSON.stringify({
              type: "player_connected",
              data: { playerNumber: 2, username: "Opponent" },
            }),
          );
        }),
      );

      await act(async () => {
        renderWithProviders(
          <Routes>
            <Route path="/game/:id" element={<GamePage />} />
          </Routes>,
          { user: mockUser, route: "/game/game-1" },
        );
      });

      await vi.waitFor(() => {
        expect(useGameStore.getState().opponentConnected).toBe(true);
      });
      expect(screen.queryByLabelText("Opponent disconnected")).toBeNull();
    });
  });

  describe("stratagem graceful degradation", () => {
    it("renders the page and shows a fallback when stratagems fail to load", async () => {
      worker.use(
        http.get(`${API_URL}/api/factions/:factionId/stratagems`, () => {
          return HttpResponse.json({ error: "boom" }, { status: 500 });
        }),
      );

      await act(async () => {
        renderGame({ activePlayer: 1, currentPhase: "shooting" });
      });

      // Page still renders the rest of the UI.
      await vi.waitFor(() => {
        expect(screen.getByText(/Battle Round/)).toBeTruthy();
      });

      // Fallback message + retry button appear; the panel button is disabled.
      // The retry chain takes ~1s (250ms + 750ms) before the query enters error state.
      await vi.waitFor(
        () => {
          expect(screen.getByText("Stratagems failed to load.")).toBeTruthy();
          expect(screen.getByText(/Stratagems unavailable/)).toBeTruthy();
          expect(screen.getByText("Retry")).toBeTruthy();
        },
        { timeout: 5000 },
      );
    });
  });

  describe("CP gain cap override", () => {
    function setupCappedGame() {
      const wsMessages: string[] = [];
      const testLink = ws.link("ws://localhost:8080/ws/game/*");
      const gs = makeGameState({
        activePlayer: 1,
        currentPhase: "command",
        players: [
          makePlayerState({ cp: 5, cpGainedThisRound: 1 }),
          makePlayerState({
            userId: "user-2",
            username: "Opponent",
            playerNumber: 2,
          }),
        ],
      });

      worker.use(
        testLink.addEventListener("connection", ({ client }) => {
          client.addEventListener("message", (event) => {
            wsMessages.push(typeof event.data === "string" ? event.data : "");
          });
          client.send(JSON.stringify({ type: "state_update", data: gs }));
        }),
      );

      useGameStore.getState().setGameState(gs);
      localStorage.setItem("token", "test-token");

      return { wsMessages };
    }

    it("opens confirmation modal when increasing CP past the cap", async () => {
      setupCappedGame();

      await act(async () => {
        renderWithProviders(
          <Routes>
            <Route path="/game/:id" element={<GamePage />} />
          </Routes>,
          { user: mockUser, route: "/game/game-1" },
        );
      });

      const user = userEvent.setup();

      await vi.waitFor(() => {
        expect(screen.getByLabelText("Increase CP")).toBeTruthy();
      });

      await user.click(screen.getByLabelText("Increase CP"));

      await vi.waitFor(() => {
        expect(screen.getByText("CP Gain Cap Reached")).toBeTruthy();
      });
    });

    it("sends adjust_cp with force=true when override is confirmed", async () => {
      const { wsMessages } = setupCappedGame();

      await act(async () => {
        renderWithProviders(
          <Routes>
            <Route path="/game/:id" element={<GamePage />} />
          </Routes>,
          { user: mockUser, route: "/game/game-1" },
        );
      });

      const user = userEvent.setup();

      await vi.waitFor(() => {
        expect(screen.getByLabelText("Increase CP")).toBeTruthy();
      });

      await user.click(screen.getByLabelText("Increase CP"));

      await vi.waitFor(() => {
        expect(screen.getByText("Increase CP")).toBeTruthy();
      });
      await user.click(screen.getByText("Increase CP"));

      await vi.waitFor(() => {
        const msg = wsMessages.find((m) => m.includes("adjust_cp"));
        expect(msg).toBeTruthy();
        const parsed = JSON.parse(msg!);
        expect(parsed.type).toBe("action");
        expect(parsed.data.type).toBe("adjust_cp");
        expect(parsed.data.delta).toBe(1);
        expect(parsed.data.force).toBe(true);
      });
    });

    it("does not send adjust_cp when override is cancelled", async () => {
      const { wsMessages } = setupCappedGame();

      await act(async () => {
        renderWithProviders(
          <Routes>
            <Route path="/game/:id" element={<GamePage />} />
          </Routes>,
          { user: mockUser, route: "/game/game-1" },
        );
      });

      const user = userEvent.setup();

      await vi.waitFor(() => {
        expect(screen.getByLabelText("Increase CP")).toBeTruthy();
      });

      await user.click(screen.getByLabelText("Increase CP"));

      await vi.waitFor(() => {
        expect(screen.getByText("CP Gain Cap Reached")).toBeTruthy();
      });
      await user.click(screen.getByText("Cancel"));

      await vi.waitFor(() => {
        expect(screen.queryByText("CP Gain Cap Reached")).toBeNull();
      });
      expect(wsMessages.find((m) => m.includes("adjust_cp"))).toBeUndefined();
    });

    it("sends adjust_cp without force when within cap", async () => {
      const wsMessages: string[] = [];
      const testLink = ws.link("ws://localhost:8080/ws/game/*");
      const gs = makeGameState({
        activePlayer: 1,
        currentPhase: "command",
        players: [
          makePlayerState({ cp: 5, cpGainedThisRound: 0 }),
          makePlayerState({
            userId: "user-2",
            username: "Opponent",
            playerNumber: 2,
          }),
        ],
      });

      worker.use(
        testLink.addEventListener("connection", ({ client }) => {
          client.addEventListener("message", (event) => {
            wsMessages.push(typeof event.data === "string" ? event.data : "");
          });
          client.send(JSON.stringify({ type: "state_update", data: gs }));
        }),
      );

      useGameStore.getState().setGameState(gs);
      localStorage.setItem("token", "test-token");

      await act(async () => {
        renderWithProviders(
          <Routes>
            <Route path="/game/:id" element={<GamePage />} />
          </Routes>,
          { user: mockUser, route: "/game/game-1" },
        );
      });

      const user = userEvent.setup();

      await vi.waitFor(() => {
        expect(screen.getByLabelText("Increase CP")).toBeTruthy();
      });

      await user.click(screen.getByLabelText("Increase CP"));

      await vi.waitFor(() => {
        const msg = wsMessages.find((m) => m.includes("adjust_cp"));
        expect(msg).toBeTruthy();
        const parsed = JSON.parse(msg!);
        expect(parsed.data.delta).toBe(1);
        expect(parsed.data.force).toBeUndefined();
      });
      // Modal should not have appeared
      expect(screen.queryByText("CP Gain Cap Reached")).toBeNull();
    });
  });
});
