/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_NAKAMA_HOST: string;
  readonly VITE_NAKAMA_PORT: string;
  readonly VITE_NAKAMA_SERVER_KEY: string;
  /** Set to "true" for https:// and wss:// (production behind TLS). */
  readonly VITE_NAKAMA_USE_SSL: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
