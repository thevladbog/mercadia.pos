# Functional Overview

## Product Goal

Mercadia POS is a multi-surface retail checkout platform for a store network. It must
support high-speed cashier sales, customer self-checkout, senior cashier cash control,
and back-office administration from one consistent operational model.

The platform is designed for stores where cash, card, QR/SBP, B2B invoice payments,
gift cards, loyalty, returns, marked goods, weighted goods, fiscal receipts, and end-of-day
cash reconciliation all coexist. The user interfaces prioritize operational speed,
clear cash accountability, auditability, and low ambiguity during exceptions.

## Product Surfaces

### POS Terminal

The POS terminal is used by cashiers for standard checkout. It supports:

- Product scanning and manual product entry.
- Quantity, weight, discount, and price adjustments according to permissions.
- Loyalty customer identification.
- Multiple payment methods in a single receipt.
- Cash, bank card/MES, QR/SBP, B2B account, bonuses, and gift card payment flows.
- Cash in, cash out, cash recount, and other service operations.
- Receipt cancellation before fiscalization.
- Returns with receipt lookup or controlled no-receipt handling.
- PIN and iButton authentication.

### Senior Cashier Terminal

The senior cashier terminal is a restricted operational console for cash and shift control.
It supports:

- Three-factor authentication: personnel ID, PIN, and iButton.
- Monitoring cashiers and self-checkout nodes during the current shift.
- Issuing change fund to cashiers.
- Receiving cash from cashiers.
- Final cashier cash collection at shift close.
- Safe recounts.
- Bank collection handoff.
- Confirming exceptions and unresolved cashier operations.
- Viewing an immutable signed operation journal.
- Handover to another senior cashier.

### Self-Checkout / KSO / SCO

The self-checkout surface lets customers complete purchases without a cashier while still
allowing staff intervention. It supports:

- Idle language selection and start scanning.
- Barcode and DataMatrix scanning.
- Cart management with a persistent receipt summary.
- Loyalty card binding by phone, QR, or app.
- Payment by bank card/MES, QR/SBP, bonuses, and gift card where enabled.
- Marked goods verification.
- Weighted goods and produce selection.
- Age verification for restricted goods.
- Random control and full rescan workflows.
- Assistant mode for staff actions.
- Blocking/unblocking, checkout cancellation, and calling an employee.

### Admin Panel

The admin panel is the store/back-office control plane. It supports:

- Real-time POS and SCO monitoring with refresh cadence and alerts.
- Safe balance, cash movements, denomination breakdown, and bank collection.
- EoD and cashier shift close workflows.
- Users, roles, and access configuration.
- Product and price catalog management.
- Reconciliation against backend systems.
- Integration health and external service configuration.
- POS/SCO registration, pairing, status, and device metadata.
- Layout templates for POS/SCO buttons and product grids.
- Payment method configuration.
- Gift card configuration.
- Receipt template configuration.
- Color themes and franchise branding.
- System settings and event logs.

## Core Roles

### Cashier

The cashier works at a POS terminal, processes sales and returns, accepts payments,
performs allowed service operations, and initiates supervisor-required actions when needed.
Cashier actions must be associated with a personal session and a cash drawer.

### Senior Cashier

The senior cashier controls cash movement and supervises sensitive actions. They can issue
cash to cashiers, accept revenue, confirm returns and cancellations, approve exceptions,
perform safe recounts, and participate in EoD closure.

### SCO Assistant

The SCO assistant supervises a group of self-checkout terminals. They can open a terminal,
approve restricted goods, run selective control, handle customer help requests, cancel
receipts, remove items, adjust allowed values, confirm marking issues, and block terminals.

### Store Manager / Admin

The store manager/admin configures the store, users, roles, devices, integrations, product
availability, payment methods, templates, and monitoring views. Admin actions must be
auditable and role-limited.

### Customer

The customer interacts directly with SCO. Customer-facing flows must be guided, localized,
resilient to mistakes, and easy to recover with staff help.

## Shared Business Objects

### Store

A physical retail location with one or more POS terminals, SCO terminals, safes, cashiers,
senior cashiers, product availability, integrations, and operational day status.

### Operational Day

The store-level business day. It begins before checkout operations and ends through EoD.
It groups shifts, receipts, cash movements, reconciliation results, and closure documents.

### Shift

A personal work period for a cashier, senior cashier, or assistant. For POS cashiers, a shift
is linked to a cash drawer and cash balance. For senior cashiers, a session signs cash-office
actions. For SCO assistants, a session signs interventions.

### Terminal

A registered device that can be a POS, SCO, assistant station, senior cashier terminal, or
admin workstation. Terminals have identifiers, store binding, hardware configuration,
network state, current user/session, software version, and operational status.

### Receipt

The legal and commercial transaction container. It includes lines, discounts, taxes,
loyalty context, payment lines, fiscalization state, status, and links to returns,
cancellations, or correction documents.

### Receipt Line

A product entry with SKU, barcode, optional DataMatrix code, quantity, weight, unit price,
discounts, tax category, marking status, age restriction, and line status.

### Payment

A payment attempt or completed payment line. It includes method, amount, status,
authorization details, processor identifiers, timestamps, and reversal/refund data.

### Cash Movement

A controlled movement of cash between cashier drawer, safe, bank collection, or expense.
Cash movements include amount, denomination breakdown, reason, source, destination,
initiator, confirmer, related shift, documents, and signatures.

### Safe

A logical cash container for a store or office. It tracks current balance, denomination
composition, collection threshold, recounts, incoming/outgoing movements, and audit history.

### Product

