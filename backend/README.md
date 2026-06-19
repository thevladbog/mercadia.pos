# Backend

This directory contains the Mercadia POS backend workspace.

## Prerequisites

- Go **1.26.2+** (toolchain **1.26.4** in `go.work` for builds)

## Layout

- `services/store-edge` - store-local operational API.
- `services/central-backend` - central operational and synchronization API.
- `services/hardware-agent` - local hardware/device API.
- `packages/platform` - shared backend helpers.
- `scripts` - backend automation.

## Commands

Run tests from the POS root:

```powershell
.\backend\scripts\test.ps1
```

CI checks (lint, vulnerability scan, OpenAPI diff) are described in [`docs/development/ci.md`](../docs/development/ci.md).

Regenerate OpenAPI contracts from the POS root:

```powershell
.\backend\scripts\export-openapi.ps1
```

## Local Infrastructure

Start PostgreSQL and NATS for local development:

```bash
docker compose -f infra/docker/docker-compose.yml up -d
```

Environment variables:

| Variable | Default | Used by |
|----------|---------|---------|
| `MERCADIA_STORE_EDGE_DATABASE_URL` | _(empty = in-memory)_ | store-edge |
| `MERCADIA_CENTRAL_BACKEND_DATABASE_URL` | _(empty = in-memory)_ | central-backend |
| `MERCADIA_STORE_EDGE_NATS_URL` | `nats://127.0.0.1:4222` | store-edge outbox publisher |
| `MERCADIA_CENTRAL_BACKEND_NATS_URL` | _(empty = consumer disabled)_ | central-backend JetStream sync consumer |
| `MERCADIA_STORE_EDGE_ADDR` | `:8081` | store-edge |
| `MERCADIA_CENTRAL_BACKEND_ADDR` | `:8082` | central-backend |
| `MERCADIA_CENTRAL_BACKEND_URL` | `http://127.0.0.1:8082` | store-edge catalog sync client |
| `MERCADIA_HARDWARE_AGENT_ADDR` | `127.0.0.1:8083` | hardware-agent |
| `MERCADIA_HARDWARE_AGENT_URL` | `http://127.0.0.1:8083` | store-edge hardware-agent client |
| `MERCADIA_STORE_EDGE_USE_HARDWARE_AGENT` | `false` | store-edge payment/fiscal via hardware-agent |
| `MERCADIA_STORE_EDGE_HARDWARE_AGENT_FALLBACK` | `true` | fallback to mock when hardware-agent command fails |
| `MERCADIA_STORE_EDGE_HARDWARE_AGENT_READINESS_PROBE` | mirrors `USE_HARDWARE_AGENT` | include hardware-agent `/healthz` in store-edge `/readyz` |
| `MERCADIA_STORE_EDGE_CATALOG_SYNC_INTERVAL` | _(empty = disabled)_ | background catalog sync interval (e.g. `5m`) |
| `MERCADIA_STORE_EDGE_DEFAULT_STORE_ID` | `store-1` | default store for background catalog sync |
| `MERCADIA_STORE_EDGE_TERMINAL_OFFLINE_AFTER` | `60s` | store-edge terminal list offline threshold from `lastSeenAt` |
| `MERCADIA_STORE_EDGE_MIGRATIONS_DIR` | walk-up `infra/migrations/store-edge` | store-edge SQL migrations override |
| `MERCADIA_CENTRAL_BACKEND_MIGRATIONS_DIR` | walk-up `infra/migrations/central-backend` | central-backend SQL migrations override |
| `MERCADIA_CENTRAL_BACKEND_SEED_ADMIN_EMAIL` | _(empty = disabled)_ | bootstrap first central admin when no users exist |
| `MERCADIA_CENTRAL_BACKEND_SEED_ADMIN_PASSWORD` | _(empty = disabled)_ | password for seeded central admin |
| `MERCADIA_CENTRAL_BACKEND_SEED_ADMIN_DISPLAY_NAME` | `Central Admin` | display name for seeded central admin |
| `MERCADIA_CENTRAL_BACKEND_SEED_ADMIN_USER_ID` | `seed-admin` | id for seeded central admin |
| `MERCADIA_CENTRAL_BACKEND_SYNC_API_KEY` | _(empty = disabled)_ | shared API key for HTTP sync ingestion and catalog reads (`X-Sync-Api-Key`); store-edge catalog sync uses the same value when set |
| `MERCADIA_OTEL_ENABLED` | `false` | all services (enables OpenTelemetry HTTP tracing) |

