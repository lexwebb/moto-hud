# Button wiring

Canonical buttons are active-low to GND with software pull-ups. Display HATs reserve some GPIOs — do not reuse those for buttons.

## Default button map (Inky / Waveshare)

| Function | BCM | Physical tip |
|----------|-----|--------------|
| Prev | 5 | Momentary NO to GND |
| Next | 6 | Momentary NO to GND |
| Action | 13 | Momentary NO to GND |

| Gesture | Effect |
|---------|--------|
| Short Prev/Next | Change screen (or skip track on Media) |
| Long Prev/Next (≥800 ms) | Always change screen |
| Short Action | Context (play/pause or force redraw) |
| Long Action | Return to Nav home |

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

SPI0 CE0. Leave buttons on 5 / 6 / 13.

### Display HAT Mini (`-host lcd`)

| Signal | BCM |
|--------|-----|
| DC | 9 |
| Backlight | **13** |
| SPI | SPI0 CE1 |

Onboard buttons: A=5, B=6, X=16, Y=24. Soft Action remaps to **16** (X) because 13 is backlight. Prev/Next stay 5/6 (A/B).
