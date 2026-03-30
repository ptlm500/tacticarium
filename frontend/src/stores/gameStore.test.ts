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
    expect(state.error).toBeNull();
    expect(state.opponentConnected).toBe(false);
  });

  describe("setGameState", () => {
    it("sets game state and clears error", () => {
      const store = useGameStore.getState();
      store.setError("some error");
      const gs = makeGameState();
      store.setGameState(gs);

      const updated = useGameStore.getState();
      expect(updated.gameState).toEqual(gs);
      expect(updated.error).toBeNull();
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
  });

  describe("setError", () => {
    it("sets and clears error", () => {
      const store = useGameStore.getState();
      store.setError("oh no");
      expect(useGameStore.getState().error).toBe("oh no");

      store.setError(null);
      expect(useGameStore.getState().error).toBeNull();
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
      store.setError("error");
      store.setOpponentConnected(true);

      store.reset();

      const reset = useGameStore.getState();
      expect(reset.gameState).toBeNull();
      expect(reset.events).toEqual([]);
      expect(reset.error).toBeNull();
      expect(reset.opponentConnected).toBe(false);
    });
  });
});
