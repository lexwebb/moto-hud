/* @ds-bundle: {"format":4,"namespace":"MotoHUDDesignSystem_632360","components":[{"name":"FooterHints","sourcePath":"components/chrome/FooterHints.jsx"},{"name":"ModeHeader","sourcePath":"components/chrome/ModeHeader.jsx"},{"name":"PixelDivider","sourcePath":"components/core/PixelDivider.jsx"},{"name":"ConnectionMark","sourcePath":"components/glyphs/ConnectionMark.jsx"},{"name":"ManeuverGlyph","sourcePath":"components/glyphs/ManeuverGlyph.jsx"},{"name":"ProgressTicks","sourcePath":"components/navigation/ProgressTicks.jsx"},{"name":"RoadRibbon","sourcePath":"components/navigation/RoadRibbon.jsx"},{"name":"DistanceReadout","sourcePath":"components/readouts/DistanceReadout.jsx"},{"name":"ETAReadout","sourcePath":"components/readouts/ETAReadout.jsx"},{"name":"MediaLine","sourcePath":"components/readouts/MediaLine.jsx"}],"sourceHashes":{"components/chrome/FooterHints.jsx":"27bcaca4de95","components/chrome/ModeHeader.jsx":"0fef930f9e51","components/core/PixelDivider.jsx":"963a0d645d80","components/glyphs/ConnectionMark.jsx":"c2514dbb29b2","components/glyphs/ManeuverGlyph.jsx":"c51ddd7af85b","components/navigation/ProgressTicks.jsx":"e7b3608a9a9d","components/navigation/RoadRibbon.jsx":"7aeb3401a6e6","components/readouts/DistanceReadout.jsx":"856ee2d739a9","components/readouts/ETAReadout.jsx":"62c356ae4dd0","components/readouts/MediaLine.jsx":"81e2b7af0d72","ui_kits/hud/HudPanel.jsx":"9ff3fb6dc653","ui_kits/hud/MediaFocus.jsx":"a3b32d95f2c1","ui_kits/hud/NavActiveNoRibbon.jsx":"3e8aef8d6cde","ui_kits/hud/NavActiveRibbon.jsx":"26dd8640bea9","ui_kits/hud/NavIdle.jsx":"dc2b10cd8c24","ui_kits/hud/NavMediaHybrid.jsx":"7472650a0e81","ui_kits/hud/StatusDiagnostics.jsx":"fe1a1c45d9a9"},"inlinedExternals":[],"unexposedExports":[]} */

