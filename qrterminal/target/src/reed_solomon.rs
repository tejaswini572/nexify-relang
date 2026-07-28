// Reed-Solomon Error Correction

pub fn append_error_correction(data: &[u8], num_ec_codewords: usize) -> Vec<u8> {
    let generator = rs_generator_polynomial(num_ec_codewords);
    let mut remainder = vec![0u8; num_ec_codewords];

    for &b in data {
        let factor = remainder[0] ^ b;
        remainder.remove(0);
        remainder.push(0);

        if factor != 0 {
            for i in 0..num_ec_codewords {
                remainder[i] ^= gf_multiply(generator[i], factor);
            }
        }
    }

    remainder
}

fn rs_generator_polynomial(degree: usize) -> Vec<u8> {
    let mut g = vec![1u8];
    let mut root = 1u8;
    for _ in 0..degree {
        let mut new_g = vec![0u8; g.len() + 1];
        for i in 0..g.len() {
            new_g[i] = g[i];
        }
        for i in 0..g.len() {
            new_g[i + 1] ^= gf_multiply(g[i], root);
        }
        g = new_g;
        root = gf_multiply(root, 2);
    }
    g[1..].to_vec()
}

fn gf_multiply(x: u8, y: u8) -> u8 {
    if x == 0 || y == 0 {
        return 0;
    }
    
    // Russian peasant multiplication in GF(2^8)
    let mut z = 0u8;
    let mut a = x;
    let mut b = y;
    for _ in 0..8 {
        if b & 1 == 1 {
            z ^= a;
        }
        let high_bit = a & 0x80;
        a <<= 1;
        if high_bit != 0 {
            a ^= 0x1D; // Modulo polynomial x^8 + x^4 + x^3 + x^2 + 1
        }
        b >>= 1;
    }
    z
}
