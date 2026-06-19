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
| `MERCADIA_STORE_EDGE_MIGRATIONS_DIR` | walk-up `infra/migrations/store-edge` | store-edge SQL migrations override |
| `MERCADIA_CENTRAL_BACKEND_MIGRATIONS_DIR` | walk-up `infra/migrations/central-backend` | central-backend SQL migrations override |
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
```

Local smoke:

1. `docker compose -f infra/docker/docker-compose.yml up -d`
2. Register a store on central-backend: `POST /v1/stores`
3. Start central-backend with `MERCADIA_CENTRAL_BACKEND_NATS_URL=nats://127.0.0.1:4222`
4. Start store-edge with `MERCADIA_STORE_EDGE_NATS_URL=nats://127.0.0.1:4222`
5. Run a checkout command that records an outbox event (for example a captured payment)
6. Confirm the event appears in central `sync_events` (via Postgres or future query API)

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
- `POST /v1/shifts` - opens a personal cashier shift for a terminal and cash drawer.
- `GET /v1/shifts/{shiftId}` - returns shift state.
- `GET /v1/shifts/{shiftId}/receipts` - lists receipts opened during the shift.
- `POST /v1/shifts/{shiftId}/close` - closes an open cashier shift with closing cash amount.
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
- `POST /v1/receipts/{receiptId}/fiscal-documents` - creates a mock fiscal document for a fully paid receipt.
- `GET /v1/receipts/{receiptId}/fiscal-documents` - lists receipt fiscal documents.
- `POST /v1/stores/{storeId}/cash-movements` - posts an immutable cash movement between cash containers.
- `GET /v1/stores/{storeId}/cash-movements` - lists cash movements posted for the store.
- `GET /v1/stores/{storeId}/cash-balances` - derives current cash container balances from posted movements.
- `POST /v1/stores/{storeId}/cash-recounts` - records a cash recount for a drawer or safe.
- `GET /v1/stores/{storeId}/cash-recounts` - lists cash recounts for the store.
- `POST /v1/stores/{storeId}/cash-recounts/{recountId}/resolve` - resolves a cash recount discrepancy.
- `POST /v1/terminals/{terminalId}/heartbeat` - records terminal heartbeat/state.
- `GET /v1/terminals/{terminalId}` - returns last known terminal state.
- `GET /v1/stores/{storeId}/terminals/events` - SSE stream of terminal heartbeat events (documented outside OpenAPI).

When `MERCADIA_STORE_EDGE_USE_HARDWARE_AGENT=true`, card payments and fiscalization send commands to
the local hardware-agent (`authorize`/`capture`, `print_receipt`) with mock fallback enabled by default.

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
shift requires an open operational day for the store. Closed shifts are removed from the
open-shift read model. A shift cannot be closed while it has unresolved receipts. Closing a
shift with `closingCashMinor > 0` requires final collection details and posts a
`drawer_to_safe` cash movement from the shift drawer to the selected safe. The final collection
requires two-person control through separate `actorId` and `approvedById` values.
Cash operations are modeled as an append-only ledger. Posted cash movements are not edited in
place; corrections must be represented by a new movement. The first control rule is separation
of duties: the actor posting a cash movement cannot approve the same movement.
Cash balances are a derived read model calculated from posted cash movements. The current
in-memory implementation intentionally allows negative balances because initial safe/drawer
opening balances are not modeled yet; production persistence should maintain the same derivation
or an auditable materialized read model backed by the ledger.
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
