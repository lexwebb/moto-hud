# Moto HUD enclosure (bench prototype)

Parametric OpenSCAD clamshell for a **Raspberry Pi Zero + e-ink HAT** stack (Inky or Waveshare 2.13″). Buttons are a right-hand rail: printed mushroom caps over 6×6 mm tactiles by default.

On-bike shape (planned): this clamshell becomes the removable **HUD pod** (Pi **Zero W**); a separate **bike plate** stays on the motorcycle. USB-C from the bike accessory port feeds the plate; magnets + keyed pogo pass 5 V into the pod. Plan and millimetre contract: [`DOCK.md`](DOCK.md), [`dock_interface.scad`](dock_interface.scad), [ADR 0014](../docs/adr/0014-two-part-magnetic-pogo-dock.md). The printable bench case is unchanged until the pod underside is cut to that interface.

## Files

| Path | Purpose |
|------|---------|
| [`moto_hud_case.scad`](moto_hud_case.scad) | Parametric bench model (`part`: `assembly` / `base` / `lid` / `caps`) |
| [`DOCK.md`](DOCK.md) | Two-part magnetic pogo plan (bike plate + HUD pod) |
| [`dock_interface.scad`](dock_interface.scad) | Shared dock-face contract (magnets, tray fence, pogo well) |
| [`preview.html`](preview.html) | Standalone Three.js orbit viewer (dev convenience) |
| [`exports/`](exports/) | STL / GLB meshes for print + site |

The project site hosts the enclosure viewer at
**https://lexwebb.github.io/moto-hud/enclosure/** (copies `exports/assembly.glb` at build time).
After regenerating meshes, commit the updated GLB so CI/Pages pick it up.

## Buttons

| `button_mode` | Lid | Base | Extra print |
|---------------|-----|------|-------------|
| **`cap`** (default) | Stem shafts + outer flange counterbore | 6×6 tactile wells + wire channels into the board cavity | `part="caps"` — three ~11 mm mushroom actuators |
| **`panel8`** | 8.2 mm holes for high-head metal panel switches | No wells | None |

Order along +Y (rail on the **right**, screen upright): **Prev → Next → Action** (BCM 5 / 6 / 13). Wiring guide: [`docs/BUTTONS.md`](../docs/BUTTONS.md).

### Cap fit loop

1. Print base, lid, and caps. Dry-fit stems — they should slide, not bind (`stem_clear` / `print_tol`).
2. Seat 6×6 momentary NO tactiles in the wells; route pigtails through the wire channels.
3. Drop caps in from the outside (flange in the counterbore). Stem should click the tactile before bottoming on the lid.
4. If the click is early/late, change **`stem_len`** or **`well_depth`** only — do not change BCM pins.
5. Caliper the real tactile height into **`tactile_h`** after you buy parts.

Caps: TPU (~95A) for a softer dome, or PETG for a rigid v0. Export with stem toward the bed (`caps_sprue` lifts geometry so stem tips sit near Z=0) or flip in the slicer.

## Dimensions (defaults)

Tune the Customizer / parameter block after caliper checks:

- Pi Zero board: **65 × 30 mm**, M2.5 holes inset **3.5 mm**
- Stack height (Pi + header + HAT): **14 mm** (`stack_h`)
- Usable glass: **48.5 × 23.8 mm** (`window_w` / `window_d`)
- Wall / floor / lid: **2 mm**; lateral clearance **0.4 mm**
- Right-hand **button rail**: **12 mm** (`button_rail`); pitch **12 mm**; cap OD **11 mm**

## OpenSCAD

1. Install [OpenSCAD](https://openscad.org/) (2021.01+).
2. Open `moto_hud_case.scad`.
3. Set **part** to `assembly` for preview (optional **explode** to lift the lid).
4. **F5** preview, **F6** render.

### Export STLs (GUI)

1. Set `part` to `base`, `lid`, or `caps`, **F6**, then **File → Export → Export as STL**.
2. For the browser preview, export `part = "assembly"` to `exports/assembly.stl`, then convert to GLB (below).

### Export STLs (CLI)

From this directory (quoting matters on Windows `cmd`):

```bat
"C:\Program Files\OpenSCAD\openscad.exe" -D "part=\"base\"" -o exports/base.stl moto_hud_case.scad
"C:\Program Files\OpenSCAD\openscad.exe" -D "part=\"lid\"" -o exports/lid.stl moto_hud_case.scad
"C:\Program Files\OpenSCAD\openscad.exe" -D "part=\"caps\"" -o exports/caps.stl moto_hud_case.scad
"C:\Program Files\OpenSCAD\openscad.exe" -D "part=\"assembly\"" -o exports/assembly.stl moto_hud_case.scad
```

macOS / Linux:

```bash
openscad -D 'part="base"' -o exports/base.stl moto_hud_case.scad
openscad -D 'part="lid"' -o exports/lid.stl moto_hud_case.scad
openscad -D 'part="caps"' -o exports/caps.stl moto_hud_case.scad
openscad -D 'part="assembly"' -o exports/assembly.stl moto_hud_case.scad
```

Or run [`export_meshes.sh`](export_meshes.sh) / [`export_meshes.ps1`](export_meshes.ps1) (includes `caps` + GLB).

### Assembly GLB for the browser viewer

OpenSCAD does not export GLB; convert the assembly STL (needs `pip install trimesh`):

```bash
python -c "import trimesh; m=trimesh.load('exports/assembly.stl', force='mesh'); from trimesh.visual.material import PBRMaterial; from trimesh.visual import TextureVisuals; m.visual=TextureVisuals(material=PBRMaterial(baseColorFactor=[0.72,0.74,0.76,1], metallicFactor=0.05, roughnessFactor=0.55)); m.export('exports/assembly.glb')"
```

## Browser preview

Primary viewer: the Astro site page (`site/` → `/enclosure/` on Pages).

Standalone convenience viewer ([`preview.html`](preview.html)) still works:

```bash
# from repo root
python -m http.server 8765 --directory enclosure
# open http://127.0.0.1:8765/preview.html
```

After changing the `.scad`, re-export meshes (including `assembly.glb`) and refresh / redeploy the site.

## Print notes

- Print **base**, **lid**, and **caps** separately. Lid export is already flipped upright with lip up.
- Start with the default clearances; if the board is tight, bump `clearance` / `print_tol` by 0.1–0.2 mm.
- USB cutout sized for Pi **Zero W** micro-USB PWR (on-bike power is USB-C into the plate, not this hole).

## Next

On-bike dock (not in the bench STL yet): [`DOCK.md`](DOCK.md).

- Print **dock-face slabs** from `dock_interface.scad` and caliper pull-off / shear before cutting the real pod floor
- Weather sealing / lens / silicone skirt over caps
- Handlebar remote pod (large switches; same BCM map) — independent of the dock
- Bike-specific clamp body after measuring bars; plate uses AMPS / RAM in the meantime
