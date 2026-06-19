import { useEffect } from 'react';

import { registerUnauthorizedHandler } from './unauthorized-handler.js';
import { useAuth } from './useAuth.js';

export function UnauthorizedBridge() {
  const { handleUnauthorized } = useAuth();

  useEffect(() => {
    registerUnauthorizedHandler(handleUnauthorized);
  }, [handleUnauthorized]);

  return null;
}
