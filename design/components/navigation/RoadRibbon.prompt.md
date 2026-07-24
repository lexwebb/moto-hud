Quiet 40px-tall band that schematizes the upcoming corridor as a bold kinked line — never a geographic map, never map tiles/streets. Oriented vertically: the spine runs bottom-to-top representing "ahead," with the line kinking sideways where a turn approaches — like a compressed forward view, not a left-right timeline.

```jsx
<RoadRibbon points={[{x:110,y:0},{x:110,y:24},{x:170,y:36}]} turnIndex={1} />
```

No `points` (or fewer than 2) renders a dashed flat line meaning "no corridor data yet" rather than an empty gap. Updates only on maneuver change or coarse distance thresholds, never a smooth scroll — the caller should only re-pass new points on those events.
