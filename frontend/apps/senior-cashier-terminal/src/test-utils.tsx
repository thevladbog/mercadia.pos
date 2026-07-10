import type { ReactElement } from 'react';
import { render } from '@testing-library/react';
import { I18nextProvider } from 'react-i18next';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ThemeProvider } from '@mercadia/ui';

import { IdleTimerProvider } from '@/auth/IdleTimerProvider.js';
import { i18n } from '@/i18n/config.js';

/**
 * Shared provider shell for component tests (plan 028 — the first
 * component-level tests in this app; every test file before this one only
 * covered pure `lib`/`*-data` helpers). Mirrors `Root.tsx`'s real provider
 * nesting (`ThemeProvider` → `QueryClientProvider` → router →
 * `IdleTimerProvider`) — minus `AuthProvider`, since callers mock
 * `@/auth/AuthProvider.js` directly instead of exercising the real
 * session/`createAuthSession` machinery it wraps.
 */
export function renderWithProviders(ui: ReactElement, options?: { route?: string }) {
  i18n.changeLanguage('en');
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });

  return render(
    <I18nextProvider i18n={i18n}>
      <ThemeProvider
        defaultTheme={{
          surface: 'senior-cashier',
          colorMode: 'dark',
          accentPreset: 'senior-cashier',
        }}
      >
        <QueryClientProvider client={queryClient}>
          <MemoryRouter initialEntries={[options?.route ?? '/']}>
            <IdleTimerProvider>{ui}</IdleTimerProvider>
          </MemoryRouter>
        </QueryClientProvider>
      </ThemeProvider>
    </I18nextProvider>,
  );
}
