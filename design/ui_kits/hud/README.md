# HUD UI Kit

Interactive walkthrough of all six delivered screens at true 250×122 (scaled 2.4× for viewing). Buttons below the panel simulate the device's three physical controls: Prev / Next / Action, with mousedown+hold (~500ms) standing in for long-press.

Screens: NavActiveNoRibbon, NavActiveRibbon, NavIdle, MediaFocus, StatusDiagnostics, NavMediaHybrid — each a thin composition of the `components/` primitives (ManeuverGlyph, DistanceReadout, RoadRibbon, etc). No new visual logic lives here beyond layout.
