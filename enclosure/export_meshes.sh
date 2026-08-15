#!/usr/bin/env bash
# Regenerates enclosure STLs + assembly.glb for preview.html
set -euo pipefail
cd "$(dirname "$0")"
mkdir -p exports

OPENSCAD="${OPENSCAD:-openscad}"
if ! command -v "$OPENSCAD" >/dev/null 2>&1; then
  echo "OpenSCAD not found. Set OPENSCAD=... or install from https://openscad.org/" >&2
  exit 1
fi

for part in base lid caps assembly; do
  echo "Rendering $part..."
  "$OPENSCAD" -D "part=\"$part\"" -o "exports/${part}.stl" moto_hud_case.scad
done

echo "Rendering dock interface slabs..."
"$OPENSCAD" -D 'iface_part="assembly"' -o exports/dock_interface.stl dock_interface.scad
"$OPENSCAD" -D 'iface_part="plate"' -o exports/dock_plate_slab.stl dock_interface.scad
"$OPENSCAD" -D 'iface_part="pod"' -o exports/dock_pod_slab.stl dock_interface.scad
"$OPENSCAD" -D 'iface_part="plate_usb"' -o exports/dock_plate_underside.stl dock_interface.scad

echo "Converting assembly.stl -> assembly.glb..."
python3 - <<'PY'
import trimesh
m = trimesh.load("exports/assembly.stl", force="mesh")
mat = trimesh.visual.material.PBRMaterial(
    baseColorFactor=[0.72, 0.74, 0.76, 1.0],
    metallicFactor=0.05,
    roughnessFactor=0.55,
)
m.visual = trimesh.visual.TextureVisuals(material=mat)
m.export("exports/assembly.glb")
PY
echo "Done."
