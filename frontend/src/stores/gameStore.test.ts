import { useGameStore } from "./gameStore";
import { makeGameState, mockEvent } from "../test/fixtures";

describe("gameStore", () => {
  beforeEach(() => {
    useGameStore.getState().reset();
  });

  it("starts with null game state", () => {
    const state = useGameStore.getState();
    expect(state.gameState).toBeNull();
    expect(state.events).toEqual([]);
    expect(state.opponentConnected).toBe(false);
  });

  describe("setGameState", () => {
    it("sets game state", () => {
      const store = useGameStore.getState();
      const gs = makeGameState();
      store.setGameState(gs);

      const updated = useGameStore.getState();
      expect(updated.gameState).toEqual(gs);
    });
  });

  describe("addEvent", () => {
    it("appends events to the log", () => {
      const store = useGameStore.getState();
      store.addEvent(mockEvent);
      store.addEvent({ ...mockEvent, eventType: "cp_adjusted" });

      expect(useGameStore.getState().events).toHaveLength(2);
      expect(useGameStore.getState().events[1].eventType).toBe("cp_adjusted");
    });

    it("dedupes events with the same id (e.g. WS event already in REST history)", () => {
      const store = useGameStore.getState();
      store.addEvent({ ...mockEvent, id: 7 });
      store.addEvent({ ...mockEvent, id: 7 });
      expect(useGameStore.getState().events).toHaveLength(1);
    });

    it("appends events that have no id without deduping", () => {
      const store = useGameStore.getState();
      store.addEvent({ ...mockEvent, id: undefined });
      store.addEvent({ ...mockEvent, id: undefined });
      expect(useGameStore.getState().events).toHaveLength(2);
    });
  });

  describe("setEvents", () => {
    it("seeds the log with historical events", () => {
      const store = useGameStore.getState();
      store.setEvents([
        { ...mockEvent, id: 1 },
        { ...mockEvent, id: 2 },
      ]);
      expect(useGameStore.getState().events.map((e) => e.id)).toEqual([1, 2]);
    });

    it("preserves live events that arrived before the history resolved", () => {
      const store = useGameStore.getState();
      store.addEvent({ ...mockEvent, id: 5, eventType: "stratagem_used" });
      store.setEvents([
        { ...mockEvent, id: 1 },
        { ...mockEvent, id: 2 },
      ]);
      expect(useGameStore.getState().events.map((e) => e.id)).toEqual([1, 2, 5]);
    });

    it("drops live events that overlap with the history (deduped by id)", () => {
      const store = useGameStore.getState();
      store.addEvent({ ...mockEvent, id: 2 });
      store.setEvents([
        { ...mockEvent, id: 1 },
        { ...mockEvent, id: 2 },
        { ...mockEvent, id: 3 },
      ]);
      expect(useGameStore.getState().events.map((e) => e.id)).toEqual([1, 2, 3]);
    });
  });

  describe("setOpponentConnected", () => {
    it("toggles opponent connection flag", () => {
      const store = useGameStore.getState();
      store.setOpponentConnected(true);
      expect(useGameStore.getState().opponentConnected).toBe(true);

      store.setOpponentConnected(false);
      expect(useGameStore.getState().opponentConnected).toBe(false);
    });
  });

  describe("reset", () => {
    it("resets all state to defaults", () => {
      const store = useGameStore.getState();
      store.setGameState(makeGameState());
      store.addEvent(mockEvent);
      store.setOpponentConnected(true);

      store.reset();

      const reset = useGameStore.getState();
      expect(reset.gameState).toBeNull();
      expect(reset.events).toEqual([]);
      expect(reset.opponentConnected).toBe(false);
    });
  });
});
