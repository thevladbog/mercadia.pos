import {
  setApiBaseUrl as setStoreEdgeApiBaseUrl,
  setSessionToken as setStoreEdgeSessionToken,
} from '@mercadia/api-clients-store-edge';
import { isRunningInTauri, resolveApiBaseUrl } from '@mercadia/receipt-kit';

export function configureApiClients(): void {
  setStoreEdgeApiBaseUrl(
    resolveApiBaseUrl('storeEdge', {
      isTauri: isRunningInTauri(),
      envValue: import.meta.env.VITE_STORE_EDGE_URL,
    }),
  );

  const storeEdgeToken = import.meta.env.VITE_STORE_EDGE_SESSION_TOKEN;
  if (storeEdgeToken) {
    setStoreEdgeSessionToken(storeEdgeToken);
  }
}

export function getStoreId(): string {
  const storeId = import.meta.env.VITE_SCO_STORE_ID;
  if (!storeId) {
    throw new Error('VITE_SCO_STORE_ID must be configured');
  }
  return storeId;
}

export function getTerminalId(): string {
  return import.meta.env.VITE_SCO_TERMINAL_ID ?? 'sco-1';
}

export function envValue(name: string, fallback: string): string {
  return (import.meta.env[name] as string | undefined) ?? fallback;
}
