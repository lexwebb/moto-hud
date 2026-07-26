import L from 'leaflet';
import markerIcon2x from 'leaflet/dist/images/marker-icon-2x.png';
import markerIcon from 'leaflet/dist/images/marker-icon.png';
import markerShadow from 'leaflet/dist/images/marker-shadow.png';
import {
  pathLength,
  pointAlong,
  nextManeuver,
  turnMarkers,
} from './nav-geometry.js';
import {
  DEVICE_PROFILES,
  HUD_W,
  HUD_H,
  getProfile,
  loadDeviceId,
  saveDeviceId,
  profileRefreshMs,
  profileFullRefreshMs,
} from './display-profiles.js';

// Fix default marker assets under Vite
delete L.Icon.Default.prototype._getIconUrl;
L.Icon.Default.mergeOptions({
  iconRetinaUrl: markerIcon2x,
  iconUrl: markerIcon,
  shadowUrl: markerShadow,
});

const BASE = globalThis.__MOTO_HUD_BASE__ || '/';

/** @type {typeof globalThis & { Go?: new () => any, MotoHUD?: any }} */
const g = globalThis;

/** E-ink presentation: nearest 50 m + ≈ (U+2248, in Terminus). */
function formatEinkDistance(m) {
  const rounded = Math.max(0, Math.round(m / 50) * 50);
  if (rounded >= 1000) {
    const km = (rounded / 1000).toFixed(1);
    return { distance_m: rounded, distance_text: `≈${km}km` };
  }
  return { distance_m: rounded, distance_text: `≈${rounded}m` };
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

/** @typedef {{ applyNav(j:string):any, applyMedia(j:string):any, button(e:string):any, renderPNG():Uint8Array, screen():string }} HudBackend */


async function createWasmBackend() {
  if (typeof g.Go !== 'function') {
    throw new Error('wasm_exec.js not loaded — check public/emulator/wasm_exec.js');
  }
  const wasmUrl = `${BASE}emulator/motohud.wasm`;
  const res = await fetch(wasmUrl);
  if (!res.ok) {
    throw new Error(
      `Failed to load ${wasmUrl} (${res.status}). Run npm run build:wasm in site/ (needs Go).`,
    );
  }
  const go = new g.Go();
  const result = await WebAssembly.instantiateStreaming(res, go.importObject);
  go.run(result.instance);
  for (let i = 0; i < 200; i++) {
    if (g.MotoHUD?.renderPNG) break;
    await new Promise((r) => setTimeout(r, 10));
  }
  if (!g.MotoHUD?.renderPNG) throw new Error('MotoHUD exports missing');
  const api = g.MotoHUD;
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
    this.#applyChrome();
  }

  #applyChrome() {
    const id = this._profile.id;
    this.screenEl?.setAttribute('data-device', id);
    this.bezelEl?.setAttribute('data-device', id);
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

  /** Waveshare-style: base + every Nth frame are full; otherwise partial. */
  #shouldPartial() {
    const p = this._profile;
    if (!p.supportPartial) return false;
    if (!this.refreshCount) return false; // first frame = full base image
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
    // Spec: no flicker — particles move quietly; ~0.3s busy.
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
    console.error(err);
    meta.textContent = String(err);
    throw err;
  }

  const route = await (await fetch(`${BASE}emulator/routes/whitehall-farringdon.json`)).json();
  let roadWays = [];
  try {
    const roads = await (await fetch(`${BASE}emulator/routes/whitehall-farringdon-roads.json`)).json();
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
