import { I18nextProvider } from 'react-i18next';
import { QueryClientProvider } from '@tanstack/react-query';
import { BrowserRouter } from 'react-router-dom';
import { ThemeProvider } from '@mercadia/ui';

import { AuthProvider } from '@/auth/AuthProvider.js';
import { i18n } from '@/i18n/config.js';
import { App } from '@/App.js';
import { queryClient } from '@/query-client.js';

export function Root() {
  return (
    <I18nextProvider i18n={i18n}>
      <ThemeProvider
        defaultTheme={{ surface: 'senior-cashier', colorMode: 'dark', accentPreset: 'neutral' }}
        persist
      >
        <QueryClientProvider client={queryClient}>
          <BrowserRouter>
            <AuthProvider>
              <App />
            </AuthProvider>
          </BrowserRouter>
        </QueryClientProvider>
      </ThemeProvider>
    </I18nextProvider>
  );
}
