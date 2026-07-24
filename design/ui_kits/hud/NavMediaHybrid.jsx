const { ManeuverGlyph, DistanceReadout, ModeHeader, PixelDivider, MediaLine } = window.MotoHUDDesignSystem_632360;
function NavMediaHybrid() {
  return (
    <HudPanel legend={{ prev: 'MODE', action: '—', next: 'MODE' }}>
      <ModeHeader mode="NAV" connected heartbeat />
      <PixelDivider variant="solid" />
      <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
        <ManeuverGlyph type="straight" size={32} />
        <DistanceReadout value="800" unit="m" />
      </div>
      <div style={{ fontSize: 'var(--text-road)', fontWeight: 400, color: 'var(--ink)' }}>Continue on Route 9</div>
      <PixelDivider variant="dither" />
      <div style={{ transform: 'scale(0.92)', transformOrigin: 'left center' }}>
        <MediaLine playing title="Night Drive" artist="Field Tapes" />
      </div>
    </HudPanel>
  );
}
window.NavMediaHybrid = NavMediaHybrid;
