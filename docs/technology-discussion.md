# Technology Discussion

This document is a placeholder for the next architecture discussion: admin panel and backend
technology choices.

## Recommended Default Stack

My recommended starting point:

- Admin panel frontend: React + TypeScript + Vite.
- Terminal frontend: React + TypeScript + Tauri.
- Senior cashier web interface: React + TypeScript + Vite as part of the admin/store-admin
  frontend.
- Senior cashier touch terminal: React + TypeScript + Tauri with Hardware Agent access.
- UI primitives: Radix UI + custom Mercadia design system.
- Tables/grids/forms: TanStack Table, TanStack Query, React Hook Form or TanStack Form after
  evaluation.
- Backend and Store Edge: Go.
- Hardware Agent: Go.
- Primary database: PostgreSQL.
- Messaging: NATS JetStream as the preferred default; RabbitMQ as acceptable fallback.
- API style: HTTP command/query APIs documented with OpenAPI, plus WebSocket/SSE for live
  monitoring.
- API contract tooling: OpenAPI generated from Go, Scalar for API reference, Orval for
  frontend clients/types/hooks/mocks.
- Internal events: versioned JSON first; Protobuf can be introduced later if performance,
  schema governance, or multi-language consumers justify it.
- Service shape for MVP: modular monoliths with strict bounded-context packages, not many
  microservices from day one.
- Observability: OpenTelemetry, Prometheus-compatible metrics, structured JSON logs, and
  Grafana-family dashboards.

The main principle: keep the first production architecture boring where money and fiscalization
need reliability, and use sharper tools only where they remove real operational complexity.

## Decisions To Make

### Admin Panel Frontend

Questions:

- Should the admin panel be a browser-only web application?
- Should store admin and central admin be one app with scoped permissions or separate apps?
- Which frontend stack should be used?
- Which UI component system should be used for dense operational interfaces?
- How should real-time monitoring be implemented in the UI?

Initial candidate:

- React + TypeScript.

Recommendation:

- Use React + TypeScript + Vite for the admin panel.
- Keep admin as a browser web app, not Tauri.
- Implement store admin and central admin as one frontend codebase with explicit scope and
  permission boundaries.
- Include the senior cashier web interface in the same browser frontend family where practical.
  It should reuse admin/store-admin components and Store Edge APIs, but must not assume direct
  access to local MSR/iButton hardware.
- Use TanStack Query for server state and cache lifecycle.
- Use TanStack Table for dense operational grids.
- Use Radix UI primitives and `@mercadia/ui` (custom Mercadia component layer with CSS-variable theming) instead of adopting a heavy
  opinionated admin template.

Why:

- Admin is an operational web UI, not a hardware-connected kiosk.
- Vite keeps the frontend simple and fast without forcing full-stack framework semantics.
- React can be shared mentally and technically with terminal UIs.
- TanStack Query fits data that is remote, shared, refetched, invalidated, and sometimes stale.
- The admin needs dense tables, filters, drawers, command panels, and realtime status rather
  than marketing-style screens.

Topics to compare:

- React + Vite.
- Next.js if server-side rendering or full-stack routing is useful.
- TanStack Router/Query/Table for operational UI.
- Component system: custom design system, shadcn/ui, Radix primitives, or enterprise grid
  components where needed.

### Backend

Questions:

- Should backend services be primarily Go, TypeScript/Node.js, Java/Kotlin, or another stack?
- Should Store Edge and central backend use the same language?
- How many services should exist at MVP?
- Should we start modular monolith plus clear bounded contexts, or separate services from day one?
- Which API style should be used for commands and real-time streams?

Initial candidate direction:

- Go for Store Edge, Hardware Agent, and high-reliability operational services.
- PostgreSQL for transactional state.
- NATS JetStream or RabbitMQ for asynchronous messaging.
- OpenAPI for command/query HTTP APIs.
- WebSocket or Server-Sent Events for live monitoring.
- Scalar for interactive API reference.
- Orval for frontend client generation.

Senior cashier interface split:

- Browser/web senior cashier functions use the same Go Store Edge APIs as store admin.
- Touch-terminal senior cashier functions use Store Edge APIs plus the Go Hardware Agent for
  MSR, iButton, and other device-backed factors.

Recommendation:

- Use Go for Store Edge and central operational backend.
- Generate OpenAPI from Go API definitions. Huma is the preferred candidate to validate in the
  backend skeleton, but the workflow matters more than the exact library.
