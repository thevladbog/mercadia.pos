import { clearSessionToken, getSessionToken } from '@mercadia/api-clients-central';

import type { AuthState } from './auth-types.js';

const USER_ID_KEY = 'mercadia.central.userId';
const ROLES_KEY = 'mercadia.central.roles';

export function readStoredAuth(): Pick<AuthState, 'userId' | 'roles'> {
  const userId = sessionStorage.getItem(USER_ID_KEY);
  const rolesRaw = sessionStorage.getItem(ROLES_KEY);

  if (!userId || !getSessionToken()) {
    return { userId: null, roles: [] };
  }

  try {
    const roles = rolesRaw ? (JSON.parse(rolesRaw) as string[]) : [];
    return { userId, roles };
  } catch {
    return { userId: null, roles: [] };
  }
}

export function persistAuth(userId: string, roles: string[]): void {
  sessionStorage.setItem(USER_ID_KEY, userId);
  sessionStorage.setItem(ROLES_KEY, JSON.stringify(roles));
}

export function clearPersistedAuth(): void {
  sessionStorage.removeItem(USER_ID_KEY);
  sessionStorage.removeItem(ROLES_KEY);
  clearSessionToken();
}
