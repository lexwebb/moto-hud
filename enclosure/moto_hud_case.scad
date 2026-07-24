// Moto HUD bench enclosure — Pi Zero + Inky pHAT clamshell
// Units: millimetres
//
// Usage (Customizer / CLI):
//   part = "assembly" | "base" | "lid"
// Preview: F5   Render + export STL: F6 → File → Export
// CLI:
//   openscad -D "part=\"base\"" -o exports/base.stl moto_hud_case.scad
//   openscad -D "part=\"lid\"" -o exports/lid.stl moto_hud_case.scad
//   openscad -D "part=\"assembly\"" -o exports/assembly.stl moto_hud_case.scad

/* [Part] */
part = "assembly"; // [assembly, base, lid]
explode = 0;       // [0:0.5:12] lid lift for assembly preview (mm)

/* [Board stack — caliper-check these] */
// Raspberry Pi Zero PCB footprint
board_w = 65.0;
board_d = 30.0;
// Pi PCB + header + Inky pHAT stack height (published Inky depth ~8.5; leave headroom)
stack_h = 14.0;
// Usable Inky glass (Pimoroni ~48.5 × 23.8)
window_w = 48.5;
window_d = 23.8;
// Window offset from board origin (Pi Zero corner = 0,0); tune after measuring Inky
window_ox = (board_w - window_w) / 2;
window_oy = (board_d - window_d) / 2 + 0.5;

/* [Enclosure] */
wall = 2.0;
floor_t = 2.0;
lid_t = 2.0;
clearance = 0.4;     // lateral fit around boards
print_tol = 0.3;     // extra cutout looseness
split_z = floor_t + stack_h; // cavity height inside base before lid

/* [Screws — Pi Zero M2.5 holes] */
// Hole centres from board corner (official Zero mechanical)
hole_inset = 3.5;
screw_d = 2.7;       // clearance for M2.5
boss_od = 6.0;
boss_id = 2.2;       // pilot / self-tap
boss_h = 3.0;

/* [USB / SD cutouts] */
// Pi Zero: micro-USB power on long edge near corner; Zero 2 W uses USB-C — widen if needed
usb_w = 10.0;
usb_h = 5.0;
sd_w = 14.0;
sd_h = 3.0;
sd_enable = true;

/* [Buttons — lid face, on right-hand side rail] */
button_count = 3;
button_d = 6.0;
button_pitch = 10.0;
button_rail = 10.0;  // extra case width on +X for buttons (SD stays on -X)
button_ox = 5.0;     // button centre X from outer right edge, inward
// button row centred on board depth (Y)

/* [Quality] */
$fn = 48;

// ---------------------------------------------------------------------------
eps = 0.02;

outer_w = board_w + 2 * (wall + clearance) + button_rail;
outer_d = board_d + 2 * (wall + clearance);
cavity_w_board = board_w + 2 * clearance;
cavity_d = board_d + 2 * clearance;

// Board XY origin: left/front inside walls; rail sits on +X
board_x = wall + clearance;
board_y = wall + clearance;

function hole_xy(i, j) = [
    hole_inset + i * (board_w - 2 * hole_inset),
    hole_inset + j * (board_d - 2 * hole_inset)
];

module at_board_xy() {
    translate([board_x, board_y, 0])
        children();
}

module screw_boss_solids(h = boss_h) {
    // Overlap floor slightly so CSG unions cleanly (avoid coplanar seams)
    translate([board_x, board_y, floor_t - eps])
        for (i = [0, 1], j = [0, 1]) {
            p = hole_xy(i, j);
            translate([p[0], p[1], 0])
                cylinder(h = h + eps, d = boss_od);
        }
}

module screw_boss_pilots(h = boss_h) {
    translate([board_x, board_y, floor_t - eps])
        for (i = [0, 1], j = [0, 1]) {
            p = hole_xy(i, j);
            translate([p[0], p[1], -eps])
                cylinder(h = h + 3 * eps, d = boss_id);
        }
}

module screw_holes_xy(z0, h, d) {
    translate([board_x, board_y, z0])
        for (i = [0, 1], j = [0, 1]) {
            p = hole_xy(i, j);
            translate([p[0], p[1], -eps])
                cylinder(h = h + 2 * eps, d = d);
        }
}

module usb_cutout() {
    // Through back wall (+Y), assuming Pi Zero power edge faces rear
    translate([
        board_x + board_w - usb_w - 1.5,
        outer_d - wall - clearance - eps,
        floor_t + 1.0
    ])
        cube([usb_w + print_tol, wall + clearance + 2 * eps, usb_h + print_tol]);
}

module sd_cutout() {
    if (sd_enable)
        translate([
            -eps,
            board_y + (board_d - sd_w) / 2,
            floor_t + 0.5
        ])
            cube([wall + clearance + 2 * eps, sd_w + print_tol, sd_h + print_tol]);
}

module base_shell() {
    // Cavity over the board only — solid rail under the side buttons
    difference() {
        union() {
            difference() {
                cube([outer_w, outer_d, split_z]);
                translate([wall, wall, floor_t])
                    cube([cavity_w_board, cavity_d, split_z - floor_t + eps]);
                usb_cutout();
                sd_cutout();
            }
            screw_boss_solids(boss_h);
        }
        screw_holes_xy(0, floor_t, screw_d + print_tol);
        screw_boss_pilots(boss_h);
    }
}

module lid_window_cut() {
    at_board_xy()
        translate([window_ox - print_tol / 2, window_oy - print_tol / 2, -eps])
            cube([
                window_w + print_tol,
                window_d + print_tol,
                lid_t + 2 * eps
            ]);
}

module lid_window_recess() {
    recess = 0.6;
    at_board_xy()
        translate([
            window_ox - 1 - print_tol / 2,
            window_oy - 1 - print_tol / 2,
            -eps
        ])
            cube([
                window_w + 2 + print_tol,
                window_d + 2 + print_tol,
                recess + eps
            ]);
}

module button_holes() {
    // Along +X side rail, centred on board depth
    span = (button_count - 1) * button_pitch;
    x = outer_w - button_ox;
    y0 = board_y + (board_d - span) / 2;
    for (n = [0 : button_count - 1]) {
        translate([x, y0 + n * button_pitch, -eps])
            cylinder(h = lid_t + 2 * eps, d = button_d + print_tol);
    }
}

module lid_shell() {
    lip_h = 1.2;
    lip_inset = 0.15;
    // Lip only around the board cavity (not the solid side rail)
    lip_w = cavity_w_board - 2 * lip_inset;
    lip_d = cavity_d - 2 * lip_inset;

    difference() {
        union() {
            cube([outer_w, outer_d, lid_t]);
            translate([wall + lip_inset, wall + lip_inset, -lip_h])
                cube([lip_w, lip_d, lip_h + eps]);
        }
        translate([
            wall + lip_inset + 0.8,
            wall + lip_inset + 0.8,
            -lip_h - eps
        ])
            cube([
                lip_w - 1.6,
                lip_d - 1.6,
                lip_h + 2 * eps
            ]);
        lid_window_cut();
        lid_window_recess();
        button_holes();
        screw_holes_xy(-lip_h, lid_t + lip_h, screw_d + print_tol);
    }
}

module assembly() {
    base_shell();
    translate([0, 0, split_z + explode])
        lid_shell();
}

if (part == "base")
    base_shell();
else if (part == "lid")
    // Sit lid flat on XY for printing / STL export
    translate([0, 0, 1.2])
        lid_shell();
else
    assembly();
