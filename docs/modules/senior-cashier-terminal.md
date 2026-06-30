# Senior Cashier Terminal Specification

## Purpose

Senior cashier functionality is available through two interfaces:

- Senior cashier web interface.
- Senior cashier touch terminal with local hardware.

Both interfaces operate on the same Store Edge state and permission model, but they have
different device capabilities and ergonomic goals.

The senior cashier touch terminal is the operational money-control workstation for the store.
It gives elevated access to cash drawer movements, safe operations, cashier shift closure,
SCO support, approvals, and EoD preparation. It is designed for touch use and hardware-backed
authentication/signing.

The senior cashier web interface is used for browser-based monitoring and management where no
local hardware interaction is required. It can expose the same operational data and many of the
same workflows, but operations requiring physical authentication factors or local devices must
be delegated to a hardware-capable terminal or require an alternative approved flow.

## Interfaces

### Senior Cashier Web Interface

The web interface runs in a browser and is intended for desks, back-office workstations, or
store admin contexts.

It supports:

- Monitoring cashiers, POS terminals, and SCO terminals.
- Reviewing safe state and cash movement history.
- Reviewing EoD readiness.
- Reviewing operation journal.
- Preparing or reviewing operations that do not require local hardware.
- Exporting reports/documents where permissions allow.

It does not directly access MSR, iButton, fiscal devices, cash drawers, printers, or other local
peripherals. If an operation requires hardware-backed authentication or local device access,
the web interface must either block the action, route it to the touch terminal, or use a
separately approved browser-compatible factor.

### Senior Cashier Touch Terminal

The touch terminal runs as a local terminal app and is intended for in-store money operations.

It supports:

- Touch-optimized login and operation flows.
- Personnel ID and PIN entry.
- iButton confirmation.
- MSR/staff magnetic card reading where configured.
- Hardware-backed operation signing.
- Local device health display.
- Fast operation handover and auto-lock.

The touch terminal communicates with:

- Store Edge for business commands and state.
- Hardware Agent for MSR, iButton, and other local devices.

The touch terminal is the preferred surface for operations that require physical presence,
two-person signing, or local hardware confirmation.

## Authentication

The design requires three-factor authorization:

1. Personnel number.
2. PIN code.
3. iButton tablet confirmation.

The terminal shows the user, role, terminal ID, store, and software version. Every session
must be personal. Shared senior cashier sessions are not allowed.

Failed login behavior:

- Every failed attempt is written to the security journal.
- After the configured number of failures, the terminal or user factor is blocked.
- Recovery requires authorized admin or senior role action.

Alternative factors shown in the design include Mercadia staff magnetic card or similar
staff token. The implementation should treat this as an MSR-backed configurable hardware
credential.

Hardware authentication factors:

- iButton reader.
- MSR for staff magnetic cards.
- Optional future staff card/token readers.

Authentication and signing events must record which factor was used, which physical terminal
was used, and which Hardware Agent/device reported the factor.

## Staff Credential Enrollment

The senior cashier touch terminal can issue employee physical credentials because it has local
Hardware Agent access. It can also revoke active credential bindings and replace a lost or
damaged physical credential. The enrollment flow is:

1. Senior cashier signs in with a Store Edge session that has credential-management permission.
2. Senior cashier selects the target employee.
3. Terminal reads the selected credential kind through Hardware Agent (`iButton`, staff MSR card,
   or barcode staff card).
4. Terminal sends the safe token returned by Hardware Agent to Store Edge once for hashing and
   binding, with an idempotency key for enrollment.
5. Store Edge returns only masked labels and token fingerprints.

The revoke/replace flow is:

1. Senior cashier selects the target employee and reviews current masked credential bindings.
2. Senior cashier explicitly confirms revocation of an active binding.
3. Terminal sends the credential kind and token fingerprint to Store Edge with an idempotency key.
4. To replace a credential, senior cashier reads the replacement through Hardware Agent and binds it
   as a new credential after revocation.

Senior cashiers cannot manage their own credential bindings from the same session; Store Edge
enforces this separation-of-duties rule server-side.

## Home Dashboard

The dashboard answers "what needs to be done now" for the senior cashier.

It shows:

- Cashiers currently on shift.
- POS cash drawer revenue and current drawer cash.
- SCO assistant scope and terminal groups.
- Operations with money.
- Safe operations.
- Monitoring entry points.
- End-of-day entry.
- Operation journal.
- Safe balance, bank collection amount, and configured safe limit.
- Open alerts and pending approvals.
- Auto-lock countdown.
- Handover action.

Primary actions:

- Issue change fund to cashier.
- Receive money from cashier.
- Confirm return or cancellation.
- Run final cashier collection.
- Recount safe.
- Start bank collection.
- Record business expense.
- Open cash monitoring.
- Open EoD.

## Issue Change Fund To Cashier

This operation transfers cash from safe to cashier drawer.

Flow:

1. Select cashier or existing POS-generated request.
2. Review operation ID, cashier, cash drawer, shift, and source safe.
3. Display denomination breakdown. If the source operation was created on POS, the breakdown
   may be read-only.
4. Senior cashier physically counts cash.
5. Senior cashier hands cash to cashier.
6. Cashier confirms receipt on the target POS or linked flow.
7. Senior cashier confirms issuance.
8. Operation is posted with both signatures if required.

