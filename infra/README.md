# Infrastructure

This directory contains Mercadia POS deployment, local environment, database, broker,
observability, and packaging assets.

## Local development

Start PostgreSQL and NATS:

```bash
docker compose -f infra/docker/docker-compose.yml up -d
```

SQL migrations live under `infra/migrations/`:

- `store-edge/` — Store Edge schema
- `central-backend/` — Central backend schema

## Layout

- `docker/` — Docker Compose for local PostgreSQL and NATS
- `migrations/` — Goose SQL migrations per service

Do not place application source code here.
