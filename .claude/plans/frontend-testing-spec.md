# Frontend Testing Strategy Specification: Tacticarium

## 1. Architectural Overview
This specification outlines the testing strategy for a React 18 frontend application (Warhammer 40K Tacticarium) built with Vite, React Router, Zustand, and TailwindCSS.

**Core Testing Stack:**

- **Test Runner:** Vitest (running in Browser Mode).
- **Browser Provider:** Playwright (headless Chromium/Firefox/Webkit).
- **Network Mocking:** MSW (Mock Service Worker v2) for REST APIs and WebSocket interceptors.
- **DOM Interactions:** `@testing-library/react` and `@testing-library/user-event` (integrated with Vitest Browser).

## 2. Environment Setup & Tooling Configuration

### 2.1 Vitest Browser Configuration
The LLM must configure `vitest.config.ts` to utilize the Playwright provider for browser mode.

- **Framework:** React.
- **Browser:** `provider: 'playwright', name: 'chromium'`.
- **Setup File:** Specify a `setupTests.ts` to initialize MSW and clean up DOM/Zustand state.

### 2.2 MSW Setup
MSW will serve as the single source of truth for all network requests. Do not use `vi.mock('fetch')` or mock the `api/client.ts` file.

- **REST API Handlers (`src/mocks/handlers/rest.ts`):** Mock `/api/auth/me`, `/api/factions`, `/api/mission-packs/*`, `/api/games`, and `/api/users/me/history`.
- **WebSocket Handlers (`src/mocks/handlers/ws.ts`):**
  - Use MSW 2.0's `ws` API to intercept connections to `ws://localhost:8080/ws/game/:gameId`.
  - Create helper functions in the tests to emit simulated server messages (`state_update`, `event`, etc.) to test the `useWebSocket` and Zustand store integrations.
- **Worker Setup:** Use `setupWorker` from `msw/browser` inside the Vitest setup file to start intercepting requests in the browser environment.

## 3. Testing Layers & Target Scenarios

### Layer 1: Global State & Hooks (Unit/Integration)
Test custom hooks and Zustand stores in isolation using the browser environment.

**Target:** `src/stores/gameStore.ts`
- **Rule:** Always reset the store between tests (LLM must write a `resetStore` utility).
- **Scenarios:**
  - Updating game state (`setGameState`).
  - Appending events to the log (`addEvent`).
  - Handling connection flags (`setOpponentConnected`).

**Target:** `src/hooks/useWebSocket.ts` & `src/hooks/useGameState.ts`
- **Strategy:** Render a test component that uses the hook. Use MSW's WebSocket mock to establish a connection.
- **Scenarios:**
  - Ensure connection is established automatically.
  - Assert that incoming `state_update` messages correctly update the `gameStore`.
  - Assert that calling `sendAction` transmits the correct JSON payload over the WebSocket.
  - Verify ping intervals and auto-reconnect logic on close.

### Layer 2: Core Components (Component Integration)
Render interactive components and assert UI changes based on props and user interactions.

**Target:** Game Controls (`CPCounter`, `VPCounter`, `PhaseTracker`)
- **Scenarios:**
  - Clicking +/- on CP/VP triggers the `onAdjust`/`onScore` callbacks with correct values.
  - Buttons are disabled when limits are reached (e.g., `cp <= 0` or `canGainCP === false`).
  - Visual representation matches the `currentPhase` in `PhaseTracker`.

**Target:** Secondary Missions (`SecondaryPanel.tsx`, `ScoringPrompt.tsx`)
- **Scenarios:**
  - Expand/collapse logic toggles visibility.
  - Achieving/Discarding secondary missions triggers correct callbacks (`onAchieve`, `onDiscard`).
  - Tactical UI variations (draw buttons, CP costs for "New Orders") behave correctly based on mode (fixed vs tactical) and CP limits.

### Layer 3: Pages & Routing (End-to-End User Flows)
Mount full page components wrapped in a `MemoryRouter` and `AuthContext`.

**Target:** `LobbyPage.tsx`
- **Scenarios:**
  - Verify API calls: fetches `/api/games` on load.
  - Create Game: Clicking "Create Game" calls the API and navigates to `/game/:id/setup`.
  - Join Game: Entering a code and clicking "Join" handles valid/invalid network responses.

**Target:** `GameSetupPage.tsx`
- **Scenarios:**
  - Loads and displays Factions, Detachments, Missions, and Twists from MSW.
  - Selecting a Faction cascades to fetching/displaying relevant Detachments.
  - "Ready Up" button is disabled until all required fields (Faction, Detachment, Mission, Mode) are selected.

**Target:** `GamePage.tsx` (The Critical Path)
- **Mock State:** Seed the `gameStore` with a full `GameState` mock before rendering.
- **Scenarios:**
  - Validates conditional rendering: shows "Victory/Defeat" if status is completed.
  - Validates "Turn Banner" indicates if it is the player's turn or opponent's turn.
  - Stratagem Panel filtering: Ensure stratagems only show if they match the current phase, turn, and detachment.
  - Advance Phase: Clicking "Advance Phase" either brings up the `ScoringPrompt` modal or directly sends the `advance_phase` WS action based on battle round/mission timing.

## 4. LLM Instruction Prompts (Rules of Engagement)
When generating code for this spec, the LLM must strictly adhere to the following rules:

- **Vitest Browser Mode Semantics:** Use `@testing-library/react` for rendering components. Because Vitest runs in a real Playwright browser, prefer `userEvent.setup()` for interactions (clicks, typing) over `fireEvent` to simulate real browser behavior.
- **Strict Mocking Rule:** DO NOT mock child components. Render the full component tree. Isolate behavior entirely via MSW network mocks and seeded Zustand state.
- **Authentication Context:** Create a `renderWithProviders` utility function that wraps tested components in React Router's `<MemoryRouter>` and the `AuthContext.Provider` (defaulting to a logged-in mock user).
- **Zustand Hygiene:** The LLM must include a `beforeEach` hook in store/hook tests that calls `useGameStore.getState().reset()` to prevent state leakage between browser tests.
- **WebSocket MSW API:** When testing WS interactions, use the `@mswjs/interceptors` or MSW's native `ws` link. Create a mock WS server instance in the test setup, capture the client connection, and use `.send()` to push mock ServerMessage objects (e.g., `state_update`) to the React app.
- **Accessibility Queries First:** Always use `getByRole`, `getByText`, or `getByLabelText` for DOM assertions. Avoid `getByTestId` unless targeting highly complex layout wrappers where semantic HTML is ambiguous.
- **File Naming:** Tests should be collocated with their targets using the `*.test.tsx` or `*.test.ts` naming convention (e.g., `GamePage.test.tsx` next to `GamePage.tsx`).

## 5. First Milestones for the LLM
*(Instruct the LLM to execute these in order)*

1. **Step 1:** Generate `vitest.workspace.ts` (or config), `setupTests.ts`, and the initial MSW handler files (`handlers/rest.ts`, `handlers/ws.ts`).
2. **Step 2:** Write the test suite for `stores/gameStore.ts` and `hooks/useWebSocket.ts`.
3. **Step 3:** Write component tests for `ScoringPrompt.tsx` and `StratagemPanel.tsx` (heavy business logic).
4. **Step 4:** Write the heavy integration test for `GameSetupPage.tsx` proving the API cascading fetches work.
5. **Step 5:** Write the integration test for `GamePage.tsx` proving WebSocket messages update the UI properly.