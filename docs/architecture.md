# Architecture Proposal

## Agreed Direction

The current architectural direction is:

- Every store has a mandatory Store Edge runtime.
- POS, SCO/KSO, senior cashier, assistant, and local admin clients work through the Store Edge
  for operational state and through a local Hardware Agent for devices.
- Terminal UI should be implemented as a local desktop/web hybrid, with React + Tauri as the
  preferred UI shell candidate.
- The Hardware Agent should be implemented in Go and expose a stable local API for fiscal
  devices, payment terminals, scanners, scales, drawers, receipt printers, iButton readers,
  and other peripherals.
- Store Edge should use PostgreSQL for local operational state, cash ledger, receipt ledger,
  terminal state, outbox/inbox, and idempotency.
- The platform should include a message broker from the start, but avoid heavyweight Kafka-like
  infrastructure for the first architecture. NATS JetStream or RabbitMQ are better candidates;
  NATS JetStream is the preferred default because it is lightweight, Go-friendly, and suitable
  for edge/central event streaming.
- The system should use an outbox pattern even when a broker is present. The database remains
  the transactional source for local commands; the broker is the delivery fabric.
- Store admin and central admin are separate scopes. They may share UI/components and backend
  modules, but they must not share the same operational authority.

## Architecture Goals

Mercadia POS needs to combine fast local checkout with reliable centralized control. The
architecture should optimize for:

- Low-latency scanning and cart calculation.
- Store-level resilience during temporary backend outages.
- Strong auditability for money, returns, discounts, and approvals.
- Consistent business rules across POS, SCO, senior cashier terminal, and admin panel.
- Clear integration boundaries for payment, fiscalization, marking, loyalty, ERP, and bank/accounting.
- Real-time monitoring without coupling UI directly to terminal internals.
- Safe synchronization and replay of operational events.

## Recommended Deployment Topology

### Central Cloud / Data Center

Central services own master data, cross-store reporting, global configuration, user directory,
integration management, long-term audit, and consolidated reconciliation.

Central services also receive synchronized operational events from every Store Edge. They should
not be required for every local checkout command, but they are the source for cross-store
analytics, global administration, long-term storage, and integration with central business
systems.

### Store Edge

Each store must run a local edge service. This service acts as the store's operational hub:

- Product and price cache.
- Local receipt ledger.
- Local cash ledger.
- Local fiscal operation ledger.
- Local payment operation ledger.
- Terminal registry and heartbeats.
- WebSocket/event gateway for live monitoring.
- Command API for POS, SCO, senior cashier, assistant station, and store admin.
- Integration proxy for store-local services where required.
- Outbox/inbox sync with central services.
- Lightweight broker bridge for local/central event delivery.

The store edge lets POS and SCO continue operating when the central backend is temporarily
unavailable, subject to legal and payment/fiscal constraints.

### Terminals

Terminals run local clients:

- POS client.
- SCO client.
- Senior cashier client.
- Assistant station client.
- Admin web client.

Clients should not own final business truth. They should maintain local UI state and submit
commands to the store edge or central backend. For scanner speed, clients may maintain a local
read cache of products/prices/layouts.

The preferred terminal app shape is:

- React for UI.
- Tauri for desktop packaging, kiosk behavior, local window management, and OS integration.
- Go Hardware Agent as a local service installed next to the app.
- Store Edge API for business commands and operational state.

The terminal client should not call fiscal, payment, or peripheral SDKs directly. Device access
goes through the Hardware Agent, and business state goes through the Store Edge.

Senior cashier is intentionally split into two surfaces:

- Senior cashier web interface: browser-based monitoring and management without direct local
  hardware access.
- Senior cashier touch terminal: local Tauri-based terminal for money operations requiring
  MSR, iButton, physical signing, and device-backed authentication.

Both surfaces use the same Store Edge commands, permissions, and audit model. The touch terminal
has additional Hardware Agent capabilities and should be the primary surface for operations that
require physical presence or local device confirmation.

### Hardware Agent

The Hardware Agent is a local Go service responsible for all terminal-connected devices.

Responsibilities:

- Normalize vendor SDKs and device protocols.
- Provide a stable local HTTP/gRPC API to terminal apps and Store Edge.
- Support Windows and Linux where device vendors allow it.
- Report health, version, and diagnostics.
- Own device-level retries and low-level error mapping.
- Keep device-specific implementation out of UI and business domains.
- Support staff credential devices such as iButton readers and MSR readers.

The first fiscal integration target is ATOL. Cash drawer control is expected to happen through
the fiscal registrar. Receipt printer support should use Epson as the baseline. Scanners should
support COM first and USB/HID where practical.

