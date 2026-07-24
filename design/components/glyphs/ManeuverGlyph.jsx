import React from 'react';

const HEAD = 7;

function arrowHead(x, y, angleDeg) {
  const a = (angleDeg * Math.PI) / 180;
  const tip = [x, y];
  const back = [x - Math.cos(a) * HEAD, y - Math.sin(a) * HEAD];
  const p1 = [back[0] + Math.cos(a + Math.PI / 2) * HEAD * 0.62, back[1] + Math.sin(a + Math.PI / 2) * HEAD * 0.62];
  const p2 = [back[0] - Math.cos(a + Math.PI / 2) * HEAD * 0.62, back[1] - Math.sin(a + Math.PI / 2) * HEAD * 0.62];
  return `${tip[0]},${tip[1]} ${p1[0]},${p1[1]} ${p2[0]},${p2[1]}`;
}

function glyphBody(type) {
  const stem = <line x1="20" y1="34" x2="20" y2="14" stroke="var(--ink)" strokeWidth="3" strokeLinecap="square" />;
  switch (type) {
    case 'straight':
      return <React.Fragment>{stem}<polygon points={arrowHead(20, 7, -90)} fill="var(--ink)" /></React.Fragment>;
    case 'left':
      return <React.Fragment>
        <line x1="20" y1="34" x2="20" y2="17" stroke="var(--ink)" strokeWidth="3" strokeLinecap="square" />
        <line x1="20" y1="17" x2="8" y2="17" stroke="var(--ink)" strokeWidth="3" strokeLinecap="square" />
        <polygon points={arrowHead(1, 17, 180)} fill="var(--ink)" />
      </React.Fragment>;
    case 'right':
      return <React.Fragment>
        <line x1="20" y1="34" x2="20" y2="17" stroke="var(--ink)" strokeWidth="3" strokeLinecap="square" />
        <line x1="20" y1="17" x2="32" y2="17" stroke="var(--ink)" strokeWidth="3" strokeLinecap="square" />
        <polygon points={arrowHead(39, 17, 0)} fill="var(--ink)" />
      </React.Fragment>;
    case 'slight-left':
      return <React.Fragment>
        <line x1="21" y1="34" x2="21" y2="24" stroke="var(--ink)" strokeWidth="3" strokeLinecap="square" />
        <line x1="21" y1="24" x2="11" y2="9" stroke="var(--ink)" strokeWidth="3" strokeLinecap="square" />
        <polygon points={arrowHead(7, 4, 214)} fill="var(--ink)" />
      </React.Fragment>;
    case 'slight-right':
      return <React.Fragment>
        <line x1="19" y1="34" x2="19" y2="24" stroke="var(--ink)" strokeWidth="3" strokeLinecap="square" />
        <line x1="19" y1="24" x2="29" y2="9" stroke="var(--ink)" strokeWidth="3" strokeLinecap="square" />
        <polygon points={arrowHead(33, 4, -34)} fill="var(--ink)" />
      </React.Fragment>;
    case 'u-turn':
      return <React.Fragment>
        <path d="M14,34 V16 A6,6 0 0 1 26,16 V26" fill="none" stroke="var(--ink)" strokeWidth="3" strokeLinecap="square" />
        <polygon points={arrowHead(26, 33, 90)} fill="var(--ink)" />
      </React.Fragment>;
    case 'roundabout':
      return <React.Fragment>
        <line x1="20" y1="34" x2="20" y2="27" stroke="var(--ink)" strokeWidth="3" strokeLinecap="square" />
        <circle cx="20" cy="17" r="9" fill="none" stroke="var(--ink)" strokeWidth="3" />
        <polygon points={arrowHead(29, 8, 0)} fill="var(--ink)" />
      </React.Fragment>;
    case 'arrive':
      return <React.Fragment>
        <line x1="13" y1="34" x2="13" y2="7" stroke="var(--ink)" strokeWidth="3" strokeLinecap="square" />
        <polygon points="13,7 29,11 13,17" fill="var(--ink)" />
      </React.Fragment>;
    case 'depart':
      return <React.Fragment>
        <circle cx="20" cy="27" r="4" fill="var(--ink)" />
        <line x1="20" y1="22" x2="20" y2="8" stroke="var(--ink)" strokeWidth="3" strokeLinecap="square" />
        <polygon points={arrowHead(20, 4, -90)} fill="var(--ink)" />
      </React.Fragment>;
    default:
      return <text x="20" y="29" fontFamily="var(--font-pixel)" fontWeight="700" fontSize="24" textAnchor="middle" fill="var(--ink)">?</text>;
  }
}

export function ManeuverGlyph(props) {
  const size = props.size || 40;
  return (
    <svg className="pixel-crisp" width={size} height={size} viewBox="0 0 40 40" style={{ display: 'block' }}>
      {glyphBody(props.type)}
    </svg>
  );
}
