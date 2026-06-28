import { setApiBaseUrl } from '@mercadia/api-clients-store-edge';

export function configureStoreEdgeClient(): void {
  setApiBaseUrl(import.meta.env.VITE_STORE_EDGE_URL ?? '');
}

const DEFAULT_STORE_ID = import.meta.env.VITE_STORE_ID ?? '';

export function getStoreId(): string {
  return DEFAULT_STORE_ID;
}
