# Two-part dock (plan)

Product split recorded in [ADR 0014](../docs/adr/0014-two-part-magnetic-pogo-dock.md) (**proposed**). Shared millimetre contract: [`dock_interface.scad`](dock_interface.scad). The current [`moto_hud_case.scad`](moto_hud_case.scad) remains the bench clamshell until the pod underside is cut to this interface.

```
Bike 12 V ── fuse ──► bike plate ── buck → 5 V ── recessed pogo
                         │ magnets + tray fence (shear)
                         ▼
                    HUD pod (Pi + HAT + buttons)
                         USB still for bench, undocked only
```

## What lives where

| | **Bike plate** (stays on the motorcycle) | **HUD pod** (comes off) |
|---|---|---|
| Structure | Tray + mount pattern (AMPS / RAM diamond) | Today’s clamshell: floor, lid, button caps |
| Electronics | Fuse, 12 V→5 V buck, optional dock-detect MOSFET | Pi Zero + e-ink HAT + tactiles |
| Dock face | Spring pogo (live, recessed) + magnets | Target pads + opposite-polarity magnets |
| Cables | Strain-relieved 12 V lead into the plate | Short pigtail pad → Pi 5 V / GND |
| User | Gloved drop-in / pull-off | Screen, three buttons, SD still in the pod |

The plate is allowed to be bike-specific later (clamp vs stem vs ram). The **dock face** is not: one pod should sit on any plate that implements `dock_interface.scad`.

## Mechanical

### Tray, then magnets, then pogo

Magnets alone are a MagSafe-style face dock. On a motorcycle they lose to **shear** (braking, wind, vibration) before they lose to pull-off. The plate is a **shallow tray**: 2 mm fence, ~3 mm tall, around the pod footprint (~82 × 35 mm with current walls). Plate floor is **4 mm** so 3 mm magnet pockets do not break through to the bike side. The pod drops in; the fence takes XY; magnets take Z and keep the pogo compressed.

Pull-off target: about **10–20 N** by hand (gloved, one motion) and high enough that idle vibration does not walk the pod up the fence. Crash: prefer the pod to pop free of the plate rather than rip the bar clamp out — do not add a rigid latch in v0.

### Magnets

| | v0 default |
|---|---|
| Parts | 4× **6 mm × 3 mm** N52 discs (nickel) |
| XY | Not the geometric corners: **−X (SD) centres at 16 mm** so they miss the Pi M2.5 bosses (~5.9 mm); **+X centres at 7 mm** in the button rail. Y inset 7 mm. See `mag_xy` in the interface file. |
| Pockets | 6.2 mm ID × 3.2 mm deep, epoxy, **flush** with the face (no proud magnets to chip) |
| Pod Z | Bench `floor_t` is only 2 mm; 3 mm magnets need local wells into the 3 mm under-PCB air gap (keep ≤ ~1.2 mm proud of the inner floor, clear of through-hole solder) **or** a 4 mm floor on the docked pod. The interface **slab** is 4 mm so we can print a pull-test article without the Pi cavity. |
| Polarity | Checkerboard **N–S / S–N** so a 180° dock is repulsive (protects 5 V / GND) |
| Keepers | Magnets in **both** faces (magnet-to-magnet). No full steel plate — that would shroud the Pi Zero W antenna in the pod floor |
| Environment | Encapsulate (epoxy + paint or a 0.4 mm printed skin only if pull still meets target). Bare Ni-Cu-Ni rusts. NdFeB fades above ~80 °C — keep the plate off black sun-soaked plastic if we can |

Place pockets using `dock_interface.scad` (`mag_xy`, `mag_d`). Keep them clear of the Pi M2.5 bosses and of the pogo well.

### Alignment

Pogo typically wants **±0.5 mm**. Magnets help close Z; they do not clock XY well. Use:

1. Tray fence (coarse).
2. Keyed pogo housing or two **Ø2 mm plastic dowels** (fine), offset so the pod only seats one way.
3. Magnet polarity as the last foolproof against 180°.

Do not rely on a round magnetic connector that can spin.

### CAD change on the existing pod

Today’s base has **through-floor M2.5 holes**. A dock face cannot have screw tips or countersinks in the magnet/pogo plane.

- Pi bosses become **blind** (pilot from inside, no through-hole), or screws sit under a 0.8 mm cover sheet that *is* the dock face.
- USB cutout on +Y stays for bench.
- SD cutout on −X stays; swapping cards still means the pod is off the bike.

## Electrical

### Budget

Pi Zero W is typically **0.15–0.35 A** at 5 V with BLE/Wi-Fi and e-ink SPI; Zero 2 W can spike toward **0.6 A+**. Design the dock for **2 A** continuous, fuse the 12 V side at **1–2 A**.

Pi wants **4.75–5.25 V** at the board. Budget contact resistance + cable: the plate buck should be set to **5.1 V** at no-load and checked **under load at the Pi**, not at the module.

### Pinout (power only)

| Pin | Plate (spring) | Pod (pad) | Notes |
|-----|----------------|-----------|--------|
| 1 | 5 V (switched if dock-detect) | 5 V | Duplicate later if we need current sharing |
| 2 | GND | GND | |
| 3 | `DETECT` (pulled up in plate, 3.3–5 V logic) | GND via 0 Ω | Plate enables 5 V only when DETECT is low |

