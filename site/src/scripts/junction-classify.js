/**
 * POC: classify a turn-centered scene into semantic junction IR.
 * Prefers OSM tags (dual_carriageway) when the bake has them; geometry is fallback.
 * Local frame: turn at origin; +Y = inbound approach (same as buildMinimap).
 */

import {
  haversineM,
  bearingBetween,
  projectOne,
  wayCoords,
  TURN_SNAP_BEHIND,
  TURN_SNAP_AHEAD,
  TURN_SNAP_HALF_W,
} from './nav-geometry.js';

const ARM_NEAR_M = 14;
const SIDE_BEFORE_Y = [-40, -12];
const SIDE_AT_Y = [-12, 12];
const SIDE_AFTER_Y = [12, 40];
/** Stricter than v1 POC — reduces false duals from alley parallels. */
const PARALLEL_MAX_ANGLE = 18;
const PARALLEL_SEP_MIN = 10;
const PARALLEL_SEP_MAX = 22;
const PARALLEL_MIN_LEN = 12;
const OPPOSITE_HEADING_MIN = 150;

function normAngle(deg) {
  return ((deg % 360) + 360) % 360;
}

function angleDelta(a, b) {
  let d = normAngle(a) - normAngle(b);
  if (d > 180) d -= 360;
  if (d < -180) d += 360;
  return d;
}

function outboundFromTurnDeg(turnDeg) {
  const a = Math.abs(turnDeg);
  if (a < 20) return 'straight';
  if (a > 150) return 'u_turn';
  if (turnDeg > 0) {
    if (a < 50) return 'slight_right';
    return 'right';
  }
  if (a < 50) return 'slight_left';
  return 'left';
}

function maneuverToOutbound(m) {
  switch (m) {
    case 'left':
    case 'right':
    case 'slight_left':
    case 'slight_right':
    case 'straight':
    case 'u_turn':
      return m;
    case 'depart':
    case 'arrive':
      return 'straight';
    case 'roundabout':
      return 'right';
    default:
      return 'straight';
  }
}

function farPoint(coords, turnIdx, dir, minDistM) {
  const turn = coords[turnIdx];
  if (dir < 0) {
    for (let i = turnIdx - 1; i >= 0; i--) {
      if (haversineM(coords[i], turn) >= minDistM) return coords[i];
    }
    return coords[0];
  }
  for (let i = turnIdx + 1; i < coords.length; i++) {
    if (haversineM(coords[i], turn) >= minDistM) return coords[i];
  }
  return coords[coords.length - 1];
}

function waySegmentsLocal(way, toLocal) {
  const geom = wayCoords(way);
  if (!geom || geom.length < 2) return [];
  const meta = Array.isArray(way)
    ? { dual_carriageway: false, oneway: null, highway: null }
    : {
        dual_carriageway: !!way.dual_carriageway,
        oneway: way.oneway || null,
        highway: way.highway || null,
        name: way.name || null,
      };
  const segs = [];
  for (let i = 0; i < geom.length - 1; i++) {
    const a = toLocal(geom[i][0], geom[i][1]);
    const b = toLocal(geom[i + 1][0], geom[i + 1][1]);
    const inBox = (p) =>
      p.y >= -TURN_SNAP_BEHIND && p.y <= TURN_SNAP_AHEAD && Math.abs(p.x) <= TURN_SNAP_HALF_W;
    if (!inBox(a) && !inBox(b)) continue;
    const mid = { x: (a.x + b.x) / 2, y: (a.y + b.y) / 2 };
    const dx = b.x - a.x;
    const dy = b.y - a.y;
    const len = Math.hypot(dx, dy);
    if (len < 2) continue;
    const heading = (Math.atan2(dx, dy) * 180) / Math.PI;
    segs.push({ a, b, mid, heading, len, ...meta });
  }
  return segs;
}

function bandForY(y) {
  if (y >= SIDE_BEFORE_Y[0] && y < SIDE_BEFORE_Y[1]) return 'before';
  if (y >= SIDE_AT_Y[0] && y <= SIDE_AT_Y[1]) return 'at';
  if (y > SIDE_AFTER_Y[0] && y <= SIDE_AFTER_Y[1]) return 'after';
  return null;
}

