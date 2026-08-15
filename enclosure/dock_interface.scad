// Dock mating contract — bike plate ↔ HUD pod
// Units: millimetres. Owned by enclosure/DOCK.md / ADR 0014.
//
// Plate: rounded plinth + U-shaped end caps that rise into matching
// rebates on the pod. Long-side middles are open for a gloved grip.
// Magnets live under the caps; pogo sits in the open middle.
// USB-C sinks into the underside centre of the plate (not a side jack).
//
//   openscad -o exports/dock_interface.stl dock_interface.scad
//   openscad -D 'iface_part="plate"' -o exports/dock_plate_slab.stl dock_interface.scad
//   openscad -D 'iface_part="pod"'   -o exports/dock_pod_slab.stl dock_interface.scad

/* [Preview] */
iface_part = "assembly"; // [assembly, plate, pod, plate_usb]
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
brim = 1.6;          // floor extra outside the pod (rounded plinth)
wall_t = 2.4;        // end-cap thickness = pod rebate depth
wrap_h = 9.0;        // how far caps rise up the pod
wrap_x = 18.0;       // remaining side-wall length from each short end (mm)
fit = 0.35;          // extra rebate vs wall (print + paint)
round_r = 2.5;       // rounded bevel on every outer edge / corner
plate_t = 13.0;      // thick enough for underside USB-C + pogo well + septum
scoop_r = 5.5;       // finger undercut in the grip gap

/* [Magnets — 6×3 mm N52 discs, checkerboard polarity] */
mag_d = 6.0;
mag_h = 3.0;
mag_pocket_clear = 0.2;
mag_pocket_extra_z = 0.2;
mag_inset_x0 = 16.0;
mag_inset_x1 = 7.0;
mag_inset_y = 7.0;

/* [USB-C inlet — underside centre of the plate. Sink from bike accessory port.] */
usb_c_w = 9.2;       // receptacle opening
usb_c_h = 3.5;
usb_c_depth = 7.5;   // into the floor from z=0 (bottom)
usb_c_shell_w = 12.5;
usb_c_shell_d = 7.0;

/* [Pogo well — keyed 3-pin magnetic housing, caliper then retune] */
pogo_well_d = 14.0;
pogo_well_clear = 0.3;
pogo_well_depth = 3.5; // from the mating face down; does not break into USB-C
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
$fn = 32;
bevel_fn = 48;

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

// Rounded box: fillet every edge and corner (hull of 8 corner spheres).
module round_cube(w, d, h, r) {
    rr = min(r, w / 2 - 0.12, d / 2 - 0.12, h / 2 - 0.12);
    hull() {
        for (x = [rr, w - rr], y = [rr, d - rr], z = [rr, h - rr])
            translate([x, y, z])
                sphere(r = rr, $fn = bevel_fn);
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
    rr = min(0.9, d / 2 - 0.15, h / 2 - 0.15);
    translate([-eps, -eps, -eps])
        round_cube(d + eps, pod_d + 2 * eps, h + eps, rr);
    translate([-eps, -eps, -eps])
        round_cube(wx + eps, d + eps, h + eps, rr);
    translate([-eps, pod_d - d, -eps])
        round_cube(wx + eps, d + eps, h + eps, rr);
    translate([pod_w - d, -eps, -eps])
        round_cube(d + eps, pod_d + 2 * eps, h + eps, rr);
    translate([pod_w - wx, -eps, -eps])
        round_cube(wx + eps, d + eps, h + eps, rr);
    translate([pod_w - wx, pod_d - d, -eps])
        round_cube(wx + eps, d + eps, h + eps, rr);
}

module pod_window_cut() {
    recess = 0.8;
    ox = wall + clearance + window_ox;
    oy = wall + clearance + window_oy;
    translate([ox, oy, pod_h - recess])
        round_cube(window_w, window_d, recess + eps, 0.6);
}

module pod_dock_slab() {
    difference() {
        round_cube(pod_w, pod_d, pod_h, round_r);
        pod_u_rebates();
        mag_pockets(mag_h + mag_pocket_extra_z);
        pogo_well_cut(mag_h + mag_pocket_extra_z + 1.2);
        pod_window_cut();
    }
}

// USB-C plug enters from below, dead-centre of the plate floor.
module usb_c_inlet_cut() {
    translate([
        (plate_w - usb_c_shell_w) / 2,
        (plate_d - usb_c_shell_d) / 2,
        -eps
    ])
        round_cube(usb_c_shell_w, usb_c_shell_d, usb_c_depth - 1.0 + eps, 0.8);
    translate([
        (plate_w - usb_c_w) / 2,
        (plate_d - usb_c_h) / 2,
        -eps
    ])
        round_cube(usb_c_w, usb_c_h, usb_c_depth + eps, 0.9);
}

module grip_cutouts() {
    cut_d = brim + wall_t + 2.5;
    cut_h = wrap_h + round_r + 2;
    x0 = brim + wrap_x;
    for (y0 = [-1.2, plate_d - cut_d + 1.2]) {
        translate([x0, y0, plate_t - 0.2])
            round_cube(grip_w, cut_d, cut_h, round_r);
    }
}

module finger_scoops() {
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
        round_cube(plate_w, plate_d, well_h, round_r);
        // Well: inner cavity sized to the pod rebate (walls fill the U-insets)
        translate([brim + wall_t, brim + wall_t, plate_t])
            round_cube(inner_w, inner_d, wrap_h + round_r + 1, 0.9);
        grip_cutouts();
        finger_scoops();
        translate([brim, brim, plate_t - (mag_h + mag_pocket_extra_z)])
            mag_pockets(mag_h + mag_pocket_extra_z + eps);
        // Pogo from the mating face down only — septum above the USB-C pocket
        translate([brim, brim, plate_t - pogo_well_depth])
            pogo_well_cut(pogo_well_depth + wrap_h + 2);
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
else if (iface_part == "plate_usb")
    // Flip so the underside USB-C opening faces +Z for the viewer.
    translate([0, plate_d, plate_t + wrap_h])
        rotate([180, 0, 0])
            bike_plate_slab();
else
    assembly();
