/**
 * Identifies which backend an API base URL is being resolved for. Each has a fixed
 * localhost port when running inside a packaged Tauri shell (see `docs/development/tauri.md`).
 */
export type ApiClientKind = 'central' | 'storeEdge' | 'hardwareAgent';

/**
 * Localhost defaults used only when running inside a packaged Tauri webview AND no explicit
 * `VITE_*_URL` env value was configured. A packaged app has no Vite dev proxy, so relative
 * paths (the browser-dev fallback) cannot resolve — these services always run on the same
 * machine as the terminal shell.
 */
const TAURI_LOCALHOST_DEFAULTS: Record<ApiClientKind, string> = {
  storeEdge: 'http://127.0.0.1:8081',
  central: 'http://127.0.0.1:8082',
  hardwareAgent: 'http://127.0.0.1:8083',
};

/**
 * Detects whether the app is running inside a Tauri webview.
 *
 * Tauri v2 injects `window.__TAURI_INTERNALS__` unconditionally before the frontend bundle
 * loads (verified against `@tauri-apps/api@2.11.1`'s `core.js`, which reads
 * `window.__TAURI_INTERNALS__` directly for `invoke`/`transformCallback`/`convertFileSrc`
 * without any feature-detection guard). Checking for that global avoids adding
 * `@tauri-apps/api` as a runtime dependency just for environment detection.
 */
export function isRunningInTauri(): boolean {
  return typeof window !== 'undefined' && '__TAURI_INTERNALS__' in window;
}

/**
 * Pure resolution function for an API client base URL.
 *
 * - An explicit, non-empty `envValue` always wins (works identically in browser and Tauri).
 * - Otherwise, in a browser (not Tauri), returns `''` — the existing behavior, which relies on
 *   the Vite dev proxy resolving same-origin relative paths.
 * - Otherwise, in Tauri, returns the fixed localhost default for `kind`.
 */
export function resolveApiBaseUrl(
  kind: ApiClientKind,
  options: { isTauri: boolean; envValue: string | undefined },
): string {
  const trimmed = options.envValue?.trim() ?? '';
  if (trimmed) {
    return trimmed;
  }
  if (options.isTauri) {
    return TAURI_LOCALHOST_DEFAULTS[kind];
  }
  return '';
}
