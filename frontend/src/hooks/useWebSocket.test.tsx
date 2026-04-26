import { render, screen, act } from "@testing-library/react";
import { useWebSocket } from "./useWebSocket";
import { ServerMessage } from "../types/ws";
import { ws } from "msw";
import { worker } from "../mocks/browser";

function TestComponent({
  gameId,
  token,
  onMessage,
  onReconnect,
}: {
  gameId: string;
  token: string;
  onMessage: (msg: ServerMessage) => void;
  onReconnect?: () => void;
}) {
  const { connected, reconnecting, sendAction } = useWebSocket({
    gameId,
    token,
    onMessage,
    onReconnect,
  });
  return (
    <div>
      <span data-testid="connected">{connected ? "true" : "false"}</span>
      <span data-testid="reconnecting">{reconnecting ? "true" : "false"}</span>
      <button onClick={() => sendAction("advance_phase")}>Send</button>
    </div>
  );
}

describe("useWebSocket", () => {
  let onMessage: (msg: ServerMessage) => void;

  beforeEach(() => {
    onMessage = vi.fn();
  });

  it("connects to WebSocket server", async () => {
    const testLink = ws.link("ws://localhost:8080/ws/game/*");
    worker.use(
      testLink.addEventListener("connection", ({ client }) => {
        client.send(JSON.stringify({ type: "pong", data: null }));
      }),
    );

    await act(async () => {
      render(<TestComponent gameId="game-1" token="tok" onMessage={onMessage} />);
    });

    await vi.waitFor(() => {
      expect(screen.getByTestId("connected").textContent).toBe("true");
    });
  });

  it("receives and routes messages through onMessage", async () => {
    const stateMsg: ServerMessage = {
      type: "state_update",
      data: { gameId: "game-1" },
    };

    const testLink = ws.link("ws://localhost:8080/ws/game/*");
    worker.use(
      testLink.addEventListener("connection", ({ client }) => {
        client.send(JSON.stringify(stateMsg));
      }),
    );

    await act(async () => {
      render(<TestComponent gameId="game-1" token="tok" onMessage={onMessage} />);
    });

    await vi.waitFor(() => {
      expect(onMessage).toHaveBeenCalledWith(stateMsg);
    });
  });

  it("calls onReconnect after the socket re-establishes", async () => {
    let connectionCount = 0;
    const testLink = ws.link("ws://localhost:8080/ws/game/*");
    worker.use(
      testLink.addEventListener("connection", ({ client }) => {
        connectionCount += 1;
        if (connectionCount === 1) {
          // Drop the first connection to force a reconnect.
          setTimeout(() => client.close(), 10);
        }
      }),
    );

    const onReconnect = vi.fn();

    await act(async () => {
      render(
        <TestComponent
          gameId="game-1"
          token="tok"
          onMessage={onMessage}
          onReconnect={onReconnect}
        />,
      );
    });

    // Wait for the first connection to flip connected → true, then close.
    await vi.waitFor(() => {
      expect(connectionCount).toBe(1);
    });

    // Reconnect delay is 1s; allow generous slack.
    await vi.waitFor(
      () => {
        expect(onReconnect).toHaveBeenCalledTimes(1);
        expect(screen.getByTestId("reconnecting").textContent).toBe("false");
        expect(screen.getByTestId("connected").textContent).toBe("true");
      },
      { timeout: 5000 },
    );
  });

  it("sends action messages over WebSocket", async () => {
    const sentMessages: string[] = [];

    const testLink = ws.link("ws://localhost:8080/ws/game/*");
    worker.use(
      testLink.addEventListener("connection", ({ client }) => {
        client.addEventListener("message", (event) => {
          sentMessages.push(typeof event.data === "string" ? event.data : "");
        });
        client.send(JSON.stringify({ type: "pong", data: null }));
      }),
    );

    await act(async () => {
      render(<TestComponent gameId="game-1" token="tok" onMessage={onMessage} />);
    });

    await vi.waitFor(() => {
      expect(screen.getByTestId("connected").textContent).toBe("true");
    });

    await act(async () => {
      screen.getByRole("button", { name: "Send" }).click();
    });

    await vi.waitFor(() => {
      const actionMsg = sentMessages.find((m) => {
        try {
          const parsed = JSON.parse(m);
          return parsed.type === "action";
        } catch {
          return false;
        }
      });
      expect(actionMsg).toBeDefined();
      const parsed = JSON.parse(actionMsg!);
      expect(parsed.data.type).toBe("advance_phase");
    });
  });
});
