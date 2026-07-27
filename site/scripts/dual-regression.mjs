/**
 * Dual-detection regression harness (tag GT vs classifyJunction).
 *
 * Labeling rule (expected_dual):
 *   Any OSM way with dual_carriageway=yes that contributes a segment into the
 *   approach corridor in turn-local frame — same box as detectDualFromTags in
 *   junction-classify.js: y ∈ [-45, 20], |x| ≤ 30 (meters; +Y = inbound).
 *   Tag GT ≠ physical truth (e.g. Embankment may be dual but untagged).
 *
 * Usage (from site/):
 *   node scripts/dual-regression.mjs
 *
 * Writes: public/emulator/junction-poc/dual-regression-whitehall.json
 */
import { readFileSync, writeFileSync, mkdirSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const root = join(__dirname, '..');

const {
  turnMarkers,
  pathLength,
  haversineM,
  bearingBetween,
  projectOne,
  wayCoords,
  TURN_SNAP_BEHIND,
  TURN_SNAP_AHEAD,
  TURN_SNAP_HALF_W,
} = await import(pathToFileURL(join(root, 'src/scripts/nav-geometry.js')).href);
const { classifyJunction } = await import(
  pathToFileURL(join(root, 'src/scripts/junction-classify.js')).href
);

/** Must stay in sync with detectDualFromTags corridor in junction-classify.js. */
const TAG_CORRIDOR = { yMin: -45, yMax: 20, xMax: 30 };

const KNOWN_NOTES = {
  'Whitehall': 'tagged dual approach — expect positive',
  'Whitehall Place': 'tagged dual — expect positive',
  'Northumberland Avenue': 'tagged dual — expect positive',
  'Victoria Embankment':
    'physically dual but often untagged; tag GT may be false or bleed from prior dual',
  'Temple Place': 'on Embankment corridor — physical dual may be untagged',
  'Fleet Street': 'must stay negative (parallel oneways, not dual_carriageway)',
};

function farPoint(coords, turnIdx, dir, minDistM) {
  const turn = coords[turnIdx];
  if (dir < 0) {
    for (let i = turnIdx - 1; i >= 0; i--) {
      if (haversineM(coords[i], turn) >= minDistM) return coords[i];
    }
    return coords[0];
  }
  for (let i = turnIdx + 1; i < coords.length; i++) {
    if (haversineM(coords[i], turn) >= minDistM) return coords[i];
  }
  return coords[coords.length - 1];
}

/** Tag-only expected dual for one turn (mirrors classifier tag path). */
function expectedDualFromTags(coords, turnAt, ways) {
  const turnIdx = Math.min(Math.max(turnAt, 0), coords.length - 1);
  const turn = coords[turnIdx];
  const approachPt = farPoint(coords, turnIdx, -1, 18);
  const inbound = bearingBetween(approachPt[0], approachPt[1], turn[0], turn[1]);
  const toLocal = (lng, lat) => projectOne(turn[1], turn[0], inbound, lng, lat);

  const hitNames = [];
  for (const way of ways || []) {
    if (!way?.dual_carriageway) continue;
    const geom = wayCoords(way);
    if (!geom || geom.length < 2) continue;
    for (let i = 0; i < geom.length - 1; i++) {
      const a = toLocal(geom[i][0], geom[i][1]);
      const b = toLocal(geom[i + 1][0], geom[i + 1][1]);
      const inBox = (p) =>
        p.y >= -TURN_SNAP_BEHIND &&
        p.y <= TURN_SNAP_AHEAD &&
        Math.abs(p.x) <= TURN_SNAP_HALF_W;
      if (!inBox(a) && !inBox(b)) continue;
      const mid = { x: (a.x + b.x) / 2, y: (a.y + b.y) / 2 };
      if (
        mid.y >= TAG_CORRIDOR.yMin &&
        mid.y <= TAG_CORRIDOR.yMax &&
        Math.abs(mid.x) <= TAG_CORRIDOR.xMax
      ) {
        const name = way.name || way.ref || String(way.id);
        if (!hitNames.includes(name)) hitNames.push(name);
      }
    }
  }
  return { expected: hitNames.length > 0, hitNames };
}

const route = JSON.parse(
  readFileSync(join(root, 'public/emulator/routes/whitehall-farringdon.json'), 'utf8'),
);
const roads = JSON.parse(
  readFileSync(join(root, 'public/emulator/routes/whitehall-farringdon-roads.json'), 'utf8'),
);

const ways = roads.ways || [];
const coords = route.coordinates;
const turns = turnMarkers(route);
const totalM = pathLength(coords);

const turnsOut = turns.map((t) => {
  const { expected, hitNames } = expectedDualFromTags(coords, t.at, ways);
  const { junction, debug } = classifyJunction({
    coords,
    turnAt: t.at,
    ways,
    maneuver: t.maneuver,
    road: t.road,
    drive: 'left',
  });
  const predicted = !!debug.dual_detected;
  const match = expected === predicted;
  const roadKey = Object.keys(KNOWN_NOTES).find((k) => (t.road || '').includes(k));
  const notes = [];
  if (roadKey) notes.push(KNOWN_NOTES[roadKey]);
  if (hitNames.length) notes.push(`tag hits: ${hitNames.join(', ')}`);
  if (!expected && predicted && debug.dual_source === 'geometry') {
    notes.push('geometry fallback positive vs tag-negative');
  }
  if (expected && !predicted) notes.push('classifier missed tag GT');

  return {
    index: t.index,
    along_m: Math.round(t.alongM),
    label: t.label,
    expected_dual: expected,
    predicted_dual: predicted,
    match,
    label_source: expected ? 'osm_dual_carriageway_tag' : 'no_tag_in_corridor',
    predicted_source: debug.dual_source || null,
    junction_kind: junction.kind,
    notes: notes.join('; ') || null,
  };
});

const pass = turnsOut.filter((r) => r.match).length;
const fail = turnsOut.length - pass;

const report = {
  route: route.name,
  total_m: Math.round(totalM),
  drive: 'left',
  generated: new Date().toISOString().slice(0, 10),
  labeling_rule:
    'expected_dual = any dual_carriageway=yes segment mid in turn-local corridor y∈[-45,20], |x|≤30 (same as classifyJunction tag path). Tag GT is incomplete vs physical duals.',
  caveat:
    'Lab-only bake. Product dual inference still uses OsmAnd. Embankment may be physically dual but untagged; Fleet Street must stay negative; Whitehall/Northumberland positive via tags.',
  summary: { total: turnsOut.length, pass, fail },
  turns: turnsOut,
};

const outDir = join(root, 'public/emulator/junction-poc');
mkdirSync(outDir, { recursive: true });
const outPath = join(outDir, 'dual-regression-whitehall.json');
writeFileSync(outPath, JSON.stringify(report, null, 2));

const readmePath = join(outDir, 'README.md');
writeFileSync(
  readmePath,
  `# Junction POC

Lab-only dual / junction experiments. Product still uses OsmAnd inference.

## Dual regression (tag GT)

\`\`\`bash
cd site
node scripts/dual-regression.mjs
\`\`\`

Writes \`dual-regression-whitehall.json\`: per-turn \`expected_dual\` (OSM \`dual_carriageway\` tags in the approach corridor), \`predicted_dual\` from \`classifyJunction\`, and match.

**Caveat:** tag GT ≠ physical truth. Embankment may be dual but untagged; Fleet Street must stay negative; Whitehall / Northumberland Avenue are positive via tags.
`,
);

console.log(`Route: ${route.name} (${Math.round(totalM)} m)`);
console.log(`Summary: ${pass} pass / ${fail} fail / ${turnsOut.length} total\n`);
console.log(
  'idx'.padEnd(4) +
    'exp'.padEnd(5) +
    'pred'.padEnd(6) +
    'ok'.padEnd(4) +
    'src'.padEnd(10) +
    'label',
);
console.log('-'.repeat(80));
for (const r of turnsOut) {
  console.log(
    String(r.index).padEnd(4) +
      (r.expected_dual ? 'Y' : 'n').padEnd(5) +
      (r.predicted_dual ? 'Y' : 'n').padEnd(6) +
      (r.match ? '✓' : '✗').padEnd(4) +
      (r.predicted_source || '—').padEnd(10) +
      r.label,
  );
}
console.log(`\nWrote ${outPath}`);
