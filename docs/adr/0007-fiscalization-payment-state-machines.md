# ADR-0007: Fiscalization And Payments As Separate State Machines

Status: Accepted

## Context

Checkout can reach difficult states: payment succeeds but fiscalization fails, payment is
cancelled before receipt completion, fiscal device is unavailable, or a correction is required.
Russia/54-FZ and ATOL behavior must be handled explicitly.

## Decision

Model payments and fiscalization as separate state machines coordinated by receipt workflow.

Payment states include:

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

Fiscal states include:

- `not_required_yet`.
- `pending`.
- `sent_to_device`.
- `fiscalized`.
- `failed_retryable`.
- `failed_blocking`.
- `correction_required`.

Payment capture does not mean the receipt is legally complete. Fiscalization failure after
payment must route to retry, correction, or manual resolution.

## Consequences

- We avoid collapsing payment and legal receipt state into one fragile status.
- Reconciliation can identify payment/fiscal mismatches.
- UI must show controlled recovery paths.
- Implementation must define sagas for payment success, payment reversal, fiscal retry, and
  correction.

## Open Points

- Exact 54-FZ handling for captured payment with failed fiscalization.
- ATOL error taxonomy and retry behavior.
- Same-day card cancellation provider details.
- Correction document rules for every failure path.
