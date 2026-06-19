# Frontend

Mercadia POS frontend workspace: applications, shared packages, and Orval-generated
API clients.

## Prerequisites

- **Node.js 24+** (see [`.node-version`](.node-version))
- **pnpm 11.5+** (pinned via `packageManager` in [`package.json`](package.json))

Package installs use the public npm registry configured in [`.npmrc`](.npmrc):

```ini
registry=https://registry.npmjs.org/
```

## Workspace layout

```text
frontend/
  apps/admin-web/                 # Central admin UI (Vite + React)
  packages/api-clients/central/   # Orval client for central-backend OpenAPI
```

## Install and verify

```bash
cd frontend
pnpm install
pnpm orval:central
pnpm typecheck
pnpm audit
```

Regenerate the central client after backend OpenAPI changes:

```bash
pnpm orval:central
```

## Local development — admin-web

1. Start **central-backend** on port `8082` with a seeded admin user:

   ```bash
   export MERCADIA_CENTRAL_BACKEND_SEED_ADMIN_EMAIL=admin@example.com
   export MERCADIA_CENTRAL_BACKEND_SEED_ADMIN_PASSWORD=change-me
   # run central-backend (see backend/README.md)
   ```

2. Start the admin UI (Vite dev server proxies `/v1` → `http://127.0.0.1:8082`):

   ```bash
   cd frontend
   pnpm --filter admin-web dev
   ```

3. Open `http://localhost:5173`, sign in, and open **Central Reporting**.

Optional: set `VITE_CENTRAL_BACKEND_URL` when the API is not same-origin (bypasses the
Vite proxy).

## Dependency policy

Use stable releases only. Before bumping pins, verify current versions via official docs
and `npm view <package> version`, then update manifests and lockfile together. See
[`docs/development/dependency-policy.md`](../docs/development/dependency-policy.md).

## Expected future apps

- `apps/pos-terminal`
- `apps/sco-terminal`
- `apps/senior-cashier-web`
- `apps/senior-cashier-terminal`
- `apps/assistant-station`

Do not place backend services or backend packages here.
