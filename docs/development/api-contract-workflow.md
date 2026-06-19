# API Contract Workflow

Mercadia uses OpenAPI as the contract between Go backends and TypeScript frontends.

## Decision

- Generate OpenAPI from Go API definitions.
- Serve interactive API documentation with Scalar.
- Generate frontend API clients, types, React Query hooks, mocks, and validators with Orval.
- Treat generated clients as the default way for frontend code to call backend APIs.

## Contract Flow

```text
Go API definitions
        |
        | generate OpenAPI
        v
openapi.json / openapi.yaml
        |
        | served by backend + rendered by Scalar
        v
interactive API docs
        |
        | consumed by Orval
        v
TypeScript clients, types, hooks, mocks, validators
        |
        v
Admin, senior cashier web, POS, SCO/KSO, assistant UIs
```

## Go OpenAPI Generation

The preferred direction is code-first OpenAPI generation from Go source.

Preferred candidate:

- Huma, because it is a Go HTTP API framework backed by OpenAPI 3.1 and JSON Schema and is
  designed to keep documentation generated from API code.

Acceptable alternative:

- Another Go OpenAPI generation tool may be used if it supports stable operation IDs,
  request/response schemas, validation, error models, examples, and CI-friendly spec export.

The exact generator remains an implementation choice until a backend skeleton proves it.
The workflow itself is mandatory: Go API source produces OpenAPI, not hand-maintained YAML.

## Scalar

Scalar is used to render OpenAPI into interactive API reference pages.

Requirements:

- Store Edge exposes its local OpenAPI document and Scalar UI in development/local admin mode.
- Central backend exposes its OpenAPI document and Scalar UI in development/staging.
- Production exposure must be controlled by authentication and environment policy.
- The OpenAPI document should be downloadable for Orval and external validation.

## Orval

Orval is used by all TypeScript frontends to generate:

- API functions.
- TypeScript request and response types.
- TanStack Query hooks where useful.
- MSW mocks for frontend tests and isolated UI work.
- Optional validators if selected for the frontend stack.

Rules:

- Do not hand-write API DTO types in frontend code when Orval can generate them.
- Generated code should live in a clearly marked generated folder.
- Generated files should not be manually edited.
- Frontend PRs that depend on API changes must include regenerated clients.
- Operation IDs must be stable and human-readable because Orval uses them for generated names.

## API Design Rules

- Every endpoint must have a stable `operationId`.
- Commands and queries should be explicit; avoid generic CRUD for operational flows.
- Command endpoints must accept an idempotency key.
- Error responses must use a shared problem/error schema.
- Long-running operations should expose status resources or event streams.
- Realtime streams should be documented separately from request/response OpenAPI if they are
  WebSocket/SSE-based.
- Breaking changes require versioning or coordinated frontend migration.

### SSE Streams (outside OpenAPI)

Store Edge exposes terminal monitoring as Server-Sent Events:

- `GET /v1/stores/{storeId}/terminals/events`
- Response `Content-Type: text/event-stream`
- Event type: `terminal_heartbeat`
- Payload fields: `terminalId`, `storeId`, `kind`, `status`, `softwareVersion`, `lastSeenAt`, `updatedAt`

This stream is intentionally excluded from generated OpenAPI because SSE event contracts differ from
JSON request/response schemas. Document changes here and in `backend/README.md`.

## Suggested Repository Layout

```text
frontend/
  apps/
    admin-web/
    pos-terminal/
    sco-terminal/
    senior-cashier-terminal/
    assistant-station/
  packages/
    api-clients/
      store-edge/
      central/
backend/
  services/
    store-edge/
    central-backend/
    hardware-agent/
  packages/
    platform/
contracts/
  openapi/
    store-edge.openapi.json
    central.openapi.json
```

Actual layout can change, but the generated contract artifacts should be easy for humans and
agents to find.

## CI Requirements

CI should verify:

- Go APIs compile.
- OpenAPI generation succeeds.
- Generated OpenAPI is valid.
- Orval generation succeeds.
- Frontend typecheck succeeds against generated clients.
- Generated files are up to date.
- Breaking API changes are visible in review.

## Agent Notes

AI agents changing API handlers must also:

- Regenerate OpenAPI.
- Regenerate Orval clients.
- Update affected frontend call sites.
- Update docs if endpoint behavior changed.
- Mention contract changes in the final summary.
