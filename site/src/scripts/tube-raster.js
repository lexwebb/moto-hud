/**
 * Labeler schematize + Bresenham raster.
 * Vertices snap to a meter grid; legs are constrained to axis + 45° diagonal
 * on that lattice (no distance-preserving octilinear rotate).
 */

export const DEFAULT_GRID_M = 5;

export function snapToGrid(p, grid = DEFAULT_GRID_M) {
  return {
    x: Math.round(p.x / grid) * grid,
    y: Math.round(p.y / grid) * grid,
  };
}

/** From prev (on-grid), project cursor onto nearest up/down/left/right/diagonal ray. */
export function snapCursor(prev, cursor, grid = DEFAULT_GRID_M) {
  const c = snapToGrid(cursor, grid);
  if (!prev) return c;
  const p = snapToGrid(prev, grid);
  const dx = c.x - p.x;
  const dy = c.y - p.y;
  if (dx === 0 && dy === 0) return p;

  const adx = Math.abs(dx);
  const ady = Math.abs(dy);
  const sx = Math.sign(dx) || 0;
  const sy = Math.sign(dy) || 0;

  // Already axis or diagonal on the grid.
  if (adx === 0 || ady === 0 || adx === ady) return c;

  // Pick nearest of: horizontal, vertical, or 45° diagonal (equal steps).
  const candidates = [
    { x: p.x + dx, y: p.y }, // horizontal
    { x: p.x, y: p.y + dy }, // vertical
  ];
  if (sx && sy) {
    const n = Math.round(Math.min(adx, ady) / grid) * grid;
    const nMax = Math.round(Math.max(adx, ady) / grid) * grid;
    candidates.push({ x: p.x + sx * n, y: p.y + sy * n });
    candidates.push({ x: p.x + sx * nMax, y: p.y + sy * nMax });
  }

  let best = candidates[0];
  let bestD = Infinity;
  for (const q of candidates) {
    const d = (q.x - c.x) ** 2 + (q.y - c.y) ** 2;
    if (d < bestD) {
      bestD = d;
      best = q;
    }
  }
  return snapToGrid(best, grid);
}

/** Insert axis/diagonal knees so every leg is H, V, or 45° on the grid. */
export function gridKnee(a, b, grid = DEFAULT_GRID_M) {
  const a0 = snapToGrid(a, grid);
  const b0 = snapToGrid(b, grid);
  const dx = b0.x - a0.x;
  const dy = b0.y - a0.y;
  const adx = Math.abs(dx);
  const ady = Math.abs(dy);
  if (adx === 0 || ady === 0 || adx === ady) return [b0];
  const sx = Math.sign(dx);
  const sy = Math.sign(dy);
  if (adx > ady) return [{ x: a0.x + sx * ady, y: a0.y + sy * ady }, b0];
  return [{ x: a0.x + sx * adx, y: a0.y + sy * adx }, b0];
}

export function schematizeGrid(pts, grid = DEFAULT_GRID_M) {
  if (!pts || pts.length < 1) return [];
  const out = [snapToGrid(pts[0], grid)];
  for (let i = 1; i < pts.length; i++) {
    const prev = out[out.length - 1];
    for (const p of gridKnee(prev, pts[i], grid)) {
      if (p.x === out[out.length - 1].x && p.y === out[out.length - 1].y) continue;
      out.push(p);
    }
  }
  return out;
}

/** @deprecated use schematizeGrid — kept as alias for call sites. */
export function schematizeTube(pts, _eps = 5, grid = DEFAULT_GRID_M) {
  return schematizeGrid(pts, grid);
}

function absInt(v) {
  return v < 0 ? -v : v;
}
function signInt(v) {
  return v < 0 ? -1 : v > 0 ? 1 : 0;
}

export function tubeKnee(a, b) {
  const dx = b[0] - a[0];
  const dy = b[1] - a[1];
  const adx = absInt(dx);
  const ady = absInt(dy);
  if (adx === 0 || ady === 0 || adx === ady) return [b];
  const sx = signInt(dx);
  const sy = signInt(dy);
  if (adx > ady) return [[a[0] + sx * ady, a[1] + sy * ady], b];
  return [[a[0] + sx * adx, a[1] + sy * adx], b];
}

export function bresenham(x0, y0, x1, y1) {
  let dx = absInt(x1 - x0);
  let dy = -absInt(y1 - y0);
  const sx = signInt(x1 - x0);
  const sy = signInt(y1 - y0);
  let err = dx + dy;
  const out = [];
  for (;;) {
    out.push([x0, y0]);
    if (x0 === x1 && y0 === y1) break;
    const e2 = 2 * err;
    if (e2 >= dy) {
      err += dy;
      x0 += sx;
    }
    if (e2 <= dx) {
      err += dx;
      y0 += sy;
    }
  }
  return out;
}

