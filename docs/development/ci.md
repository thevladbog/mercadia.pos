# Continuous Integration

Mercadia POS uses a single GitHub Actions workflow ([`.github/workflows/ci.yml`](../../.github/workflows/ci.yml))
with path-based job selection. The workflow always starts on pull requests and pushes to
`main`; individual jobs are skipped when their paths did not change.

## What runs when

| Changed paths | Jobs |
| --- | --- |
| `backend/**`, `contracts/openapi/**`, `infra/migrations/**` | `backend` |
| `frontend/**`, `contracts/openapi/**` | `frontend` |
| `backend/**`, `frontend/**` | `codeql` |
| `.github/workflows/**` | all jobs (validates CI itself) |
| docs-only (`docs/**`, `*.md`) | `changes` + `ci-gate` only |

OpenAPI changes trigger both backend and frontend jobs because contracts are generated on the
backend and consumed by Orval on the frontend.

## Jobs

### `backend`

- `go test` for all backend modules
- `go vet`
- `golangci-lint` ([`backend/.golangci.yml`](../../backend/.golangci.yml))
- `govulncheck` per module
- OpenAPI export and diff check against committed artifacts in `contracts/openapi/`

### `frontend`

- `pnpm install --frozen-lockfile`
- Orval generation for central, store-edge, and hardware-agent clients
- Generated client diff check
- Typecheck, ESLint, Prettier, admin-web build
- `pnpm audit --audit-level=high` (blocking on high/critical)

### `codeql`

Static analysis for Go (`backend/`) and JavaScript/TypeScript (`frontend/`) when application
code changes.

### `ci-gate`

Final required check. Fails only when a non-skipped job failed or was cancelled. Skipped jobs
are treated as success so docs-only pull requests can merge without running backend or
frontend work.

## Local commands

Backend:

```powershell
.\backend\scripts\test.ps1
```

```bash
cd backend
go vet ./packages/platform/... ./services/...
golangci-lint run ./packages/platform/... ./services/store-edge/... ./services/central-backend/... ./services/hardware-agent/...
```

Frontend:

```bash
cd frontend
pnpm verify
pnpm audit --audit-level=high
```

OpenAPI regeneration:

```powershell
.\backend\scripts\export-openapi.ps1
```

## Branch protection

Recommended required status check:

- **`CI / ci-gate`**

Optional advisory checks (GitHub App integrations):

- GitGuardian Security Checks
- Socket Security
- Debricked vulnerability analysis

Also recommended: require pull request reviews, dismiss stale approvals on new commits, and
block force pushes to `main`.

## Dependency updates

Dependabot opens weekly update PRs for Go modules, npm packages, and GitHub Actions. See
[`.github/dependabot.yml`](../../.github/dependabot.yml).

## Pull requests

New pull requests use [`.github/pull_request_template.md`](../../.github/pull_request_template.md)
for summary, test plan, security impact, and contract change notes.
