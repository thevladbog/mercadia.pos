# ADR-0009: OpenAPI Generation, Scalar Docs, And Orval Clients

Status: Accepted

## Context

Mercadia has multiple frontends: admin web, senior cashier web, POS terminal, SCO/KSO terminal,
assistant station, and senior cashier touch terminal. These frontends must consume Store Edge
and central backend APIs without drifting from backend request/response contracts.

Manual TypeScript DTOs and hand-maintained API documentation would create mismatch risk,
especially with AI agents and multiple UI surfaces.

## Decision

Use OpenAPI as the API contract workflow:

- Generate OpenAPI from Go API definitions.
- Render interactive API documentation with Scalar.
- Generate frontend TypeScript clients/types/hooks/mocks with Orval.

The workflow is mandatory. The exact Go OpenAPI generator is an implementation detail, with
Huma as the preferred candidate for the backend skeleton because it is a Go framework backed
by OpenAPI 3.1 and JSON Schema.

## Consequences

- Backend API changes produce a visible contract artifact.
- Frontend clients and types are generated instead of hand-written.
- Admin, POS, SCO/KSO, assistant, and senior cashier UIs share the same API contract source.
- API docs stay close to executable backend code.
- CI must enforce OpenAPI generation, validation, Orval generation, and frontend typecheck.

## Rules

- Every endpoint must have a stable `operationId`.
- Command endpoints must support idempotency.
- Error responses must use shared schemas.
- Generated frontend code must not be edited manually.
- API-breaking changes require coordinated frontend updates.

## Open Points

- Final Go OpenAPI generation library.
- Exact generated client layout.
- Whether Orval generates React Query hooks for all APIs or only query-heavy APIs.
- API diff/breaking-change tooling in CI.
