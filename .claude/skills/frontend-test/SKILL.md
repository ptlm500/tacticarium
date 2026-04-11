---
name: frontend-test
description: |
  Writing and fixing frontend tests for this project's React/Vitest browser-mode test suite.
  Use this skill whenever you need to: write new component or hook tests, fix failing tests,
  debug act() warnings, set up WebSocket or REST API mocks, or understand the test infrastructure.
  Triggers on: "write tests for", "fix tests", "test is failing", "act warning", "add test coverage",
  or any work touching *.test.tsx / *.test.ts files in the frontend.
---

# Frontend Testing Guide

## Running tests

```bash
cd frontend && vp test          # run all tests
vp test src/path/to/file.test.tsx  # run a specific test file
```

`vp` is the Vite+ CLI. Do NOT use `npx vitest` or `pnpm test` directly.

## Imports

Vitest globals (`vi`, `describe`, `it`, `expect`, `beforeEach`, `afterEach`, `beforeAll`, `afterAll`) are available without import (`globals: true` in config).

For lifecycle hooks used in setup files, import from `vite-plus/test`:
```ts
import { beforeAll, afterAll, afterEach } from "vite-plus/test";
```

For rendering and DOM queries, import from `@testing-library/react`:
```ts
import { render, screen, act } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
```

## Test structure

### Component tests (pages)

Page components need routing and auth context. Use `renderWithProviders` from `src/test/renderWithProviders.tsx`:

```tsx
import { screen, act } from "@testing-library/react";
import { renderWithProviders } from "../test/renderWithProviders";
import { Route, Routes } from "react-router-dom";

// Wrap render in await act(async () => {...}) to flush async effects
await act(async () => {
  renderWithProviders(
    <Routes>
      <Route path="/game/:id" element={<MyPage />} />
    </Routes>,
    { user: mockUser, route: "/game/game-1" },
  );
});
```

### Component tests (non-page)

Simple components that don't need routing can use `render` directly:

```tsx
import { render, screen } from "@testing-library/react";

render(<MyComponent prop="value" />);
```

### Hook tests

Test hooks via a thin wrapper component:

```tsx
function TestComponent({ gameId }: { gameId: string }) {
  const { data } = useMyHook(gameId);
  return <span data-testid="result">{JSON.stringify(data)}</span>;
}

await act(async () => {
  render(<TestComponent gameId="game-1" />);
});
```

## Async rendering and act()

Components that trigger async effects on mount (WebSocket connections, API calls, store updates) **must** be rendered inside `await act(async () => {...})`. This flushes microtasks like MSW's WebSocket `onopen` dispatch, preventing act() warnings and ensuring state is settled before assertions.

```tsx
// CORRECT
await act(async () => {
  render(<MyComponent />);
});

// WRONG - will produce act() warnings from async state updates
render(<MyComponent />);
```

Use `vi.waitFor()` for assertions that depend on async state:

```tsx
await vi.waitFor(() => {
  expect(screen.getByText("Ready")).toBeTruthy();
});
```

## Mocking

### MSW setup

MSW (Mock Service Worker) intercepts HTTP and WebSocket requests in the browser.

- **Worker**: `src/mocks/browser.ts` — `setupWorker` with REST + WS handlers
- **REST handlers**: `src/mocks/handlers/rest.ts` — default API responses using fixtures
- **WS handlers**: `src/mocks/handlers/ws.ts` — exports `gameWs` link and `wsHandlers`
- **Setup**: `src/test/setupTests.ts` — starts worker, cleans up between tests

### REST API mocking

Default handlers cover common endpoints (auth, factions, missions, games). Override per-test with `worker.use()`:

```tsx
import { http, HttpResponse } from "msw";
import { worker } from "../mocks/browser";

worker.use(
  http.get("http://localhost:8080/api/factions/:factionId/stratagems", () => {
    return HttpResponse.json(myCustomData);
  }),
);
// worker.resetHandlers() in afterEach restores defaults
```

### WebSocket mocking

**Use `ws.link()` + `worker.use()` for per-test WebSocket handlers.** This is critical — never call `gameWs.addEventListener("connection", ...)` directly in tests. Direct listeners accumulate across tests because `worker.resetHandlers()` doesn't clean them up, causing phantom messages and act() warnings.

```tsx
import { ws } from "msw";
import { worker } from "../mocks/browser";

const testLink = ws.link("ws://localhost:8080/ws/game/*");
worker.use(
  testLink.addEventListener("connection", ({ client }) => {
    client.send(JSON.stringify({ type: "state_update", data: gameState }));
  }),
);

await act(async () => {
  render(<TestComponent gameId="game-1" token="tok" />);
});
```