## High-Level Component Model

```text
POS/SCO/Senior/Admin Clients
        |
        | HTTPS/WebSocket to Store Edge
        | localhost HTTP/gRPC to Hardware Agent
        v
Terminal Hardware Agent
        |
        | device SDKs/protocols
        v
ATOL, payment terminals, scanners, scales, drawers, printers, iButton

POS/SCO/Senior/Admin Clients
        |
        | operational commands/events
        v
Store Edge API and Event Gateway
        |
        | local commands, events, cache, outbox
        v
Store PostgreSQL + Lightweight Broker Bridge
        |
        | Sync
        v
Central Platform Services + Broker
        |
        | Integrations
        v
ERP, Payments, Fiscal, Marking, Loyalty, Gift Cards, Bank/Accounting
```

## Bounded Contexts

### Identity And Access

Owns users, roles, permissions, sessions, PIN policy, iButton binding, staff cards, and
authentication logs.

Key entities:

- User.
- Role.
- Permission.
- Credential factor.
- Session.
- Login attempt.

### Store Operations

Owns operational day, terminal registry, terminal status, heartbeats, shift state, and
store-level monitoring.

Key entities:

- Store.
- Operational day.
- Terminal.
- Terminal session.
- Shift.
- Alert.

### Checkout

Owns receipt lifecycle, cart calculation, product scan behavior, discounts, taxes, returns,
and cancellation before fiscalization.

Key entities:

- Receipt.
- Receipt line.
- Discount.
- Tax line.
- Return.
- Cancellation.

### Payments

Owns payment attempts, split payments, payment status, reversals, refunds, and provider
references.

Key entities:

- Payment.
- Payment method.
- Payment provider.
- Authorization.
- Refund.
- Reversal.

### Cash Office

Owns drawer/safe balances, cash movements, denomination breakdown, cash recounts, bank
collection, and business expenses.

Key entities:

- Cash container.
- Cash movement.
- Denomination count.
- Safe recount.
- Bank collection.
- Cash document.

### Fiscalization

Owns fiscal receipt requests, legal receipt states, correction receipts, printer/fiscal
device communication, and retry/error state. For the first target market this context must be
designed around Russian 54-FZ requirements and ATOL fiscal devices.

Key entities:

- Fiscal document.
- Fiscal operation.
- Fiscal device.
- Fiscal error.

### Catalog And Pricing

Owns products, barcodes, prices, tax categories, restrictions, layouts, and catalog versions.

Key entities:

- Product.
- Barcode.
- Price.
- Tax category.
- Restriction.
- Layout template.
- Catalog version.

### Loyalty And Promotions

Owns customer lookup, loyalty account state, bonus accrual/write-off, customer tier, and
promotion application.

Key entities:

- Customer.
- Loyalty account.
- Bonus transaction.
- Promotion.

### Marking And Compliance

Owns DataMatrix validation, marked goods state, age-restricted goods, and compliance overrides.
Marked-goods behavior must support Russian Chestny ZNAK requirements. Whether a sale can
continue when marking services are unavailable is a legal/product decision and must be kept
configurable by policy.

Key entities:

- Marking code.
- Marking validation.
- Age verification.
- Compliance override.

### Reconciliation

Owns cross-source comparisons, mismatch detection, resolution status, and EoD readiness checks.

Key entities:

- Reconciliation run.
- Reconciliation source.
- Mismatch.
- Resolution.

### Audit

Owns append-only operational event history and reportable audit trails.

Key entities:

- Audit event.
- Actor.
- Signature.
- Attachment.
- Export package.

## Command/Event Pattern

The platform should use command handling plus append-only domain events for sensitive flows.
This is not full event sourcing for every screen. It is a pragmatic operational event model:
commands update transactional tables, write immutable audit/domain events, and publish those
events through an outbox.

Example POS sale:

1. `OpenReceipt` command.
2. `ReceiptOpened` event.
3. `AddReceiptLine` command.
4. `ReceiptLineAdded` event.
5. `SelectPaymentMethod` command.
6. `PaymentAttemptCreated` event.
7. `PaymentCaptured` event.
8. `FiscalizeReceipt` command.
9. `ReceiptFiscalized` event.
10. `ReceiptCompleted` event.

This model supports audit, replay for monitoring, local sync, and downstream reconciliation.

Not every read model needs full event sourcing. A pragmatic hybrid is recommended:

- Append-only event log for audit and synchronization.
- Relational tables for current state and queries.
- Outbox table for reliable publishing.
- Lightweight broker for asynchronous delivery after the local transaction commits.

