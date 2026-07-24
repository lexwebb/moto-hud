import React from 'react';

export function DistanceReadout(props) {
  const value = props.value ?? '—';
  const unit = props.unit || '';
  const size = props.size || 'hero';
  const fontSize = size === 'hero' ? 'var(--text-hero)' : 'var(--text-eta)';
  return (
    <div style={{ display: 'flex', alignItems: 'baseline', gap: 4, fontFamily: 'var(--font-pixel)', color: 'var(--ink)', lineHeight: 'var(--lh-tight)' }}>
      <span style={{ fontSize, fontWeight: 700 }}>{value}</span>
      {unit && <span style={{ fontSize: 'var(--text-road)', fontWeight: 700 }}>{unit}</span>}
    </div>
  );
}
