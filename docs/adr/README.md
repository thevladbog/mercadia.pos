# Architecture Decision Records

This folder contains architecture decisions for the Mercadia POS platform.

ADR status values:

- `Accepted` - decision is agreed for the current architecture.
- `Proposed` - likely direction, still open for discussion.
- `Superseded` - replaced by a newer decision.

## Decisions

- [ADR-0001: Store Edge per store](0001-store-edge-per-store.md)
- [ADR-0002: Terminal application and hardware agent split](0002-terminal-app-and-hardware-agent.md)
- [ADR-0003: PostgreSQL as Store Edge database](0003-postgresql-store-edge.md)
- [ADR-0004: Lightweight broker with transactional outbox](0004-lightweight-broker-with-outbox.md)
- [ADR-0005: Separate store admin and central admin scopes](0005-store-admin-and-central-admin.md)
- [ADR-0006: Cash ledger immutability and separation of duties](0006-cash-ledger-and-separation-of-duties.md)
- [ADR-0007: Fiscalization and payments as separate state machines](0007-fiscalization-payment-state-machines.md)
- [ADR-0008: SCO cash payments out of first stage](0008-sco-cashless-first-stage.md)
- [ADR-0009: OpenAPI generation, Scalar docs, and Orval clients](0009-openapi-scalar-orval.md)

## Upcoming Decisions

- Admin panel frontend technology.
- Backend language/runtime and service boundaries.
- Store Edge packaging and update mechanism.
- NATS JetStream vs RabbitMQ.
- ATOL integration mode and supported models.
- Payment terminal protocol baseline.
- Observability stack.
