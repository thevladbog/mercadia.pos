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
    port: 5174,
    proxy: {
      '^/v1/(auth|catalog|operational-days|receipts|returns|shifts|store-edge|terminals)(/.*)?$': {
        target: 'http://127.0.0.1:8081',
        changeOrigin: true,
      },
      '^/v1/stores/[^/]+/(bank-|business-|cash-|catalog|monitoring|operation-journal|operational-days|returns|shifts|terminals)':
        {
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
