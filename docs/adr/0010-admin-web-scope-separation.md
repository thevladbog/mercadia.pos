# ADR-0010: Admin-Web Scope Separation (Store vs Central)

Status: Proposed

## Context

ADR-0005 separates store admin authority from central admin authority: "The same UI
system may reuse screens and components, but command targets and permissions must
be explicit." ADR-0005 also left an open point: "Whether central office gets a
separate product UI or a scoped mode in the same admin app." `docs/open-questions.md`
records the same unresolved intent: "в идеале - у центрального офиса будет
отдельная админка, данные в которую будут стекаться из магазинных" (ideally the
central office gets its own admin panel, fed by data synced from stores).

In practice, `admin-web` (`frontend/apps/admin-web`) is a single React app already
serving both scopes. Its `src/pages/` directory holds 32 page components, routed in
`src/App.tsx`:

- 16 pages under `/central/**`: `CentralDashboardPage`, `CentralReportingPage`,
  `CentralStoresPage`, `CentralSyncExplorerPage`, `CentralCatalogPage`,
  `CentralUsersPage`, `CreateCentralUserPage`, `EditCentralUserPage`,
  `CentralColorSchemesPage`, `CreateColorSchemePage`, `EditColorSchemePage`,
  `CentralLayoutTemplatesPage`, `CreateLayoutTemplatePage`,
  `EditLayoutTemplatePage`, `RegisterStorePage`, and `SyncEntityDetailPage`
  (mounted five times, once per synced entity kind).
- 6 pages under `/store/**`: `StoreMonitoringPage`, `StoreSafePage`,
  `StoreEodPage`, `StoreCredentialManagementPage`, `StoreSettingsPage`, and
  `TerminalMonitoringDetailPage`.
- 9 pages under `/senior-cashier/**`: `SeniorCashierDashboardPage`,
  `IssueChangeFundPage`, `ReceiveCashPage`, `SafeRecountPage`,
  `BankCollectionPage`, `BusinessExpensePage`, `FinalCollectionPage`,
  `OperationJournalPage`, `ShiftHandoverPage`. Eight of these nine (all but the
  dashboard) exist as separately-implemented but identically-named page
  components in the standalone `senior-cashier-terminal` app
  (`frontend/apps/senior-cashier-terminal/src/pages/`) — the same
  cash-operation screens are built twice.
- `StoreReportingPage` is the outlier: named like a store page, but its only
  route is `/central/reporting/stores/:storeId`, nested under the central
  reporting tree. The `Store*`/`Central*` filename convention and the actual URL
  scope already disagree in at least this one place, which shows a naming
  convention alone is not a reliable scope boundary.

