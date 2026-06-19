import type { ListCentralUsers200UsersItem } from '@mercadia/api-clients-central';

export const CENTRAL_ROLE_ADMIN = 'central_admin' satisfies ListCentralUsers200UsersItem['roles'][number];
export const CENTRAL_ROLE_VIEWER = 'central_viewer' satisfies ListCentralUsers200UsersItem['roles'][number];

export const CENTRAL_ROLE_OPTIONS = [CENTRAL_ROLE_VIEWER, CENTRAL_ROLE_ADMIN] as const;

export function canManageCentralUsers(roles: string[]): boolean {
  return roles.includes(CENTRAL_ROLE_ADMIN);
}
