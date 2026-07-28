#!/usr/bin/env python3
"""Stress tests for marked.py — stability under edge-case conditions."""
import sys, time, traceback
sys.path.insert(0, 'target')
import marked

m = marked.Marked()
results = []

def run_test(name, inp, timeout_sec=5):
    """Run a single test, catching crashes and hangs."""
    t0 = time.perf_counter()
    try:
        result = m.parse(inp)
        elapsed = time.perf_counter() - t0
        if elapsed > timeout_sec:
            results.append((name, "SLOW", f"{elapsed:.1f}s (>{timeout_sec}s threshold)"))
        else:
            results.append((name, "PASS", f"{elapsed*1000:.1f}ms, {len(result)} chars output"))
    except RecursionError as e:
        results.append((name, "FAIL", f"RecursionError: {e}"))
    except RuntimeError as e:
        # marked.py raises RuntimeError for infinite loops
        results.append((name, "FAIL", f"RuntimeError: {e}"))
    except Exception as e:
        results.append((name, "FAIL", f"{type(e).__name__}: {e}"))

# ══════════════════════════════════════════════════════════════════════
# 1. MALFORMED / UNCLOSED CONSTRUCTS
# ══════════════════════════════════════════════════════════════════════
print("=== 1. Malformed/Unclosed Constructs ===")

run_test("Unclosed code fence (backtick)", "```python\nprint('hello')\n# no closing fence")
run_test("Unclosed code fence (tilde)", "~~~\ncode here\n# no closing fence")
run_test("Unclosed emphasis *", "*unclosed emphasis here")
run_test("Unclosed emphasis **", "**unclosed strong here")
run_test("Unclosed emphasis _", "_unclosed emphasis here")
run_test("Unclosed emphasis __", "__unclosed strong here")
run_test("Unclosed link bracket [", "[unclosed link")
run_test("Unclosed link bracket ![", "![unclosed image")
run_test("Unclosed link paren", "[link](http://example.com")
run_test("Unmatched table pipes", "| a | b |\n| - | - |\n| x | y | z | extra |")
run_test("Table with no delimiter row", "| a | b |\n| x | y |")
run_test("Unclosed HTML tag", "<div>\n<p>unclosed")
run_test("Unclosed HTML comment", "<!-- unclosed comment")
run_test("Unclosed inline code", "`unclosed code span")
run_test("Mismatched emphasis", "*bold** and **italic*")
run_test("Dangling backslash at EOF", "text\\")
run_test("Only opening brackets", "[[[[[[")
run_test("Deeply nested unclosed brackets", "[a[b[c[d[e[f")

# ══════════════════════════════════════════════════════════════════════
# 2. DEEPLY NESTED STRUCTURES
# ══════════════════════════════════════════════════════════════════════
print("=== 2. Deeply Nested Structures ===")

# Nested lists 10 deep
nested_list = ""
for i in range(10):
    nested_list += "    " * i + "- level " + str(i) + "\n"
run_test("Nested list 10 levels", nested_list)

# Nested lists 20 deep
nested_list_20 = ""
for i in range(20):
    nested_list_20 += "    " * i + "- level " + str(i) + "\n"
run_test("Nested list 20 levels", nested_list_20)

# Nested blockquotes 10 deep
nested_bq = ""
for i in range(10):
    nested_bq += "> " * (i + 1) + "level " + str(i) + "\n"
run_test("Nested blockquotes 10 deep", nested_bq)

# Nested blockquotes 50 deep
nested_bq_50 = "> " * 50 + "deep\n"
run_test("Nested blockquotes 50 deep", nested_bq_50)

# Deeply nested emphasis
nested_em = "*" * 20 + "deep" + "*" * 20
run_test("Nested emphasis 20 deep", nested_em)

nested_em_100 = "*" * 100 + "deep" + "*" * 100
run_test("Nested emphasis 100 deep", nested_em_100)

# Nested strong+em combos
nested_combo = "***" * 10 + "text" + "***" * 10
run_test("Nested ***strong+em*** 10 deep", nested_combo)

# Nested links
nested_links = "[" * 10 + "text" + "]" * 10 + "(url)" * 10
run_test("Nested link brackets 10 deep", nested_links)

# ══════════════════════════════════════════════════════════════════════
# 3. EXTREME SIZES
# ══════════════════════════════════════════════════════════════════════
print("=== 3. Extreme Sizes ===")

run_test("10,000 char single line", "a" * 10000)
run_test("50,000 char single line", "a" * 50000)
run_test("100,000 char single line", "a" * 100000)

# Long line with markdown syntax
run_test("10,000 char line with *emphasis*", "a" * 5000 + " *emphasis* " + "b" * 5000)

# 1000 paragraphs
run_test("1000 paragraphs", "\n\n".join(["paragraph " + str(i) for i in range(1000)]))

# 5000 lines
run_test("5000 simple lines", "\n".join(["line " + str(i) for i in range(5000)]))

# 10000 list items
run_test("10000 list items", "\n".join(["- item " + str(i) for i in range(10000)]))

# Large table
header = "| " + " | ".join(["H" + str(i) for i in range(50)]) + " |"
delim = "| " + " | ".join(["---" for _ in range(50)]) + " |"
rows = "\n".join(["| " + " | ".join(["c" for _ in range(50)]) + " |" for _ in range(200)])
run_test("Large table 50x200", header + "\n" + delim + "\n" + rows)

# ══════════════════════════════════════════════════════════════════════
# 4. MINIMAL / DEGENERATE INPUT
# ══════════════════════════════════════════════════════════════════════
print("=== 4. Minimal/Degenerate Input ===")

