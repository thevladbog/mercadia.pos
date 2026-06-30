<!-- DECISION-STORE-EDGE-LOCAL-AUTHORITY -->

## Decision: Store Edge owns store-local operational state

**Status**: Active
**Date**: 2026-06-30
**Severity**: Critical

**Files**:

- `backend/services/store-edge/**`
- `contracts/openapi/store-edge.openapi.json`
- `frontend/packages/api-clients/store-edge/**`

### Context

ADR-0001 requires every store to run Store Edge as the authority for store-local operational state.
POS, SCO/KSO, senior cashier, assistant, and local admin clients must submit operational commands
to Store Edge rather than owning final business truth locally or depending synchronously on central
services for store-local operations.

Reference: `docs/adr/0001-store-edge-per-store.md`

---

<!-- DECISION-HARDWARE-AGENT-DEVICE-BOUNDARY -->

## Decision: Hardware Agent is the only layer that talks to local devices

**Status**: Active
**Date**: 2026-06-30
**Severity**: Critical

**Files**:

- `backend/services/hardware-agent/**`
- `frontend/apps/pos-terminal/**`
- `frontend/apps/sco-terminal/**`
- `frontend/apps/senior-cashier-terminal/**`
- `frontend/apps/assistant-station/**`
- `frontend/packages/api-clients/hardware-agent/**`
- `contracts/openapi/hardware-agent.openapi.json`

### Context

ADR-0002 keeps fiscal devices, payment terminals, scanners, scales, drawers, printers, MSR,
iButton, and vendor SDK access behind the local Hardware Agent. Terminal UI code can use the
Hardware Agent API, but must not call device SDKs or protocols directly.

Reference: `docs/adr/0002-terminal-app-and-hardware-agent.md`

---

<!-- DECISION-OPENAPI-SCALAR-ORVAL-CONTRACTS -->

## Decision: Public HTTP APIs flow through generated OpenAPI and Orval clients

**Status**: Active
**Date**: 2026-06-30
**Severity**: Warning

**Files**:

- `backend/services/*/internal/api/**`
- `backend/packages/platform/httpapi/**`
- `contracts/openapi/**`
- `frontend/packages/api-clients/**`
- `frontend/apps/**/src/api*.ts`
- `frontend/apps/**/src/**/*api*.ts`

### Context

ADR-0009 requires public HTTP APIs to be generated from Go API definitions into OpenAPI, served
with Scalar, and consumed by TypeScript frontends through Orval-generated clients and types. API
handler changes should regenerate OpenAPI and affected generated clients; frontend code should not
hand-write duplicate DTOs when generated types exist.

Reference: `docs/adr/0009-openapi-scalar-orval.md`

---

<!-- DECISION-CASH-LEDGER-SEPARATION-OF-DUTIES -->

## Decision: Cash ledger entries are immutable and SoD is server-side

**Status**: Active
**Date**: 2026-06-30
**Severity**: Critical

**Files**:

- `backend/services/store-edge/internal/app/*cash*`
- `backend/services/store-edge/internal/domain/*cash*`
- `backend/services/store-edge/internal/api/*cash*`
- `frontend/apps/admin-web/src/pages/*cash*`
- `frontend/apps/admin-web/src/pages/Cash*`
- `frontend/apps/senior-cashier-terminal/src/pages/*Cash*`
- `frontend/apps/senior-cashier-terminal/src/pages/Safe*`
- `frontend/apps/senior-cashier-terminal/src/pages/Bank*`

### Context

ADR-0006 models cash as immutable ledger movements. Posted movements are not edited; corrections
are separate operations. Critical operations must enforce separation of duties server-side, not only
in UI controls.

Reference: `docs/adr/0006-cash-ledger-and-separation-of-duties.md`

---

<!-- DECISION-PAYMENT-FISCAL-SEPARATE-STATES -->

## Decision: Payment and fiscalization are separate state machines

**Status**: Active
**Date**: 2026-06-30
**Severity**: Critical

**Files**:

- `backend/services/store-edge/internal/app/*payment*`
- `backend/services/store-edge/internal/app/*fiscal*`
- `backend/services/store-edge/internal/domain/*payment*`
- `backend/services/store-edge/internal/domain/*fiscal*`
- `backend/services/store-edge/internal/api/*payment*`
- `backend/services/store-edge/internal/api/*fiscal*`
- `frontend/apps/pos-terminal/src/**/*payment*`
- `frontend/apps/pos-terminal/src/**/*fiscal*`
- `frontend/apps/sco-terminal/src/**/*payment*`
- `frontend/apps/sco-terminal/src/**/*fiscal*`

### Context

ADR-0007 requires payment and fiscalization to be modeled as separate state machines. Payment
success does not mean a receipt is legally complete; fiscalization failures need explicit retry,
correction, or manual resolution paths.

Reference: `docs/adr/0007-fiscalization-payment-state-machines.md`

---

<!-- DECISION-STORE-ADMIN-CENTRAL-ADMIN-SCOPES -->

## Decision: Store admin and central admin authority stay separate

**Status**: Active
**Date**: 2026-06-30
**Severity**: Warning

**Files**:

- `frontend/apps/admin-web/**`
- `backend/services/store-edge/internal/api/**`
- `backend/services/central-backend/internal/api/**`
- `frontend/packages/api-clients/store-edge/**`
- `frontend/packages/api-clients/central/**`

### Context

ADR-0005 separates store-local admin authority from central admin authority. Store admin operations
go through Store Edge for local operational state; central admin handles network-level configuration,
reporting, and policy. UI and API changes should keep command targets and permission scopes explicit.

Reference: `docs/adr/0005-store-admin-and-central-admin.md`

---
