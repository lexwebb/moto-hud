Bottom-edge button legend, 11px, right-aligned so each hint sits near its physical button, bold button name + plain action word.

```jsx
<FooterHints hints={[{btn:'PREV',label:'Mode'},{btn:'NEXT',label:'Skip'},{btn:'HOLD',label:'Home'}]} />
```

Keep to 2–3 entries — this row must never compete with the hero content above it. Omit entirely on screens where the legend would be redundant (e.g. once the user has learned it).
