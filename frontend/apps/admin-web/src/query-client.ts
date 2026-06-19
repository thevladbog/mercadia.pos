import { MutationCache, QueryCache, QueryClient } from '@tanstack/react-query';

import { isUnauthorizedError } from '@/auth/api-errors.js';
import { notifyUnauthorized } from '@/auth/unauthorized-handler.js';

function handleQueryError(error: unknown): void {
  if (isUnauthorizedError(error)) {
    notifyUnauthorized();
  }
}

export const queryClient = new QueryClient({
  queryCache: new QueryCache({
    onError: handleQueryError,
  }),
  mutationCache: new MutationCache({
    onError: handleQueryError,
  }),
  defaultOptions: {
    queries: {
      retry: false,
      refetchOnWindowFocus: false,
    },
  },
});
