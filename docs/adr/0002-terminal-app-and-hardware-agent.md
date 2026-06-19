# ADR-0002: Terminal Application And Hardware Agent Split

Status: Accepted

## Context

Terminals must run rich operational UIs and integrate with local hardware: ATOL fiscal devices,
payment terminals, scanners, scales, cash drawers, receipt printers, and iButton readers.
Device SDKs and protocols should not leak into UI or domain logic.

Senior cashier functionality exists in two forms:

- Browser-based web interface for monitoring and management.
- Touch terminal interface for in-store money operations that require local hardware.

## Decision

Use a split terminal architecture:

- React + Tauri as the preferred terminal UI shell.
- Go Hardware Agent as a local service installed next to the terminal app.
- Store Edge API for business commands and operational state.
- Hardware Agent API for local device operations and health.

Terminal UI must not call fiscal, payment, scanner, printer, scale, drawer, or iButton SDKs
directly.

Senior cashier hardware-backed operations must run on the touch terminal surface. The web
interface may show the same operational state and initiate safe browser-compatible workflows,
but MSR, iButton, and local signing go through the Hardware Agent on the touch terminal.

## Consequences

- UI remains portable and focused on interaction logic.
- Device integrations can evolve without rewriting POS/SCO screens.
- Go is a good fit for long-running local services, device protocols, concurrency, and
  single-binary deployment.
- Tauri gives a lighter desktop shell than Electron while still allowing a modern React UI.
- We must define and version the local Hardware Agent API early.
- Senior cashier UX must be designed as two related surfaces, not one screen merely resized.
- Permission checks must understand whether the current surface has hardware-backed factors
  available.

## Open Points

- Whether all terminal types use Tauri, or admin remains browser-only.
- Local API protocol: HTTP, gRPC, or both.
- Service supervision/install/update mechanism.
- Supported Windows/Linux matrix per device vendor.
- MSR reader model/protocol.
- iButton reader model/protocol.
