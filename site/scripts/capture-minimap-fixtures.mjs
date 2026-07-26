/**
 * Capture minimap fixtures from the Whitehall→Farringdon route for the lab + golden tests.
 *
 * Usage (from site/):
 *   node scripts/capture-minimap-fixtures.mjs
 */
import { readFileSync, writeFileSync, mkdirSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const root = join(__dirname, '..');
const repo = join(root, '..');
const geomPath = join(root, 'src/scripts/nav-geometry.js');
const {
  nextManeuver,
  turnMarkers,
  pathLength,
  pointAlong,
  haversineM,
  bearingBetween,
  minimapViewRadius,
} = await import(pathToFileURL(geomPath).href);

const route = JSON.parse(
  readFileSync(join(root, 'public/emulator/routes/whitehall-farringdon.json'), 'utf8'),
);
const roads = JSON.parse(
  readFileSync(join(root, 'public/emulator/routes/whitehall-farringdon-roads.json'), 'utf8'),
);
const ways = roads.ways || [];
const turns = turnMarkers(route);
const totalM = pathLength(route.coordinates);
const coords = route.coordinates;

function turnFrame(alongM, turnAt) {
  const turnIdx = Math.min(Math.max(turnAt, 0), coords.length - 1);
  const turn = coords[turnIdx];
  let fromIdx = Math.max(0, turnIdx - 1);
  for (let i = turnIdx - 1; i >= 0; i--) {
    if (haversineM(coords[i], turn) >= 12) {
      fromIdx = i;
      break;
    }
    fromIdx = i;
  }
  const inbound = bearingBetween(coords[fromIdx][0], coords[fromIdx][1], turn[0], turn[1]);
  const rider = pointAlong(coords, alongM);
  return {
    turn_lng: turn[0],
    turn_lat: turn[1],
    turn_at: turnIdx,
    inbound_bearing_deg: inbound,
    rider_lng: rider.lng,
    rider_lat: rider.lat,
  };
}

function nextTurnAt(alongM) {
  let traveled = 0;
  let segIndex = 0;
  for (; segIndex < coords.length - 1; segIndex++) {
    const seg = haversineM(coords[segIndex], coords[segIndex + 1]);
    if (traveled + seg >= alongM) break;
    traveled += seg;
  }
  const mans = route.maneuvers;
  let next = mans[mans.length - 1];
  for (let i = 0; i < mans.length; i++) {
    if (mans[i].at > segIndex) {
      next = mans[i];
      break;
    }
  }
  return next.at;
}

/** Presets: approach just before each named turn (unhappy cases first). */
const presets = [
  {
    id: 'northumberland-ave-near',
    title: 'Whitehall Place → Northumberland Avenue (near)',
    notes:
      'Unhappy junction from 2026-07-25 screenshot. Approach ~3 m from turn; route should read as a clean left without a box kink.',
    pick: (mans) => {
      const t = mans.find((m) => /northumberland/i.test(m.road));
      return t ? Math.max(0, t.alongM - 3) : 425;
    },
    acceptance: [
      'Single continuous route through turn mark (no rectangular loop)',
      'Exit direction matches left onto Northumberland Ave',
      'Context dashes are 1px and do not swallow the route',
      'Rider sits on approach below the turn',
    ],
  },
  {
    id: 'northumberland-ave-mid',
    title: 'Whitehall Place → Northumberland Avenue (~95 m)',
    notes: 'Earlier approach distance — useful for zoom/radius checks.',
    pick: (mans) => {
      const t = mans.find((m) => /northumberland/i.test(m.road));
      return t ? Math.max(0, t.alongM - 95) : 330;
    },
    acceptance: [
      'Route still fills the pane without looking tiny',
      'Turn stays centered',
    ],
  },
  {
    id: 'victoria-embankment-near',
    title: 'Northumberland Avenue → Victoria Embankment (near)',
    notes: 'Second major turn on the baked route.',
    pick: (mans) => {
      const t = mans.find((m) => /victoria embankment/i.test(m.road));
      return t ? Math.max(0, t.alongM - 5) : null;
    },
    acceptance: ['Clean bend; no spur back to origin'],
  },
];

const outPi = join(repo, 'pi/internal/hud/testdata/minimap');
const outSite = join(root, 'public/emulator/minimap-fixtures');
mkdirSync(outPi, { recursive: true });
mkdirSync(outSite, { recursive: true });

const index = [];
for (const p of presets) {
  const along = p.pick(turns);
  if (along == null || Number.isNaN(along)) {
    console.warn('skip', p.id, '(could not resolve along_m)');
    continue;
  }
  const nav = nextManeuver(route, along, ways);
  if (!nav.minimap) {
    console.warn('skip', p.id, '(no minimap at', along, 'm)');
    continue;
  }
  const frame = turnFrame(along, nextTurnAt(along));
  const fixture = {
    id: p.id,
    title: p.title,
    notes: p.notes,
    along_m: Math.round(along * 10) / 10,
    total_m: Math.round(totalM),
    acceptance: p.acceptance,
    ...frame,
    view_radius_m: minimapViewRadius(nav.minimap),
    minimap: nav.minimap,
    nav: {
      active: nav.active,
      maneuver: nav.maneuver,
      road: nav.road,
      distance_m: nav.distance_m,
      distance_text: nav.distance_text,
      eta_min: nav.eta_min,
      instruction: nav.instruction,
      minimap: nav.minimap,
    },
  };
  const body = `${JSON.stringify(fixture, null, 2)}\n`;
  writeFileSync(join(outPi, `${p.id}.json`), body);
  writeFileSync(join(outSite, `${p.id}.json`), body);
  index.push({
    id: p.id,
    title: p.title,
    along_m: fixture.along_m,
    file: `${p.id}.json`,
  });
  console.log('wrote', p.id, `@ ${fixture.along_m} m · ${nav.maneuver} · ${nav.road}`);
}

const idx = `${JSON.stringify({ route: route.name, fixtures: index }, null, 2)}\n`;
writeFileSync(join(outPi, 'index.json'), idx);
writeFileSync(join(outSite, 'index.json'), idx);
console.log(`captured ${index.length} fixtures → ${outPi}`);
