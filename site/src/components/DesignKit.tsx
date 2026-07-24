import React, { useRef, useState } from 'react';
import { ManeuverGlyph } from '@design/components/glyphs/ManeuverGlyph.jsx';
import { ConnectionMark } from '@design/components/glyphs/ConnectionMark.jsx';
import { ModeHeader } from '@design/components/chrome/ModeHeader.jsx';
import { PixelDivider } from '@design/components/core/PixelDivider.jsx';
import { DistanceReadout } from '@design/components/readouts/DistanceReadout.jsx';
import { ETAReadout } from '@design/components/readouts/ETAReadout.jsx';
import { MediaLine } from '@design/components/readouts/MediaLine.jsx';
import { ProgressTicks } from '@design/components/navigation/ProgressTicks.jsx';
import { RoadRibbon } from '@design/components/navigation/RoadRibbon.jsx';

function HudPanel({
  legend,
  children,
}: {
  legend: { prev: string; action: string; next: string };
  children: React.ReactNode;
}) {
  return (
    <div
      className="pixel-crisp"
      style={{
        width: 250,
        height: 122,
        background: 'var(--paper)',
        border: '1px solid var(--ink)',
        padding: 3,
        boxSizing: 'border-box',
        display: 'flex',
        gap: 4,
        overflow: 'hidden',
        fontFamily: 'var(--font-pixel)',
        lineHeight: 1,
      }}
    >
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 2, overflow: 'hidden', minWidth: 0 }}>
        {React.Children.map(children, (c) =>
          c && React.isValidElement(c)
            ? React.cloneElement(c as React.ReactElement<{ style?: React.CSSProperties }>, {
                style: {
                  ...((c.props as { style?: React.CSSProperties }).style || {}),
                  flexShrink: 0,
                },
              })
            : c,
        )}
      </div>
      <div style={{ width: 1, alignSelf: 'stretch', background: 'var(--ink)' }} />
      <div
        style={{
          width: 42,
          flexShrink: 0,
          display: 'flex',
          flexDirection: 'column',
          justifyContent: 'space-between',
          alignItems: 'flex-end',
          textAlign: 'right',
        }}
      >
        <span style={{ fontSize: 10, fontWeight: 700, color: 'var(--ink)' }}>{legend.prev}</span>
        <span style={{ fontSize: 10, fontWeight: 700, color: 'var(--ink)' }}>{legend.action}</span>
        <span style={{ fontSize: 10, fontWeight: 700, color: 'var(--ink)' }}>{legend.next}</span>
      </div>
    </div>
  );
}

function NavActiveNoRibbon() {
  return (
    <HudPanel legend={{ prev: 'MODE', action: '—', next: 'MODE' }}>
      <ModeHeader mode="NAV" connected heartbeat />
      <PixelDivider variant="solid" />
      <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
        <ManeuverGlyph type="left" size={38} />
        <DistanceReadout value="350" unit="m" />
      </div>
      <div
        style={{
          fontSize: 'var(--text-road)',
          fontWeight: 400,
          color: 'var(--ink)',
          whiteSpace: 'nowrap',
          overflow: 'hidden',
          textOverflow: 'ellipsis',
          lineHeight: 1,
        }}
      >
        onto Harbor Blvd
      </div>
      <div style={{ flex: 1 }} />
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline' }}>
        <ETAReadout etaMin={12} />
        <ProgressTicks total={5} filled={3} />
      </div>
    </HudPanel>
  );
}

