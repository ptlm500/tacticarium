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
            activeSecondaries: [],
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

      // Click the scoring button inside the modal (indigo-800 bg), not the Quick Score one
      const buttons = screen.getAllByText("3+ objectives (+10)");
      const modalButton = buttons.find((btn) =>
        btn.closest("button")?.classList.contains("bg-indigo-800"),
      );
      await user.click(modalButton!);

      // Verify score_vp action was sent via WebSocket
      await vi.waitFor(() => {
        const scoreMsg = wsMessages.find((m) => m.includes("score_vp"));
        expect(scoreMsg).toBeTruthy();
        const parsed = JSON.parse(scoreMsg!);
        expect(parsed.data.category).toBe("primary");
        expect(parsed.data.delta).toBe(10);
      });
    });
  });
});
