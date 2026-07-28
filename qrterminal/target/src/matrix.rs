use crate::qr::{Code, Level};
use crate::tables;

pub fn build_matrix(version: usize, data: &[u8], level: Level) -> Code {
    let size = version * 4 + 17;
    let mut code = Code::new(size);
    
    // 1. Finder patterns
    draw_finder(&mut code, 0, 0);
    draw_finder(&mut code, size - 7, 0);
    draw_finder(&mut code, 0, size - 7);
    
    // Separators
    draw_separators(&mut code);
    
    // 2. Alignment patterns
    let aligns = crate::tables::get_alignment_pattern_positions(version);
    let num_aligns = aligns.len();
    for i in 0..num_aligns {
        for j in 0..num_aligns {
            let (x, y) = (aligns[i], aligns[j]);
            if (x == 6 && y == 6) || (x == 6 && y == size - 7) || (x == size - 7 && y == 6) {
                continue; // Overlaps with finder patterns
            }
            draw_alignment_pattern(&mut code, x, y);
        }
    }
    
    // 3. Timing patterns
    for i in 8..size - 8 {
        if !code.is_dark[6 * size + i] {
            code.set(i, 6, i % 2 == 0);
        }
        if !code.is_dark[i * size + 6] {
            code.set(6, i, i % 2 == 0);
        }
    }
    
    // 4. Version info
    if version >= 7 {
        // Reserve version info
        for i in 0..6 {
            for j in 0..3 {
                code.is_dark[(size - 11 + j) * size + i] = true;
                code.is_dark[i * size + (size - 11 + j)] = true;
            }
        }
    }
    
    // 5. Format info
    for i in 0..9 {
        code.is_dark[8 * size + i] = true;
        code.is_dark[i * size + 8] = true;
    }
    for i in size - 8..size {
        code.is_dark[8 * size + i] = true;
        code.is_dark[i * size + 8] = true;
    }
    code.is_dark[8 * size + 8] = true;
    
    // Always dark module
    code.set(8, size - 8, true);
    
    // 6. Data placement (zigzag)
    let mut bit_idx = 0;
    let mut right = size - 1;
    let total_bits = data.len() * 8;
    
    while right > 0 {
        if right == 6 {
            right -= 1;
        }
        
        let upward = ((size - 1 - right) / 2) % 2 == 0;
        
        for i in 0..size {
            let y = if upward { size - 1 - i } else { i };
            for j in 0..2 {
                let x = right - j;
                if !code.is_dark[y * size + x] {
                    let bit = if bit_idx < total_bits {
                        ((data[bit_idx / 8] >> (7 - (bit_idx % 8))) & 1) == 1
                    } else {
                        false
                    };
                    code.modules[y * size + x] = bit;
                    bit_idx += 1;
                }
            }
        }
        if right < 2 {
            break;
        }
        right -= 2;
    }
    
    // Apply masking and find best
    crate::masking::apply_best_mask(&mut code, version, level);
    code
}

fn draw_finder(code: &mut Code, ox: usize, oy: usize) {
    for y in 0..7 {
        for x in 0..7 {
            let is_border = x == 0 || x == 6 || y == 0 || y == 6;
            let is_center = x >= 2 && x <= 4 && y >= 2 && y <= 4;
            code.modules[(oy + y) * code.size + (ox + x)] = is_border || is_center;
        }
    }
    
    // Mark 8x8 area as reserved (including separator)
    let start_x = if ox == 0 { 0 } else { ox - 1 };
    let end_x = if ox == 0 { 8 } else { code.size };
    let start_y = if oy == 0 { 0 } else { oy - 1 };
    let end_y = if oy == 0 { 8 } else { code.size };
    
    for y in start_y..end_y {
        for x in start_x..end_x {
            code.is_dark[y * code.size + x] = true;
        }
    }
}

fn draw_separators(code: &mut Code) {
    let size = code.size;
    // Top-left
    for i in 0..8 {
        code.set(7, i, false);
        code.set(i, 7, false);
    }
    // Top-right
    for i in 0..8 {
        code.set(size - 8, i, false);
        code.set(size - 1 - i, 7, false);
    }
    // Bottom-left
    for i in 0..8 {
        code.set(7, size - 1 - i, false);
        code.set(i, size - 8, false);
    }
}

fn draw_alignment_pattern(code: &mut Code, cx: usize, cy: usize) {
    for y in 0..5 {
        for x in 0..5 {
            let px = cx - 2 + x;
            let py = cy - 2 + y;
            let is_border = x == 0 || x == 4 || y == 0 || y == 4;
            let is_center = x == 2 && y == 2;
            code.set(px, py, is_border || is_center);
        }
    }
}
