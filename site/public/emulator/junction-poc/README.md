# Junction POC

Lab-only dual / junction experiments. Product still uses OsmAnd inference.

## Dual regression (tag GT)

```bash
cd site
node scripts/dual-regression.mjs
```

Writes `dual-regression-whitehall.json`: per-turn `expected_dual` (OSM `dual_carriageway` tags in the approach corridor), `predicted_dual` from `classifyJunction`, and match.

**Caveat:** tag GT ≠ physical truth. Embankment may be dual but untagged; Fleet Street must stay negative; Whitehall / Northumberland Avenue are positive via tags.
