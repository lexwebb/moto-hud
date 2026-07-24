import React from 'react';

export function MediaLine(props) {
  const playing = !!props.playing;
  return (
    <div style={{ fontFamily: 'var(--font-pixel)', color: 'var(--ink)', display: 'flex', alignItems: 'center', gap: 6 }}>
      <svg className="pixel-crisp" width="10" height="10" viewBox="0 0 10 10" style={{ flexShrink: 0 }}>
        {playing
          ? <polygon points="1,1 9,5 1,9" fill="var(--ink)" />
          : <React.Fragment><rect x="1" y="1" width="3" height="8" fill="var(--ink)" /><rect x="6" y="1" width="3" height="8" fill="var(--ink)" /></React.Fragment>}
      </svg>
      <div style={{ overflow: 'hidden' }}>
        <div style={{ fontSize: 'var(--text-road)', fontWeight: 700, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis', lineHeight: 1 }}>{props.title || '—'}</div>
        <div style={{ fontSize: 'var(--text-meta)', fontWeight: 400, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis', lineHeight: 1 }}>{props.artist || ''}</div>
      </div>
    </div>
  );
}
