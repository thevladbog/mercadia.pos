import { QueryCache, QueryClient } from '@tanstack/react-query';

import { isUnauthorizedError } from './auth/api-errors.js';
import { notifyUnauthorized } from './auth/unauthorized-handler.js';

export const queryClient = new QueryClient({
  queryCache: new QueryCache({
    onError: (error) => {
      if (isUnauthorizedError(error)) {
        notifyUnauthorized();
      }
    },
  }),
  defaultOptions: {
    queries: {
      retry: false,
      refetchOnWindowFocus: false,
    },
    mutations: {
      onError: (error) => {
        if (isUnauthorizedError(error)) {
          notifyUnauthorized();
        }
      },
    },
  },
});
