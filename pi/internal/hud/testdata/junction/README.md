# Junction IR templates

Semantic (not meter) drawer that will replace `MinimapNodes` once
`PreferJunctionTemplates` is the live default.

## Hook points

| Path | Role |
|------|------|
| `pi/internal/hud/junction.go` | `JunctionNodes` templates + `PreferJunctionTemplates` flag |
| `pi/internal/hud/minimap.go` | Legacy meter polyline path (keep until flag flips) |
| `pi/internal/hudui/compose/nav_live.go` | Left column prefers junction when deps.HasJunction |
| `site/src/pages/junction-lab.astro` | Preview POC JSON (`emulator/junction-poc/*.json`) |
| `MotoHUD.junctionSVG` (WASM) | Optional SVG fragment from junction IR JSON |

## Kinds drawn (v1)

`simple`, `t_junction`, `crossroads`, `fork`, `merge`, `dual_carriageway`,
`roundabout`, `ramp_exit`, `ramp_enter`, `u_turn`, `arrive`, `depart`.

Site mirror: `site/src/scripts/junction-draw.js` (`IMPLEMENTED_KINDS`).

## Preview

```bash
# Go unit tests (no WASM rebuild required)
cd pi && go test ./internal/hud/ -run Junction -count=1

# Lab UI (Astro site)
cd site && npm run dev
# open /moto-hud/junction-lab/
```

Set `hud.PreferJunctionTemplates = true` in a local build to route live HUD
through IR templates (synthesizes from `maneuver` when `nav.junction` is absent).
