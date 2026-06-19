import { useEffect } from 'react';

import {
  registerUnauthorizedHandler,
  unregisterUnauthorizedHandler,
} from './unauthorized-handler.js';
import { useAuth } from './useAuth.js';

export function UnauthorizedBridge() {
  const { handleUnauthorized } = useAuth();

  useEffect(() => {
    registerUnauthorizedHandler(handleUnauthorized);
    return () => {
      unregisterUnauthorizedHandler();
    };
  }, [handleUnauthorized]);

  return null;
}
