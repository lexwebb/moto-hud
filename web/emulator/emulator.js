const NATIVE_W = 250;
const NATIVE_H = 122;

function haversineM(a, b) {
  const R = 6371000;
  const toRad = (d) => (d * Math.PI) / 180;
  const dLat = toRad(b[1] - a[1]);
  const dLon = toRad(b[0] - a[0]);
  const lat1 = toRad(a[1]);
  const lat2 = toRad(b[1]);
  const h =
    Math.sin(dLat / 2) ** 2 +
    Math.cos(lat1) * Math.cos(lat2) * Math.sin(dLon / 2) ** 2;
  return 2 * R * Math.asin(Math.sqrt(h));
}

function pathLength(coords, from = 0, to = coords.length - 1) {
  let m = 0;
  for (let i = from; i < to; i++) m += haversineM(coords[i], coords[i + 1]);
  return m;
}

function pointAlong(coords, distM) {
  let left = distM;
  for (let i = 0; i < coords.length - 1; i++) {
    const seg = haversineM(coords[i], coords[i + 1]);
    if (left <= seg) {
      const t = seg === 0 ? 0 : left / seg;
      return {
        index: i,
        lng: coords[i][0] + (coords[i + 1][0] - coords[i][0]) * t,
        lat: coords[i][1] + (coords[i + 1][1] - coords[i][1]) * t,
      };
    }
    left -= seg;
  }
  const last = coords[coords.length - 1];
  return { index: coords.length - 2, lng: last[0], lat: last[1] };
}

function nextManeuver(route, alongM) {
  const coords = route.coordinates;
  let traveled = 0;
  let segIndex = 0;
  for (; segIndex < coords.length - 1; segIndex++) {
    const seg = haversineM(coords[segIndex], coords[segIndex + 1]);
    if (traveled + seg >= alongM) break;
    traveled += seg;
  }
  const mans = route.maneuvers;
  let cur = mans[0];
  let next = mans[mans.length - 1];
  for (let i = 0; i < mans.length; i++) {
    if (mans[i].at <= segIndex) cur = mans[i];
    if (mans[i].at > segIndex) {
      next = mans[i];
      break;
    }
  }
  const distToNext = pathLength(coords, segIndex, next.at);
  const remainOnSeg =
    haversineM(coords[segIndex], coords[Math.min(segIndex + 1, coords.length - 1)]) -
    (alongM - traveled);
  const distanceM = Math.max(0, Math.round(distToNext + Math.max(0, remainOnSeg)));
  const total = pathLength(coords);
  const done = Math.min(1, alongM / total);
  const etaMin = Math.max(1, Math.round(route.eta_min_start * (1 - done)));
  const arrived = next.maneuver === 'arrive' && distanceM < 15;
  return {
    active: !arrived,
    maneuver: arrived ? 'arrive' : next.maneuver,
    road: next.road,
    distance_m: distanceM,
    distance_text: distanceM >= 1000 ? `${(distanceM / 1000).toFixed(1)} km` : `${distanceM} m`,
    eta_min: etaMin,
    instruction: next.road,
  };
}

/** E-ink presentation: nearest 50 m + ≈ (U+2248, in Terminus). */
function formatEinkDistance(m) {
  const rounded = Math.max(0, Math.round(m / 50) * 50);
  if (rounded >= 1000) {
    const km = (rounded / 1000).toFixed(1);
    return { distance_m: rounded, distance_text: `≈ ${km} km` };
  }
  return { distance_m: rounded, distance_text: `≈ ${rounded} m` };
}

function forPanel(nav, eink) {
  if (!eink) return nav;
  const d = formatEinkDistance(nav.distance_m);
  return { ...nav, ...d };
}