The command handler must enforce permissions, idempotency, and business invariants before
state changes are committed.

## Data Storage

### Store Edge Database

Use PostgreSQL as the default transactional relational database at the Store Edge. It must
support:

- Receipt state.
- Cash ledger.
- Fiscal operation state.
- Payment operation state.
- Terminal state.
- Local catalog cache.
- Outbox/inbox queues.
- Idempotency records.
- Sync checkpoints.

SQLite is not recommended as the primary Store Edge database because the store has multiple
POS terminals, SCO terminals, a senior cashier terminal, assistant station, store admin views,
monitoring, and EoD workflows. SQLite may still be useful for terminal-local UI cache.

### Message Broker

Use a lightweight broker from the start, but keep it secondary to transactional persistence.

Preferred direction:

- NATS JetStream as the default candidate for central and optional local event streaming.
- PostgreSQL outbox/inbox as the source of reliable command/event publication.
- Broker consumers are idempotent.
- Store Edge can continue local operations if the central broker is unavailable.

RabbitMQ remains an acceptable alternative if operations prefer it. Kafka/Redpanda should not
be introduced for the initial architecture unless scale or platform constraints later justify it.

### Central Database

Use relational storage for operational truth plus analytical storage for reporting.

Suggested split:

- PostgreSQL for transactional central services.
- NATS JetStream or RabbitMQ for service/event delivery.
- Object storage for documents, receipt PDFs, exported reports, and attachments.
- Data warehouse/lake for analytics and long-term reporting.

## API Strategy

### Client APIs

Expose explicit command endpoints instead of generic CRUD for operational flows.

API contracts are generated from Go API definitions into OpenAPI. The OpenAPI documents are
rendered with Scalar for interactive documentation and consumed by Orval to generate frontend
TypeScript clients, types, hooks, mocks, and validators.

Examples:

- `POST /receipts`
- `POST /receipts/{id}/lines`
- `POST /receipts/{id}/payments`
- `POST /receipts/{id}/cancel`
- `POST /cash-movements`
- `POST /cash-movements/{id}/confirm`
- `POST /terminals/{id}/heartbeat`
- `POST /sco/{id}/assistant-actions`

All command endpoints should require idempotency keys.

Contract rules:

- Every endpoint has a stable `operationId`.
- Every command accepts an idempotency key.
- Error responses use a shared schema.
- Generated OpenAPI is validated in CI.
- Frontend code uses Orval-generated clients by default.

### Real-Time APIs

Use WebSocket or Server-Sent Events for:

- Terminal heartbeat stream.
- Receipt status updates.
- Payment status updates.
- SCO help queue.
- Assistant station terminal cards.
- Admin monitoring.
- EoD progress.

### Integration APIs

Provider-specific adapters should hide external protocol differences behind internal ports:

- Payment adapter.
- Fiscal adapter.
- Marking adapter.
- Loyalty adapter.
- Gift card adapter.
- ERP/catalog adapter.
- Bank/accounting adapter.

Adapters should be isolated so failed external dependencies do not break unrelated checkout
capabilities.

### Store Admin And Central Admin APIs

Admin APIs should be split by authority:

- Store admin APIs operate against the Store Edge and can manage store-local state: monitoring,
  operational day, safe, local terminals, local cash movements, and store-level configuration.
- Central admin APIs operate against central services and manage network-level state: global
  catalog policy, central reporting, cross-store users, integrations, themes, templates, and
  aggregated reconciliation.

Some UI screens may look identical across scopes, but the command target and permission model
must be explicit.

## Offline And Synchronization Strategy

### Operating Modes

- `online`: central and store edge are reachable.
- `central_offline`: terminals can reach store edge, but store edge cannot reach central.
- `edge_degraded`: terminal has limited local fallback or cannot reach store edge.
- `blocked`: legal/operational requirements prevent continued checkout.

### What Can Continue Locally

Subject to business/legal confirmation:

- Scanning products from cached catalog.
- Calculating totals from cached price version.
- Cash payment.
- Draft receipt creation.
- Cash drawer service operations.
- Local monitoring inside the store.

### What May Need Online Confirmation

- Bank card authorization.
- QR/SBP payment.
- Fiscalization.
- Marking validation.
- Loyalty bonus write-off.
- Gift card balance usage.
- B2B credit limit check.

If an operation cannot legally or financially proceed offline, the UI must show a controlled
block or fallback path.

Current product decisions:

- SCO/KSO accepts only cashless payment in the first stage.
- QR/SBP always requires online confirmation.
- Same-day card cancellation must be supported.
- Split payments are required.
- B2B is represented as a payment method in POS, but invoices and settlement live in an
  external system.

