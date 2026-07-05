# SCO Terminal Implementation Design (Spike)

This document maps the self-checkout (SCO/KSO) specification onto the current Store Edge
and Central Backend APIs, proposes an app architecture and auth model, and slices the build
into milestones. It is a design spike, not an implementation plan: no code changes are made
here. Each milestone in section 5 is meant to become its own ordinary implementation plan.

Sources: [Self-checkout specification](modules/self-checkout.md) (298 lines),
[ADR-0008](adr/0008-sco-cashless-first-stage.md), `contracts/openapi/store-edge.openapi.json`,
`contracts/openapi/central.openapi.json`, `frontend/apps/pos-terminal/`,
`frontend/apps/senior-cashier-terminal/`, `backend/services/store-edge/internal/app/rbac.go`,
`backend/services/store-edge/internal/domain/session.go`, `docs/open-questions.md`.

## 1. Scope and non-goals

Stage 1 SCO is cashless-only by decision, not by omission. ADR-0008
(`docs/adr/0008-sco-cashless-first-stage.md`, Accepted) restricts SCO to bank
card/terminal, QR/SBP, bonuses, and gift card where configured; cash stays POS-only. This
document does not design cash-on-SCO. Cash support is explicitly deferred to "a separate
hardware and cash-ledger project" per the ADR's consequences section.

Additional non-goals for this spike:

- **Real payment-terminal hardware drivers.** Which acquirer/terminal protocol SCO uses is
  an open question (`docs/open-questions.md`, "Платежи" section, items 1 and 4: acquirer and
  SBP/QR provider both "to be decided"). If a hardware support matrix exists under
  `docs/hardware-support-matrix.md` by the time this is read, cross-check its payment-terminal
  row against section 3's cashless-payment row below; it does not exist yet in this repo
  snapshot, so this document does not block on it.
- **Loyalty integrations.** The spec's loyalty flow (phone/QR/card identification, bonus
  balance, write-off) has zero corresponding backend surface today (see section 3, row 7).
  It is scoped out of the milestone list except as an optional stretch milestone.
- **Scaffolding the app itself.** No `frontend/apps/sco-terminal` directory, package.json, or
  component is created by this document.

## 2. App architecture

### New workspace app

`frontend/apps/sco-terminal` would join the pnpm workspace the same way `pos-terminal` and
`senior-cashier-terminal` already do — `frontend/pnpm-workspace.yaml` globs `apps/*`, so no
workspace config change is needed, only a new directory with its own `package.json`.

### Reuse table

| pos-terminal source | Reuse strategy | Notes |
|---|---|---|
| `src/api-client-config.ts` (`configureApiClients`, `getStoreId`, `getTerminalId`) | Copy, then diverge | SCO needs the same store/terminal identity plumbing but no cashier-session env vars; add a `VITE_SCO_LAYOUT_PROFILE` (`h`/`v`/`hd`) terminal-configuration variable per the spec's layout-variant requirement. |
| `src/query-client.ts` (3 lines, bare `QueryClient`) | Copy | Trivial; not worth extracting. |
| `src/i18n/config.ts` + `src/i18n/locales/{ru,en}.json` | Copy, then diverge | Same ru/en bootstrap (`i18next` + `react-i18next`), but SCO's copy is customer-facing (idle/help/error text) rather than staff-facing, so the locale JSON content itself must be written fresh, not copied. |
| `src/auth/credential.ts` (`readStaffCredential`, iButton/MSR/barcode via hardware-agent `listDevices`/`sendDeviceCommand`) | Copy | Directly reusable for assistant login on the SCO terminal or assistant station; hardware-agent contract is app-agnostic. |
| `src/auth/AuthProvider.tsx`, `src/auth/types.ts` (`canUsePosSession` gating on `cashier`/`senior_cashier`/`admin`) | Diverge, do not copy verbatim | SCO has no cashier login by default (see section 4); the assistant-login overlay needs its own role check once an assistant role exists, and the "no session" default state is customer-idle, not a login screen. |
| `src/lib/receipt-utils.ts` (`formatMinorAmount`, `filterGridByCategory`, `parseAmountToMinor`, `formatInputAmount`, `settledPaymentAmountMinor`) | **Move to a shared package** | These are pure functions with no POS-specific state. `senior-cashier-terminal` already has its own parallel `src/lib/cash-utils.ts`; a third from-scratch copy in `sco-terminal` would be the second duplication of the same amount-formatting logic. Extract to a new shared package (e.g. `frontend/packages/receipt-kit`) before `sco-terminal` is scaffolded. |
| Idempotency-key helpers `createIdempotencyKey`/`createIdempotencyHeaders` (inline in `src/Root.tsx:135-141`, not even extracted within pos-terminal itself) | **Move to a shared package** | Both pos-terminal and sco-terminal issue the same `Idempotency-Key` header convention on every mutating call (`openReceipt`, `scanReceiptLine`, `createReceiptPayment`, `createReceiptFiscalDocument`, heartbeat). Extract alongside `receipt-utils` rather than re-inlining a third time. |

