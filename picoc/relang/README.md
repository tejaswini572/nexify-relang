# picoc — Test Suite

C interpreter

## Contents

| File/Dir | Purpose |
|---|---|
| `input/` | Test input files (`.json`), one per test case |
| `output/` | Expected output files (`.json`), one per test case |
| `project_config.json` | Project metadata: input type, name, ID |
| `validate.py` | Local test runner — run your tool against the test suite |
| `tester.py` | CLI adapter — receives batch via stdin, outputs results |

## Test case format (`input/{id}.json`)

```json
{
  "id": "test_001",
  "type": "file",
  "data": "...C source code...",
  "timeout": 30
}
```

- `type`: `"file"` — write `data` to a temp `.c` file and pass its path as the program argument
- `data`: the raw C source code your interpreter should execute
- `timeout`: max seconds for execution

## Expected output format (`output/{id}.json`)

```json
{
  "id": "test_001",
  "output": "...expected stdout..."
}
```

## How to test locally

```bash
python3 validate.py "<your-tool-command>"
```

Examples:

```bash
# Python interpreter
python3 validate.py "python3 picoc.py"

# Node.js interpreter
python3 validate.py "node picoc.js"

# Compiled binary
python3 validate.py "./picoc_cpp"

# With arguments
python3 validate.py "java -jar picoc.jar"
```

The script runs your tool against every test case, hashes the output, and compares against the expected hash. Pass/fail per test is printed, with a summary at the end.

## Test selection

This folder contains **307/1228 test cases** (25% of the full suite). Use `build_deliverables.py --percent N` to regenerate with a different percentage.

## Hash-based comparison

Output correctness is verified by **SHA256 hash comparison**, not direct text diff. Your tool must produce **byte-identical** output to the reference C compiler. This is deterministic across languages for the same C source.
