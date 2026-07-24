import { ConnectionMarkProps } from '../glyphs/ConnectionMark';
export interface ModeHeaderProps {
  /** Current screen name — NAV / MEDIA / STATUS. */
  mode: 'NAV' | 'MEDIA' | 'STATUS';
  connected: boolean;
  heartbeat?: boolean;
}
/** Top chrome row: mode label + BLE mark. Present on every screen for orientation. */
export function ModeHeader(props: ModeHeaderProps): JSX.Element;
