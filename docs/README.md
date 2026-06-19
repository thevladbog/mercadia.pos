# Mercadia POS Platform Documentation

This folder describes the first implementation scope for the Mercadia POS platform.
The source designs are stored under `pos/docs/Design/Design` and cover four product
surfaces:

- POS terminal for cashier-operated sales.
- Senior cashier terminal for cash control, shift operations, and elevated approvals.
- Self-checkout, also called KSO/SCO, for customer-operated checkout and assistant mode.
- Admin panel for operations, configuration, monitoring, reconciliation, and master data.

## Documents

- [Functional overview](functional-overview.md) - shared concepts, roles, business objects, and end-to-end workflows.
- [POS terminal specification](modules/pos-terminal.md) - cashier checkout, payments, service operations, returns, and authentication.
- [Senior cashier terminal specification](modules/senior-cashier-terminal.md) - three-factor login, safe operations, cashier handover, EoD support, and audit journal.
- [Self-checkout specification](modules/self-checkout.md) - customer journey, assistant mode, age checks, marking, weighing, and error handling.
- [Admin panel specification](modules/admin-panel.md) - monitoring, cash office, EoD, users, catalog, integrations, devices, templates, and branding.
- [Architecture proposal](architecture.md) - suggested system architecture, bounded contexts, data model, deployment topology, integration patterns, and reliability strategy.
- [Architecture decisions](adr/README.md) - accepted and proposed architecture decision records.
- [Technology discussion](technology-discussion.md) - pending discussion for admin panel, backend, Store Edge packaging, and observability technology choices.
- [Development documentation](development/README.md) - AI agent rules, API contract workflow, and documentation backlog.
- [Open questions](open-questions.md) - decisions that should be resolved before implementation hardening.

## Reading Order

Start with the functional overview, then read the module documents for product behavior.
Use the architecture proposal to discuss implementation choices and integration boundaries.
The open questions document should become the working backlog for business, compliance,
hardware, and engineering decisions.

AI coding agents should also read `../AGENTS.md` and
`development/ai-agent-rules.md` before changing behavior. API-related work must follow
`development/api-contract-workflow.md`.

## Design Sources Reviewed

- Main POS PDF and exported POS PNG screens.
- Senior cashier PDF and exported senior cashier PNG screens.
- Self-checkout/KSO PDF and exported SCO PNG screens.
- Admin panel PDF and exported admin PNG screens.

Some exported file names are partially normalized by the operating system, so the
documentation references behavior visible in the screens rather than relying on file names.
