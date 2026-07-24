import React from 'react';
import { ConnectionMark } from '../glyphs/ConnectionMark.jsx';

export function ModeHeader(props) {
  return (
    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontFamily: 'var(--font-pixel)', color: 'var(--ink)', lineHeight: 1 }}>
      <span style={{ fontSize: 'var(--text-meta)', fontWeight: 700, letterSpacing: 1 }}>{props.mode}</span>
      <ConnectionMark connected={props.connected} heartbeat={props.heartbeat} size={11} />
    </div>
  );
}
