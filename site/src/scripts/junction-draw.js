/**
 * Idealized junction templates from semantic IR (not meter polylines).
 * Mirrors pi/internal/hud/junction.go — keep kinds/fields aligned with protocol/junction.ts.
 */

const STROKE = 2;
const THICK = 3;

function normalizeOutbound(o) {
  switch (o) {
    case 'left':
    case 'right':
    case 'slight_left':
    case 'slight_right':
    case 'straight':
    case 'u_turn':
      return o;
    default:
      return 'straight';
  }
}

function isLeftish(o) {
  return o === 'left' || o === 'slight_left';
}
function isRightish(o) {
  return o === 'right' || o === 'slight_right';
}
function isSlight(o) {
  return o === 'slight_left' || o === 'slight_right';
}

function joinSide(j) {
  if (j.side === 'left' || j.side === 'right') return j.side;
  const ob = normalizeOutbound(j.outbound);
  if (isLeftish(ob)) return 'left';
  if (isRightish(ob)) return 'right';
  return j.drive === 'left' ? 'left' : 'right';
}

function line(ctx, x1, y1, x2, y2, { thick = false, dashed = false } = {}) {
  ctx.save();
  ctx.strokeStyle = '#000';
  ctx.lineWidth = thick ? THICK : STROKE;
  ctx.lineCap = 'square';
  if (dashed) ctx.setLineDash([3, 3]);
  else ctx.setLineDash([]);
  ctx.beginPath();
  ctx.moveTo(x1 + 0.5, y1 + 0.5);
  ctx.lineTo(x2 + 0.5, y2 + 0.5);
  ctx.stroke();
  ctx.restore();
}

function mark(ctx, x, y) {
  ctx.fillStyle = '#000';
  ctx.fillRect(x - 2, y - 2, 4, 4);
}

function circle(ctx, cx, cy, r) {
  ctx.beginPath();
  ctx.strokeStyle = '#000';
  ctx.lineWidth = STROKE;
  ctx.setLineDash([]);
  ctx.arc(cx, cy, r, 0, Math.PI * 2);
  ctx.stroke();
}

function appendSides(ctx, sides, cx, cy, w, h) {
  for (const s of sides || []) {
    let x = cx;
    if (s.side === 'left') x = 8;
    else if (s.side === 'right') x = w - 8;
    else continue;
    let y = cy;
    if (s.at === 'before') y = cy + Math.floor(h / 5);
    else if (s.at === 'after') y = cy - Math.floor(h / 5);
    line(ctx, cx, y, x, y, { dashed: s.style !== 'solid' });
  }
}

function drawSimple(ctx, j, w, h) {
  const cx = Math.floor(w / 2);
  const cy = Math.floor(h / 2);
  const top = 6;
  const bot = h - 6;
  const ob = normalizeOutbound(j.outbound);
  line(ctx, cx, bot, cx, cy, { thick: true });
  if (ob === 'straight' || !ob) {
    line(ctx, cx, cy, cx, top, { thick: true });
  } else if (isLeftish(ob)) {
    if (isSlight(ob)) line(ctx, cx, cy, cx - Math.floor(w / 5), top + 8, { thick: true });
    else {
      line(ctx, cx, cy, cx - Math.floor((w * 2) / 5), cy, { thick: true });
      if (j.through) line(ctx, cx, cy, cx, top, { dashed: true });
    }
  } else if (isRightish(ob)) {
    if (isSlight(ob)) line(ctx, cx, cy, cx + Math.floor(w / 5), top + 8, { thick: true });
    else {
      line(ctx, cx, cy, cx + Math.floor((w * 2) / 5), cy, { thick: true });
      if (j.through) line(ctx, cx, cy, cx, top, { dashed: true });
    }
  } else if (ob === 'u_turn') {
    const ux = j.drive === 'left' ? cx - 14 : cx + 14;
    ctx.beginPath();
    ctx.strokeStyle = '#000';
    ctx.lineWidth = THICK;
    ctx.setLineDash([]);
    ctx.moveTo(cx, cy);
    ctx.lineTo(cx, cy - 10);
    ctx.arcTo(ux, cy - 10, ux, cy + 4, 8);
    ctx.lineTo(ux, cy + 4);
    ctx.stroke();
  }
  appendSides(ctx, j.sides, cx, cy, w, h);
  mark(ctx, cx, cy);
}