function classifySides(segs, routeOutbound) {
  const sides = [];
  const seen = new Set();

  for (const s of segs) {
    const nearOrigin = Math.hypot(s.mid.x, s.mid.y) <= ARM_NEAR_M;
    const crossesSpine =
      Math.min(Math.abs(s.a.x), Math.abs(s.b.x), Math.abs(s.mid.x)) <= 10 &&
      Math.abs(s.heading) > 35 &&
      Math.abs(Math.abs(s.heading) - 180) > 35;

    if (!nearOrigin && !crossesSpine) continue;

    const side = s.mid.x >= 0 ? 'right' : 'left';
    const far = Math.abs(s.a.x) > Math.abs(s.b.x) ? s.a : s.b;
    const side2 = far.x >= 0 ? 'right' : 'left';
    const useSide = Math.abs(far.x) > 8 ? side2 : side;

    const outMatch =
      (routeOutbound === 'right' || routeOutbound === 'slight_right') && useSide === 'right';
    const outMatchL =
      (routeOutbound === 'left' || routeOutbound === 'slight_left') && useSide === 'left';
    if ((outMatch || outMatchL) && nearOrigin && Math.hypot(s.mid.x, s.mid.y) < 8) {
      continue;
    }

    if (Math.abs(s.heading) < 25 || Math.abs(Math.abs(s.heading) - 180) < 25) {
      if (Math.abs(s.mid.x) < 8) continue;
    }

    const at = bandForY(s.mid.y) || (nearOrigin ? 'at' : null);
    if (!at) continue;

    const key = `${useSide}:${at}`;
    if (seen.has(key)) continue;
    seen.add(key);
    sides.push({ side: useSide, at, style: 'dashed' });
  }

  return sides;
}

/** Tag path: any dual_carriageway=yes segment on approach / at junction. */
function detectDualFromTags(segs) {
  const hit = segs.find(
    (s) =>
      s.dual_carriageway &&
      s.mid.y >= -45 &&
      s.mid.y <= 20 &&
      Math.abs(s.mid.x) <= 30,
  );
  return !!hit;
}

/**
 * Geometry / tag fallback for dual carriageway.
 * Aligns with OsmAnd POC rules:
 * 1) dual_carriageway tag (handled separately)
 * 2) opposite oneway + parallel sep in [10,22] on major highways
 * 3) opposite heading + parallel sep (never same-direction)
 */
function isMajorHighway(hwy) {
  if (!hwy) return false;
  return /^(motorway|trunk|primary|secondary)(_link)?$/.test(hwy);
}

function onewaySign(v) {
  if (v === true || v === 'yes' || v === 1 || v === '1') return 1;
  if (v === -1 || v === '-1' || v === 'reverse') return -1;
  if (v === false || v === 'no' || v === 0 || v === '0' || v == null) return 0;
  return 0;
}

function isCarHighway(hwy) {
  if (!hwy) return true; // bare geometry: allow
  return /^(motorway|trunk|primary|secondary|tertiary|unclassified|residential)(_link)?$/.test(hwy);
}

function detectDualFromGeometry(segs) {
  const approach = segs.filter(
    (s) =>
      s.mid.y < -5 &&
      s.mid.y > -45 &&
      s.len >= PARALLEL_MIN_LEN &&
      isCarHighway(s.highway) &&
      (Math.abs(s.heading) < PARALLEL_MAX_ANGLE ||
        Math.abs(Math.abs(s.heading) - 180) < PARALLEL_MAX_ANGLE),
  );
  if (approach.length < 2) return false;

  for (let i = 0; i < approach.length; i++) {
    for (let j = i + 1; j < approach.length; j++) {
      const a = approach[i];
      const b = approach[j];
      const sep = Math.abs(a.mid.x - b.mid.x);
      if (sep < PARALLEL_SEP_MIN || sep > PARALLEL_SEP_MAX) continue;

      const headDiff = Math.abs(angleDelta(a.heading, b.heading));
      const oppositeHeading = headDiff >= OPPOSITE_HEADING_MIN;
      if (!oppositeHeading) continue;

      const oa = onewaySign(a.oneway);
      const ob = onewaySign(b.oneway);
      // Same-direction oneway parallels (Fleet Street) — both oneway and NOT opposite heading already excluded;
      // still reject if headings are parallel (same dir) — handled above.
      // Prefer both-oneway opposite travel; else opposite heading on majors/same name.
      const bothOneway = oa !== 0 && ob !== 0;
      const eitherOneway = oa !== 0 || ob !== 0;
      const named =
        a.name && b.name && String(a.name).toLowerCase() === String(b.name).toLowerCase();
      const majors = isMajorHighway(a.highway) && isMajorHighway(b.highway);

      // Both-bidirectional parallels are too weak without a oneway signal.
      if (!eitherOneway) continue;
      if (bothOneway && (named || majors)) return true;
      if (oppositeHeading && (named || majors || !a.highway)) return true;
    }
  }
  return false;
}

function countCardinalArms(segs, outbound) {
  let left = false;
  let right = false;
  let through = false;
  for (const s of segs) {
    if (Math.hypot(s.mid.x, s.mid.y) > ARM_NEAR_M + 6) continue;
    if (s.mid.y > 8 && Math.abs(s.mid.x) < 12) through = true;
    if (s.mid.x < -8 && Math.abs(s.mid.y) < 14) left = true;
    if (s.mid.x > 8 && Math.abs(s.mid.y) < 14) right = true;
  }
  if (outbound === 'left' || outbound === 'slight_left') left = true;
  if (outbound === 'right' || outbound === 'slight_right') right = true;
  if (outbound === 'straight') through = true;
  return { left, right, through };
}

