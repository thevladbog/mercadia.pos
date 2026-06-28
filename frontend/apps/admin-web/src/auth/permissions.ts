import type { ListCentralUsers200UsersItem } from '@mercadia/api-clients-central';

export const CENTRAL_ROLE_ADMIN =
  'central_admin' satisfies ListCentralUsers200UsersItem['roles'][number];
export const CENTRAL_ROLE_VIEWER =
  'central_viewer' satisfies ListCentralUsers200UsersItem['roles'][number];
export const SENIOR_ROLE = 'senior_cashier' satisfies ListCentralUsers200UsersItem['roles'][number];

export const CENTRAL_ROLE_OPTIONS = [CENTRAL_ROLE_VIEWER, CENTRAL_ROLE_ADMIN, SENIOR_ROLE] as const;

export function canManageCentralUsers(roles: string[]): boolean {
  return roles.includes(CENTRAL_ROLE_ADMIN);
}

/** Admin-only cash operations (safe, EoD, admin pages). */
export function canWriteCashOperations(roles: string[]): boolean {
  return roles.includes(CENTRAL_ROLE_ADMIN);
}

/** Senior cashier pages or admin. */
export function canWriteStoreOperations(roles: string[]): boolean {
  return roles.includes(CENTRAL_ROLE_ADMIN) || roles.includes(SENIOR_ROLE);
}

export function isSeniorCashier(roles: string[]): boolean {
  return roles.includes(SENIOR_ROLE);
}
