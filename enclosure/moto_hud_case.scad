// Moto HUD bench enclosure — Pi Zero + display HAT clamshell
// Units: millimetres
//
// On-bike dock (plate + magnets + pogo) is planned separately:
//   DOCK.md, dock_interface.scad, docs/adr/0014-two-part-magnetic-pogo-dock.md
// Do not cut this floor to the dock face until those pockets are calipered.
//
// Usage (Customizer / CLI):
//   part = "assembly" | "base" | "lid" | "caps"
// Preview: F5   Render + export STL: F6 → File → Export
// CLI:
//   openscad -D 'part="base"' -o exports/base.stl moto_hud_case.scad
//   openscad -D 'part="lid"' -o exports/lid.stl moto_hud_case.scad
//   openscad -D 'part="caps"' -o exports/caps.stl moto_hud_case.scad
//   openscad -D 'part="assembly"' -o exports/assembly.stl moto_hud_case.scad

/* [Part] */
part = "assembly"; // [assembly, base, lid, caps]
explode = 0;       // [0:0.5:12] lid lift for assembly preview (mm)
cap_index = -1;    // [-1:1:2] caps export: -1 = all three on a sprue; 0/1/2 = one cap

/* [Board stack — caliper-check these] */
// Raspberry Pi Zero PCB footprint
board_w = 65.0;
board_d = 30.0;
// Pi PCB + header + HAT stack height (Inky ~8.5; leave headroom for Waveshare too)
stack_h = 14.0;
// Usable glass (Inky / Waveshare 2.13″ ~48.5 × 23.8)
window_w = 48.5;
window_d = 23.8;
// Window offset from board origin (Pi Zero corner = 0,0); tune after measuring HAT
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
// Pi Zero W: micro-USB PWR on long edge near corner (bench). On-bike 5 V is USB-C into the plate.
usb_w = 10.0;
usb_h = 5.0;
sd_w = 14.0;
sd_h = 3.0;
sd_enable = true;

/* [Buttons — lid face, on right-hand side rail] */
// cap = printed mushroom + 6×6 tactile wells; panel8 = 8 mm high-head panel holes
button_mode = "cap"; // [cap, panel8]
button_count = 3;
button_pitch = 12.0; // centres; leave clearance between ~11 mm caps
button_rail = 12.0;  // extra case width on +X for buttons (SD stays on -X)
button_ox = 6.0;     // button centre X from outer right edge, inward
// Order along +Y (rail on right, screen upright): Prev, Next, Action

// Cap mode (Path A′) — glove target is the print; electrical is the tactile
cap_od = 11.0;
cap_dome_h = 2.5;
stem_d = 3.6;
stem_clear = 0.35;
stem_len = 5.5;      // through lid + travel into tactile
flange_d = 5.2;      // retention flange under dome (wider than shaft)
flange_h = 0.8;
retention_h = 1.0;   // underside counterbore depth in lid
travel = 1.2;
tactile_foot = 6.2;  // 6×6 body + fit
tactile_h = 3.5;     // body height — caliper after purchase and retune
well_depth = 2.0;
wire_ch_w = 2.5;
wire_ch_h = 2.0;

// panel8 fallback
button_d = 8.2;

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

shaft_d = stem_d + print_tol + stem_clear;
rail_x0 = wall + cavity_w_board; // start of solid +X rail (inside cavity wall)

function hole_xy(i, j) = [
    hole_inset + i * (board_w - 2 * hole_inset),
    hole_inset + j * (board_d - 2 * hole_inset)
];

