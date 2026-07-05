# Tauri Desktop Shell

`frontend/apps/pos-terminal/` is the reference (template) Tauri v2 desktop shell for Mercadia
terminal apps, per ADR-0002 (`docs/adr/0002-terminal-app-and-hardware-agent.md`). It has been
replicated onto `frontend/apps/senior-cashier-terminal/` and `frontend/apps/sco-terminal/`, so
all three terminal apps now ship as Tauri v2 desktop shells; `admin-web` stays browser-only per
ADR-0002's open point. The browser dev workflow (`pnpm --filter <app> dev`, proxied through
Vite) is unchanged for every app; Tauri wraps the same app for packaged, kiosk-capable desktop
distribution.

## Why

A packaged web app has no Vite dev proxy, so it cannot reach the local Hardware Agent
(`127.0.0.1:8083` — credential readers: iButton/MSR/barcode) or Store Edge (`127.0.0.1:8081`)
using relative paths, and there is no kiosk/fullscreen capability. The Tauri shell fixes both:
it packages a native window around the built frontend, and the app resolves absolute
`http://127.0.0.1:<port>` base URLs for its API clients when it detects it is running inside
Tauri (see "URL resolution" below).

## Prerequisites

- Rust toolchain (`cargo`, `rustc`) — see the official
  [Tauri prerequisites guide](https://v2.tauri.app/start/prerequisites/) for your OS.
- Platform system dependencies per the same guide. On Debian/Ubuntu:

  ```bash
  sudo apt update
  sudo apt install libwebkit2gtk-4.1-dev \
    build-essential \
    curl \
    wget \
    file \
    libxdo-dev \
    libssl-dev \
    libayatana-appindicator3-dev \
    librsvg2-dev
  ```

- `pnpm install` from `frontend/` (installs `@tauri-apps/cli`, a pinned devDependency of each
  terminal app).

## Commands

Run all commands from `frontend/`, substituting the target app (`pos-terminal`,
`senior-cashier-terminal`, or `sco-terminal`):

| Purpose | Command |
| --- | --- |
| Dev (opens a native window against the Vite dev server) | `pnpm --filter <app> tauri dev` |
| Rust compile check only | `cd apps/<app>/src-tauri && cargo check` |
| Release build, no installer bundle | `pnpm --filter <app> tauri build --no-bundle` |
| Release build with installer bundle(s) | `pnpm --filter <app> tauri build` |
| Kiosk build (fullscreen, no window chrome, always-on-top) | `pnpm --filter <app> tauri build --config src-tauri/tauri.kiosk.conf.json` |

Each app's Vite dev server port: pos-terminal `5174`, senior-cashier-terminal `5175`,
sco-terminal `5176`.

`tauri dev` requires the app's backends running locally if you want live data (Store Edge
`:8081` for all three; Central Backend `:8082` for pos-terminal only; Hardware Agent `:8083` for
pos-terminal and senior-cashier-terminal); otherwise the window opens but API calls fail, same
as browser dev.

## URL resolution: dev vs packaged

`@mercadia/receipt-kit` (`frontend/packages/receipt-kit/src/api-base-url.ts`) exports a pure
function, `resolveApiBaseUrl(kind, { isTauri, envValue })`, and `isRunningInTauri()`, shared by
all three terminal apps' `src/api-client-config.ts`. Each app calls it once per API client it
uses (`central`/`storeEdge`/`hardwareAgent` for pos-terminal; `storeEdge`/`hardwareAgent` for
senior-cashier-terminal; `storeEdge` only for sco-terminal, which has no hardware-agent client
dependency yet):

1. If `envValue` (the corresponding `VITE_*_URL`) is set (non-empty after trimming), it always
   wins — this works identically in the browser and in Tauri.
2. Otherwise, in a **browser** (not Tauri), it returns `''`. Clients then issue relative-path
   requests, which each app's Vite dev proxy in `vite.config.ts` resolves to
   `http://127.0.0.1:8081` / `:8082` / `:8083`. This is the existing, unchanged behavior.
3. Otherwise, in a **packaged Tauri webview** (origin `tauri://localhost` /
   `http://tauri.localhost`), there is no dev proxy, so it returns a fixed localhost default:
   Store Edge `http://127.0.0.1:8081`, Central Backend `http://127.0.0.1:8082`, Hardware Agent
   `http://127.0.0.1:8083`.

Tauri detection (`isRunningInTauri()`) checks `'__TAURI_INTERNALS__' in window`. Tauri v2
injects `window.__TAURI_INTERNALS__` before the frontend bundle loads; it is used unconditionally
by `@tauri-apps/api`'s `core.js` (verified against the installed `@tauri-apps/api@2.11.1`
package) for `invoke`/`transformCallback`/`convertFileSrc`, so its presence is a reliable,
dependency-free signal. This repo does **not** depend on `@tauri-apps/api` — the check only reads
a global the Tauri runtime already sets.

Each app's `.env.example` documents this: leaving the relevant `VITE_*_URL` empty means "use the
Vite proxy" in the browser and "use the 127.0.0.1 default" when packaged. Setting it explicitly
overrides both. All three apps' Vite configs route `/v1/devices` and `/v1/hardware` to
`http://127.0.0.1:8083`, and the `@mercadia/api-clients-hardware-agent` mutator accepts an empty
base URL (relative-path mode) the same way the store-edge client does, so an empty
`VITE_HARDWARE_AGENT_URL` no longer throws — it correctly falls through to the proxy or the
Tauri localhost default.

## Config files

Each app's `src-tauri/` is structurally identical to make future kiosk hardening a mechanical
diff across all three:

