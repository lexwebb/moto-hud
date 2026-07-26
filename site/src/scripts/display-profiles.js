/** Display profiles for the ride emulator (mirrors -host backends). */

export const HUD_W = 250;
export const HUD_H = 122;

/**
 * Inky pHAT 4-colour (PIM784 / JD79661 Spectra BWRY).
 * Pimoroni quotes ~20s full refresh. BWRY OTP waveforms run
 * inverse → flashing (red/yellow particle drive) → imaging even when
 * the framebuffer is only black/white.
 * Stage sum 12100ms + fade 7900ms ≈ 20s.
 */
const INKY_4COLOUR_SEQUENCE = [
  // Inverse
  { fill: '#000000', ms: 700 },
  { fill: '#ffffff', ms: 700 },
  { fill: '#000000', ms: 600 },
  { fill: '#ffffff', ms: 600 },
  // Flashing — coloured particles visible as full-field flashes
  { fill: '#c41e3a', ms: 1600 },
  { fill: '#000000', ms: 500 },
  { fill: '#ffffff', ms: 500 },
  { fill: '#f0c400', ms: 1800 },
  { fill: '#000000', ms: 500 },
  { fill: '#c41e3a', ms: 1200 },
  { fill: '#ffffff', ms: 600 },
  { fill: '#f0c400', ms: 1200 },
  { fill: '#000000', ms: 700 },
  { fill: '#ffffff', ms: 900 },
];

/**
 * Waveshare 2.13″ B/W V3/V4 full refresh (~2s).
 * Spec: full flickers several times; fast flashes once (~1.8s);
 * partial has no flicker (~0.3s). Full after every ~5 partials
 * to clear ghosting (Waveshare V4 datasheet).
 * Stage sum 1300ms + fade 700ms = 2000ms.
 */
const WAVESHARE_FULL_SEQUENCE = [
  { fill: '#000000', ms: 200 },
  { fill: '#ffffff', ms: 200 },
  { fill: '#000000', ms: 250 },
  { fill: '#ffffff', ms: 250 },
  { fill: '#000000', ms: 200 },
  { fill: '#ffffff', ms: 200 },
];

export const DEVICE_PROFILES = {
  inky: {
    id: 'inky',
    label: 'Inky pHAT (4-colour)',
    kind: 'eink',
    panelW: HUD_W,
    panelH: HUD_H,
    sequence: INKY_4COLOUR_SEQUENCE,
    fadeMs: 7900,
    distanceStep: true,
    letterbox: false,
    hint: '4-colour BWRY · ~20s (inverse → R/Y flash → settle) · ≈ 50 m steps · HUD stays 1-bit',
  },
  waveshare: {
    id: 'waveshare',
    label: 'Waveshare 2.13″ B/W',
    kind: 'eink',
    panelW: HUD_W,
    panelH: HUD_H,
    sequence: WAVESHARE_FULL_SEQUENCE,
    fadeMs: 700,
    supportPartial: true,
    partialMs: 300,
    fullEveryN: 5,
    distanceStep: true,
    letterbox: false,
    hint: 'B/W · partial ~0.3s (no flicker) · full ~2s every 5 · ≈ 50 m steps',
  },
  lcd: {
    id: 'lcd',
    label: 'Display HAT Mini (LCD)',
    kind: 'lcd',
    panelW: 320,
    panelH: 240,
    flashMs: 0,
    fadeMs: 0,
    distanceStep: false,
    letterbox: true,
    hint: 'instant frames · 250×122 letterboxed on 320×240 · no e-ink gate',
  },
};

export const DEFAULT_DEVICE = 'waveshare';
export const DEVICE_STORAGE_KEY = 'moto-hud-emulator-device';

export function loadDeviceId() {
  try {
    const id = localStorage.getItem(DEVICE_STORAGE_KEY);
    if (id && DEVICE_PROFILES[id]) return id;
  } catch {
    /* ignore */
  }
  return DEFAULT_DEVICE;
}

export function saveDeviceId(id) {
  try {
    localStorage.setItem(DEVICE_STORAGE_KEY, id);
  } catch {
    /* ignore */
  }
}

export function getProfile(id) {
  return DEVICE_PROFILES[id] || DEVICE_PROFILES[DEFAULT_DEVICE];
}

/** Typical wipe+settle duration for an e-ink profile (ms). Prefers partial when available. */
export function profileRefreshMs(profile) {
  if (!profile || profile.kind !== 'eink') return 0;
  if (profile.supportPartial && profile.partialMs != null) {
    return profile.partialMs;
  }
  if (profile.sequence?.length) {
    const flash = profile.sequence.reduce((s, step) => s + step.ms, 0);
    return flash + (profile.fadeMs || 0);
  }
  return (profile.flashMs ?? 0) * 2 + (profile.fadeMs ?? 0);
}

/** Full-refresh duration (ms) when the profile distinguishes partial vs full. */
export function profileFullRefreshMs(profile) {
  if (!profile || profile.kind !== 'eink') return 0;
  if (profile.sequence?.length) {
    const flash = profile.sequence.reduce((s, step) => s + step.ms, 0);
    return flash + (profile.fadeMs || 0);
  }
  return (profile.flashMs ?? 0) * 2 + (profile.fadeMs ?? 0);
}