// Button centres on the +X rail; n = 0..button_count-1 → Prev, Next, Action
function button_xy(n) = [
    outer_w - button_ox,
    board_y + (board_d - (button_count - 1) * button_pitch) / 2 + n * button_pitch
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

// --- Buttons ----------------------------------------------------------------

module button_panel8_holes() {
    for (n = [0 : button_count - 1]) {
        p = button_xy(n);
        translate([p[0], p[1], -eps])
            cylinder(h = lid_t + 2 * eps, d = button_d + print_tol);
    }
}

module button_cap_shafts() {
    // Lid coords: z=0 underside (toward cavity), z=lid_t outer face.
    // Cap drops in from outside: flange seats in outer counterbore; stem hangs toward tactile.
    for (n = [0 : button_count - 1]) {
        p = button_xy(n);
        translate([p[0], p[1], -eps])
            cylinder(h = lid_t + 2 * eps, d = shaft_d);
        translate([p[0], p[1], lid_t - retention_h])
            cylinder(h = retention_h + eps, d = flange_d + print_tol);
    }
}

module tactile_wells() {
    for (n = [0 : button_count - 1]) {
        p = button_xy(n);
        translate([
            p[0] - tactile_foot / 2,
            p[1] - tactile_foot / 2,
            split_z - well_depth
        ])
            cube([tactile_foot, tactile_foot, well_depth + eps]);
    }
}

module wire_channels() {
    // Channel from each well toward the board cavity (−X), open at top of rail
    for (n = [0 : button_count - 1]) {
        p = button_xy(n);
        ch_x0 = rail_x0 - eps;
        ch_len = p[0] - tactile_foot / 2 - ch_x0 + eps;
        if (ch_len > 0)
            translate([
                ch_x0,
                p[1] - wire_ch_w / 2,
                split_z - wire_ch_h
            ])
                cube([ch_len, wire_ch_w, wire_ch_h + eps]);
    }
}

module button_cap_solid() {
    // Origin at flange top (= lid outer face when seated). Dome +Z; stem −Z through lid.
    union() {
        // Domed head
        intersection() {
            translate([0, 0, -cap_od / 2 + cap_dome_h])
                sphere(d = cap_od);
            translate([-cap_od, -cap_od, 0])
                cube([2 * cap_od, 2 * cap_od, cap_dome_h + eps]);
        }
        // Flange seats in outer counterbore
        translate([0, 0, -flange_h])
            cylinder(h = flange_h + eps, d = flange_d);
        // Stem through lid toward tactile
        translate([0, 0, -stem_len])
            cylinder(h = stem_len, d = stem_d);
    }
}

module caps_sprue() {
    // Print flange-down (stem into bed support / or float with brim); dome up.
    pitch = cap_od + 4;
    lift = stem_len; // sit stem tips on Z=0 for flat bed contact option
    if (cap_index >= 0 && cap_index < button_count) {
        translate([0, 0, lift])
            button_cap_solid();
    } else {
        for (n = [0 : button_count - 1]) {
            translate([n * pitch, 0, lift])
                button_cap_solid();
            if (n < button_count - 1)
                translate([n * pitch + cap_od / 2 - 0.5, -0.6, lift - flange_h])
                    cube([pitch - cap_od + 1, 1.2, 0.8]);
        }
    }
}

module caps_in_assembly() {
    // Seat caps in lid shafts: flange in retention recess, dome proud of outer face
    for (n = [0 : button_count - 1]) {
        p = button_xy(n);
        translate([p[0], p[1], split_z + explode + lid_t])
            button_cap_solid();
    }
}

// --- Shells -----------------------------------------------------------------

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
        if (button_mode == "cap") {
            tactile_wells();
            wire_channels();
        }
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
        if (button_mode == "cap")
            button_cap_shafts();
        else
            button_panel8_holes();
        screw_holes_xy(-lip_h, lid_t + lip_h, screw_d + print_tol);
    }
}

module assembly() {
    base_shell();
    translate([0, 0, split_z + explode])
        lid_shell();
    if (button_mode == "cap")
        caps_in_assembly();
}

if (part == "base")
    base_shell();
else if (part == "lid")
    // Sit lid flat on XY for printing / STL export (lip up)
    translate([0, 0, 1.2])
        lid_shell();
else if (part == "caps")
    // Dome up for TPU; stem hangs below — or flip in slicer if preferred
    caps_sprue();
else
    assembly();
