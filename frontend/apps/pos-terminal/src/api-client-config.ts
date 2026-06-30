import {
  setApiBaseUrl as setCentralApiBaseUrl,
  setSessionToken as setCentralSessionToken,
} from '@mercadia/api-clients-central';
import { setApiBaseUrl as setHardwareAgentApiBaseUrl } from '@mercadia/api-clients-hardware-agent';
import {
  setApiBaseUrl as setStoreEdgeApiBaseUrl,
  setSessionToken as setStoreEdgeSessionToken,
} from '@mercadia/api-clients-store-edge';

export function configureApiClients(): void {
  setCentralApiBaseUrl(import.meta.env.VITE_CENTRAL_BACKEND_URL ?? '');
  setStoreEdgeApiBaseUrl(import.meta.env.VITE_STORE_EDGE_URL ?? '');
  setHardwareAgentApiBaseUrl(import.meta.env.VITE_HARDWARE_AGENT_URL ?? '');

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
