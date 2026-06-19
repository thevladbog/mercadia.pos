import { useEffect } from 'react';

import { registerUnauthorizedHandler } from '../main.js';
import { useAuth } from './AuthProvider.js';

export function UnauthorizedBridge() {
  const { handleUnauthorized } = useAuth();

  useEffect(() => {
    registerUnauthorizedHandler(handleUnauthorized);
  }, [handleUnauthorized]);

  return null;
}
