# Moto HUD enclosure (bench prototype)

Parametric OpenSCAD clamshell for a **Raspberry Pi Zero + Inky pHAT** stack. No bike mount or weather sealing yet — fit the boards first, then add mount geometry once you have on-bike measurements.

## Files

| Path | Purpose |
|------|---------|
| [`moto_hud_case.scad`](moto_hud_case.scad) | Parametric model (`part`: `assembly` / `base` / `lid`) |
| [`preview.html`](preview.html) | Browser orbit viewer ([model-viewer](https://modelviewer.dev/)) |
| [`exports/`](exports/) | STL meshes for print + preview |

## Dimensions (defaults)

Tune the Customizer / parameter block after caliper checks:

- Pi Zero board: **65 × 30 mm**, M2.5 holes inset **3.5 mm**
- Stack height (Pi + header + Inky): **14 mm** (`stack_h`)
- Inky usable glass: **48.5 × 23.8 mm** (`window_w` / `window_d`)
- Wall / floor / lid: **2 mm**; lateral clearance **0.4 mm**
- Right-hand **button rail**: **10 mm** (`button_rail`) — three holes along the side, clear of the glass and SD slot

## OpenSCAD

1. Install [OpenSCAD](https://openscad.org/) (2021.01+).
2. Open `moto_hud_case.scad`.
3. Set **part** to `assembly` for preview (optional **explode** to lift the lid).
4. **F5** preview, **F6** render.

### Export STLs (GUI)

1. Set `part` to `base` or `lid`, **F6**, then **File → Export → Export as STL**.
2. For the browser preview, export `part = "assembly"` to `exports/assembly.stl`, then convert to GLB (below).

### Export STLs (CLI)

From this directory (quoting matters on Windows `cmd`):

```bat
"C:\Program Files\OpenSCAD\openscad.exe" -D "part=\"base\"" -o exports/base.stl moto_hud_case.scad
"C:\Program Files\OpenSCAD\openscad.exe" -D "part=\"lid\"" -o exports/lid.stl moto_hud_case.scad
"C:\Program Files\OpenSCAD\openscad.exe" -D "part=\"assembly\"" -o exports/assembly.stl moto_hud_case.scad
```

macOS / Linux:

```bash
openscad -D 'part="base"' -o exports/base.stl moto_hud_case.scad
openscad -D 'part="lid"' -o exports/lid.stl moto_hud_case.scad
openscad -D 'part="assembly"' -o exports/assembly.stl moto_hud_case.scad
```

### Assembly GLB for the browser viewer

OpenSCAD does not export GLB; convert the assembly STL (needs `pip install trimesh`):

```bash
python -c "import trimesh; m=trimesh.load('exports/assembly.stl', force='mesh'); from trimesh.visual.material import PBRMaterial; from trimesh.visual import TextureVisuals; m.visual=TextureVisuals(material=PBRMaterial(baseColorFactor=[0.72,0.74,0.76,1], metallicFactor=0.05, roughnessFactor=0.55)); m.export('exports/assembly.glb')"
```

Or run [`export_meshes.ps1`](export_meshes.ps1) / [`export_meshes.sh`](export_meshes.sh).

## Browser preview

The preview is a Three.js orbit viewer (directional key light + ground shadows, solid background — no HDR skybox). Serve over HTTP:

```bash
# from repo root
python -m http.server 8765 --directory enclosure
# open http://127.0.0.1:8765/preview.html
```

After changing the `.scad`, re-export meshes and refresh the page.

## Print notes

- Print **base** and **lid** separately, flat on the bed (lid export is already flipped upright with lip up).
- Start with the default clearances; if the board is tight, bump `clearance` / `print_tol` by 0.1–0.2 mm.
- Buttons are **holes only** — pick switch diameter and set `button_d` / `button_pitch`.
- USB cutout sized for Pi Zero micro-USB power; widen `usb_w` / `usb_h` for Zero 2 W USB-C if needed.

## Next (not in this pass)

- Bike mount interface (after you measure the mount area)
- Weather sealing / lens
- Exact switch footprint and power scheme
