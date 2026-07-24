const { ManeuverGlyph, DistanceReadout, ETAReadout, ModeHeader, PixelDivider, ProgressTicks } = window.MotoHUDDesignSystem_632360;
function NavActiveNoRibbon() {
  return (
    <HudPanel legend={{ prev: 'MODE', action: '—', next: 'MODE' }}>
      <ModeHeader mode="NAV" connected heartbeat />
      <PixelDivider variant="solid" />
      <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
        <ManeuverGlyph type="left" size={38} />
        <DistanceReadout value="350" unit="m" />
      </div>
      <div style={{ fontSize: 'var(--text-road)', fontWeight: 400, color: 'var(--ink)', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis', lineHeight: 1 }}>onto Harbor Blvd</div>
      <div style={{ flex: 1 }} />
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline' }}>
        <ETAReadout etaMin={12} />
        <ProgressTicks total={5} filled={3} />
      </div>
    </HudPanel>
  );
}
window.NavActiveNoRibbon = NavActiveNoRibbon;
