# Two-part dock (plan)

Product split recorded in [ADR 0014](../docs/adr/0014-two-part-magnetic-pogo-dock.md) (**proposed**). Shared millimetre contract: [`dock_interface.scad`](dock_interface.scad). The current [`moto_hud_case.scad`](moto_hud_case.scad) remains the bench clamshell until the pod underside is cut to this interface.

```
Bike accessory USB ── USB-C ──► bike plate (sink + polyfuse) ── recessed pogo 5 V
                                      │ magnets + tray fence (shear)
                                      ▼
                                 HUD pod (Pi Zero W + HAT + buttons)
                                      micro-USB PWR for bench, undocked only
```

## What lives where

| | **Bike plate** (stays on the motorcycle) | **HUD pod** (comes off) |
|---|---|---|
| Structure | Tray + USB-C inlet + mount pattern (AMPS / RAM diamond) | Today’s clamshell: floor, lid, button caps |
| Electronics | USB-C sink (CC), VBUS polyfuse, optional dock-detect MOSFET | **Pi Zero W** + e-ink HAT + tactiles |
| Dock face | Spring pogo (live, recessed) + magnets | Target pads + opposite-polarity magnets |
| Cables | Accessory USB-C into the plate (strain-relieved jack) | Short pigtail pad → Pi GPIO 5 V / GND |
| User | Gloved drop-in / pull-off | Screen, three buttons, SD still in the pod |

The plate is allowed to be bike-specific later (clamp vs stem vs ram). The **dock face** is not: one pod should sit on any plate that implements `dock_interface.scad`.

## Mechanical

### U-caps, grip gaps, then magnets and pogo

Magnets alone are a MagSafe-style face dock. On a motorcycle they lose to **shear** before pull-off. The plate is a rounded plinth with **U-shaped end caps** that rise `wrap_h` (~9 mm) into matching **ridged insets** on the pod. The middle of each long side is open (~46 mm) so a glove can pinch the pod; a finger scoop undercuts the **+Y** (top) long edge. Caps take XY; magnets take Z and keep the pogo compressed.

Every outer edge and corner is a **2.5 mm rounded bevel** (fillet, not a 45° chamfer). Plate floor is **13 mm** so a USB-C jack fits in the **rider-facing (−Y) long edge** (bottom of the HUD as the rider sees it, buttons on the right) — the plug comes in from that vertical face, not up through the underside. Pogo stays on the top of that floor.

Pull-off target: about **10–20 N** by hand (gloved, one motion). Crash: prefer the pod to pop free of the plate rather than rip the bar clamp out — no rigid latch in v0.

### Magnets

| | v0 default |
|---|---|
| Parts | 4× **6 mm × 3 mm** N52 discs (nickel) |
| XY | Not the geometric corners: **−X (SD) centres at 16 mm** so they miss the Pi M2.5 bosses (~5.9 mm); **+X centres at 7 mm** in the button rail. Y inset 7 mm. See `mag_xy` in the interface file. |
| Pockets | 6.2 mm ID × 3.2 mm deep, epoxy, **flush** with the face (no proud magnets to chip) |
| Pod Z | Bench `floor_t` is only 2 mm; 3 mm magnets need local wells into the 3 mm under-PCB air gap **or** a thicker docked-pod floor. The interface preview pod is a 16 mm stand-in; the plate floor is 13 mm. |
| Polarity | Checkerboard **N–S / S–N** so a 180° dock is repulsive (protects 5 V / GND) |
| Keepers | Magnets in **both** faces (magnet-to-magnet). No full steel plate — that would shroud the Pi Zero W antenna in the pod floor |
| Environment | Encapsulate (epoxy + paint or a 0.4 mm printed skin only if pull still meets target). Bare Ni-Cu-Ni rusts. NdFeB fades above ~80 °C — keep the plate off black sun-soaked plastic if we can |

Place pockets using `dock_interface.scad` (`mag_xy`, `mag_d`). Keep them clear of the Pi M2.5 bosses and of the pogo well.

### Alignment

Pogo typically wants **±0.5 mm**. Magnets help close Z; they do not clock XY well. Use:

1. U-shaped end caps (coarse XY) + ridged insets on the pod.
2. Keyed pogo housing or two **Ø2 mm plastic dowels** (fine), offset so the pod only seats one way.
3. Magnet polarity as the last foolproof against 180°.

