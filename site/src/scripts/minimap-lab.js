import L from 'leaflet';
import markerIcon2x from 'leaflet/dist/images/marker-icon-2x.png';
import markerIcon from 'leaflet/dist/images/marker-icon.png';
import markerShadow from 'leaflet/dist/images/marker-shadow.png';
import {
  paintMeterTruth,
  minimapViewRadius,
  packBits,
  unpackBits,
} from './nav-geometry.js';
import {
  snapCursor,
  schematizeGrid,
  rasterizeVectors,
  DEFAULT_GRID_M,
} from './tube-raster.js';

delete L.Icon.Default.prototype._getIconUrl;
L.Icon.Default.mergeOptions({
  iconRetinaUrl: markerIcon2x,
  iconUrl: markerIcon,
  shadowUrl: markerShadow,
});

const BASE = globalThis.__MOTO_HUD_BASE__ || '/';
/** @type {typeof globalThis & { Go?: new () => any, MotoHUD?: any }} */
const g = globalThis;

const W = 70;
const H = 80;

async function createWasmBackend() {
  if (typeof g.Go !== 'function') throw new Error('wasm_exec.js not loaded');
  const res = await fetch(`${BASE}emulator/motohud.wasm`);
  if (!res.ok) throw new Error(`Failed to load wasm (${res.status})`);
  const go = new g.Go();
  const result = await WebAssembly.instantiateStreaming(res, go.importObject);
  go.run(result.instance);
  for (let i = 0; i < 200; i++) {
    if (g.MotoHUD?.renderMinimapPNG) break;
    await new Promise((r) => setTimeout(r, 10));
  }
  if (!g.MotoHUD?.renderMinimapPNG) throw new Error('renderMinimapPNG missing — rebuild WASM');
  return g.MotoHUD;
}

async function paintPng(canvas, bytes) {
  const bitmap = await createImageBitmap(new Blob([bytes], { type: 'image/png' }));
  const ctx = canvas.getContext('2d', { alpha: false });
  canvas.width = bitmap.width;
  canvas.height = bitmap.height;
  ctx.imageSmoothingEnabled = false;
  ctx.fillStyle = '#fff';
  ctx.fillRect(0, 0, canvas.width, canvas.height);
  ctx.drawImage(bitmap, 0, 0);
  bitmap.close();
}

function paintBits(canvas, bits, { color = '#000', dashed = false } = {}) {
  const ctx = canvas.getContext('2d', { alpha: false });
  canvas.width = W;
  canvas.height = H;
  ctx.imageSmoothingEnabled = false;
  ctx.fillStyle = '#fff';
  ctx.fillRect(0, 0, W, H);
  ctx.fillStyle = color;
  for (let y = 0; y < H; y++) {
    for (let x = 0; x < W; x++) {
      if (!bits[y * W + x]) continue;
      if (dashed && (x + y) % 2 === 1) continue;
      ctx.fillRect(x, y, 1, 1);
    }
  }
}

function paintComposite(canvas, routeBits, contextBits) {
  const ctx = canvas.getContext('2d', { alpha: false });
  canvas.width = W;
  canvas.height = H;
  ctx.imageSmoothingEnabled = false;
  ctx.fillStyle = '#fff';
  ctx.fillRect(0, 0, W, H);
  ctx.fillStyle = '#000';
  for (let y = 0; y < H; y++) {
    for (let x = 0; x < W; x++) {
      const i = y * W + x;
      if (routeBits[i] || contextBits[i]) ctx.fillRect(x, y, 1, 1);
    }
  }
}

function cloneWay(way) {
  return (way || []).map((p) => ({ x: p.x, y: p.y }));
}

function zoomForRadius(lat, radiusM, pixelExtent) {
  const mpp = (2 * radiusM) / Math.max(1, pixelExtent);
  const cos = Math.max(0.2, Math.cos((lat * Math.PI) / 180));
  return Math.log2((156543.03392 * cos) / mpp);
}

