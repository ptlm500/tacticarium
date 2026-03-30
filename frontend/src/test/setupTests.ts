import { beforeAll, afterAll, afterEach } from "vite-plus/test";
import { cleanup } from "@testing-library/react";
import { worker } from "../mocks/browser";
import { useGameStore } from "../stores/gameStore";

beforeAll(async () => {
  await worker.start({ quiet: true });
  console.log('worker started')
});

afterEach(() => {
  cleanup();
  useGameStore.getState().reset();
  worker.resetHandlers();
});

afterAll(() => {
  worker.stop();
});