/**
 * @param {object} opts
 * @param {number[][]} opts.coords
 * @param {number} opts.turnAt
 * @param {Array} opts.ways tagged or bare ways
 * @param {string} [opts.maneuver]
 * @param {string} [opts.road]
 * @param {'left'|'right'} [opts.drive]
 */
export function classifyJunction({ coords, turnAt, ways, maneuver, road, drive = 'left' }) {
  const turnIdx = Math.min(Math.max(turnAt, 0), coords.length - 1);
  const turn = coords[turnIdx];
  const approachPt = farPoint(coords, turnIdx, -1, 18);
  const departPt = farPoint(coords, turnIdx, 1, 18);
  const inbound = bearingBetween(approachPt[0], approachPt[1], turn[0], turn[1]);
  const outboundBrg = bearingBetween(turn[0], turn[1], departPt[0], departPt[1]);
  const turnDeg = angleDelta(outboundBrg, inbound);

  const geomOutbound = outboundFromTurnDeg(turnDeg);
  const labeledOutbound = maneuver ? maneuverToOutbound(maneuver) : null;
  let outbound = geomOutbound;
  if (labeledOutbound) {
    if (
      maneuver === 'slight_left' ||
      maneuver === 'slight_right' ||
      maneuver === 'straight' ||
      maneuver === 'arrive' ||
      maneuver === 'depart' ||
      maneuver === 'u_turn' ||
      maneuver === 'roundabout'
    ) {
      outbound = labeledOutbound;
    } else if (
      (labeledOutbound === 'left' || labeledOutbound === 'right') &&
      (geomOutbound === 'left' ||
        geomOutbound === 'right' ||
        geomOutbound === 'slight_left' ||
        geomOutbound === 'slight_right')
    ) {
      const labSide = labeledOutbound.includes('left') ? 'left' : 'right';
      const geoSide = geomOutbound.includes('left') ? 'left' : 'right';
      outbound = labSide === geoSide ? labeledOutbound : geomOutbound;
    }
  }

  const toLocal = (lng, lat) => projectOne(turn[1], turn[0], inbound, lng, lat);
  const segs = [];
  for (const way of ways || []) {
    segs.push(...waySegmentsLocal(way, toLocal));
  }

  const sides = classifySides(segs, outbound);
  const arms = countCardinalArms(segs, outbound);
  const dualTag = detectDualFromTags(segs);
  const dualGeom = !dualTag && detectDualFromGeometry(segs);
  const dual = dualTag || dualGeom;
  const dualSource = dualTag ? 'tag' : dualGeom ? 'geometry' : null;

  let kind = 'simple';
  let through = arms.through || outbound === 'straight';
  let cross_median = false;
  let exits;
  let exit;

  if (maneuver === 'arrive') {
    kind = 'arrive';
    through = false;
  } else if (maneuver === 'depart') {
    kind = 'depart';
    through = true;
  } else if (maneuver === 'roundabout') {
    kind = 'roundabout';
    exits = 4;
    exit = 2;
  } else if (maneuver === 'u_turn' || outbound === 'u_turn') {
    kind = 'u_turn';
    through = false;
  } else if (
    maneuver === 'slight_left' ||
    maneuver === 'slight_right' ||
    outbound.startsWith('slight_')
  ) {
    kind = 'fork';
    through = false;
  } else if (dual) {
    kind = 'dual_carriageway';
    through = outbound === 'straight' || arms.through;
    cross_median =
      (outbound === 'left' || outbound === 'right') && !outbound.startsWith('slight');
  } else if (arms.left && arms.right && arms.through) {
    kind = 'crossroads';
    through = true;
  } else if ((arms.left || arms.right) && !arms.through && outbound !== 'straight') {
    kind = 't_junction';
    through = false;
  } else {
    kind = 'simple';
    through = arms.through || outbound === 'straight';
  }

  const junction = {
    kind,
    drive,
    outbound,
    through,
  };
  if (sides.length) junction.sides = sides;
  if (kind === 'dual_carriageway') junction.cross_median = cross_median;
  if (kind === 'roundabout') {
    junction.exits = exits;
    junction.exit = exit;
  }

  return {
    junction,
    debug: {
      road,
      maneuver_label: maneuver,
      turn_deg: Math.round(turnDeg * 10) / 10,
      geom_outbound: geomOutbound,
      inbound_bearing_deg: Math.round(inbound * 10) / 10,
      outbound_bearing_deg: Math.round(outboundBrg * 10) / 10,
      arms,
      dual_detected: dual,
      dual_source: dualSource,
      context_segs: segs.length,
      sides_raw: sides.length,
    },
  };
}