### Design-system pieces that already anticipate SCO

`@mercadia/ui`'s theme types already model an SCO surface:
`frontend/packages/ui/src/theme/types.ts:1` defines
`Surface = 'admin' | 'terminal' | 'sco' | 'senior-cashier'` and `:5` defines
`AccentPreset = 'sale' | 'return' | 'sco' | 'neutral'`. pos-terminal already switches into
this surface today — `frontend/apps/pos-terminal/src/Root.tsx:434-439` calls
`applyTheme({ surface: template.kind === 'sco' ? 'sco' : 'terminal', ... })` whenever the
layout template's `kind` is `'sco'`. `LayoutGrid`, `Numpad`, `Tabs`, `Card`, `Badge`, `Field`
components apply directly to the idle, scanning, and product-grid screens without new
components.

### Layout profiles (H/V/HD) as terminal configuration

The central-backend `LayoutTemplate` schema (`contracts/openapi/central.openapi.json`,
`createLayoutTemplate`/`getLayoutTemplate`/`listLayoutTemplates` operations) has a `kind`
field and a `terminalType` field, both untyped strings with no enum — there is no dedicated
"layout profile" value today. Encoding horizontal/vertical/HD as terminal configuration (the
spec's explicit requirement) needs either a naming convention (e.g. `kind: "sco-horizontal"`)
or a schema addition; this is recorded as a gap-table row (section 3, row 5/1 combined) rather
than invented here.

Recommended architecture: one workflow/state machine
(idle → scanning → receipt → payment → done, matching the "current stage" list in the spec's
"Scanning Flow" section) driving three CSS layout shells selected by a `layoutProfile`
terminal-config value — mirroring how pos-terminal already branches its theme by
`template.kind`, just one more axis (profile) alongside surface.

## 3. Spec → API gap table

Every verdict below is derived from grepping `operationId` values out of
`contracts/openapi/store-edge.openapi.json` (66 operations total) and
`contracts/openapi/central.openapi.json`, then reading the referenced operation's request/
response schema inline in the file (both contracts define schemas inline per-operation, not
in `components/schemas`).

| # | Spec capability | Verdict | Operation(s) / evidence | Notes |
|---|---|---|---|---|
| 1 | Idle screen & terminal availability | Partial | `recordTerminalHeartbeat` (`POST /v1/terminals/{terminalId}/heartbeat`), `listStoreTerminals`, `getStoreMonitoringSummary` | Heartbeat reports `kind`/`softwareVersion`/`storeId` and returns `lastSeenAt`; there is no dedicated "in service / out of service" readiness endpoint (see row 14). |
| 2 | Barcode scan / add receipt line | Exists | `scanReceiptLine` (`POST /v1/receipts/{receiptId}/scan`), `addReceiptLine` (`POST /v1/receipts/{receiptId}/lines`) | Used as-is by pos-terminal (`Root.tsx` `scanProduct`). |
| 3 | DataMatrix marking validation | Partial | `validateReceiptMarking` (`POST /v1/receipts/{receiptId}/marking/validate`) | Request is `{code}`, response is `{valid, code, message, productId}` — no `lineId` in either direction, so "link DataMatrix code to the receipt line" and "prevent duplicate/invalid code usage" (spec requirements) are not modeled at the contract level. |
| 4 | Weighted goods lookup / scale integration | Missing | none | Zero "weigh"/"scale" hits anywhere in `store-edge.openapi.json`. Needs a new contract (line-level weight field, scale-reading command likely via hardware-agent, weight-vs-expected comparison result) before this milestone can start. |
| 5 | Product grid (produce/manual selection) | Exists for layout, missing for weight-aware tiles | `getLayoutTemplate`/`listLayoutTemplates` (central), tile schema has `label`/`color`/`productId`/`empty`/`categoryId`/`iconUrl` | Grid rendering is fully reusable (`LayoutGrid` + `filterGridByCategory`); tiles have no weight/scale fields, so produce-specific behavior ties back to row 4. |
| 6 | Quantity change / line delete | Missing | none | Only `addReceiptLine` and `applyReceiptLineDiscount` touch `/lines/{lineId}`-adjacent paths; there is no `removeReceiptLine`, `updateReceiptLineQuantity`, or void-line operation among the 66 operationIds. |
| 7 | Loyalty flow (phone/QR/card ID, balance, bonus write-off) | Missing | none | Zero "loyalty" hits in any of the three contract files (`store-edge`, `central`, `hardware-agent`). Full new backend surface. |
| 8 | Age verification (18+) | Missing | none | No age/restricted-goods operation or field found. Needs a dedicated pause-and-approve flow (block checkout, notify assistant, record identity/timestamp/result) — nothing today models this. |
| 9 | Assistant intervention actions (10 listed in spec: selective control, cancel receipt, remove item, price/discount override, confirm 18+, accept marking issue, manual entry, start return, print copy, block terminal) | Partial | `cancelReceipt` (`POST /v1/receipts/{receiptId}/cancel`), `applyReceiptLineDiscount`, `createReceiptReturn` (`POST /v1/receipts/{receiptId}/returns`) | 3 of 10 listed actions have a direct operation. "Remove item," "confirm 18+," "accept marking issue," "manual item entry," "block terminal," "print receipt copy" have none. |
| 10 | Selective control / full rescan / audit escalation | Missing | `listOperationJournal` (`GET /v1/stores/{storeId}/operation-journal`) exists but is read-only history | No operation starts, drives, or resolves a control workflow; the journal only records events after the fact. |
| 11 | Cashless payment (card, QR/SBP, bonuses, gift card per ADR-0008) | Partial | `createReceiptPayment` (`POST /v1/receipts/{receiptId}/payments`) | `method` is an untyped free string with no enum anywhere in the schema; pos-terminal only ever sends `'cash'` or `'card_mock'`. Whether the SCO-specific methods (`qr_sbp`, `bonus`, `gift_card` or similar) are accepted values is unverified against the contract. |
| 12 | Fiscalization | Exists | `createReceiptFiscalDocument` (`POST /v1/receipts/{receiptId}/fiscal-documents`), `listReceiptFiscalDocuments` | Identical shape to pos-terminal's flow; reusable as-is. |
| 13 | Receipt completion / reset (finish, abandon, return to idle) | Exists | `getReceipt`, `cancelReceipt` | "Finish" is a client-side state reset (as in pos-terminal's `finishSale()`); "abandon" maps to `cancelReceipt`. |
| 14 | Remove terminal from service / block | Missing | `getTerminal`/`listStoreTerminals` expose a free-string `status` field | No `PUT`/`PATCH`/mutation operation sets terminal `status`; heartbeat only carries `kind`/`softwareVersion`/`storeId`. No "reason" field anywhere. |
| 15 | Assistant station monitoring (zone/group, help queue, checks today, average check, audit %, response time, alerts) | Partial | `getStoreMonitoringSummary` (`GET /v1/stores/{storeId}/monitoring/summary`), `listStoreMonitoringTerminals` (`GET /v1/stores/{storeId}/monitoring/terminals`) | These supply store-wide revenue/terminal-count/attention-needed metrics — `senior-cashier-terminal`'s `MonitoringPage.tsx` already consumes the simpler `listStoreTerminals` for a similar per-terminal card view. Neither operation has zone/terminal-group scoping, a help queue, an audit-percentage field, or a response-time field. |
| 16 | Assistant/customer role model | Missing | `backend/services/store-edge/internal/app/rbac.go`, `internal/domain/session.go` | Only `RoleCashier`, `RoleSeniorCashier`, `RoleAdmin` exist (`session.go:11-13`). `docs/open-questions.md` line 45 records the intended role ("помощник на КСО" — SCO assistant) as part of "Роли и права" item 1, but it is not implemented. |
| 17 | Cash-on-SCO | **Out of scope by decision**, not a gap | — | ADR-0008 explicitly restricts SCO to cashless payments in stage 1; do not design for it. |

Summary: of 16 in-scope capabilities, 4 exist, 6 are partial, 6 are missing outright — the
majority of the spec's assistant-facing and compliance-sensitive behavior (age check,
weighing, loyalty, line editing, terminal-status control) has no backend contract today.

## 4. Auth & session model (proposal + open questions)

### Current constraints

`createAuthSession` (`POST /v1/auth/sessions`) requires `{actorId, pin, storeId}` — every
session belongs to one authenticated human actor. `openReceipt` (`POST /v1/receipts`)
requires `cashierId`. There is no anonymous or terminal-scoped session/receipt type in the
contract. This is a real constraint for a customer-facing terminal that, per the spec's idle
screen, must let a customer start scanning without any login.

### Proposal A — provisioned terminal-service actor (recommended, smaller diff)

Provision one long-lived service `Actor` per SCO terminal (e.g. `sco-terminal-{id}`) that
authenticates once at boot (like the heartbeat loop already does) and stays logged in for the
terminal's operating lifetime. Its session's `cashierId` becomes the receipt's `cashierId` for
every customer-initiated receipt on that terminal — the human customer is never an `Actor` in
this model.

This requires a new role (tentatively `sco_service`) with permissions scoped to
`receipts.open`/`receipts.scan`/`payments.create`/`fiscal.create` only — explicitly **not**
`PermissionReturnsCreate`, `PermissionDiscountApply`, or `PermissionRecountApprove` (the
existing `rolePermissions` map in `rbac.go:19-23`), consistent with ADR-0008's cashless-only,
no-discount-authority spirit for stage 1. This is recorded here as a proposal, not invented
silently in code — see open question 2 below.

### Proposal B — terminal-credential auth path (larger diff)

Extend `createAuthSession` (or add a new endpoint) to accept a terminal-scoped credential
(e.g. `terminalId` + a provisioned terminal secret) instead of `actorId` + `pin`, producing a
session whose identity is the terminal itself rather than an impersonated human actor. This
avoids inventing a fake "actor" per terminal but is a bigger contract change and touches
`createAuthSession`'s required-field set for every existing caller.

Recommend Proposal A for the smaller footprint, but this is explicitly an open question for
backend ownership (open question 1), not a decision made here.

### Assistant intervention

The assistant is a real human staff member who authenticates using the same credential
mechanism already implemented for POS: `readStaffCredential` in
`frontend/apps/pos-terminal/src/auth/credential.ts` reads iButton/MSR/barcode-card factors via
hardware-agent's `listDevices`/`sendDeviceCommand`, unchanged for SCO. The blocking gap is not
the mechanism — it's the missing role. `rbac.go` has no assistant role, so an assistant who
authenticates today would only ever be recognized as `cashier`/`senior_cashier`/`admin`
(whatever role their actor record already carries), not as an SCO-specific assistant with
SCO-specific permissions. Do not silently reuse `senior_cashier` as a stand-in; record the
missing role as backend work (gap-table row 16) and resolve it as part of milestone 3 (or
earlier, per open question 2).

Session lifetime: pos-terminal's model (`sessionStorage` + auto-lock timers,
`Root.tsx:610-653`) assumes exactly one human session per shift. SCO instead needs a
long-lived terminal-service session (Proposal A) that survives the whole operating day, with
an assistant's PIN/credential session layered temporarily **on top** during an intervention —
an overlay, not a replacement, so the customer's in-progress receipt is not disturbed by the
assistant logging in and back out.