function NavActiveRibbon() {
  return (
    <HudPanel legend={{ prev: 'MODE', action: '—', next: 'MODE' }}>
      <ModeHeader mode="NAV" connected heartbeat />
      <PixelDivider variant="solid" />
      <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
        <ManeuverGlyph type="right" size={34} />
        <DistanceReadout value="120" unit="m" />
      </div>
      <div
        style={{
          fontSize: 'var(--text-road)',
          fontWeight: 400,
          color: 'var(--ink)',
          whiteSpace: 'nowrap',
          overflow: 'hidden',
          textOverflow: 'ellipsis',
          lineHeight: 1,
        }}
      >
        Ridge Rd
      </div>
      <div style={{ flex: 1 }} />
      <RoadRibbon
        points={[
          { x: 110, y: 0 },
          { x: 110, y: 22 },
          { x: 175, y: 34 },
        ]}
        turnIndex={1}
        height={44}
      />
    </HudPanel>
  );
}

function NavIdle() {
  return (
    <HudPanel legend={{ prev: 'MEDIA', action: '—', next: 'STATUS' }}>
      <ModeHeader mode="NAV" connected heartbeat />
      <PixelDivider variant="dashed" />
      <div
        style={{
          flex: 1,
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          gap: 6,
        }}
      >
        <div style={{ fontSize: 17, fontWeight: 700, color: 'var(--ink)', letterSpacing: 1 }}>MOTO HUD</div>
        <div style={{ fontSize: 'var(--text-road)', fontWeight: 400, color: 'var(--ink)' }}>Waiting for route…</div>
      </div>
    </HudPanel>
  );
}

function MediaFocus({ playing = true }: { playing?: boolean }) {
  return (
    <HudPanel legend={{ prev: 'SKIP', action: playing ? 'PAUSE' : 'PLAY', next: 'SKIP' }}>
      <ModeHeader mode="MEDIA" connected heartbeat />
      <PixelDivider variant="solid" />
      <div style={{ flex: 1, display: 'flex', alignItems: 'center' }}>
        <div style={{ transform: 'scale(1.35)', transformOrigin: 'left center' }}>
          <MediaLine playing={playing} title="Night Drive" artist="Field Tapes" />
        </div>
      </div>
    </HudPanel>
  );
}

function StatusDiagnostics() {
  const row = {
    display: 'flex',
    justifyContent: 'space-between',
    fontSize: 'var(--text-road)',
    color: 'var(--ink)',
  } as const;
  return (
    <HudPanel legend={{ prev: 'MODE', action: 'REDRAW', next: 'MODE' }}>
      <ModeHeader mode="STATUS" connected heartbeat />
      <PixelDivider variant="solid" />
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 5, justifyContent: 'center' }}>
        <div style={row}>
          <span>LINK</span>
          <ConnectionMark connected heartbeat size={12} />
        </div>
        <div style={row}>
          <span>LAST BEAT</span>
          <b>0.4s</b>
        </div>
        <div style={row}>
          <span>PACKETS</span>
          <b>OK</b>
        </div>
      </div>
    </HudPanel>
  );
}

function NavMediaHybrid() {
  return (
    <HudPanel legend={{ prev: 'MODE', action: '—', next: 'MODE' }}>
      <ModeHeader mode="NAV" connected heartbeat />
      <PixelDivider variant="solid" />
      <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
        <ManeuverGlyph type="straight" size={32} />
        <DistanceReadout value="800" unit="m" />
      </div>
      <div style={{ fontSize: 'var(--text-road)', fontWeight: 400, color: 'var(--ink)' }}>Continue on Route 9</div>
      <PixelDivider variant="dither" />
      <div style={{ transform: 'scale(0.92)', transformOrigin: 'left center' }}>
        <MediaLine playing title="Night Drive" artist="Field Tapes" />
      </div>
    </HudPanel>
  );
}

const ORDER = [
  'NavActiveNoRibbon',
  'NavActiveRibbon',
  'NavIdle',
  'MediaFocus',
  'StatusDiagnostics',
  'NavMediaHybrid',
] as const;

const LABELS: Record<(typeof ORDER)[number], string> = {
  NavActiveNoRibbon: 'Nav active — no ribbon',
  NavActiveRibbon: 'Nav active — with ribbon',
  NavIdle: 'Nav idle / waiting for route',
  MediaFocus: 'Media focus',
  StatusDiagnostics: 'Status / link diagnostics',
  NavMediaHybrid: 'Nav + media hybrid',
};

