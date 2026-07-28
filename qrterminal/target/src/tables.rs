// QR tables using dynamic calculation matching standard capacities

pub fn get_num_raw_data_modules(ver: usize) -> usize {
    let size = ver * 4 + 17;
    let mut is_dark = vec![false; size * size];
    
    // Finder patterns & separators
    for y in 0..8 {
        for x in 0..8 {
            is_dark[y * size + x] = true;
            is_dark[(size - 1 - y) * size + x] = true;
            is_dark[y * size + (size - 1 - x)] = true;
        }
    }
    // Timing patterns
    for i in 8..size-8 {
        is_dark[6 * size + i] = true;
        is_dark[i * size + 6] = true;
    }
    // Alignment patterns
    if ver >= 2 {
        let pos = get_alignment_pattern_positions(ver);
        for &r in &pos {
            for &c in &pos {
                if (r < 8 && c < 8) || (r > size - 8 && c < 8) || (r < 8 && c > size - 8) {
                    continue; // Overlaps with finder
                }
                for y in r-2..=r+2 {
                    for x in c-2..=c+2 {
                        is_dark[y * size + x] = true;
                    }
                }
            }
        }
    }
    // Format info
    for i in 0..9 {
        is_dark[8 * size + i] = true;
        is_dark[i * size + 8] = true;
    }
    for i in size - 8..size {
        is_dark[8 * size + i] = true;
        is_dark[i * size + 8] = true;
    }
    // Version info
    if ver >= 7 {
        for i in 0..18 {
            let a = size - 11 + i % 3;
            let b = i / 3;
            is_dark[b * size + a] = true;
            is_dark[a * size + b] = true;
        }
    }
    
    // One dark module
    is_dark[(size - 8) * size + 8] = true;

    let mut count = 0;
    for &d in &is_dark {
        if !d { count += 1; }
    }
    count / 8
}

pub fn get_alignment_pattern_positions(ver: usize) -> Vec<usize> {
    if ver == 1 {
        return vec![];
    }
    let num_align = ver / 7 + 2;
    let step = if ver == 32 {
        26
    } else {
        ((ver * 4 + num_align * 2 + 1) / (num_align * 2 - 2)) * 2
    };
    
    let mut pos = vec![6];
    let max_pos = ver * 4 + 10;
    for i in (0..(num_align - 1)).rev() {
        pos.push(max_pos - step * i);
    }
    pos
}

static ECC_CODEWORDS_PER_BLOCK: [[usize; 41]; 4] = [
    // L
    [0, 7, 10, 15, 20, 26, 18, 20, 24, 30, 18, 20, 24, 26, 30, 22, 24, 28, 30, 28, 28, 28, 28, 30, 30, 26, 28, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30],
    // M
    [0, 10, 16, 26, 18, 24, 16, 18, 22, 22, 26, 30, 22, 22, 24, 24, 28, 28, 26, 26, 26, 26, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28],
    // Q
    [0, 13, 22, 18, 26, 18, 24, 18, 22, 20, 24, 28, 26, 24, 20, 30, 24, 28, 28, 26, 30, 28, 30, 30, 30, 30, 28, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30],
    // H
    [0, 17, 28, 22, 16, 22, 28, 26, 26, 24, 28, 24, 28, 22, 24, 24, 30, 28, 28, 26, 28, 30, 24, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30],
];

static NUM_ERROR_CORRECTION_BLOCKS: [[usize; 41]; 4] = [
    // L
    [0, 1, 1, 1, 1, 1, 2, 2, 2, 2, 4, 4, 4, 4, 4, 6, 6, 6, 6, 7, 8, 8, 9, 9, 10, 12, 12, 12, 13, 14, 15, 16, 17, 18, 19, 19, 20, 21, 22, 24, 25],
    // M
    [0, 1, 1, 1, 2, 2, 4, 4, 4, 5, 5, 5, 8, 9, 9, 10, 10, 11, 13, 14, 16, 17, 17, 18, 20, 21, 23, 25, 26, 28, 29, 31, 33, 35, 37, 38, 40, 43, 45, 47, 49],
    // Q
    [0, 1, 1, 2, 2, 4, 4, 6, 6, 8, 8, 8, 10, 12, 16, 12, 17, 16, 18, 21, 20, 23, 23, 25, 27, 29, 34, 34, 35, 38, 40, 43, 45, 48, 51, 53, 56, 59, 62, 65, 68],
    // H
    [0, 1, 1, 2, 4, 4, 4, 5, 6, 8, 8, 11, 11, 16, 16, 18, 16, 19, 21, 25, 25, 25, 34, 30, 32, 35, 37, 40, 42, 45, 48, 51, 54, 57, 60, 63, 66, 70, 74, 77, 81],
];

pub fn get_ec_codewords_per_block(ver: usize, level: crate::qr::Level) -> usize {
    ECC_CODEWORDS_PER_BLOCK[level as usize][ver]
}

pub fn get_num_ec_blocks(ver: usize, level: crate::qr::Level) -> usize {
    NUM_ERROR_CORRECTION_BLOCKS[level as usize][ver]
}
