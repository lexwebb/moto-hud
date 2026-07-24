export interface ConnectionMarkProps {
  /** BLE link to phone established. */
  connected: boolean;
  /** Blink the heartbeat dot (only meaningful when connected). Discrete on/off, never a smooth fade. */
  heartbeat?: boolean;
  size?: number;
}
/** BLE status mark: bolt = linked, X = lost. Blinking square dot = heartbeat received. */
export function ConnectionMark(props: ConnectionMarkProps): JSX.Element;
