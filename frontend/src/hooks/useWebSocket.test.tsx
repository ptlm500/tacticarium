import { useWebSocket } from "./useWebSocket";
import { ServerMessage } from "../types/ws";
import { gameWs } from "../mocks/handlers/ws";

function TestComponent({
  gameId,
  token,
  onMessage,
}: {
  gameId: string;
  token: string;
  onMessage: (msg: ServerMessage) => void;
}) {
  const { connected, sendAction } = useWebSocket({ gameId, token, onMessage });
  return (
    <div>
      <span data-testid="connected">{connected ? "true" : "false"}</span>
      <button onClick={() => sendAction("advance_phase")}>Send</button>
    </div>
  );
}

describe("useWebSocket", () => {
  let onMessage: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    onMessage = vi.fn();
  });

  it("connects to WebSocket server", async () => {
    const serverMessagePromise = new Promise<void>((resolve) => {
      gameWs.addEventListener("connection", ({ client }) => {
        client.send(JSON.stringify({ type: "pong", data: null }));
        resolve();
      });
    });

    render(<TestComponent gameId="game-1" token="tok" onMessage={onMessage} />);

    await serverMessagePromise;

    await vi.waitFor(() => {
      expect(screen.getByTestId("connected").textContent).toBe("true");
    });
  });

  it("receives and routes messages through onMessage", async () => {
    const stateMsg: ServerMessage = {
      type: "state_update",
      data: { gameId: "game-1" },
    };

    const connectionPromise = new Promise<void>((resolve) => {
      gameWs.addEventListener("connection", ({ client }) => {
        client.send(JSON.stringify(stateMsg));
        resolve();
      });
    });

    render(<TestComponent gameId="game-1" token="tok" onMessage={onMessage} />);

    await connectionPromise;

    await vi.waitFor(() => {
      expect(onMessage).toHaveBeenCalledWith(stateMsg);
    });
  });

  it("sends action messages over WebSocket", async () => {
    const sentMessages: string[] = [];

    const connectionPromise = new Promise<void>((resolve) => {
      gameWs.addEventListener("connection", ({ client }) => {
        client.addEventListener("message", (event) => {
          sentMessages.push(typeof event.data === "string" ? event.data : "");
        });
        client.send(JSON.stringify({ type: "pong", data: null }));
        resolve();
      });
    });

    render(<TestComponent gameId="game-1" token="tok" onMessage={onMessage} />);

    await connectionPromise;

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