An item sellable at POS/SCO. It includes SKU, name, barcode(s), price, tax, category,
unit/weight behavior, images, restrictions, marking requirements, available channels,
and layout placement.

### User

A person with credentials, role assignments, authentication factors, personnel ID,
iButton/card identifiers, active status, and audit history.

## Shared Status Vocabulary

### Terminal Statuses

- `active`: terminal is online and in normal operation.
- `idle`: terminal is available but no active receipt is in progress.
- `busy`: receipt, payment, marking, or service operation is in progress.
- `needs_attention`: customer or cashier action needs senior/assistant attention.
- `control`: selective control or assistant verification is active.
- `blocked`: terminal cannot serve customers until released.
- `offline`: terminal is not communicating with the store platform.
- `out_of_service`: terminal is intentionally removed from service.

### Receipt Statuses

- `draft`: receipt is being edited.
- `payment_pending`: checkout entered payment stage.
- `partially_paid`: at least one payment is captured, but amount remains.
- `paid`: commercial payment is complete.
- `fiscal_pending`: payment complete, fiscal receipt not yet confirmed.
- `fiscalized`: legal receipt is issued.
- `cancelled`: receipt was cancelled before completion.
- `return_pending`: return flow is started.
- `returned`: return was completed.
- `exception`: receipt needs manual resolution.

### Cash Movement Statuses

- `created`: operation was started.
- `pending_counterparty`: another user must confirm or receive cash.
- `pending_documents`: required order, correction receipt, or attachment is missing.
- `count_mismatch`: declared and counted amounts differ.
- `confirmed`: operation is accepted and posted.
- `cancelled`: operation was cancelled before posting.
- `rejected`: operation was rejected after review.

## End-to-End Workflows

### Store Opening

1. Admin or senior cashier opens the operational day.
2. Terminals load configuration, product cache, price cache, payment methods, templates,
   and fiscal device state.
3. Cashiers authenticate at POS terminals.
4. Senior cashier issues starting change fund where required.
5. SCO terminals are moved from out-of-service or idle state to customer-ready state.
6. Monitoring starts reporting live state for all POS/SCO nodes.

### Standard Cashier Sale

1. Cashier scans products or adds them manually.
2. POS validates product availability, price, restrictions, taxes, and marking requirements.
3. Cashier identifies loyalty customer if applicable.
4. POS calculates subtotal, discounts, VAT, bonuses, and total.
5. Cashier selects payment method(s).
6. Payment is authorized or cash is accepted.
7. POS fiscalizes the receipt.
8. Receipt, fiscal data, payment result, inventory movement, and loyalty accrual are posted.
9. Monitoring and reports update in near real time.

### SCO Customer Sale

1. Customer starts from idle screen and selects language if needed.
2. Customer scans products.
3. SCO interrupts the flow for marked goods, weighted goods, restricted goods, or unknown items.
4. Customer applies loyalty if desired.
5. Customer selects payment.
6. SCO waits for payment confirmation.
7. SCO fiscalizes and shows success.
8. Customer exits; terminal resets to idle.
9. Any control, age verification, or exception is signed by an assistant.

### Cashier Change Fund

1. Cashier requests or senior cashier initiates change fund.
2. Denomination breakdown is shown exactly as created by the source operation.
3. Senior cashier counts and confirms cash from safe to drawer.
4. Cashier receives and signs on POS or related terminal.
5. Safe and drawer balances are updated.
6. Cash order and fiscal/correction documents are printed or attached where required.

### Cash Collection From Cashier

1. Cashier initiates cash surrender or senior cashier opens the operation.
2. POS provides expected denomination breakdown when available.
3. Senior cashier counts cash.
4. Differences are recorded as count mismatch and routed to investigation or correction.
5. Confirmed amount moves from drawer to safe.
6. Operation is signed by both cashier and senior cashier.

### End Of Day

1. System identifies all open shifts, unresolved receipts, pending returns, pending B2B
   documents, cash movements, safe balance, and integration errors.
2. Cashiers close shifts and perform final cash collection.
3. Senior cashier performs or confirms safe recount.
4. Bank collection is prepared when threshold or schedule requires it.
5. Admin/EoD workflow validates critical operations and documents.
6. Reconciliation compares POS totals with backend, fiscal, payment, bank, loyalty,
   marking, and inventory sources.
7. Operational day is closed only when blocking issues are resolved or explicitly approved.

## Non-Functional Requirements

### Reliability

Checkout must continue during temporary backend outages whenever legally and operationally
allowed. Local terminal/store services should keep product, price, payment, fiscal, and
receipt state consistent enough to complete the customer interaction and synchronize later.

### Auditability

Every sensitive action must be recorded with actor, terminal, timestamp, reason, before/after
values, related receipt or cash movement, authentication factors, and approval chain.

### Security

Role-based access control must be enforced on the server side and in terminal UI state.
PIN and iButton authentication must be treated as strong operational controls for money
and exception workflows. Administrative actions need permission separation.

### Observability

Admin monitoring needs live terminal state, action logs, payment and fiscal errors,
integration health, queue depth, sync status, and cash balance anomalies.

### Localization

Customer-facing SCO flows must support at least Russian and English based on design.
Operational UIs may initially be Russian, but the implementation should keep all labels
externalized.

### Performance

POS and SCO interactions must feel instant for scanning, quantity changes, payment selection,
and cart recalculation. Network calls must not block the scanning loop unless required by
marking, payment, or fiscalization rules.

### Accessibility And Ergonomics

SCO buttons must remain large and touch-friendly. POS and senior cashier screens must support
fast keyboard/numpad operation. Critical operations must use clear color states, explicit
confirmation, and persistent totals.
