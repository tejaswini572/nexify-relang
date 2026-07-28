pub struct BitStream {
    bits: Vec<u8>,
    len: usize,
}

impl BitStream {
    pub fn new() -> Self {
        Self { bits: Vec::new(), len: 0 }
    }

    pub fn push(&mut self, value: u32, length: usize) {
        for i in (0..length).rev() {
            let bit = ((value >> i) & 1) as u8;
            if self.len % 8 == 0 {
                self.bits.push(0);
            }
            if bit == 1 {
                let idx = self.bits.len() - 1;
                self.bits[idx] |= 1 << (7 - (self.len % 8));
            }
            self.len += 1;
        }
    }

    pub fn pad_to_bytes(&mut self, capacity_bytes: usize) {
        // Terminator
        let remaining_bits = capacity_bytes * 8 - self.len;
        let terminator_len = remaining_bits.min(4);
        self.push(0, terminator_len);

        // Bit padding to byte boundary
        while self.len % 8 != 0 {
            self.push(0, 1);
        }

        // Byte padding
        let padding_bytes = [0xEC, 0x11];
        let mut idx = 0;
        while self.len / 8 < capacity_bytes {
            self.push(padding_bytes[idx % 2] as u32, 8);
            idx += 1;
        }
    }

    pub fn into_bytes(self) -> Vec<u8> {
        self.bits
    }
}

pub fn encode_byte_mode(data: &[u8], version: usize, capacity_bytes: usize) -> Vec<u8> {
    let mut stream = BitStream::new();
    
    // Mode indicator for Byte Mode: 0100 (4 bits)
    stream.push(0b0100, 4);

    // Character count indicator
    let cci_len = if version < 10 {
        8
    } else {
        16
    };
    stream.push(data.len() as u32, cci_len);

    // Data
    for &b in data {
        stream.push(b as u32, 8);
    }

    stream.pad_to_bytes(capacity_bytes);
    stream.into_bytes()
}
