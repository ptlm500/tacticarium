import { create } from "zustand";
import { GameState, GameEvent } from "../types/game";

interface GameStore {
  gameState: GameState | null;
  events: GameEvent[];
  error: string | null;
  opponentConnected: boolean;

  setGameState: (state: GameState) => void;
  addEvent: (event: GameEvent) => void;
  setError: (error: string | null) => void;
  setOpponentConnected: (connected: boolean) => void;
  reset: () => void;
}

export const useGameStore = create<GameStore>((set) => ({
  gameState: null,
  events: [],
  error: null,
  opponentConnected: false,

  setGameState: (gameState) => set({ gameState, error: null }),
  addEvent: (event) => set((s) => ({ events: [...s.events, event] })),
  setError: (error) => set({ error }),
  setOpponentConnected: (opponentConnected) => set({ opponentConnected }),
  reset: () =>
    set({
      gameState: null,
      events: [],
      error: null,
      opponentConnected: false,
    }),
}));
