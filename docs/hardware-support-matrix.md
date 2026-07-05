# Hardware Support Matrix And Real-Driver Architecture (Spike)

Status: Draft (design gate for the first real driver implementation).

## Why this document exists

Every device path in the Hardware Agent
(`backend/services/hardware-agent/internal/infra/simulated/`) is simulated today —
there is no real driver for any device kind. A production store cannot ring up its
first sale (fiscal registrar) or take a card (payment terminal) until real
integrations exist. This document inventories what exists in code today, proposes
how a real driver plugs into the existing seam without breaking the ADR-0002
boundary, and lists what must be verified with vendors before implementation
starts. It is a design document; no code changes are included or implied by
merging it.

## 1. Device support matrix

`DeviceKind` is defined in
`backend/services/hardware-agent/internal/domain/device.go:11-17` and currently
has six values. The cash drawer is not a `DeviceKind` — per
`docs/open-questions.md:67` ("Как управляется денежный ящик? - фискальным
регистратором") it is a function of the fiscal registrar, so it is listed as a
row that maps onto the `fiscal` kind rather than a device of its own.

| Device kind | First target vendor/model | Protocol/SDK | Current status | Store Edge consumer | Priority |
|---|---|---|---|---|---|
| `fiscal` | ATOL (`docs/open-questions.md:6`, `:65`). Simulated seed uses model string `ATOL 42F` (`infra/memory/store.go:32`) — this is a placeholder label chosen for the simulator, not a confirmed model number. | **TBD/verify** — ATOL driver version (Driver 10/11), transport (USB vs TCP/IP), DTO/JSON vs COM API. Nothing about ATOL's SDK is vendored in this repo, so no protocol detail below "ATOL" is verifiable from code. | Simulated only (`infra/simulated/fiscal.go`) | `print_receipt` (receipt + return fiscalization, `internal/app/fiscalization.go:153`, `:264`, wired via `WithFiscalReceiptPrinter` in `internal/api/server.go:715`) | **P0** — blocks first sale |
| `payment_terminal` | **TBD/verify** — not decided. `docs/open-questions.md:16` ("Какой эквайер и какой вендор платежных терминалов должны поддерживаться первыми? - Популярные в РФ") only says "popular in Russia," no vendor named. Simulated seed uses model string `Ingenico iPP350` (`infra/memory/store.go:33`) as a placeholder, not a decision. | **TBD/verify** — acquirer + terminal vendor unresolved; protocol per `docs/open-questions.md:66` ("поддерживаем популярные протоколы") is equally unresolved. | Simulated only (`infra/simulated/payment_terminal.go`) | `authorize`, `capture`, `cancel`, `refund` (`internal/app/payments.go`, hardcoded device id `sim-payment-1` via `WithCardPaymentTerminal` in `internal/api/server.go:714`) | **P0** — blocks first card sale |
| `msr` (magnetic stripe reader) | Placeholder seed model `MagTek MSR605` (`infra/memory/store.go:34`) — not a vendor decision, ADR-0002 Open Points still lists "MSR reader model/protocol" as unresolved (`docs/adr/0002-terminal-app-and-hardware-agent.md:50`). | **TBD/verify** | Simulated only (`infra/simulated/msr.go`) | Not called by store-edge. Consumed directly by terminal frontends (`frontend/apps/senior-cashier-terminal/src/auth/ibutton.ts` and sibling MSR auth code) through the Hardware Agent HTTP API, per ADR-0002's rule that MSR/iButton go through the Hardware Agent on the touch terminal. | P1 — senior cashier auth factor |
| `ibutton` | Placeholder seed model `DS9490R` (`infra/memory/store.go:35`) — ADR-0002 Open Points still lists "iButton reader model/protocol" as unresolved (`docs/adr/0002-terminal-app-and-hardware-agent.md:51`); `docs/open-questions.md:69` says "не готов ответить" (not ready to answer). | **TBD/verify** | Simulated only (`infra/simulated/ibutton.go`) | Not called by store-edge. Consumed by terminal frontends (senior cashier three-factor login) via the Hardware Agent HTTP API. | P1 — senior cashier auth factor |
| `scanner` | Placeholder seed model `Honeywell 1900` (`infra/memory/store.go:36`). `docs/open-questions.md:64` says scanners are usually COM, possibly USB — no specific model chosen. | **TBD/verify** — COM vs USB HID keyboard-wedge behavior differs materially for a real driver. | Simulated only (`infra/simulated/scanner.go`) | Not called by store-edge directly; consumed by POS/SCO terminal frontends via the Hardware Agent HTTP API for line-item and staff-card scanning. | P0 — blocks scan-based checkout |
| `printer` (receipt printer, separate from the fiscal document) | Epson, per `docs/open-questions.md:70` ("возьмем Epson за основу"). Seed model string `Epson TM-T88` (`infra/memory/store.go:37`) is consistent with this decision but the exact model/firmware is unconfirmed. | **TBD/verify** — ESC/POS command set, ribbon/paper sensor exposure. | Simulated only (`infra/simulated/printer.go`) | Not called by store-edge today (fiscal flow uses `print_receipt` on the `fiscal` kind, not this `printer` kind). Non-fiscal printer use (if any, e.g. kitchen/service slips) is not implemented. | P2 |
| cash drawer (not a `DeviceKind`) | Driven by the fiscal registrar, i.e. ATOL, per `docs/open-questions.md:67`. | **TBD/verify** — ATOL drawer-kick command as part of the fiscal driver. | Not modeled as a device; no dedicated command exists in any simulated adapter today. | None directly — would ride on the `fiscal` executor once a real ATOL driver exists. | P1 — needed once a real fiscal driver opens/closes the drawer as a side effect |

Scales/weighing devices are **absent from `DeviceKind` entirely** — see Open
Questions §5.4 below.

## 2. Current architecture

- `DeviceExecutor` is the seam a real driver implements:
  `Execute(ctx, device domain.Device, commandType string, payload map[string]any) (map[string]any, error)`
  (`internal/app/devices.go:33-35`). `DeviceService` takes exactly one executor
  (`NewDeviceService(...)`, `internal/app/devices.go:51-68`); today that one
  executor is `simulated.DefaultRegistry()` (`internal/api/server.go:66`), a
  `Registry` that dispatches by `DeviceKind` to one `Adapter` per kind
  (`internal/infra/simulated/adapter.go:13-47`).
- Command lifecycle: `SendCommand` validates idempotency, persists the command
  as `accepted`, and asynchronously runs it (`internal/app/devices.go:152-205`,
  `:211-237`), transitioning to `running` then `completed`/`failed`
  (`domain.CommandStatus`, `internal/domain/device.go:29-36`). Callers poll
  `GET /v1/devices/{deviceId}/commands/{commandId}` or block via the
  store-edge client's `WaitCommand` (`internal/infra/hardwareagent/client.go:153-181`).
- Idempotency: `SendCommand` requires an `Idempotency-Key` header
  (`internal/api/server.go:186-189`, enforced via `RequiresIdempotency`) and
  fingerprints `(deviceID, commandType, payload)` so a retried request with the
  same key and the same payload returns the original result, while a reused key
  with a different payload is rejected (`internal/app/devices.go:239-256`,
  `:258-263`).
- `DeviceStatus` includes a dedicated `simulated` value alongside `ready`,
  `busy`, `offline`, `error` (`internal/domain/device.go:21-27`); every seeded
  device is created with `Status: domain.DeviceStatusSimulated`
  (`internal/infra/memory/store.go:32-37`).
- Store Edge's fallback: when `MERCADIA_STORE_EDGE_USE_HARDWARE_AGENT=true`,
  `store-edge` wires the Hardware Agent client into payments and fiscalization
  (`internal/api/server.go:712-715`); a separate
  `MERCADIA_STORE_EDGE_HARDWARE_AGENT_FALLBACK` flag (default `true`,
  `backend/README.md:55`) lets card payments and fiscalization fall back to
  the pre-existing mock path (`card_mock`) if the Hardware Agent call fails, so
  a store can keep selling even if the agent or device is down
  (`backend/README.md:196-202`).

## 3. Real-driver architecture (proposal)

**One executor per vendor, registered per `DeviceKind`, same as today's
simulated registry.** The existing shape already argues for this: `Registry` in
`internal/infra/simulated/adapter.go:18-28` maps one `Adapter` per
`domain.DeviceKind`, and `DeviceExecutor.Execute` receives `commandType` as a
plain string — it has no notion of "vendor," only "kind of device plus command
name." The cleanest extension is a **mixed registry**: keep the `Adapter`
interface, add a real adapter (e.g. `atol.NewFiscalAdapter()`) that implements
it for `DeviceKindFiscal`, and construct the registry with a mix of real and
simulated adapters per environment:

```go
simulated.NewRegistry(
    atol.NewFiscalAdapter(cfg.ATOL),      // real
    simulated.NewPaymentTerminalAdapter(), // still simulated until a vendor is chosen
    simulated.NewMSRAdapter(),
    simulated.NewIButtonAdapter(),
    simulated.NewScannerAdapter(),
    simulated.NewPrinterAdapter(),
)
```

This requires no change to `DeviceExecutor`, `DeviceService`, or the command
lifecycle — `DeviceService.SendCommand` and `runCommand`
(`internal/app/devices.go:152-237`) are agnostic to what implements
`DeviceExecutor`. A per-kind driver registry (one executor implementation per
vendor per kind, not a single monolithic `DeviceExecutor` per store) fits the
existing code shape better than any single-executor-does-everything design,
because the current `Registry` already dispatches on `device.Kind`
(`internal/infra/simulated/adapter.go:31-36`) and each simulated adapter file
is already scoped to exactly one kind.

**Config-driven device inventory.** Devices are seeded in code today
(`Store.seedDevices`, `internal/infra/memory/store.go:29-42`) with hardcoded
IDs like `sim-fiscal-1` and `sim-payment-1` that store-edge also hardcodes
(`internal/api/server.go:714`, device id `sim-payment-1`). A real deployment
needs devices loaded from configuration (device id, kind, model, connection
string/COM port/IP, driver selection) rather than compiled into the binary,
so a store can be provisioned with its actual device list without a code
change. This is a new `DeviceRepository` implementation (interface already
exists at `internal/app/devices.go:23-26`) backed by config or a small local
store, not a change to the domain model.

**Error/status mapping into existing enums.** A real driver's failures (paper
out, device offline, timeout, protocol NAK) must map into the existing
`DeviceStatus` (`ready`/`busy`/`offline`/`error`, plus `simulated` for devices
still running the simulator) and `CommandStatus` (`accepted`/`running`/
`completed`/`failed`, `internal/domain/device.go:21-36`). No new enum values
are proposed here; a real driver returning `error` from `Execute` already
surfaces as `CommandStatusFailed` with `Error` populated
(`internal/app/devices.go:229-233`), so the mapping problem is really "what
`DeviceStatus` does the device report at rest," which is a per-adapter
`get_status` implementation concern (see fiscal's `status()` helper,
`internal/infra/simulated/fiscal.go:72-88`, as the shape to replicate with real
driver introspection instead of hardcoded fields).

