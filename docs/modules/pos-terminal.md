# POS Terminal Specification

## Purpose

The POS terminal is the primary cashier workstation for sales, returns, payments, and
limited service operations. It must be optimized for speed, keyboard operation, scanner input,
clear totals, and recovery from payment or fiscal exceptions.

## Users

- Cashier: performs sales, returns, and allowed service operations.
- Senior cashier or manager: approves restricted actions and money operations.
- Customer: indirectly represented through loyalty, payments, and receipt delivery.

## Entry And Authentication

The POS supports personal authentication by:

- Personnel ID.
- PIN code.
- iButton or equivalent staff token.

Authentication must create a signed cashier session. Failed login attempts must be logged.
After a configurable number of failures, the terminal or user factor can be locked pending
supervisor/admin action.

The Store Edge-backed POS terminal login slice creates a Store Edge auth session from personnel ID,
PIN, and a local Hardware Agent staff credential read (`iButton`, staff MSR card, or barcode staff
card). The terminal stores the POS session separately from other apps, uses the session actor as the
cashier for shift and receipt commands, and clears sale state on logout. POS access is limited in
the UI to cashier, senior cashier, or admin roles; server-side permissions remain authoritative.
Store-level authentication hardening settings are managed in admin-web Store Settings and exposed by
Store Edge for terminals. These settings control failed-attempt lockout and POS idle auto-lock.

The terminal must show:

- Store and terminal identity.
- Current cashier.
- Shift start time.
- Terminal readiness.
- Current time.
- Auto-lock or timeout state where configured.

## Main Sale Screen

The sale screen consists of:

- Left navigation for checkout, customers, orders, returns, promotions, reports, bonuses,
  and other configured POS areas.
- Receipt line table with product name, quantity, price, and amount.
- Current receipt number.
- Cashier and loyalty customer summary.
- Totals area with subtotal, discounts, VAT, paid amount, and remaining amount.
- Payment method panel.
- Digital keypad and function key shortcuts.
- Operational side panels for payments, terminal state, and contextual actions.

The POS must support both scanner-driven and manual flows. Scanning a barcode should add
the matching product line or increment the existing line according to configured behavior.
Manual entry should allow barcode/SKU/search-based item selection.

The first Store Edge-backed POS sale slice supports terminal preparation, operational day and
cashier shift readiness, opening a receipt, scanning catalog products by barcode, capturing cash
or mock card payments, creating a fiscal document through Store Edge, and finishing the sale so the
next receipt starts from a clean terminal state. Payment history and fiscal document state are
reloaded from Store Edge when receipt state is refreshed.

## Receipt Line Behavior

For each line the POS must track:

- Product SKU and barcode.
- Product display name.
- Quantity or weight.
- Unit price.
- Line amount.
- Discount amount and reason.
- Tax/VAT calculation.
- Marking state for marked goods.
- Age restriction state.
- Manual override flags.
- Removed/cancelled state.

Allowed line actions:

- Increase quantity.
- Decrease quantity.
- Enter quantity through keypad.
- Enter weighted quantity from scale or manual fallback.
- Remove line.
- Change price or discount if permission allows.
- View product details.
- Resolve marking requirement.
- Resolve age restriction where applicable.

All line modifications after initial scan must be auditable.

## Product Scanning

The scanning loop must handle:

- Standard EAN/UPC barcode.
- Internal barcode.
- SKU lookup.
- Weighted barcode where price or weight is encoded.
- DataMatrix for marked goods.
- Duplicate scans.
- Unknown barcode.
- Product unavailable for the current store/channel.
- Product requiring senior approval.

When a product requires DataMatrix, the POS must prompt the cashier to scan the code and
validate it before payment or fiscalization. The system should allow a configured supervisor
override only when legally permitted.

## Loyalty And Customer Context

The POS must allow identifying a customer through:

- Loyalty card.
- Phone number.
- QR/app code.
- Manual search where permitted.

When a customer is attached, the POS shows:

- Customer name or masked identity.
- Loyalty tier, for example Gold.
- Available bonuses.
- Applicable discounts.
- Bonus write-off eligibility.

The POS must recalculate totals after customer binding, unbinding, or bonus write-off.

## Payment Methods

The design shows these payment methods:

- Cash.
- Bank card/MES.
- QR/SBP.
- B2B account/invoice.
- Bonuses.
- Gift card.

The POS must support split payments across multiple methods. Each payment line includes:

- Method.
- Amount.
- Status.
- Processor/terminal identifier.
- Authorization code or RRN where available.
- Timestamp.
- Reversal/refund references.

The UI must show:

- Total due.
- Paid amount.
- Remaining amount.
- Active payment method.
- Available function key shortcuts.
- Reset payments action.
- Split receipt action where enabled.

## Cash Payment

Cash flow requirements:

1. Cashier selects cash.
2. POS shows amount due.
3. Cashier enters received amount through keypad or quick amount buttons.
4. POS calculates change.
5. POS blocks finalization if received amount is below due amount unless split payment remains active.
6. Cash drawer opens only at the configured point.
7. Receipt is fiscalized and printed/sent.

Cash payment must update the drawer balance and cash movement ledger.

## Bank Card / MES Payment

Card flow requirements:

1. Cashier selects bank card/MES.
2. POS assigns or displays the active payment terminal.
3. POS prompts customer to tap/insert/swipe card or phone.
4. Payment status is updated from the terminal integration.
5. Approved payment stores authorization details.
6. Declined/timeout/cancelled payment can be retried, changed to another method, or removed.
7. Receipt finalization waits for payment confirmation.