(() => {

const __ds_ns = (window.MotoHUDDesignSystem_632360 = window.MotoHUDDesignSystem_632360 || {});

const __ds_scope = {};

(__ds_ns.__errors = __ds_ns.__errors || []);

// components/chrome/FooterHints.jsx
try { (() => {
function FooterHints(props) {
  const hints = props.hints || [];
  return /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      gap: 10,
      fontFamily: 'var(--font-pixel)',
      fontSize: 'var(--text-meta)',
      color: 'var(--ink)',
      lineHeight: 1,
      justifyContent: 'flex-end',
      textAlign: 'right'
    }
  }, hints.map((h, i) => /*#__PURE__*/React.createElement("span", {
    key: i,
    style: {
      display: 'flex',
      gap: 3
    }
  }, /*#__PURE__*/React.createElement("span", {
    style: {
      fontWeight: 700
    }
  }, h.btn), /*#__PURE__*/React.createElement("span", {
    style: {
      fontWeight: 400
    }
  }, h.label))));
}
Object.assign(__ds_scope, { FooterHints });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/chrome/FooterHints.jsx", error: String((e && e.message) || e) }); }

// components/core/PixelDivider.jsx
try { (() => {
function PixelDivider(props) {
  const variant = props.variant || 'solid';
  const style = {
    width: '100%',
    height: 0,
    borderTop: '1.5px solid var(--ink)'
  };
  if (variant === 'dashed') style.borderTopStyle = 'dashed';
  if (variant === 'dither') return /*#__PURE__*/React.createElement("div", {
    className: "pixel-crisp",
    style: {
      width: '100%',
      height: 2,
      background: 'var(--dither-25)'
    }
  });
  return /*#__PURE__*/React.createElement("div", {
    style: style
  });
}
Object.assign(__ds_scope, { PixelDivider });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/core/PixelDivider.jsx", error: String((e && e.message) || e) }); }

// components/glyphs/ConnectionMark.jsx
try { (() => {
function ConnectionMark(props) {
  const connected = !!props.connected;
  const heartbeat = !!props.heartbeat;
  const size = props.size || 12;
  return /*#__PURE__*/React.createElement("span", {
    style: {
      display: 'inline-flex',
      alignItems: 'center',
      gap: 4
    }
  }, connected ? /*#__PURE__*/React.createElement("svg", {
    className: "pixel-crisp",
    width: size,
    height: size,
    viewBox: "0 0 12 12"
  }, /*#__PURE__*/React.createElement("path", {
    d: "M2,9 L6,2 L6,6 L10,6 L6,10 L6,6",
    fill: "none",
    stroke: "var(--ink)",
    strokeWidth: "1.6",
    strokeLinejoin: "miter"
  })) : /*#__PURE__*/React.createElement("svg", {
    className: "pixel-crisp",
    width: size,
    height: size,
    viewBox: "0 0 12 12"
  }, /*#__PURE__*/React.createElement("line", {
    x1: "2",
    y1: "2",
    x2: "10",
    y2: "10",
    stroke: "var(--ink)",
    strokeWidth: "1.6"
  }), /*#__PURE__*/React.createElement("line", {
    x1: "10",
    y1: "2",
    x2: "2",
    y2: "10",
    stroke: "var(--ink)",
    strokeWidth: "1.6"
  })), connected && /*#__PURE__*/React.createElement("span", {
    className: "pixel-crisp",
    style: {
      width: 4,
      height: 4,
      background: 'var(--ink)',
      display: 'inline-block',
      animation: heartbeat ? 'moto-hb 1.6s steps(1) infinite' : 'none'
    }
  }), /*#__PURE__*/React.createElement("style", null, '@keyframes moto-hb{0%,49%{opacity:1}50%,100%{opacity:0}}'));
}
Object.assign(__ds_scope, { ConnectionMark });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/glyphs/ConnectionMark.jsx", error: String((e && e.message) || e) }); }

// components/chrome/ModeHeader.jsx
try { (() => {
function ModeHeader(props) {
  return /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      justifyContent: 'space-between',
      alignItems: 'center',
      fontFamily: 'var(--font-pixel)',
      color: 'var(--ink)',
      lineHeight: 1
    }
  }, /*#__PURE__*/React.createElement("span", {
    style: {
      fontSize: 'var(--text-meta)',
      fontWeight: 700,
      letterSpacing: 1
    }
  }, props.mode), /*#__PURE__*/React.createElement(__ds_scope.ConnectionMark, {
    connected: props.connected,
    heartbeat: props.heartbeat,
    size: 11
  }));
}
Object.assign(__ds_scope, { ModeHeader });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/chrome/ModeHeader.jsx", error: String((e && e.message) || e) }); }

// components/glyphs/ManeuverGlyph.jsx
try { (() => {
const HEAD = 7;
function arrowHead(x, y, angleDeg) {
  const a = angleDeg * Math.PI / 180;
  const tip = [x, y];
  const back = [x - Math.cos(a) * HEAD, y - Math.sin(a) * HEAD];
  const p1 = [back[0] + Math.cos(a + Math.PI / 2) * HEAD * 0.62, back[1] + Math.sin(a + Math.PI / 2) * HEAD * 0.62];
  const p2 = [back[0] - Math.cos(a + Math.PI / 2) * HEAD * 0.62, back[1] - Math.sin(a + Math.PI / 2) * HEAD * 0.62];
  return `${tip[0]},${tip[1]} ${p1[0]},${p1[1]} ${p2[0]},${p2[1]}`;
}
function glyphBody(type) {
  const stem = /*#__PURE__*/React.createElement("line", {
    x1: "20",
    y1: "34",
    x2: "20",
    y2: "14",
    stroke: "var(--ink)",
    strokeWidth: "3",
    strokeLinecap: "square"
  });
  switch (type) {
    case 'straight':
      return /*#__PURE__*/React.createElement(React.Fragment, null, stem, /*#__PURE__*/React.createElement("polygon", {
        points: arrowHead(20, 7, -90),
        fill: "var(--ink)"
      }));
    case 'left':
      return /*#__PURE__*/React.createElement(React.Fragment, null, /*#__PURE__*/React.createElement("line", {
        x1: "20",
        y1: "34",
        x2: "20",
        y2: "17",
        stroke: "var(--ink)",
        strokeWidth: "3",
        strokeLinecap: "square"
      }), /*#__PURE__*/React.createElement("line", {
        x1: "20",
        y1: "17",
        x2: "8",
        y2: "17",
        stroke: "var(--ink)",
        strokeWidth: "3",
        strokeLinecap: "square"
      }), /*#__PURE__*/React.createElement("polygon", {
        points: arrowHead(1, 17, 180),
        fill: "var(--ink)"
      }));
    case 'right':
      return /*#__PURE__*/React.createElement(React.Fragment, null, /*#__PURE__*/React.createElement("line", {
        x1: "20",
        y1: "34",
        x2: "20",
        y2: "17",
        stroke: "var(--ink)",
        strokeWidth: "3",
        strokeLinecap: "square"
      }), /*#__PURE__*/React.createElement("line", {
        x1: "20",
        y1: "17",
        x2: "32",
        y2: "17",
        stroke: "var(--ink)",
        strokeWidth: "3",
        strokeLinecap: "square"
      }), /*#__PURE__*/React.createElement("polygon", {
        points: arrowHead(39, 17, 0),
        fill: "var(--ink)"
      }));
    case 'slight-left':
      return /*#__PURE__*/React.createElement(React.Fragment, null, /*#__PURE__*/React.createElement("line", {
        x1: "21",
        y1: "34",
        x2: "21",
        y2: "24",
        stroke: "var(--ink)",
        strokeWidth: "3",
        strokeLinecap: "square"
      }), /*#__PURE__*/React.createElement("line", {
        x1: "21",
        y1: "24",
        x2: "11",
        y2: "9",
        stroke: "var(--ink)",
        strokeWidth: "3",
        strokeLinecap: "square"
      }), /*#__PURE__*/React.createElement("polygon", {
        points: arrowHead(7, 4, 214),
        fill: "var(--ink)"
      }));
    case 'slight-right':
      return /*#__PURE__*/React.createElement(React.Fragment, null, /*#__PURE__*/React.createElement("line", {
        x1: "19",
        y1: "34",
        x2: "19",
        y2: "24",
        stroke: "var(--ink)",
        strokeWidth: "3",
        strokeLinecap: "square"
      }), /*#__PURE__*/React.createElement("line", {
        x1: "19",
        y1: "24",
        x2: "29",
        y2: "9",
        stroke: "var(--ink)",
        strokeWidth: "3",
        strokeLinecap: "square"
      }), /*#__PURE__*/React.createElement("polygon", {
        points: arrowHead(33, 4, -34),
        fill: "var(--ink)"
      }));
    case 'u-turn':
      return /*#__PURE__*/React.createElement(React.Fragment, null, /*#__PURE__*/React.createElement("path", {
        d: "M14,34 V16 A6,6 0 0 1 26,16 V26",
        fill: "none",
        stroke: "var(--ink)",
        strokeWidth: "3",
        strokeLinecap: "square"
      }), /*#__PURE__*/React.createElement("polygon", {
        points: arrowHead(26, 33, 90),
        fill: "var(--ink)"
      }));
    case 'roundabout':
      return /*#__PURE__*/React.createElement(React.Fragment, null, /*#__PURE__*/React.createElement("line", {
        x1: "20",
        y1: "34",
        x2: "20",
        y2: "27",
        stroke: "var(--ink)",
        strokeWidth: "3",
        strokeLinecap: "square"
      }), /*#__PURE__*/React.createElement("circle", {
        cx: "20",
        cy: "17",
        r: "9",
        fill: "none",
        stroke: "var(--ink)",
        strokeWidth: "3"
      }), /*#__PURE__*/React.createElement("polygon", {
        points: arrowHead(29, 8, 0),
        fill: "var(--ink)"
      }));
    case 'arrive':
      return /*#__PURE__*/React.createElement(React.Fragment, null, /*#__PURE__*/React.createElement("line", {
        x1: "13",
        y1: "34",
        x2: "13",
        y2: "7",
        stroke: "var(--ink)",
        strokeWidth: "3",
        strokeLinecap: "square"
      }), /*#__PURE__*/React.createElement("polygon", {
        points: "13,7 29,11 13,17",
        fill: "var(--ink)"
      }));
    case 'depart':
      return /*#__PURE__*/React.createElement(React.Fragment, null, /*#__PURE__*/React.createElement("circle", {
        cx: "20",
        cy: "27",
        r: "4",
        fill: "var(--ink)"
      }), /*#__PURE__*/React.createElement("line", {
        x1: "20",
        y1: "22",
        x2: "20",
        y2: "8",
        stroke: "var(--ink)",
        strokeWidth: "3",
        strokeLinecap: "square"
      }), /*#__PURE__*/React.createElement("polygon", {
        points: arrowHead(20, 4, -90),
        fill: "var(--ink)"
      }));
    default:
      return /*#__PURE__*/React.createElement("text", {
        x: "20",
        y: "29",
        fontFamily: "var(--font-pixel)",
        fontWeight: "700",
        fontSize: "24",
        textAnchor: "middle",
        fill: "var(--ink)"
      }, "?");
  }
}
function ManeuverGlyph(props) {
  const size = props.size || 40;
  return /*#__PURE__*/React.createElement("svg", {
    className: "pixel-crisp",
    width: size,
    height: size,
    viewBox: "0 0 40 40",
    style: {
      display: 'block'
    }
  }, glyphBody(props.type));
}
Object.assign(__ds_scope, { ManeuverGlyph });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/glyphs/ManeuverGlyph.jsx", error: String((e && e.message) || e) }); }

// components/navigation/ProgressTicks.jsx
try { (() => {
function ProgressTicks(props) {
  const total = props.total || 5;
  const filled = Math.max(0, Math.min(total, props.filled || 0));
  const ticks = [];
  for (let i = 0; i < total; i++) ticks.push(i < filled);
  return /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      gap: 3
    }
  }, ticks.map((on, i) => /*#__PURE__*/React.createElement("span", {
    key: i,
    style: {
      width: 8,
      height: 10,
      background: on ? 'var(--ink)' : 'var(--paper)',
      border: '1.5px solid var(--ink)',
      boxSizing: 'border-box'
    }
  })));
}
Object.assign(__ds_scope, { ProgressTicks });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/navigation/ProgressTicks.jsx", error: String((e && e.message) || e) }); }

// components/navigation/RoadRibbon.jsx
try { (() => {
function RoadRibbon(props) {
  const w = props.width || 220;
  const h = props.height || 40;
  const points = props.points;
  if (!points || points.length < 2) {
    return /*#__PURE__*/React.createElement("svg", {
      className: "pixel-crisp",
      width: w,
      height: h,
      viewBox: `0 0 ${w} ${h}`
    }, /*#__PURE__*/React.createElement("line", {
      x1: w / 2,
      y1: "4",
      x2: w / 2,
      y2: h - 4,
      stroke: "var(--ink)",
      strokeWidth: "2",
      strokeDasharray: "4 5"
    }));
  }
  const xs = points.map(p => p.x);
  const ys = points.map(p => p.y);
  const minX = Math.min(...xs),
    maxX = Math.max(...xs);
  const minY = Math.min(...ys),
    maxY = Math.max(...ys);
  const pad = 5;
  const rangeX = maxX - minX || 1;
  const rangeY = maxY - minY || 1;
  const scale = Math.min((w - pad * 2) / rangeX, (h - pad * 2) / rangeY);
  const plottedW = rangeX * scale;
  const offsetX = pad + (w - pad * 2 - plottedW) / 2;
  const screenBottom = h - pad;
  const sx = x => offsetX + (x - minX) * scale;
  const sy = y => screenBottom - (y - minY) * scale;
  const d = points.map((p, i) => `${i === 0 ? 'M' : 'L'}${sx(p.x)},${sy(p.y)}`).join(' ');
  const turn = props.turnIndex != null ? points[props.turnIndex] : null;
  return /*#__PURE__*/React.createElement("svg", {
    className: "pixel-crisp",
    width: w,
    height: h,
    viewBox: `0 0 ${w} ${h}`
  }, /*#__PURE__*/React.createElement("path", {
    d: d,
    fill: "none",
    stroke: "var(--ink)",
    strokeWidth: "3",
    strokeLinejoin: "miter",
    strokeLinecap: "square"
  }), turn && /*#__PURE__*/React.createElement("rect", {
    x: sx(turn.x) - 3,
    y: sy(turn.y) - 3,
    width: "6",
    height: "6",
    fill: "var(--ink)"
  }));
}
Object.assign(__ds_scope, { RoadRibbon });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/navigation/RoadRibbon.jsx", error: String((e && e.message) || e) }); }

// components/readouts/DistanceReadout.jsx
try { (() => {
function DistanceReadout(props) {
  const value = props.value ?? '—';
  const unit = props.unit || '';
  const size = props.size || 'hero';
  const fontSize = size === 'hero' ? 'var(--text-hero)' : 'var(--text-eta)';
  return /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      alignItems: 'baseline',
      gap: 4,
      fontFamily: 'var(--font-pixel)',
      color: 'var(--ink)',
      lineHeight: 'var(--lh-tight)'
    }
  }, /*#__PURE__*/React.createElement("span", {
    style: {
      fontSize,
      fontWeight: 700
    }
  }, value), unit && /*#__PURE__*/React.createElement("span", {
    style: {
      fontSize: 'var(--text-road)',
      fontWeight: 700
    }
  }, unit));
}
Object.assign(__ds_scope, { DistanceReadout });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/readouts/DistanceReadout.jsx", error: String((e && e.message) || e) }); }

// components/readouts/ETAReadout.jsx
try { (() => {
function ETAReadout(props) {
  if (props.etaMin == null) return null;
  return /*#__PURE__*/React.createElement("div", {
    style: {
      fontFamily: 'var(--font-pixel)',
      fontSize: 'var(--text-road)',
      fontWeight: 400,
      color: 'var(--ink)',
      display: 'flex',
      gap: 4,
      lineHeight: 1
    }
  }, /*#__PURE__*/React.createElement("span", null, "ETA"), /*#__PURE__*/React.createElement("span", {
    style: {
      fontWeight: 700
    }
  }, props.etaMin), /*#__PURE__*/React.createElement("span", null, "min"));
}
Object.assign(__ds_scope, { ETAReadout });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/readouts/ETAReadout.jsx", error: String((e && e.message) || e) }); }

// components/readouts/MediaLine.jsx
try { (() => {
function MediaLine(props) {
  const playing = !!props.playing;
  return /*#__PURE__*/React.createElement("div", {
    style: {
      fontFamily: 'var(--font-pixel)',
      color: 'var(--ink)',
      display: 'flex',
      alignItems: 'center',
      gap: 6
    }
  }, /*#__PURE__*/React.createElement("svg", {
    className: "pixel-crisp",
    width: "10",
    height: "10",
    viewBox: "0 0 10 10",
    style: {
      flexShrink: 0
    }
  }, playing ? /*#__PURE__*/React.createElement("polygon", {
    points: "1,1 9,5 1,9",
    fill: "var(--ink)"
  }) : /*#__PURE__*/React.createElement(React.Fragment, null, /*#__PURE__*/React.createElement("rect", {
    x: "1",
    y: "1",
    width: "3",
    height: "8",
    fill: "var(--ink)"
  }), /*#__PURE__*/React.createElement("rect", {
    x: "6",
    y: "1",
    width: "3",
    height: "8",
    fill: "var(--ink)"
  }))), /*#__PURE__*/React.createElement("div", {
    style: {
      overflow: 'hidden'
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      fontSize: 'var(--text-road)',
      fontWeight: 700,
      whiteSpace: 'nowrap',
      overflow: 'hidden',
      textOverflow: 'ellipsis',
      lineHeight: 1
    }
  }, props.title || '—'), /*#__PURE__*/React.createElement("div", {
    style: {
      fontSize: 'var(--text-meta)',
      fontWeight: 400,
      whiteSpace: 'nowrap',
      overflow: 'hidden',
      textOverflow: 'ellipsis',
      lineHeight: 1
    }
  }, props.artist || '')));
}
Object.assign(__ds_scope, { MediaLine });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/readouts/MediaLine.jsx", error: String((e && e.message) || e) }); }

// ui_kits/hud/HudPanel.jsx
try { (() => {
function HudPanel(props) {
  const legend = props.legend || {
    prev: '',
    action: '',
    next: ''
  };
  return /*#__PURE__*/React.createElement("div", {
    className: "pixel-crisp",
    style: {
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
      lineHeight: 1
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      flex: 1,
      display: 'flex',
      flexDirection: 'column',
      gap: 2,
      overflow: 'hidden',
      minWidth: 0
    }
  }, React.Children.map(props.children, c => c && React.cloneElement(c, {
    style: {
      ...(c.props.style || {}),
      flexShrink: 0
    }
  }))), /*#__PURE__*/React.createElement("div", {
    style: {
      width: 1,
      alignSelf: 'stretch',
      background: 'var(--ink)'
    }
  }), /*#__PURE__*/React.createElement("div", {
    style: {
      width: 42,
      flexShrink: 0,
      display: 'flex',
      flexDirection: 'column',
      justifyContent: 'space-between',
      alignItems: 'flex-end',
      textAlign: 'right'
    }
  }, /*#__PURE__*/React.createElement("span", {
    style: {
      fontSize: 10,
      fontWeight: 700,
      color: 'var(--ink)'
    }
  }, legend.prev), /*#__PURE__*/React.createElement("span", {
    style: {
      fontSize: 10,
      fontWeight: 700,
      color: 'var(--ink)'
    }
  }, legend.action), /*#__PURE__*/React.createElement("span", {
    style: {
      fontSize: 10,
      fontWeight: 700,
      color: 'var(--ink)'
    }
  }, legend.next)));
}
window.HudPanel = HudPanel;
})(); } catch (e) { __ds_ns.__errors.push({ path: "ui_kits/hud/HudPanel.jsx", error: String((e && e.message) || e) }); }

// ui_kits/hud/MediaFocus.jsx
try { (() => {
const {
  ModeHeader,
  PixelDivider,
  MediaLine
} = window.MotoHUDDesignSystem_632360;
function MediaFocus(props) {
  const playing = props && props.playing != null ? props.playing : true;
  return /*#__PURE__*/React.createElement(HudPanel, {
    legend: {
      prev: 'SKIP',
      action: playing ? 'PAUSE' : 'PLAY',
      next: 'SKIP'
    }
  }, /*#__PURE__*/React.createElement(ModeHeader, {
    mode: "MEDIA",
    connected: true,
    heartbeat: true
  }), /*#__PURE__*/React.createElement(PixelDivider, {
    variant: "solid"
  }), /*#__PURE__*/React.createElement("div", {
    style: {
      flex: 1,
      display: 'flex',
      alignItems: 'center'
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      transform: 'scale(1.35)',
      transformOrigin: 'left center'
    }
  }, /*#__PURE__*/React.createElement(MediaLine, {
    playing: playing,
    title: "Night Drive",
    artist: "Field Tapes"
  }))));
}
window.MediaFocus = MediaFocus;
})(); } catch (e) { __ds_ns.__errors.push({ path: "ui_kits/hud/MediaFocus.jsx", error: String((e && e.message) || e) }); }

// ui_kits/hud/NavActiveNoRibbon.jsx
try { (() => {
const {
  ManeuverGlyph,
  DistanceReadout,
  ETAReadout,
  ModeHeader,
  PixelDivider,
  ProgressTicks
} = window.MotoHUDDesignSystem_632360;
function NavActiveNoRibbon() {
  return /*#__PURE__*/React.createElement(HudPanel, {
    legend: {
      prev: 'MODE',
      action: '—',
      next: 'MODE'
    }
  }, /*#__PURE__*/React.createElement(ModeHeader, {
    mode: "NAV",
    connected: true,
    heartbeat: true
  }), /*#__PURE__*/React.createElement(PixelDivider, {
    variant: "solid"
  }), /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      alignItems: 'center',
      gap: 10
    }
  }, /*#__PURE__*/React.createElement(ManeuverGlyph, {
    type: "left",
    size: 38
  }), /*#__PURE__*/React.createElement(DistanceReadout, {
    value: "350",
    unit: "m"
  })), /*#__PURE__*/React.createElement("div", {
    style: {
      fontSize: 'var(--text-road)',
      fontWeight: 400,
      color: 'var(--ink)',
      whiteSpace: 'nowrap',
      overflow: 'hidden',
      textOverflow: 'ellipsis',
      lineHeight: 1
    }
  }, "onto Harbor Blvd"), /*#__PURE__*/React.createElement("div", {
    style: {
      flex: 1
    }
  }), /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      justifyContent: 'space-between',
      alignItems: 'baseline'
    }
  }, /*#__PURE__*/React.createElement(ETAReadout, {
    etaMin: 12
  }), /*#__PURE__*/React.createElement(ProgressTicks, {
    total: 5,
    filled: 3
  })));
}
window.NavActiveNoRibbon = NavActiveNoRibbon;
})(); } catch (e) { __ds_ns.__errors.push({ path: "ui_kits/hud/NavActiveNoRibbon.jsx", error: String((e && e.message) || e) }); }

// ui_kits/hud/NavActiveRibbon.jsx
try { (() => {
const {
  ManeuverGlyph,
  DistanceReadout,
  ModeHeader,
  PixelDivider,
  RoadRibbon
} = window.MotoHUDDesignSystem_632360;
function NavActiveRibbon() {
  return /*#__PURE__*/React.createElement(HudPanel, {
    legend: {
      prev: 'MODE',
      action: '—',
      next: 'MODE'
    }
  }, /*#__PURE__*/React.createElement(ModeHeader, {
    mode: "NAV",
    connected: true,
    heartbeat: true
  }), /*#__PURE__*/React.createElement(PixelDivider, {
    variant: "solid"
  }), /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      alignItems: 'center',
      gap: 8
    }
  }, /*#__PURE__*/React.createElement(ManeuverGlyph, {
    type: "right",
    size: 34
  }), /*#__PURE__*/React.createElement(DistanceReadout, {
    value: "120",
    unit: "m"
  })), /*#__PURE__*/React.createElement("div", {
    style: {
      fontSize: 'var(--text-road)',
      fontWeight: 400,
      color: 'var(--ink)',
      whiteSpace: 'nowrap',
      overflow: 'hidden',
      textOverflow: 'ellipsis',
      lineHeight: 1
    }
  }, "Ridge Rd"), /*#__PURE__*/React.createElement("div", {
    style: {
      flex: 1
    }
  }), /*#__PURE__*/React.createElement(RoadRibbon, {
    points: [{
      x: 110,
      y: 0
    }, {
      x: 110,
      y: 22
    }, {
      x: 175,
      y: 34
    }],
    turnIndex: 1,
    height: 44
  }));
}
window.NavActiveRibbon = NavActiveRibbon;
})(); } catch (e) { __ds_ns.__errors.push({ path: "ui_kits/hud/NavActiveRibbon.jsx", error: String((e && e.message) || e) }); }

// ui_kits/hud/NavIdle.jsx
try { (() => {
const {
  ModeHeader,
  PixelDivider,
  ConnectionMark
} = window.MotoHUDDesignSystem_632360;
function NavIdle() {
  return /*#__PURE__*/React.createElement(HudPanel, {
    legend: {
      prev: 'MEDIA',
      action: '—',
      next: 'STATUS'
    }
  }, /*#__PURE__*/React.createElement(ModeHeader, {
    mode: "NAV",
    connected: true,
    heartbeat: true
  }), /*#__PURE__*/React.createElement(PixelDivider, {
    variant: "dashed"
  }), /*#__PURE__*/React.createElement("div", {
    style: {
      flex: 1,
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      justifyContent: 'center',
      gap: 6
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      fontSize: 17,
      fontWeight: 700,
      color: 'var(--ink)',
      letterSpacing: 1
    }
  }, "MOTO HUD"), /*#__PURE__*/React.createElement("div", {
    style: {
      fontSize: 'var(--text-road)',
      fontWeight: 400,
      color: 'var(--ink)'
    }
  }, "Waiting for route\u2026")));
}
window.NavIdle = NavIdle;
})(); } catch (e) { __ds_ns.__errors.push({ path: "ui_kits/hud/NavIdle.jsx", error: String((e && e.message) || e) }); }

// ui_kits/hud/NavMediaHybrid.jsx
try { (() => {
const {
  ManeuverGlyph,
  DistanceReadout,
  ModeHeader,
  PixelDivider,
  MediaLine
} = window.MotoHUDDesignSystem_632360;
function NavMediaHybrid() {
  return /*#__PURE__*/React.createElement(HudPanel, {
    legend: {
      prev: 'MODE',
      action: '—',
      next: 'MODE'
    }
  }, /*#__PURE__*/React.createElement(ModeHeader, {
    mode: "NAV",
    connected: true,
    heartbeat: true
  }), /*#__PURE__*/React.createElement(PixelDivider, {
    variant: "solid"
  }), /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      alignItems: 'center',
      gap: 8
    }
  }, /*#__PURE__*/React.createElement(ManeuverGlyph, {
    type: "straight",
    size: 32
  }), /*#__PURE__*/React.createElement(DistanceReadout, {
    value: "800",
    unit: "m"
  })), /*#__PURE__*/React.createElement("div", {
    style: {
      fontSize: 'var(--text-road)',
      fontWeight: 400,
      color: 'var(--ink)'
    }
  }, "Continue on Route 9"), /*#__PURE__*/React.createElement(PixelDivider, {
    variant: "dither"
  }), /*#__PURE__*/React.createElement("div", {
    style: {
      transform: 'scale(0.92)',
      transformOrigin: 'left center'
    }
  }, /*#__PURE__*/React.createElement(MediaLine, {
    playing: true,
    title: "Night Drive",
    artist: "Field Tapes"
  })));
}
window.NavMediaHybrid = NavMediaHybrid;
})(); } catch (e) { __ds_ns.__errors.push({ path: "ui_kits/hud/NavMediaHybrid.jsx", error: String((e && e.message) || e) }); }

// ui_kits/hud/StatusDiagnostics.jsx
try { (() => {
const {
  ModeHeader,
  PixelDivider,
  ConnectionMark
} = window.MotoHUDDesignSystem_632360;
function StatusDiagnostics() {
  const row = {
    display: 'flex',
    justifyContent: 'space-between',
    fontSize: 'var(--text-road)',
    color: 'var(--ink)'
  };
  return /*#__PURE__*/React.createElement(HudPanel, {
    legend: {
      prev: 'MODE',
      action: 'REDRAW',
      next: 'MODE'
    }
  }, /*#__PURE__*/React.createElement(ModeHeader, {
    mode: "STATUS",
    connected: true,
    heartbeat: true
  }), /*#__PURE__*/React.createElement(PixelDivider, {
    variant: "solid"
  }), /*#__PURE__*/React.createElement("div", {
    style: {
      flex: 1,
      display: 'flex',
      flexDirection: 'column',
      gap: 5,
      justifyContent: 'center'
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: row
  }, /*#__PURE__*/React.createElement("span", null, "LINK"), /*#__PURE__*/React.createElement(ConnectionMark, {
    connected: true,
    heartbeat: true,
    size: 12
  })), /*#__PURE__*/React.createElement("div", {
    style: row
  }, /*#__PURE__*/React.createElement("span", null, "LAST BEAT"), /*#__PURE__*/React.createElement("b", null, "0.4s")), /*#__PURE__*/React.createElement("div", {
    style: row
  }, /*#__PURE__*/React.createElement("span", null, "PACKETS"), /*#__PURE__*/React.createElement("b", null, "OK"))));
}
window.StatusDiagnostics = StatusDiagnostics;
})(); } catch (e) { __ds_ns.__errors.push({ path: "ui_kits/hud/StatusDiagnostics.jsx", error: String((e && e.message) || e) }); }

__ds_ns.FooterHints = __ds_scope.FooterHints;

__ds_ns.ModeHeader = __ds_scope.ModeHeader;

__ds_ns.PixelDivider = __ds_scope.PixelDivider;

__ds_ns.ConnectionMark = __ds_scope.ConnectionMark;

__ds_ns.ManeuverGlyph = __ds_scope.ManeuverGlyph;

__ds_ns.ProgressTicks = __ds_scope.ProgressTicks;

__ds_ns.RoadRibbon = __ds_scope.RoadRibbon;

__ds_ns.DistanceReadout = __ds_scope.DistanceReadout;

__ds_ns.ETAReadout = __ds_scope.ETAReadout;

__ds_ns.MediaLine = __ds_scope.MediaLine;

})();
