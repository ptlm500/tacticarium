import { render, act } from "@testing-library/react";
import { useGameConnection } from "./useGameState";
import { useGameStore } from "../stores/gameStore";
import { makeGameState, mockEvent } from "../test/fixtures";
import { ws } from "msw";
import { worker } from "../mocks/browser";

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

    const testLink = ws.link("ws://localhost:8080/ws/game/*");
    worker.use(
      testLink.addEventListener("connection", ({ client }) => {
        client.send(JSON.stringify({ type: "state_update", data: gs }));
      }),
    );

    await act(async () => {
      render(<TestComponent gameId="game-1" token="tok" />);
    });

    await vi.waitFor(() => {
      expect(useGameStore.getState().gameState).toEqual(gs);
    });
  });

  it("routes event messages to the store", async () => {
    const testLink = ws.link("ws://localhost:8080/ws/game/*");
    worker.use(
      testLink.addEventListener("connection", ({ client }) => {
        client.send(JSON.stringify({ type: "event", data: mockEvent }));
      }),
    );

    await act(async () => {
      render(<TestComponent gameId="game-1" token="tok" />);
    });

    await vi.waitFor(() => {
      expect(useGameStore.getState().events).toHaveLength(1);
      expect(useGameStore.getState().events[0].eventType).toBe("phase_advanced");
    });
  });

  it("routes error messages to the store and auto-clears", async () => {
    const testLink = ws.link("ws://localhost:8080/ws/game/*");
    worker.use(
      testLink.addEventListener("connection", ({ client }) => {
        client.send(JSON.stringify({ type: "error", data: { message: "Bad move" } }));
      }),
    );

    await act(async () => {
      render(<TestComponent gameId="game-1" token="tok" />);
    });

    await vi.waitFor(() => {
      expect(useGameStore.getState().error).toBe("Bad move");
    });
  });

  it("routes player_connected/disconnected messages", async () => {
    const testLink = ws.link("ws://localhost:8080/ws/game/*");
    worker.use(
      testLink.addEventListener("connection", ({ client }) => {
        client.send(
          JSON.stringify({
            type: "player_connected",
            data: { playerNumber: 2, username: "Opponent" },
          }),
        );
      }),
    );

    await act(async () => {
      render(<TestComponent gameId="game-1" token="tok" />);
    });

    await vi.waitFor(() => {
      expect(useGameStore.getState().opponentConnected).toBe(true);
    });
  });
});