run_test("Empty string", "")
run_test("Single space", " ")
run_test("Single newline", "\n")
run_test("Only whitespace (spaces)", "     ")
run_test("Only whitespace (tabs)", "\t\t\t")
run_test("Only whitespace (mixed)", "  \t  \n  \t  \n")
run_test("Single char 'a'", "a")
run_test("Single char '#'", "#")
run_test("Single char '*'", "*")
run_test("Single char '>'", ">")
run_test("Single char '-'", "-")
run_test("Single char '|'", "|")
run_test("Single char '`'", "`")
run_test("Null byte in input", "hello\x00world")
run_test("Only newlines", "\n\n\n\n\n")

# ══════════════════════════════════════════════════════════════════════
# 5. MIXED LINE ENDINGS
# ══════════════════════════════════════════════════════════════════════
print("=== 5. Mixed Line Endings ===")

run_test("CRLF only", "hello\r\nworld\r\n")
run_test("CR only", "hello\rworld\r")
run_test("Mixed CRLF and LF", "line1\r\nline2\nline3\r\nline4\n")
run_test("CRLF in code fence", "```\r\ncode\r\n```\r\n")
run_test("CRLF in list", "- item1\r\n- item2\r\n- item3\r\n")
run_test("CRLF in blockquote", "> quote1\r\n> quote2\r\n")

# ══════════════════════════════════════════════════════════════════════
# 6. UNICODE EDGE CASES
# ══════════════════════════════════════════════════════════════════════
print("=== 6. Unicode Edge Cases ===")

run_test("Emoji in text", "Hello 😀 World 🌍")
run_test("Emoji in emphasis", "*hello 😀 world*")
run_test("Emoji in heading", "# Hello 😀")
run_test("Emoji in link text", "[click 😀](http://example.com)")
run_test("RTL text (Arabic)", "مرحبا بالعالم")
run_test("RTL in emphasis", "*مرحبا بالعالم*")
run_test("Chinese text", "你好世界")
run_test("Chinese in heading", "# 你好世界")
run_test("Zero-width space in emphasis", "*hello\u200bworld*")
run_test("Zero-width joiner in text", "hello\u200dworld")
run_test("Zero-width non-joiner", "hello\u200cworld")
run_test("Combining diacriticals", "héllo wörld")
run_test("Surrogate-adjacent chars", "\U0001F600 text \U0001F4A9")
run_test("Mixed scripts", "Hello мир 世界 مرحبا")
run_test("Math symbols in emphasis", "*∑∫∂*")
run_test("Full-width chars", "＊not emphasis＊")

# ══════════════════════════════════════════════════════════════════════
# 7. HTML INJECTION / MALFORMED HTML
# ══════════════════════════════════════════════════════════════════════
print("=== 7. HTML Injection / Malformed HTML ===")

run_test("Script tag inline", "<script>alert('xss')</script>")
run_test("Script tag in paragraph", "Hello <script>alert(1)</script> world")
run_test("Onclick attribute", '<a href="#" onclick="alert(1)">click</a>')
run_test("Nested script tags", "<script><script>alert(1)</script></script>")
run_test("Malformed HTML tag", "<div class=>broken</div>")
run_test("Unclosed self-closing tag", "<img src=x onerror=alert(1)")
run_test("HTML entities", "&amp; &lt; &gt; &quot; &#39;")
run_test("Invalid HTML entity", "&notavalidentry;")
run_test("CDATA section", "<![CDATA[<sender>John</sender>]]>")
run_test("Processing instruction", "<?xml version='1.0'?>")
run_test("HTML comment with dashes", "<!-- -- -- -->")
run_test("Deeply nested HTML", "<div>" * 100 + "content" + "</div>" * 100)
run_test("Tag with many attributes", '<div ' + ' '.join([f'data-{i}="v"' for i in range(100)]) + '>text</div>')

# ══════════════════════════════════════════════════════════════════════
# 8. ADVERSARIAL / PATHOLOGICAL PATTERNS
# ══════════════════════════════════════════════════════════════════════
print("=== 8. Adversarial/Pathological Patterns ===")

run_test("Many consecutive #", "#" * 100)
run_test("Many consecutive >", ">" * 100)
run_test("Many consecutive *", "*" * 100)
run_test("Many consecutive ~", "~" * 100)
run_test("Many consecutive -", "-" * 100)
run_test("Many consecutive |", "|" * 100)
run_test("Many consecutive [", "[" * 100)
run_test("Many consecutive ]", "]" * 100)
run_test("Many consecutive backticks", "`" * 100)
run_test("Alternating */_ emphasis", "*_*_*_*_*_*_*_*_*_*_" * 10)
run_test("Many reference-style links to missing refs", "\n".join([f"[link{i}][ref{i}]" for i in range(100)]))
run_test("Backslash flood", "\\" * 1000)
run_test("Tab flood", "\t" * 1000)
run_test("Only horizontal rules", "\n\n".join(["---" for _ in range(100)]))
run_test("Interleaved bold/italic", "***a]** b* c ***d** e*")

# ══════════════════════════════════════════════════════════════════════
# REPORT
# ══════════════════════════════════════════════════════════════════════
print("\n" + "=" * 70)
print("STRESS TEST RESULTS")
print("=" * 70)

pass_count = sum(1 for _, s, _ in results if s == "PASS")
fail_count = sum(1 for _, s, _ in results if s == "FAIL")
slow_count = sum(1 for _, s, _ in results if s == "SLOW")

for name, status, detail in results:
    icon = {"PASS": "✓", "FAIL": "✗", "SLOW": "⚠"}[status]
    print(f"  {icon} [{status}] {name}: {detail}")

print(f"\nTotal: {len(results)} tests | PASS: {pass_count} | FAIL: {fail_count} | SLOW: {slow_count}")