function drawTJunction(ctx, j, w, h) {
  const cx = Math.floor(w / 2);
  const cy = Math.floor(h / 2);
  const bot = h - 6;
  const left = 8;
  const right = w - 8;
  const ob = normalizeOutbound(j.outbound);
  line(ctx, cx, bot, cx, cy, { thick: true });
  const goLeft = isLeftish(ob) || (!isRightish(ob) && ob !== 'straight');
  if (goLeft) {
    line(ctx, cx, cy, left, cy, { thick: true });
    line(ctx, cx, cy, right, cy, { dashed: true });
  } else {
    line(ctx, cx, cy, right, cy, { thick: true });
    line(ctx, cx, cy, left, cy, { dashed: true });
  }
  appendSides(ctx, j.sides, cx, cy, w, h);
  mark(ctx, cx, cy);
}

function drawCrossroads(ctx, j, w, h) {
  const cx = Math.floor(w / 2);
  const cy = Math.floor(h / 2);
  const top = 6;
  const bot = h - 6;
  const left = 8;
  const right = w - 8;
  const ob = normalizeOutbound(j.outbound);
  const arm = (x1, y1, x2, y2, ours) => line(ctx, x1, y1, x2, y2, { thick: ours, dashed: !ours });
  arm(cx, bot, cx, cy, true);
  arm(cx, cy, cx, top, ob === 'straight' || !ob);
  arm(cx, cy, left, cy, isLeftish(ob) && !isSlight(ob));
  arm(cx, cy, right, cy, isRightish(ob) && !isSlight(ob));
  if (isSlight(ob) && isLeftish(ob)) arm(cx, cy, cx - Math.floor(w / 5), top + 6, true);
  if (isSlight(ob) && isRightish(ob)) arm(cx, cy, cx + Math.floor(w / 5), top + 6, true);
  appendSides(ctx, j.sides, cx, cy, w, h);
  mark(ctx, cx, cy);
}

function drawFork(ctx, j, w, h) {
  const cx = Math.floor(w / 2);
  const cy = Math.floor(h / 2) + 4;
  const top = 6;
  const bot = h - 6;
  const ob = normalizeOutbound(j.outbound);
  let lx = cx - Math.floor(w / 5);
  let ly = top + 4;
  let rx = cx + Math.floor(w / 5);
  let ry = top + 4;
  if (!isSlight(ob)) {
    lx = cx - Math.floor((w * 2) / 5);
    rx = cx + Math.floor((w * 2) / 5);
    ly = cy - 8;
    ry = cy - 8;
  }
  line(ctx, cx, bot, cx, cy, { thick: true });
  let leftOurs = isLeftish(ob);
  if (!leftOurs && !isRightish(ob)) leftOurs = true;
  if (leftOurs) {
    line(ctx, cx, cy, lx, ly, { thick: true });
    line(ctx, cx, cy, rx, ry, { dashed: true });
  } else {
    line(ctx, cx, cy, rx, ry, { thick: true });
    line(ctx, cx, cy, lx, ly, { dashed: true });
  }
  appendSides(ctx, j.sides, cx, cy, w, h);
  mark(ctx, cx, cy);
}

function drawMerge(ctx, j, w, h) {
  const cx = Math.floor(w / 2);
  const cy = Math.floor(h / 2);
  const top = 6;
  const bot = h - 6;
  const side = joinSide(j);
  line(ctx, cx, bot, cx, top, { thick: true });
  const sx = side === 'right' ? w - 8 : 8;
  line(ctx, sx, bot - 2, cx, cy, { thick: true });
  appendSides(ctx, j.sides, cx, cy, w, h);
  mark(ctx, cx, cy);
}

