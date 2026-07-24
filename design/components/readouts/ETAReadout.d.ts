export interface ETAReadoutProps {
  /** Minutes to arrival, straight from the nav packet's optional eta_min. Omit/undefined renders nothing. */
  etaMin?: number;
}
/** Small "ETA N min" line — secondary to the hero distance, optional per packet. */
export function ETAReadout(props: ETAReadoutProps): JSX.Element | null;
