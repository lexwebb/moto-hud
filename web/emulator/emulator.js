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

/** Degrees clockwise from north — mirrors Android OsrmRibbon.bearingBetween. */
function bearingBetween(lng1, lat1, lng2, lat2) {
  const φ1 = (lat1 * Math.PI) / 180;
  const φ2 = (lat2 * Math.PI) / 180;
  const Δλ = ((lng2 - lng1) * Math.PI) / 180;
  const y = Math.sin(Δλ) * Math.cos(φ2);
  const x = Math.cos(φ1) * Math.sin(φ2) - Math.sin(φ1) * Math.cos(φ2) * Math.cos(Δλ);
  return (((Math.atan2(y, x) * 180) / Math.PI) + 360) % 360;
}

function projectOne(lat0, lon0, bearingDeg, lon, lat) {
  const br = (bearingDeg * Math.PI) / 180;
  const north = (lat - lat0) * 111320;
  const east = (lon - lon0) * 111320 * Math.cos((lat0 * Math.PI) / 180);
  const ahead = north * Math.cos(br) + east * Math.sin(br);
  const right = east * Math.cos(br) - north * Math.sin(br);
  return { x: Math.round(right), y: Math.round(ahead) };
}

function downsampleRibbon(pts, max = 6) {
  if (pts.length <= max) return pts;
  const out = [pts[0]];
  const mid = max - 2;
  for (let i = 1; i <= mid; i++) {
    const t = i / (mid + 1);
    const idx = Math.min(pts.length - 2, Math.max(1, Math.floor(t * (pts.length - 1))));
    out.push(pts[idx]);
  }
  out.push(pts[pts.length - 1]);
  return out;
}

/**
 * Project the baked OSRM route polyline ahead of the bike into ribbon local space
 * (Y ahead, X right) — same contract as Android OsrmRibbon / protocol ribbon_points.
 */
function corridorRibbon(coords, origin, bearingDeg, turnAt, distanceM) {
  const turnIdx = Math.min(Math.max(turnAt, 0), coords.length - 1);
  const end = Math.min(coords.length - 1, Math.max(turnIdx + 2, origin.index + 2));
  const start = Math.max(0, origin.index);
  const raw = [[origin.lng, origin.lat]];
  for (let i = start + 1; i <= end; i++) raw.push(coords[i]);
  if (raw.length < 2) return null;

  const projected = raw.map(([lon, lat]) =>
    projectOne(origin.lat, origin.lng, bearingDeg, lon, lat),
  );
  const turnCoord = coords[turnIdx];
  const turnProj = projectOne(origin.lat, origin.lng, bearingDeg, turnCoord[0], turnCoord[1]);
  const yMax = Math.max(turnProj.y + 40, distanceM + 80, 80);
  let clipped = projected.filter((p) => p.y >= -25 && p.y <= yMax);
  if (clipped.length < 2) clipped = projected;
  const points = downsampleRibbon(clipped, 6);
  if (points.length < 2) return null;

  let ribbonTurn = 0;
  let best = Infinity;
  points.forEach((p, i) => {
    const d = Math.hypot(p.x - turnProj.x, p.y - turnProj.y);
    if (d < best) {
      best = d;
      ribbonTurn = i;
    }
  });
  return { ribbon_points: points, ribbon_turn: ribbonTurn };
}

const TURN_SNAP_BEHIND = 50; // meters before turn (approach)
const TURN_SNAP_AHEAD = 50; // meters after turn
const TURN_SNAP_HALF_W = 50;

/**
 * Top-down snapshot centered on the next turn.
 * Local frame: turn at origin; +Y = inbound approach direction (rider below).
 */
function wayCoords(way) {
  if (!way) return null;
  if (Array.isArray(way)) return way;
  if (Array.isArray(way.coords)) return way.coords;
  return null;
}