function drawDual(ctx, j, w, h) {
  const cx = Math.floor(w / 2);
  const cy = Math.floor(h / 2);
  const top = 6;
  const bot = h - 6;
  const sep = 7;
  const driveLeft = j.drive === 'left';
  let ourX = driveLeft ? cx - sep : cx + sep;
  let oppX = driveLeft ? cx + sep : cx - sep;
  const ob = normalizeOutbound(j.outbound);
  if (j.cross_median && (isLeftish(ob) || isRightish(ob))) {
    const gap = 10;
    line(ctx, ourX, bot, ourX, cy + gap, { thick: true });
    line(ctx, oppX, bot, oppX, cy + gap, { dashed: true });
    line(ctx, ourX, cy + gap, oppX, cy - gap, { thick: true });
    if (isLeftish(ob)) {
      line(ctx, oppX, cy - gap, 8, cy - gap, { thick: true });
      if (j.through) line(ctx, oppX, cy - gap, oppX, top, { dashed: true });
    } else {
      line(ctx, oppX, cy - gap, w - 8, cy - gap, { thick: true });
      if (j.through) line(ctx, oppX, cy - gap, oppX, top, { dashed: true });
    }
  } else {
    line(ctx, ourX, bot, ourX, top, { thick: true });
    line(ctx, oppX, bot, oppX, top, { dashed: true });
    if (isLeftish(ob) && !j.cross_median) line(ctx, ourX, cy, 8, cy, { thick: true });
    else if (isRightish(ob) && !j.cross_median) line(ctx, ourX, cy, w - 8, cy, { thick: true });
  }
  appendSides(ctx, j.sides, ourX, cy, w, h);
  mark(ctx, ourX, cy);
}

function drawRoundabout(ctx, j, w, h) {
  const cx = w / 2;
  const cy = h / 2 - 2;
  let r = Math.min(w, h) * 0.28;
  if (r < 10) r = 10;
  let exits = j.exits || 4;
  if (exits < 2) exits = 2;
  if (exits > 6) exits = 6;
  let exit = j.exit || 1;
  if (exit < 1) exit = 1;
  if (exit > exits) exit = exits;
  line(ctx, Math.round(cx), h - 6, Math.round(cx), Math.round(cy + r), { thick: true });
  circle(ctx, cx, cy, r);
  const driveLeft = j.drive === 'left';
  const step = (2 * Math.PI) / exits;
  for (let i = 1; i <= exits; i++) {
    const ang = driveLeft ? Math.PI / 2 - i * step : Math.PI / 2 + i * step;
    const ex = cx + Math.cos(ang) * (r + 14);
    const ey = cy - Math.sin(ang) * (r + 14);
    const ix = cx + Math.cos(ang) * r;
    const iy = cy - Math.sin(ang) * r;
    line(ctx, Math.round(ix), Math.round(iy), Math.round(ex), Math.round(ey), {
      thick: i === exit,
      dashed: i !== exit,
    });
  }
  mark(ctx, Math.round(cx), Math.round(cy + r));
}

function drawRampExit(ctx, j, w, h) {
  const cx = Math.floor(w / 2);
  const cy = Math.floor(h / 2);
  const top = 6;
  const bot = h - 6;
  const ob = normalizeOutbound(j.outbound);
  const side = joinSide(j);
  line(ctx, cx, bot, cx, top, { thick: true });
  let sx = side === 'right' || isRightish(ob) ? w - 8 : 8;
  let sy = top + 10;
  if (isSlight(ob)) {
    sy = top + 4;
    sx = isLeftish(ob) || side === 'left' ? cx - Math.floor(w / 5) : cx + Math.floor(w / 5);
  }
  line(ctx, cx, cy, sx, sy, { thick: true });
  appendSides(ctx, j.sides, cx, cy, w, h);
  mark(ctx, cx, cy);
}

