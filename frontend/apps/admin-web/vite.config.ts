import react from '@vitejs/plugin-react';
import { defineConfig } from 'vite';

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '^/v1/stores/[^/]+/monitoring': {
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
