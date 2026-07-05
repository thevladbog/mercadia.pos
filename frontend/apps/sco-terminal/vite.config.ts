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
    port: 5176,
    proxy: {
      '^/v1/(auth|operational-days|receipts|shifts|terminals)(/.*)?$': {
        target: 'http://127.0.0.1:8081',
        changeOrigin: true,
      },
      '/v1/stores': {
        target: 'http://127.0.0.1:8081',
        changeOrigin: true,
      },
      '/v1/devices': {
        target: 'http://127.0.0.1:8083',
        changeOrigin: true,
      },
      '/v1/hardware': {
        target: 'http://127.0.0.1:8083',
        changeOrigin: true,
      },
    },
  },
});
