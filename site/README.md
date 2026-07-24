# Moto HUD site

Astro static project site deployed to GitHub Pages at
**https://lexwebb.github.io/moto-hud/**.

## Local

```bash
cd site
npm install
npm run build:wasm   # needs Go on PATH — builds public/emulator/motohud.wasm
npm run dev          # http://localhost:4321/moto-hud/
```

Without wasm, home / enclosure / design still work; the emulator page shows an error until `build:wasm` (or CI) produces the binary.

```bash
npm run build        # syncs GLB + route JSON, then astro build → dist/
npm run preview      # preview the production build
```

## Pages deploy

Push to `main` runs [`.github/workflows/pages.yml`](../.github/workflows/pages.yml):
Go WASM build → copy assets → `astro build` → GitHub Pages.

One-time: repo **Settings → Pages → Source: GitHub Actions**.
