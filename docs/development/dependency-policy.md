# Dependency And Version Policy

Mercadia POS must use current, actively maintained, and secure versions of all runtime,
framework, build, and library components.

## Baseline Rules

- Use stable releases only. Do not introduce alpha, beta, RC, nightly, or abandoned packages
  unless an ADR explicitly accepts the risk.
- Prefer actively supported LTS or current stable versions.
- Before pinning a new major runtime/framework/tool version, verify the current upstream support
  status from official sources.
- Keep version declarations close to the code that uses them: `go.mod`, `go.work`,
  `package.json`, lockfiles, Dockerfiles, and CI images.
- Do not copy version numbers from old examples, blog posts, or generated snippets without
  checking that they are still supported.
- Security fixes are allowed to move ahead of normal planning.

## Backend

- Go services must target an actively supported Go release.
- Go modules must not depend on unmaintained libraries for security-sensitive areas such as
  authentication, authorization, cryptography, payment, fiscalization, or transport security.
- New third-party packages require a reason: what problem they solve, why the standard library
  is not enough, and whether the package is maintained.
- Run `go test` after dependency changes.
- Run vulnerability scanning once tooling is added to the project.

## Frontend

- Frontend runtimes and frameworks must use supported stable versions.
- Node.js should use an active LTS/current line chosen for the project and pinned through the
  frontend workspace.
- TypeScript, React, Tauri, Vite, TanStack Query, Orval, and testing tools must be pinned in
  package manifests and lockfiles.
- Frontend code must use generated API clients/types instead of duplicating DTOs.
- Dependency updates must run typecheck, tests, linting, and Orval generation once those commands
  exist.

## API Tooling

- OpenAPI generation must come from Go API definitions.
- Scalar must be kept on a supported version when bundled or pinned.
- Orval must be kept on a supported version and regenerated after API changes.
- Generated OpenAPI and generated frontend clients must be reproducible from scripts.

## Security Review Triggers

A dependency change needs extra review when it touches:

- authentication or sessions;
- permissions or RBAC;
- payment flows;
- fiscalization;
- marked goods compliance;
- cash ledger logic;
- cryptography;
- local hardware communication;
- update/install mechanisms;
- logging or telemetry of sensitive data.

## Update Cadence

- Patch and security updates: apply as soon as practical after verification.
- Minor updates: review regularly during active development.
- Major updates: require a short ADR or technology note when they affect architecture,
  generated code, packaging, or runtime behavior.

## Agent Rules

AI agents must:

- Check existing version policy before adding packages.
- Prefer no dependency over a weak dependency.
- Use official documentation or release/support pages when deciding a version.
- Update manifests, lockfiles, generated clients, and docs together.
- Mention dependency and version changes in the final summary.
