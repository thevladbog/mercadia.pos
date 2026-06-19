# ADR-0008: SCO Cashless First Stage

Status: Accepted

## Context

SCO/KSO designs include checkout and payment flows. Supporting cash on SCO would require
additional cash hardware, cash state, replenishment/collection flows, jams/errors, change
handling, and more EoD complexity.

## Decision

SCO/KSO supports only cashless payments in the first stage.

Supported SCO payment methods can include:

- Bank card/payment terminal.
- QR/SBP.
- Bonuses.
- Gift card where configured.

Cash payment remains available on POS.

## Consequences

- SCO hardware and cash operations are simpler for MVP.
- SCO does not participate in cash drawer/safe movement flows.
- Customer UX must clearly indicate available payment methods before payment.
- Future SCO cash support can be added as a separate hardware and cash-ledger project.

## Open Points

- Exact first-stage SCO payment method list.
- Whether "payment through cashier" is a first-stage fallback.
- Future cash module vendor and integration model.
