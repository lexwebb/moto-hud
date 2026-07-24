const { ModeHeader, PixelDivider, ConnectionMark } = window.MotoHUDDesignSystem_632360;
function StatusDiagnostics() {
  const row = { display: 'flex', justifyContent: 'space-between', fontSize: 'var(--text-road)', color: 'var(--ink)' };
  return (
    <HudPanel legend={{ prev: 'MODE', action: 'REDRAW', next: 'MODE' }}>
      <ModeHeader mode="STATUS" connected heartbeat />
      <PixelDivider variant="solid" />
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 5, justifyContent: 'center' }}>
        <div style={row}><span>LINK</span><ConnectionMark connected heartbeat size={12} /></div>
        <div style={row}><span>LAST BEAT</span><b>0.4s</b></div>
        <div style={row}><span>PACKETS</span><b>OK</b></div>
      </div>
    </HudPanel>
  );
}
window.StatusDiagnostics = StatusDiagnostics;
