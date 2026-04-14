# Admin Frontend

Standalone React application for managing Tacticarium reference data (factions, stratagems, missions, etc.).

## Tech Stack

- pnpm
- React 18 + TypeScript
- Vite+
- Tailwind CSS 4 (via `@tailwindcss/vite` plugin)
- React Router v7

<!--VITE PLUS START-->

# Using Vite+, the Unified Toolchain for the Web

This project is using Vite+, a unified toolchain built on top of Vite, Rolldown, Vitest, tsdown, Oxlint, Oxfmt, and Vite Task. Vite+ wraps runtime management, package management, and frontend tooling in a single global CLI called `vp`. Vite+ is distinct from Vite, but it invokes Vite through `vp dev` and `vp build`.

## Vite+ Workflow

`vp` is a global binary that handles the full development lifecycle. Run `vp help` to print a list of commands and `vp <command> --help` for information about a specific command.

### Start

- create - Create a new project from a template
- migrate - Migrate an existing project to Vite+
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

## CI Integration

For GitHub Actions, consider using [`voidzero-dev/setup-vp`](https://github.com/voidzero-dev/setup-vp) to replace separate `actions/setup-node`, package-manager setup, cache, and install steps with a single action.

```yaml
- uses: voidzero-dev/setup-vp@v1
  with:
    cache: true
- run: vp check
- run: vp test
```

## Review Checklist for Agents

- [ ] Run `vp install` after pulling remote changes and before getting started.
- [ ] Run `vp check` and `vp test` to validate changes.
<!--VITE PLUS END-->

## Development

```bash
# From repo root:
make dev-admin       # Runs on port 5174

# Or directly:
cd admin && pnpm run dev
```

Requires the Go backend running on port 8080 (`make dev-backend`).

## Auth

Uses GitHub OAuth (separate from the player Discord auth). The backend checks the GitHub user ID against the `ADMIN_GITHUB_IDS` env var allowlist. Admin JWTs have a `role: "admin"` claim.

Token is stored in `localStorage` under the key `admin_token`.

## Project Structure

```
src/
├── api/
│   ├── client.ts      # HTTP client (Bearer token from admin_token)
│   ├── auth.ts        # GitHub auth endpoints
│   └── admin.ts       # Generic CRUD factory + import helpers
├── hooks/
│   └── useAuth.ts     # GitHub auth context
├── components/
│   ├── Layout.tsx     # Sidebar nav + main content
│   ├── DataTable.tsx  # Reusable table with search/filter/delete
│   └── ImportDialog.tsx # File upload modal
└── pages/
    ├── LoginPage.tsx
    ├── AuthCallbackPage.tsx
    ├── DashboardPage.tsx
    └── {entity}/          # One folder per entity type
        ├── {Entity}ListPage.tsx
        └── {Entity}EditPage.tsx
```

## Data Fetching

Currently uses manual `useState` + `useEffect` + `fetch` patterns. A migration to TanStack Query is planned (the player frontend has already been migrated).

## Conventions

- Each entity has a list page and a shared create/edit page
- List pages use `DataTable` component with optional filter dropdowns
- Edit pages use controlled form state, `useParams` for edit vs create
- JSONB fields (scoring rules, scoring options) use structured sub-forms with add/remove buttons
- Import uses `multipart/form-data` upload via `ImportDialog` component
- API client in `admin.ts` uses a generic `crud()` factory — each entity gets `list`, `get`, `create`, `update`, `delete` methods

## Environment Variables

- `VITE_API_URL` — Backend URL (default: `http://localhost:8080`)
