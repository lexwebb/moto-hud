// Dock mating contract — bike plate ↔ HUD pod underside
// Units: millimetres. Owned by enclosure/DOCK.md / ADR 0014.
//
// Include from the pod floor and the bike plate so magnet XY, pogo well,
// tray fence, and the plate USB-C inlet cannot drift. This file is a
// previewable slab; it does not replace moto_hud_case.scad (bench clamshell)
// until the pod underside is cut to match.
//
//   openscad -o exports/dock_interface.stl dock_interface.scad
//   openscad -D 'iface_part="plate"' -o exports/dock_plate_slab.stl dock_interface.scad
//   openscad -D 'iface_part="pod"'   -o exports/dock_pod_slab.stl dock_interface.scad

/* [Preview] */
iface_part = "assembly"; // [assembly, plate, pod]
iface_explode = 8;       // [0:1:20] plate/pod separation for assembly preview (mm)

/* [Pod footprint — must match moto_hud_case.scad outer_*] */
// Duplicated (not include) so this file stays standalone until the case is split.
board_w = 65.0;
board_d = 30.0;
wall = 2.0;
clearance = 0.4;
button_rail = 12.0;
outer_w = board_w + 2 * (wall + clearance) + button_rail;
outer_d = board_d + 2 * (wall + clearance);

/* [Tray] */
fence_t = 2.0;     // wall around the pod
fence_h = 3.0;     // shear capture; magnets still do Z
fence_fit = 0.4;   // extra opening vs pod outer (print + paint)
plate_t = 4.0;     // floor thicker than mag_h so pockets do not break through

/* [Magnets — 6×3 mm N52 discs, checkerboard polarity] */
mag_d = 6.0;
mag_h = 3.0;
mag_pocket_clear = 0.2; // extra ID
mag_pocket_extra_z = 0.2;
// Centres from pod outer edges. −X (SD) is past the Pi M2.5 boss (centre ~5.9 mm);
// +X sits in the button rail, clear of the board.
mag_inset_x0 = 16.0;
mag_inset_x1 = 7.0;
mag_inset_y = 7.0;

/* [USB-C inlet — plate only, +Y edge. Sink from bike accessory port.] */
usb_c_block_w = 16.0;
usb_c_block_d = 10.0;
usb_c_block_h = 8.0;
usb_c_w = 9.2;   // receptacle opening; caliper the real shell
usb_c_h = 3.5;
usb_c_z = 2.0;   // opening bottom above plate z=0

/* [Pogo well — keyed 3-pin magnetic housing, caliper then retune] */
pogo_well_d = 14.0;
pogo_well_clear = 0.3;
// Spring pins live on the plate, recessed below the fence top
pogo_recess = 1.0;
// 3 pads on the pod, 2.54 mm pitch, along +X (long axis), centred
pogo_pitch = 2.54;
pogo_pad_d = 2.0;
pogo_pin_d = 1.5;

/* [Quality] */
$fn = 48;

eps = 0.02;

pod_w = outer_w;
pod_d = outer_d;

plate_w = pod_w + 2 * fence_t + 2 * fence_fit;
plate_d = pod_d + 2 * fence_t + 2 * fence_fit;

// Pod-local XY: origin at pod outer corner, same as moto_hud_case.scad.
function mag_xy(i, j) = [
    i == 0 ? mag_inset_x0 : pod_w - mag_inset_x1,
    mag_inset_y + j * (pod_d - 2 * mag_inset_y)
];

// Checkerboard: (0,0)=N, (1,0)=S, (0,1)=S, (1,1)=N — 180° is all-repel.
function mag_north(i, j) = (i + j) % 2 == 0;

function pogo_center() = [pod_w / 2, pod_d / 2];

function pogo_pad_xy(n) = [
    pogo_center()[0] + (n - 1) * pogo_pitch, // n = 0,1,2 → pins 1,2,3
    pogo_center()[1]
];

module mag_pockets(h, extra_r = 0) {
    for (i = [0, 1], j = [0, 1]) {
        p = mag_xy(i, j);
        translate([p[0], p[1], -eps])
            cylinder(h = h + 2 * eps, d = mag_d + mag_pocket_clear + extra_r);
    }
}

module pogo_well_cut(h) {
    p = pogo_center();
    translate([p[0], p[1], -eps])
        cylinder(h = h + 2 * eps, d = pogo_well_d + pogo_well_clear);
}

module pogo_pads_2d() {
    for (n = [0, 1, 2]) {
        p = pogo_pad_xy(n);
        translate([p[0], p[1], 0])
            circle(d = pogo_pad_d);
    }
}

// --- Pod dock face (underside of the HUD). z=0 is the outer floor (mates plate).
module pod_dock_slab() {
    t = mag_h + mag_pocket_extra_z + 0.8; // 0.8 mm skin over magnets if we encapsulate
    difference() {
        cube([pod_w, pod_d, t]);
        mag_pockets(mag_h + mag_pocket_extra_z);
        // Pads sit in a shallow well so the housing can enter
        pogo_well_cut(t);
    }
}

module usb_c_inlet_cut() {
    // Through the +Y face of the inlet block (cable plugs from outside).
    translate([
        (plate_w - usb_c_w) / 2,
        plate_d - eps,
        usb_c_z
    ])
        cube([usb_c_w, usb_c_block_d + 2 * eps, usb_c_h]);
}

module usb_c_inlet_block() {
    translate([
        (plate_w - usb_c_block_w) / 2,
        plate_d - eps,
        0
    ])
        cube([usb_c_block_w, usb_c_block_d + eps, usb_c_block_h]);
}

// --- Bike plate: floor + fence; springs in the well, recessed.
module bike_plate_slab() {
    well_h = plate_t + fence_h;
    difference() {
        union() {
            cube([plate_w, plate_d, plate_t]);
            // Fence
            difference() {
                translate([0, 0, plate_t - eps])
                    cube([plate_w, plate_d, fence_h + eps]);
                translate([
                    fence_t,
                    fence_t,
                    plate_t - 2 * eps
                ])
                    cube([
                        pod_w + 2 * fence_fit,
                        pod_d + 2 * fence_fit,
                        fence_h + 3 * eps
                    ]);
            }
            usb_c_inlet_block();
        }
        // Magnet pockets in the floor, opening toward the pod (up)
        translate([
            fence_t + fence_fit,
            fence_t + fence_fit,
            plate_t - (mag_h + mag_pocket_extra_z)
        ])
            mag_pockets(mag_h + mag_pocket_extra_z + eps);
        // Pogo well through floor; recess measured from fence top
        translate([
            fence_t + fence_fit,
            fence_t + fence_fit,
            -eps
        ])
            pogo_well_cut(well_h + 2 * eps);
        usb_c_inlet_cut();
    }
}

module assembly() {
    bike_plate_slab();
    translate([
        fence_t + fence_fit,
        fence_t + fence_fit,
        plate_t + iface_explode
    ])
        pod_dock_slab();
}

if (iface_part == "plate")
    bike_plate_slab();
else if (iface_part == "pod")
    pod_dock_slab();
else
    assembly();
