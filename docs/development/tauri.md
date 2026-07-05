# Tauri Desktop Shell

`frontend/apps/pos-terminal/` is the reference (template) Tauri v2 desktop shell for Mercadia
terminal apps, per ADR-0002 (`docs/adr/0002-terminal-app-and-hardware-agent.md`). The browser
dev workflow (`pnpm --filter pos-terminal dev`, proxied through Vite) is unchanged; Tauri wraps
the same app for packaged, kiosk-capable desktop distribution.

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

- `pnpm install` from `frontend/` (installs `@tauri-apps/cli`, a pinned devDependency of
  `pos-terminal`).

## Commands

Run all commands from `frontend/`:

| Purpose | Command |
| --- | --- |
| Dev (opens a native window against the Vite dev server) | `pnpm --filter pos-terminal tauri dev` |
| Rust compile check only | `cd apps/pos-terminal/src-tauri && cargo check` |
| Release build, no installer bundle | `pnpm --filter pos-terminal tauri build --no-bundle` |
| Release build with installer bundle(s) | `pnpm --filter pos-terminal tauri build` |
| Kiosk build (fullscreen, no window chrome, always-on-top) | `pnpm --filter pos-terminal tauri build --config src-tauri/tauri.kiosk.conf.json` |

`tauri dev` requires Store Edge (`:8081`) and Hardware Agent (`:8083`) running locally if you
want live data; otherwise the window opens but API calls fail, same as browser dev.

## URL resolution: dev vs packaged

`frontend/apps/pos-terminal/src/api-client-config.ts` exports a pure function,
`resolveApiBaseUrl(kind, { isTauri, envValue })`, used for all three API clients
(`central`, `storeEdge`, `hardwareAgent`):

1. If `envValue` (the corresponding `VITE_*_URL`) is set (non-empty after trimming), it always
   wins — this works identically in the browser and in Tauri.
2. Otherwise, in a **browser** (not Tauri), it returns `''`. Clients then issue relative-path
   requests, which the Vite dev proxy in `vite.config.ts` resolves to
   `http://127.0.0.1:8081` / `:8082`. This is the existing, unchanged behavior.
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

`.env.example` documents this: leaving `VITE_STORE_EDGE_URL` / `VITE_CENTRAL_BACKEND_URL` /
`VITE_HARDWARE_AGENT_URL` empty means "use the Vite proxy" in the browser and "use the
127.0.0.1 default" when packaged. Setting any of them explicitly overrides both.

## Config files

- `src-tauri/tauri.conf.json` — base config: `identifier: dev.mercadia.pos-terminal`,
  `build.devUrl` pointed at the app's Vite dev server (`http://localhost:5174`),
  `build.frontendDist: "../dist"`, `beforeDevCommand`/`beforeBuildCommand` delegate to the
  app's own `pnpm dev` / `pnpm build` scripts, and `app.security.csp` restricts
  `connect-src` to `'self'` plus `http://127.0.0.1:8081` (Store Edge),
  `http://127.0.0.1:8082` (Central Backend), `http://127.0.0.1:8083` (Hardware Agent), and
  `ws://127.0.0.1:*` (for any local websocket use). The window starts non-fullscreen
  (`fullscreen: false`).
- `src-tauri/tauri.kiosk.conf.json` — a config **overlay** applied with `--config` at build
  time. Tauri merges `--config` files using
  [JSON Merge Patch (RFC 7396)](https://v2.tauri.app/develop/configuration-files/): arrays are
  replaced wholesale, not deep-merged, so this file repeats the full `windows[0]` object (same
  `label`/`title`/`width`/`height` as the base config) with `fullscreen: true`,
  `decorations: false`, and `alwaysOnTop: true` added. Use it only for the packaged kiosk
  build; `tauri dev` should stay in the normal windowed base config for development ergonomics.
- `src-tauri/capabilities/default.json` — grants only `core:default` to the main window. No
  custom Tauri commands or plugin permissions are registered in this plan; the shell is a
  minimal wrapper around the existing web app (see ADR-0002's boundary: devices talk **only** to
  the Hardware Agent — the Tauri shell must not become a second device layer).
- `src-tauri/icons/` — generated with `pnpm tauri icon <path-to-square-png>` from
  `docs/Design/logo-pos-square.png`. Regenerate the same way if the logo changes.

## Replication checklist (next app)

To wrap another terminal app (`senior-cashier-terminal`, `sco-terminal`) in the same shell:

1. Copy `frontend/apps/pos-terminal/src-tauri/` into the target app, then:
   - Update `Cargo.toml` `package.name`/`[lib].name` to the new app (keep the `tauri`/
     `tauri-build` version pins in sync with this file).
   - Update `tauri.conf.json`: `productName`, `identifier`
     (`dev.mercadia.<app-name>`), `build.devUrl` (match the app's Vite `server.port`), window
     `title`.
   - Regenerate icons for the target app if it has its own logo; otherwise reuse
     `docs/Design/logo-pos-square.png`.
   - Update `tauri.kiosk.conf.json` `title` to match.
2. Add the same `devDependencies["@tauri-apps/cli"]` pin and `"tauri": "tauri"` script to the
   target app's `package.json`.
3. Apply the same `resolveApiBaseUrl`/`isRunningInTauri` pattern to the target app's API client
   configuration module, with that app's own set of `VITE_*_URL` env vars and Tauri localhost
   defaults (Store Edge `:8081` / Central `:8082` / Hardware Agent `:8083` are shared across all
   terminals — only the app-specific env var names differ).
4. Extend `.github/workflows/tauri.yml`'s path filter
   (`frontend/apps/*/src-tauri/**` already matches any app) and, if the new app should be
   built in CI immediately, add a parallel `cargo check` / `tauri build --no-bundle` step (or a
   matrix) for it.
5. Run `cargo check` and `pnpm --filter <app> tauri build --no-bundle` for the new app before
   opening a PR.

admin-web stays browser-only per ADR-0002's open point, resolved in
`docs/adr/technology-discussion.md` — flag this for maintainer confirmation in review if it
hasn't been formally closed yet.

## CI

`.github/workflows/tauri.yml` runs on `workflow_dispatch` and on pull requests touching
`frontend/apps/*/src-tauri/**`. It is a separate workflow from `.github/workflows/ci.yml`
(which stays Rust-free) and is **not** wired into `ci-gate`'s required checks — it installs
Linux WebKitGTK system dependencies, then runs `cargo check` and
`pnpm --filter pos-terminal tauri build --no-bundle --ci` (release build without bundling). Full
installer bundling (`tauri build` without `--no-bundle`) was verified locally on macOS
(produces a `.app` and a `.dmg`) but is intentionally not exercised in this Linux CI job, since
Linux bundle formats (`.deb`/`.rpm`/AppImage) need additional packaging tools this plan does not
provision; `cargo check` + `--no-bundle` is the accepted minimum per plan 013.
