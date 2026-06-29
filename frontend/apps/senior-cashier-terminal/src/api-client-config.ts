import { setApiBaseUrl as setHardwareAgentApiBaseUrl } from '@mercadia/api-clients-hardware-agent';
import { setApiBaseUrl as setStoreEdgeApiBaseUrl } from '@mercadia/api-clients-store-edge';

export function configureApiClients(): void {
  const storeEdgeUrl = import.meta.env.VITE_STORE_EDGE_URL;
  if (!storeEdgeUrl) {
    throw new Error('VITE_STORE_EDGE_URL is required');
  }
  setStoreEdgeApiBaseUrl(storeEdgeUrl);
  setHardwareAgentApiBaseUrl(import.meta.env.VITE_HARDWARE_AGENT_URL ?? '');
}

const DEFAULT_STORE_ID = import.meta.env.VITE_STORE_ID;

if (!DEFAULT_STORE_ID) {
  throw new Error('VITE_STORE_ID is required');
}

export function getStoreId(): string {
  return DEFAULT_STORE_ID;
}
