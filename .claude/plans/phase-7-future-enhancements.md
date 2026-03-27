# Phase 7: Future Enhancements

## Features Deferred from Initial Scope

These were discussed during planning but intentionally left out of the MVP. They represent natural next steps once the core application is stable.

---

### 1. Social Features
**Deferred because**: User said "leave out of scope for now"

Potential additions:
- **Friends list**: Add friends by Discord username, see their online status
- **Win/loss record**: Track record against specific opponents
- **ELO/ranking**: Competitive ranking system based on game outcomes
- **Rematch**: Quick "play again" button at game end

### 2. Army List Integration
**Deferred because**: User chose "just faction + detachment" for now

Potential additions:
- **Full army list input**: Import army lists from BattleScribe XML or New Recruit
- **Unit tracking**: Track which units are destroyed / in reserves
- **Points validation**: Ensure army list is valid for the selected battle size (1000/2000/3000)
- **BSData XML parsing**: Use the `BSData/wh40k-10e` repo data for unit definitions

### 3. Full Detachment Rules
**Deferred because**: User said "just the name is sufficient"

Potential additions:
- **Detachment abilities**: Display the detachment's special rule text
- **Enhancement selection**: Pick enhancements from the detachment's options
- **Detachment-specific scoring**: Some detachments modify how secondaries work

### 4. Spectator Mode
- Allow third parties to watch a game in real-time via a spectator link
- Read-only WebSocket connection (no actions allowed)
- Could enable tournament streaming / coaching

### 5. Tournament Support
- **Multi-game brackets**: Create a tournament with multiple rounds
- **Swiss pairing**: Auto-pair players based on W/L record
- **Scoring**: Tournament points, strength of schedule
- **Export**: Generate results sheets

### 6. Offline/Local Mode
- Play without network connection (both players on same device)
- Sync results when back online
- Useful at game stores with poor WiFi

### 7. Battle Size Support
The current implementation assumes Strike Force (2000 points). Could add:
- Incursion (1000 points)
- Onslaught (3000 points)
- Each with different mission selections

### 8. Crusade Mode
10th edition has a narrative play format called Crusade with:
- Experience tracking across games
- Battle Honours and Battle Scars
- Unit progression
- Crusade-specific missions

This would be a significant expansion requiring its own game mode and data model.

### 9. Data Management Admin Panel
- Admin UI for managing mission packs, missions, secondaries
- Add new content when GW releases updates without re-running scrapers
- User-contributed mission packs (community content)

### 10. Notifications
- Push notifications when it's your turn (if opponent advances on their device)
- Discord webhook integration (post game results to a Discord channel)
- Email summary of completed games

---

## Technical Debt to Address

### Before Scaling
- **go.sum file**: Currently missing — needs `go mod tidy` after Go is installed
- **sqlc integration**: Current handlers use raw SQL queries; migrate to sqlc for type-safe generated code
- **Test suite**: No tests exist yet — need unit tests for game engine, integration tests for handlers
- **CI/CD pipeline**: GitHub Actions for lint, test, build on PR
- **Rate limiting**: Add rate limiting to REST endpoints and WebSocket actions
- **Logging**: Structured logging (replace `log.Printf` with `slog` or `zerolog`)

### Database
- **Connection pooling config**: Tune pgxpool settings for production load
- **Query optimization**: Add indexes on game_events(game_id, created_at) for history queries
- **Data retention**: Policy for cleaning up abandoned games older than N days
- **Backup strategy**: Automated Postgres backups on Railway

### Security
- **CSRF protection**: Add CSRF tokens for state-changing REST requests
- **Rate limiting on auth**: Prevent brute-force on invite codes
- **Input sanitization**: Validate all action data fields server-side
- **WebSocket auth refresh**: Handle JWT expiry during long games (>7 days)