function drawRampEnter(ctx, j, w, h) {
  const cx = Math.floor(w / 2);
  const cy = Math.floor(h / 2);
  const top = 6;
  const bot = h - 6;
  const side = joinSide(j);
  line(ctx, cx, bot, cx, top, { thick: true });
  const sx = side === 'right' ? w - 8 : 8;
  line(ctx, sx, cy + 12, cx, cy, { thick: true });
  appendSides(ctx, j.sides, cx, cy, w, h);
  mark(ctx, cx, cy);
}

function drawUTurn(ctx, j, w, h) {
  const cx = Math.floor(w / 2);
  const cy = Math.floor(h / 2) + 6;
  const bot = h - 6;
  const ux = j.drive === 'left' ? cx - 14 : cx + 14;
  line(ctx, cx, bot, cx, cy, { thick: true });
  ctx.beginPath();
  ctx.strokeStyle = '#000';
  ctx.lineWidth = THICK;
  ctx.setLineDash([]);
  ctx.moveTo(cx, cy);
  ctx.lineTo(cx, cy - 14);
  ctx.arcTo(ux, cy - 14, ux, cy + 8, 8);
  ctx.lineTo(ux, cy + 8);
  ctx.stroke();
  appendSides(ctx, j.sides, cx, cy, w, h);
  mark(ctx, cx, cy);
}

function drawArrive(ctx, j, w, h) {
  const cx = Math.floor(w / 2);
  const bot = h - 6;
  const endY = Math.floor(h / 2) - 4;
  line(ctx, cx, bot, cx, endY, { thick: true });
  circle(ctx, cx, endY, 5);
  mark(ctx, cx, endY);
}

function drawDepart(ctx, j, w, h) {
  const cx = Math.floor(w / 2);
  const top = 6;
  const startY = Math.floor(h / 2) + 10;
  circle(ctx, cx, startY, 5);
  line(ctx, cx, startY, cx, top, { thick: true });
  mark(ctx, cx, startY);
}

/**
 * Paint a junction IR template into a 2d canvas context (clears to white first).
 * @param {CanvasRenderingContext2D} ctx
 * @param {object} junction — protocol JunctionMessage shape
 * @param {number} [w]
 * @param {number} [h]
 */
export function paintJunction(ctx, junction, w = 70, h = 80) {
  ctx.fillStyle = '#fff';
  ctx.fillRect(0, 0, w, h);
  if (!junction || !junction.kind) {
    line(ctx, Math.floor(w / 2), 4, Math.floor(w / 2), h - 4);
    return;
  }
  const j = junction;
  switch (j.kind) {
    case 'simple':
      drawSimple(ctx, j, w, h);
      break;
    case 't_junction':
      drawTJunction(ctx, j, w, h);
      break;
    case 'crossroads':
      drawCrossroads(ctx, j, w, h);
      break;
    case 'fork':
      drawFork(ctx, j, w, h);
      break;
    case 'merge':
      drawMerge(ctx, j, w, h);
      break;
    case 'dual_carriageway':
      drawDual(ctx, j, w, h);
      break;
    case 'roundabout':
      drawRoundabout(ctx, j, w, h);
      break;
    case 'ramp_exit':
      drawRampExit(ctx, j, w, h);
      break;
    case 'ramp_enter':
      drawRampEnter(ctx, j, w, h);
      break;
    case 'u_turn':
      drawUTurn(ctx, j, w, h);
      break;
    case 'arrive':
      drawArrive(ctx, j, w, h);
      break;
    case 'depart':
      drawDepart(ctx, j, w, h);
      break;
    default:
      drawSimple(ctx, { ...j, kind: 'simple' }, w, h);
  }
}

export const IMPLEMENTED_KINDS = [
  'simple',
  't_junction',
  'crossroads',
  'fork',
  'merge',
  'dual_carriageway',
  'roundabout',
  'ramp_exit',
  'ramp_enter',
  'u_turn',
  'arrive',
  'depart',
];
export const TODO_KINDS = [];
