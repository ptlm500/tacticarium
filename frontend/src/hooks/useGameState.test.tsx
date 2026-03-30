import { render } from "@testing-library/react";
import { useGameConnection } from "./useGameState";
import { useGameStore } from "../stores/gameStore";
import { makeGameState, mockEvent } from "../test/fixtures";
import { gameWs } from "../mocks/handlers/ws";

function TestComponent({ gameId, token }: { gameId: string; token: string }) {
  const { connected } = useGameConnection(gameId, token);
  return <span data-testid="connected">{connected ? "true" : "false"}</span>;
}

describe("useGameConnection", () => {
  beforeEach(() => {
    useGameStore.getState().reset();
  });

  it("routes state_update messages to the store", async () => {
    const gs = makeGameState();

    const connectionPromise = new Promise<void>((resolve) => {
      gameWs.addEventListener("connection", ({ client }) => {
        client.send(JSON.stringify({ type: "state_update", data: gs }));
        resolve();
      });
    });

    render(<TestComponent gameId="game-1" token="tok" />);
    await connectionPromise;

    await vi.waitFor(() => {
      expect(useGameStore.getState().gameState).toEqual(gs);
    });
  });

  it("routes event messages to the store", async () => {
    const connectionPromise = new Promise<void>((resolve) => {
      gameWs.addEventListener("connection", ({ client }) => {
        client.send(JSON.stringify({ type: "event", data: mockEvent }));
        resolve();
      });
    });

    render(<TestComponent gameId="game-1" token="tok" />);
    await connectionPromise;

    await vi.waitFor(() => {
      expect(useGameStore.getState().events).toHaveLength(1);
      expect(useGameStore.getState().events[0].eventType).toBe("phase_advanced");
    });
  });

  it("routes error messages to the store and auto-clears", async () => {
    const connectionPromise = new Promise<void>((resolve) => {
      gameWs.addEventListener("connection", ({ client }) => {
        client.send(JSON.stringify({ type: "error", data: { message: "Bad move" } }));
        resolve();
      });
    });

    render(<TestComponent gameId="game-1" token="tok" />);
    await connectionPromise;

    await vi.waitFor(() => {
      expect(useGameStore.getState().error).toBe("Bad move");
    });
  });

  it("routes player_connected/disconnected messages", async () => {
    const connectionPromise = new Promise<void>((resolve) => {
      gameWs.addEventListener("connection", ({ client }) => {
        client.send(
          JSON.stringify({
            type: "player_connected",
            data: { playerNumber: 2, username: "Opponent" },
          }),
        );
        resolve();
      });
    });

    render(<TestComponent gameId="game-1" token="tok" />);
    await connectionPromise;

    await vi.waitFor(() => {
      expect(useGameStore.getState().opponentConnected).toBe(true);
    });
  });
});
