import { useEffect, useRef, useCallback, useState } from "react";
import { ClientMessage, ServerMessage } from "../types/ws";

const WS_URL = import.meta.env.VITE_WS_URL || "ws://localhost:8080";

interface UseWebSocketOptions {
  gameId: string;
  token: string;
  onMessage: (msg: ServerMessage) => void;
  onReconnect?: () => void;
}

export function useWebSocket({ gameId, token, onMessage, onReconnect }: UseWebSocketOptions) {
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimer = useRef<number | undefined>(undefined);
  const reconnectDelay = useRef(1000);
  const pingInterval = useRef<number | undefined>(undefined);
  const mountedRef = useRef(false);
  const hasConnectedRef = useRef(false);
  const onReconnectRef = useRef(onReconnect);
  const [connected, setConnected] = useState(false);
  const [reconnecting, setReconnecting] = useState(false);

  useEffect(() => {
    onReconnectRef.current = onReconnect;
  }, [onReconnect]);

  const connect = useCallback(() => {
    if (!mountedRef.current) return;
    if (wsRef.current?.readyState === WebSocket.OPEN) return;

    const ws = new WebSocket(`${WS_URL}/ws/game/${gameId}?token=${token}`);
    wsRef.current = ws;

    ws.onopen = () => {
      setConnected(true);
      setReconnecting(false);
      reconnectDelay.current = 1000;

      if (hasConnectedRef.current) {
        onReconnectRef.current?.();
      }
      hasConnectedRef.current = true;

      // Start ping interval
      pingInterval.current = window.setInterval(() => {
        if (ws.readyState === WebSocket.OPEN) {
          ws.send(JSON.stringify({ type: "ping" }));
        }
      }, 30000);
    };

    ws.onmessage = (event) => {
      try {
        const msg: ServerMessage = JSON.parse(event.data);
        onMessage(msg);
      } catch {
        console.error("Failed to parse WS message");
      }
    };

    ws.onclose = () => {
      setConnected(false);
      if (pingInterval.current) {
        clearInterval(pingInterval.current);
      }

      // Only reconnect if still mounted
      if (mountedRef.current) {
        if (hasConnectedRef.current) {
          setReconnecting(true);
        }
        reconnectTimer.current = window.setTimeout(() => {
          reconnectDelay.current = Math.min(reconnectDelay.current * 2, 30000);
          connect();
        }, reconnectDelay.current);
      }
    };

    ws.onerror = () => {
      ws.close();
    };
  }, [gameId, token, onMessage]);

  useEffect(() => {
    mountedRef.current = true;
    connect();

    return () => {
      mountedRef.current = false;
      if (reconnectTimer.current) clearTimeout(reconnectTimer.current);
      if (pingInterval.current) clearInterval(pingInterval.current);
      wsRef.current?.close();
    };
  }, [connect]);

  const sendMessage = useCallback((msg: ClientMessage) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(msg));
    }
  }, []);

  const sendAction = useCallback(
    (type: string, data?: Record<string, unknown>) => {
      sendMessage({ type: "action", data: { type, ...data } });
    },
    [sendMessage],
  );

  return { connected, reconnecting, sendAction, sendMessage };
}
