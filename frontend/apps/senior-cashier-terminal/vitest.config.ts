import path from 'node:path';
import { fileURLToPath } from 'node:url';

import { defineConfig } from 'vitest/config';

const appRoot = path.dirname(fileURLToPath(import.meta.url));

export default defineConfig({
  // Mirrors `vite.config.ts`'s `@` alias — this config is separate from
  // that one (vitest doesn't merge them), and every prior test file here
  // only used relative imports, so the alias was never needed until plan
  // 028's component tests started importing via `@/...`.
  resolve: {
    alias: {
      '@': path.resolve(appRoot, 'src'),
    },
  },
  test: {
    environment: 'happy-dom',
    // `api-client-config.ts` throws at import time if `VITE_STORE_ID` is
    // unset (no real `.env` file is checked in, only `.env.example`) — any
    // component test that transitively imports it (e.g. `SafeRecountPage.tsx`
    // via `getStoreId()`) needs this defined for the module to load at all.
    env: {
      VITE_STORE_ID: 'store-1',
    },
  },
});