Open legal/compliance decisions:

- Exact offline fiscalization behavior under 54-FZ.
- Whether marked goods can be sold when marking validation is unavailable.
- Exact fiscal correction flow after payment succeeds but fiscalization fails.

### Sync Rules

- Every command has an idempotency key.
- Every domain event has globally unique ID, store ID, terminal ID, and monotonic local sequence.
- Store edge writes local transaction and outbox atomically.
- Sync worker sends events to central platform.
- Central platform deduplicates by event ID and idempotency key.
- Conflicts are resolved by workflow-specific rules, never by blind last-write-wins for money.

## Terminal Hardware Integration

### POS Hardware

Target devices:

- Barcode scanner.
- Cash drawer.
- Customer display.
- Receipt printer.
- Fiscal registrar/device.
- Payment terminal.
- Scales.
- iButton reader.
- MSR reader for staff magnetic cards.
- Staff card reader.

Use a local hardware abstraction layer:

- Client calls the local Hardware Agent or Store Edge hardware facade.
- Device agent normalizes vendor SDK calls.
- Device state is reported to terminal and admin monitoring.
- Device errors are mapped to user-safe messages and technical diagnostics.

Initial assumptions:

- ATOL is the first fiscal device family.
- Cash drawer is controlled by the fiscal registrar.
- Scanner support should prioritize COM and also support USB/HID where possible.
- Epson is the baseline receipt printer family.
- iButton reader model is still open and must be confirmed.
- MSR reader model/protocol is still open and must be confirmed.

### SCO Hardware

Target devices:

- Barcode scanner.
- Payment terminal.
- Receipt printer.
- Scale or weight sensor.
- Status light.
- Optional camera/video link.

SCO must report health for each critical device and block customer flow when a required
device is unavailable.

SCO cash hardware is out of scope for the first stage.

## Fiscalization

Fiscalization must be modeled as a separate state machine:

- `not_required_yet`.
- `pending`.
- `sent_to_device`.
- `fiscalized`.
- `failed_retryable`.
- `failed_blocking`.
- `correction_required`.

Payments and fiscalization must be coordinated carefully:

- Avoid fiscalizing unpaid receipts.
- Avoid leaving captured payments without legal receipt.
- Support retry and correction workflows.
- Preserve all intermediate states for audit.

Exact fiscal rules must be confirmed for the target jurisdiction and fiscal device provider.
For Russia, the fiscal behavior must be validated against 54-FZ and ATOL integration behavior.

## Payment State Machine

Suggested states:

- `created`.
- `sent_to_provider`.
- `awaiting_customer`.
- `authorized`.
- `captured`.
- `declined`.
- `cancelled`.
- `expired`.
- `reversed`.
- `refunded`.
- `failed_unknown`.

Payment operations must be idempotent and reconcileable. Provider references, RRN,
authorization codes, and terminal IDs should be mandatory when provided by the processor.

Additional payment decisions:

- Card same-day cancellation is required.
- Refund to the same card should use original authorization/reference data where the provider
  supports it.
- Refund to a different method than the original payment is a critical operation and requires
  supervisor approval.
- QR/SBP cannot be treated as offline-capable.
- Preauthorization is not assumed for MVP unless a selected acquirer requires it.

## Cash Ledger Model

Cash should be represented as ledger movements between containers.

Cash containers:

- POS drawer.
- SCO cash module, if cash SCO is ever supported.
- Store safe.
- Bank collection bag.
- Expense/outgoing counterparty.

Movement examples:

- Safe to drawer: change fund.
- Drawer to safe: revenue collection.
- Safe to bank: bank collection.
- Safe to expense: business expense.
- Drawer recount adjustment: discrepancy operation.

Every posted movement changes balances only through immutable ledger entries. Corrections are
new movements.

Operational rules:

- Cash in and cash out are created by the cashier and confirmed by the senior cashier.
- Regular cash in/out confirmation is not editable by the senior cashier; mismatch is handled
  through a correction path.
- Cashier final collection is a cashier-shift close operation and includes withdrawal of all
  cash from the drawer.
- Safe operations are initiated from admin/senior cashier surfaces.
- Safe recount with discrepancy requires two-person control.
- Russian denominations must be supported; coins can be entered by denomination and also through
  an "other coins amount" fallback field.

## Security

### Authentication

- Cashier: personnel ID/PIN and optional token.
- Senior cashier touch terminal: personnel ID/PIN/iButton, with MSR staff card as a configurable
  alternative or additional factor.
- Senior cashier web interface: browser authentication; hardware-backed operations must be
  routed to the touch terminal or use an approved browser-compatible factor.
