/**
 * Export a small OSM XML around Whitehall→Farringdon for OsmAnd OBF baking.
 * Usage: node scripts/export-whitehall-osm.mjs
 */
import { writeFileSync, mkdirSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const outDir = join(__dirname, '../public/emulator/osmand-poc');
mkdirSync(outDir, { recursive: true });

// Tight bbox covering the route (same pad as tagged roads bake).
const south = 51.502;
const west = -0.130;
const north = 51.522;
const east = -0.100;

const query = `
[out:xml][timeout:120];
(
  way["highway"](${south},${west},${north},${east});
  node(w);
);
out body;
`.trim();

const endpoints = [
  'https://overpass-api.de/api/interpreter',
  'https://overpass.kumi.systems/api/interpreter',
];

let xml = null;
for (const url of endpoints) {
  console.log('Trying', url);
  const res = await fetch(url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded',
      Accept: 'application/xml, text/xml, */*',
      'User-Agent': 'moto-hud/0.1 (osmand-poc; local)',
    },
    body: 'data=' + encodeURIComponent(query),
  });
  const text = await res.text();
  if (!res.ok) {
    console.warn(url, res.status, text.slice(0, 100));
    continue;
  }
  if (!text.includes('<osm')) {
    console.warn(url, 'not osm xml', text.slice(0, 100));
    continue;
  }
  xml = text;
  break;
}
if (!xml) {
  console.error('Overpass failed');
  process.exit(1);
}

const out = join(outDir, 'whitehall-farringdon.osm');
writeFileSync(out, xml);
console.log('Wrote', out, `(${(xml.length / 1e6).toFixed(2)} MB)`);
