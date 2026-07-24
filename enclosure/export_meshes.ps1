# Regenerates enclosure STLs + assembly.glb for preview.html
# Requires: OpenSCAD on PATH or at the default Windows install path; Python + trimesh

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $Root

$openscad = @(
  "${env:ProgramFiles}\OpenSCAD\openscad.exe",
  "${env:ProgramFiles(x86)}\OpenSCAD\openscad.exe"
) | Where-Object { Test-Path $_ } | Select-Object -First 1

if (-not $openscad) {
  $cmd = Get-Command openscad -ErrorAction SilentlyContinue
  if ($cmd) { $openscad = $cmd.Source }
}

if (-not $openscad) {
  Write-Error "OpenSCAD not found. Install from https://openscad.org/ or winget install OpenSCAD.OpenSCAD"
}

New-Item -ItemType Directory -Force -Path "exports" | Out-Null

foreach ($part in @("base", "lid", "assembly")) {
  Write-Host "Rendering $part..."
  cmd /c "`"$openscad`" -D ""part=\""$part\"""" -o exports\$part.stl moto_hud_case.scad"
}

Write-Host "Converting assembly.stl -> assembly.glb..."
python -c @"
import trimesh
m = trimesh.load('exports/assembly.stl', force='mesh')
# Matte plastic so directional IBL / shadows read in the browser viewer
mat = trimesh.visual.material.PBRMaterial(
    baseColorFactor=[0.72, 0.74, 0.76, 1.0],
    metallicFactor=0.05,
    roughnessFactor=0.55,
)
m.visual = trimesh.visual.TextureVisuals(material=mat)
m.export('exports/assembly.glb')
"@
Write-Host "Done."
