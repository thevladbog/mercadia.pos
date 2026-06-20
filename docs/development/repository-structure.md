# Repository Structure

All Mercadia POS work lives under the `pos/` directory. The parent repository root is reserved
for other Mercadia ecosystem projects and must not receive POS-specific code, scripts, generated
contracts, or documentation.

## Top-Level Directories

- `frontend/` - frontend applications, frontend packages, and generated frontend clients.
- `backend/` - backend workspace, services, shared backend packages, and backend scripts.
- `contracts/` - generated and source contract artifacts.
- `docs/` - product, architecture, development, and ADR documentation.
- `infra/` - deployment, local environment, database, observability, and packaging assets.
- `.cache/` - local generated cache, ignored by git.

Create a directory only when it has a clear owner and purpose. Do not put source code directly
in the `pos/` root except workspace files and project-level metadata.

## Frontend Layout

Frontend code lives under `frontend/`.

Frontend applications go under `frontend/apps/`.

Expected applications:

- `frontend/apps/admin-web`
- `frontend/apps/pos-terminal`
- `frontend/apps/sco-terminal`
- `frontend/apps/senior-cashier-web`
- `frontend/apps/senior-cashier-terminal`
- `frontend/apps/assistant-station`

Frontend apps must use generated API clients from `frontend/packages/api-clients` or another
generated client package defined by the frontend workspace. They must not hand-roll backend DTOs
when OpenAPI and Orval can generate them.

Shared UI lives in `frontend/packages/ui` (`@mercadia/ui`). See [`ui-components.md`](ui-components.md).

## Backend Layout

Backend code lives under `backend/`.

Deployable backend processes go under `backend/services/`.

Current services:

- `backend/services/store-edge`
- `backend/services/central-backend`
- `backend/services/hardware-agent`

Each service should keep this shape:

```text
backend/services/<service-name>/
  cmd/
    <service-name>/
      main.go
    export-openapi/
      main.go
  internal/
    api/
    app/
    domain/
    infra/
  go.mod
```

Rules:

- `cmd/` contains process entrypoints only.
- `internal/api` exposes transport adapters and OpenAPI registration.
- `internal/app` contains command/query handlers and use cases.
- `internal/domain` contains business rules, state machines, and invariants.
- `internal/infra` contains database, broker, hardware, and provider adapters.

## Shared Packages

Shared backend code goes under `backend/packages/`.

Rules:

- Shared packages must be small and boring.
- Do not put product-specific business rules into generic platform packages.
- Prefer service-local code until reuse is real.
- Keep generated frontend API clients under `frontend/packages/api-clients`, separate from
  hand-written shared UI packages.

## Contracts

Contracts go under `contracts/`.

Current generated contracts:

- `contracts/openapi/store-edge.openapi.json`
- `contracts/openapi/central.openapi.json`
- `contracts/openapi/hardware-agent.openapi.json`

Generated contracts must be updated by scripts, not edited manually.

## Infrastructure

Infrastructure files go under `infra/`.

Expected future areas:

- `infra/docker`
- `infra/postgres`
- `infra/nats`
- `infra/migrations`
- `infra/observability`
- `infra/packaging`

Do not mix infrastructure assets into service or app directories unless they are service-local
development fixtures.

## Backend Scripts

Backend automation goes under `backend/scripts/`.

Scripts should:

- Work from the `pos/` project root.
- Use local project cache directories when possible.
- Fail fast on errors.
- Be documented in `README.md`.

## Agent Rule

Before adding a new top-level directory under `pos/`, update this document and explain the
ownership boundary.
