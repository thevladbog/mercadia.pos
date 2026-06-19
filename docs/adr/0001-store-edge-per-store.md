# ADR-0001: Store Edge Per Store

Status: Accepted

## Context

Mercadia POS must operate in physical stores with POS terminals, SCO/KSO terminals, senior
cashier terminal, assistant station, local cash operations, fiscal devices, payment terminals,
and EoD workflows. Many critical operations are local to the store and must remain observable
even when the central platform is temporarily unavailable.

## Decision

Every store must run a mandatory Store Edge runtime.

The Store Edge owns store-local operational state:

- Current operational day.
- Terminal registry and heartbeats.
- POS/SCO receipt state.
- Cash ledger and safe state.
- Local fiscal and payment operation state.
- Store monitoring streams.
- Outbox/inbox synchronization with central services.

## Consequences

- POS/SCO/senior cashier flows are not directly dependent on central backend latency.
- Store-level monitoring and cash control can continue during central outage.
- We need packaging, deployment, update, backup, and observability for Store Edge.
- Central platform becomes the aggregator and policy/master-data plane, not the synchronous
  execution path for every checkout command.

## Open Points

- Store Edge packaging format.
- Store Edge OS target.
- Store Edge high-availability expectation for larger stores.
- Exact offline limits under 54-FZ and marking rules.
