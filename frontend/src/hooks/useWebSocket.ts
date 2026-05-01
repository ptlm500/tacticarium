import { useEffect, useRef, useCallback, useState } from "react";
import { ClientMessage, ServerMessage } from "../types/ws";

const WS_URL = import.meta.env.VITE_WS_URL || "ws://localhost:8080";

const PING_INTERVAL_MS = 20000;
const WATCHDOG_INTERVAL_MS = 5000;
// If we haven't heard anything from the server in this long, the connection
// is presumed dead even if the browser still says OPEN — force a reconnect.
const STALE_THRESHOLD_MS = 45000;
// On tab refocus, anything older than this is suspicious enough to recycle.
const VISIBILITY_STALE_THRESHOLD_MS = 30000;

interface UseWebSocketOptions {
  gameId: string;
  token?: string;
  spectator?: boolean;
  onMessage: (msg: ServerMessage) => void;
  onReconnect?: () => void;
}

export function useWebSocket({
  gameId,
  token,
  spectator = false,
  onMessage,
  onReconnect,
}: UseWebSocketOptions) {
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimer = useRef<number | undefined>(undefined);
  const reconnectDelay = useRef(1000);
  const pingInterval = useRef<number | undefined>(undefined);
  const watchdogInterval = useRef<number | undefined>(undefined);
  const lastMessageAt = useRef(Date.now());
  const mountedRef = useRef(false);
  const hasConnectedRef = useRef(false);
  const onReconnectRef = useRef(onReconnect);
  const [connected, setConnected] = useState(false);
  const [reconnecting, setReconnecting] = useState(false);

  useEffect(() => {
    onReconnectRef.current = onReconnect;
  }, [onReconnect]);

  const clearTimers = useCallback(() => {
    if (pingInterval.current) {
      clearInterval(pingInterval.current);
      pingInterval.current = undefined;
    }
    if (watchdogInterval.current) {
      clearInterval(watchdogInterval.current);
      watchdogInterval.current = undefined;
    }
  }, []);

  const connect = useCallback(() => {
    if (!mountedRef.current) return;
    if (wsRef.current?.readyState === WebSocket.OPEN) return;
    if (wsRef.current?.readyState === WebSocket.CONNECTING) return;

    const url = spectator
      ? `${WS_URL}/ws/game/${gameId}/spectate`
      : `${WS_URL}/ws/game/${gameId}?token=${token}`;
    const ws = new WebSocket(url);
    wsRef.current = ws;

    ws.onopen = () => {
      setConnected(true);
      setReconnecting(false);
      reconnectDelay.current = 1000;
      lastMessageAt.current = Date.now();

      if (hasConnectedRef.current) {
        onReconnectRef.current?.();
      }
      hasConnectedRef.current = true;

      pingInterval.current = window.setInterval(() => {
        if (ws.readyState === WebSocket.OPEN) {
          ws.send(JSON.stringify({ type: "ping" }));
        }
      }, PING_INTERVAL_MS);

      // Watchdog: if no server traffic for STALE_THRESHOLD_MS, the socket is
      // likely a zombie (mobile NAT dropped, wifi died with no RST, etc.).
      // Closing fires onclose, which schedules a reconnect.
      watchdogInterval.current = window.setInterval(() => {
        if (Date.now() - lastMessageAt.current > STALE_THRESHOLD_MS) {
          ws.close();
        }
      }, WATCHDOG_INTERVAL_MS);
    };

    ws.onmessage = (event) => {
      lastMessageAt.current = Date.now();
      try {
        const msg: ServerMessage = JSON.parse(event.data);
        onMessage(msg);
      } catch {
        console.error("Failed to parse WS message");
      }
    };

    ws.onclose = () => {
      setConnected(false);
      clearTimers();

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
  }, [gameId, token, spectator, onMessage, clearTimers]);

  useEffect(() => {
    mountedRef.current = true;
    connect();

    return () => {
      mountedRef.current = false;
      if (reconnectTimer.current) clearTimeout(reconnectTimer.current);
      clearTimers();
      wsRef.current?.close();
    };
  }, [connect, clearTimers]);

  // When the tab becomes visible, mobile browsers may have throttled the
  // ping interval to death. Eagerly reconnect (or recycle a stale socket)
  // so the user doesn't have to act first to discover they're offline.
  useEffect(() => {
    const handleVisibilityChange = () => {
      if (document.visibilityState !== "visible") return;
      if (!mountedRef.current) return;

      const ws = wsRef.current;
      const stale = Date.now() - lastMessageAt.current > VISIBILITY_STALE_THRESHOLD_MS;

      if (!ws || ws.readyState === WebSocket.CLOSED) {
        if (reconnectTimer.current) {
          clearTimeout(reconnectTimer.current);
          reconnectTimer.current = undefined;
        }
        reconnectDelay.current = 1000;
        connect();
      } else if (ws.readyState === WebSocket.OPEN && stale) {
        ws.close();
      }
    };

    document.addEventListener("visibilitychange", handleVisibilityChange);
    return () => {
      document.removeEventListener("visibilitychange", handleVisibilityChange);
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
