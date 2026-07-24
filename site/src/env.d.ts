/// <reference types="astro/client" />

interface ImportMetaEnv {
  readonly BASE_URL: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}

interface Window {
  __MOTO_HUD_BASE__?: string;
  Go?: new () => {
    importObject: WebAssembly.Imports;
    run(instance: WebAssembly.Instance): void;
  };
  MotoHUD?: {
    applyNav(json: string): { ok?: boolean; error?: string };
    applyMedia(json: string): { ok?: boolean; error?: string };
    button(ev: string): { ok?: boolean; error?: string };
    renderPNG(): Uint8Array | { ok: false; error?: string };
    screen(): string;
  };
}
