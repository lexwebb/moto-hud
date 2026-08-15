// Dock mating contract — bike plate ↔ HUD pod
// Units: millimetres. Owned by enclosure/DOCK.md / ADR 0014.
//
// Plate: chamfered plinth + U-shaped end caps that rise into matching
// rebates on the pod. Long-side middles are open for a gloved grip.
// Magnets live under the caps; pogo sits in the open middle.
//
//   openscad -o exports/dock_interface.stl dock_interface.scad
//   openscad -D 'iface_part="plate"' -o exports/dock_plate_slab.stl dock_interface.scad
//   openscad -D 'iface_part="pod"'   -o exports/dock_pod_slab.stl dock_interface.scad

/* [Preview] */
iface_part = "assembly"; // [assembly, plate, pod]
iface_explode = 10;      // [0:1:24] plate/pod separation (mm)

/* [Pod footprint — must match moto_hud_case.scad outer_*] */
board_w = 65.0;
board_d = 30.0;
wall = 2.0;
clearance = 0.4;
button_rail = 12.0;
outer_w = board_w + 2 * (wall + clearance) + button_rail;
outer_d = board_d + 2 * (wall + clearance);

/* [Cradle] */
brim = 1.6;          // floor extra outside the pod (chamfered plinth)
wall_t = 2.4;        // end-cap thickness = pod rebate depth
wrap_h = 9.0;        // how far caps rise up the pod
wrap_x = 18.0;       // remaining side-wall length from each short end (mm)
fit = 0.35;          // extra rebate vs wall (print + paint)
chamfer = 1.2;       // every outer edge / corner
plate_t = 4.0;       // floor thicker than mag_h so pockets do not break through
scoop_r = 5.5;       // finger undercut in the grip gap

/* [Magnets — 6×3 mm N52 discs, checkerboard polarity] */
mag_d = 6.0;
mag_h = 3.0;
mag_pocket_clear = 0.2;
mag_pocket_extra_z = 0.2;
// Centres from pod outer edges. −X (SD) past the Pi M2.5 boss; +X in the rail.
mag_inset_x0 = 16.0;
mag_inset_x1 = 7.0;
mag_inset_y = 7.0;

/* [USB-C inlet — plate +X end cap. Sink from bike accessory port.] */
usb_c_block_w = 16.0;
usb_c_block_d = 10.0;
usb_c_block_h = 8.0;
usb_c_w = 9.2;
usb_c_h = 3.5;
usb_c_z = 2.0;

/* [Pogo well — keyed 3-pin magnetic housing, caliper then retune] */
pogo_well_d = 14.0;
pogo_well_clear = 0.3;
pogo_recess = 1.0;
pogo_pitch = 2.54;
pogo_pad_d = 2.0;
pogo_pin_d = 1.5;

/* [Pod preview body] */
pod_h = 16.0;
window_w = 48.5;
window_d = 23.8;
window_ox = (board_w - window_w) / 2;
window_oy = (board_d - window_d) / 2 + 0.5;

/* [Quality] */
$fn = 36;

eps = 0.02;

pod_w = outer_w;
pod_d = outer_d;

plate_w = pod_w + 2 * brim;
plate_d = pod_d + 2 * brim;

grip_w = pod_w - 2 * wrap_x;

function mag_xy(i, j) = [
    i == 0 ? mag_inset_x0 : pod_w - mag_inset_x1,
    mag_inset_y + j * (pod_d - 2 * mag_inset_y)
];

function mag_north(i, j) = (i + j) % 2 == 0;

function pogo_center() = [pod_w / 2, pod_d / 2];

function pogo_pad_xy(n) = [
    pogo_center()[0] + (n - 1) * pogo_pitch,
    pogo_center()[1]
];

// Chamfered box: 45° on every edge and corner (hull of three slabs + caps).
module chamfer_cube(w, d, h, c) {
    cc = min(c, w / 2 - 0.15, d / 2 - 0.15, h / 2 - 0.15);
    hull() {
        translate([cc, cc, 0])
            cube([w - 2 * cc, d - 2 * cc, 0.05]);
        translate([0, cc, cc])
            cube([w, d - 2 * cc, h - 2 * cc]);
        translate([cc, 0, cc])
            cube([w - 2 * cc, d, h - 2 * cc]);
        translate([cc, cc, h - 0.05])
            cube([w - 2 * cc, d - 2 * cc, 0.05]);
    }
}

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

