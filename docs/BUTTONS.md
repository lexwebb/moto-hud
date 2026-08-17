# Button wiring

Canonical buttons are active-low to GND with software pull-ups. Display HATs reserve some GPIOs — do not reuse those for buttons.

The Waveshare 2.13″ **PH2.0 8-pin cable is display-only** (VCC/GND/SPI/DC/RST/BUSY). It is **not** a button bus. Wire external momentary NO switches to free GPIOs instead.

## Default button map (Inky / Waveshare)

| Function | BCM | Physical pin (Pi Zero underside) | Suggested wire |
|----------|-----|----------------------------------|----------------|
| Prev | 5 | **29** | Green |
| Next | 6 | **31** | Yellow |
| Action | 13 | **33** | Blue |
| Common GND | — | **30** or **34** | Black |

| Gesture | Effect |
|---------|--------|
| Short Prev/Next | Change screen (or skip track on Media) |
| Long Prev/Next (≥800 ms) | Always change screen |
| Short Action | Context (play/pause or force redraw) |
| Long Action | Return to Nav home |

Enclosure rail order (+Y, rail on the right): **Prev → Next → Action**. See [`enclosure/README.md`](../enclosure/README.md).

## Switch BOM (matches enclosure SCAD)

| Tier | What | Notes |
|------|------|--------|
| **HUD lid (default)** | 6×6 mm **sealed / IP67** momentary NO tactile ×3 + printed caps (`part="caps"`) | Glove target is the ~11 mm print; electrical is the tactile |
| **Fallback** | 8 mm **high-head** metal momentary 1NO, IP65–IP67 | Set `button_mode = "panel8"` in OpenSCAD; no caps/wells |
| **Bar pod only (later)** | OTTO P9 / 12 mm high-head | Too big for the Pi Zero lid |
| **Avoid** | Flush open tactiles with no proud cap; latching (on/off) switches | Misses in gloves; wrong semantics |

Electrical: momentary NO → GND only. Soft pull-ups are enabled in software. Measure real tactile height and update `tactile_h` in the SCAD after purchase.

## Dummies soldering guide (Pi Zero + stacked HAT)

### What you need

- Fine tip soldering iron (~350 °C), thin solder, flux
- 4 lengths of flexible wire (~24–28 AWG, stranded), ~15–30 cm — colour-code them
- Multimeter (continuity); heat-shrink or Kapton
- Three 6×6 mm momentary NO tactiles (or 8 mm panel switches for `panel8`)

### Orient the board

1. Power off, unplug USB. Pads are on the **bottom** of the Zero (header solder side).
2. Find **pin 1**: the only **square** pad, nearest the **microSD** end.
3. With the SD end toward you and the header along the long edge facing you:

```
        SD card end ← toward you
   (odd row, toward board centre)     (even row, toward board edge)
        1  3  5  …  27  29  31  33  35  37  39
        2  4  6  …  28  30  32  34  36  38  40
```

Even pins sit along the **outer board edge**. Odds sit one row inward. Cross-check: [pinout.xyz](https://pinout.xyz/).

**Do not solder** Waveshare display pins: physical **19** (MOSI), **23** (SCLK), **24** (CE0), **22** (BCM25/DC), **11** (BCM17/RST), **18** (BCM24/BUSY).

### Solder steps

1. Strip ~2 mm; tin the wire tip.
2. Flux the pad. Iron to pad + wire; tiny solder — shiny cone, no bridge.
3. Solder **GND first** (30 or 34), then 29, 31, 33.
4. Loupe-check: no bridge to the neighbour (especially 31↔32).
5. Continuity (power still off): wire ↔ pad beep; wire ↔ neighbour **no** beep.
6. Heat-shrink / strain-relieve where the four wires leave the board.
7. Restack the HAT. Route wires through the enclosure **wire channels** into the tactile wells (or out a notch for bare-board tests).

### Connect the switches

Each tactile: terminal A → coloured signal; terminal B → shared black GND (daisy-chain). Ignore LED leads on illuminated parts for now.

When pressed, the GPIO is pulled to GND; `buttons.Start` treats low as pressed.

### Bench smoke test

```bash
./bin/motohud -host waveshare -http :8787
```

Expect logs `buttons: watching prev (GPIO5)` / `next (GPIO6)` / `action (GPIO13)`. Short Prev/Next cycles screens; long Action returns to Nav. Control without GPIO:

```bash
curl -X POST -d next http://127.0.0.1:8787/button
```

## Per-display reserved pins (BCM)

### Inky pHAT (`-host inky`)

| Signal | BCM |
|--------|-----|
| BUSY | 17 |
| RESET | 27 |
| DC | 22 |

SPI0 CE0. Leave buttons on 5 / 6 / 13.

### Waveshare 2.13″ B/W (`-host waveshare`)

| Signal | BCM |
|--------|-----|
| DC | 25 |
| RST | 17 |
| BUSY | 24 |

SPI0 CE0. Leave buttons on 5 / 6 / 13. The HAT’s 8-pin PH2.0 is the same SPI/control set for remote mounting — not buttons.

### Display HAT Mini (`-host lcd`)

| Signal | BCM |
|--------|-----|
| DC | 9 |
| Backlight | **13** |
| SPI | SPI0 CE1 |

Onboard buttons: A=5, B=6, X=16, Y=24. Soft Action remaps to **16** (X) because 13 is backlight. Prev/Next stay 5/6 (A/B).

## Waterproofing (later)

Printed caps are actuators only — not the electrical contact. Next steps: IP67 sealed tactiles in the same wells and/or a silicone skirt over the caps; base/lid gasket; optional remote handlebar pod with larger switches on the same BCM map.

On-bike power is **not** a USB cable out the side: the pod sits on a magnetic pogo **dock** ([`enclosure/DOCK.md`](../enclosure/DOCK.md)). Button BCM map does not change when docked.