Example PostgreSQL URL: `postgres://mercadia:mercadia@127.0.0.1:5433/mercadia_pos?sslmode=disable`

## Database Migrations

When `MERCADIA_STORE_EDGE_DATABASE_URL` or `MERCADIA_CENTRAL_BACKEND_DATABASE_URL` is set,
SQL migrations run automatically during service startup via `platform/migrate` (goose).

Startup log examples:

```text
INFO ✅ migrations applied service=store-edge directory=.../infra/migrations/store-edge from_version=0 to_version=3
INFO ⏭️ migrations already up to date service=central-backend directory=.../infra/migrations/central-backend version=1
```

Override migration directories with `MERCADIA_STORE_EDGE_MIGRATIONS_DIR` or
`MERCADIA_CENTRAL_BACKEND_MIGRATIONS_DIR` when the service is started outside the repo root.

## Store Edge Sync Pipeline

When NATS is enabled on both services, operational events flow from store-edge to central-backend:

```text
store-edge command -> outbox row -> JetStream (mercadia.store-edge.sync.{storeId})
  -> central-backend consumer -> POST-equivalent AcceptEvents -> sync_events table
  -> read-model projection (catalog products, synced payments with lifecycle, synced cash movements, synced fiscal documents, synced returns, synced operational days)
```

On PostgreSQL, command handlers that emit outbox events persist business state and the outbox row in a single database transaction (ADR-0004 transactional outbox). Command paths that write the operation journal (returns, discounts, cash recounts, and related cash movements) commit journal entries atomically with business state and idempotency on PostgreSQL as well. In-memory mode remains single-process and does not use multi-statement transactions.

Local smoke:

1. `docker compose -f infra/docker/docker-compose.yml up -d`
2. Register a store on central-backend: `POST /v1/stores` with `X-Session-Token` (`users.manage`, central admin) when sync key is unset, or `X-Sync-Api-Key` when `MERCADIA_CENTRAL_BACKEND_SYNC_API_KEY` is set
3. Start central-backend with `MERCADIA_CENTRAL_BACKEND_NATS_URL=nats://127.0.0.1:4222`
4. Start store-edge with `MERCADIA_STORE_EDGE_NATS_URL=nats://127.0.0.1:4222` and the same `MERCADIA_CENTRAL_BACKEND_SYNC_API_KEY` when catalog sync or HTTP sync ingestion uses the key
5. Run a checkout command that records an outbox event (for example a captured payment)
6. Confirm the event appears in central sync events: `GET /v1/stores/{storeId}/sync-events` with `X-Session-Token` (`reporting.read`)
7. Confirm projected read models with the same session token: `GET /v1/stores/{storeId}/payments`, `GET /v1/stores/{storeId}/cash-movements`, `GET /v1/stores/{storeId}/fiscal-documents`, `GET /v1/stores/{storeId}/returns`, and `GET /v1/stores/{storeId}/operational-days`
8. After cancel/refund, return settlement, or EoD close on store-edge, confirm the corresponding central read models are updated

When using NATS (steps 3–4 above), sync ingestion does not use the HTTP API key. For manual `POST /v1/stores/{storeId}/sync-events` calls, set `MERCADIA_CENTRAL_BACKEND_SYNC_API_KEY` and pass `X-Sync-Api-Key` with the same value. When the env var is unset, HTTP sync ingestion remains open for local development.

The consumer uses durable name `central-backend-sync` and idempotency keys `nats:{storeId}:{eventId}` so JetStream redelivery is safe.

## OpenAPI And Scalar

- OpenAPI is generated from Go handler registrations via `export-openapi.ps1`.
- Each service serves `/openapi.json` and interactive Scalar docs at `/docs`.
- Scalar is pinned to `@scalar/api-reference@1.60.0` in `packages/platform/httpapi`.
- Huma was evaluated; the project keeps the custom `httpapi` OpenAPI builder because it already
  covers idempotency headers, shared Problem schemas, and stable operation IDs across services.

## Dependencies

