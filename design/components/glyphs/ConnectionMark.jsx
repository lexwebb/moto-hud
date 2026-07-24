import React from 'react';

export function ConnectionMark(props) {
  const connected = !!props.connected;
  const heartbeat = !!props.heartbeat;
  const size = props.size || 12;
  return (
    <span style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
      {connected ? (
        <svg className="pixel-crisp" width={size} height={size} viewBox="0 0 12 12">
          <path d="M2,9 L6,2 L6,6 L10,6 L6,10 L6,6" fill="none" stroke="var(--ink)" strokeWidth="1.6" strokeLinejoin="miter" />
        </svg>
      ) : (
        <svg className="pixel-crisp" width={size} height={size} viewBox="0 0 12 12">
          <line x1="2" y1="2" x2="10" y2="10" stroke="var(--ink)" strokeWidth="1.6" />
          <line x1="10" y1="2" x2="2" y2="10" stroke="var(--ink)" strokeWidth="1.6" />
        </svg>
      )}
      {connected && (
        <span
          className="pixel-crisp"
          style={{
            width: 4,
            height: 4,
            background: 'var(--ink)',
            display: 'inline-block',
            animation: heartbeat ? 'moto-hb 1.6s steps(1) infinite' : 'none',
          }}
        />
      )}
      <style>{'@keyframes moto-hb{0%,49%{opacity:1}50%,100%{opacity:0}}'}</style>
    </span>
  );
}
