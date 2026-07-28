import marked
t = marked.Tokenizer()
t.lexer = marked.Lexer()
t.lexer.tokenizer = t
src = r'~not strike\~~'

def debug_deltok(src, masked_src):
    i = 0
    while i < len(src) and src[i] == '~': i += 1
    a = i
    t = masked_src[len(masked_src) - len(src) + i:]
    idx = 0
    print("i:", i)
    while idx < len(t):
        ch = t[idx]
        if ch != '~':
            idx += 1
            continue
        u = 0
        while idx + u < len(t) and t[idx + u] == '~': u += 1
        print("Found run of ~ at", idx, "len", u)
        if u != i:
            idx += u
            continue
        before = t[idx - 1] if idx > 0 else ''
        after_ch = t[idx + u] if idx + u < len(t) else ''
        before_ws = not before or marked._is_unicode_ws(before)
        before_ps = before and marked._is_unicode_punct_sym(before)
        after_ws = not after_ch or marked._is_unicode_ws(after_ch)
        after_ps = after_ch and marked._is_unicode_punct_sym(after_ch)
        right = not before_ws and (not before_ps or after_ws or after_ps)
        left = not after_ws and (not after_ps or before_ws or before_ps)
        print(f"before={repr(before)} after={repr(after_ch)} right={right} left={left}")
        idx += u

debug_deltok(src, src)
