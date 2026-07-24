import React from 'react';

export function ETAReadout(props) {
  if (props.etaMin == null) return null;
  return (
    <div style={{ fontFamily: 'var(--font-pixel)', fontSize: 'var(--text-road)', fontWeight: 400, color: 'var(--ink)', display: 'flex', gap: 4, lineHeight: 1 }}>
      <span>ETA</span>
      <span style={{ fontWeight: 700 }}>{props.etaMin}</span>
      <span>min</span>
    </div>
  );
}
