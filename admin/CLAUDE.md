# Admin Frontend

Standalone React application for managing Tacticarium reference data (factions, stratagems, missions, etc.).

## Tech Stack

- pnpm
- React 18 + TypeScript
- Vite 5 (standard, not Vite+)
- Tailwind CSS 4 (via `@tailwindcss/vite` plugin)
- React Router v7

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

## Conventions

- Each entity has a list page and a shared create/edit page
- List pages use `DataTable` component with optional filter dropdowns
- Edit pages use controlled form state, `useParams` for edit vs create
- JSONB fields (scoring rules, scoring options) use structured sub-forms with add/remove buttons
- Import uses `multipart/form-data` upload via `ImportDialog` component
- API client in `admin.ts` uses a generic `crud()` factory — each entity gets `list`, `get`, `create`, `update`, `delete` methods

## Environment Variables

- `VITE_API_URL` — Backend URL (default: `http://localhost:8080`)
