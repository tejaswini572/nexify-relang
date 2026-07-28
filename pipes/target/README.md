# Pipes - Go Reimplementation

This is a Go port of the Python animated pipes terminal screensaver. It reproduces the exact observable behavior of the original Python script, including pipe movement logic, terminal rendering, configuration, and keyboard controls.

## Prerequisites

- **Go 1.21+**
  - **Ubuntu 24.04 Install Instructions**:
    ```bash
    sudo apt update
    sudo apt install golang-go
    ```

## Build

To build the executable from source, run:

```bash
cd target
go build -o pipes main.go types.go config.go renderer.go pipes.go
```

## Run

Run the generated executable:

```bash
./pipes
```

To run it via the Hackathon `relang` wrapper:
```bash
source ../setup.sh
relang "./pipes"
```

### Supported CLI Flags

The Go port mirrors the exact same CLI flags and default behavior as the original Python script:

- `-p`, `-pipes`: number of pipes (default: 1)
- `-f`, `-fps`: frames per second (20-100) (default: 75)
- `-s`, `-steady`: steadiness (5-15) (default: 13)
- `-r`, `-limit`: character limit before reset (default: 2000)
- `-R`, `-random`: random start (default: false)
- `-B`, `-no-bold`: disable bold (default: false)
- `-C`, `-no-color`: disable color (default: false)
- `-P`, `-pipe-style`: change pipe style (0-9)
- `-K`, `-keep-style`: keep style on wrap (default: false)
- `-S`, `-save-config`: save current settings as default
- `-v`, `-version`: show version

### Keyboard Controls

While running, the following interactive controls are supported:
- `P`/`O`: Increase/decrease steadiness
- `F`/`D`: Increase/decrease FPS
- `B`: Toggle bold rendering
- `C`: Toggle colors
- `K`: Toggle keeping the style on wrap
- `?` or `ESC`: Quit

## Porting Notes

- **Terminal Rendering**: The original Python script leverages `curses`, which abstracts away terminal raw-mode handling and cursor positioning. To eliminate heavy CGO or 3rd-party dependencies, this Go port utilizes pure ANSI escape sequences and `golang.org/x/term` for raw mode terminal manipulation.
- **Pipe Movement Math**: The random walk algorithm and wrapping behavior match the Python equivalent entirely (e.g., `-oldDir + 2` for horizontal movement, and probability math using `rand.Intn`).
- **Color Cycling**: Matches the mapped palette mechanism in Python.
- **Clean Shutdown**: Implements a dedicated non-blocking goroutine to listen for OS signals (like `SIGINT`) and user inputs, ensuring `term.Restore()` is called and the cursor is un-hidden smoothly upon exiting.
