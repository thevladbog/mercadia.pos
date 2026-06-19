import path from 'node:path';
import { fileURLToPath } from 'node:url';

import react from '@vitejs/plugin-react';
import { defineConfig } from 'vite';

const appRoot = path.dirname(fileURLToPath(import.meta.url));

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(appRoot, 'src'),
    },
  },
  server: {
    port: 5173,
    proxy: {
      '^/v1/stores/[^/]+/(monitoring|terminals|cash-|bank-|business-|operation-journal|operational-days|shifts)':
        {
          target: 'http://127.0.0.1:8081',
          changeOrigin: true,
        },
      '/v1/operational-days': {
        target: 'http://127.0.0.1:8081',
        changeOrigin: true,
      },
      '/v1/shifts': {
        target: 'http://127.0.0.1:8081',
        changeOrigin: true,
      },
      '/v1/receipts': {
        target: 'http://127.0.0.1:8081',
        changeOrigin: true,
      },
      '/v1/terminals': {
        target: 'http://127.0.0.1:8081',
        changeOrigin: true,
      },
      '/v1': {
        target: 'http://127.0.0.1:8082',
        changeOrigin: true,
      },
    },
  },
});