function tubePixels(pts, project) {
  const raw = [];
  for (const p of pts) {
    const pr = project(p.x, p.y);
    if (!pr) continue;
    if (raw.length && raw[raw.length - 1][0] === pr[0] && raw[raw.length - 1][1] === pr[1]) continue;
    raw.push(pr);
  }
  if (raw.length < 2) return raw;
  const out = [raw[0]];
  for (let i = 1; i < raw.length; i++) {
    const a = out[out.length - 1];
    for (const p of tubeKnee(a, raw[i])) {
      if (out[out.length - 1][0] === p[0] && out[out.length - 1][1] === p[1]) continue;
      out.push(p);
    }
  }
  return out;
}

function paintDashed(bits, w, h, pix, dash = 3, gap = 3) {
  if (pix.length < 2) return;
  let on = true;
  let left = dash;
  for (let i = 0; i < pix.length - 1; i++) {
    const line = bresenham(pix[i][0], pix[i][1], pix[i + 1][0], pix[i + 1][1]);
    const start = i > 0 && line.length ? 1 : 0;
    for (const [x, y] of line.slice(start)) {
      if (on && x >= 0 && y >= 0 && x < w && y < h) bits[y * w + x] = 1;
      left--;
      if (left <= 0) {
        on = !on;
        left = on ? dash : gap;
      }
    }
  }
}

function paintSolid(bits, w, h, pix, thickness = 3) {
  if (pix.length < 2) return;
  const r = Math.floor(thickness / 2);
  const seen = new Set();
  const paint = (x, y) => {
    if (x < 0 || y < 0 || x >= w || y >= h) return;
    const key = `${x},${y}`;
    if (seen.has(key)) return;
    seen.add(key);
    bits[y * w + x] = 1;
  };
  for (let i = 0; i < pix.length - 1; i++) {
    const line = bresenham(pix[i][0], pix[i][1], pix[i + 1][0], pix[i + 1][1]);
    const start = i > 0 && line.length ? 1 : 0;
    for (const [x, y] of line.slice(start)) {
      for (let ox = -r; ox <= r; ox++) {
        for (let oy = -r; oy <= r; oy++) {
          if (absInt(ox) + absInt(oy) <= r) paint(x + ox, y + oy);
        }
      }
    }
  }
}

export function viewRadius(minimap, { min = 25, max = 50 } = {}) {
  let maxR = 0;
  const note = (x, y) => {
    const r = Math.hypot(x, y);
    if (r > maxR) maxR = r;
  };
  for (const p of minimap?.route || []) note(p.x, p.y);
  if (minimap?.rider) note(minimap.rider.x, minimap.rider.y);
  for (const way of minimap?.context || []) {
    for (const p of way) {
      if (Math.hypot(p.x, p.y) <= max) note(p.x, p.y);
    }
  }
  maxR += 6;
  return Math.min(max, Math.max(min, maxR || min));
}

/**
 * Rasterize labeled vector ways on a meter grid, then Bresenham to pixels.
 */
export function rasterizeVectors({
  routeWays = [],
  contextWays = [],
  rider = null,
  w = 70,
  h = 80,
  radius = null,
  grid = DEFAULT_GRID_M,
} = {}) {
  const probe = {
    route: routeWays[0] || [
      { x: 0, y: -40 },
      { x: 0, y: 0 },
    ],
    context: contextWays,
    rider,
  };
  const R = radius ?? viewRadius(probe);
  const pad = 3;
  const cx = w / 2;
  const cy = h / 2;
  const scale = Math.min((w - pad * 2) / (2 * R), (h - pad * 2) / (2 * R));
  const project = (x, y) => {
    if (Math.hypot(x, y) > R * 1.2) return null;
    const sx = Math.round(cx + x * scale);
    const sy = Math.round(cy - y * scale);
    if (sx < -1 || sx > w || sy < -1 || sy > h) return null;
    return [sx, sy];
  };

  const routeBits = new Uint8Array(w * h);
  const contextBits = new Uint8Array(w * h);

  for (const way of contextWays) {
    if (!way || way.length < 2) continue;
    const pix = tubePixels(schematizeGrid(way, grid), project);
    paintDashed(contextBits, w, h, pix, 3, 3);
  }
  for (const way of routeWays) {
    if (!way || way.length < 2) continue;
    const pix = tubePixels(schematizeGrid(way, grid), project);
    paintSolid(routeBits, w, h, pix, 3);
  }

  if (routeWays.some((way) => way && way.length >= 2)) {
    for (let dy = -2; dy <= 1; dy++) {
      for (let dx = -2; dx <= 1; dx++) {
        const x = Math.round(cx) + dx;
        const y = Math.round(cy) + dy;
        if (x >= 0 && y >= 0 && x < w && y < h) routeBits[y * w + x] = 1;
      }
    }
  }
  if (rider) {
    const pr = project(rider.x, rider.y);
    if (pr) {
      for (let dy = -2; dy <= 2; dy++) {
        for (let dx = -2; dx <= 2; dx++) {
          const x = pr[0] + dx;
          const y = pr[1] + dy;
          if (x >= 0 && y >= 0 && x < w && y < h) routeBits[y * w + x] = 1;
        }
      }
    }
  }

  for (let i = 0; i < w * h; i++) {
    if (routeBits[i]) contextBits[i] = 0;
  }
  return { routeBits, contextBits, radius: R, grid };
}
