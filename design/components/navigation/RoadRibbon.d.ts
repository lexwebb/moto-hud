export interface RoadPoint { x: number; y: number; }
export interface RoadRibbonProps {
  /** Simplified polyline of the next ~300–500m of route geometry, already projected to local screen-space units — not lat/lng. Omit while waiting on data. */
  points?: RoadPoint[];
  /** Index into points marking the upcoming maneuver kink, drawn as a filled square. */
  turnIndex?: number;
  width?: number;
  height?: number;
}
/** Schematic corridor ribbon — a shape, not a map. Reserved band for the road-view stretch feature. */
export function RoadRibbon(props: RoadRibbonProps): JSX.Element;
