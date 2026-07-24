export interface MediaLineProps {
  playing: boolean;
  title?: string;
  artist?: string;
}
/** Two-line title/artist row with a play/pause glyph. No album art — the phone doesn't send any. */
export function MediaLine(props: MediaLineProps): JSX.Element;