// U-rebate on the pod: step-in at each short end + wrap along the long sides.
module pod_u_rebates() {
    d = wall_t + fit;
    h = wrap_h + fit;
    wx = wrap_x + fit;
    // −X end + wraps
    translate([-eps, -eps, -eps])
        cube([d + eps, pod_d + 2 * eps, h + eps]);
    translate([-eps, -eps, -eps])
        cube([wx + eps, d + eps, h + eps]);
    translate([-eps, pod_d - d, -eps])
        cube([wx + eps, d + eps, h + eps]);
    // +X end + wraps
    translate([pod_w - d, -eps, -eps])
        cube([d + eps, pod_d + 2 * eps, h + eps]);
    translate([pod_w - wx, -eps, -eps])
        cube([wx + eps, d + eps, h + eps]);
    translate([pod_w - wx, pod_d - d, -eps])
        cube([wx + eps, d + eps, h + eps]);
}

module pod_window_cut() {
    recess = 0.8;
    ox = wall + clearance + window_ox;
    oy = wall + clearance + window_oy;
    translate([ox, oy, pod_h - recess])
        cube([window_w, window_d, recess + eps]);
}

module pod_dock_slab() {
    difference() {
        chamfer_cube(pod_w, pod_d, pod_h, chamfer);
        pod_u_rebates();
        mag_pockets(mag_h + mag_pocket_extra_z);
        pogo_well_cut(mag_h + mag_pocket_extra_z + 1.2);
        pod_window_cut();
    }
}

module usb_c_inlet_block() {
    translate([
        plate_w - eps,
        (plate_d - usb_c_block_w) / 2,
        0
    ])
        chamfer_cube(usb_c_block_d + eps, usb_c_block_w, usb_c_block_h, 0.8);
}

module usb_c_inlet_cut() {
    translate([
        plate_w - usb_c_block_d - eps,
        (plate_d - usb_c_w) / 2,
        usb_c_z
    ])
        cube([usb_c_block_d + 2 * eps, usb_c_w, usb_c_h]);
}

module grip_cutouts() {
    // Chamfered cutters so the remaining arm ends are 45°, not square.
    cut_d = brim + wall_t + 2.5;
    cut_h = wrap_h + chamfer + 2;
    x0 = brim + wrap_x;
    for (y0 = [-1.2, plate_d - cut_d + 1.2]) {
        translate([x0, y0, plate_t - 0.2])
            chamfer_cube(grip_w, cut_d, cut_h, chamfer);
    }
}

module finger_scoops() {
    // Undercut the plinth in the grip gap so a glove can hook the pod ridge.
    for (y = [0, plate_d]) {
        translate([plate_w / 2, y, plate_t])
            rotate([0, 90, 0])
                cylinder(h = grip_w - 2, r = scoop_r, center = true);
    }
}

module bike_plate_slab() {
    well_h = plate_t + wrap_h;
    inner_w = pod_w - 2 * wall_t;
    inner_d = pod_d - 2 * wall_t;
    difference() {
        union() {
            // One chamfered solid: floor + full-height walls, then we notch grips.
            chamfer_cube(plate_w, plate_d, well_h, chamfer);
            usb_c_inlet_block();
        }
        // Well: inner cavity sized to the pod rebate (walls fill the U-insets)
        translate([brim + wall_t, brim + wall_t, plate_t])
            chamfer_cube(inner_w, inner_d, wrap_h + chamfer + 1, 0.8);
        // Open the long-side middles for grip
        grip_cutouts();
        finger_scoops();
        // Magnet pockets in the floor, opening toward the pod (up)
        translate([brim, brim, plate_t - (mag_h + mag_pocket_extra_z)])
            mag_pockets(mag_h + mag_pocket_extra_z + eps);
        // Pogo well through floor (open middle, between the caps)
        translate([brim, brim, -eps])
            pogo_well_cut(well_h + 2 * eps);
        usb_c_inlet_cut();
    }
}

module assembly() {
    bike_plate_slab();
    translate([brim, brim, plate_t + iface_explode])
        pod_dock_slab();
}

if (iface_part == "plate")
    bike_plate_slab();
else if (iface_part == "pod")
    pod_dock_slab();
else
    assembly();
