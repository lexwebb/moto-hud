const { ManeuverGlyph, DistanceReadout, ModeHeader, PixelDivider, RoadRibbon } = window.MotoHUDDesignSystem_632360;
function NavActiveRibbon() {
  return (
    <HudPanel legend={{ prev: 'MODE', action: '—', next: 'MODE' }}>
      <ModeHeader mode="NAV" connected heartbeat />
      <PixelDivider variant="solid" />
      <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
        <ManeuverGlyph type="right" size={34} />
        <DistanceReadout value="120" unit="m" />
      </div>
      <div style={{ fontSize: 'var(--text-road)', fontWeight: 400, color: 'var(--ink)', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis', lineHeight: 1 }}>Ridge Rd</div>
      <div style={{ flex: 1 }} />
      <RoadRibbon points={[{ x: 110, y: 0 }, { x: 110, y: 22 }, { x: 175, y: 34 }]} turnIndex={1} height={44} />
    </HudPanel>
  );
}
window.NavActiveRibbon = NavActiveRibbon;
