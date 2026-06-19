# Self-Checkout Specification

## Purpose

Self-checkout, called KSO/SCO in the designs, lets customers scan and pay independently while
staff supervise exceptions. The product must support horizontal, vertical, and HD layouts
without changing the business workflow.

## Layout Variants

The exported designs show:

- Horizontal customer checkout.
- Vertical customer checkout.
- HD customer checkout.
- Assistant station for a group of SCO terminals.

Implementation should use one workflow engine and responsive/adaptive UI shells for different
screen formats. Device capabilities and layout profile should be terminal configuration.

## Customer Idle Screen

Idle screen requirements:

- Brand and store context.
- Language selector, at least Russian and English.
- "Start scanning" primary action.
- Short customer guidance: scan, check, pay.
- Current availability state.
- Optional staff call/help action.
- Reset state when previous receipt is complete or abandoned.

## Scanning Flow

Customer scanning must support:

- Standard product barcode.
- DataMatrix for marked goods.
- Weighted product lookup.
- Product grid for produce/manual selection.
- Quantity changes where allowed.
- Deleting a line only through allowed customer or staff flow.

The active receipt area shows:

- Receipt number.
- Product list.
- Quantity and price.
- Discounts.
- Subtotal.
- VAT.
- Amount due.
- Current stage: scanning, receipt, payment, done.

## Marked Goods / DataMatrix

When a product requires marking validation, SCO must pause the normal flow and ask the
customer to scan DataMatrix.

Requirements:

- Show the affected product.
- Show example of where the DataMatrix code is located.
- Provide "where is the code?" help.
- Provide "cannot scan" path.
- Validate DataMatrix before payment where legally required.
- Link DataMatrix code to the receipt line.
- Prevent duplicate or invalid code usage.
- Escalate to assistant if customer cannot complete the scan.

## Loyalty Flow

The design shows a Mercadia loyalty prompt with bonus accrual and write-off.

Supported identification methods:

- Phone number.
- QR from mobile app.
- Card scan.

After binding a loyalty account, SCO must show:

- Customer display name or masked identity.
- Bonus balance.
- Discount tier.
- Bonus amount available for write-off.
- Applied loyalty discount.

The customer must be able to continue without loyalty.

## Payment Selection

Supported payment methods in the design:

- Bank card/MES.
- QR/SBP.
- Bonuses.
- Gift card.
- Payment through cashier or staff-assisted path where configured.

Payment selection must show amount due and available method cards. Unavailable methods should
be hidden or disabled with staff-readable reason.

## Bank Card Payment

Flow:

1. Customer selects bank card.
2. SCO shows "tap card or phone" instructions.
3. Payment terminal status is displayed.
4. Customer can change payment method or request help before authorization.
5. On approved payment, SCO proceeds to fiscalization and success.
6. On decline/timeout, SCO shows retry, change method, or call employee.

## QR / SBP Payment

Flow:

1. Customer selects QR/SBP.
2. SCO displays QR code and bank app instructions.
3. QR expiration timer is visible.
4. SCO waits for confirmation.
5. Customer can change method or request help.
6. On success, SCO finalizes the receipt.

## Success Screen

After payment and fiscalization, SCO must show:

- Clear success confirmation.
- Total paid.
- Loyalty bonuses accrued or written off.
- Receipt delivery options: QR, email, print, or configured options.
- "Finish" action.

After finish or timeout, terminal returns to idle.

## Weighted Goods And Produce Selection

The designs show produce tiles and weight confirmation.

Requirements:

- Product search/filter by category.
- Product cards with image, name, unit price, and color-coded tile.
- Scale integration where available.
- Manual fallback only in assistant mode or when policy allows.
- Weight comparison screen: measured/expected values and pass/fail status.
- Ability to change item parameters when allowed.
- Audit entry for assistant override.

## Item Parameter Editing

Some screens show editing parameters for a line, such as price/quantity/weight. Requirements:

- Customer can only edit safe fields, such as quantity for eligible packaged goods.
- Assistant can edit controlled fields based on permissions.
- Changes recalculate totals immediately.
- Original values remain in audit history.
- Restricted changes require reason.

## Age Verification

For alcohol/tobacco/18+ goods, SCO must:

- Detect restricted goods in cart.
- Stop checkout or payment until verification.
- Show age verification required.
- Notify assistant station.
- Let assistant verify document and approve or reject.
- Record assistant identity, timestamp, result, and terminal.
- Remove restricted items or cancel receipt if verification fails.

## Selective Control And Full Rescan

The assistant station and customer screens show selective control, full rescan, and audit
states.

Requirements:

- System can randomly select a receipt for control based on configurable rules.
- Control can also be triggered by risk signals, age goods, marking failures, manual changes,
  payment anomalies, or assistant action.
- Customer sees "please wait" or "employee will come" state.
- Assistant sees required control actions.
- Control types can include random item count, full rescan, weight check, and receipt review.
- Result can be passed, failed, or escalated.
- Failure can require item correction, receipt cancellation, or manager approval.

## Assistant Mode On SCO

Assistant mode is entered after staff authentication and overlays operational actions on the
customer checkout.

Available assistant actions shown in the designs:

- Start selective control.
- Cancel receipt.
- Remove item.
- Change price or discount.
- Confirm 18+.
- Accept marking issue.
- Manual item entry.
- Start return.
- Print receipt copy.
- Block terminal.
- Exit assistant mode.

All assistant mode actions must be logged with assistant ID and session.

## Customer Help

The customer can call an employee. The SCO must:

- Show waiting state.
- Notify assistant station.
- Display approximate or actual response state if available.
- Let assistant claim the request.
- Allow customer to cancel request if no longer needed.

## Remove From Service

The design includes "take KSO number out of work" screen.

Requirements:

- Only staff can remove terminal from service.
- Reason is required.
- Current receipt must be absent, completed, or explicitly cancelled.
- Terminal enters out-of-service state.
- Admin monitoring updates immediately.
- Returning to service requires staff action and health checks.

## Assistant Station

Assistant station supervises multiple SCO terminals.

It displays:

- Zone and terminal group.
- Number of working terminals.
- Help queue.
- Checks today.
- Average check.
- Audit percentage.
- Response time.
- Cards/list of terminals with status and current action.
- Alerts such as alcohol 18+, selective control, out of tape, blocked terminal.

Assistant actions:

- Open terminal details.
- Claim help request.
- Approve age check.
- Start/finish control.
- Block/unblock terminal.
- Send terminal out of service.
- View shift controls.

## Error Handling

SCO must provide friendly recovery for:

- Unknown barcode.
- Product cannot be sold on SCO.
- DataMatrix invalid/unavailable.
- Scale unavailable.
- Weight mismatch.
- Payment declined.
- QR expired.
- Fiscalization failed.
- Printer unavailable.
- Network offline.
- Customer inactivity.
- Assistant timeout.

Customer-facing text should avoid technical details. Staff screens should expose detailed
diagnostics.

## Data And Audit

Each SCO receipt must store:

- Terminal ID.
- Layout/profile.
- Customer language.
- Receipt lines and scan source.
- Marking validations.
- Weight validations.
- Loyalty events.
- Payment attempts.
- Assistant interventions.
- Control results.
- Fiscalization result.

SCO audit is especially important because customer actions and staff interventions share
the same receipt. The audit model must distinguish customer input, system automation, and
authenticated staff actions.
