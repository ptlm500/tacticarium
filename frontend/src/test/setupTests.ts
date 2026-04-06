import { beforeAll, afterAll, afterEach } from "vite-plus/test";
import { cleanup } from "@testing-library/react";
import { worker } from "../mocks/browser";
import { useGameStore } from "../stores/gameStore";

// Suppress React act() warnings from async WebSocket state updates (MSW
// dispatches onopen via queueMicrotask which fires after act() completes).
const _origError = console.error;
console.error = (...args: unknown[]) => {
  if (typeof args[0] === "string" && args[0].includes("not wrapped in act")) return;
  _origError.apply(console, args);
};

beforeAll(async () => {
  await worker.start({ quiet: true });
  console.log("worker started");
});

afterEach(() => {
  cleanup();
  useGameStore.getState().reset();
  worker.resetHandlers();
});

afterAll(() => {
  worker.stop();
});
