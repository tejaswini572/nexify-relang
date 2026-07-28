# qrterminal

QR code generator for the terminal, implemented in Rust from scratch.

## Prerequisites

- Rust (Cargo) 1.70+

**Install Rust (Ubuntu 24.04):**
```bash
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source $HOME/.cargo/env
```

## Build

```bash
cd qrterminal/target && cargo build --release
```

## Run

```bash
relang "target/release/qrterminal 'https://example.com'"
```
Or directly:
```bash
./target/release/qrterminal "https://example.com"
```

## Supported Flags

- `-l [L, M, Q, H]` : Error correction level (default L).
- `-q [size]` : Quiet zone size in blocks (default 2).
- `-v` : Verbose output.
- `-s` : Disable sixel (Ignored; terminal output defaults to Half-Blocks).

## Behavior Matching Notes

- **Encoding**: Uses standard Byte Mode encoding and computes required version size (V1-V40) dynamically based on input length.
- **Error Correction**: Implemented Reed-Solomon GF(2^8) math identically to ISO 18004 standards to match EC levels.
- **Masking**: Computes all 4 standard penalties (N1 to N4) and picks the best mask to guarantee identical layout patterns to the reference implementation.
- **Rendering**: Emulates the Go reference's `writeHalfBlocks` using the `▀`, `▄`, `█`, ` ` characters and accurately calculates the padded quiet zone borders.