**Mixed fleets (real fiscal + simulated scanner) work today by construction.**
Because the registry dispatches per `device.Kind` and each `Device` record
carries its own `Kind` and `Status`
(`internal/domain/device.go:43-49`), a store can run a real ATOL fiscal driver
side by side with simulated MSR/scanner/iButton adapters simply by seeding a
mixed device list and a mixed registry — no architectural change is needed,
only (a) the config-driven inventory above and (b) real adapters for the
kinds that have graduated. The `simulated` `DeviceStatus` value should be kept
exactly as "this device is a stand-in, not connected to hardware" — it becomes
a useful signal in mixed fleets for admin/monitoring UIs to flag which devices
in a store are not yet real.

## 4. ATOL-first integration plan

Fiscal is P0 (blocks the first sale) and is the only kind with a named vendor
decision (`docs/open-questions.md:6`, `:65`), so it is the first real driver.

**(a) Transport/protocol decision — to verify with ATOL documentation:**
- Driver generation/version (ATOL Driver 10 vs 11, or a newer SDK) — **TBD/verify**.
- Transport: USB vs TCP/IP (network-attached fiscal registrars are common for
  ATOL but must be confirmed per target hardware model) — **TBD/verify**.
- Whether integration goes through ATOL's native COM/DLL driver (Windows-only)
  or a documented HTTP/JSON gateway — this materially affects whether the
  Hardware Agent needs a Windows-only build tag or CGO bridge. **TBD/verify**,
  and it interacts with ADR-0002's open point "Supported Windows/Linux matrix
  per device vendor" (`docs/adr/0002-terminal-app-and-hardware-agent.md:49`),
  which is still unresolved.
