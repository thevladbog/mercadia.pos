import { setApiBaseUrl, setSessionToken } from '@mercadia/api-clients-central';

export function configureCentralApiClient(): void {
  setApiBaseUrl(import.meta.env.VITE_CENTRAL_BACKEND_URL ?? '');
  const token = import.meta.env.VITE_CENTRAL_SESSION_TOKEN;
  if (token) {
    setSessionToken(token);
  }
}
