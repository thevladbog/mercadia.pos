# AI Agent Rules

These rules apply to AI coding agents working on Mercadia POS.

## Core Rule

Do not invent behavior when product, fiscal, cash, or hardware rules are unclear. Use existing
docs first, then ask for clarification or add an explicit open question.

## Required Reading Order

Before implementing a feature, read:

1. `pos/docs/README.md`.
2. The relevant module spec in `pos/docs/modules`.
3. `pos/docs/architecture.md`.
4. Relevant ADRs in `pos/docs/adr`.
5. `pos/docs/development/api-contract-workflow.md` for API/client work.
6. `pos/docs/open-questions.md` when the feature touches unresolved business or legal rules.

## Architecture Rules

- Keep Store Edge as the authority for store-local operational state.
- Keep Hardware Agent as the only layer that talks to local hardware SDKs/protocols.
- Do not let UI clients call fiscal, payment terminal, scanner, scale, drawer, printer, MSR,
  or iButton SDKs directly.
- Keep payment state and fiscal state separate.
- Keep cash ledger immutable. Corrections are new operations, not edits to posted operations.
- Enforce separation of duties server-side. A user cannot approve their own critical operation.
- Use the same Store Edge command model for POS, SCO/KSO, senior cashier, assistant, and store
  admin where the business operation is the same.

## API Rules

- Public HTTP APIs must be described by generated OpenAPI.
- Frontend code must consume generated clients/types from Orval.
- Do not hand-write duplicate request/response TypeScript types when they can be generated.
- Every command endpoint must support idempotency.
- Every command must return stable error codes suitable for UI and support teams.
- Breaking API changes require updating the OpenAPI spec, generated clients, and affected docs.

## Frontend Rules

- Admin and senior cashier web interfaces are browser apps.
- POS, SCO/KSO, assistant station, and senior cashier touch terminal are terminal apps.
- Senior cashier web interface must not assume local MSR/iButton access.
- Senior cashier touch terminal may use Hardware Agent for MSR, iButton, and local signing.
- UIs must be dense and operational, not marketing-like.
- Critical actions require explicit confirmation and clear reason capture.
- **Admin-web (`frontend/apps/admin-web`):** all user-facing strings must go through i18n. See [`admin-web-i18n.md`](admin-web-i18n.md). Never add hardcoded UI labels in TSX.
- **Admin-web store-edge commands:** send `Idempotency-Key` on every POST command; invalidate React Query caches after `202`. See [`admin-web-i18n.md`](admin-web-i18n.md) §Store-edge integration.
- **Admin-web dev proxy:** route store-edge `/v1` prefixes to `:8081` before the central catch-all. Document new prefixes when adding store-edge API usage.
- **Cash UI:** collect `actorId` and `approvedById` for separation-of-duties; permissions are enforced server-side, not UI-only.
- **UI components:** use `@mercadia/ui` for interactive controls and theming. See [`ui-components.md`](ui-components.md). Do not add global button styles in apps; use `--ui-*` tokens or package components.

## Backend Rules

- Use Go for Store Edge, central operational backend, workers, integrations, and Hardware Agent
  unless an ADR says otherwise.
- Start with modular monolith boundaries, not premature microservices.
- Keep business logic out of HTTP handlers.
- Put command handlers, ledgers, state machines, permissions, and policies in application/domain
  packages.
- Use PostgreSQL transactions for state changes that produce events.
- Use outbox/inbox for reliable event publication and synchronization.

## Testing Rules

- Money, payment, fiscalization, returns, and permissions require tests.
- State machines must have transition tests.
- Cash ledger posting must have invariant tests.
- API changes must regenerate clients and compile affected frontend code.
- Hardware integrations must have mock/simulator tests before real hardware tests.
- Add regression tests for every fixed bug in checkout, cash, fiscal, payment, or sync logic.

## Documentation Rules

- Update module specs when behavior changes.
- Add or update ADRs when an architectural decision changes.
- Add open questions when product/legal/hardware behavior is unknown.
- Do not leave important behavior only in chat.

## Safety Rules

- Never remove audit history.
- Never rewrite posted cash movements.
- Never collapse payment success and fiscal success into one status.
- Never bypass role/permission checks in UI only.
- Never store raw PAN/card data.
- Never log secrets, PINs, full card numbers, or raw credential tokens.

## Useful Agent Workflow

1. Read relevant docs.
2. Identify the bounded context.
3. Check whether an ADR exists.
4. Check if OpenAPI/client generation is affected.
5. Implement the smallest coherent slice.
6. Add tests around state changes and permissions.
7. Regenerate OpenAPI/Orval clients when APIs changed.
8. Update docs and open questions.
9. Summarize changed behavior and verification.
