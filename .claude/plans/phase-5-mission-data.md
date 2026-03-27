# Phase 5: Mission Data Scraper & Integration

## Goal
Scrape Warhammer 40K mission pack data from Wahapedia and integrate it into the turn tracker so players can select real missions, secondary objectives, and gambit cards during game setup.

## Context
The BSData repo (wh40k-10e) does NOT contain mission data — only army building data. Mission packs, primary missions, secondary objectives, deployment maps, and gambit cards must be scraped from Wahapedia. The site uses client-side JS rendering, so a headless browser (Playwright) is required.

## Target Mission Packs
1. **Leviathan** — `https://wahapedia.ru/wh40k10ed/the-rules/leviathan-battles/`
2. **Pariah Nexus** — `https://wahapedia.ru/wh40k10ed/the-rules/pariah-nexus-battles/`
3. **Chapter Approved 2025-26** — `https://wahapedia.ru/wh40k10ed/the-rules/chapter-approved-2025-26/`

## Data to Extract

### Per Mission Pack
- Pack name and ID
- List of deployment maps (5 types: Search and Destroy, Dawn of War, Sweeping Engagement, Crucible of Battle, Hammer and Anvil)

### Per Primary Mission (typically 9 per pack)
- Name
- Description / rules text
- Scoring rules
- Associated deployment map(s)

### Per Secondary Objective
- Name
- Category: "fixed" or "tactical"
- Description / scoring conditions
- Max VP (typically 8 per secondary)
- Scoring phase (when VP is awarded)

### Per Gambit Card
- Name
- Description
- VP value (typically 12)
- Conditions

## Implementation Steps

### 1. Scraper Setup (`scraper/`)
```
scraper/
├── package.json
├── tsconfig.json
└── src/
    ├── scrape-missions.ts       # Main entry point
    ├── parsers/
    │   ├── common.ts            # Shared DOM parsing utilities
    │   ├── leviathan.ts         # Leviathan-specific parsing
    │   ├── pariah-nexus.ts      # Pariah Nexus-specific parsing
    │   └── chapter-approved.ts  # Chapter Approved-specific parsing
    └── output/
        └── missions.json        # Structured output
```

- Initialize with `npm init` + install `playwright`, `typescript`
- Use Playwright to launch headless Chromium
- Navigate to each mission pack URL
- Wait for JS rendering (likely need `waitForSelector` on mission cards)
- Parse DOM for structured data

### 2. Parser Design
Each mission pack page has different HTML structure, so each gets its own parser. Common patterns:
- Mission cards are likely in expandable sections or tabs
- Secondary objectives may be in a separate section from primaries
- Gambit cards are late-game content (round 3+)

The parsers should:
- Use CSS selectors to find mission/secondary/gambit sections
- Extract text content, strip HTML formatting
- Handle edge cases (multi-part descriptions, nested rules)

### 3. Output Format (`missions.json`)
```json
{
  "missionPacks": [
    {
      "id": "pariah-nexus",
      "name": "Pariah Nexus",
      "description": "...",
      "deploymentMaps": ["search-and-destroy", "dawn-of-war", ...],
      "missions": [
        {
          "name": "Take and Hold",
          "description": "...",
          "deploymentMap": "search-and-destroy",
          "rulesText": "..."
        }
      ],
      "secondaries": [
        {
          "name": "Assassination",
          "category": "fixed",
          "description": "...",
          "maxVp": 8
        }
      ],
      "gambits": [
        {
          "name": "Proceed as Planned",
          "description": "...",
          "vpValue": 12
        }
      ]
    }
  ]
}
```

### 4. Seed Tool Extension
Extend `backend/cmd/seed/main.go` to accept `--missions path/to/missions.json`:
- Parse JSON
- Upsert mission_packs, missions, secondaries, gambits tables
- Idempotent with ON CONFLICT

### 5. Frontend Integration

#### MissionPackPicker Component
- Fetch mission packs from `GET /api/mission-packs`
- Display as selectable cards
- On select: load missions for that pack

#### MissionPicker Component
- Fetch missions from `GET /api/mission-packs/{id}/missions`
- Display with deployment map info
- On select: send `select_mission` action via WebSocket

#### SecondaryPicker Component
- Fetch secondaries from `GET /api/mission-packs/{id}/secondaries`
- Display grouped by category (fixed vs tactical)
- Allow selection of predefined secondaries
- "Custom" option: text input for name + max VP
- Send `select_secondary` / `remove_secondary` actions

#### DeploymentMap Component (stretch)
- Visual representation of deployment zones
- Could be SVG-based or image-based

### 6. GameSetupPage Integration
Add mission/secondary selection steps between faction selection and ready-up:

```
1. Select Faction → 2. Select Detachment → 3. Select Mission Pack
→ 4. Select Mission → 5. Select Secondaries → 6. Ready Up
```

## Technical Notes
- Wahapedia pages use JS rendering — simple HTTP fetches return incomplete HTML
- Playwright is a dev-time dependency only, not needed in production
- The scraper should be idempotent and re-runnable
- Consider caching scraped HTML to avoid hitting Wahapedia repeatedly during development
- GW updates mission packs periodically — the scraper should be easy to re-run

## Risks
- Wahapedia HTML structure may change without notice
- Some content may be behind interactive UI elements (tabs, dropdowns)
- Copyright considerations: storing full rules text vs. just names and scoring summaries
- Mission packs rotate — need a process to add new ones

## Estimated Effort
- Scraper setup + first parser: 4-6 hours
- Additional parsers: 2-3 hours each
- Seed tool extension: 1-2 hours
- Frontend components: 4-6 hours
- Integration + testing: 2-3 hours