- Fiscal document numbering/duplicate-shift protection semantics — **TBD/verify**,
  needed because `DeviceService.SendCommand` idempotency only protects the
  Hardware Agent's own command layer (`internal/app/devices.go:152-167`), not
  the fiscal registrar's own document sequence.

**(b) Minimal command set needed by store-edge's current fiscal flow.**
Store-edge only issues one fiscal command type today: `print_receipt`, called
from `FiscalizationService.Fiscalize`/return fiscalization
(`internal/app/fiscalization.go:153`, `:264`) through the
`FiscalReceiptPrinter` interface (`internal/app/fiscalization.go:28-29`,
`PrintReceipt(ctx, deviceID, totalMinor) (string, error)`), which itself calls
the Hardware Agent's `print_receipt` command
(`internal/infra/hardwareagent/client.go:261-280`). A real ATOL driver's
minimal viable command set to unblock store-edge is therefore just
**`print_receipt`** — matching the simulated adapter's existing contract
(`internal/infra/simulated/fiscal.go:44-61`: accepts `totalMinor`, must return
a non-empty `fiscalSign`, per the client's own validation at
`internal/infra/hardwareagent/client.go:275-278`). The simulated adapter also
implements `get_status`, `open_shift`, `close_shift`, and `cancel_receipt`
(`internal/infra/simulated/fiscal.go:27-69`) — none of these are invoked by
store-edge today, but shift lifecycle (`open_shift`/`close_shift`) will be
needed once EoD/shift-open flows drive the real registrar directly rather than
only recording business-layer shift state, and the cash-drawer-kick behavior
(§1 above) likely attaches to `open_shift`/`print_receipt` on a real driver
even though no command surfaces it explicitly today.

