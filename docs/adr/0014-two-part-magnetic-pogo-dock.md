---
status: proposed
date: 2026-08-15
---

# Two-part enclosure: bike plate + magnetic pogo HUD pod

The on-bike enclosure splits into a **bike plate** that stays on the motorcycle and a removable **HUD pod** (Pi + display + buttons). The pod docks with magnets; **power-only** 5 V passes through keyed magnetized pogo pins. Take the electronics off the bike; leave the mount and fused supply.

## Considered options

- **Chosen — two-part magnetic dock, 5 V pogo** — theft/weather/bench use want the Pi off the bike; the plate can be bike-specific; magnets make dock/undock gloved and tool-free.
- **Rejected: one-piece case bolted to the bars** — SD, USB, and overnight storage fight a permanently mounted stack; every bike change reprints the whole HUD.
- **Rejected: USB cable between plate and pod** — cables snag, wet connectors fail, and the “removable” story still needs a plug.
- **Rejected: 12 V on the pogo** — a live motorcycle rail on exposed pins is a worse short/corrosion hazard than a fused 5 V well; the buck converter belongs in the plate.
- **Rejected: power + USB data on the same pogo** — extra pins and a gadget-mode conflict; BLE/Wi-Fi already carry the phone link. Revisit only if we need a docked debug channel.

## Consequences

- CAD splits: keep the current clamshell as the **pod**; add a **bike plate** and a shared underside **dock interface** ([`enclosure/DOCK.md`](../../enclosure/DOCK.md), [`enclosure/dock_interface.scad`](../../enclosure/dock_interface.scad)).
- Through-floor Pi screws in the bench case cannot poke the dock face — bosses stay blind or are covered.
- Bike electrical: fused 12 V in, automotive-ish buck to 5 V in the plate, recessed pogo, optional dock-detect so pins are not hot when empty.
- USB cutout on the pod stays for bench power when undocked; do not backfeed GPIO 5 V and USB at once without OR-ing.
- Magnets sit in the floor, away from a full steel sheet over the Pi Zero W antenna.