function buildMinimap(coords, riderPos, turnAt, ways) {
  if (!ways || !ways.length) return null;
  const turnIdx = Math.min(Math.max(turnAt, 0), coords.length - 1);
  const turn = coords[turnIdx];
  const turnLng = turn[0];
  const turnLat = turn[1];

  // Inbound bearing: toward the turn along the route.
  let fromIdx = Math.max(0, turnIdx - 1);
  for (let i = turnIdx - 1; i >= 0; i--) {
    if (haversineM(coords[i], turn) >= 12) {
      fromIdx = i;
      break;
    }
    fromIdx = i;
  }
  const inbound = bearingBetween(coords[fromIdx][0], coords[fromIdx][1], turnLng, turnLat);

  const toLocal = (lng, lat) =>
    projectOne(turnLat, turnLng, inbound, lng, lat);

  // Route as approach → turn → departure (origin must sit in path order).
  const approach = [];
  const departure = [];
  for (let i = 0; i < coords.length; i++) {
    const p = toLocal(coords[i][0], coords[i][1]);
    if (p.y < -TURN_SNAP_BEHIND || p.y > TURN_SNAP_AHEAD || Math.abs(p.x) > TURN_SNAP_HALF_W * 1.3) {
      if (departure.length >= 2 && i > turnIdx) break;
      if (i < turnIdx) {
        approach.length = 0;
        continue;
      }
      continue;
    }
    const pt = { x: Math.round(p.x), y: Math.round(p.y) };
    if (i < turnIdx) approach.push(pt);
    else if (i > turnIdx) departure.push(pt);
  }
  const routePts = downsampleRibbon(
    [...downsampleRibbon(approach, 8), { x: 0, y: 0 }, ...downsampleRibbon(departure, 8)],
    16,
  );
  if (routePts.length < 2) return null;

  const riderLocal = toLocal(riderPos.lng, riderPos.lat);
  const rider = {
    x: Math.round(riderLocal.x),
    y: Math.round(riderLocal.y),
  };

  const cosLat = Math.cos((turnLat * Math.PI) / 180);
  const dLat = (TURN_SNAP_BEHIND + 30) / 111320;
  const dLon = (TURN_SNAP_BEHIND + 30) / (111320 * Math.max(0.2, cosLat));

  const context = [];
  for (const way of ways) {
    if (context.length >= 40) break;
    const geom = wayCoords(way);
    if (!geom || geom.length < 2) continue;
    let near = false;
    for (const [lng, lat] of geom) {
      if (Math.abs(lat - turnLat) <= dLat && Math.abs(lng - turnLng) <= dLon) {
        near = true;
        break;
      }
    }
    if (!near) continue;

    const projected = [];
    const flush = () => {
      if (projected.length >= 2) {
        context.push(downsampleRibbon(projected.splice(0, projected.length), 8));
      } else {
        projected.length = 0;
      }
    };
    for (const [lng, lat] of geom) {
      const p = toLocal(lng, lat);
      if (p.y >= -TURN_SNAP_BEHIND && p.y <= TURN_SNAP_AHEAD && Math.abs(p.x) <= TURN_SNAP_HALF_W) {
        projected.push({ x: Math.round(p.x), y: Math.round(p.y) });
      } else {
        flush();
      }
    }
    flush();
  }

  return { route: routePts, context, rider };
}