v0 can ship **2-pin** (always-on 5 V) for bench; **3-pin + MOSFET** is the on-bike default so an empty wet tray is not a live short.

Do **not** put motorcycle 12 V on the pogo.

### Connector style

“Magnetized pogo” means a **keyed** spring-pin + pad pair whose housing may include a ring magnet. Treat that ring magnet as **alignment**, not retention.

| Approach | When |
|----------|------|
| **Hybrid (chosen)** | Keyed 3-pin magnetic pogo **in the centre well** + the four corner discs + tray. Connector SKU can change; the well diameter is a parameter (`pogo_well_d`, default 14 mm). |
| Off-the-shelf circular only | Fine for a desk mock; too spinny / too weak on a bike by itself. |
| Custom Mill-Max / 2.54 mm pogo strip | If we cannot buy a keyed 3-pin that fits the 35 mm-deep pod. |

Spring pins live on the **plate** (replaceable, recessed below the fence). Pads live on the **pod** (flat, wipeable). Recess live pins ~1 mm below fence top.

### 12 V → 5 V in the plate

```
Battery / SAE ── inline mini-blade fuse ── reverse diode / TVS ── buck (8–36 V in → 5.1 V)
                                                                      │
                                                         DETECT MOSFET ── pogo 5 V
```

- Accessory tap: **SAE 2-pin (powerlet)** if the bike has one; otherwise a fused pigtail to the battery, not an unfused tap at the tail light.
- Prototype buck: a sealed 2–3 A module is enough. Road: wide-input, load-dump tolerant (ISO 7637-ish), or a pre-made USB-C PD pigtail **only if** we then regulate to a clean 5.1 V for the Pi — do not feed a PD trigger into GPIO 5 V and hope.
- Strain-relieve the 12 V lead in the plate (gland or printed clamp). No load on solder joints.

### Into the Pi

| Mode | Path |
|------|------|
| Docked | Pogo 5 V / GND → short pigtail → **GPIO physical 2 and 6** (or the 5 V / GND test pads). Pi Zero has **no polyfuse**. |
| Bench, undocked | Existing USB PWR (micro-USB on Zero W; USB-C on Zero 2 W). |
| Both | **Forbidden** unless the pod floor PCB has ideal-diode / MOSFET OR-ing. v0: unplug USB when docking. |

Do not feed 5 V into the USB gadget **data** port (the one next to HDMI on a Zero W) as if it were PWR.

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
- Printed caps remain actuators, not seals.

## BOM (v0, buy-to-fit)

| Qty | Item | Role |
|-----|------|------|
| 8 | 6×3 mm N52 discs | 4 plate + 4 pod |
| 1 pair | Keyed 3-pin magnetic pogo, ≥2 A, ~12–14 mm housing | Centre well; confirm footprint before cutting `pogo_well_d` |
| 1 | 8–36 V → 5 V 3 A buck | Inside plate |
| 1 | Mini-blade fuse holder + 2 A fuse | 12 V input |
| 1 | SAE 2-pin pigtail or fused battery lead | Bike tap |
| 1 | TVS + series diode (or a protected buck that includes them) | Polarity / spikes |
| — | 24–26 AWG silicone to Pi 5 V/GND | Pod pigtail |
| — | AMPS / RAM hardware | Plate to bike |

Caliper the real pogo housing and magnet OD before the first dock-face print. Same loop as `tactile_h` on the button rail.

## OpenSCAD work (next implementation pass)

Do **not** silently break the printable bench case. Sequence:

1. Keep `part = base\|lid\|caps\|assembly` behaviour.
2. Add `part = "plate"` (bike plate) and `part = "dock"` (underside preview) that `include` [`dock_interface.scad`](dock_interface.scad).
3. New pod floor: magnet pockets, pogo well, blind bosses, tray-compatible outer footprint (same `outer_w` / `outer_d` as today unless the fence needs a 0.3 mm shrink).
4. Export `plate.stl` + a docked `assembly` that shows plate + pod + explode.

Parameter ownership: **interface file** owns magnet XY, pogo XY, well diameter, fence height. Case file owns walls, buttons, window.

## Phases

1. **This plan** — lock split, 5 V pogo, tray+magnets, plate-side buck.
2. **Print the dock face only** — two slabs with pockets + well, no Pi cavity. Measure pull-off, shear on a fence, and whether polarity actually rejects 180°.
3. **Electrical bench** — fuse + buck + pogo into a Zero on the desk; load-test 5.1 V at the board; then add DETECT.
4. **Pod CAD** — blind bosses, pockets in the real floor, keep lid/buttons.
5. **Plate CAD** — tray, well, AMPS, cable clamp, electronics pocket.
6. **On-bike** — orientation, drain holes, sun/heat, real clamp geometry.

## Open questions (do not block CAD slabs)

- Where does 12 V actually come from on the first bike (existing SAE vs battery pigtail)?
- Zero W (micro-USB) or Zero 2 W (USB-C, more current, antenna still on-board)?
- Handlebar / stem / ram — measure after the dock face exists; AMPS is the stand-in.
- Is ~15 N pull-off the right “won’t launch, still removable with gloves” feel once we have a printed fence?
