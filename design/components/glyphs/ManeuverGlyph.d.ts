export interface ManeuverGlyphProps {
  /** Maneuver type from the phone's nav packet. */
  type: 'left' | 'right' | 'slight-left' | 'slight-right' | 'straight' | 'u-turn' | 'roundabout' | 'arrive' | 'depart' | 'unknown';
  /** Square glyph size in px. Default 40. */
  size?: number;
}
/** Pixel-grid maneuver arrow, the hero glyph on the Nav screen. */
export function ManeuverGlyph(props: ManeuverGlyphProps): JSX.Element;
