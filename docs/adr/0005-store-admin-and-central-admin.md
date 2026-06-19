# ADR-0005: Separate Store Admin And Central Admin Scopes

Status: Accepted

## Context

The admin design covers local store operations and broader configuration/reporting. Some
actions are store-local and operational, while others are network-level and central-office
controlled.

## Decision

Separate admin authority into two scopes:

- Store admin scope.
- Central admin scope.

The same UI system may reuse screens and components, but command targets and permissions must
be explicit.

Store admin operates through Store Edge and can manage:

- Store monitoring.
- Operational day.
- Safe and cash movements.
- Local terminals.
- Local cash operations.
- Store-level settings where allowed.

Central admin operates through central services and can manage:

- Cross-store reporting.
- Global integrations.
- Master data policies.
- Central user and permission policy.
- Global themes/templates.
- Aggregated reconciliation.

## Consequences

- Local operations remain available even during central outages.
- Central office can aggregate and govern without directly owning every local command.
- Permissions must include scope, not just role name.
- Some screens need dual mode: local store view and central aggregated view.

## Open Points

- Whether central office gets a separate product UI or a scoped mode in the same admin app.
- Exact store-local vs central-only section matrix.
- Conflict behavior between central policy and local override.
