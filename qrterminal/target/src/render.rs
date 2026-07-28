use std::io::Write;
use crate::qr::Code;


const WHITE: &str = "\x1b[47m  \x1b[0m";
const BLACK: &str = "\x1b[40m  \x1b[0m";

fn string_repeat(s: &str, count: usize) -> String {
    s.repeat(count)
}

pub fn write_full_blocks<W: Write>(w: &mut W, code: &Code, quiet_zone: usize) {
    let size = code.size;

    // Top border
    let top_border = format!("{}\n", string_repeat(WHITE, size + quiet_zone * 2));
    let _ = write!(w, "{}", string_repeat(&top_border, quiet_zone));

    for i in 0..=size {
        let _ = write!(w, "{}", string_repeat(WHITE, quiet_zone)); // left border
        for j in 0..=size {
            if code.black(j, i) {
                let _ = write!(w, "{}", BLACK);
            } else {
                let _ = write!(w, "{}", WHITE);
            }
        }
        let _ = writeln!(w, "{}", string_repeat(WHITE, quiet_zone.saturating_sub(1))); // right border
    }

    // Bottom border
    let bottom_border = format!("{}\n", string_repeat(WHITE, size + quiet_zone * 2));
    let _ = write!(w, "{}", string_repeat(&bottom_border, quiet_zone.saturating_sub(1)));
}