Data captured:

- Operation ID.
- Source safe.
- Destination drawer.
- Cashier.
- Senior cashier.
- Denomination breakdown.
- Total amount.
- Reason, usually change fund.
- Printed documents such as cash order.
- Signature status.

Mismatch path:

- If counted cash does not match expected breakdown, senior cashier chooses "does not match".
- Operation moves to investigation or correction state and is not posted as normal.

## Receive Cash From Cashier

This operation transfers revenue or excess cash from cashier drawer to safe.

Flow:

1. Select cashier or pending POS request.
2. Review expected drawer amount, revenue since last collection, and reason.
3. Count denominations and coins.
4. Compare against amount declared by cashier/POS.
5. Confirm receipt if matching.
6. If mismatch exists, record discrepancy and route to investigation.
7. Update safe and drawer balances.
8. Print or attach required documents.

Common reasons:

- Revenue collection due to drawer limit.
- Shift-end collection.
- Manual cash withdrawal.
- Correction after mismatch.

## Final Cash Collection For Cashier Shift

The final collection closes the cashier's money accountability for the shift.

Requirements:

- Show cashier identity, cash drawer, shift start time, and checks since last collection.
- Show critical operations that need attention before close.
- Require denomination recount.
- Compare POS expected amount against counted amount.
- Allow document upload or rejection for operations that lack required proof.
- Save draft while unresolved.
- Confirm collection only when blocking issues are resolved or authorized override is recorded.
- Produce act/order documents.
- Update EoD readiness.

The design shows B2B invoice items, returns, missing documents, and signed/unsigned states
inside final collection. These must be first-class closure checks, not free-form notes.

## Safe Recount

Safe recount validates the physical contents of the safe.

Flow:

1. Senior cashier starts recount.
2. System shows current expected balance and denomination composition.
3. Senior cashier enters actual denominations and coins.
4. System compares expected vs actual.
5. If equal, recount is confirmed.
6. If different, discrepancy is recorded with reason and optional documents.
7. Recount becomes part of EoD and audit journal.

Safe recount should be possible during the day and as an EoD prerequisite.

## Bank Collection

Bank collection moves cash from safe to collector/bank.

Requirements:

- Collection can be scheduled or triggered by safe limit.
- Show collector/vendor, contract, schedule, recipient bank account, and bag/seal number.
- Enter declared denominations.
- Enter actual denominations during handoff.
- Detect mismatch.
- Print collection act.
- Confirm and close collection.
- Update safe balance.
- Keep operation in `in_progress` until handoff confirmation is complete.

## Business Expense

The senior cashier can record controlled cash expense from safe.

Requirements:

- Required expense reason.
- Recipient/payee.
- Amount and denomination breakdown.
- Optional comment.
- Required supporting document when policy demands it.
- Signature and audit entry.
- Safe balance update.

## Cashier And SCO Monitoring

The terminal provides operational monitoring:

- POS terminal cards with cashier, status, current activity, receipt count, revenue, drawer amount.
- SCO terminal cards with status, current action, help requests, selective control, tape/paper alerts.
- Filters by active, attention, blocked, offline, or control status.
- Click-through to terminal detail.
- Ability to open pending help/control operations.

## Shift Handover

The design includes "transfer terminal/shift" behavior.

Requirements:

- Current senior cashier starts handover.
- New senior cashier authenticates with required factors.
- System shows current active operations, pending approvals, safe state, and session journal.
- Both users confirm handover.
- Previous session is closed or marked handed over.
- New session inherits operational responsibility only after confirmation.

Open point: whether handover transfers only the terminal session or the formal senior cashier
responsibility for the operational day must be decided.

## Operation Journal

The journal is immutable and signed.

It must include:

- Session start/end.
- Change fund issuance.
- Cash receipt from cashier.
- Returns confirmed.
- Cancellations confirmed.
- Bank collection.
- Business expense.
- Safe recount.
- EoD steps.
- Documents attached or rejected.
- Mismatches and overrides.

Each record includes:

- Timestamp.
- Operation ID.
- Actor.
- Counterparty user if any.
- Source and destination cash container.
- Amount.
- Status.
- Signature status.
- Linked documents.
- Export to PDF.

The design states entries are signed by iButton and retained for years. Exact retention period
must be confirmed by legal/accounting requirements.

## Permissions

Senior cashier permissions should be granular:

- View monitoring.
- Issue change fund.
- Receive cash.
- Confirm returns.
- Confirm cancellations.
- Start safe recount.
- Confirm safe recount.
- Start bank collection.
- Confirm bank collection.
- Record expense.
- Override mismatch.
- Close cashier shift.
- Start/advance EoD.
- Handover session.
- Export journal.

## Audit And Compliance

Every senior cashier action must capture:

- User ID.
- Role.
- Authentication factors used.
- Terminal ID.
- Timestamp.
- Store and operational day.
- Operation ID.
- Amount and denominations where applicable.
- Reason.
- Before and after balances.
- Related POS/SCO receipt or shift.
- Documents printed or uploaded.

No posted money operation should be physically deleted. Corrections must be modeled as
additional signed operations.
