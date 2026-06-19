# Mercadia POS

Mercadia POS contains the store-facing platform implementation and documentation.

All POS-specific work must stay inside this directory. The repository root is reserved for other
Mercadia ecosystem projects.

## Current Backend Skeleton

- `backend/services/store-edge` - store-local operational API.
- `backend/services/central-backend` - central operational and synchronization API.
- `backend/services/hardware-agent` - local device API for fiscal devices, payment terminals, scanners,
  scales, drawers, printers, MSR, iButton, and other peripherals.
- `backend/packages/platform` - shared backend helpers for system endpoints, JSON responses, OpenAPI,
  and Scalar documentation.
- `contracts/openapi` - generated OpenAPI contract artifacts.

Backend details are documented in `backend/README.md`.

## Project Layout

All source code, generated contracts, scripts, and documentation for POS must stay under this
directory.

- `frontend/` - frontend applications, frontend packages, and generated frontend clients.
- `backend/` - backend workspace, services, shared backend packages, and backend scripts.
- `contracts/` - generated API contracts and future source contracts.
- `docs/` - documentation and ADRs.
- `infra/` - deployment and local infrastructure assets.

See `docs/development/repository-structure.md` for ownership rules.

## Versions And Security

Use current, actively maintained, and secure versions for runtimes, frameworks, tools, and
libraries. New dependencies must be justified, pinned, and verified against official support
status before they become part of the platform.

See `docs/development/dependency-policy.md` before adding or upgrading components.

## Commands

Run backend tests:

```powershell
.\backend\scripts\test.ps1
```

Regenerate OpenAPI contracts:

```powershell
.\backend\scripts\export-openapi.ps1
```

Run local services:

```powershell
Push-Location .\backend
go run .\services\store-edge\cmd\store-edge
go run .\services\central-backend\cmd\central-backend
go run .\services\hardware-agent\cmd\hardware-agent
Pop-Location
```

Default ports:

- Store Edge: `:8081`
- Central Backend: `:8082`
- Hardware Agent: `127.0.0.1:8083`

Each service exposes:

- `/healthz`
- `/readyz`
- `/openapi.json`
- `/docs`