const SCREENS = {
  NavActiveNoRibbon,
  NavActiveRibbon,
  NavIdle,
  MediaFocus,
  StatusDiagnostics,
  NavMediaHybrid,
};

export default function DesignKit() {
  const [i, setI] = useState(0);
  const [playing, setPlaying] = useState(true);
  const [flash, setFlash] = useState(false);
  const pressTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const longRef = useRef(false);

  function down(fn: (isLong: boolean) => void) {
    longRef.current = false;
    pressTimer.current = setTimeout(() => {
      longRef.current = true;
      fn(true);
    }, 500);
  }
  function up(fn: (isLong: boolean) => void) {
    if (pressTimer.current) clearTimeout(pressTimer.current);
    if (!longRef.current) fn(false);
  }
  function prev() {
    if (ORDER[i] === 'MediaFocus') return;
    setI((v) => (v - 1 + ORDER.length) % ORDER.length);
  }
  function next() {
    if (ORDER[i] === 'MediaFocus') return;
    setI((v) => (v + 1) % ORDER.length);
  }
  function action(isLong: boolean) {
    if (isLong) {
      setI(0);
      return;
    }
    if (ORDER[i] === 'MediaFocus') setPlaying((p) => !p);
    if (ORDER[i] === 'StatusDiagnostics') {
      setFlash(true);
      setTimeout(() => setFlash(false), 150);
    }
  }

  const name = ORDER[i];
  const Screen = SCREENS[name];

  return (
    <div className="design-kit">
      <p className="tag">{LABELS[name]}</p>
      <div className="panel-row">
        <div className="stage-wrap">
          <div className="stage" style={{ opacity: flash ? 0.3 : 1 }}>
            {name === 'MediaFocus' ? <MediaFocus playing={playing} /> : <Screen />}
          </div>
        </div>
        <div className="btncol">
          <button type="button" onMouseDown={() => down(prev)} onMouseUp={() => up(prev)}>
            PREV
          </button>
          <button type="button" onMouseDown={() => down(action)} onMouseUp={() => up(action)}>
            ACTION
          </button>
          <button type="button" onMouseDown={() => down(next)} onMouseUp={() => up(next)}>
            NEXT
          </button>
        </div>
      </div>
      <style>{`
        .design-kit {
          display: flex;
          flex-direction: column;
          align-items: center;
          gap: 16px;
          padding: 8px 0 24px;
          --ink: #000;
          --paper: #fff;
          --font-pixel: ui-monospace, "Cascadia Mono", monospace;
          --text-meta: 12px;
          --text-road: 16px;
          --text-eta: 16px;
          --text-hero: 32px;
          --lh-tight: 1;
          --dither-25: repeating-linear-gradient(90deg, #000 0 1px, #fff 1px 2px);
        }
        .tag { margin: 0; font-size: 0.9rem; color: #9aa0a6; }
        .panel-row { display: flex; align-items: flex-start; gap: 24px; flex-wrap: wrap; justify-content: center; }
        .stage-wrap { width: 600px; height: 293px; background: #eee; padding: 12px; border-radius: 4px; }
        .stage { display: inline-block; width: 250px; height: 122px; transform: scale(2.4); transform-origin: top left; }
        .btncol { display: flex; flex-direction: column; gap: 10px; }
        .btncol button {
          font-family: ui-monospace, monospace;
          font-size: 12px;
          padding: 10px 16px;
          border: 2px solid #e8eaed;
          background: #1c1f24;
          color: #e8eaed;
          cursor: pointer;
          width: 90px;
        }
        .btncol button:active { background: #c4a35a; color: #0e1012; border-color: #c4a35a; }
      `}</style>
    </div>
  );
}