Roles are scoped per backend, not shared. store-edge
(`backend/services/store-edge/internal/app/rbac.go`) defines `cashier`,
`senior_cashier`, `admin` with its own `rolePermissions`/`HasPermission` table.
central-backend defines `central_viewer` and `central_admin`
(`backend/README.md`: "Central roles in v1: `central_viewer` (reporting read) and
`central_admin` (reporting + user management)"). `senior_cashier` also appears as
a role value on the *central* user-list type
(`ListCentralUsers200UsersItem['roles']`), which
`frontend/apps/admin-web/src/auth/permissions.ts` reads from to gate cash
operations (`canWriteCashOperations`, `canWriteStoreOperations`,
`isSeniorCashier`) — so pages that act against store-edge are gated by a role
carried on the central session.

Session handling is already split between backends, but not as a designed scope
boundary. `admin-web` configures two independent generated API clients,
`@mercadia/api-clients-central` and `@mercadia/api-clients-store-edge`
(`frontend/apps/admin-web/src/auth/api-client-config.ts`), each with its own
`sessionStorage` key and its own `setSessionToken`/`getSessionToken` pair
(`frontend/packages/api-clients/central/src/session.ts` vs
`frontend/packages/api-clients/store-edge/src/session.ts`). App-level login
(`frontend/apps/admin-web/src/auth/LoginPage.tsx`, `AuthProvider.tsx`) only ever
establishes a central session, via `useCreateCentralAuthSession` and the central
client's `setSessionToken`. There is no equivalent app-level login for
store-edge: `StoreCredentialManagementPage.tsx` and `StoreSettingsPage.tsx` each
call `createAuthSession`/`setSessionToken` from
`@mercadia/api-clients-store-edge` directly, inline in the page component, to
obtain a store-edge session on demand. This is a genuine separation of transport
and credentials, but it is incidental per-page plumbing rather than an
architected boundary — nothing guarantees a new page picks the right client or
gets scope-explicit navigation.

## Options

**A. Single app, hard internal scope split.** Keep one `admin-web` deployable,
but make the store/central boundary a first-class internal contract: route
trees strictly under `/store/**` and `/central/**` (already mostly true —
`StoreReportingPage` is the one exception to fix), a scope-aware session/auth
layer in place of ad hoc per-page token calls, and navigation that only surfaces
links inside the current scope. Shared visual/behavioral components stay in
`@mercadia/ui`; only page-level code is scope-exclusive. Lowest short-term
migration cost, since routing already leans this way; the risk is that an
"internal contract" erodes over time without tooling (lint rule, code owners)
to enforce it.

**B. Two apps.** Split into `admin-store` and `admin-central`, each its own
build/deployable, both depending on `@mercadia/ui` and the generated Orval API
clients (`@mercadia/api-clients-central`, `@mercadia/api-clients-store-edge`).
Gives the cleanest boundary — independent auth, independent bundles,
independent release cadence — and is the direct realization of "central office
gets its own admin panel" from `docs/open-questions.md`. Cost: two build/deploy
pipelines, duplicated app-shell code (layout, routing, auth bootstrap) unless
that shell is itself extracted to a shared package, and a migration that
touches every existing page.

**C. Status quo.** Keep the single role-filtered route tree as-is, with no
enforced internal boundary beyond current `Central*`/`Store*` naming and the
`RequireCentralAdmin`/`RequireSeniorCashierOrAdmin` route guards. Cheapest
today. The cost compounds silently: every new admin feature is one more page
that can import across scopes or reuse the wrong session, and an eventual split
(Option B) only gets more expensive the longer this continues — already visible
in the `StoreReportingPage` routing mismatch and the ad hoc store-edge session
calls described above.

## Decision

Adopt **Option A now**: keep `admin-web` as one app, but treat the store/central
scope boundary as an enforced internal contract rather than a naming
convention. Concretely — and as the baseline for the Consequences below —
route trees stay strictly under `/store/**` and `/central/**` (fixing the
`StoreReportingPage` exception), session/auth handling for the two backends is
made explicit and scope-aware instead of ad hoc, and no page imports another
scope's page-level code.

Graduate to **Option B (two apps)** when any of the following triggers is hit,
whichever comes first:

- Central admin needs an independent release cadence from store admin (for
  example, central rolls out weekly while store admin is pinned to
  hardware-qualified builds).
- Central admin gets its own auth domain or SSO requirement that store-edge
  sessions do not need (for example, a corporate IdP login required only for
  central users).
- Store-side bundle size or load time becomes a problem on in-store hardware
  because central-only code (catalog, cross-store reporting, layout templates)
  ships in the same bundle as store screens.

This is a proposal for maintainer review; Options B and C remain live
alternatives until this ADR is marked Accepted.

## Consequences

- New admin-web pages must declare which scope (`store` or `central`) they
  belong to through their route path (`/store/**` or `/central/**`); no new
  page may be reachable from both trees the way `StoreReportingPage` is today.
- No page-level module may import another scope's page-level module. Shared UI
  goes through `@mercadia/ui`; shared non-visual logic goes through a shared
  package, not a cross-scope import.
- Store-edge and central sessions must be obtained through one explicit,
  scope-aware auth flow, not ad hoc per-page `createAuthSession` calls like the
  ones in `StoreCredentialManagementPage.tsx` and `StoreSettingsPage.tsx` today.
  Existing ad hoc call sites should converge on this flow as they are touched.
- The cash-operation pages duplicated between `admin-web`'s
  `/senior-cashier/**` tree and `frontend/apps/senior-cashier-terminal` are a
  known cost of the current architecture, independent of this decision. Per
  `AGENTS.md`'s shared-component guidance, they should eventually converge on
  shared package components rather than be re-implemented per app; this ADR
  does not mandate a timeline but flags it as work that should not get harder
  while Option A is in effect.
- If a graduation trigger fires, splitting becomes a routing/deployment change
  rather than an architecture change, because the scope boundary was already
  enforced.
