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
pnpm verify
```

`pnpm verify` runs, in order: Orval generation, typecheck, ESLint, Prettier check, and
admin-web production build.

Individual checks:

```bash
pnpm orval:central
pnpm typecheck
pnpm lint              # ESLint
pnpm lint:fix          # ESLint with auto-fix
pnpm format:check      # Prettier
pnpm format            # Prettier write
pnpm audit
```

Orval-generated files under `packages/api-clients/**/src/generated/` are excluded from ESLint
and Prettier — regenerate them with `pnpm orval:central` instead of editing manually.

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

3. Open `http://localhost:5173`, sign in, and open **Central Reporting** or **Users**.

### Central users smoke test

Requires a seeded user with the `central_admin` role (default seed admin):

1. Sign in and open **Users** in the header.
2. Confirm the user list loads.
3. Create a viewer account (`central_viewer` role) via **Create user**.
4. Edit the new user: toggle **Active**, adjust roles, or set a new password.
5. Sign in as a `central_viewer` user: the **Users** nav link is hidden and direct
   `/central/users` URLs redirect back to reporting.

Optional: set `VITE_CENTRAL_BACKEND_URL` when the API is not same-origin (bypasses the
Vite proxy).

## Dependency policy

Use stable releases only. Before bumping pins, verify current versions via official docs
and `npm view <package> version`, then update manifests and lockfile together. See
[`docs/development/dependency-policy.md`](../docs/development/dependency-policy.md).

## Continuous integration

GitHub Actions workflow [`.github/workflows/frontend.yml`](../.github/workflows/frontend.yml)
runs on changes to `frontend/**` and `contracts/openapi/**`. It verifies Orval output is
committed, then runs typecheck, lint, format check, build, and dependency audit.

## Expected future apps

- `apps/pos-terminal`
- `apps/sco-terminal`
- `apps/senior-cashier-web`
- `apps/senior-cashier-terminal`
- `apps/assistant-station`

Do not place backend services or backend packages here.