Credential enrollment is unaffected either way: `addActorCredentialBinding`,
`setActorCredentialPolicy`, `getCredentialManagement` are already role-agnostic operations and
apply to an assistant actor without change.

## 5. Milestone slicing

**M1 — App scaffold + idle + scan + receipt view (existing APIs only).**
New `frontend/apps/sco-terminal` package; reuse `api-client-config`/`query-client`/i18n
bootstrap pattern (section 2); idle screen with language selector; barcode scan via
`scanReceiptLine`; live receipt view via `getReceipt`. Blocked on *some* answer to the
session question even for a spike — a hardcoded/demo service actor (mirroring pos-terminal's
own hardcoded `openedById` fallback) is an acceptable placeholder until Proposal A/B is
decided. Gap rows to close first: **16** (at least provisionally).

**M2 — Cashless payment + fiscalization (mirrors pos-terminal).**
`createReceiptPayment` against the cashless methods ADR-0008 lists; `createReceiptFiscalDocument`
unchanged from pos-terminal's flow. Gap rows to close first: **11** (agree on accepted
`method` values for SCO); row 12 is already closed.

**M3 — Assistant mode + assistant station.**
Assistant login overlay (needs the real role from row 16, not a placeholder); the 3 of 10
intervention actions that already have operations (cancel, discount, return) go live;
the remaining 7 are surfaced as visibly disabled with a staff-readable reason, per the spec's
own error-handling philosophy ("staff screens should expose detailed diagnostics"), until
backend closes them. Gap rows to close first: **9, 10, 14, 15, 16**.

