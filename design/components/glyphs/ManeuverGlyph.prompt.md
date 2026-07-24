Hero pixel-grid arrow glyph for the current navigation maneuver — the first thing the eye lands on in a glance.

```jsx
<ManeuverGlyph type="left" size={40} />
```

Variants: left, right, slight-left, slight-right, straight, u-turn, roundabout, arrive, depart, unknown (renders a bold "?"). All strokes are 3px and square-capped — no rounded caps, no anti-aliasing tricks. Drawn as vector for preview; production raster pipeline thresholds it to 1-bit.
