import type { CreateAuthSession201Session } from '@mercadia/api-clients-store-edge';

export type SessionResult = CreateAuthSession201Session;

export function canUsePosSession(session: SessionResult): boolean {
  return session.roles.some(
    (role) => role === 'cashier' || role === 'senior_cashier' || role === 'admin',
  );
}