Do not rely on a round magnetic connector that can spin.

### CAD change on the existing pod

Today’s base has **through-floor M2.5 holes**. A dock face cannot have screw tips or countersinks in the magnet/pogo plane.

- Pi bosses become **blind** (pilot from inside, no through-hole), or screws sit under a 0.8 mm cover sheet that *is* the dock face.
- USB cutout on +Y stays for **bench** (Zero W micro-USB PWR). On-bike 5 V does not use that hole.
- SD cutout on −X stays; swapping cards still means the pod is off the bike.

## Electrical

Board is **Raspberry Pi Zero W** (not Zero 2 W). Accessory power is **USB-C 5 V into the plate**.

### Budget

Zero W is typically **0.15–0.35 A** at 5 V with BLE/Wi-Fi and e-ink SPI. Size the plate polyfuse at **1 A** (the Pi has no polyfuse of its own; accessory sockets are often 2–3 A). Pogo still **≥2 A** rated so contact resistance is not the weak point.

Pi wants **4.75–5.25 V** at the board. Check **under load at the Pi**, not at the accessory socket — USB-C cable + pogo + pigtail all drop voltage. Zero W current is low enough that this should stay in spec on a short cable.

### USB-C into the plate

The plate is a USB-C **UFP / sink**, not a charger and not PD.

```
Accessory port ── USB-C cable ──► plate receptacle
                                    CC1, CC2 → 5.1 kΩ to GND  (enable 5 V VBUS)
                                    VBUS → 1 A polyfuse → DETECT MOSFET → pogo 5 V
                                    D+/D− unconnected
```

- **Do not** fit a PD trigger that requests 9 / 12 / 20 V. Default USB-C 5 V is what Zero W needs.
- If the bike accessory is USB-A, use A-to-C; the plate inlet stays USB-C.
- Strain-relieve the jack in the **rider-facing long edge** of the plate (−Y, middle of the floor thickness; `usb_c_*` in `dock_interface.scad`). The plug comes in from below the screen, toward the rider — not up through the underside and not in a short-end wall. Rubber dust cap when the cable is out.
- No 12 V buck, SAE pigtail, or load-dump TVS in v0 — the accessory port already did that conversion. A USB VBUS ESD diode on the inlet is still worth fitting.

### Pinout (power only)

| Pin | Plate (spring) | Pod (pad) | Notes |
|-----|----------------|-----------|--------|
| 1 | 5 V (switched if dock-detect) | 5 V | From fused VBUS |
| 2 | GND | GND | |
| 3 | `DETECT` (pulled up in plate) | GND via 0 Ω | Plate enables 5 V only when DETECT is low |

v0 can ship **2-pin** (always-on 5 V) for bench; **3-pin + MOSFET** is the on-bike default so an empty wet tray is not a live short.

Do **not** put motorcycle 12 V on the pogo. Do **not** pass USB data through the pogo.

### Connector style

“Magnetized pogo” means a **keyed** spring-pin + pad pair whose housing may include a ring magnet. Treat that ring magnet as **alignment**, not retention.

| Approach | When |
|----------|------|
| **Hybrid (chosen)** | Keyed 3-pin magnetic pogo **in the centre well** + the four corner discs + tray. Connector SKU can change; the well diameter is a parameter (`pogo_well_d`, default 14 mm). |
| Off-the-shelf circular only | Fine for a desk mock; too spinny / too weak on a bike by itself. |
| Custom Mill-Max / 2.54 mm pogo strip | If we cannot buy a keyed 3-pin that fits the 35 mm-deep pod. |

Spring pins live on the **plate** (replaceable, recessed below the fence). Pads live on the **pod** (flat, wipeable). Recess live pins ~1 mm below fence top.

### Into the Pi (Zero W)

| Mode | Path |
|------|------|
| Docked | Pogo 5 V / GND → short pigtail → **GPIO physical 2 and 6** (or the 5 V / GND test pads). Zero W has **no polyfuse**. |
| Bench, undocked | micro-USB **PWR** on the Zero W (the port inward of HDMI). |
| Both | **Forbidden** unless the pod floor PCB has ideal-diode / MOSFET OR-ing. v0: unplug USB when docking. |

Do not feed 5 V into the USB gadget **data** port (the one next to HDMI) as if it were PWR. Do not widen the pod USB cutout for USB-C — that inlet lives on the plate.

