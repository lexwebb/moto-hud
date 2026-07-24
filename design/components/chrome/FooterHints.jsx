import React from 'react';

export function FooterHints(props) {
  const hints = props.hints || [];
  return (
    <div style={{ display: 'flex', gap: 10, fontFamily: 'var(--font-pixel)', fontSize: 'var(--text-meta)', color: 'var(--ink)', lineHeight: 1, justifyContent: 'flex-end', textAlign: 'right' }}>
      {hints.map((h, i) => (
        <span key={i} style={{ display: 'flex', gap: 3 }}>
          <span style={{ fontWeight: 700 }}>{h.btn}</span>
          <span style={{ fontWeight: 400 }}>{h.label}</span>
        </span>
      ))}
    </div>
  );
}