/** @typedef {{ applyNav(j:string):any, applyMedia(j:string):any, button(e:string):any, renderPNG():Uint8Array, screen():string }} HudBackend */

  async function createWasmBackend() {
  if (typeof Go !== 'function') throw new Error('wasm_exec.js not loaded');
  const res = await fetch('motohud.wasm');
  if (!res.ok) throw new Error(`wasm ${res.status}`);
  const go = new Go();
  const result = await WebAssembly.instantiateStreaming(res, go.importObject);
  go.run(result.instance);
  for (let i = 0; i < 200; i++) {
    if (globalThis.MotoHUD?.renderPNG) break;
    await new Promise((r) => setTimeout(r, 10));
  }
  if (!globalThis.MotoHUD?.renderPNG) throw new Error('MotoHUD exports missing');
  const api = globalThis.MotoHUD;
  return {
    name: 'wasm',
    async applyNav(msg) {
      const r = api.applyNav(JSON.stringify({ type: 'nav', ...msg }));
      if (r && r.ok === false) throw new Error(r.error || 'applyNav failed');
    },
    async applyMedia(msg) {
      const r = api.applyMedia(JSON.stringify({ type: 'media', ...msg }));
      if (r && r.ok === false) throw new Error(r.error || 'applyMedia failed');
    },
    async button(ev) {
      const r = api.button(ev);
      if (r && r.ok === false) throw new Error(r.error || 'button failed');
    },
    async renderPNG() {
      const bytes = api.renderPNG();
      if (bytes && bytes.ok === false) throw new Error(bytes.error || 'render failed');
      return bytes;
    },
    async screen() {
      return api.screen();
    },
  };
}

function createHttpBackend() {
  return {
    name: 'http',
    async applyNav(msg) {
      await fetch('/nav', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ type: 'nav', ...msg }),
      });
    },
    async applyMedia(msg) {
      await fetch('/media', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ type: 'media', ...msg }),
      });
    },
    async button(ev) {
      await fetch('/button', { method: 'POST', body: ev });
    },
    async renderPNG() {
      const res = await fetch(`/frame.png?${Date.now()}`);
      if (!res.ok) throw new Error(`frame.png ${res.status}`);
      return new Uint8Array(await res.arrayBuffer());
    },
    async screen() {
      return '—';
    },
  };
}

async function paintHudInstant(canvas, pngBytes) {
  const bitmap = await createImageBitmap(new Blob([pngBytes], { type: 'image/png' }));
  canvas.width = NATIVE_W;
  canvas.height = NATIVE_H;
  const ctx = canvas.getContext('2d', { alpha: false });
  ctx.imageSmoothingEnabled = false;
  ctx.fillStyle = '#fff';
  ctx.fillRect(0, 0, NATIVE_W, NATIVE_H);
  ctx.drawImage(bitmap, 0, 0);
  bitmap.close();
}

function sleep(ms) {
  return new Promise((r) => setTimeout(r, ms));
}

/** Mirrors pi/internal/hud.RefreshGate — redraw about every 50 m. */
const DISTANCE_STEP_M = 50;

function bucketForDistance(m) {
  if (m <= 0) return 0;
  return Math.round(m / DISTANCE_STEP_M) * DISTANCE_STEP_M;
}

class RefreshGate {
  constructor() {
    this.hasLast = false;
  }

  wouldRedraw(screen, nav, force) {
    if (force || !this.hasLast) return true;
    if (screen !== this.lastScreen) return true;
    if (screen !== 'NAV') return false;
    if (
      nav.active !== this.lastActive ||
      nav.maneuver !== this.lastManeuver ||
      nav.road !== this.lastRoad
    ) {
      return true;
    }
    return bucketForDistance(nav.distance_m) !== this.lastBucket;
  }

  shouldRedraw(screen, nav, force) {
    if (!this.wouldRedraw(screen, nav, force)) return false;
    this.remember(screen, nav);
    return true;
  }

  remember(screen, nav) {
    this.lastScreen = screen;
    this.lastManeuver = nav.maneuver;
    this.lastRoad = nav.road;
    this.lastActive = nav.active;
    this.lastBucket = bucketForDistance(nav.distance_m);
    this.hasLast = true;
  }

  reset() {
    this.hasLast = false;
  }
}

/**
 * Inky-style full refresh: all pixels on → all off → fade in new frame (~1s).
 * When disabled, paints instantly every call.
 */
class EinkPanel {
  constructor(canvas, opts = {}) {
    this.canvas = canvas;
    this.screenEl = canvas.closest('.screen');
    this.enabled = opts.enabled !== false;
    this.flashMs = opts.flashMs ?? 120;
    this.fadeMs = opts.fadeMs ?? 1000;
    this.gate = new RefreshGate();
    this.busy = false;
    this.pending = null;
    this.lastRefreshAt = 0;
    this.refreshCount = 0;
  }