**(c) Test strategy without hardware.**
- Contract tests: assert a real ATOL adapter satisfies the same
  request/response shape the simulated adapter already guarantees — same
  `commandType` switch cases, same required result fields (`fiscalSign`,
  `driverState`, `shiftState`) as `internal/infra/simulated/fiscal.go:26-69`,
  so `FiscalizationService` and the Hardware Agent HTTP contract
  (`internal/api/server.go:83-208`) don't need to change when the adapter is
  swapped.
- Golden-path integration test: reuse `store-edge`'s existing
  `MERCADIA_STORE_EDGE_USE_HARDWARE_AGENT=true` + fallback wiring
  (`backend/README.md:196-202`) against a real Hardware Agent process running
  the ATOL adapter, asserting the same store-edge API responses as the
  simulated path today.
- Hardware-in-the-loop checklist (for later, once physical devices are
  available): power-cycle recovery, paper-out/cover-open error surfacing into
  `DeviceStatus`, shift-open/close against the actual OFD, drawer kick timing,
  and a duplicate-command replay test against the idempotency layer
  (`internal/app/devices.go:239-256`) to confirm a network retry does not
  double-print a fiscal document.

## 5. Open questions

1. **Payment terminal vendor and acquirer** (owner: product). Not decided
   anywhere in this repo — `docs/open-questions.md:16` only says "popular in
   Russia" ("Популярные в РФ"). This blocks starting a real payment terminal
   driver the same way ATOL unblocks fiscal; until a vendor is named, the
   `payment_terminal` row in §1 has no target beyond the placeholder simulator
   model string.