async function main() {
  const listEl = document.getElementById('fixtureList');
  const metaEl = document.getElementById('fixtureMeta');
  const acceptEl = document.getElementById('acceptance');
  const statusEl = document.getElementById('status');
  const comparePane = document.getElementById('comparePane');
  const labelPane = document.getElementById('labelPane');
  const truth = document.getElementById('truth');
  const wasmPane = document.getElementById('wasmPane');
  const hud = document.getElementById('hud');
  const layerRoute = document.getElementById('layerRoute');
  const layerContext = document.getElementById('layerContext');
  const layerComposite = document.getElementById('layerComposite');
  const labelNotes = document.getElementById('labelNotes');
  const vectorPad = document.getElementById('vectorPad');
  const rotator = document.getElementById('labelMapRotator');
  const shell = document.getElementById('labelMapShell');

  let api;
  try {
    api = await createWasmBackend();
    statusEl.textContent = 'WASM ready';
  } catch (err) {
    listEl.textContent = String(err);
    statusEl.textContent = String(err);
    throw err;
  }

  const index = await (await fetch(`${BASE}emulator/minimap-fixtures/index.json`)).json();
  /** @type {any} */
  let current = null;
  let mode = 'compare';
  let drawLayer = 'route';
  /** @type {{x:number,y:number}[][]} */
  let routeWays = [];
  /** @type {{x:number,y:number}[][]} */
  let contextWays = [];
  /** @type {{x:number,y:number}[]} */
  let draftStroke = [];
  /** @type {{x:number,y:number}|null} */
  let cursorLocal = null;
  let routeBits = new Uint8Array(W * H);
  let contextBits = new Uint8Array(W * H);
  let viewRadiusM = 50;

  /** @type {L.Map | null} */
  let map = null;
  /** @type {L.Polyline | null} */
  let routeLine = null;
  /** @type {L.CircleMarker | null} */
  let turnMark = null;
  /** @type {L.CircleMarker | null} */
  let riderMark = null;
  /** @type {any} */
  let bakedRoute = null;

  async function loadBakedRoute() {
    if (bakedRoute) return bakedRoute;
    bakedRoute = await (await fetch(`${BASE}emulator/routes/whitehall-farringdon.json`)).json();
    return bakedRoute;
  }

  function setMode(next) {
    mode = next;
    comparePane.hidden = mode !== 'compare';
    labelPane.hidden = mode !== 'label';
    for (const btn of document.querySelectorAll('.mode-tabs [data-mode]')) {
      btn.classList.toggle('active', btn.getAttribute('data-mode') === mode);
    }
    if (mode === 'label' && current) {
      void (async () => {
        ensureMap();
        await loadBakedRoute();
        syncMapFrame();
        redrawPad();
        refreshLayerCanvases();
        setTimeout(() => {
          map?.invalidateSize();
          syncMapFrame();
          redrawPad();
        }, 60);
      })();
    }
  }

  for (const btn of document.querySelectorAll('.mode-tabs [data-mode]')) {
    btn.addEventListener('click', () => setMode(btn.getAttribute('data-mode')));
  }

  for (const btn of document.querySelectorAll('.tool-row [data-layer]')) {
    btn.addEventListener('click', () => {
      if (draftStroke.length) finishStroke();
      drawLayer = btn.getAttribute('data-layer');
      for (const b of document.querySelectorAll('.tool-row [data-layer]')) {
        b.classList.toggle('active', b === btn);
      }
      statusEl.textContent = `drawing ${drawLayer} · ${DEFAULT_GRID_M} m grid · dbl-click/Enter to finish`;
    });
  }

  function draftKey(id) {
    return `minimap-label:${id}`;
  }

  async function loadAcceptedLabel(id) {
    try {
      const res = await fetch(`${BASE}emulator/minimap-labels/${id}.accepted.json`);
      if (!res.ok) return null;
      return await res.json();
    } catch {
      return null;
    }
  }

  function recomputeRaster() {
    const mm = current?.minimap || current?.nav?.minimap;
    const rider = mm?.rider || null;
    const out = rasterizeVectors({
      routeWays,
      contextWays,
      rider,
      w: W,
      h: H,
      radius: viewRadiusM,
    });
    routeBits = out.routeBits;
    contextBits = out.contextBits;
  }

  function refreshLayerCanvases() {
    recomputeRaster();
    paintBits(layerRoute, routeBits, { color: '#111' });
    paintBits(layerContext, contextBits, { color: '#444', dashed: true });
    paintComposite(layerComposite, routeBits, contextBits);
  }

  function ensureMap() {
    if (map) return;
    map = L.map('labelMap', {
      zoomControl: false,
      attributionControl: true,
      dragging: false,
      scrollWheelZoom: false,
      doubleClickZoom: false,
      boxZoom: false,
      keyboard: false,
    });
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      maxZoom: 19,
      attribution: '&copy; OpenStreetMap',
    }).addTo(map);
  }

  function syncMapFrame() {
    if (!map || !current?.turn_lat) return;
    const mm = current.minimap || current.nav?.minimap;
    viewRadiusM = current.view_radius_m || minimapViewRadius(mm);
    const bearing = current.inbound_bearing_deg ?? 0;
    // Rotate map so inbound (+Y / approach) points up on screen.
    rotator.style.transform = `translate(-50%, -50%) rotate(${-bearing}deg)`;

    const padH = shell.clientHeight || 400;
    const z = zoomForRadius(current.turn_lat, viewRadiusM * 1.05, padH);
    map.setView([current.turn_lat, current.turn_lng], z, { animate: false });
    map.invalidateSize();

    // Same gold route polyline as the emulator, so you can see where to stroke.
    if (bakedRoute?.coordinates?.length) {
      const latlngs = bakedRoute.coordinates.map(([lng, lat]) => [lat, lng]);
      if (routeLine) routeLine.setLatLngs(latlngs);
      else {
        routeLine = L.polyline(latlngs, {
          color: '#c4a35a',
          weight: 5,
          opacity: 0.95,
          lineJoin: 'round',
          lineCap: 'round',
        }).addTo(map);
      }
    }

    if (turnMark) turnMark.setLatLng([current.turn_lat, current.turn_lng]);
    else {
      turnMark = L.circleMarker([current.turn_lat, current.turn_lng], {
        radius: 7,
        color: '#111',
        weight: 2,
        fillColor: '#c4a35a',
        fillOpacity: 1,
      }).addTo(map);
    }

    if (current.rider_lat != null) {
      if (riderMark) riderMark.setLatLng([current.rider_lat, current.rider_lng]);
      else {
        riderMark = L.circleMarker([current.rider_lat, current.rider_lng], {
          radius: 6,
          color: '#111',
          weight: 2,
          fillColor: '#fff',
          fillOpacity: 1,
        }).addTo(map);
      }
    }
  }

  function localToPad(x, y) {
    const pw = vectorPad.width;
    const ph = vectorPad.height;
    const scale = Math.min(pw, ph) / (2 * viewRadiusM);
    return {
      x: pw / 2 + x * scale,
      y: ph / 2 - y * scale,
    };
  }

  function padToLocal(px, py) {
    const pw = vectorPad.width;
    const ph = vectorPad.height;
    const scale = Math.min(pw, ph) / (2 * viewRadiusM);
    return {
      x: (px - pw / 2) / scale,
      y: (ph / 2 - py) / scale,
    };
  }

  function eventToLocal(ev) {
    const rect = vectorPad.getBoundingClientRect();
    const scaleX = vectorPad.width / rect.width;
    const scaleY = vectorPad.height / rect.height;
    const px = (ev.clientX - rect.left) * scaleX;
    const py = (ev.clientY - rect.top) * scaleY;
    return padToLocal(px, py);
  }

  function drawWay(ctx, way, { color, width, dash = null, schematize = true, grid = DEFAULT_GRID_M }) {
    if (!way || way.length < 1) return;
    const pts = schematize && way.length >= 2 ? schematizeGrid(way, grid) : way;
    ctx.beginPath();
    pts.forEach((p, i) => {
      const s = localToPad(p.x, p.y);
      if (i === 0) ctx.moveTo(s.x, s.y);
      else ctx.lineTo(s.x, s.y);
    });
    ctx.strokeStyle = color;
    ctx.lineWidth = width;
    ctx.lineCap = 'square';
    ctx.lineJoin = 'miter';
    ctx.setLineDash(dash || []);
    ctx.stroke();
    ctx.setLineDash([]);
    for (const p of pts) {
      const s = localToPad(p.x, p.y);
      ctx.fillStyle = color;
      ctx.fillRect(s.x - 3, s.y - 3, 6, 6);
    }
  }

  function redrawPad() {
    const rect = shell.getBoundingClientRect();
    const dpr = Math.min(2, window.devicePixelRatio || 1);
    vectorPad.width = Math.max(1, Math.round(rect.width * dpr));
    vectorPad.height = Math.max(1, Math.round(rect.height * dpr));

    const ctx = vectorPad.getContext('2d');
    ctx.clearRect(0, 0, vectorPad.width, vectorPad.height);

    // Meter grid (label snap lattice)
    const g = DEFAULT_GRID_M;
    ctx.strokeStyle = 'rgba(255,255,255,0.07)';
    ctx.lineWidth = 1 * dpr;
    for (let m = -Math.ceil(viewRadiusM); m <= Math.ceil(viewRadiusM); m += g) {
      const v0 = localToPad(m, -viewRadiusM);
      const v1 = localToPad(m, viewRadiusM);
      ctx.beginPath();
      ctx.moveTo(v0.x, v0.y);
      ctx.lineTo(v1.x, v1.y);
      ctx.stroke();
      const h0 = localToPad(-viewRadiusM, m);
      const h1 = localToPad(viewRadiusM, m);
      ctx.beginPath();
      ctx.moveTo(h0.x, h0.y);
      ctx.lineTo(h1.x, h1.y);
      ctx.stroke();
    }

    // HUD frame
    const tl = localToPad(-viewRadiusM, viewRadiusM);
    const br = localToPad(viewRadiusM, -viewRadiusM);
    ctx.strokeStyle = 'rgba(196,163,90,0.95)';
    ctx.lineWidth = 2 * dpr;
    ctx.strokeRect(tl.x, tl.y, br.x - tl.x, br.y - tl.y);

    // Origin / turn
    const o = localToPad(0, 0);
    ctx.fillStyle = '#c4a35a';
    ctx.fillRect(o.x - 4, o.y - 4, 8, 8);

    // Ghost original geometry (faint) + stronger route corridor for tracing
    const mm = current?.minimap;
    if (mm) {
      for (const way of mm.context || []) {
        drawWay(ctx, way, { color: 'rgba(255,255,255,0.18)', width: 1.5 * dpr, dash: [4, 4], schematize: false });
      }
      drawWay(ctx, mm.route, { color: 'rgba(196,163,90,0.9)', width: 5 * dpr, schematize: false });
    }

    for (const way of contextWays) {
      drawWay(ctx, way, { color: 'rgba(180,180,180,0.95)', width: 2 * dpr, dash: [6, 6] });
    }
    for (const way of routeWays) {
      drawWay(ctx, way, { color: '#111', width: 4 * dpr });
    }

    // In-progress stroke + rubber band
    if (draftStroke.length || cursorLocal) {
      const prev = draftStroke.length ? draftStroke[draftStroke.length - 1] : null;
      const rubber = cursorLocal ? (prev ? snapCursor(prev, cursorLocal) : snapCursor(null, cursorLocal)) : null;
      const live = rubber ? [...draftStroke, rubber] : [...draftStroke];
      drawWay(ctx, live, {
        color: drawLayer === 'route' ? '#000' : 'rgba(200,200,200,0.9)',
        width: drawLayer === 'route' ? 4 * dpr : 2 * dpr,
        dash: drawLayer === 'context' ? [6, 6] : null,
      });
      // Grid-direction guides from last vertex
      if (prev && rubber) {
        ctx.strokeStyle = 'rgba(196,163,90,0.35)';
        ctx.lineWidth = 1 * dpr;
        for (let k = 0; k < 8; k++) {
          const ang = (k * Math.PI) / 4;
          const far = {
            x: prev.x + Math.cos(ang) * viewRadiusM,
            y: prev.y + Math.sin(ang) * viewRadiusM,
          };
          const a = localToPad(prev.x, prev.y);
          const b = localToPad(far.x, far.y);
          ctx.beginPath();
          ctx.moveTo(a.x, a.y);
          ctx.lineTo(b.x, b.y);
          ctx.stroke();
        }
      }
    }
  }

  function finishStroke() {
    if (draftStroke.length < 2) {
      draftStroke = [];
      redrawPad();
      return;
    }
    const snapped = schematizeGrid(draftStroke, DEFAULT_GRID_M);
    if (drawLayer === 'route') routeWays.push(snapped);
    else contextWays.push(snapped);
    draftStroke = [];
    refreshLayerCanvases();
    redrawPad();
    statusEl.textContent = `${drawLayer} stroke saved (${snapped.length} verts)`;
  }

  vectorPad.addEventListener('pointermove', (ev) => {
    if (mode !== 'label') return;
    cursorLocal = eventToLocal(ev);
    redrawPad();
  });
  vectorPad.addEventListener('pointerleave', () => {
    cursorLocal = null;
    redrawPad();
  });
  vectorPad.addEventListener('click', (ev) => {
    if (mode !== 'label') return;
    const raw = eventToLocal(ev);
    const prev = draftStroke.length ? draftStroke[draftStroke.length - 1] : null;
    const pt = prev ? snapCursor(prev, raw) : snapCursor(null, raw);
    if (prev && Math.hypot(pt.x - prev.x, pt.y - prev.y) < DEFAULT_GRID_M * 0.5) return;
    draftStroke.push(pt);
    redrawPad();
    statusEl.textContent = `${drawLayer}: ${draftStroke.length} vertices (dbl-click / Enter to finish)`;
  });
  vectorPad.addEventListener('dblclick', (ev) => {
    ev.preventDefault();
    finishStroke();
  });
  window.addEventListener('keydown', (ev) => {
    if (mode !== 'label') return;
    if (ev.key === 'Enter') {
      ev.preventDefault();
      finishStroke();
    } else if (ev.key === 'Escape') {
      draftStroke = [];
      redrawPad();
    } else if (ev.key === 'Backspace' && draftStroke.length) {
      ev.preventDefault();
      draftStroke.pop();
      redrawPad();
    }
  });

  function seedFromGeometry() {
    if (!current) return;
    const mm = current.minimap || current.nav?.minimap;
    routeWays = mm?.route?.length >= 2 ? [schematizeGrid(cloneWay(mm.route), DEFAULT_GRID_M)] : [];
    contextWays = (mm?.context || [])
      .filter((w) => w && w.length >= 2)
      .map((w) => schematizeGrid(cloneWay(w), DEFAULT_GRID_M));
    draftStroke = [];
    refreshLayerCanvases();
    redrawPad();
    statusEl.textContent = `seeded ${routeWays.length} route + ${contextWays.length} context strokes (${DEFAULT_GRID_M} m grid)`;
  }

  function buildLabelPayload(status) {
    recomputeRaster();
    return {
      id: current.id,
      status,
      title: current.title,
      along_m: current.along_m,
      notes: labelNotes.value.trim(),
      turn_lat: current.turn_lat,
      turn_lng: current.turn_lng,
      inbound_bearing_deg: current.inbound_bearing_deg,
      view_radius_m: viewRadiusM,
      grid_m: DEFAULT_GRID_M,
      vectors: {
        route: routeWays,
        context: contextWays,
      },
      glyph: {
        w: W,
        h: H,
        route_bits: packBits(routeBits, W, H),
        context_bits: packBits(contextBits, W, H),
      },
      minimap: current.minimap,
      source_fixture: current.id,
      saved_at: new Date().toISOString(),
    };
  }

  function downloadLabel(payload) {
    const blob = new Blob([`${JSON.stringify(payload, null, 2)}\n`], { type: 'application/json' });
    const a = document.createElement('a');
    a.href = URL.createObjectURL(blob);
    a.download = `${payload.id}.${payload.status}.json`;
    a.click();
    URL.revokeObjectURL(a.href);
  }

  function applyPayload(payload) {
    routeWays = (payload.vectors?.route || []).map(cloneWay);
    contextWays = (payload.vectors?.context || []).map(cloneWay);
    if (!routeWays.length && payload.glyph?.route_bits) {
      // Old pixel-only drafts: keep bits for preview but no vectors.
      routeBits = unpackBits(payload.glyph.route_bits, W, H);
      contextBits = unpackBits(payload.glyph.context_bits, W, H);
      paintBits(layerRoute, routeBits, { color: '#111' });
      paintBits(layerContext, contextBits, { color: '#444', dashed: true });
      paintComposite(layerComposite, routeBits, contextBits);
    } else {
      refreshLayerCanvases();
    }
    labelNotes.value = payload.notes || '';
    draftStroke = [];
    redrawPad();
  }

  document.getElementById('seedGuess').addEventListener('click', () => seedFromGeometry());
  document.getElementById('undoVertex').addEventListener('click', () => {
    draftStroke.pop();
    redrawPad();
  });
  document.getElementById('finishStroke').addEventListener('click', () => finishStroke());
  document.getElementById('deleteLastStroke').addEventListener('click', () => {
    if (drawLayer === 'route') routeWays.pop();
    else contextWays.pop();
    refreshLayerCanvases();
    redrawPad();
  });
  document.getElementById('clearLayer').addEventListener('click', () => {
    if (drawLayer === 'context') contextWays = [];
    else routeWays = [];
    draftStroke = [];
    refreshLayerCanvases();
    redrawPad();
  });
  document.getElementById('clearAll').addEventListener('click', () => {
    routeWays = [];
    contextWays = [];
    draftStroke = [];
    refreshLayerCanvases();
    redrawPad();
  });
  document.getElementById('acceptLabel').addEventListener('click', () => {
    if (!current) return;
    if (draftStroke.length) finishStroke();
    const payload = buildLabelPayload('accepted');
    localStorage.setItem(draftKey(current.id), JSON.stringify(payload));
    downloadLabel(payload);
    statusEl.textContent = `accepted ${current.id} — downloaded + saved draft`;
  });
  document.getElementById('saveDraft').addEventListener('click', () => {
    if (!current) return;
    if (draftStroke.length) finishStroke();
    const payload = buildLabelPayload('draft');
    localStorage.setItem(draftKey(current.id), JSON.stringify(payload));
    statusEl.textContent = `draft saved (${current.id})`;
  });
  document.getElementById('loadDraft').addEventListener('click', () => {
    if (!current) return;
    const raw = localStorage.getItem(draftKey(current.id));
    if (!raw) {
      statusEl.textContent = 'no draft for this fixture';
      return;
    }
    applyPayload(JSON.parse(raw));
    statusEl.textContent = 'loaded draft';
  });

  async function loadFixture(file) {
    const fix = await (await fetch(`${BASE}emulator/minimap-fixtures/${file}`)).json();
    current = fix;
    metaEl.textContent = `${fix.title} · along ${fix.along_m} m`;
    acceptEl.replaceChildren(
      ...(fix.acceptance || []).map((line) => {
        const li = document.createElement('li');
        li.textContent = line;
        return li;
      }),
    );

    const mm = fix.minimap || fix.nav?.minimap;
    viewRadiusM = fix.view_radius_m || minimapViewRadius(mm);
    const tctx = truth.getContext('2d', { alpha: false });
    truth.width = W;
    truth.height = H;
    paintMeterTruth(tctx, mm, W, H, viewRadiusM);

    const paneBytes = api.renderMinimapPNG(JSON.stringify(mm), W, H, 'all');
    if (paneBytes?.ok === false) throw new Error(paneBytes.error || 'renderMinimapPNG failed');
    await paintPng(wasmPane, paneBytes);

    if (fix.nav) {
      const r = api.applyNav(JSON.stringify({ type: 'nav', ...fix.nav }));
      if (r?.ok === false) throw new Error(r.error || 'applyNav failed');
      const hudBytes = api.renderPNG();
      if (hudBytes?.ok === false) throw new Error(hudBytes.error || 'renderPNG failed');
      await paintPng(hud, hudBytes);
    }

    const accepted = await loadAcceptedLabel(fix.id);
    const draft = localStorage.getItem(draftKey(fix.id));
    if (accepted?.vectors) {
      applyPayload(accepted);
      statusEl.textContent = `loaded ${fix.id} (accepted label)`;
    } else if (draft) {
      applyPayload(JSON.parse(draft));
      statusEl.textContent = `loaded ${fix.id} (draft)`;
    } else {
      labelNotes.value = '';
      seedFromGeometry();
      statusEl.textContent = `loaded ${fix.id}`;
    }

    if (fix.turn_lat != null && mode === 'label') {
      ensureMap();
      await loadBakedRoute();
      syncMapFrame();
      redrawPad();
    }

    for (const btn of listEl.querySelectorAll('button')) {
      btn.classList.toggle('active', btn.dataset.file === file);
    }
  }

  listEl.replaceChildren(
    ...index.fixtures.map((f) => {
      const btn = document.createElement('button');
      btn.type = 'button';
      btn.dataset.file = f.file;
      btn.textContent = `${f.title} (${Math.round(f.along_m)} m)`;
      btn.addEventListener('click', () => void loadFixture(f.file));
      return btn;
    }),
  );

  document.getElementById('copyFixture').addEventListener('click', async () => {
    if (!current) return;
    await navigator.clipboard.writeText(JSON.stringify(current, null, 2));
    statusEl.textContent = 'fixture JSON copied';
  });
  document.getElementById('copySvg').addEventListener('click', async () => {
    if (!current) return;
    const mm = current.minimap || current.nav?.minimap;
    const frag = api.minimapSVG(JSON.stringify(mm), W, H);
    await navigator.clipboard.writeText(String(frag));
    statusEl.textContent = 'SVG fragment copied';
  });

  window.addEventListener('resize', () => {
    if (mode === 'label') {
      syncMapFrame();
      redrawPad();
    }
  });

  setMode('compare');
  if (index.fixtures?.[0]) await loadFixture(index.fixtures[0].file);
}

main().catch((err) => {
  console.error(err);
  const statusEl = document.getElementById('status');
  if (statusEl) statusEl.textContent = String(err);
});