Third-party Go modules are pinned in each service `go.mod`. Verify versions against
[pkg.go.dev](https://pkg.go.dev) and run `govulncheck ./...` after changes.

| Package | Version | Service | Purpose |
|---------|---------|---------|---------|
| `github.com/jackc/pgx/v5` | v5.10.0 | store-edge, central-backend | PostgreSQL driver |
| `github.com/pressly/goose/v3` | v3.27.1 | store-edge, central-backend | SQL migrations |
| `github.com/nats-io/nats.go` | v1.52.0 | store-edge, central-backend | NATS JetStream bridge |
| `github.com/prometheus/client_golang` | v1.23.2 | platform (all services) | Prometheus `/metrics` |
| `go.opentelemetry.io/otel` | v1.44.0 | platform (all services) | OpenTelemetry tracing |
| `go.opentelemetry.io/otel/sdk` | v1.44.0 | platform (all services) | OpenTelemetry SDK |
| `go.opentelemetry.io/otel/exporters/stdout/stdouttrace` | v1.44.0 | platform (all services) | OTEL stdout trace exporter |
| `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp` | v0.69.0 | platform (all services) | HTTP tracing middleware |

See `docs/development/dependency-policy.md` for audit rules.

## Current Store Edge Slice

The Store Edge service has the first checkout and terminal monitoring paths:

- `POST /v1/operational-days` - opens the store operational day.
- `GET /v1/operational-days/{operationalDayId}` - returns operational day state.
- `GET /v1/operational-days/{operationalDayId}/summary` - returns EoD readiness, shift counters, receipt counters, payment totals, fiscal totals, cash balances, and cash recount counters.
- `GET /v1/operational-days/{operationalDayId}/receipts` - lists receipts linked to the operational day.
- `GET /v1/operational-days/{operationalDayId}/shifts` - lists cashier shifts linked to the operational day.
- `GET /v1/stores/{storeId}/operational-days/current` - returns the current open operational day for the store.
- `POST /v1/operational-days/{operationalDayId}/close-check` - returns EoD close readiness and blockers.
- `POST /v1/operational-days/{operationalDayId}/close` - closes the operational day when blockers are resolved or overridden.
- `POST /v1/shifts` - opens a personal cashier shift for a terminal and cash drawer. When `openingCashMinor > 0`, `sourceSafeId` is required and posts a `change_fund` movement from the safe into the drawer.
- `GET /v1/shifts/{shiftId}` - returns shift state.
- `GET /v1/shifts/{shiftId}/receipts` - lists receipts opened during the shift.
- `POST /v1/shifts/{shiftId}/close` - closes an open cashier shift with closing cash amount.
- `POST /v1/shifts/{shiftId}/cash-in` - posts a shift-scoped `cash_in` movement into the shift drawer (defaults from external customer source).
- `POST /v1/shifts/{shiftId}/cash-out` - posts a shift-scoped `cash_out` movement from the shift drawer to a safe; requires two-person control via `actorId` and `approvedById`.
- `GET /v1/stores/{storeId}/shifts/open` - lists currently open shifts for the store.
- `POST /v1/receipts` - opens a draft receipt.
- `GET /v1/receipts/{receiptId}` - returns current receipt state.
- `POST /v1/receipts/{receiptId}/lines` - adds an item line to a draft receipt.
- `POST /v1/receipts/{receiptId}/scan` - looks up a product by barcode and adds it to a draft receipt.
- `POST /v1/receipts/{receiptId}/cancel` - cancels a draft receipt with a reason and actor.
- `GET /v1/catalog/products/by-barcode/{barcode}` - returns product data from the local catalog cache.
- `POST /v1/stores/{storeId}/catalog/sync` - pulls catalog delta from central backend into the local cache.
- `POST /v1/receipts/{receiptId}/payments` - creates a captured mock payment.
- `GET /v1/receipts/{receiptId}/payments` - lists receipt payments.
- `POST /v1/receipts/{receiptId}/payments/{paymentId}/cancel` - cancels a captured same-day payment on the receipt business date and rolls receipt payment progress back. Supports `card_mock` (hardware-agent `cancel` when enabled) and `cash` (posts compensating `cash_sale_reversal` ledger movement).
- `POST /v1/receipts/{receiptId}/payments/{paymentId}/refund` - refunds a captured payment after fiscalization or on a later business date; same-day pre-fiscal payments must use cancel instead. Supports `card_mock` (hardware-agent `refund` when enabled) and `cash` (posts `cash_sale_reversal` ledger movement). Optional `amountMinor` performs a partial refund; omitted amount refunds the full remaining balance.
- `POST /v1/receipts/{receiptId}/fiscal-documents` - creates a mock fiscal document for a fully paid receipt.
- `GET /v1/receipts/{receiptId}/fiscal-documents` - lists receipt fiscal documents.
- `POST /v1/receipts/{receiptId}/returns` - creates a with-receipt return against a fiscalized receipt. Per-line quantities are capped across prior with-receipt returns in `completed` or `settled` status on the same receipt.
- `GET /v1/receipts/{receiptId}/returns` - lists returns for the receipt (paginated, newest first).
- `GET /v1/returns/{returnId}` - returns return state.
- `GET /v1/stores/{storeId}/returns` - lists returns for the store (paginated, newest first).
- `POST /v1/stores/{storeId}/returns/no-receipt` - creates a no-receipt return with approval.
- `POST /v1/returns/{returnId}/settle` - settles a with-receipt return by refunding captured payments on the original receipt proportionally across payment methods, or disburses cash for an approved no-receipt return. Supports partial line returns when the return total is less than the receipt total. Optional `drawerId` selects the payout drawer for no-receipt returns (otherwise resolved from the actor's open shift). Cumulative settled return totals for a receipt cannot exceed the receipt total.
- `POST /v1/returns/{returnId}/fiscal-documents` - creates a mock fiscal return/correction document for a settled with-receipt return. One fiscal document per return; requires the original receipt to be fiscalized.
- `GET /v1/returns/{returnId}/fiscal-documents` - lists fiscal documents linked to the return (0 or 1 document).
- `POST /v1/stores/{storeId}/cash-movements` - posts an immutable cash movement between cash containers.
- `GET /v1/stores/{storeId}/cash-movements` - lists cash movements posted for the store.
- `POST /v1/stores/{storeId}/bank-collections` - posts a `safe_to_bank` collection from a safe to a bank container; requires two-person control.
- `POST /v1/stores/{storeId}/business-expenses` - posts an `expense` disbursement from a safe to a payee expense container; requires two-person control.
- `GET /v1/stores/{storeId}/cash-balances` - derives current cash container balances from posted movements.
- `POST /v1/stores/{storeId}/cash-recounts` - records a cash recount for a drawer or safe.
- `GET /v1/stores/{storeId}/cash-recounts` - lists cash recounts for the store.
- `POST /v1/stores/{storeId}/cash-recounts/{recountId}/resolve` - resolves a cash recount discrepancy.
- `POST /v1/terminals/{terminalId}/heartbeat` - records terminal heartbeat/state.
- `GET /v1/terminals/{terminalId}` - returns last known terminal state.
- `GET /v1/stores/{storeId}/terminals` - paginated terminal fleet snapshot with offline derivation from `lastSeenAt`.
- `GET /v1/stores/{storeId}/monitoring/terminals` - paginated terminal monitoring cards with shift, receipt, and drawer KPIs.
- `GET /v1/stores/{storeId}/monitoring/summary` - store-level monitoring counters and today's fiscalized receipt totals.
- `GET /v1/stores/{storeId}/terminals/events` - SSE stream of terminal heartbeat events (documented outside OpenAPI).

Use `GET /v1/stores/{storeId}/monitoring/*` for REST polling of terminal tiles and store KPIs; use SSE for live heartbeat push.

When `MERCADIA_STORE_EDGE_USE_HARDWARE_AGENT=true`, card payments and fiscalization send commands to
the local hardware-agent (`authorize`/`capture`, `cancel`, `refund`, `print_receipt`) with mock fallback enabled by default.
Same-day card payment cancel uses the hardware-agent `cancel` command when a terminal is configured.
Same-day cash payment cancel posts a compensating `cash_sale_reversal` movement from the receipt drawer back to the external customer container.
Post-sale card refunds use the hardware-agent `refund` command with the original provider reference.
Post-fiscal cash refunds post a compensating `cash_sale_reversal` movement from the receipt drawer back to the external customer container.
Return settlement refunds captured payments on the original receipt through the existing refund paths (card via hardware-agent when enabled, cash via ledger reversal) and transitions the return to `settled`. Partial returns allocate refund amounts proportionally across refundable payment balances. Cumulative settled return totals for a receipt are capped at the receipt total at settlement time; per-line return quantities are capped at create time across prior with-receipt returns in `completed` or `settled` status. No-receipt returns settle with a `no_receipt_return_payout` cash movement from the drawer; the approver recorded on the return is stored on the movement and cannot be the disbursing actor. After settlement, with-receipt returns can be fiscalized with `POST /v1/returns/{returnId}/fiscal-documents` (mock `print_receipt` via hardware-agent when enabled); receipt listing includes both sale and return fiscal documents.

Command endpoints require `Idempotency-Key`. Reusing the same key for the same command returns
the same result; reusing it with a different command payload returns an idempotency conflict.

The current persistence adapter is in-memory when no database URL is configured. PostgreSQL
repositories and migrations are used when `MERCADIA_STORE_EDGE_DATABASE_URL` is set.
The current catalog contains demo seed products so local scan workflows can be exercised before
running catalog sync against central-backend.
Operational day operations are the first EoD foundation. Only one operational day can be open
per store. Closing uses the same readiness checks exposed by `close-check`: open cashier shifts
unresolved receipts, unresolved cash recount discrepancies, and non-zero drawer balances block
closure, and a day with no fiscalized sales receipts requires an explicit admin override.
The operational day summary combines readiness blockers with shift counters, receipt counters,
payment totals by method, fiscal totals, cash balances, non-zero drawer count, and cash recount
counters so senior cashier and admin clients can render EoD progress without stitching together
low-level calls.
Payments are modeled separately from receipts. The current implementation supports captured mock
cash/card payments and prevents paying more than the receipt remaining amount.
Cash payments are also posted into the cash ledger as `cash_sale` movements from an external
customer source into the receipt drawer. External containers are ignored by the controlled cash
balance read model, so the drawer balance increases without exposing a synthetic customer
container in cash-office views.
Fiscalization is modeled separately from payments. The current implementation supports mock
fiscal documents only after the receipt is fully paid.
Shift operations are the first Store Operations foundation. The current implementation enforces
one open shift per terminal and one open shift per cashier. Shifts are linked to a cash drawer,
the current operational day, and the business date. In the Store Edge API runtime, opening a
shift requires an open operational day for the store. When `openingCashMinor > 0`, shift open
requires `sourceSafeId` and posts a `change_fund` ledger movement from the safe into the drawer.
Closed shifts are removed from the open-shift read model. A shift cannot be closed while it has
unresolved receipts. Closing a shift with `closingCashMinor > 0` requires final collection details
and posts a `drawer_to_safe` cash movement from the shift drawer to the selected safe. On PostgreSQL,
shift open and close persist shift state, cash movements, and idempotency in a single transaction.
The final collection requires two-person control through separate `actorId` and `approvedById` values.
Mid-shift cash in and cash out are shift-scoped commands that post `cash_in` and `cash_out` ledger movements with journal entries inside the same PostgreSQL transaction as idempotency when persistence is enabled. Shift-scoped cash movements also enqueue `cash.movement.posted` outbox events for central sync. Bank collection and business expense commands post typed safe operations with journal and outbox recording.
Cash operations are modeled as an append-only ledger. Posted cash movements are not edited in
place; corrections must be represented by a new movement. The first control rule is separation
of duties: the actor posting a cash movement cannot approve the same movement.
Cash balances are a derived read model calculated from posted cash movements. Drawer and safe
balances reflect posted movements including shift opening `change_fund` and closing collection.
The in-memory implementation uses the same derivation as PostgreSQL.
Cash recounts compare a counted amount with the derived expected balance. Balanced recounts can
be recorded by one actor. Recounts with a discrepancy require a second approving actor, and the
same actor cannot approve their own discrepancy. A discrepancy remains open until it is resolved
with a resolution note and a second approving actor; unresolved discrepancies block EoD closure.
Receipt lifecycle is coordinated with those separate state machines: receipts start as `draft`,
move to `payment_started` after partial payment, `paid` after full payment, and `fiscalized`
after fiscal document creation. Draft receipts can be cancelled with a reason and actor before
payment starts. Receipt lines can be changed only while the receipt is `draft`.
In the Store Edge API runtime, receipt opening requires an open operational day and an open
cashier shift for the same store, terminal, and cashier. Accepted receipts carry
`operationalDayId`, `businessDate`, `shiftId`, and `drawerId` so sales, payments, fiscalization,
cash accountability, and EoD checks can be joined without guessing from timestamps.

## Current Central Backend Slice

The central backend service exposes store registration, sync ingestion, synced read models, and cross-store reporting:

- `GET /v1/central/status` - returns region status and registered store count. Requires `X-Session-Token` with `reporting.read`.
- `POST /v1/stores` - registers a store (idempotent). Requires `X-Sync-Api-Key` when configured, otherwise `X-Session-Token` with `users.manage`.
- `GET /v1/stores` - lists registered stores. Requires `X-Session-Token` with `reporting.read`.
- `POST /v1/stores/{storeId}/sync-events` - accepts synchronized Store Edge event batches. When `MERCADIA_CENTRAL_BACKEND_SYNC_API_KEY` is set, requires `X-Sync-Api-Key`.
- `GET /v1/stores/{storeId}/sync-events` - lists accepted sync events (paginated, newest first). Requires `X-Session-Token` with `reporting.read`.
- `GET /v1/stores/{storeId}/catalog/products` - lists catalog products for a store. Requires `X-Sync-Api-Key` when configured.
- `GET /v1/stores/{storeId}/catalog/delta` - returns catalog products updated since a timestamp. Requires `X-Sync-Api-Key` when configured.
- `GET /v1/stores/{storeId}/payments` - lists synchronized payments (paginated). Requires `X-Session-Token` with `reporting.read`.
- `GET /v1/stores/{storeId}/payments/{paymentId}` - returns a synchronized payment. Requires `X-Session-Token` with `reporting.read`.
- `GET /v1/stores/{storeId}/cash-movements` - lists synchronized cash movements (paginated). Requires `X-Session-Token` with `reporting.read`.
- `GET /v1/stores/{storeId}/cash-movements/{cashMovementId}` - returns a synchronized cash movement. Requires `X-Session-Token` with `reporting.read`.
- `GET /v1/stores/{storeId}/fiscal-documents` - lists synchronized fiscal documents (paginated). Requires `X-Session-Token` with `reporting.read`.
- `GET /v1/stores/{storeId}/fiscal-documents/{fiscalDocumentId}` - returns a synchronized fiscal document. Requires `X-Session-Token` with `reporting.read`.
- `GET /v1/stores/{storeId}/returns` - lists synchronized returns (paginated). Requires `X-Session-Token` with `reporting.read`.
- `GET /v1/stores/{storeId}/returns/{returnId}` - returns a synchronized return. Requires `X-Session-Token` with `reporting.read`.
- `GET /v1/stores/{storeId}/operational-days` - lists synchronized closed operational days (paginated). Requires `X-Session-Token` with `reporting.read`.
- `GET /v1/stores/{storeId}/operational-days/{operationalDayId}` - returns a synchronized closed operational day. Requires `X-Session-Token` with `reporting.read`.
- `GET /v1/stores/{storeId}/reporting/summary?since=&until=` - store KPI snapshot from synced projections (RFC3339 inclusive window). Requires `X-Session-Token` with `reporting.read`.
- `GET /v1/central/reporting/summary?since=&until=&region=` - network KPI aggregate across registered stores. Requires `X-Session-Token` with `reporting.central.read`.
- `GET /v1/central/reporting/stores?since=&until=&region=&limit=&offset=` - paginated per-store reporting rows for drill-down. Requires `X-Session-Token` with `reporting.central.read`.
- `POST /v1/auth/sessions` - creates a central admin session from email and password; returns opaque token for `X-Session-Token`.
- `GET /v1/central/users` - lists central users (`users.manage`, central admin role).
- `POST /v1/central/users` - creates a central user (`users.manage`).
- `GET /v1/central/users/{userId}` - returns a central user (`users.manage`).
- `PATCH /v1/central/users/{userId}` - updates roles, password, or active state (`users.manage`).

Reporting aggregates use synced fiscal documents (`kind=receipt` for revenue proxy), payments, returns, cash movements, and closed operational days within the requested time window.

Central roles in v1: `central_viewer` (reporting read) and `central_admin` (reporting + user management). NATS sync ingestion does not use the HTTP API key. HTTP sync ingestion and catalog reads require `X-Sync-Api-Key` when `MERCADIA_CENTRAL_BACKEND_SYNC_API_KEY` is set; synced read-model GET endpoints require `X-Session-Token` with `reporting.read`.

Auth smoke:

1. Start central-backend with seed env vars, for example `MERCADIA_CENTRAL_BACKEND_SEED_ADMIN_EMAIL=admin@example.com` and `MERCADIA_CENTRAL_BACKEND_SEED_ADMIN_PASSWORD=change-me`
2. `POST /v1/auth/sessions` with `{"email":"admin@example.com","password":"change-me"}`
3. Call reporting endpoints with `X-Session-Token: <token from step 2>`
4. Confirm `GET /v1/central/reporting/summary` returns 401 without the header

After a captured payment or NATS-delivered sync event, use `GET /v1/stores/{storeId}/sync-events` with `X-Session-Token`
to confirm central ingestion without querying Postgres directly. Then query `GET /v1/stores/{storeId}/reporting/summary`
or the central reporting endpoints for cross-store KPIs.