- SCO assistant: staff authentication when entering assistant mode.
- Admin: strong account authentication, preferably SSO/MFA.

### Authorization

Use server-side RBAC with fine-grained permissions. UI state is only convenience and must not
be trusted as enforcement.

Authorization must also enforce separation of duties:

- A user cannot approve their own critical operation.
- If a supervisor is logged in as a cashier on POS, they are acting as the cashier for that
  session and cannot approve that session's critical operations.
- A store manager cannot edit already posted senior-cashier decisions; corrections must be new
  documents/operations.

### Sensitive Data

- Mask card numbers.
- Store payment tokens, not raw PAN.
- Encrypt credentials and integration secrets.
- Protect iButton/staff token identifiers.
- Protect MSR/staff card identifiers.
- Log access to personal/customer data.

## Observability

Collect:

- Terminal heartbeat and software version.
- Device health.
- Command latency.
- Payment/fiscal/marking/loyalty provider latency and failure rates.
- Outbox queue depth.
- Sync lag.
- Receipt exceptions.
- Cash mismatches.
- EoD blocking issues.

Monitoring should support store dashboards and central operations dashboards.

Store Edge should expose a local health dashboard and machine-readable health endpoint covering:

- PostgreSQL availability.
- Broker bridge status.
- Outbox/inbox lag.
- Hardware Agent reachability.
- Fiscal device status.
- Payment terminal status.
- Central sync status.

## Testing Strategy

### Unit Tests

- Cart calculation.
- Discounts and taxes.
- Payment state transitions.
- Cash ledger posting.
- Permission checks.
- Marking and age-restriction rules.

### Integration Tests

- Payment provider simulators.
- Fiscal device simulator.
- Marking service simulator.
- Loyalty/gift card simulator.
- Store edge sync.
- Terminal heartbeat and monitoring.

### End-To-End Tests

- Standard POS sale.
- Split payment sale.
- SCO sale with loyalty.
- SCO age verification.
- DataMatrix product sale.
- Return with receipt.
- No-receipt return approval.
- Cash in/out.
- Cashier shift close.
- EoD close.
- Network outage and recovery.

### Hardware-In-The-Loop Tests

Before production, test real scanners, fiscal devices, payment terminals, printers, drawers,
scales, and iButton readers.

## Suggested Implementation Slices

### Slice 1: Core Store Checkout

- Store edge skeleton.
- PostgreSQL schema and migration flow.
- Outbox/inbox foundation.
- Lightweight broker integration.
- POS sale screen.
- Catalog cache.
- Receipt lifecycle.
- Cash and card mock payments.
- Basic fiscal mock.
- Terminal heartbeat.
- Go Hardware Agent skeleton with mock devices.

### Slice 2: Cash Office

- Cash drawer ledger.
- Cash in/out.
- Senior cashier login.
- Change fund.
- Receive cash.
- Operation journal.

### Slice 3: SCO Customer Flow

- Idle/scanning/cart/payment/success.
- Assistant station basics.
- Help request.
- Assistant mode actions.

### Slice 4: Admin Monitoring And Devices

- Real-time terminal monitoring.
- Terminal registration.
- Device health.
- Event log.

### Slice 5: Compliance And EoD

- Returns.
- Marking.
- Age verification.
- 54-FZ/ATOL fiscal hardening.
- Safe recount.
- Bank collection.
- EoD blockers.
- Reconciliation.

### Slice 6: Configuration And Templates

- Users/roles.
- Products/prices.
- Layout templates.
- Receipt templates.
- Payment methods.
- Themes/franchise branding.

## Architectural Risks

- Fiscalization and payment ordering can create hard-to-resolve edge cases.
- Offline mode may be legally constrained for marked goods, fiscal receipts, and payments.
- Cash movement workflows require strong immutability and dual-signature clarity.
- SCO assistant actions share a receipt with customer actions and need careful audit separation.
- Product/price changes must be versioned so receipts remain explainable later.
- Integration failures must not cascade into full store outage.
- Hardware SDKs may force OS-specific decisions.

## Recommended Near-Term Decisions

- Confirm Store Edge packaging and operational ownership.
- Confirm React + Tauri + Go Hardware Agent as the terminal implementation stack.
- Confirm fiscal and payment providers.
- Confirm whether local offline fiscal/payment is required or allowed.
- Define the first legally valid MVP receipt lifecycle.
- Define RBAC matrix before implementing admin configuration.
- Define event schema and idempotency standard early.
- Choose NATS JetStream vs RabbitMQ as the initial broker.
- Confirm ATOL model/protocol and iButton reader model.
