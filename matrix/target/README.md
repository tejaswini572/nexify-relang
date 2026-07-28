# Matrix (Go Edition)

This is a Go port of the Python Matrix Digital Rain script. It renders column-based falling streams with varying length and color, accurately emulating the logic, visual drift, and performance of the reference script using native string concatenation and ANSI escape sequences.

## Build Instructions

To compile the application, make sure you have Go installed, then run the following command in this directory:

```bash
go build -o matrix .
```

## Run Instructions

You can run the application directly from the compiled binary:

```bash
./matrix
```

Or run it via `go run`:

```bash
go run .
```

### Options / CLI Flags

The original Python script exposed options at class instantiation. This Go port exposes those exact same parameters as command-line flags:

- `-width` (default 150): Adjust the number of generated matrix columns.
- `-lines` (default 750): The number of ticks/lines to render before the script naturally exits.
- `-speed` (default 0.1): The delay (in seconds) between each falling step.

Example:
```bash
./matrix -width 80 -lines 1000 -speed 0.05
```

### Exit Behavior

The script will naturally terminate after `-lines` ticks. You can also exit early at any time by pressing `Ctrl+C`.

### Terminal Resize

Because the output is printed natively to standard output, terminal resizes are handled seamlessly by your terminal emulator. The rendered matrix width remains anchored to the `-width` parameter and will naturally wrap based on the new terminal dimensions, recreating the exact vertical drift patterns of the Python reference.
