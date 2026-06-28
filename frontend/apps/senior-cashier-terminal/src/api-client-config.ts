import { setApiBaseUrl } from '@mercadia/api-clients-store-edge';

export function configureStoreEdgeClient(): void {
  const baseUrl = import.meta.env.VITE_STORE_EDGE_URL;
  if (!baseUrl) {
    throw new Error('VITE_STORE_EDGE_URL is required');
  }
  setApiBaseUrl(baseUrl);
}

const DEFAULT_STORE_ID = import.meta.env.VITE_STORE_ID;

if (!DEFAULT_STORE_ID) {
  throw new Error('VITE_STORE_ID is required');
}

export function getStoreId(): string {
  return DEFAULT_STORE_ID;
}
