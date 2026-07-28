import time, marked, re, functools

# ── 1. escape_html_entities: double-pass (search then sub) vs single-pass ─────

ESCAPE_REPLACEMENTS = {"&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;"}

def _rep(m): return ESCAPE_REPLACEMENTS[m.group(0)]

ENCODE_PAT  = re.compile(r"[&<>\"']")
NO_ENC_PAT  = re.compile(r"""[<>"']|&(?!(#\d{1,7}|#[Xx][a-fA-F0-9]{1,6}|\w+);)""")

def current_escape(text, encode=False):
    if encode:
        if ENCODE_PAT.search(text): return ENCODE_PAT.sub(_rep, text)
    elif NO_ENC_PAT.search(text):   return NO_ENC_PAT.sub(_rep, text)
    return text

def direct_escape(text, encode=False):
    if encode: return ENCODE_PAT.sub(_rep, text)
    return NO_ENC_PAT.sub(_rep, text)

# text WITH matches — search is wasted work
test_special = '<b>Hello &amp; world</b> with "quotes" ' * 200

N = 1000
t0 = time.perf_counter()
for _ in range(N): current_escape(test_special)
print(f"escape WITH matches  — current (search+sub): {(time.perf_counter()-t0)*1000:.1f}ms")

t0 = time.perf_counter()
for _ in range(N): direct_escape(test_special)
print(f"escape WITH matches  — direct  (sub only)  : {(time.perf_counter()-t0)*1000:.1f}ms")

# text WITHOUT matches — search short-circuits; direct sub also fast
test_clean = "hello world no special chars " * 200

t0 = time.perf_counter()
for _ in range(N): current_escape(test_clean)
print(f"escape NO  matches   — current (search+sub): {(time.perf_counter()-t0)*1000:.1f}ms")

t0 = time.perf_counter()
for _ in range(N): direct_escape(test_clean)
print(f"escape NO  matches   — direct  (sub only)  : {(time.perf_counter()-t0)*1000:.1f}ms")

# ── 2. split_cells: O(n^2) char-by-char loop ──────────────────────────────────
print()
row_100 = " | ".join("cell " + str(i) for i in range(100))   # 100 columns
row_10  = " | ".join("cell " + str(i) for i in range(10))

t0 = time.perf_counter()
for _ in range(N): marked.split_cells(row_100)
print(f"split_cells  100-col row  {N}x: {(time.perf_counter()-t0)*1000:.1f}ms")

t0 = time.perf_counter()
for _ in range(N): marked.split_cells(row_10)
print(f"split_cells   10-col row  {N}x: {(time.perf_counter()-t0)*1000:.1f}ms")

# ── 3. Regex recompilation inside list hot path ───────────────────────────────
print()
rules = marked.RULES

t0 = time.perf_counter()
for _ in range(10000):
    rules.next_bullet_regex(4)
    rules.hr_regex(4)
    rules.fences_begin_regex(4)
    rules.heading_begin_regex(4)
    rules.html_begin_regex(4)
    rules.blockquote_begin_regex(4)
print(f"6 regex compilations x10000 (list inner loop): {(time.perf_counter()-t0)*1000:.1f}ms")

# With lru_cache — safe because indent values are tiny ints (0-4)
cached_nb  = functools.lru_cache(maxsize=8)(rules.next_bullet_regex)
cached_hr  = functools.lru_cache(maxsize=8)(rules.hr_regex)
cached_fb  = functools.lru_cache(maxsize=8)(rules.fences_begin_regex)
cached_hb  = functools.lru_cache(maxsize=8)(rules.heading_begin_regex)
cached_htm = functools.lru_cache(maxsize=8)(rules.html_begin_regex)
cached_bq  = functools.lru_cache(maxsize=8)(rules.blockquote_begin_regex)

t0 = time.perf_counter()
for _ in range(10000):
    cached_nb(4); cached_hr(4); cached_fb(4)
    cached_hb(4); cached_htm(4); cached_bq(4)
print(f"6 cached lookups      x10000 (list inner loop): {(time.perf_counter()-t0)*1000:.1f}ms")

# ── 4. run_parse output building ─────────────────────────────────────────────
print()
parser = marked.Parser()
tokens = marked.Lexer().run_lex("# heading\n\nparagraph\n\n" * 200)
t0 = time.perf_counter()
for _ in range(100):
    parser.run_parse(tokens)
print(f"run_parse 400 tokens x100: {(time.perf_counter()-t0)*1000:.1f}ms")

# ── 5. Adversarial regex backtracking probes ──────────────────────────────────
print()
import signal, sys

def alarm(sig, frame): raise TimeoutError()

# lheading is the most complex regex — test on long ambiguous content
lheading = marked.RULES.block.lheading
test_lh = ("x " * 40 + "\n") * 20  # lots of lines, no setext marker
t0 = time.perf_counter()
lheading.match(test_lh)
elapsed = (time.perf_counter() - t0) * 1000
print(f"lheading regex on 20-line ambiguous text: {elapsed:.2f}ms")

# paragraph regex on text with many near-matches for the interruption lookahead
para_pat = marked.RULES.block.paragraph
test_para = ("x " * 80 + "\n") * 20
t0 = time.perf_counter()
para_pat.match(test_para)
elapsed = (time.perf_counter() - t0) * 1000
print(f"paragraph regex on 20-line text: {elapsed:.2f}ms")

# inline text regex on a line with many backtick-like chars
text_pat = marked.RULES.inline.text
test_inline = ("`word` " * 100)
t0 = time.perf_counter()
for _ in range(1000):
    text_pat.match(test_inline)
elapsed = (time.perf_counter() - t0) * 1000
print(f"inline.text regex x1000 on 700-char backtick line: {elapsed:.1f}ms")