function nextManeuver(route, alongM, roadWays) {
  const coords = route.coordinates;
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
  const distToNext = pathLength(coords, segIndex, next.at);
  const remainOnSeg =
    haversineM(coords[segIndex], coords[Math.min(segIndex + 1, coords.length - 1)]) -
    (alongM - traveled);
  const distanceM = Math.max(0, Math.round(distToNext + Math.max(0, remainOnSeg)));
  const total = pathLength(coords);
  const done = Math.min(1, alongM / total);
  const etaMin = Math.max(1, Math.round(route.eta_min_start * (1 - done)));
  const arrived = next.maneuver === 'arrive' && distanceM < 15;
  const nav = {
    active: !arrived,
    maneuver: arrived ? 'arrive' : next.maneuver,
    road: next.road,
    distance_m: distanceM,
    distance_text: distanceM >= 1000 ? `${(distanceM / 1000).toFixed(1)} km` : `${distanceM} m`,
    eta_min: etaMin,
    instruction: next.road,
  };
  if (!arrived) {
    const pos = pointAlong(coords, alongM);
    const a = coords[pos.index];
    const b = coords[Math.min(pos.index + 1, coords.length - 1)];
    const bearing = bearingBetween(a[0], a[1], b[0], b[1]);
    const minimap = buildMinimap(coords, pos, next.at, roadWays);
    if (minimap) {
      nav.minimap = minimap;
    } else {
      const ribbon = corridorRibbon(coords, pos, bearing, next.at, distanceM);
      if (ribbon) Object.assign(nav, ribbon);
    }
  }
  return nav;
}

/** E-ink presentation: nearest 50 m + ≈ (U+2248, in Terminus). */
function formatEinkDistance(m) {
  const rounded = Math.max(0, Math.round(m / 50) * 50);
  if (rounded >= 1000) {
    const km = (rounded / 1000).toFixed(1);
    return { distance_m: rounded, distance_text: `≈${km}km` };
  }
  return { distance_m: rounded, distance_text: `≈${rounded}m` };
}

const HUD_W = 250;
const HUD_H = 122;

/** Inky 4-colour BWRY (~20s). Stage sum 12100ms + fade 7900ms. */
const INKY_4COLOUR_SEQUENCE = [
  { fill: '#000000', ms: 700 },
  { fill: '#ffffff', ms: 700 },
  { fill: '#000000', ms: 600 },
  { fill: '#ffffff', ms: 600 },
  { fill: '#c41e3a', ms: 1600 },
  { fill: '#000000', ms: 500 },
  { fill: '#ffffff', ms: 500 },
  { fill: '#f0c400', ms: 1800 },
  { fill: '#000000', ms: 500 },
  { fill: '#c41e3a', ms: 1200 },
  { fill: '#ffffff', ms: 600 },
  { fill: '#f0c400', ms: 1200 },
  { fill: '#000000', ms: 700 },
  { fill: '#ffffff', ms: 900 },
];

/**
 * Waveshare 2.13″ B/W V3/V4 full refresh (~2s).
 * Spec: full flickers several times; fast flashes once (~1.8s);
 * partial has no flicker (~0.3s). Full after every ~5 partials.
 * Stage sum 1300ms + fade 700ms = 2000ms.
 */
const WAVESHARE_FULL_SEQUENCE = [
  { fill: '#000000', ms: 200 },
  { fill: '#ffffff', ms: 200 },
  { fill: '#000000', ms: 250 },
  { fill: '#ffffff', ms: 250 },
  { fill: '#000000', ms: 200 },
  { fill: '#ffffff', ms: 200 },
];

const DEVICE_PROFILES = {
  inky: {
    id: 'inky',
    label: 'Inky pHAT (4-colour)',
    kind: 'eink',
    panelW: HUD_W,
    panelH: HUD_H,
    sequence: INKY_4COLOUR_SEQUENCE,
    fadeMs: 7900,
    distanceStep: true,
    letterbox: false,
    hint: '4-colour BWRY · ~20s (inverse → R/Y flash → settle) · ≈ 50 m steps · HUD stays 1-bit',
  },
  waveshare: {
    id: 'waveshare',
    label: 'Waveshare 2.13″ B/W',
    kind: 'eink',
    panelW: HUD_W,
    panelH: HUD_H,
    sequence: WAVESHARE_FULL_SEQUENCE,
    fadeMs: 700,
    supportPartial: true,
    partialMs: 300,
    fullEveryN: 5,
    distanceStep: true,
    letterbox: false,
    hint: 'B/W · partial ~0.3s (no flicker) · full ~2s every 5 · ≈ 50 m steps',
  },
  lcd: {
    id: 'lcd',
    label: 'Display HAT Mini (LCD)',
    kind: 'lcd',
    panelW: 320,
    panelH: 240,
    flashMs: 0,
    fadeMs: 0,
    distanceStep: false,
    letterbox: true,
    hint: 'instant frames · 250×122 letterboxed on 320×240 · no e-ink gate',
  },
};
const DEFAULT_DEVICE = 'waveshare';
const DEVICE_STORAGE_KEY = 'moto-hud-emulator-device';

