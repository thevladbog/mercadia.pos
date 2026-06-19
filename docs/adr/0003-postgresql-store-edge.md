# ADR-0003: PostgreSQL As Store Edge Database

Status: Accepted

## Context

Store Edge must coordinate multiple POS terminals, SCO terminals, a senior cashier terminal,
assistant station, local admin views, cash ledger, receipts, fiscal operations, payment
operations, monitoring, EoD, and synchronization.

## Decision

Use PostgreSQL as the primary Store Edge database.

PostgreSQL stores:

- Receipt state and receipt ledger.
- Cash ledger and safe state.
- Fiscal/payment operation state.
- Terminal state.
- Local catalog and price cache.
- Outbox/inbox records.
- Idempotency keys.
- Sync checkpoints.

SQLite may be used only for terminal-local UI cache, not as the main Store Edge database.

## Consequences

- We get transactional integrity for multi-terminal store operations.
- We can implement outbox/inbox reliably inside the same database transaction.
- Operational setup is heavier than a file database and must be automated.
- Backup, migrations, and local diagnostics must be part of Store Edge packaging.

## Open Points

- PostgreSQL installation model: bundled, managed by installer, Docker/container, or OS package.
- Backup/restore process.
- Migration tooling.
- Store Edge database encryption requirements.
