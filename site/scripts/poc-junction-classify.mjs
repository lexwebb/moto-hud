/**
 * POC runner: classify every maneuver on Whitehall→Farringdon.
 *
 * Usage (from site/):
 *   node scripts/poc-junction-classify.mjs
 */
import { readFileSync, writeFileSync, mkdirSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const root = join(__dirname, '..');

const { turnMarkers, pathLength } = await import(
  pathToFileURL(join(root, 'src/scripts/nav-geometry.js')).href
);
const { classifyJunction } = await import(
  pathToFileURL(join(root, 'src/scripts/junction-classify.js')).href
);

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

const results = turns.map((t) => {
  const { junction, debug } = classifyJunction({
    coords,
    turnAt: t.at,
    ways,
    maneuver: t.maneuver,
    road: t.road,
    drive: 'left',
  });
  return {
    index: t.index,
    along_m: Math.round(t.alongM),
    label: t.label,
    junction,
    debug,
  };
});

const outDir = join(root, 'public/emulator/junction-poc');
mkdirSync(outDir, { recursive: true });
const outPath = join(outDir, 'whitehall-farringdon.json');
writeFileSync(
  outPath,
  JSON.stringify(
    {
      route: route.name,
      total_m: Math.round(totalM),
      drive: 'left',
      generated: new Date().toISOString().slice(0, 10),
      note: 'POC classifier — OSM dual_carriageway tags first; stricter opposite-heading geometry fallback.',
      maneuvers: results,
    },
    null,
    2,
  ),
);

console.log(`Route: ${route.name} (${Math.round(totalM)} m)\n`);
console.log(
  'idx'.padEnd(4) +
    'along'.padEnd(8) +
    'label'.padEnd(40) +
    'kind'.padEnd(18) +
    'out'.padEnd(14) +
    'dual'.padEnd(10) +
    'sides',
);
console.log('-'.repeat(110));
for (const r of results) {
  const j = r.junction;
  const sides = (j.sides || []).map((s) => `${s.side}@${s.at}`).join(',') || '—';
  const dual = r.debug.dual_source || '—';
  console.log(
    String(r.index).padEnd(4) +
      `${r.along_m}m`.padEnd(8) +
      r.label.slice(0, 38).padEnd(40) +
      j.kind.padEnd(18) +
      j.outbound.padEnd(14) +
      dual.padEnd(10) +
      sides,
  );
}

console.log(`\nWrote ${outPath}`);
console.log('\nFull IR per maneuver:\n');
for (const r of results) {
  console.log(`#${r.index} ${r.label}`);
  console.log(JSON.stringify(r.junction, null, 2));
  console.log(
    `  debug: turn=${r.debug.turn_deg}° geom=${r.debug.geom_outbound} dual=${r.debug.dual_source || 'no'} segs=${r.debug.context_segs}`,
  );
  console.log();
}
