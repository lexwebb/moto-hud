Optional small ETA line, 13px, sits below the hero distance.

```jsx
<ETAReadout etaMin={12} />
```

Renders nothing when `eta_min` wasn't in the packet — never show a placeholder dash for a field the phone didn't send.
