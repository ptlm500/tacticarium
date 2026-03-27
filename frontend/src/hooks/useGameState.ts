import { useCallback } from 'react';
import { useGameStore } from '../stores/gameStore';
import { useWebSocket } from './useWebSocket';
import { ServerMessage } from '../types/ws';
import { GameState, GameEvent } from '../types/game';

export function useGameConnection(gameId: string, token: string) {
  const { setGameState, addEvent, setError, setOpponentConnected } =
    useGameStore();

  const handleMessage = useCallback(
    (msg: ServerMessage) => {
      switch (msg.type) {
        case 'state_update':
          setGameState(msg.data as GameState);
          break;
        case 'event':
          addEvent(msg.data as GameEvent);
          break;
        case 'error': {
          const err = msg.data as { message: string };
          setError(err.message);
          setTimeout(() => setError(null), 3000);
          break;
        }
        case 'player_connected':
          setOpponentConnected(true);
          break;
        case 'player_disconnected':
          setOpponentConnected(false);
          break;
      }
    },
    [setGameState, addEvent, setError, setOpponentConnected]
  );

  const { connected, sendAction } = useWebSocket({
    gameId,
    token,
    onMessage: handleMessage,
  });

  return { connected, sendAction };
}
