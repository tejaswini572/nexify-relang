# Asciiquarium (Go Port)

This is a complete Go port of the original Perl Asciiquarium animation, written from scratch and utilizing `tcell` for terminal rendering.

## Requirements
- Go 1.20 or newer
- A terminal with 256-color support

## Build Instructions

To build the executable, run the following command in this directory (`asciiquarium/target/`):

```bash
# This will download the required dependencies (tcell) and build the binary
go build -o asciiquarium .
```

## Run Instructions

After building, run the executable:

```bash
./asciiquarium
```

### Controls:
- **`q`** or **`ESC`**: Quit the aquarium.
- **`p`**: Pause or unpause the animation.
- **`r`**: Redraw and respawn all entities.

## Dependencies

The program uses Go modules to manage dependencies. The required dependency is:
- `github.com/gdamore/tcell/v2` (for cross-platform terminal rendering)

If you haven't already, ensure dependencies are downloaded using:
```bash
go mod tidy
```
