import { setApiBaseUrl } from '@mercadia/api-clients-central';
import { setApiBaseUrl as setStoreEdgeApiBaseUrl } from '@mercadia/api-clients-store-edge';

// Empty base URL keeps requests relative so Vite dev-server proxy routes /v1 traffic.
export function configureCentralApiClient(): void {
  setApiBaseUrl(import.meta.env.VITE_CENTRAL_BACKEND_URL ?? '');
}

export function configureStoreEdgeApiClient(): void {
  setStoreEdgeApiBaseUrl(import.meta.env.VITE_STORE_EDGE_URL ?? '');
}
