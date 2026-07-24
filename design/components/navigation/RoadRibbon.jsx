import React from 'react';

export function RoadRibbon(props) {
  const w = props.width || 220;
  const h = props.height || 40;
  const points = props.points;
  if (!points || points.length < 2) {
    return (
      <svg className="pixel-crisp" width={w} height={h} viewBox={`0 0 ${w} ${h}`}>
        <line x1={w / 2} y1="4" x2={w / 2} y2={h - 4} stroke="var(--ink)" strokeWidth="2" strokeDasharray="4 5" />
      </svg>
    );
  }
  const xs = points.map((p) => p.x);
  const ys = points.map((p) => p.y);
  const minX = Math.min(...xs), maxX = Math.max(...xs);
  const minY = Math.min(...ys), maxY = Math.max(...ys);
  const pad = 5;
  const rangeX = maxX - minX || 1;
  const rangeY = maxY - minY || 1;
  const scale = Math.min((w - pad * 2) / rangeX, (h - pad * 2) / rangeY);
  const plottedW = rangeX * scale;
  const offsetX = pad + ((w - pad * 2) - plottedW) / 2;
  const screenBottom = h - pad;
  const sx = (x) => offsetX + (x - minX) * scale;
  const sy = (y) => screenBottom - (y - minY) * scale;
  const d = points.map((p, i) => `${i === 0 ? 'M' : 'L'}${sx(p.x)},${sy(p.y)}`).join(' ');
  const turn = props.turnIndex != null ? points[props.turnIndex] : null;
  return (
    <svg className="pixel-crisp" width={w} height={h} viewBox={`0 0 ${w} ${h}`}>
      <path d={d} fill="none" stroke="var(--ink)" strokeWidth="3" strokeLinejoin="miter" strokeLinecap="square" />
      {turn && <rect x={sx(turn.x) - 3} y={sy(turn.y) - 3} width="6" height="6" fill="var(--ink)" />}
    </svg>
  );
}
