---
name: game-issue
description: |
  Fix issues where game rules documentation and implementation are out of sync, or where either
  contains a bug. Takes an initial issue description, investigates both docs and code, asks the
  user which is correct when they conflict, then fixes both sides and writes/updates tests.
  Triggers on: "/game-issue <description of problem>"
---

# Game Issue Resolution

You are fixing a game rule issue in the Tacticarium turn tracker. The user has described a problem
that may involve the documentation, the backend game engine, the frontend, the API layer, or any
combination of these.

## Input

The user has provided an initial description of the issue after `/game-issue`. This may be vague
or specific. Your first job is to understand the problem fully before making any changes.

## Step 1: Investigate

Read the relevant game rules documentation and the corresponding implementation code. Always
consult both sides:

**Documentation** (in `docs/`):
- `docs/game-overview.md` — Game lifecycle and Warhammer 40K concepts
- `docs/game-setup.md` — Game creation, joining, setup configuration
- `docs/turn-structure.md` — Rounds, turns, phases, phase advancement
- `docs/scoring.md` — VP categories, scoring mechanics, win conditions
- `docs/secondary-objectives.md` — Fixed/tactical modes, deck management
- `docs/special-mechanics.md` — Stratagems, CP, gambits, challenger cards, adapt or die

**Backend game engine** (in `backend/internal/game/`):
- `engine.go` — Core action handlers (setup, phase advance, CP, VP, stratagems, gambits, concede)
- `engine_missions.go` — Mission system actions (secondaries, tactical deck, challenger, adapt or die)
- `state.go` — Game state, player state, phase/status definitions
- `rules.go` — Constants (max VP, CP per phase, challenger threshold) and helper functions
- `actions.go` — Action and event type definitions
- `twists.go` — Twist-specific logic (adapt or die, new orders CP cost)

**Backend handlers** (in `backend/internal/handler/`):
- `game_handler.go` — HTTP endpoints, game persistence, history

**Frontend** (in `frontend/src/`):
- `stores/gameStore.ts` — Zustand state management
- `pages/GameSetupPage.tsx`, `pages/GamePage.tsx` — Main game UI
- `hooks/useGameState.ts` — WebSocket/store bridge
- `components/game/` — Phase tracker, VP counter, stratagem panel, secondary panel, game log

Not every issue affects all layers. Focus your investigation on the layers relevant to the issue —
it's fine to skip frontend investigation for a pure backend logic change, or skip handlers if the
issue is only in the game engine. Don't waste time reading code that clearly isn't involved.

Read any files that are relevant to the issue the user described. Do not guess — read the actual
code before forming conclusions.

## Step 2: Clarify

After investigating, present your findings to the user. If you find a conflict between docs and
code, or if the issue is ambiguous, **ask the user for guidance**. Structure your question like:

```
I found a discrepancy between the docs and the implementation:

**Documentation says:** <what the docs say>
**Code does:** <what the code actually does>

Which is the intended behaviour?
1. The docs are correct — I'll update the code to match
2. The code is correct — I'll update the docs to match
3. Neither — <ask the user to describe the correct behaviour>
```

If the issue is clearly a bug in one place (e.g., a typo in docs, an off-by-one in code), you can
state what you intend to fix and ask for confirmation rather than presenting options.

Always ask at least one clarifying question before making changes. Even if the fix seems obvious,
confirm your understanding of the intended behaviour. Don't assume — the user may have context you
don't.

## Step 3: Plan the fix

Once you have clarity on the intended behaviour, outline what you will change:

- Which doc files need updating
- Which code files need updating (backend engine, handlers, frontend)
- Which existing tests need updating (flag any test changes outside the direct scope of the issue)
- What new tests are needed to cover the corrected behaviour

Present this plan to the user and get confirmation before proceeding.

## Step 4: Implement

Make the changes. Follow this order:

1. **Fix the code** (backend engine, handlers, frontend — whatever is needed)
2. **Update the docs** to match the corrected behaviour
3. **Write or update tests**

### Test guidelines

**Before writing or updating any tests**, read the existing test helpers and fixture functions in
the relevant test file. Understanding what `newTestState()` vs `newActiveTestState()` return
(including default field values like CP=0) is critical for getting assertions right. Misunderstanding
a helper's defaults leads to wrong expected values.

**Backend tests** (`backend/internal/game/engine_test.go` and `backend/internal/handler/`):
- Framework: standard Go `testing` package with `github.com/stretchr/testify`
- Use existing helpers: `newTestState()`, `newActiveTestState()`, `makeDeck()` in engine tests
- Use `testutil` package helpers for handler tests: `MustSetupTestEnv()`, `CleanDatabase()`,
  `CreateTestUser()`, `CreateTestGame()`, `DoRequest()`, etc.
- Pattern: set up state → apply action → assert events and resulting state

**Frontend tests** (`frontend/src/**/*.test.{ts,tsx}`):
- Framework: Vitest via `vp test` (NOT `npx vitest`)
- Follow the patterns in the `frontend-test` skill (see `.claude/skills/frontend-test/SKILL.md`)
- Use fixtures from `src/test/fixtures.ts`: `makeGameState()`, `makePlayerState()`, etc.
- Use MSW for API/WebSocket mocking with the `ws.link()` + `worker.use()` pattern
- Always `await act(async () => {...})` for components with async effects

### Test scope rules

- **New tests**: Write tests that specifically cover the behaviour being fixed. Each test should
  validate one aspect of the corrected rule.
- **Existing test updates**: Only modify tests that are directly affected by the change (e.g.,
  a test that was asserting the old incorrect behaviour). If you find yourself needing to change
  a test that seems unrelated to the issue, **stop and ask the user for confirmation** before
  modifying it. Explain why you think the test needs changing and let them decide.
- **Do not refactor tests** that are not part of the fix. Don't rename, reorganise, or "improve"
  test files beyond what the issue requires.

### Watch for cascading test impacts

Changing core mechanics (e.g., how often CP is gained, VP caps, phase order) can cascade into
tests that aren't directly about the mechanic you changed. For example, a test that plays through
multiple rounds and asserts final CP totals will break if CP is now gained more frequently — even
if that test is "about" something else like round advancement.

After making your code change, scan existing tests for assertions on values affected by your change
(e.g., grep for `CP` assertions if you changed CP logic). Update these tests' expected values with
clear comments explaining the new arithmetic. Include these in your plan so the user can review them.

## Step 5: Verify

Run the tests to confirm everything passes:

```bash
# Backend
cd backend && go test ./...

# Frontend
cd frontend && vp test
```

If tests fail, diagnose and fix. If a failure is in an area unrelated to your change, flag it to
the user rather than silently fixing it.

## Step 6: Summary

Provide a concise summary of what was changed and why:
- Files modified (with brief description of each change)
- Tests added or updated
- Any remaining concerns or follow-up items
