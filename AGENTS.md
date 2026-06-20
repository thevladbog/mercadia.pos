# Mercadia POS Agent Guide

These instructions apply to all files under `pos/`.

## Required Reading

Before implementing or changing behavior, read:

1. `README.md`
2. `docs/README.md`
3. `docs/development/ai-agent-rules.md`
4. `docs/development/api-contract-workflow.md` when APIs or frontend data access are involved
5. `docs/development/repository-structure.md`
6. `docs/development/dependency-policy.md`
7. The relevant module spec in `docs/modules`
8. Relevant ADRs in `docs/adr`
9. `docs/development/admin-web-i18n.md` when changing `frontend/apps/admin-web`
10. `docs/development/ui-components.md` when changing `@mercadia/ui` or app theming

## Non-Negotiable Rules

- Store Edge is the authority for store-local operational state.
- Hardware Agent is the only layer that talks to fiscal devices, payment terminals, scanners,
  scales, drawers, printers, MSR, iButton, or vendor SDKs.
- Public HTTP APIs are generated from Go API definitions into OpenAPI.
- Scalar is used for interactive API reference.
- TypeScript frontends use Orval-generated clients and types.
- Frontend code lives under `frontend/`; backend code lives under `backend/`; generated
  contracts live under `contracts/`; infrastructure lives under `infra/`.
- Use current, actively maintained, and secure versions of all runtimes, frameworks, tools, and
  libraries.
- Cash ledger entries are immutable.
- Payment and fiscalization are separate state machines.
- Permission and separation-of-duties checks must be enforced server-side.
- Admin-web UI strings must use i18n (`docs/development/admin-web-i18n.md`); never hardcode user-facing labels in TSX.
- Shared interactive UI and theming live in `@mercadia/ui` (`frontend/packages/ui`);
  apps should use package components and `--ui-*` tokens instead of adding new global
  button/control styles.

## Commands

- Run backend tests with `.\backend\scripts\test.ps1`.
- Regenerate OpenAPI contracts with `.\backend\scripts\export-openapi.ps1`.
- Keep generated OpenAPI files in `contracts/openapi` up to date when API handlers change.

When a behavior is unclear, update `docs/open-questions.md` or ask for clarification before
inventing a rule.
