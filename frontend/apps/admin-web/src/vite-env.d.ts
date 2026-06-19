/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_CENTRAL_BACKEND_URL?: string;
  readonly VITE_STORE_EDGE_URL?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
