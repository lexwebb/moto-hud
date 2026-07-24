import React from 'react';

export function ProgressTicks(props) {
  const total = props.total || 5;
  const filled = Math.max(0, Math.min(total, props.filled || 0));
  const ticks = [];
  for (let i = 0; i < total; i++) ticks.push(i < filled);
  return (
    <div style={{ display: 'flex', gap: 3 }}>
      {ticks.map((on, i) => (
        <span key={i} style={{ width: 8, height: 10, background: on ? 'var(--ink)' : 'var(--paper)', border: '1.5px solid var(--ink)', boxSizing: 'border-box' }} />
      ))}
    </div>
  );
}
