import React from 'react';

export function PixelDivider(props) {
  const variant = props.variant || 'solid';
  const style = { width: '100%', height: 0, borderTop: '1.5px solid var(--ink)' };
  if (variant === 'dashed') style.borderTopStyle = 'dashed';
  if (variant === 'dither') return <div className="pixel-crisp" style={{ width: '100%', height: 2, background: 'var(--dither-25)' }} />;
  return <div style={style} />;
}
