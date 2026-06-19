# Backend

This directory contains the Mercadia POS backend workspace.

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

Command endpoints require `Idempotency-Key`. Reusing the same key for the same command returns
the same result; reusing it with a different command payload returns an idempotency conflict.

The current persistence adapter is in-memory and intended for the first skeleton only. PostgreSQL
repositories and migrations should replace it when the Store Edge persistence slice starts.
The current catalog contains demo seed products so local scan workflows can be exercised before
the catalog sync slice exists.
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