The cashier must be able to change the payment terminal or enter manual fallback data only
when permissions and integration policy allow it.

## QR / SBP Payment

QR flow requirements:

1. Cashier selects QR/SBP.
2. POS requests a payment QR for the exact amount.
3. QR is displayed with instructions for bank app payment.
4. QR has an expiration timer.
5. Cashier can refresh the QR or send it by SMS where configured.
6. POS waits for asynchronous confirmation.
7. On confirmation, POS marks payment as captured.
8. On expiration or cancellation, POS allows retry or method change.

## B2B Account Payment

B2B flow requirements:

1. Cashier selects B2B account.
2. POS displays selected counterparty, tax identifier, contract, limit, and available debt balance.
3. Cashier can change counterparty where permission allows.
4. POS validates that the total is within account rules.
5. Payment is posted as receivable/invoice rather than immediate tender.
6. Required documents must be attached, printed, signed, or marked pending according to policy.

The B2B payment method must be visible in reconciliation and EoD because documents may
remain pending after receipt fiscalization.

## Bonus Payment

Bonus flow requirements:

- Show available bonus balance.
- Show maximum amount allowed for the receipt.
- Apply bonus write-off.
- Recalculate discounts and payment due.
- Support rollback before finalization.
- Store loyalty authorization and balance result.

## Gift Card Payment

Gift card flow requirements:

- Scan or enter gift card.
- Validate balance and status.
- Apply gift card amount.
- Support partial usage.
- Store gift card transaction reference.
- Reverse gift card authorization if receipt is cancelled before completion.

## Service Operation: Cash In

Cash-in is a controlled operation that adds cash to the drawer, usually as change fund.

The POS must support:

- Denomination input.
- Additional arbitrary amount.
- Optional comment.
- Expected drawer balance after operation.
- Link to senior cashier/safe operation if initiated externally.
- Confirmation and audit record.

The design shows denomination tiles, quick amount entry, comment field, and resulting drawer
balance. This operation should require appropriate permissions or senior cashier approval.

## Service Operation: Cash Out

Cash-out removes cash from the drawer for revenue collection, safe transfer, expense, or
other configured reasons.

The POS must support:

- Denomination input.
- Required reason.
- Optional comment.
- Expected drawer balance after operation.
- Confirmation and audit record.
- Counterparty confirmation by senior cashier when moving cash to safe.

## Cash Recount

The POS includes cash recount flow. Requirements:

- Cashier counts denominations and coins.
- POS calculates declared total.
- POS compares declared amount with expected drawer amount.
- If amounts match, status becomes confirmed.
- If mismatch exists, POS creates discrepancy requiring senior review.
- Recount result is included in EoD and cash audit.

## Receipt Cancellation

Before finalization/fiscalization, a cashier or assistant may cancel a receipt according to
permissions.

Cancellation flow:

1. User starts cancel receipt.
2. POS shows receipt number, amount, line count, and effect on totals.
3. User selects cancellation reason, for example customer changed mind, cashier error,
   duplicate/random scan, marking issue, or technical failure.
4. POS records cancellation, releases payment holds, reverses discounts/bonuses, and resets UI.
5. If payments were already captured, reversal or refund flow is required.

Cancellation must be clearly separated from post-fiscal return.

## Returns

Return requirements:

- Search original receipt by number, date, fiscal identifier, customer, payment card mask,
  or scanned receipt code where available.
- Select full receipt or specific lines.
- Support return with receipt and controlled no-receipt return.
- Require reason.
- Require senior approval for no-receipt, high value, age-restricted, or exceptional returns.
- Refund to original payment method where possible.
- Support cash refund, card refund, gift card reversal, bonus correction, and B2B correction.
- Generate legal fiscal return/correction documents.

The POS design indicates return certificate and staff approval concepts; exact legal document
types must be confirmed for target jurisdiction.

## Supervisor Intervention

The POS must be able to request or require senior approval for:

- Receipt cancellation.
- Returns.
- No-receipt returns.
- Price override.
- Discount override.
- Manual product entry.
- Marking bypass if allowed.
- Age restriction if the POS cashier role cannot approve alone.
- Cash drawer mismatch.
- Payment reversal.

Approval must capture approver identity and authentication factor.

## Error Handling

The POS must handle:

- Product not found.
- Price missing or stale.
- Network unavailable.
- Payment terminal offline.
- Payment timeout.
- Fiscal device error.
- Cash drawer unavailable.
- Printer out of paper.
- Marking validation unavailable.
- Loyalty service unavailable.
- Duplicate DataMatrix.
- Offline sync conflict.

Each error must provide an operator-safe recovery path: retry, change method, request
supervisor, save as exception, or cancel.

## Keyboard And Shortcut Requirements

The UI shows function key ranges for payment methods and operational actions. Implementation
should support:

- Numeric keypad entry.
- Enter to confirm.
- Escape/back where safe.
- Function key mappings for payment methods.
- Scanner input without focusing a text field.
- Permission-aware disabling of unsafe shortcuts.

## Audit Requirements

Audit entries must be written for:

- Login and logout.
- Shift open/close.
- Product add/remove/change.
- Manual price/discount changes.
- Loyalty binding and bonus usage.
- Payment creation, capture, cancellation, reversal.
- Receipt cancellation.
- Return creation and approval.
- Cash in/out/recount.
- Supervisor approvals.
- Fiscal errors and recovery steps.

Audit entries should be immutable and synchronized to the store/backend event log.
