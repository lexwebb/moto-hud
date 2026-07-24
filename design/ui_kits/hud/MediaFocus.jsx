const { ModeHeader, PixelDivider, MediaLine } = window.MotoHUDDesignSystem_632360;
function MediaFocus(props) {
  const playing = props && props.playing != null ? props.playing : true;
  return (
    <HudPanel legend={{ prev: 'SKIP', action: playing ? 'PAUSE' : 'PLAY', next: 'SKIP' }}>
      <ModeHeader mode="MEDIA" connected heartbeat />
      <PixelDivider variant="solid" />
      <div style={{ flex: 1, display: 'flex', alignItems: 'center' }}>
        <div style={{ transform: 'scale(1.35)', transformOrigin: 'left center' }}>
          <MediaLine playing={playing} title="Night Drive" artist="Field Tapes" />
        </div>
      </div>
    </HudPanel>
  );
}
window.MediaFocus = MediaFocus;
