import os

def generate_alignment_patterns():
    # ISO/IEC 18004:2015(E) Table E.1
    # Version 1 has none, V2-40 have them.
    # We can compute them: 
    # V1: []
    # V2-V40: 
    # first is always 6, last is always 4 * V + 17
    # step is roughly (last - 6) / (num_patterns - 1), rounded to nearest even integer
    # Actually, the exact calculation per standard:
    # intervals are even. 
    # Let's just generate the known list.
    def get_alignments(version):
        if version == 1: return []
        first = 6
        last = 4 * version + 17
        num = version // 7 + 2
        if num == 2:
            return [first, last]
        step = (last - first + num - 2) // (num - 1)
        if step % 2 != 0:
            step = (last - first + num - 2) // (num - 1) // 2 * 2
            if step == 0: step = 2 # fallback, but formula is actually complex.
            
    # Better to just use the exact known table.
    alignments = [
        [], # v1
        [6, 18], # v2
        [6, 22],
        [6, 26],
        [6, 30],
        [6, 34],
        [6, 22, 38],
        [6, 24, 42],
        [6, 26, 46],
        [6, 28, 50],
        [6, 30, 54],
        [6, 32, 58],
        [6, 34, 62],
        [6, 26, 46, 66],
        [6, 26, 48, 70],
        [6, 26, 50, 74],
        [6, 30, 54, 78],
        [6, 30, 56, 82],
        [6, 30, 58, 86],
        [6, 34, 62, 90],
        [6, 28, 50, 72, 94],
        [6, 26, 50, 74, 98],
        [6, 30, 54, 78, 102],
        [6, 28, 54, 80, 106],
        [6, 32, 58, 84, 110],
        [6, 30, 58, 86, 114],
        [6, 34, 62, 90, 118],
        [6, 26, 50, 74, 98, 122],
        [6, 30, 54, 78, 102, 126],
        [6, 26, 52, 78, 104, 130],
        [6, 30, 56, 82, 108, 134],
        [6, 34, 60, 86, 112, 138],
        [6, 30, 58, 86, 114, 142],
        [6, 34, 62, 90, 118, 146],
        [6, 30, 54, 78, 102, 126, 150],
        [6, 24, 50, 76, 102, 128, 154],
        [6, 28, 54, 80, 106, 132, 158],
        [6, 32, 58, 84, 110, 136, 162],
        [6, 26, 54, 82, 110, 138, 166],
        [6, 30, 58, 86, 114, 142, 170]
    ]
    return alignments

def get_ec_blocks():
    # Tables for EC blocks per version and EC level (L, M, Q, H).
    # Format: [Total Data Codewords, EC Codewords per Block, (num_blocks_1, data_cw_1), (num_blocks_2, data_cw_2)]
    # We will generate this from a compact string representation.
    
    # We can just write out the capacities for a few common versions, or generate all 40.
    # To keep the AI token limit and generation time small, let's include V1 to V40 exactly as in standard.
    
    # I'll just write the Rust file string directly!
    rust_code = """
pub struct ECBlock {
    pub num_blocks: usize,
    pub data_codewords: usize,
}

pub struct ECInfo {
    pub ec_codewords_per_block: usize,
    pub blocks1: ECBlock,
    pub blocks2: Option<ECBlock>,
}

pub fn get_alignment_patterns(version: usize) -> &'static [usize] {
    const ALIGNMENTS: &[&[usize]] = &[
        &[], // v0
        &[], // v1
        &[6, 18], // v2
        &[6, 22],
        &[6, 26],
        &[6, 30],
        &[6, 34],
        &[6, 22, 38],
        &[6, 24, 42],
        &[6, 26, 46],
        &[6, 28, 50],
        &[6, 30, 54],
        &[6, 32, 58],
        &[6, 34, 62],
        &[6, 26, 46, 66],
        &[6, 26, 48, 70],
        &[6, 26, 50, 74],
        &[6, 30, 54, 78],
        &[6, 30, 56, 82],
        &[6, 30, 58, 86],
        &[6, 34, 62, 90],
        &[6, 28, 50, 72, 94],
        &[6, 26, 50, 74, 98],
        &[6, 30, 54, 78, 102],
        &[6, 28, 54, 80, 106],
        &[6, 32, 58, 84, 110],
        &[6, 30, 58, 86, 114],
        &[6, 34, 62, 90, 118],
        &[6, 26, 50, 74, 98, 122],
        &[6, 30, 54, 78, 102, 126],
        &[6, 26, 52, 78, 104, 130],
        &[6, 30, 56, 82, 108, 134],
        &[6, 34, 60, 86, 112, 138],
        &[6, 30, 58, 86, 114, 142],
        &[6, 34, 62, 90, 118, 146],
        &[6, 30, 54, 78, 102, 126, 150],
        &[6, 24, 50, 76, 102, 128, 154],
        &[6, 28, 54, 80, 106, 132, 158],
        &[6, 32, 58, 84, 110, 136, 162],
        &[6, 26, 54, 82, 110, 138, 166],
        &[6, 30, 58, 86, 114, 142, 170]
    ];
    ALIGNMENTS[version]
}
"""
    with open("src/tables.rs", "w") as f:
        f.write(rust_code)
    
os.makedirs("src", exist_ok=True)
generate_alignment_patterns()
