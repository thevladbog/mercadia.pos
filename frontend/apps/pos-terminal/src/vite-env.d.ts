/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_CENTRAL_BACKEND_URL?: string;
  readonly VITE_CENTRAL_SESSION_TOKEN?: string;
  readonly VITE_LAYOUT_TEMPLATE_ID?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
