/**
 * Junction IR on `nav.junction` — replaces geographic `nav.minimap` polylines.
 * Pi draws idealized templates by `kind`; phone fills when possible, else Pi
 * synthesizes from `maneuver`. See protocol/README.md and ADR 0013.
 */

/** Same vocabulary as protocol `maneuver` (directional subset + u_turn). */
export type JunctionOutbound =
  | 'left'
  | 'right'
  | 'slight_left'
  | 'slight_right'
  | 'straight'
  | 'u_turn';

export type DriveSide = 'left' | 'right';

export type SideArm = {
  side: 'left' | 'right';
  at: 'before' | 'at' | 'after';
  style: 'dashed' | 'solid';
};

/** v1 implemented templates. */
export type JunctionKindV1 =
  | 'simple'
  | 't_junction'
  | 'crossroads'
  | 'fork'
  | 'merge'
  | 'dual_carriageway'
  | 'roundabout'
  | 'ramp_exit'
  | 'ramp_enter'
  | 'u_turn'
  | 'arrive'
  | 'depart';

/** Reserved — accept on wire; render as `simple` until a template exists. */
export type JunctionKindReserved = 'jughandle' | 'interchange' | 'gyratory';

export type JunctionKind = JunctionKindV1 | JunctionKindReserved | (string & {});

/** Shared frame: approach from bottom, ahead toward top. Proportions = renderer. */
export type JunctionBase = {
  kind: JunctionKind;
  /** Omit → `right` (continental US / most EU). */
  drive?: DriveSide;
  outbound: JunctionOutbound;
  /** Main corridor continues past the decision (T vs cross / ramp with mainline). */
  through: boolean;
  /** Extra arms along the approach / at the node. */
  sides?: SideArm[];
};

export type JunctionSimple = JunctionBase & { kind: 'simple' };
export type JunctionTJunction = JunctionBase & { kind: 't_junction'; through: false };
export type JunctionCrossroads = JunctionBase & { kind: 'crossroads'; through: true };
export type JunctionFork = JunctionBase & {
  kind: 'fork';
  outbound: 'slight_left' | 'slight_right' | 'left' | 'right';
};
export type JunctionMerge = JunctionBase & {
  kind: 'merge';
  /** Which side the slip joins from when not implied by outbound. */
  side?: 'left' | 'right';
};
export type JunctionDualCarriageway = JunctionBase & {
  kind: 'dual_carriageway';
  /** True when outbound is a hard left/right across the median (not slight/straight). */
  cross_median?: boolean;
};
export type JunctionRoundabout = JunctionBase & {
  kind: 'roundabout';
  /** Cap 2–6 on the diagram. */
  exits: number;
  /** 1-based exit index from entry, respecting `drive`. */
  exit: number;
};
export type JunctionRampExit = JunctionBase & {
  kind: 'ramp_exit';
  through: true;
};
export type JunctionRampEnter = JunctionBase & {
  kind: 'ramp_enter';
  side?: 'left' | 'right';
};
export type JunctionUTurn = JunctionBase & { kind: 'u_turn' };
export type JunctionArrive = JunctionBase & { kind: 'arrive' };
export type JunctionDepart = JunctionBase & { kind: 'depart' };

export type JunctionMessage =
  | JunctionSimple
  | JunctionTJunction
  | JunctionCrossroads
  | JunctionFork
  | JunctionMerge
  | JunctionDualCarriageway
  | JunctionRoundabout
  | JunctionRampExit
  | JunctionRampEnter
  | JunctionUTurn
  | JunctionArrive
  | JunctionDepart
  | JunctionBase;

/**
 * Pi fallback when phone omits `junction`: map maneuver → minimal IR.
 * Rich producers (Full Library / classifier) should prefer topology over this.
 */
export function synthesizeJunctionFromManeuver(
  maneuver: string,
  drive: DriveSide = 'right',
): JunctionMessage {
  switch (maneuver) {
    case 'arrive':
      return { kind: 'arrive', drive, outbound: 'straight', through: false };
    case 'depart':
      return { kind: 'depart', drive, outbound: 'straight', through: true };
    case 'u_turn':
      return { kind: 'u_turn', drive, outbound: 'u_turn', through: false };
    case 'roundabout':
      return { kind: 'roundabout', drive, outbound: 'right', through: false, exits: 4, exit: 2 };
    case 'slight_left':
      return { kind: 'fork', drive, outbound: 'slight_left', through: false };
    case 'slight_right':
      return { kind: 'fork', drive, outbound: 'slight_right', through: false };
    case 'straight':
      return { kind: 'simple', drive, outbound: 'straight', through: true };
    case 'left':
      return { kind: 'simple', drive, outbound: 'left', through: false };
    case 'right':
      return { kind: 'simple', drive, outbound: 'right', through: false };
    default:
      return { kind: 'simple', drive, outbound: 'straight', through: true };
  }
}