- Start as a modular monolith per deployment scope:
  - Store Edge modular monolith.
  - Central backend modular monolith.
  - Hardware Agent separate Go service.
- Split into separate services only when a bounded context has independent scaling,
  deployment, security, or ownership needs.
- Keep business logic out of HTTP handlers and UI clients.
- Put command handlers, state machines, ledgers, and policy checks in domain/application
  packages.
- Use Scalar to expose API references in development/staging.
- Use Orval on all TypeScript frontends so API calls, DTO types, React Query hooks, and mocks
  are generated from OpenAPI.

Why:

- Go is a strong fit for long-running services, concurrency, networking, device agents, and
  simple deployment.
- A modular monolith avoids early distributed-system tax while still preserving boundaries.
- Store Edge must be easy to install, observe, and recover in a store; fewer processes is a
  feature.
- Central services can evolve into services later once operational boundaries are proven.

Suggested backend package boundaries:

- identity-access.
- store-operations.
- checkout.
- payments.
- fiscalization.
- cash-office.
- catalog-pricing.
- loyalty-promotions.
- marking-compliance.
- reconciliation.
- audit.
- integration-adapters.

### API Contract Tooling

Questions:

- Which Go OpenAPI generator should be used?
- Should Huma be adopted as the HTTP API framework for Store Edge and central backend?
- Should Orval generate fetch clients, React Query hooks, MSW mocks, and validators for every
  frontend package?
- Where should generated OpenAPI and frontend clients live in the repository?

Recommendation:

- Use code-first OpenAPI generation from Go.
- Use Huma as the first candidate for a backend skeleton because it generates OpenAPI 3.1 and
  JSON Schema from Go API definitions.
- Serve OpenAPI through Scalar in local development and staging.
- Use Orval to generate TypeScript clients and TanStack Query hooks for admin, senior cashier
  web, terminal UIs, and assistant station.
- Make CI fail when OpenAPI or generated clients are stale.

Why:

- OpenAPI gives one contract for backend, frontend, tests, mocks, docs, and agents.
- Scalar gives developers and QA a browsable API reference without maintaining separate docs.
- Orval prevents frontend DTO drift and can generate mocks for UI work before backend flows are
  complete.
- AI agents are much safer when API shapes are machine-readable and generated.

### Store Edge Packaging

Questions:

- Is Store Edge installed directly on the store server/terminal OS?
- Is Docker/Podman acceptable in stores?
- How are PostgreSQL, Store Edge, broker bridge, and Hardware Agent installed and updated?
- Who operates local backups and diagnostics?

Recommendation:

- Package Store Edge as a managed local service bundle.
- Avoid Kubernetes in stores for MVP.
- Prefer direct OS services or a simple container bundle depending on target store operations.
- Bundle or provision PostgreSQL with automated migrations and backup scripts.
- Keep the local broker optional until we confirm whether Store Edge needs local stream replay;
  outbox workers are mandatory either way.

Packaging candidates:

- Windows service + Linux systemd service.
- Installer-managed PostgreSQL.
- Optional Docker/Podman only if store IT is comfortable operating containers.

My bias:

- For MVP: OS services are safer than asking every store to operate containers.
- For controlled pilots: Docker Compose can be acceptable if deployment is handled centrally.

### Observability

Questions:

- Which logs/metrics/traces stack should be used centrally?
- How much observability must work locally when central is unavailable?
- What is the minimum health dashboard for store technicians?

Recommendation:

- Use OpenTelemetry conventions for traces/metrics/log correlation.
- Emit Prometheus-compatible metrics from Store Edge and Hardware Agent.
- Use structured JSON logs everywhere.
- Provide a local health UI for store technicians and a central dashboard for operations.
- Track outbox lag, sync lag, device health, fiscal failures, payment failures, receipt
  exceptions, cash mismatches, and EoD blockers from day one.

Initial tools to consider:

- Prometheus or VictoriaMetrics for metrics.
- Grafana for dashboards.
- Loki or another log backend for centralized logs.
- OpenTelemetry Collector for local/central forwarding.

Avoid making observability dependent only on central connectivity. Store Edge must expose
enough local health information to debug a store during an outage.

## ADR Candidates

- Admin panel frontend stack.
- Backend language/runtime.
- MVP service decomposition.
- Store Edge packaging.
- Broker selection.
- Observability stack.
