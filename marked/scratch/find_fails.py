import json, subprocess, os, sys
from pathlib import Path

test_dir = Path('relang/input')
out_dir = Path('relang/output')
fails = []
passes = 0

for jf in sorted(test_dir.glob('*.json')):
    t = jf.stem
    try:
        tc = json.loads(jf.read_text(encoding='utf-8'))
        ex = json.loads((out_dir / jf.name).read_text(encoding='utf-8'))['output']
        r = subprocess.run(
            ['python', 'target/marked.py'],
            input=tc['data'], capture_output=True, text=True, encoding='utf-8',
            timeout=10  # 10 second timeout per test
        )
        ac = r.stdout
        if ac != ex:
            fails.append({'test': t, 'input': tc['data'], 'exp': ex, 'act': ac})
        else:
            passes += 1
    except subprocess.TimeoutExpired:
        fails.append({'test': t, 'input': tc.get('data','?'), 'error': 'TIMEOUT'})
    except Exception as e:
        fails.append({'test': t, 'error': str(e)})

print(f"PASS: {passes}/{passes+len(fails)}")
print(f"FAIL: {len(fails)}")
for f in fails:
    print(f"\n=== {f['test']} ===")
    print(f"IN : {repr(f.get('input','?'))}")
    print(f"EXP: {repr(f.get('exp','?'))}")
    print(f"ACT: {repr(f.get('act','?'))}")
    if 'error' in f:
        print(f"ERR: {f['error']}")
