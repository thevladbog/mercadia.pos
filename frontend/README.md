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
  apps/pos-terminal/                   # POS terminal dev shell (Vite + React)
  packages/ui/                         # @mercadia/ui — Radix components + theme system
  packages/api-clients/central/        # Orval client for central-backend OpenAPI
  packages/api-clients/store-edge/     # Orval client for store-edge OpenAPI
```

## Install and verify

```bash
cd frontend
pnpm install
pnpm verify
```

`pnpm verify` runs, in order: Orval generation, typecheck, ESLint, Prettier check, `@mercadia/ui` unit tests, and
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

### Central catalog smoke test

Requires central-backend (store list and catalog read-model). Catalog data is populated via
store-edge sync events (see dev seed flow).

1. Sign in as seeded `central_viewer` and open **Catalog** in the header.
2. Select a store — the product table loads (or empty state if no catalog sync yet).
3. Use the search box to filter rows by product ID, name, barcode, or tax category.
4. Confirm **Refresh** reloads products for the selected store.

### Central sync explorer smoke test

Requires central-backend (store list and sync read-models). Data is populated via store-edge sync.

1. Sign in and open **Sync** in the header.
2. Select a store — sync events and entity tabs load (or empty state if no sync yet).
3. On **Sync events**, click a source event ID or event type — read-only detail modal shows payload JSON.
4. Change tab or store — URL updates with `tab` and `store` query params; shared links restore state.
5. Open a payment or return entity detail page — click `receiptId` to open the central receipt summary modal; on returns,
   `paymentIds` link to payment entity pages.
6. On the **Returns** tab, click a receipt ID — central receipt summary modal lists related sync projections.

### Store monitoring smoke test

Requires central-backend (store list) and store-edge (monitoring KPIs/terminals):

1. Register at least one store in central-backend (or use existing seed data).
2. Sign in and open **Monitoring** in the header.
3. Select a store from the dropdown — KPI cards and terminal table should load.
4. Confirm data refreshes automatically (every 5 seconds) or via **Refresh**.
5. Confirm the **Terminal events** panel shows **Connected** and lists `terminal_heartbeat`
   events when store-edge emits terminal heartbeats.
6. Toggle **List** / **Tiles** view and use the search box to filter terminals client-side.
7. When a terminal has an active receipt, confirm current receipt ID, status, and total appear
   in list columns and tile cards; click the receipt ID to open the read-only receipt detail modal.
8. Click a terminal ID to open terminal detail — current receipt fields use the same drill-down;
   click shift ID to open read-only shift detail modal.

### Store Safe smoke test

Requires central-backend (store list) and store-edge (cash-office APIs). Write operations require
`central_admin` role.

1. Sign in as seeded `central_admin` and open **Safe** in the header.
2. Select a store — the cash overview KPI panel shows safe/drawer/bank totals, container counts,
   movement and recount totals, and open recount discrepancies; per-container balances and
   journal tables load below.
3. Confirm data refreshes automatically (every 5 seconds) or via **Refresh**.
4. Post **Issue change fund** with actor `senior-1` and approver `admin-1` — balances and movements refresh.
5. Click movement and recount IDs in the journal tables — read-only detail modals show all row fields.
6. Use search boxes above movements and recounts tables — filters apply to the current page only.
7. As `central_viewer`, follow an EoD blocker deep-link to Safe with `?recount=` — read-only recount detail opens (not resolve).
8. Sign in as `central_viewer` — action buttons on Safe are hidden (read-only).

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
6. On the **Open shifts** tab, close an open shift (0.00 cash or with safe + actors when
   collecting cash) — row disappears and overview blockers refresh.
7. On overview blockers, click a reference ID or action — open shift/receipt detail modals,
   navigate to Safe (drawer balance / recount), or switch to the Open shifts tab.
8. On **Open shifts**, click a shift ID — read-only shift detail modal opens.
9. On **Journal**, click reference IDs — shift/receipt/return detail modals, or navigate to Safe
   for movement/recount entries (movement detail via `?movement=` deep link).
10. Expanded overview KPI panels show receipt, cash, payment, and fiscal rollups from the summary API.
11. Sign in as `central_viewer` — open/close action panels and shift row actions are hidden;
    blocker and journal drill-down remains read-only (detail modals, no write CTAs).

Optional env vars when APIs are not same-origin (bypass Vite proxy):

- `VITE_CENTRAL_BACKEND_URL` — central-backend base URL
- `VITE_STORE_EDGE_URL` — store-edge base URL

## Local development — pos-terminal

The POS terminal app is a Vite dev shell with a Store Edge checkout demo and optional central
layout-template loading.

1. Start **store-edge** on port `8081` and open an operational local catalog (see backend README).
2. Optionally start **central-backend** on port `8082` when using `?templateId=` layout templates.
3. Start the POS UI:

   ```bash
   cd frontend
   pnpm --filter pos-terminal dev
   ```

4. Open `http://localhost:5174`. Use **Prepare terminal**, **Open receipt**, **Scan product**,
   **Capture payment**, and **Fiscalize** to run the mock Store Edge flow.

Default demo env vars:

- `VITE_POS_STORE_ID=store-1`
- `VITE_POS_TERMINAL_ID=pos-1`
- `VITE_POS_CASHIER_ID=cashier-1`
- `VITE_POS_DRAWER_ID=drawer-1`
- `VITE_POS_OPENED_BY_ID=admin-1`
- `VITE_POS_FISCAL_DEVICE_ID=fiscal-1`
- `VITE_POS_DEMO_BARCODE=4600000000000`
- `VITE_POS_STORE_TIME_ZONE` — optional IANA store time zone for operational day date selection
- `VITE_STORE_EDGE_URL` — optional Store Edge base URL when bypassing the Vite proxy
- `VITE_STORE_EDGE_SESSION_TOKEN` — optional Store Edge session token
- `VITE_LAYOUT_TEMPLATE_ID` and `VITE_CENTRAL_SESSION_TOKEN` — optional central layout template access

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
