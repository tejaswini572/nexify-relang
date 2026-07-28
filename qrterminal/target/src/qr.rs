

#[derive(Clone, Copy, PartialEq, Eq)]
pub enum Level {
    L = 0,
    M = 1,
    Q = 2,
    H = 3,
}

pub struct Code {
    pub size: usize,
    pub modules: Vec<bool>,
    pub is_dark: Vec<bool>,
}

impl Code {
    pub fn new(size: usize) -> Self {
        Self {
            size,
            modules: vec![false; size * size],
            is_dark: vec![false; size * size],
        }
    }

    pub fn set(&mut self, x: usize, y: usize, v: bool) {
        if x < self.size && y < self.size {
            self.modules[y * self.size + x] = v;
            self.is_dark[y * self.size + x] = true;
        }
    }

    pub fn get(&self, x: usize, y: usize) -> bool {
        if x < self.size && y < self.size {
            self.modules[y * self.size + x]
        } else {
            false
        }
    }

    pub fn black(&self, x: usize, y: usize) -> bool {
        self.get(x, y)
    }
}

pub fn get_data_capacity(version: usize, level: Level) -> usize {
    let raw_modules = crate::tables::get_num_raw_data_modules(version);
    let ec_blocks = crate::tables::get_num_ec_blocks(version, level);
    let ec_codewords_per_block = crate::tables::get_ec_codewords_per_block(version, level);
    let total_ec_codewords = ec_blocks * ec_codewords_per_block;
    
    raw_modules - total_ec_codewords
}

pub fn encode(text: &str, level: Level) -> Code {
    let data = text.as_bytes();
    
    for version in 1..=40 {
        if let Some(code) = try_encode(data, version, level) {
            return code;
        }
    }
    
    panic!("Data too long to fit in any QR code version.");
}

fn try_encode(data: &[u8], version: usize, level: Level) -> Option<Code> {
    let capacity_bytes = get_data_capacity(version, level);
    
    let len_bits = if version < 10 { 8 } else { 16 };
    let required_bits = 4 + len_bits + data.len() * 8;
    if required_bits > capacity_bytes * 8 {
        return None;
    }
    
    let bits = crate::encoding::encode_byte_mode(data, version, capacity_bytes);
    let codewords = generate_ec(&bits, version, level, capacity_bytes);
    
    let matrix = crate::matrix::build_matrix(version, &codewords, level);
    Some(matrix)
}

fn generate_ec(data: &[u8], version: usize, level: Level, total_data_codewords: usize) -> Vec<u8> {
    let num_blocks = crate::tables::get_num_ec_blocks(version, level);
    let block_ec_len = crate::tables::get_ec_codewords_per_block(version, level);
    let raw_modules = crate::tables::get_num_raw_data_modules(version);
    
    let num_short_blocks = num_blocks - raw_modules % num_blocks;
    let short_block_len = raw_modules / num_blocks;
    
    let mut blocks_data: Vec<Vec<u8>> = Vec::new();
    let mut blocks_ec: Vec<Vec<u8>> = Vec::new();
    
    let mut k = 0;
    for i in 0..num_blocks {
        let block_len = short_block_len - block_ec_len + if i < num_short_blocks { 0 } else { 1 };
        let block_data = &data[k..k + block_len];
        k += block_len;
        
        blocks_data.push(block_data.to_vec());
        let ec = crate::reed_solomon::append_error_correction(block_data, block_ec_len);
        blocks_ec.push(ec);
    }
    
    let mut result = Vec::new();
    // Interleave data
    let max_data = blocks_data.iter().map(|b| b.len()).max().unwrap_or(0);
    for i in 0..max_data {
        for block in &blocks_data {
            if i < block.len() {
                result.push(block[i]);
            }
        }
    }
    
    // Interleave EC
    let max_ec = blocks_ec.iter().map(|b| b.len()).max().unwrap_or(0);
    for i in 0..max_ec {
        for block in &blocks_ec {
            if i < block.len() {
                result.push(block[i]);
            }
        }
    }
    
    result
}