The pattern: create a temporary `ws.link()` with the same URL, add your connection handler, register it via `worker.use()`. The runtime handler fires alongside the initial handler. `worker.resetHandlers()` in `afterEach` removes it automatically.

To listen for client-to-server messages:

```tsx
const sentMessages: string[] = [];
testLink.addEventListener("connection", ({ client }) => {
  client.addEventListener("message", (event) => {
    sentMessages.push(typeof event.data === "string" ? event.data : "");
  });
  client.send(JSON.stringify({ type: "pong", data: null }));
});
```

## Store (Zustand)

The game store (`src/stores/gameStore.ts`) holds `gameState`, `events`, `error`, and `opponentConnected`.

Reset in `beforeEach` (this is also done globally in setupTests.ts, but explicit resets make tests clearer):

```tsx
import { useGameStore } from "../stores/gameStore";

beforeEach(() => {
  useGameStore.getState().reset();
});
```

Pre-populate the store for tests that need existing state:

```tsx
const gs = makeGameState({ status: "active", currentRound: 2 });
useGameStore.getState().setGameState(gs);
```

## Fixtures

`src/test/fixtures.ts` provides factory functions and mock data:

| Export | Description |
|---|---|
| `mockUser` | Default test user (`id: "user-1"`, `username: "TestPlayer"`) |
| `makePlayerState(overrides?)` | Creates a `PlayerState` with sensible defaults |
| `makeGameState(overrides?)` | Creates a `GameState` with two players, active status |
| `mockFactions` | Space Marines, Chaos Space Marines, Orks |
| `mockDetachments` | Gladius Task Force, Ironstorm Spearhead |
| `mockStratagems` | Command Re-roll, Storm of Fire, Heroic Intervention |
| `mockMissions` | Supply Drop, Scorched Earth |
| `mockRules` | Hidden Supplies, Chilling Rain |
| `mockSecondaries` | Behind Enemy Lines, Assassination |
| `mockEvent` | A `phase_advanced` event |

## Complete test example

```tsx
import { screen, act } from "@testing-library/react";
import { renderWithProviders } from "../test/renderWithProviders";
import { useGameStore } from "../stores/gameStore";
import { makeGameState, mockUser } from "../test/fixtures";
import { ws } from "msw";
import { worker } from "../mocks/browser";
import { Route, Routes } from "react-router-dom";

function renderGame(overrides?: Partial<GameState>) {
  const gs = makeGameState(overrides);
  useGameStore.getState().setGameState(gs);
  localStorage.setItem("token", "test-token");

  const testLink = ws.link("ws://localhost:8080/ws/game/*");
  worker.use(
    testLink.addEventListener("connection", ({ client }) => {
      client.send(JSON.stringify({ type: "state_update", data: gs }));
    }),
  );

  return renderWithProviders(
    <Routes>
      <Route path="/game/:id" element={<GamePage />} />
    </Routes>,
    { user: mockUser, route: "/game/game-1" },
  );
}

describe("GamePage", () => {
  beforeEach(() => {
    useGameStore.getState().reset();
    localStorage.clear();
  });

  it("shows the current phase", async () => {
    await act(async () => {
      renderGame({ currentPhase: "shooting" });
    });

    await vi.waitFor(() => {
      expect(screen.getByText(/Shooting Phase/)).toBeTruthy();
    });
  });
});
```

## Gotchas

1. **Never import from `vitest`** — use `vite-plus/test` for setup-file imports, and globals everywhere else.
2. **Never add WS listeners directly to `gameWs`** — use the `ws.link()` + `worker.use()` pattern.
3. **Always `await act(async () => {...})` around renders** for components with async effects.
4. **Set `localStorage.setItem("token", "test-token")`** before rendering pages that call `getToken()`.
5. **`vi.waitFor` default timeout is 1000ms** — increase with `{ timeout: 5000 }` for slower operations.
6. **act() warnings are suppressed** in `setupTests.ts` for the known false-positive from MSW's async WebSocket `onopen`. If you see new act warnings, investigate the root cause rather than adding more suppression.
7. **Wait for async data before interacting** — if a test depends on data loaded via an API call (e.g., mission scoring rules), wait for a UI element that proves the data has loaded before clicking buttons. For example, wait for `screen.getByText("Quick Score")` (rendered by `MissionScoring` when mission data loads) before clicking "Advance Phase". Without this, the click fires before `currentMission` is set and the expected behavior won't trigger.
8. **Duplicate text across page and modals** — elements like scoring buttons or secondary names appear in both the main page components and in modals (e.g., `ScoringPrompt`). Use `getAllByText` / `getAllByRole` instead of `getByText` when asserting, and scope clicks to the right element (e.g., find buttons by class: `buttons.find(btn => btn.closest("button")?.classList.contains("bg-indigo-800"))`).
