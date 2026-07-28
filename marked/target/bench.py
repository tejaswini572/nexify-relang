import time, marked, re

# escape_html_entities - two-pass search+sub
test = '<b>Hello &amp; world</b> with "quotes" and more <stuff>' * 200
t0 = time.perf_counter()
for _ in range(1000):
    marked.escape_html_entities(test)
t1 = time.perf_counter()
elapsed = (t1 - t0) * 1000
print(f"escape_html_entities (no encode) 1000x on 10k string: {elapsed:.1f}ms")

encode_test = '<>&"' * 500
t0 = time.perf_counter()
for _ in range(1000):
    marked.escape_html_entities(encode_test, True)
t1 = time.perf_counter()
elapsed = (t1 - t0) * 1000
print(f"escape_html_entities (encode) 1000x on 2k string: {elapsed:.1f}ms")

# inline_tokens on emphasis-heavy text
lexer = marked.Lexer()
test3 = '*word* ' * 500
t0 = time.perf_counter()
lexer.inline_tokens(test3)
t1 = time.perf_counter()
elapsed = (t1 - t0) * 1000
print(f"inline_tokens 500 emphasis tokens: {elapsed:.1f}ms")

# inline_tokens on many links
lexer2 = marked.Lexer()
test4 = '[link](http://example.com) ' * 200
t0 = time.perf_counter()
lexer2.inline_tokens(test4)
t1 = time.perf_counter()
elapsed = (t1 - t0) * 1000
print(f"inline_tokens 200 link tokens: {elapsed:.1f}ms")

# split_cells character-by-character loop on wide row
test_row = ' | '.join(['cell content ' + str(i) for i in range(100)])
t0 = time.perf_counter()
for _ in range(1000):
    marked.split_cells(test_row)
t1 = time.perf_counter()
elapsed = (t1 - t0) * 1000
print(f"split_cells 100-col row 1000x: {elapsed:.1f}ms")

# Renderer string building check - check Parser.run_parse
parser = marked.Parser()
tokens = marked.Lexer().run_lex('# h1\n\nparagraph\n\n' * 200)
t0 = time.perf_counter()
parser.run_parse(tokens)
t1 = time.perf_counter()
elapsed = (t1 - t0) * 1000
print(f"run_parse on 400 tokens (200 h1 + 200 para): {elapsed:.1f}ms")
