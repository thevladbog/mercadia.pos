import { QueryClient, QueryClientProvider, QueryCache } from '@tanstack/react-query';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';
import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';

import { App } from './App.js';
import { configureCentralApiClient, isUnauthorizedError } from './auth/AuthProvider.js';
import './index.css';

configureCentralApiClient();

let unauthorizedHandler: (() => void) | null = null;

export function registerUnauthorizedHandler(handler: () => void): void {
  unauthorizedHandler = handler;
}

const queryClient = new QueryClient({
  queryCache: new QueryCache({
    onError: (error) => {
      if (isUnauthorizedError(error)) {
        unauthorizedHandler?.();
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
          unauthorizedHandler?.();
        }
      },
    },
  },
});

function Root() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <App />
      </BrowserRouter>
      <ReactQueryDevtools initialIsOpen={false} />
    </QueryClientProvider>
  );
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <Root />
  </StrictMode>,
);