### EMI / antenna

Pi Zero W chip antenna is a PCB zigzag on a board corner. Four 6 mm discs in the **printed floor** are acceptable; a steel dock plate the size of the pod is not. Keep magnet steel local to the pockets.

## Bike mount (plate back side)

Independent of the dock face so we can reprint clamps without touching the pod.

v0: **AMPS 4-hole** (30.4 × 38.6 mm) through the plate, plus a **RAM diamond** 2-hole slot (48.6 mm) if thickness allows. That covers most bar/stem balls. Do not bake a 22 mm clamp into the first print — measure the bike, then add a clamp body as a *second* SCAD part that bolts to AMPS.

## Weather (interface is the weak point)

The split is a rain gutter, not an IP67 plane.

- Drain the tray (two 3 mm holes at the low edge once we know on-bike orientation).
- Gasket optional around the pogo well only (closed-cell strip), not the whole fence — water should leave, not pool on live pins.
- DETECT + recessed pins are the electrical weather story for v0; conformal coat and lid gasket stay on the [buttons / sealing list](../docs/BUTTONS.md).
- USB-C inlet gets a rubber dust cap when the accessory cable is out.
- Printed caps remain actuators, not seals.

## BOM (v0, buy-to-fit)

| Qty | Item | Role |
|-----|------|------|
| 8 | 6×3 mm N52 discs | 4 plate + 4 pod |
| 1 pair | Keyed 3-pin magnetic pogo, ≥2 A, ~12–14 mm housing | Centre well; confirm footprint before cutting `pogo_well_d` |
| 1 | USB-C receptacle (panel / breakout) + dust cap | Plate inlet; power only |
| 2 | 5.1 kΩ 1% (CC1, CC2 to GND) | Advertise 5 V sink to the accessory port |
| 1 | ~1 A polyfuse (hold ≥0.5 A) | VBUS; Zero W has no onboard fuse |
| 1 | USB VBUS ESD / TVS (optional but cheap) | Inlet protection |
| 1 | DETECT MOSFET + pull-up (on-bike) | Dead tray when undocked |
| — | USB-C cable from accessory port (or A-to-C) | Bike → plate |
| — | 24–26 AWG silicone to Pi GPIO 5 V/GND | Pod pigtail |
| — | AMPS / RAM hardware | Plate to bike |

Caliper the real pogo housing and magnet OD before the first dock-face print. Same loop as `tactile_h` on the button rail.

## OpenSCAD work (next implementation pass)

Do **not** silently break the printable bench case. Sequence:

1. Keep `part = base\|lid\|caps\|assembly` behaviour.
2. Add `part = "plate"` (bike plate) and `part = "dock"` (underside preview) that `include` [`dock_interface.scad`](dock_interface.scad).
3. New pod floor: magnet pockets, pogo well, blind bosses, tray-compatible outer footprint (same `outer_w` / `outer_d` as today unless the fence needs a 0.3 mm shrink).
4. Export `plate.stl` + a docked `assembly` that shows plate + pod + explode.

Parameter ownership: **interface file** owns magnet XY, pogo XY, well diameter, U-cap wrap, grip gap, rounded bevel, and the rider-facing USB-C inlet. Case file owns walls, buttons, window.

## Phases

1. **This plan** — lock split, Zero W, USB-C accessory into the plate, 5 V pogo, tray+magnets.
2. **Print the dock face only** — two slabs with pockets + well, no Pi cavity. Measure pull-off, shear on a fence, and whether polarity actually rejects 180°.
3. **Electrical bench** — USB-C PSU → plate inlet (CC sink) → pogo into a Zero W; load-test 5 V at the board; then add DETECT.
4. **Pod CAD** — blind bosses, pockets in the real floor, keep lid/buttons.
5. **Plate CAD** — U-caps, rider-facing USB-C, AMPS, electronics pocket for polyfuse / MOSFET.
6. **On-bike** — orientation, drain holes, sun/heat, real clamp geometry.

## Open questions (do not block CAD slabs)

- Handlebar / stem / ram — measure after the dock face exists; AMPS is the stand-in.
- Is ~15 N pull-off the right “won’t launch, still removable with gloves” feel once we have a printed fence?
- Accessory socket USB-A vs USB-C on the bike itself — plate inlet stays USB-C either way.
