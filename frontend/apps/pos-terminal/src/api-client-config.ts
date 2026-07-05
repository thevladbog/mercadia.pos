import {
  setApiBaseUrl as setCentralApiBaseUrl,
  setSessionToken as setCentralSessionToken,
} from '@mercadia/api-clients-central';
import { setApiBaseUrl as setHardwareAgentApiBaseUrl } from '@mercadia/api-clients-hardware-agent';
import {
  setApiBaseUrl as setStoreEdgeApiBaseUrl,
  setSessionToken as setStoreEdgeSessionToken,
} from '@mercadia/api-clients-store-edge';

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

export function configureApiClients(): void {
  const tauri = isRunningInTauri();

  setCentralApiBaseUrl(
    resolveApiBaseUrl('central', {
      isTauri: tauri,
      envValue: import.meta.env.VITE_CENTRAL_BACKEND_URL,
    }),
  );
  setStoreEdgeApiBaseUrl(
    resolveApiBaseUrl('storeEdge', {
      isTauri: tauri,
      envValue: import.meta.env.VITE_STORE_EDGE_URL,
    }),
  );
  setHardwareAgentApiBaseUrl(
    resolveApiBaseUrl('hardwareAgent', {
      isTauri: tauri,
      envValue: import.meta.env.VITE_HARDWARE_AGENT_URL,
    }),
  );

  const centralToken = import.meta.env.VITE_CENTRAL_SESSION_TOKEN;
  if (centralToken) {
    setCentralSessionToken(centralToken);
  }

  const storeEdgeToken = import.meta.env.VITE_STORE_EDGE_SESSION_TOKEN;
  if (storeEdgeToken) {
    setStoreEdgeSessionToken(storeEdgeToken);
  }
}

export function getStoreId(): string {
  const storeId = import.meta.env.VITE_POS_STORE_ID;
  if (!storeId) {
    throw new Error('VITE_POS_STORE_ID must be configured');
  }
  return storeId;
}

export function getTerminalId(): string {
  const terminalId = import.meta.env.VITE_POS_TERMINAL_ID;
  if (!terminalId) {
    throw new Error('VITE_POS_TERMINAL_ID must be configured');
  }
  return terminalId;
}
