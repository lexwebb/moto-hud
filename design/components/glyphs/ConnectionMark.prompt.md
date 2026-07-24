Compact BLE link/heartbeat status mark for the corner of every screen.

```jsx
<ConnectionMark connected={true} heartbeat={true} />
```

Bolt glyph = connected, X = disconnected. The trailing square dot blinks in discrete steps (never a smooth fade/pulse — matches the slow-refresh, no-smooth-animation rule) when a heartbeat is present.
