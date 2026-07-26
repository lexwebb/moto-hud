# Minimap fixtures

JSON snapshots of turn-centered `nav.minimap` geometry for golden raster tests and the site minimap lab.

## Label tool

Open `/moto-hud/minimap-lab/` → **Label** tab.

- **Route** and **Context** are separate **vector** layers (octilinear H/V/45° snap — same as Go).
- Map is **heading-up** (approach ↑) to match the HUD frame.
- Click vertices → rubber-band snaps to 45°; double-click / Enter finishes a stroke.
- Preview bitmaps are rasterized with the same Bresenham tube pipeline as the Pi.
- **Accept & download** writes vectors + packed `route_bits` / `context_bits`.

Put accepted files in `site/public/emulator/minimap-labels/` when you want them in-repo.

## Iterate renderer

1. Change `pi/internal/hud/minimap.go`.
2. `npm run build:wasm` in `site/`.
3. Refresh lab Compare tab.
4. When locking raster output: `UPDATE_MINIMAP_GOLDEN=1 go test ./internal/hud/ -run MinimapGolden` from `pi/`.
5. After projection changes: `npm run capture:minimap` in `site/`.
