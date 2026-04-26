import { useCallback, useEffect, useRef } from "react";
import { toast } from "sonner";
import { useGameStore } from "../stores/gameStore";
import { useWebSocket } from "./useWebSocket";
import { ClientMessage, ServerMessage } from "../types/ws";
import { GameState, GameEvent } from "../types/game";

export function useGameConnection(gameId: string, token: string) {
  const { setGameState, addEvent, setOpponentConnected } = useGameStore();

  const handleMessage = useCallback(
    (msg: ServerMessage) => {
      switch (msg.type) {
        case "state_update":
          setGameState(msg.data as GameState);
          break;
        case "event":
          addEvent(msg.data as GameEvent);
          break;
        case "error": {
          const err = msg.data as { message: string };
          toast.error(err.message);
          break;
        }
        case "player_connected":
          setOpponentConnected(true);
          break;
        case "player_disconnected":
          setOpponentConnected(false);
          break;
      }
    },
    [setGameState, addEvent, setOpponentConnected],
  );

  // Indirection so the stable onReconnect callback can call the latest
  // sendMessage without retriggering the useWebSocket hook every render.
  const sendMessageRef = useRef<((msg: ClientMessage) => void) | null>(null);

  const handleReconnect = useCallback(() => {
    sendMessageRef.current?.({ type: "sync_request" });
  }, []);

  const { connected, reconnecting, sendAction, sendMessage } = useWebSocket({
    gameId,
    token,
    onMessage: handleMessage,
    onReconnect: handleReconnect,
  });

  useEffect(() => {
    sendMessageRef.current = sendMessage;
  }, [sendMessage]);

  return { connected, reconnecting, sendAction };
}
