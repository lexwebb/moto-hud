export interface ProgressTicksProps {
  /** Number of coarse ticks, default 5. */
  total?: number;
  /** How many ticks are filled — advances only at distance thresholds, never smoothly. */
  filled?: number;
}
/** Coarse progress-to-turn bar of discrete tick blocks — no smooth-fill animation. */
export function ProgressTicks(props: ProgressTicksProps): JSX.Element;
