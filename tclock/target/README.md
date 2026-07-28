# tclock (Rust Port)

## Requirements
- Rust toolchain (cargo, rustc).
- To install on Ubuntu 24.04, run:
  ```bash
  curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
  source $HOME/.cargo/env
  ```

## Build
```bash
cargo build --release
```
The executable will be located at `target/release/tclock`.

## Run via relang
To run this project via the `relang` hackathon testing setup:
```bash
source ../setup.sh
relang ./target/release/tclock
```
(Or run it directly as `./target/release/tclock`)
