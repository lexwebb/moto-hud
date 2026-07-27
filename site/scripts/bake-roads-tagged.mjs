/**
 * Re-bake Whitehall→Farringdon OSM highways WITH tags for junction POC.
 * Writes web/emulator/routes/whitehall-farringdon-roads.json and copies to site/public.
 *
 * Usage (from site/):
 *   node scripts/bake-roads-tagged.mjs
 */
import { readFileSync, writeFileSync, copyFileSync, mkdirSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const siteRoot = join(__dirname, '..');
const repoRoot = join(siteRoot, '..');

const route = JSON.parse(
  readFileSync(join(siteRoot, 'public/emulator/routes/whitehall-farringdon.json'), 'utf8'),
);
const coords = route.coordinates;

let minLat = Infinity;
let maxLat = -Infinity;
let minLng = Infinity;
let maxLng = -Infinity;
for (const [lng, lat] of coords) {
  if (lat < minLat) minLat = lat;
  if (lat > maxLat) maxLat = lat;
  if (lng < minLng) minLng = lng;
  if (lng > maxLng) maxLng = lng;
}
const pad = 0.002; // ~220 m
const south = minLat - pad;
const west = minLng - pad;
const north = maxLat + pad;
const east = maxLng + pad;

const query = `
[out:json][timeout:90];
(
  way["highway"~"^(motorway|motorway_link|trunk|trunk_link|primary|primary_link|secondary|secondary_link|tertiary|tertiary_link|residential|unclassified|living_street|service)$"](${south},${west},${north},${east});
);
out body geom;
`.trim();

const endpoints = [
  'https://overpass-api.de/api/interpreter',
  'https://overpass.kumi.systems/api/interpreter',
];

async function fetchOverpass(url) {
  const res = await fetch(url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded',
      Accept: 'application/json',
      'User-Agent': 'moto-hud/0.1 (junction-poc bake; local dev)',
    },
    body: 'data=' + encodeURIComponent(query),
  });
  const text = await res.text();
  if (!res.ok) throw new Error(`${url} → ${res.status}: ${text.slice(0, 120)}`);
  if (text.trimStart().startsWith('<')) throw new Error(`${url} returned HTML/XML`);
  return JSON.parse(text);
}

function haversineM(a, b) {
  const R = 6371000;
  const toRad = (d) => (d * Math.PI) / 180;
  const dLat = toRad(b[1] - a[1]);
  const dLon = toRad(b[0] - a[0]);
  const lat1 = toRad(a[1]);
  const lat2 = toRad(b[1]);
  const h =
    Math.sin(dLat / 2) ** 2 + Math.cos(lat1) * Math.cos(lat2) * Math.sin(dLon / 2) ** 2;
  return 2 * R * Math.asin(Math.sqrt(h));
}

function nearRoute(geom, maxM = 180) {
  for (const [lng, lat] of geom) {
    for (const c of coords) {
      if (haversineM(c, [lng, lat]) <= maxM) return true;
    }
    // subsample: check every ~8th route point already covered by loop; ok for bake
  }
  // denser check against downsampled route
  for (let i = 0; i < coords.length; i += 3) {
    for (const [lng, lat] of geom) {
      if (haversineM(coords[i], [lng, lat]) <= maxM) return true;
    }
  }
  return false;
}

let data = null;
let used = null;
for (const url of endpoints) {
  try {
    console.log('Trying', url);
    data = await fetchOverpass(url);
    used = url;
    break;
  } catch (e) {
    console.warn(String(e.message || e));
  }
}
if (!data) {
  console.error('All Overpass endpoints failed');
  process.exit(1);
}

const ways = [];
let dualTagged = 0;
for (const el of data.elements || []) {
  if (el.type !== 'way' || !el.geometry?.length) continue;
  const geom = el.geometry.map((p) => [p.lon, p.lat]);
  if (geom.length < 2) continue;
  if (!nearRoute(geom, 180)) continue;
  const tags = el.tags || {};
  const dual = tags.dual_carriageway === 'yes';
  if (dual) dualTagged++;
  ways.push({
    id: el.id,
    coords: geom,
    highway: tags.highway || null,
    oneway: tags.oneway || null,
    dual_carriageway: dual,
    name: tags.name || null,
    ref: tags.ref || null,
  });
}

const payload = {
  name: 'Whitehall–Farringdon OSM highways',
  source: {
    overpass: true,
    highways: true,
    tagged: true,
    endpoint: used,
    baked: new Date().toISOString().slice(0, 10),
    filtered_near_route_m: 180,
    bbox: [south, west, north, east],
  },
  ways,
};

const webPath = join(repoRoot, 'web/emulator/routes/whitehall-farringdon-roads.json');
const sitePath = join(siteRoot, 'public/emulator/routes/whitehall-farringdon-roads.json');
mkdirSync(dirname(webPath), { recursive: true });
writeFileSync(webPath, JSON.stringify(payload));
copyFileSync(webPath, sitePath);

console.log(`Wrote ${ways.length} ways (${dualTagged} dual_carriageway=yes)`);
console.log(webPath);
console.log(sitePath);
