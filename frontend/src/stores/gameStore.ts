import { create } from "zustand";
import { GameState, GameEvent } from "../types/game";

interface GameStore {
  gameState: GameState | null;
  events: GameEvent[];
  error: string | null;
  opponentConnected: boolean;

  setGameState: (state: GameState) => void;
  setEvents: (events: GameEvent[]) => void;
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
  setEvents: (events) =>
    set((s) => {
      // Preserve any live events the WS pushed before the REST history resolved.
      const seen = new Set<number>();
      const merged: GameEvent[] = [];
      for (const e of events) {
        if (e.id != null) seen.add(e.id);
        merged.push(e);
      }
      for (const e of s.events) {
        if (e.id != null && seen.has(e.id)) continue;
        merged.push(e);
      }
      return { events: merged };
    }),
  addEvent: (event) =>
    set((s) => {
      if (event.id != null && s.events.some((e) => e.id === event.id)) {
        return s;
      }
      return { events: [...s.events, event] };
    }),
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
