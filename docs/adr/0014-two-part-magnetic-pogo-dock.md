---
status: proposed
date: 2026-08-15
---

# Two-part enclosure: bike plate + magnetic pogo HUD pod

The on-bike enclosure splits into a **bike plate** that stays on the motorcycle and a removable **HUD pod** (Pi Zero W + display + buttons). The pod docks with magnets; **power-only** 5 V passes through keyed magnetized pogo pins. The plate is a USB-C **sink** fed from the bike accessory port. Take the electronics off the bike; leave the mount and the USB-C inlet.

## Considered options

- **Chosen — two-part magnetic dock, 5 V pogo, Zero W** — theft/weather/bench use want the Pi off the bike; the plate can be bike-specific; magnets make dock/undock gloved and tool-free.
- **Chosen — USB-C from the accessory port into the plate** — the bike already supplies 5 V USB; the plate is a CC-wired sink, not a 12 V DC-DC.
- **Rejected: one-piece case bolted to the bars** — SD, USB, and overnight storage fight a permanently mounted stack; every bike change reprints the whole HUD.
- **Rejected: USB cable between plate and pod** — cables snag, wet connectors fail, and the “removable” story still needs a plug.
- **Rejected: 12 V SAE + buck in the plate** — extra heat, load-dump parts, and a second conversion when the accessory port is already USB.
- **Rejected: 12 V on the pogo** — a live motorcycle rail on exposed pins is a worse short than fused 5 V VBUS.
- **Rejected: Pi Zero 2 W** — USB-C on the Pi, higher current, different bench cutout; Zero W is enough for the hub and keeps the pod micro-USB.
- **Rejected: power + USB data on the same pogo** — extra pins and a gadget-mode conflict; BLE/Wi-Fi already carry the phone link. Revisit only if we need a docked debug channel.

## Consequences

- CAD splits: keep the current clamshell as the **pod**; add a **bike plate** and a shared underside **dock interface** ([`enclosure/DOCK.md`](../../enclosure/DOCK.md), [`enclosure/dock_interface.scad`](../../enclosure/dock_interface.scad)).
- Through-floor Pi screws in the bench case cannot poke the dock face — bosses stay blind or are covered.
- Bike electrical: USB-C inlet in the **rider-facing long edge** of the plate (UFP, 5.1 kΩ on CC), polyfuse on VBUS, recessed pogo on the mating face, dock-detect so pins are not hot when empty. No 12 V buck.
- Pod is **Pi Zero W**: docked power on GPIO 5 V/GND; bench power on micro-USB **PWR** (not the gadget port). Do not backfeed both without OR-ing.
- Magnets sit in the floor, away from a full steel sheet over the Zero W antenna.
