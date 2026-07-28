import json, subprocess, os

tests = [
    '01_basic_inline_0088', '11_edge_cases_0871', '11_edge_cases_0968', '11_edge_cases_0974',
    '12_combinatorial_1020', '12_combinatorial_1028', '12_combinatorial_1103', '12_combinatorial_1104',
    '12_combinatorial_1727', '12_combinatorial_2190', '12_combinatorial_2579', '12_combinatorial_2830',
    '12_combinatorial_2839', '12_combinatorial_3179', '13_combinatorial2_3390', '13_combinatorial2_3393',
    '13_combinatorial2_3418', '13_combinatorial2_3420', '13_combinatorial2_3421', '13_combinatorial2_3499',
    '13_combinatorial2_3505', '13_combinatorial2_3504', '13_combinatorial2_3864'
]

fails = []
os.makedirs("scratch", exist_ok=True)
for t in tests:
    tc = json.load(open('relang/input/' + t + '.json', encoding='utf-8'))
    ex = json.load(open('relang/output/' + t + '.json', encoding='utf-8'))['output']
    r = subprocess.run(['python', 'target/marked.py'], input=tc['data'], capture_output=True, text=True, encoding='utf-8')
    ac = r.stdout
    fails.append(f"{t}\nIN : {repr(tc['data'])}\nEXP: {repr(ex)}\nACT: {repr(ac)}\n---")

with open('scratch/failures.txt', 'w', encoding='utf-8') as f:
    f.write('\n'.join(fails))
