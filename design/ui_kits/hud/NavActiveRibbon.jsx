const { DistanceReadout, ModeHeader, PixelDivider, ETAReadout } = window.MotoHUDDesignSystem_632360;

function MiniMapSketch({ width = 70, height = 72 }) {
  const cx = width / 2;
  const cy = height / 2;
  const riderY = cy + 22;
  return (
    <svg className="pixel-crisp" width={width} height={height} viewBox={`0 0 ${width} ${height}`}>
      <path d={`M8,${cy + 8} L12,${cy - 22} L6,${cy - 30}`} fill="none" stroke="var(--ink)" strokeWidth="1" strokeDasharray="3 3" />
      <path d={`M${width - 10},${cy - 10} L${width - 6},${cy + 24}`} fill="none" stroke="var(--ink)" strokeWidth="1" strokeDasharray="3 3" />
      <path d={`M${cx},${riderY} L${cx},${cy} L${cx + 16},${cy - 14}`} fill="none" stroke="var(--ink)" strokeWidth="2" strokeLinejoin="miter" strokeLinecap="square" />
      <rect x={cx - 2} y={cy - 2} width="4" height="4" fill="var(--ink)" />
      <rect x={cx - 2.5} y={riderY - 2.5} width="5" height="5" fill="var(--ink)" />
    </svg>
  );
}

function NavActiveRibbon() {
  return (
    <HudPanel legend={{ prev: 'MODE', action: '—', next: 'MODE' }}>
      <ModeHeader mode="NAV" connected heartbeat />
      <PixelDivider variant="solid" />
      <div style={{ flex: 1, display: 'flex', gap: 4, minHeight: 0, overflow: 'hidden' }}>
        <div style={{ width: '36%', flexShrink: 0, minWidth: 0, display: 'flex', alignItems: 'stretch' }}>
          <MiniMapSketch height={72} width={70} />
        </div>
        <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 4, minWidth: 0 }}>
          <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
            <DistanceReadout value="≈120m" />
          </div>
          <div style={{ fontSize: 'var(--text-road)', fontWeight: 400, color: 'var(--ink)', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis', lineHeight: 1 }}>Ridge Rd</div>
          <ETAReadout etaMin={8} />
        </div>
      </div>
    </HudPanel>
  );
}
window.NavActiveRibbon = NavActiveRibbon;
