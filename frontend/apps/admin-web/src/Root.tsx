import { ThemeProvider } from '@mercadia/ui';
import { QueryClientProvider } from '@tanstack/react-query';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';
import { I18nextProvider } from 'react-i18next';
import { BrowserRouter } from 'react-router-dom';

import { App } from '@/App.js';
import { i18n } from '@/i18n/index.js';
import { queryClient } from '@/query-client.js';

export function Root() {
  return (
    <ThemeProvider defaultTheme={{ surface: 'admin', colorMode: 'light', accentPreset: 'neutral' }}>
      <I18nextProvider i18n={i18n}>
        <QueryClientProvider client={queryClient}>
          <BrowserRouter>
            <App />
          </BrowserRouter>
          <ReactQueryDevtools initialIsOpen={false} />
        </QueryClientProvider>
      </I18nextProvider>
    </ThemeProvider>
  );
}
