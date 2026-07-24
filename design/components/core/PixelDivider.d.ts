export interface PixelDividerProps {
  variant?: 'solid' | 'dashed' | 'dither';
}
/** Hairline row separator. `dither` variant is the panel's stand-in for gray on 1-bit e-ink. */
export function PixelDivider(props: PixelDividerProps): JSX.Element;