**M4 — Marking, weighted goods, age verification.**
DataMatrix pause-flow against `validateReceiptMarking` once line-linkage is added (row 3);
weighted-goods flow once a scale contract exists (row 4); age-verification pause/approve flow
once a dedicated operation exists (row 8). Gap rows to close first: **3, 4, 8**.

**M5 — Loyalty integration (stretch, optional).**
Entirely new backend surface (row 7). Deliberately last: ADR-0008's stage-1 payment list does
not require loyalty to ship SCO, and the spec itself says "the customer must be able to
continue without loyalty."

## 6. Open questions

1. **(backend)** Which session/actor model authenticates an unattended SCO terminal to Store
   Edge — a provisioned service actor with a new `sco_service` role (Proposal A), or a new
   terminal-credential auth path (Proposal B)? See section 4.
2. **(backend)** Should a formal SCO-assistant role be added to `rbac.go`/`session.go` now, or
   deferred until M3 starts? `docs/open-questions.md` line 45 records the intent ("помощник на
   КСО") but no scope or permission list.
3. **(product)** What is the definitive first-stage cashless `method` value list for SCO (card,
   QR/SBP, bonuses, gift card per ADR-0008), and should `createReceiptPayment`'s `method` field
   become an enum rather than a free string?
4. **(hardware)** What device/protocol reads a DataMatrix code on SCO hardware, and does it
   route through hardware-agent (like credential reads) or a new device class? Feeds rows 3
   and 4.
5. **(hardware)** What scale hardware and protocol are used for weighted goods on SCO, and
   where does the abstraction live (hardware-agent vs. a new service)? Cross-check against a
   future hardware support matrix's payment-terminal and scale rows once it exists.
6. **(product)** What triggers/rules define "selective control" (row 10) — percentage-random,
   risk-score, or manual assistant trigger — and who configures it?
7. **(backend/product)** Does "remove KSO from service" (row 14) need a new terminal-status
   mutation endpoint, or can it be modeled as a client-side state change plus an audit entry
   via `listOperationJournal`?
8. **(product)** Is loyalty (row 7, M5) truly deferrable, or does the business need it inside
   M1-M4 rather than as a stretch milestone?

## Maintenance notes

Each milestone above is meant to become one ordinary implementation plan; gap-table rows
marked "missing" become store-edge (or hardware-agent) backlog items before their dependent
milestone starts. When a hardware support matrix lands elsewhere in `docs/`, its
payment-terminal and scale rows must agree with rows 11 and 4 here — reviewers should
cross-check both documents before either milestone (M2, M4) starts implementation.
