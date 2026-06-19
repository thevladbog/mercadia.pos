# ADR-0006: Cash Ledger Immutability And Separation Of Duties

Status: Accepted

## Context

Cash operations include cash in, cash out, cashier final collection, safe recount, bank
collection, and discrepancies. These operations are financially sensitive and must be auditable.

## Decision

Represent cash as immutable ledger movements between containers.

Cash containers include:

- POS drawer.
- Store safe.
- Bank collection bag.
- Expense/outgoing counterparty.

Posted movements are not edited. Corrections are modeled as new correction/reversal movements.

Critical operations must enforce separation of duties:

- A user cannot approve their own critical operation.
- If a supervisor is logged in as a cashier on POS, they act as cashier for that session and
  cannot approve that session's critical operations.
- Store manager cannot edit already posted senior-cashier decisions; they create correction
  documents/operations instead.

## Consequences

- Cash history remains explainable and auditable.
- UI must expose correction flows instead of edit-in-place behavior.
- Permission checks must be server-side and include actor/session context.
- EoD can reason over posted movements, pending movements, and corrections.

## Open Points

- Full cash movement type catalog.
- Which operations require exactly two signatures.
- Denomination policy for coins and "other coins amount".
- Legal/accounting document mapping for every movement.
