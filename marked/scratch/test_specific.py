import subprocess, json
from pathlib import Path

tests = [
    '01_basic_inline_0088',
]

for t in tests:
    inp = json.loads(Path(f'relang/input/{t}.json').read_text(encoding='utf-8'))['data']
    exp = json.loads(Path(f'relang/output/{t}.json').read_text(encoding='utf-8'))['output']
    r = subprocess.run(['python', 'target/marked.py'], input=inp, capture_output=True, text=True, encoding='utf-8')
    act = r.stdout
    status = 'PASS' if act == exp else 'FAIL'
    print(f'[{status}] {t}')
    if act != exp:
        print(f'  IN : {repr(inp)}')
        print(f'  EXP: {repr(exp)}')
        print(f'  ACT: {repr(act)}')
