import { useContext } from 'react';

import { AuthContext } from './auth-context.js';
import type { AuthContextValue } from './auth-types.js';

export function useAuth(): AuthContextValue {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
}
