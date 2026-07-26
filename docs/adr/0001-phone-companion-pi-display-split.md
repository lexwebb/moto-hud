---
status: accepted
date: 2026-07-01
---

# Phone companion owns nav/media; Pi is display + buttons hub

The motorcycle HUD splits work: an Android (later iOS) **companion** reads navigation and media on the phone and pushes state over BLE; a Go **hub** on a Raspberry Pi Zero composites the 250×122 frame, drives the panel, and sends button `cmd`s back for media control. Audio stays on the phone → helmet Bluetooth.

## Considered options

- **Chosen** — thin Pi, smart phone (scrapers / nav engines already live there).
- **Rejected: all routing on Pi** — Pi Zero is a weak host for maps/search/reroute; phone already has GPS and app ecosystems.
- **Rejected: phone-only HUD** — e-ink / handlebar mount wants a dedicated low-power panel off the phone UI.

## Consequences

- Protocol and companion quality dominate nav richness; the Pi must not assume a single phone OS or engine.
- Dev HTTP injector can stand in for BLE while bringing up the hub.
