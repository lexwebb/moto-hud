export interface DistanceReadoutProps {
  /** Pre-formatted numeric distance string, e.g. "350" or "1.2". Formatting/rounding happens on the phone. */
  value: string;
  /** Unit shown as small type beside the number, e.g. "m" or "km" — unit switching is typographic, never a control. */
  unit?: string;
  /** hero = 26px main readout, secondary = 17px (e.g. ETA row). Default hero. */
  size?: 'hero' | 'secondary';
}
/** The distance-to-turn number, the single largest element on the Nav screen. */
export function DistanceReadout(props: DistanceReadoutProps): JSX.Element;
