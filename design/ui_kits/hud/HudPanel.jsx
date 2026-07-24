function HudPanel(props) {
  const legend = props.legend || { prev: '', action: '', next: '' };
  return (
    <div className="pixel-crisp" style={{ width: 250, height: 122, background: 'var(--paper)', border: '1px solid var(--ink)', padding: 3, boxSizing: 'border-box', display: 'flex', gap: 4, overflow: 'hidden', fontFamily: 'var(--font-pixel)', lineHeight: 1 }}>
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 2, overflow: 'hidden', minWidth: 0 }}>
        {React.Children.map(props.children, (c) => c && React.cloneElement(c, { style: { ...(c.props.style || {}), flexShrink: 0 } }))}
      </div>
      <div style={{ width: 1, alignSelf: 'stretch', background: 'var(--ink)' }} />
      <div style={{ width: 42, flexShrink: 0, display: 'flex', flexDirection: 'column', justifyContent: 'space-between', alignItems: 'flex-end', textAlign: 'right' }}>
        <span style={{ fontSize: 10, fontWeight: 700, color: 'var(--ink)' }}>{legend.prev}</span>
        <span style={{ fontSize: 10, fontWeight: 700, color: 'var(--ink)' }}>{legend.action}</span>
        <span style={{ fontSize: 10, fontWeight: 700, color: 'var(--ink)' }}>{legend.next}</span>
      </div>
    </div>
  );
}
window.HudPanel = HudPanel;
