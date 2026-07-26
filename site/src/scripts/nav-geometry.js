/**
 * Shared route / minimap projection helpers for emulator + fixture capture + lab.
 * Local minimap frame: turn at origin; +Y = inbound approach (rider usually below).
 */

export function haversineM(a, b) {
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

export function pathLength(coords, from = 0, to = coords.length - 1) {
  let m = 0;
  for (let i = from; i < to; i++) m += haversineM(coords[i], coords[i + 1]);
  return m;
}

export function pointAlong(coords, distM) {
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
export function bearingBetween(lng1, lat1, lng2, lat2) {
  const φ1 = (lat1 * Math.PI) / 180;
  const φ2 = (lat2 * Math.PI) / 180;
  const Δλ = ((lng2 - lng1) * Math.PI) / 180;
  const y = Math.sin(Δλ) * Math.cos(φ2);
  const x = Math.cos(φ1) * Math.sin(φ2) - Math.sin(φ1) * Math.cos(φ2) * Math.cos(Δλ);
  return (((Math.atan2(y, x) * 180) / Math.PI) + 360) % 360;
}

export function projectOne(lat0, lon0, bearingDeg, lon, lat) {
  const br = (bearingDeg * Math.PI) / 180;
  const north = (lat - lat0) * 111320;
  const east = (lon - lon0) * 111320 * Math.cos((lat0 * Math.PI) / 180);
  const ahead = north * Math.cos(br) + east * Math.sin(br);
  const right = east * Math.cos(br) - north * Math.sin(br);
  return { x: Math.round(right), y: Math.round(ahead) };
}

export function downsampleRibbon(pts, max = 6) {
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

export const TURN_SNAP_BEHIND = 50;
export const TURN_SNAP_AHEAD = 50;
export const TURN_SNAP_HALF_W = 50;

/**
 * Top-down snapshot centered on the next turn.
 * Local frame: turn at origin; +Y = inbound approach direction (rider below).
 */
export function buildMinimap(coords, riderPos, turnAt, ways) {
  if (!ways || !ways.length) return null;
  const turnIdx = Math.min(Math.max(turnAt, 0), coords.length - 1);
  const turn = coords[turnIdx];
  const turnLng = turn[0];
  const turnLat = turn[1];

  let fromIdx = Math.max(0, turnIdx - 1);
  for (let i = turnIdx - 1; i >= 0; i--) {
    if (haversineM(coords[i], turn) >= 12) {
      fromIdx = i;
      break;
    }
    fromIdx = i;
  }
  const inbound = bearingBetween(coords[fromIdx][0], coords[fromIdx][1], turnLng, turnLat);

  const toLocal = (lng, lat) => projectOne(turnLat, turnLng, inbound, lng, lat);

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
    let near = false;
    for (const [lng, lat] of way) {
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
    for (const [lng, lat] of way) {
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

export function corridorRibbon(coords, origin, bearingDeg, turnAt, distanceM) {
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

export function turnMarkers(route) {
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

/** Nav payload at a given along-m (same contract as emulator nextManeuver). */
export function nextManeuver(route, alongM, roadWays) {
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

/** Draw meter-space geometry into a canvas (truth overlay, no schematize). */
export function paintMeterTruth(ctx, minimap, w, h, radius = 50) {
  const pad = 3;
  const cx = w / 2;
  const cy = h / 2;
  const scale = Math.min((w - pad * 2) / (2 * radius), (h - pad * 2) / (2 * radius));
  const to = (p) => [cx + p.x * scale, cy - p.y * scale];

  ctx.fillStyle = '#fff';
  ctx.fillRect(0, 0, w, h);
  ctx.lineCap = 'square';
  ctx.lineJoin = 'miter';

  ctx.strokeStyle = '#999';
  ctx.lineWidth = 1;
  ctx.setLineDash([2, 2]);
  for (const way of minimap.context || []) {
    if (way.length < 2) continue;
    ctx.beginPath();
    way.forEach((p, i) => {
      const [x, y] = to(p);
      if (i === 0) ctx.moveTo(x, y);
      else ctx.lineTo(x, y);
    });
    ctx.stroke();
  }
  ctx.setLineDash([]);

  ctx.strokeStyle = '#000';
  ctx.lineWidth = 3;
  const route = minimap.route || [];
  if (route.length >= 2) {
    ctx.beginPath();
    route.forEach((p, i) => {
      const [x, y] = to(p);
      if (i === 0) ctx.moveTo(x, y);
      else ctx.lineTo(x, y);
    });
    ctx.stroke();
  }

  ctx.fillStyle = '#000';
  ctx.fillRect(cx - 2, cy - 2, 4, 4);
  if (minimap.rider) {
    const [rx, ry] = to(minimap.rider);
    ctx.fillRect(rx - 2, ry - 2, 5, 5);
  }
}

/** Adaptive view radius mirroring Go minimapViewRadius. */
export function minimapViewRadius(minimap, { min = 25, max = 50 } = {}) {
  let maxR = 0;
  const note = (x, y) => {
    const r = Math.hypot(x, y);
    if (r > maxR) maxR = r;
  };
  for (const p of minimap.route || []) note(p.x, p.y);
  if (minimap.rider) note(minimap.rider.x, minimap.rider.y);
  for (const way of minimap.context || []) {
    for (const p of way) {
      if (Math.hypot(p.x, p.y) <= max) note(p.x, p.y);
    }
  }
  maxR += 6;
  return Math.min(max, Math.max(min, maxR));
}

/** Inverse of projectOne: local meters → lat/lng. */
export function localToLatLng(turnLat, turnLng, bearingDeg, x, y) {
  const br = (bearingDeg * Math.PI) / 180;
  const north = y * Math.cos(br) - x * Math.sin(br);
  const east = y * Math.sin(br) + x * Math.cos(br);
  const lat = turnLat + north / 111320;
  const lng = turnLng + east / (111320 * Math.max(0.2, Math.cos((turnLat * Math.PI) / 180)));
  return { lat, lng };
}

export function pixelScale(w, h, radius, pad = 3) {
  return Math.min((w - pad * 2) / (2 * radius), (h - pad * 2) / (2 * radius));
}

export function localToPixel(x, y, w, h, radius) {
  const cx = w / 2;
  const cy = h / 2;
  const scale = pixelScale(w, h, radius);
  return {
    sx: Math.round(cx + x * scale),
    sy: Math.round(cy - y * scale),
  };
}

export function pixelToLocal(sx, sy, w, h, radius) {
  const cx = w / 2;
  const cy = h / 2;
  const scale = pixelScale(w, h, radius);
  return {
    x: (sx - cx) / scale,
    y: (cy - sy) / scale,
  };
}

/** Pack a Uint8ClampedArray of 0/1 into a compact base64 bitstring (MSB first). */
export function packBits(bits, w, h) {
  const bytes = new Uint8Array(Math.ceil((w * h) / 8));
  for (let i = 0; i < w * h; i++) {
    if (bits[i]) bytes[i >> 3] |= 0x80 >> (i & 7);
  }
  let s = '';
  for (const b of bytes) s += String.fromCharCode(b);
  return btoa(s);
}

export function unpackBits(b64, w, h) {
  const raw = atob(b64);
  const bytes = new Uint8Array(raw.length);
  for (let i = 0; i < raw.length; i++) bytes[i] = raw.charCodeAt(i);
  const bits = new Uint8Array(w * h);
  for (let i = 0; i < w * h; i++) {
    bits[i] = bytes[i >> 3] & (0x80 >> (i & 7)) ? 1 : 0;
  }
  return bits;
}
