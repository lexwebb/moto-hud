export interface FooterHint { btn: 'PREV' | 'NEXT' | 'HOLD'; label: string; }
export interface FooterHintsProps {
  /** 2–3 short hints max — this row must stay ≤11px and non-competing. */
  hints: FooterHint[];
}
/** Tiny bottom-row button legend, e.g. "PREV Mode  NEXT Skip  HOLD Home". */
export function FooterHints(props: FooterHintsProps): JSX.Element;
