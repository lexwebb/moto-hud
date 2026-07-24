const { ModeHeader, PixelDivider, ConnectionMark } = window.MotoHUDDesignSystem_632360;
function NavIdle() {
  return (
    <HudPanel legend={{ prev: 'MEDIA', action: '—', next: 'STATUS' }}>
      <ModeHeader mode="NAV" connected heartbeat />
      <PixelDivider variant="dashed" />
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', gap: 6 }}>
        <div style={{ fontSize: 17, fontWeight: 700, color: 'var(--ink)', letterSpacing: 1 }}>MOTO HUD</div>
        <div style={{ fontSize: 'var(--text-road)', fontWeight: 400, color: 'var(--ink)' }}>Waiting for route…</div>
      </div>
    </HudPanel>
  );
}
window.NavIdle = NavIdle;