function loadDeviceId() {
  try {
    const id = localStorage.getItem(DEVICE_STORAGE_KEY);
    if (id && DEVICE_PROFILES[id]) return id;
  } catch {
    /* ignore */
  }
  return DEFAULT_DEVICE;
}

function saveDeviceId(id) {
  try {
    localStorage.setItem(DEVICE_STORAGE_KEY, id);
  } catch {
    /* ignore */
  }
}

function getProfile(id) {
  return DEVICE_PROFILES[id] || DEVICE_PROFILES[DEFAULT_DEVICE];
}

function profileRefreshMs(profile) {
  if (!profile || profile.kind !== 'eink') return 0;
  if (profile.supportPartial && profile.partialMs != null) {
    return profile.partialMs;
  }
  if (profile.sequence?.length) {
    const flash = profile.sequence.reduce((s, step) => s + step.ms, 0);
    return flash + (profile.fadeMs || 0);
  }
  return (profile.flashMs ?? 0) * 2 + (profile.fadeMs ?? 0);
}

function profileFullRefreshMs(profile) {
  if (!profile || profile.kind !== 'eink') return 0;
  if (profile.sequence?.length) {
    const flash = profile.sequence.reduce((s, step) => s + step.ms, 0);
    return flash + (profile.fadeMs || 0);
  }
  return (profile.flashMs ?? 0) * 2 + (profile.fadeMs ?? 0);
}

function forPanel(nav, profile) {
  if (!profile?.distanceStep) return nav;
  const d = formatEinkDistance(nav.distance_m);
  return { ...nav, ...d };
}

function formatClock(sec) {
  const s = Math.max(0, Math.round(sec));
  const m = Math.floor(s / 60);
  const r = s % 60;
  return `${m}:${String(r).padStart(2, '0')}`;
}

