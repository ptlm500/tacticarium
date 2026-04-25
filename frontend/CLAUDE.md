# Player Frontend

React application for playing Tacticarium games.

## Tech Stack

- pnpm
- React 18 + TypeScript
- Vite+
- Tailwind CSS 4 (via `@tailwindcss/vite` plugin)
- React Router v7
- TanStack Query v5 (REST API data fetching + caching)
- Zustand (real-time game state via WebSocket)
- shadcn/ui + [thegridcn](https://thegridcn.com) (Tron: Ares themed HUD primitives) for components
- lucide-react for icons

## UI Components

Components are installed via the shadcn CLI from two registries, configured in `components.json`:

- **default** shadcn registry — base primitives (`button`, `badge`, `dialog`, `dropdown-menu`, `input`, `label`, `scroll-area`, `separator`, `table`, etc.)
- **`@thegridcn`** — HUD-styled primitives (e.g. `hud-frame`, `spinner`, `alert`, the animated `scan` scanline)

All primitives live in `src/components/ui/` and are imported via the `@/` alias (e.g. `import { Button } from "@/components/ui/button"`). Do not edit these files by hand for theming — restyle via the CSS variables in `src/index.css`.

### Adding components

```bash
vp dlx shadcn@latest add button              # default registry
vp dlx shadcn@latest add @thegridcn/hud-frame # gridcn registry
```

Only add what you will actually use — the registries pull in transitive Radix peer deps. Prune unused primitives (and their Radix deps) when removing features.

### Theming

The app is dark-only. The base palette and four theme variants (`scorpion` (default), `spacewolf`, `blood`, `badmoon`) live in `src/index.css` as `[data-theme="..."]` overrides of `--primary`, `--ring`, `--accent`, and `--chart-{1..5}`. The active theme is controlled by `ThemeSwitcher` (in the page header), which sets `data-theme` on `<html>`.

When adding a new page, follow the established header pattern: grid background + radial fade overlay, `ThemeSwitcher` + back `Button` in the top bar, `HUDFrame` for content panels, and `font-mono uppercase tracking-widest` for labels.

## Data Fetching

REST API calls use **TanStack Query**. WebSocket-driven game state uses **Zustand**.

### Query Hooks

All query/mutation hooks live in `src/hooks/queries/`:

| File                   | Hooks                                                                      |
| ---------------------- | -------------------------------------------------------------------------- |
| `useGamesQueries.ts`   | `useGameList()`, `useGame(id)`, `useGameEvents(id)`                        |
| `useHistoryQueries.ts` | `useGameHistory(filters?)`, `useUserStats()`                               |
| `useFactionQueries.ts` | `useFactions()`, `useDetachments(factionId?)`, `useStratagems(factionId?)` |
| `useMissionQueries.ts` | `useMissions(packId)`, `useMissionRules(packId)`, `useSecondaries(packId)` |
| `useGameMutations.ts`  | `useCreateGame()`, `useJoinGame()`, `useHideGame()`                        |

Query keys are defined in `src/hooks/queryKeys.ts` as a factory object.

### Patterns

- **Dependent queries**: Use `enabled: !!value` (e.g., `useDetachments` only fetches when factionId exists)
- **Parallel queries**: Multiple `useQuery` calls run in parallel automatically
- **Mutations**: Use `mutate()` with `onSuccess`/`onSettled` callbacks — not `mutateAsync` — to avoid unhandled rejections
- **Error messages**: Pages map mutation errors to user-friendly strings rather than surfacing raw API errors
- **Error boundaries**: `QueryErrorBoundary` wraps each route in `App.tsx`; query errors throw to the boundary via `throwOnError: true`

### What NOT to use TanStack Query for

- Real-time game state — handled by WebSocket + Zustand (`useGameConnection` / `useGameStore`)
- Auth state — handled by React Context (`useAuth`)

## API Client

Thin fetch wrapper in `src/api/client.ts` with `api.get<T>()` / `api.post<T>()`. Domain-specific modules in `src/api/` (auth, games, factions, missions). Query hooks call these modules — they are the `queryFn` implementations.

## Testing

- Vitest in browser mode (Chromium via Playwright) — run with `vp test`
- MSW v2 for API mocking (`src/mocks/handlers/`)
- `renderWithProviders()` in `src/test/renderWithProviders.tsx` wraps with QueryClientProvider, AuthContext, and MemoryRouter
- Each test render gets a fresh QueryClient (retry disabled for fast failures)
- Test fixtures in `src/test/fixtures.ts`

## Environment Variables

- `VITE_API_URL` — Backend URL (default: `http://localhost:8080`)

<!--VITE PLUS START-->

# Using Vite+, the Unified Toolchain for the Web

This project is using Vite+, a unified toolchain built on top of Vite, Rolldown, Vitest, tsdown, Oxlint, Oxfmt, and Vite Task. Vite+ wraps runtime management, package management, and frontend tooling in a single global CLI called `vp`. Vite+ is distinct from Vite, but it invokes Vite through `vp dev` and `vp build`.

## Vite+ Workflow

`vp` is a global binary that handles the full development lifecycle. Run `vp help` to print a list of commands and `vp <command> --help` for information about a specific command.

### Start

- config - Configure hooks and agent integration
- staged - Run linters on staged files
- install (`i`) - Install dependencies
- env - Manage Node.js versions

### Develop

- dev - Run the development server
- check - Run format, lint, and TypeScript type checks
- lint - Lint code
- fmt - Format code
- test - Run tests

### Execute

- run - Run monorepo tasks
- exec - Execute a command from local `node_modules/.bin`
- dlx - Execute a package binary without installing it as a dependency
- cache - Manage the task cache

### Build

- build - Build for production
- pack - Build libraries
- preview - Preview production build

### Manage Dependencies

Vite+ automatically detects and wraps the underlying package manager such as pnpm, npm, or Yarn through the `packageManager` field in `package.json` or package manager-specific lockfiles.

- add - Add packages to dependencies
- remove (`rm`, `un`, `uninstall`) - Remove packages from dependencies
- update (`up`) - Update packages to latest versions
- dedupe - Deduplicate dependencies
- outdated - Check for outdated packages
- list (`ls`) - List installed packages
- why (`explain`) - Show why a package is installed
- info (`view`, `show`) - View package information from the registry
- link (`ln`) / unlink - Manage local package links
- pm - Forward a command to the package manager

### Maintain

- upgrade - Update `vp` itself to the latest version

These commands map to their corresponding tools. For example, `vp dev --port 3000` runs Vite's dev server and works the same as Vite. `vp test` runs JavaScript tests through the bundled Vitest. The version of all tools can be checked using `vp --version`. This is useful when researching documentation, features, and bugs.

## Common Pitfalls

- **Using the package manager directly:** Do not use pnpm, npm, or Yarn directly. Vite+ can handle all package manager operations.
- **Always use Vite commands to run tools:** Don't attempt to run `vp vitest` or `vp oxlint`. They do not exist. Use `vp test` and `vp lint` instead.
- **Running scripts:** Vite+ built-in commands (`vp dev`, `vp build`, `vp test`, etc.) always run the Vite+ built-in tool, not any `package.json` script of the same name. To run a custom script that shares a name with a built-in command, use `vp run <script>`. For example, if you have a custom `dev` script that runs multiple services concurrently, run it with `vp run dev`, not `vp dev` (which always starts Vite's dev server).
- **Do not install Vitest, Oxlint, Oxfmt, or tsdown directly:** Vite+ wraps these tools. They must not be installed directly. You cannot upgrade these tools by installing their latest versions. Always use Vite+ commands.
- **Use Vite+ wrappers for one-off binaries:** Use `vp dlx` instead of package-manager-specific `dlx`/`npx` commands.
- **Import JavaScript modules from `vite-plus`:** Instead of importing from `vite` or `vitest`, all modules should be imported from the project's `vite-plus` dependency. For example, `import { defineConfig } from 'vite-plus';` or `import { expect, test, vi } from 'vite-plus/test';`. You must not install `vitest` to import test utilities.
- **Type-Aware Linting:** There is no need to install `oxlint-tsgolint`, `vp lint --type-aware` works out of the box.
