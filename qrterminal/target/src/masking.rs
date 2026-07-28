use crate::qr::{Code, Level};

pub fn apply_best_mask(code: &mut Code, version: usize, level: Level) {
    let best_mask = 0;
    apply_mask(code, best_mask);
    draw_format_info(code, level, best_mask);
    if version >= 7 {
        draw_version_info(code, version);
    }
}


fn apply_mask(code: &mut Code, mask: usize) {
    let size = code.size;
    for y in 0..size {
        for x in 0..size {
            if !code.is_dark[y * size + x] {
                let invert = match mask {
                    0 => (x + y) % 2 == 0,
                    1 => y % 2 == 0,
                    2 => x % 3 == 0,
                    3 => (x + y) % 3 == 0,
                    4 => (y / 2 + x / 3) % 2 == 0,
                    5 => (x * y) % 2 + (x * y) % 3 == 0,
                    6 => ((x * y) % 2 + (x * y) % 3) % 2 == 0,
                    7 => ((x + y) % 2 + (x * y) % 3) % 2 == 0,
                    _ => false,
                };
                if invert {
                    code.modules[y * size + x] = !code.modules[y * size + x];
                }
            }
        }
    }
}

fn draw_format_info(code: &mut Code, level: Level, mask: usize) {
    let ecl_format_bits = match level {
        Level::L => 1,
        Level::M => 0,
        Level::Q => 3,
        Level::H => 2,
    };
    let data = (ecl_format_bits << 3) | mask;
    
    let mut rem = data << 10;
    for i in (10..=14).rev() {
        if (rem >> i) & 1 == 1 {
            rem ^= 0x537 << (i - 10);
        }
    }
    let bits = ((data << 10) | rem) ^ 0x5412;
    
    let size = code.size;
    for i in 0..15 {
        let bit = ((bits >> i) & 1) == 1;
        
        let (x1, y1) = if i < 6 {
            (8, i)
        } else if i < 8 {
            (8, i + 1)
        } else if i < 9 {
            (7, 8)
        } else {
            (14 - i, 8)
        };
        code.modules[y1 * size + x1] = bit;
        
        let (x2, y2) = if i < 8 {
            (size - 1 - i, 8)
        } else {
            (8, size - 1 - (14 - i))
        };
        code.modules[y2 * size + x2] = bit;
    }
}

fn draw_version_info(code: &mut Code, version: usize) {
    let mut rem = version;
    for _ in 0..12 {
        rem = (rem << 1) ^ (if (rem >> 5) == 1 { 0x1F25 } else { 0 });
    }
    let bits = (version << 12) | (rem & 0xFFF);
    
    let size = code.size;
    for i in 0..18 {
        let bit = ((bits >> i) & 1) == 1;
        let a = size - 11 + i % 3;
        let b = i / 3;
        code.modules[b * size + a] = bit;
        code.modules[a * size + b] = bit;
    }
}
