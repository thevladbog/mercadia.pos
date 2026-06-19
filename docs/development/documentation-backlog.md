# Documentation Backlog

This document tracks important topics that are not yet described deeply enough.

## Highest Priority

### 54-FZ And Fiscalization Rules

Needed:

- Legal receipt lifecycle for Russia.
- ATOL integration behavior.
- Payment captured but fiscalization failed.
- Fiscal correction scenarios.
- Offline fiscal behavior.
- Receipt delivery through OFD, email, SMS, print, and optional Mercadia delivery.

### Marking / Chestny ZNAK

Needed:

- Product categories requiring DataMatrix.
- Online/offline validation rules.
- Supervisor override legality.
- Failure states and customer/cashier UX.
- Reconciliation against marking service.

### Payment Provider Baseline

Needed:

- First acquirer.
- Payment terminal protocol.
- Card same-day cancellation.
- Refund references and RRN/authorization requirements.
- QR/SBP provider and timeout behavior.
- Split payment edge cases.

### Cash Operation Catalogue

Needed:

- Final list of cash movement types.
- Which operations are created by cashier vs senior cashier vs admin.
- Which operations require two signatures.
- Correction movement rules.
- Required printed/accounting documents.

### RBAC Matrix

Needed:

- Final role list.
- Permissions by module.
- Critical operation approvals.
- Store vs central scopes.
- Temporary staff and trainee flags.
- iButton/MSR issuance and revocation rules.

## Medium Priority

### API Standards

Needed:

- Endpoint naming conventions.
- Error schema.
- Idempotency key format.
- Pagination/filtering/sorting conventions.
- WebSocket/SSE event schemas.
- API versioning rules.

### Event Standards

Needed:

- Event envelope.
- Event naming.
- Event versioning.
- Correlation/causation IDs.
- Store/terminal/session metadata.
- Retention and replay rules.

### Data Model

Needed:

- Conceptual ERD.
- Cash ledger schema.
- Receipt/payment/fiscal schema.
- Terminal/session schema.
- Product/catalog/versioning schema.

### Testing Strategy

Needed:

- Unit/integration/E2E boundaries.
- Hardware simulator strategy.
- Fiscal/payment provider simulator strategy.
- Contract testing.
- Offline/sync testing.
- Regression requirements for money/fiscal bugs.

### Deployment And Operations

Needed:

- Store Edge packaging.
- Hardware Agent packaging.
- PostgreSQL install/backup/restore.
- Terminal update mechanism.
- Local diagnostics.
- Central observability.

## Lower Priority

### Frontend Design System

Needed:

- Component inventory.
- Density/layout rules.
- Form/table/drawer/modal patterns.
- Touch terminal ergonomics.
- Accessibility baseline.

### Reporting

Needed:

- Production acceptance report list.
- Report data sources.
- Export formats.
- Retention policy.

### AI Agent Development

Needed:

- Repo structure once code exists.
- Build/test command map.
- Common implementation recipes.
- Known generated folders.
- Agent-safe refactoring boundaries.
