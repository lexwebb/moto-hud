Discrete tick-block bar standing in for a progress-to-turn indicator.

```jsx
<ProgressTicks total={5} filled={3} />
```

Each tick is a small filled/empty square, matching the instrument-panel bold-stroke language. Only step the `filled` count at coarse distance thresholds (e.g. every 100m) — this is explicitly not a smooth-filling bar, since the panel can't refresh fast enough to animate one.
