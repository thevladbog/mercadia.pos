# ADR-0004: Lightweight Broker With Transactional Outbox

Status: Accepted

## Context

Mercadia needs asynchronous delivery for monitoring, synchronization, integration events,
background processing, and central aggregation. However, the first architecture should avoid
heavy operational infrastructure such as Kafka unless scale later justifies it.

## Decision

Use a lightweight broker from the start, together with a transactional outbox.

Preferred broker candidate:

- NATS JetStream.

Accepted alternative:

- RabbitMQ.

Avoid Kafka/Redpanda in the initial architecture.

The database outbox remains mandatory even when a broker is used. Business commands commit
state and outbox records atomically; workers publish to the broker after commit.

## Consequences

- We get asynchronous event delivery without turning broker availability into command
  transaction availability.
- Store Edge can continue local operations during central broker outage.
- Consumers must be idempotent.
- NATS JetStream fits Go and edge scenarios well, but operational familiarity must be checked.

## Open Points

- Final choice: NATS JetStream vs RabbitMQ.
- Whether Store Edge runs a local broker or only a bridge/worker to central broker.
- Retention policy for event streams.
- Dead-letter and retry strategy.
