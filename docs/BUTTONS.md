# Button wiring

Inky pHAT uses SPI0 plus:

| Signal | BCM |
|--------|-----|
| BUSY | 17 |
| RESET | 27 |
| DC | 22 |

Do not reuse those for buttons.

## Default button map

| Function | BCM | Physical tip |
|----------|-----|--------------|
| Prev | 5 | Momentary NO to GND |
| Next | 6 | Momentary NO to GND |
| Action | 13 | Momentary NO to GND |

Software enables pull-ups; press = low.

| Gesture | Effect |
|---------|--------|
| Short Prev/Next | Change screen (or skip track on Media) |
| Long Prev/Next (≥800 ms) | Always change screen |
| Short Action | Context (play/pause or force redraw) |
| Long Action | Return to Nav home |