- `src-tauri/tauri.conf.json` — base config: `identifier: dev.mercadia.<app-name>` (`pos-terminal`,
  `senior-cashier-terminal`, `sco-terminal`), `build.devUrl` pointed at the app's own Vite dev
  server, `build.frontendDist: "../dist"`, `beforeDevCommand`/`beforeBuildCommand` delegate to
  the app's own `pnpm dev` / `pnpm build` scripts, and `app.security.csp` restricts
  `connect-src` to `'self'` plus only the backends that app actually calls:
  - pos-terminal: `http://127.0.0.1:8081` (Store Edge), `http://127.0.0.1:8082` (Central
    Backend), `http://127.0.0.1:8083` (Hardware Agent).
  - senior-cashier-terminal: `http://127.0.0.1:8081` (Store Edge),
    `http://127.0.0.1:8083` (Hardware Agent) — no Central Backend dependency.
  - sco-terminal: `http://127.0.0.1:8081` (Store Edge), `http://127.0.0.1:8083` (Hardware
    Agent) — included ahead of the client dependency because scanner/hardware integration is
    the next SCO milestone.

  Every CSP also includes `ws://127.0.0.1:*` for any local websocket use. The window starts
  non-fullscreen (`fullscreen: false`).
- `src-tauri/tauri.kiosk.conf.json` — a config **overlay** applied with `--config` at build
  time. Tauri merges `--config` files using
  [JSON Merge Patch (RFC 7396)](https://v2.tauri.app/develop/configuration-files/): arrays are
  replaced wholesale, not deep-merged, so this file repeats the full `windows[0]` object (same
  `label`/`title`/`width`/`height` as the base config) with `fullscreen: true`,
  `decorations: false`, and `alwaysOnTop: true` added. Use it only for the packaged kiosk
  build; `tauri dev` should stay in the normal windowed base config for development ergonomics.
- `src-tauri/capabilities/default.json` — grants only `core:default` to the main window. No
  custom Tauri commands or plugin permissions are registered in any of the three shells; each is
  a minimal wrapper around the existing web app (see ADR-0002's boundary: devices talk **only**
  to the Hardware Agent — the Tauri shell must not become a second device layer).
- `src-tauri/icons/` — generated with `pnpm exec tauri icon <path-to-square-png> -o
  src-tauri/icons` from `docs/Design/logo-pos-square.png` (all three apps currently share this
  logo). The `ios/`, `android/`, and `64x64.png` outputs the command also generates are pruned —
  none of the three shells target mobile.

## Replication checklist (next app)

All three planned terminal apps (`pos-terminal`, `senior-cashier-terminal`, `sco-terminal`) now
have the shell. To wrap a future terminal app in the same pattern:

1. Copy `frontend/apps/pos-terminal/src-tauri/` into the target app, then:
   - Update `Cargo.toml` `package.name`/`[lib].name` to the new app (keep the `tauri`/
     `tauri-build` version pins in sync with this file).
   - Update `tauri.conf.json`: `productName`, `identifier`
     (`dev.mercadia.<app-name>`), `build.devUrl` (match the app's Vite `server.port`), window
     `title`, and `app.security.csp` (only the backends the app actually calls).
   - Regenerate icons for the target app if it has its own logo; otherwise reuse
     `docs/Design/logo-pos-square.png`.
   - Update `tauri.kiosk.conf.json` `title` to match.
2. Add the same `devDependencies["@tauri-apps/cli"]` pin and `"tauri": "tauri"` script to the
   target app's `package.json`.
3. Apply the same `resolveApiBaseUrl`/`isRunningInTauri` pattern (imported from
   `@mercadia/receipt-kit`) to the target app's API client configuration module, with that app's
   own set of `VITE_*_URL` env vars (Store Edge `:8081` / Central `:8082` / Hardware Agent
   `:8083` are shared across all terminals — only the app-specific env var names differ). If the
   app calls the Hardware Agent, also add `'/v1/devices'` and `'/v1/hardware'` proxy rules
   pointed at `http://127.0.0.1:8083` to its `vite.config.ts`.
4. Extend `.github/workflows/tauri.yml`'s path filter
   (`frontend/apps/*/src-tauri/**` already matches any app), add `cargo check` +
   `tauri build --no-bundle --ci` steps for the new app, and add its `src-tauri` directory to the
   `Swatinem/rust-cache` `workspaces` list.
5. Run `cargo check` and `pnpm --filter <app> tauri build --no-bundle` for the new app before
   opening a PR.

admin-web stays browser-only per ADR-0002's open point, resolved in
`docs/adr/technology-discussion.md` — flag this for maintainer confirmation in review if it
hasn't been formally closed yet.

## CI

`.github/workflows/tauri.yml` runs on `workflow_dispatch` and on pull requests touching
`frontend/apps/*/src-tauri/**`. It is a separate workflow from `.github/workflows/ci.yml`
(which stays Rust-free) and is **not** wired into `ci-gate`'s required checks — it installs
Linux WebKitGTK system dependencies once, then runs `cargo check` followed by
`tauri build --no-bundle --ci` (release build without bundling) sequentially for each of the
three shells (pos-terminal, senior-cashier-terminal, sco-terminal) in a single job, with
`Swatinem/rust-cache` scoped to all three `src-tauri` workspaces. Full installer bundling
(`tauri build` without `--no-bundle`) was verified locally on macOS for pos-terminal (produces a
`.app` and a `.dmg`) but is intentionally not exercised in this Linux CI job, since Linux bundle
formats (`.deb`/`.rpm`/AppImage) need additional packaging tools this plan does not provision;
`cargo check` + `--no-bundle` is the accepted minimum per plan 013.