  get refreshMs() {
    return this.flashMs * 2 + this.fadeMs;
  }

  setEnabled(on) {
    this.enabled = on;
    if (!on) this.gate.reset();
  }

  async show(pngBytes, { screen = 'nav', nav = {}, force = false } = {}) {
    if (!this.enabled) {
      await paintHudInstant(this.canvas, pngBytes);
      this.lastRefreshAt = performance.now();
      this.refreshCount += 1;
      return { refreshed: true, wiped: false };
    }

    if (!this.gate.shouldRedraw(screen, nav, force)) {
      return { refreshed: false, wiped: false };
    }

    if (this.busy) {
      this.pending = { pngBytes, screen, nav, force: true };
      return { refreshed: false, wiped: false, queued: true };
    }

    this.busy = true;
    return this.#runRefresh(pngBytes);
  }

  async #runRefresh(pngBytes) {
    this.screenEl?.classList.add('wiping');
    const ctx = this.canvas.getContext('2d', { alpha: false });
    this.canvas.width = NATIVE_W;
    this.canvas.height = NATIVE_H;
    ctx.imageSmoothingEnabled = false;

    // Drive waveform: all pixels on, then all off (clear).
    ctx.fillStyle = '#000';
    ctx.fillRect(0, 0, NATIVE_W, NATIVE_H);
    await sleep(this.flashMs);
    ctx.fillStyle = '#fff';
    ctx.fillRect(0, 0, NATIVE_W, NATIVE_H);
    await sleep(this.flashMs);

    // Fade new frame in over ~1s (pigment settling).
    const bitmap = await createImageBitmap(new Blob([pngBytes], { type: 'image/png' }));
    const t0 = performance.now();
    await new Promise((resolve) => {
      const step = (now) => {
        const t = Math.min(1, (now - t0) / this.fadeMs);
        // Ease-out: fast start, settle at the end like e-ink.
        const a = 1 - (1 - t) ** 2;
        ctx.globalAlpha = 1;
        ctx.fillStyle = '#fff';
        ctx.fillRect(0, 0, NATIVE_W, NATIVE_H);
        ctx.globalAlpha = a;
        ctx.drawImage(bitmap, 0, 0);
        ctx.globalAlpha = 1;
        if (t < 1) {
          requestAnimationFrame(step);
        } else {
          resolve();
        }
      };
      requestAnimationFrame(step);
    });
    ctx.fillStyle = '#fff';
    ctx.fillRect(0, 0, NATIVE_W, NATIVE_H);
    ctx.drawImage(bitmap, 0, 0);
    bitmap.close();

    this.screenEl?.classList.remove('wiping');
    this.busy = false;
    this.lastRefreshAt = performance.now();
    this.refreshCount += 1;

    if (this.pending) {
      const next = this.pending;
      this.pending = null;
      this.gate.reset();
      return this.show(next.pngBytes, next);
    }
    return { refreshed: true, wiped: true };
  }

  statusLine() {
    if (!this.enabled) return 'e-ink off (instant)';
    const ago = this.lastRefreshAt
      ? `${Math.round((performance.now() - this.lastRefreshAt) / 100) / 10}s ago`
      : 'never';
    const state = this.busy ? 'refreshing…' : `idle · last ${ago}`;
    return `e-ink on · flash+fade ${this.refreshMs}ms · ${this.refreshCount} refreshes · ${state}`;
  }
}