/** Maneuver markers with along-route meters for the scrubber. */
function turnMarkers(route) {
  const coords = route.coordinates;
  const total = pathLength(coords);
  return (route.maneuvers || []).map((m, i) => {
    const alongM = Math.min(total, pathLength(coords, 0, m.at));
    return {
      index: i,
      at: m.at,
      alongM,
      maneuver: m.maneuver,
      road: m.road,
      label: `${m.maneuver.replace(/_/g, ' ')} · ${m.road}`,
    };
  });
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
 * Device panel: e-ink wipe+fade (Inky / Waveshare) or instant LCD letterbox.
 */
class DevicePanel {
  constructor(canvas, profile) {
    this.canvas = canvas;
    this.screenEl = canvas.closest('.screen');
    this.bezelEl = canvas.closest('.bezel');
    this.gate = new RefreshGate();
    this.busy = false;
    this.pending = null;
    this.lastRefreshAt = 0;
    this.refreshCount = 0;
    this.partialCount = 0;
    this.lastMode = null;
    this.setProfile(profile);
  }

  get profile() {
    return this._profile;
  }

  get gated() {
    return this._profile.kind === 'eink';
  }

  get refreshMs() {
    return profileRefreshMs(this._profile);
  }

  setProfile(profile) {
    this._profile = profile;
    this.gate.reset();
    this.partialCount = 0;
    this.lastMode = null;
    this.screenEl?.setAttribute('data-device', profile.id);
    this.bezelEl?.setAttribute('data-device', profile.id);
  }

  async #paintFrame(pngBytes) {
    const { panelW, panelH, letterbox } = this._profile;
    const bitmap = await createImageBitmap(new Blob([pngBytes], { type: 'image/png' }));
    const ctx = this.canvas.getContext('2d', { alpha: false });
    this.canvas.width = panelW;
    this.canvas.height = panelH;
    ctx.imageSmoothingEnabled = false;
    ctx.fillStyle = '#fff';
    ctx.fillRect(0, 0, panelW, panelH);
    if (letterbox) {
      const ox = Math.floor((panelW - HUD_W) / 2);
      const oy = Math.floor((panelH - HUD_H) / 2);
      ctx.drawImage(bitmap, ox, oy);
    } else {
      ctx.drawImage(bitmap, 0, 0);
    }
    bitmap.close();
  }

  async show(pngBytes, { screen = 'nav', nav = {}, force = false } = {}) {
    if (!this.gated) {
      await this.#paintFrame(pngBytes);
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
    const usePartial = this.#shouldPartial();
    return usePartial ? this.#runPartial(pngBytes) : this.#runRefresh(pngBytes);
  }

  #shouldPartial() {
    const p = this._profile;
    if (!p.supportPartial) return false;
    if (!this.refreshCount) return false;
    const every = p.fullEveryN ?? 5;
    return this.partialCount < every;
  }

  #flashStages() {
    const { sequence, flashMs } = this._profile;
    if (sequence?.length) return sequence;
    const ms = flashMs ?? 120;
    return [
      { fill: '#000000', ms },
      { fill: '#ffffff', ms },
    ];
  }

  async #runPartial(pngBytes) {
    this.screenEl?.classList.add('partial');
    const settle = this._profile.partialMs ?? 300;
    const bitmap = await createImageBitmap(new Blob([pngBytes], { type: 'image/png' }));
    const { panelW, panelH } = this._profile;
    const ctx = this.canvas.getContext('2d', { alpha: false });
    const t0 = performance.now();
    await new Promise((resolve) => {
      const step = (now) => {
        const t = Math.min(1, (now - t0) / settle);
        const a = 1 - (1 - t) ** 2;
        ctx.globalAlpha = a;
        ctx.drawImage(bitmap, 0, 0, panelW, panelH);
        ctx.globalAlpha = 1;
        if (t < 1) requestAnimationFrame(step);
        else resolve();
      };
      requestAnimationFrame(step);
    });
    ctx.drawImage(bitmap, 0, 0, panelW, panelH);
    bitmap.close();

    this.screenEl?.classList.remove('partial');
    this.busy = false;
    this.lastRefreshAt = performance.now();
    this.refreshCount += 1;
    this.partialCount += 1;
    this.lastMode = 'partial';

    if (this.pending) {
      const next = this.pending;
      this.pending = null;
      this.gate.reset();
      return this.show(next.pngBytes, next);
    }
    return { refreshed: true, wiped: false, mode: 'partial' };
  }

  async #runRefresh(pngBytes) {
    const { panelW, panelH, fadeMs } = this._profile;
    this.screenEl?.classList.add('wiping');
    const ctx = this.canvas.getContext('2d', { alpha: false });
    this.canvas.width = panelW;
    this.canvas.height = panelH;
    ctx.imageSmoothingEnabled = false;

    for (const step of this.#flashStages()) {
      ctx.fillStyle = step.fill;
      ctx.fillRect(0, 0, panelW, panelH);
      await sleep(step.ms);
    }

    const bitmap = await createImageBitmap(new Blob([pngBytes], { type: 'image/png' }));
    const settle = fadeMs || 1000;
    const t0 = performance.now();
    await new Promise((resolve) => {
      const step = (now) => {
        const t = Math.min(1, (now - t0) / settle);
        const a = 1 - (1 - t) ** 2;
        ctx.globalAlpha = 1;
        ctx.fillStyle = '#fff';
        ctx.fillRect(0, 0, panelW, panelH);
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
    ctx.fillRect(0, 0, panelW, panelH);
    ctx.drawImage(bitmap, 0, 0);
    bitmap.close();

    this.screenEl?.classList.remove('wiping');
    this.busy = false;
    this.lastRefreshAt = performance.now();
    this.refreshCount += 1;
    this.partialCount = 0;
    this.lastMode = 'full';

    if (this.pending) {
      const next = this.pending;
      this.pending = null;
      this.gate.reset();
      return this.show(next.pngBytes, next);
    }
    return { refreshed: true, wiped: true, mode: 'full' };
  }

  statusLine() {
    const ago = this.lastRefreshAt
      ? `${Math.round((performance.now() - this.lastRefreshAt) / 100) / 10}s ago`
      : 'never';
    const state = this.busy ? 'refreshing…' : `idle · last ${ago}`;
    if (!this.gated) {
      return `${this._profile.label} · instant · ${this.refreshCount} frames · ${state}`;
    }
    if (this._profile.supportPartial) {
      const partSec = Math.round((this._profile.partialMs ?? 300) / 100) / 10;
      const fullSec = Math.round(profileFullRefreshMs(this._profile) / 100) / 10;
      const mode = this.lastMode ? ` · last ${this.lastMode}` : '';
      const every = this._profile.fullEveryN ?? 5;
      return (
        `${this._profile.label} · partial ~${partSec}s / full ~${fullSec}s` +
        ` (every ${every}) · ${this.refreshCount} refreshes${mode} · ${state}`
      );
    }
    const sec = Math.round(this.refreshMs / 100) / 10;
    return (
      `${this._profile.label} · ~${sec}s wipe · ` +
      `${this.refreshCount} refreshes · ${state}`
    );
  }
}

async function main() {
  const canvas = document.getElementById('hud');
  const stats = document.getElementById('stats');
  const meta = document.getElementById('backendMeta');
  const deviceSelect = document.getElementById('deviceSelect');
  const deviceHint = document.getElementById('deviceHint');

  for (const [id, p] of Object.entries(DEVICE_PROFILES)) {
    const opt = document.createElement('option');
    opt.value = id;
    opt.textContent = p.label;
    deviceSelect.appendChild(opt);
  }
  deviceSelect.value = loadDeviceId();
  const panel = new DevicePanel(canvas, getProfile(deviceSelect.value));
  const updateDeviceHint = () => {
    deviceHint.textContent = panel.profile.hint;
  };
  updateDeviceHint();
  deviceSelect.addEventListener('change', () => {
    const id = deviceSelect.value;
    saveDeviceId(id);
    panel.setProfile(getProfile(id));
    updateDeviceHint();
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
  let roadWays = [];
  try {
    const roads = await (await fetch('routes/whitehall-farringdon-roads.json')).json();
    roadWays = roads.ways || [];
  } catch (e) {
    console.warn('OSM context roads missing', e);
  }
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
  const speed = route.speed_mps || 8;
  const totalSec = totalM / speed;
  const turns = turnMarkers(route);
  let along = 0;
  let playing = false;
  let lastTs = 0;
  let lastNav = nextManeuver(route, 0, roadWays);
  let scrubbing = false;

  const scrub = document.getElementById('scrub');
  const timelinePos = document.getElementById('timelinePos');
  const timelineTime = document.getElementById('timelineTime');
  const timelineMarks = document.getElementById('timelineMarks');
  const timelineTurns = document.getElementById('timelineTurns');

  scrub.max = String(Math.max(1, Math.round(totalM)));
  scrub.value = '0';

  if (turns.length && timelineMarks && timelineTurns) {
    timelineMarks.hidden = false;
    timelineMarks.replaceChildren(
      ...turns.map((t) => {
        const tick = document.createElement('span');
        tick.style.left = `${(t.alongM / totalM) * 100}%`;
        tick.title = t.label;
        return tick;
      }),
    );
    timelineTurns.replaceChildren(
      ...turns.map((t) => {
        const btn = document.createElement('button');
        btn.type = 'button';
        btn.dataset.along = String(Math.round(t.alongM));
        btn.dataset.turn = String(t.index);
        btn.title = `${Math.round(t.alongM)} m · ${formatClock(t.alongM / speed)}`;
        btn.textContent = `${t.index + 1}. ${t.maneuver.replace(/_/g, ' ')}`;
        btn.addEventListener('click', () => {
          void seekTo(t.alongM, { pause: true, force: true });
        });
        return btn;
      }),
    );
  }

  function syncTimeline(nav) {
    if (!scrubbing) scrub.value = String(Math.round(along));
    timelinePos.textContent = `${Math.round(along)} / ${Math.round(totalM)} m`;
    timelineTime.textContent =
      `${formatClock(along / speed)} / ${formatClock(totalSec)}` +
      (nav?.eta_min != null ? ` · ETA ${nav.eta_min} min` : '');
    if (timelineTurns) {
      let active = -1;
      for (let i = 0; i < turns.length; i++) {
        if (turns[i].alongM <= along + 1) active = i;
      }
      for (const btn of timelineTurns.querySelectorAll('button')) {
        btn.classList.toggle('active', Number(btn.dataset.turn) === active);
      }
    }
  }

  async function paintFromState({ force = false } = {}) {
    const screen = await backend.screen();
    const displayNav = forPanel(lastNav, panel.profile);
    if (panel.gated && !force && !panel.gate.wouldRedraw(screen, displayNav, false)) {
      return { refreshed: false, wiped: false };
    }
    const png = await backend.renderPNG();
    return panel.show(png, { screen, nav: displayNav, force });
  }

  async function pushNav(nav) {
    lastNav = nav;
    const displayNav = forPanel(nav, panel.profile);
    await backend.applyNav(displayNav);
    return displayNav;
  }

  async function seekTo(meters, { pause = false, force = true } = {}) {
    if (pause) {
      playing = false;
      document.getElementById('play').disabled = false;
    }
    along = Math.max(0, Math.min(totalM, meters));
    const pos = pointAlong(route.coordinates, along);
    bike.setLatLng([pos.lat, pos.lng]);
    panel.gate.reset();
    const displayNav = await pushNav(nextManeuver(route, along, roadWays));
    await paintFromState({ force });
    const screen = await backend.screen();
    syncTimeline(displayNav);
    stats.textContent =
      `along ${Math.round(along)} / ${Math.round(totalM)} m · ` +
      `${displayNav.maneuver} · ${displayNav.distance_text} · ${displayNav.road} · screen ${screen}\n` +
      panel.statusLine();
    return displayNav;
  }

  async function tick(ts) {
    if (!playing) return;
    if (!lastTs) lastTs = ts;
    const dt = Math.min(0.1, (ts - lastTs) / 1000);
    lastTs = ts;
    along = Math.min(totalM, along + speed * dt);
    const pos = pointAlong(route.coordinates, along);
    bike.setLatLng([pos.lat, pos.lng]);
    const displayNav = await pushNav(nextManeuver(route, along, roadWays));
    syncTimeline(displayNav);
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
    await pushNav(nextManeuver(route, along, roadWays));
    await paintFromState({ force });
    syncTimeline(lastNav);
    stats.textContent = panel.statusLine();
  }

  scrub.addEventListener('pointerdown', () => {
    scrubbing = true;
    playing = false;
    document.getElementById('play').disabled = false;
  });
  scrub.addEventListener('input', () => {
    void seekTo(Number(scrub.value), { pause: true, force: true });
  });
  scrub.addEventListener('change', () => {
    scrubbing = false;
    void seekTo(Number(scrub.value), { pause: true, force: true });
  });
  scrub.addEventListener('pointerup', () => {
    scrubbing = false;
  });

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
    await seekTo(0, { pause: true, force: true });
  });

  for (const btn of document.querySelectorAll('[data-btn]')) {
    btn.addEventListener('click', async () => {
      await backend.button(btn.getAttribute('data-btn'));
      await pushNav(nextManeuver(route, along, roadWays));
      await paintFromState({ force: true });
      syncTimeline(lastNav);
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