2. **Payment terminal protocol** (owner: vendor docs, gated by #1). Also
   unresolved (`docs/open-questions.md:66`, "поддерживаем популярные
   протоколы" — "we support popular protocols," no protocol named).
3. **Scales/weighing devices have no `DeviceKind` at all** (owner:
   engineering + product). `docs/modules/self-checkout.md` requires scale
   integration for weighed goods (`docs/modules/self-checkout.md:39`, `:146`,
   `:148`, `:267`), but `domain.DeviceKind`
   (`internal/domain/device.go:11-17`) has no `scale`/`scales` value and no
   simulated adapter exists for one. `docs/open-questions.md:68` confirms the
   model/protocol is also undecided ("используем популярный" — "we'll use a
   popular one"). This is a real gap: a `DeviceKindScale` (or equivalent) and
   at least a simulated adapter are needed before self-checkout weighing can
   be implemented against the Hardware Agent, not just before a real driver
   exists.
4. **Cash drawer: stays fiscal-registrar-driven, or becomes its own
   `DeviceKind`?** (owner: engineering, informed by #5 ATOL verification).
   Today's answer (`docs/open-questions.md:67`) is that the fiscal registrar
   drives the drawer, and no command in any adapter models a drawer kick
   explicitly. If a store's ATOL unit does not expose drawer control (or a
   future store uses a registrar that doesn't), a standalone drawer device/
   command will be needed. Recommendation in this doc: keep drawer-via-fiscal
   as the default and revisit only if ATOL verification (§4a) shows the
   registrar can't drive the drawer Mercadia needs.
5. **MSR and iButton reader models/protocols** (owner: vendor docs). Still
   open per ADR-0002's own Open Points
   (`docs/adr/0002-terminal-app-and-hardware-agent.md:50-51`) and
   `docs/open-questions.md:69` ("не готов ответить" for iButton). Lower
   priority than fiscal/payment/scanner because these gate senior-cashier
   authentication factors, not the base sale flow.
6. **Scanner model and connection mode** (owner: vendor docs).
   `docs/open-questions.md:64` says "chaще всего COM, можно поддержать и USB"
   (usually COM, USB support possible) but names no specific model. This
   affects whether the real scanner driver is a COM/serial reader or a USB
   HID keyboard-wedge pass-through, which are architecturally different
   integrations.
7. **ATOL transport and driver generation** (owner: vendor docs). See §4a —
   nothing beyond "ATOL" is confirmed; USB vs TCP/IP and driver version must
   be verified against ATOL's current documentation before writing the ATOL
   adapter's implementation plan.
8. **Windows/Linux support matrix per device vendor** (owner: engineering +
   vendor docs). Still an open point on ADR-0002 itself
   (`docs/adr/0002-terminal-app-and-hardware-agent.md:49`); directly affects
   whether the Hardware Agent needs per-OS build tags once real drivers land,
   especially if ATOL's driver is Windows-only (COM/DLL) rather than a
   cross-platform TCP/IP gateway.

## Maintenance

This document gates the first real driver implementation plan (ATOL). When
that plan is written, §4 ("minimal command set") becomes its scope. Update
this matrix in the same PR whenever a `DeviceKind` or vendor decision changes
— treat it like code, not a one-time snapshot.