async function main() {
  const canvas = document.getElementById('hud');
  const stats = document.getElementById('stats');
  const meta = document.getElementById('backendMeta');
  const einkToggle = document.getElementById('einkToggle');
  const einkHint = document.getElementById('einkHint');

  const panel = new EinkPanel(canvas, { enabled: einkToggle.checked });
  const updateEinkHint = () => {
    einkHint.textContent = panel.enabled
      ? 'flash all on → off, then ~1s fade-in · ≈ nearest 50 m · redraw ~every 50 m'
      : 'instant frames (no gate, no flash)';
  };
  updateEinkHint();
  einkToggle.addEventListener('change', () => {
    panel.setEnabled(einkToggle.checked);
    updateEinkHint();
    void refreshOnce(true);
  });

  let backend;
  try {
    backend = await createWasmBackend();
    meta.textContent = 'Backend: Go WASM (in-browser HUD core)';
  } catch (err) {
    console.warn('WASM unavailable, using HTTP', err);
    backend = createHttpBackend();
    meta.textContent =
      'Backend: HTTP → motohud (-host emu|png). Build WASM with scripts/build-wasm.sh for offline core.';
  }

  const route = await (await fetch('routes/whitehall-farringdon.json')).json();
  const latlngs = route.coordinates.map(([lng, lat]) => [lat, lng]);

  const map = L.map('map', { zoomControl: true }).setView(latlngs[0], 15);
  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    maxZoom: 19,
    attribution: '&copy; OpenStreetMap',
  }).addTo(map);
  const line = L.polyline(latlngs, { color: '#c4a35a', weight: 5 }).addTo(map);
  map.fitBounds(line.getBounds(), { padding: [24, 24] });
  const bike = L.circleMarker(latlngs[0], {
    radius: 7,
    color: '#111',
    fillColor: '#c4a35a',
    fillOpacity: 1,
    weight: 2,
  }).addTo(map);

  if (route.media) await backend.applyMedia(route.media);

  const totalM = pathLength(route.coordinates);
  let along = 0;
  let playing = false;
  let lastTs = 0;
  let lastNav = nextManeuver(route, 0);

  async function paintFromState({ force = false } = {}) {
    const screen = await backend.screen();
    const displayNav = forPanel(lastNav, panel.enabled);
    if (panel.enabled && !force && !panel.gate.wouldRedraw(screen, displayNav, false)) {
      return { refreshed: false, wiped: false };
    }
    const png = await backend.renderPNG();
    return panel.show(png, { screen, nav: displayNav, force });
  }

  async function pushNav(nav) {
    lastNav = nav;
    const displayNav = forPanel(nav, panel.enabled);
    await backend.applyNav(displayNav);
    return displayNav;
  }

  async function tick(ts) {
    if (!playing) return;
    if (!lastTs) lastTs = ts;
    const dt = Math.min(0.1, (ts - lastTs) / 1000);
    lastTs = ts;
    along = Math.min(totalM, along + route.speed_mps * dt);
    const pos = pointAlong(route.coordinates, along);
    bike.setLatLng([pos.lat, pos.lng]);
    const displayNav = await pushNav(nextManeuver(route, along));
    // Don't await wipe — map keeps moving while the panel refreshes.
    void paintFromState({ force: false });
    const screen = await backend.screen();
    stats.textContent =
      `along ${Math.round(along)} / ${Math.round(totalM)} m · ` +
      `${displayNav.maneuver} · ${displayNav.distance_text} · ${displayNav.road} · screen ${screen}\n` +
      panel.statusLine();
    if (along >= totalM) {
      playing = false;
      document.getElementById('play').disabled = false;
      void paintFromState({ force: true });
      return;
    }
    requestAnimationFrame(tick);
  }

  async function refreshOnce(force = true) {
    await pushNav(nextManeuver(route, along));
    await paintFromState({ force });
    stats.textContent = panel.statusLine();
  }

  document.getElementById('play').addEventListener('click', () => {
    if (along >= totalM) along = 0;
    playing = true;
    lastTs = 0;
    document.getElementById('play').disabled = true;
    requestAnimationFrame(tick);
  });
  document.getElementById('pause').addEventListener('click', () => {
    playing = false;
    document.getElementById('play').disabled = false;
  });
  document.getElementById('reset').addEventListener('click', async () => {
    playing = false;
    along = 0;
    lastTs = 0;
    document.getElementById('play').disabled = false;
    bike.setLatLng(latlngs[0]);
    panel.gate.reset();
    await refreshOnce(true);
  });

  for (const btn of document.querySelectorAll('[data-btn]')) {
    btn.addEventListener('click', async () => {
      await backend.button(btn.getAttribute('data-btn'));
      await pushNav(nextManeuver(route, along));
      await paintFromState({ force: true });
      stats.textContent =
        `screen ${await backend.screen()} (manual)\n` + panel.statusLine();
    });
  }

  await refreshOnce(true);
}

main().catch((err) => {
  document.getElementById('backendMeta').textContent = String(err);
  console.error(err);
});
