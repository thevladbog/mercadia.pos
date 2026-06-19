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
  apps/admin-web/                      # Central admin UI (Vite + React)
  packages/api-clients/central/        # Orval client for central-backend OpenAPI
  packages/api-clients/store-edge/     # Orval client for store-edge OpenAPI
```

## Install and verify

```bash
cd frontend
pnpm install
pnpm verify
```

`pnpm verify` runs, in order: Orval generation, typecheck, ESLint, Prettier check, and
admin-web production build.

CI workflow details: [`docs/development/ci.md`](../docs/development/ci.md).

Individual checks:

```bash
pnpm orval:central
pnpm orval:store-edge
pnpm typecheck
pnpm lint              # ESLint
pnpm lint:fix          # ESLint with auto-fix
pnpm format:check      # Prettier
pnpm format            # Prettier write
pnpm audit
```

Orval-generated files under `packages/api-clients/**/src/generated/` are excluded from ESLint
and Prettier — regenerate them with `pnpm orval:central` or `pnpm orval:store-edge` instead of
editing manually.

Regenerate clients after backend OpenAPI changes:

```bash
pnpm orval:central
pnpm orval:store-edge
```

## Local development — admin-web

1. Start **central-backend** on port `8082` with a seeded admin user:

   ```bash
   export MERCADIA_CENTRAL_BACKEND_SEED_ADMIN_EMAIL=admin@example.com
   export MERCADIA_CENTRAL_BACKEND_SEED_ADMIN_PASSWORD=change-me
   # run central-backend (see backend/README.md)
   ```

2. Start **store-edge** on port `8081` for store monitoring (see `backend/README.md`).

3. Start the admin UI (Vite dev server proxies most `/v1` traffic to central `:8082`; store-edge
   paths under `/v1/stores/{storeId}/monitoring/*`, `/v1/stores/{storeId}/terminals/*`,
   `/v1/stores/{storeId}/cash-*`, `/v1/stores/{storeId}/bank-*`, `/v1/stores/{storeId}/business-*`,
   `/v1/stores/{storeId}/operation-journal`, `/v1/stores/{storeId}/operational-days`,
   `/v1/stores/{storeId}/shifts`, `/v1/operational-days/*`, `/v1/shifts/*`, and `/v1/terminals/*`
   go to store-edge `:8081`):

   ```bash
   cd frontend
   pnpm --filter admin-web dev
   ```

4. Open `http://localhost:5173`, sign in, and open **Central Reporting**, **Monitoring**, **Safe**,
   **EoD**, or **Users**.

The admin UI defaults to **Russian** (`ru`); use the language switcher in the header to switch to
**English** (`en`). Locale files live in `apps/admin-web/src/i18n/locales/`. Agent i18n rules:
[`docs/development/admin-web-i18n.md`](../docs/development/admin-web-i18n.md).

### Central users smoke test

Requires a seeded user with the `central_admin` role (default seed admin):

1. Sign in and open **Users** in the header.
2. Confirm the user list loads.
3. Create a viewer account (`central_viewer` role) via **Create user**.
4. Edit the new user: toggle **Active**, adjust roles, or set a new password.
5. Sign in as a `central_viewer` user: the **Users** nav link is hidden and direct
   `/central/users` URLs redirect back to `/central/dashboard`.

### Store monitoring smoke test

Requires central-backend (store list) and store-edge (monitoring KPIs/terminals):

1. Register at least one store in central-backend (or use existing seed data).
2. Sign in and open **Monitoring** in the header.
3. Select a store from the dropdown — KPI cards and terminal table should load.
4. Confirm data refreshes automatically (every 5 seconds) or via **Refresh**.
5. Confirm the **Terminal events** panel shows **Connected** and lists `terminal_heartbeat`
   events when store-edge emits terminal heartbeats.
6. Toggle **List** / **Tiles** view and use the search box to filter terminals client-side.

### Store Safe smoke test

Requires central-backend (store list) and store-edge (cash-office APIs). Write operations require
`central_admin` role.

1. Sign in as seeded `central_admin` and open **Safe** in the header.
2. Select a store — balance KPI cards, cash movements table, and recounts table should load.
3. Confirm data refreshes automatically (every 5 seconds) or via **Refresh**.
4. Post **Issue change fund** with actor `senior-1` and approver `admin-1` — balances and movements refresh.
5. Sign in as `central_viewer` — action buttons on Safe are hidden (read-only).

### Store EoD smoke test

Requires central-backend (store list) and store-edge (store-operations APIs):

1. Sign in and open **EoD** in the header.
2. Select a store with an open operational day — overview KPIs, blockers, open shifts, and
   operation journal tabs should load.
3. If no operational day is open, confirm the empty state message appears (not an error panel).
4. As `central_admin` on a store with no open day, use **Open operational day** — confirm with
   today's business date; success notice appears and overview loads.
5. As `central_admin`, when the day can close (or only `no_sales_receipts` requires override),
   use **Close operational day** — confirm in the modal; success notice appears and the page
   shows the no-open-day state.
6. Sign in as `central_viewer` — open/close action panels are hidden (read-only).

Optional env vars when APIs are not same-origin (bypass Vite proxy):

- `VITE_CENTRAL_BACKEND_URL` — central-backend base URL
- `VITE_STORE_EDGE_URL` — store-edge base URL

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
