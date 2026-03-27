# Phase 6: Polish, History & Mobile UX

## Goal
Bring the application to production quality: game completion flow, history page refinement, mobile UX polish, error recovery, and performance optimization.

## Tasks

### 1. Game Completion Flow
Currently the game ends when round 5 completes or a player concedes. Polish needed:

- **End-of-game summary screen**: Show both players' VP breakdowns side by side, highlight winner, show key stats (stratagems used, rounds played)
- **Concede confirmation**: Replace `window.confirm` with a styled modal dialog
- **Abandon game**: Add ability for both players to agree to abandon (distinct from concede — no winner)
- **Timeout handling**: If a player disconnects for >5 minutes during an active game, allow the other player to claim victory or pause the game

### 2. Game History Improvements
- **Detailed game view**: Click a completed game to see full event timeline, VP progression chart, stratagems used
- **Stats summary**: Win/loss record, most-played factions, average VP scored
- **Filter/sort**: Filter by faction, date range, win/loss

### 3. Mobile UX Refinement

#### Bottom Sheet for Stratagems
Replace the current collapsible section with a proper bottom sheet:
- Drag handle at top
- Slides up from bottom on tap
- Scrollable list inside
- Dismiss by swiping down or tapping overlay
- Consider using a library like `react-spring` for smooth animations

#### Touch Targets
- Ensure all interactive elements are ≥44px (iOS) / ≥48px (Android)
- CP +/- buttons should be large and easy to tap during gameplay
- VP score buttons should be comfortably spaced

#### Swipe Gestures
- Swipe between your state and opponent state panels
- Swipe to reveal stratagem panel

#### Viewport Management
- Prevent zoom on double-tap (interferes with gameplay)
- Handle virtual keyboard appearance (doesn't push content off screen)
- Lock to portrait orientation on mobile

### 4. Reconnection & Error Recovery

#### Reconnection UX
- Show "Reconnecting..." banner with spinner when WebSocket disconnects
- Auto-request `sync_request` on reconnect to get latest state
- Show "Opponent disconnected" indicator with reconnection status

#### Stale State Detection
- Include a state version/timestamp in GameState
- If client detects a gap in event sequence, request full sync
- Handle case where server restarts mid-game (state reloaded from DB)

#### Error Handling
- Toast notifications for action rejections (instead of alert bars)
- Retry logic for REST API calls (exponential backoff)
- Graceful degradation if stratagems fail to load

### 5. Performance Optimization

#### Memoization
- `useMemo` for stratagem filtering (recompute only when phase/faction/turn changes)
- `React.memo` on StratagemCard, VPRow components (prevent re-render on parent state change)
- Memoize the sorted/filtered game list on LobbyPage

#### Virtualized Lists
- If stratagem list exceeds ~20 items, use `react-window` for virtualized scrolling
- Game history list should be virtualized for users with many completed games

#### Bundle Optimization
- Code-split the GamePage (it's the heaviest page)
- Lazy-load the history page
- Preload stratagem data when entering setup (before game starts)

### 6. Visual Polish

#### Dark Theme (Grimdark)
Refine the color palette to be more thematically appropriate:
- Deep blacks and dark grays for backgrounds
- Gold/amber accents for Imperial factions
- Red accents for Chaos
- Consider faction-specific accent colors on the game page

#### Animations
- Phase advancement: slide transition between phases
- VP score: number tick-up animation
- CP spend: brief flash/pulse on the CP counter
- Round change: dramatic transition between rounds

#### Loading States
- Skeleton screens instead of "Loading..." text
- Smooth transitions between pages

### 7. PWA Support (Optional)
- Add `manifest.json` for "Add to Home Screen"
- Service worker for offline caching of static assets
- App icon and splash screen

## Technical Notes

### Bottom Sheet Implementation Options
1. **CSS-only**: `transform: translateY()` with transition, `touch-action: none` for drag
2. **Library**: `@gorhom/bottom-sheet` (React Native) or roll custom with `react-spring`
3. **Native dialog**: `<dialog>` element with CSS animations

### State Versioning
Add a `version` field to GameState (incremented on each Apply):
```go
type GameState struct {
    Version int `json:"version"`
    // ...
}
```
Client tracks last-seen version. On reconnect, if versions don't match, request full sync.

### Toast Notification System
Simple implementation:
```typescript
// stores/toastStore.ts
interface Toast { id: string; message: string; type: 'error' | 'info' | 'success' }
// Auto-dismiss after 3 seconds
```

## Estimated Effort
- Game completion flow: 3-4 hours
- History improvements: 4-5 hours
- Mobile UX: 6-8 hours
- Reconnection/error handling: 3-4 hours
- Performance: 2-3 hours
- Visual polish: 4-6 hours
- PWA (optional): 2-3 hours
