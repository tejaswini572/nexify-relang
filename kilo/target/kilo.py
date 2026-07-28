#!/usr/bin/env python3
"""
kilo.py — A minimal terminal text editor, Python 3 port of antirez's kilo.

Idiomatic Python rewrite: termios/tty/os for raw terminal control,
dataclasses for state, Pythonic control flow throughout.

Usage: python3 kilo.py <filename>
"""

import os
import sys
import time
import termios
import tty
import signal
import struct
import fcntl
from dataclasses import dataclass, field
from typing import List, Optional

# ---------------------------------------------------------------------------
# Version
# ---------------------------------------------------------------------------
KILO_VERSION = "0.0.1"
KILO_QUIT_TIMES = 3
KILO_QUERY_LEN = 256
TAB_STOP = 8

# ---------------------------------------------------------------------------
# Key codes (mirroring the C enum)
# ---------------------------------------------------------------------------
KEY_NULL    = 0
CTRL_C      = 3
CTRL_D      = 4
CTRL_F      = 6
CTRL_H      = 8
TAB         = 9
CTRL_L      = 12
ENTER       = 13
CTRL_Q      = 17
CTRL_S      = 19
CTRL_U      = 21
ESC         = 27
BACKSPACE   = 127

# Soft key codes (values above normal byte range)
ARROW_LEFT  = 1000
ARROW_RIGHT = 1001
ARROW_UP    = 1002
ARROW_DOWN  = 1003
DEL_KEY     = 1004
HOME_KEY    = 1005
END_KEY     = 1006
PAGE_UP     = 1007
PAGE_DOWN   = 1008

# ---------------------------------------------------------------------------
# Syntax highlight types
# ---------------------------------------------------------------------------
HL_NORMAL   = 0
HL_NONPRINT = 1
HL_COMMENT  = 2
HL_MLCOMMENT= 3
HL_KEYWORD1 = 4
HL_KEYWORD2 = 5
HL_STRING   = 6
HL_NUMBER   = 7
HL_MATCH    = 8

HL_HIGHLIGHT_STRINGS = 1 << 0
HL_HIGHLIGHT_NUMBERS = 1 << 1

# ---------------------------------------------------------------------------
# Syntax highlight colour map: hl_type -> ANSI colour code
# ---------------------------------------------------------------------------
HL_COLOR = {
    HL_COMMENT:    36,   # cyan
    HL_MLCOMMENT:  36,   # cyan
    HL_KEYWORD1:   33,   # yellow
    HL_KEYWORD2:   32,   # green
    HL_STRING:     35,   # magenta
    HL_NUMBER:     31,   # red
    HL_MATCH:      34,   # blue
}

# ---------------------------------------------------------------------------
# Syntax database
# ---------------------------------------------------------------------------
import re as _re

@dataclass
class SyntaxDef:
    name: str
    filematch: List[str]
    keywords: List[str]        # type-2 keywords end with '|'
    singleline_comment: str
    multiline_comment_start: str
    multiline_comment_end: str
    flags: int
    # Compiled keyword regex — built once after construction
    _kw1_pattern: object = field(default=None, init=False, repr=False, compare=False)
    _kw2_pattern: object = field(default=None, init=False, repr=False, compare=False)

    def __post_init__(self) -> None:
        """Pre-compile keyword patterns so per-character matching is O(1) via regex."""
        kw1 = [kw for kw in self.keywords if not kw.endswith('|')]
        kw2 = [kw[:-1] for kw in self.keywords if kw.endswith('|')]
        sep = r'(?=[\s,\.\(\)\+\-\/\*=~%\[\];]|$)'
        if kw1:
            self._kw1_pattern = _re.compile(
                r'(?:' + '|'.join(_re.escape(w) for w in sorted(kw1, key=len, reverse=True)) + r')' + sep
            )
        if kw2:
            self._kw2_pattern = _re.compile(
                r'(?:' + '|'.join(_re.escape(w) for w in sorted(kw2, key=len, reverse=True)) + r')' + sep
            )

HLDB: List[SyntaxDef] = [
    SyntaxDef(
        name="C/C++",
        filematch=[".c", ".h", ".cpp", ".hpp", ".cc"],
        keywords=[
            # C keywords
            "auto", "break", "case", "continue", "default", "do", "else",
            "enum", "extern", "for", "goto", "if", "register", "return",
            "sizeof", "static", "struct", "switch", "typedef", "union",
            "volatile", "while", "NULL",
            # C++ keywords
            "alignas", "alignof", "and", "and_eq", "asm", "bitand", "bitor",
            "class", "compl", "constexpr", "const_cast", "deltype", "delete",
            "dynamic_cast", "explicit", "export", "false", "friend", "inline",
            "mutable", "namespace", "new", "noexcept", "not", "not_eq",
            "nullptr", "operator", "or", "or_eq", "private", "protected",
            "public", "reinterpret_cast", "static_assert", "static_cast",
            "template", "this", "thread_local", "throw", "true", "try",
            "typeid", "typename", "virtual", "xor", "xor_eq",
            # C types (keyword2, marked with trailing '|')
            "int|", "long|", "double|", "float|", "char|", "unsigned|",
            "signed|", "void|", "short|", "auto|", "const|", "bool|",
        ],
        singleline_comment="//",
        multiline_comment_start="/*",
        multiline_comment_end="*/",
        flags=HL_HIGHLIGHT_STRINGS | HL_HIGHLIGHT_NUMBERS,
    ),
]

# ---------------------------------------------------------------------------
# Row representation
# ---------------------------------------------------------------------------
@dataclass
class Row:
    idx: int
    chars: str
    render: str = ""
    hl: List[int] = field(default_factory=list)
    hl_open_comment: bool = False   # was a multiline comment open at end?

# ---------------------------------------------------------------------------
# Editor state
# ---------------------------------------------------------------------------
@dataclass
class Editor:
    cx: int = 0
    cy: int = 0
    rowoff: int = 0
    coloff: int = 0
    screenrows: int = 24
    screencols: int = 80
    rows: List[Row] = field(default_factory=list)
    dirty: int = 0
    filename: Optional[str] = None
    statusmsg: str = ""
    statusmsg_time: float = 0.0
    syntax: Optional[SyntaxDef] = None
    quit_times: int = KILO_QUIT_TIMES

    @property
    def numrows(self) -> int:
        return len(self.rows)

# Single global editor state
E = Editor()

# ---------------------------------------------------------------------------
# Raw terminal mode
# ---------------------------------------------------------------------------
_orig_termios = None

def enable_raw_mode() -> None:
    global _orig_termios
    fd = sys.stdin.fileno()
    _orig_termios = termios.tcgetattr(fd)
    new = termios.tcgetattr(fd)
    # input: no break, no CR-to-NL, no parity, no strip, no flow control
    new[0] &= ~(termios.BRKINT | termios.ICRNL | termios.INPCK |
                termios.ISTRIP | termios.IXON)
    # output: no post-processing
    new[1] &= ~termios.OPOST
    # control: 8-bit chars
    new[2] |= termios.CS8
    # local: no echo, no canonical, no extended, no signals
    new[3] &= ~(termios.ECHO | termios.ICANON | termios.IEXTEN |
                termios.ISIG)
    # read: return per-byte or on 100 ms timeout
    new[6][termios.VMIN]  = 0
    new[6][termios.VTIME] = 1
    termios.tcsetattr(fd, termios.TCSAFLUSH, new)

def disable_raw_mode() -> None:
    if _orig_termios is not None:
        termios.tcsetattr(sys.stdin.fileno(), termios.TCSAFLUSH, _orig_termios)

def at_exit_cleanup() -> None:
    """Restore terminal and clear the screen on exit."""
    disable_raw_mode()
    sys.stdout.write("\x1b[2J\x1b[H")
    sys.stdout.flush()

# ---------------------------------------------------------------------------
# Terminal size detection
# ---------------------------------------------------------------------------
def get_window_size():
    """Return (rows, cols). Falls back to cursor-position query."""
    try:
        buf = fcntl.ioctl(sys.stdout.fileno(), termios.TIOCGWINSZ, b'\x00' * 8)
        rows, cols = struct.unpack('HHHH', buf)[:2]
        if cols > 0 and rows > 0:
            return rows, cols
    except Exception:
        pass
    return _get_cursor_pos_fallback()

def _get_cursor_pos_fallback():
    try:
        sys.stdout.write("\x1b[999C\x1b[999B")
        sys.stdout.flush()
        sys.stdout.write("\x1b[6n")
        sys.stdout.flush()
        # Use bytearray to avoid O(n²) bytes concatenation in the loop
        buf = bytearray()
        while True:
            ch = os.read(sys.stdin.fileno(), 1)
            buf += ch
            if ch == b'R':
                break
        text = buf.decode('ascii', errors='ignore')
        if text.startswith('\x1b['):
            parts = text[2:-1].split(';')
            return int(parts[0]), int(parts[1])
    except Exception:
        pass
    return 24, 80

# ---------------------------------------------------------------------------
# Key reading
# ---------------------------------------------------------------------------
def read_key() -> int:
    """Block until a keypress, decode escape sequences, return key code."""
    fd = sys.stdin.fileno()
    while True:
        try:
            c = os.read(fd, 1)
        except OSError:
            continue
        if c:
            break

    byte = c[0]

    if byte != ESC:
        return byte

    # Try to read escape sequence
    try:
        seq0 = os.read(fd, 1)
        if not seq0:
            return ESC
        seq1 = os.read(fd, 1)
        if not seq1:
            return ESC
    except OSError:
        return ESC

    s0 = seq0.decode('latin-1')
    s1 = seq1.decode('latin-1')

    if s0 == '[':
        if '0' <= s1 <= '9':
            try:
                seq2 = os.read(fd, 1)
                s2 = seq2.decode('latin-1') if seq2 else ''
            except OSError:
                s2 = ''
            if s2 == '~':
                return {
                    '1': HOME_KEY,
                    '3': DEL_KEY,
                    '4': END_KEY,
                    '5': PAGE_UP,
                    '6': PAGE_DOWN,
                    '7': HOME_KEY,
                    '8': END_KEY,
                }.get(s1, ESC)
        else:
            return {
                'A': ARROW_UP,
                'B': ARROW_DOWN,
                'C': ARROW_RIGHT,
                'D': ARROW_LEFT,
                'H': HOME_KEY,
                'F': END_KEY,
            }.get(s1, ESC)
    elif s0 == 'O':
        return {
            'H': HOME_KEY,
            'F': END_KEY,
        }.get(s1, ESC)

    return ESC

# ---------------------------------------------------------------------------
# Syntax highlighting helpers
# ---------------------------------------------------------------------------
def is_separator(ch: str) -> bool:
    return ch == '' or ch.isspace() or ch in ',.()+-/*=~%[];'

def _syntax_to_color(hl: int) -> int:
    return HL_COLOR.get(hl, 37)

def update_syntax(row: Row) -> None:
    """Compute syntax highlights for a single row."""
    render = row.render
    n = len(render)
    hl = [HL_NORMAL] * n
    row.hl = hl

    syntax = E.syntax
    if syntax is None:
        row.hl_open_comment = False
        return

    scs = syntax.singleline_comment
    mcs = syntax.multiline_comment_start
    mce = syntax.multiline_comment_end
    keywords = syntax.keywords
    highlight_strings = bool(syntax.flags & HL_HIGHLIGHT_STRINGS)
    highlight_numbers = bool(syntax.flags & HL_HIGHLIGHT_NUMBERS)

    # Inherit open comment from previous row
    in_comment = (row.idx > 0 and E.rows[row.idx - 1].hl_open_comment)
    in_string = 0   # 0 = not in string; otherwise the quote char ordinal
    prev_sep = True
    i = 0

    while i < n:
        c = render[i]

        # --- Single-line comment ---
        if not in_string and not in_comment:
            if scs and render[i:i + len(scs)] == scs:
                for k in range(i, n):
                    hl[k] = HL_COMMENT
                break

        # --- Multi-line comment ---
        if in_comment:
            hl[i] = HL_MLCOMMENT
            if render[i:i + len(mce)] == mce:
                for k in range(i, i + len(mce)):
                    if k < n:
                        hl[k] = HL_MLCOMMENT
                i += len(mce)
                in_comment = False
                prev_sep = True
            else:
                prev_sep = False
                i += 1
            continue
        elif not in_string and mcs and render[i:i + len(mcs)] == mcs:
            for k in range(i, i + len(mcs)):
                if k < n:
                    hl[k] = HL_MLCOMMENT
            i += len(mcs)
            in_comment = True
            prev_sep = False
            continue

        # --- Strings ---
        if highlight_strings:
            if in_string:
                hl[i] = HL_STRING
                if c == '\\' and i + 1 < n:
                    hl[i + 1] = HL_STRING
                    i += 2
                    prev_sep = False
                    continue
                if ord(c) == in_string:
                    in_string = 0
                i += 1
                continue
            else:
                if c in ('"', "'"):
                    in_string = ord(c)
                    hl[i] = HL_STRING
                    i += 1
                    prev_sep = False
                    continue

        # --- Non-printable ---
        if not c.isprintable():
            hl[i] = HL_NONPRINT
            i += 1
            prev_sep = False
            continue

        # --- Numbers ---
        if highlight_numbers:
            if (c.isdigit() and (prev_sep or (i > 0 and hl[i - 1] == HL_NUMBER))) or \
               (c == '.' and i > 0 and hl[i - 1] == HL_NUMBER):
                hl[i] = HL_NUMBER
                i += 1
                prev_sep = False
                continue

        # --- Keywords — use pre-compiled regex for O(1) lookup per position ---
        if prev_sep:
            matched = False
            for pat, color in (
                (syntax._kw2_pattern, HL_KEYWORD2),
                (syntax._kw1_pattern, HL_KEYWORD1),
            ):
                if pat is None:
                    continue
                m = pat.match(render, i)
                if m:
                    klen = len(m.group(0).rstrip())  # strip lookahead padding
                    # group() includes the lookahead? No — use m.end() - m.start()
                    klen = m.end() - m.start()
                    # The lookahead sep is zero-width — m spans just the word
                    hl[i:i + klen] = [color] * klen
                    i += klen
                    prev_sep = False
                    matched = True
                    break
            if matched:
                continue

        prev_sep = is_separator(c)
        i += 1

    # Propagate open-comment state
    open_comment = in_comment
    if row.hl_open_comment != open_comment:
        row.hl_open_comment = open_comment
        if row.idx + 1 < E.numrows:
            update_syntax(E.rows[row.idx + 1])
    else:
        row.hl_open_comment = open_comment

def select_syntax_highlight(filename: str) -> None:
    """Choose the syntax definition based on the file extension."""
    E.syntax = None
    for syn in HLDB:
        for pat in syn.filematch:
            if pat.startswith('.'):
                if filename.endswith(pat):
                    E.syntax = syn
                    return
            else:
                if pat in filename:
                    E.syntax = syn
                    return

# ---------------------------------------------------------------------------
# Row operations
# ---------------------------------------------------------------------------
def _build_render(chars: str) -> str:
    """Expand tabs to the next TAB_STOP column boundary."""
    out = []
    col = 0
    for ch in chars:
        if ch == '\t':
            out.append(' ')
            col += 1
            while col % TAB_STOP != 0:
                out.append(' ')
                col += 1
        else:
            out.append(ch)
            col += 1
    return ''.join(out)

def update_row(row: Row) -> None:
    row.render = _build_render(row.chars)
    update_syntax(row)

def insert_row(at: int, s: str) -> None:
    if at > E.numrows:
        return
    new_row = Row(idx=at, chars=s)
    E.rows.insert(at, new_row)
    # Only renumber the rows that were actually shifted (at+1 onward).
    # This is still O(n) worst-case but only touches affected rows.
    for i in range(at + 1, len(E.rows)):
        E.rows[i].idx = i
    update_row(new_row)
    E.dirty += 1

def delete_row(at: int) -> None:
    if at >= E.numrows:
        return
    del E.rows[at]
    # Renumber only from the deletion point forward.
    for i in range(at, len(E.rows)):
        E.rows[i].idx = i
    E.dirty += 1

def row_insert_char(row: Row, at: int, ch: str) -> None:
    # Mirror C: if at > current length, pad with spaces first
    if at > len(row.chars):
        pad = at - len(row.chars)
        row.chars = row.chars + ' ' * pad
    at = max(0, at)
    row.chars = row.chars[:at] + ch + row.chars[at:]
    update_row(row)
    E.dirty += 1

def row_del_char(row: Row, at: int) -> None:
    if at < 0 or at >= len(row.chars):
        return
    row.chars = row.chars[:at] + row.chars[at + 1:]
    update_row(row)
    E.dirty += 1

def row_append_string(row: Row, s: str) -> None:
    row.chars += s
    update_row(row)
    E.dirty += 1

def rows_to_string() -> str:
    return '\n'.join(r.chars for r in E.rows) + ('\n' if E.rows else '')

# ---------------------------------------------------------------------------
# Editor-level insert / delete
# ---------------------------------------------------------------------------
def insert_char(ch: str) -> None:
    filerow = E.rowoff + E.cy
    filecol = E.coloff + E.cx
    while E.numrows <= filerow:
        insert_row(E.numrows, '')
    row = E.rows[filerow]
    row_insert_char(row, filecol, ch)
    if E.cx == E.screencols - 1:
        E.coloff += 1
    else:
        E.cx += 1
    E.dirty += 1

def insert_newline() -> None:
    filerow = E.rowoff + E.cy
    filecol = E.coloff + E.cx
    if filerow >= E.numrows:
        if filerow == E.numrows:
            insert_row(filerow, '')
            _fix_cursor_after_newline()
        return
    row = E.rows[filerow]
    filecol = min(filecol, len(row.chars))
    if filecol == 0:
        insert_row(filerow, '')
    else:
        new_chars = row.chars[filecol:]
        row.chars = row.chars[:filecol]
        update_row(row)
        insert_row(filerow + 1, new_chars)
    _fix_cursor_after_newline()

def _fix_cursor_after_newline() -> None:
    if E.cy == E.screenrows - 1:
        E.rowoff += 1
    else:
        E.cy += 1
    E.cx = 0
    E.coloff = 0

def delete_char() -> None:
    filerow = E.rowoff + E.cy
    filecol = E.coloff + E.cx
    if filerow >= E.numrows:
        return
    row = E.rows[filerow]
    if filecol == 0 and filerow == 0:
        return
    if filecol == 0:
        # Backspace at column 0: merge current line into previous
        prev = E.rows[filerow - 1]
        new_cx = len(prev.chars)           # cursor lands at old end of prev line
        row_append_string(prev, row.chars)
        delete_row(filerow)
        # Move cursor up one row
        if E.cy == 0:
            # cy==0 means we were scrolled, so rowoff > 0 (guaranteed: filerow>0)
            E.rowoff -= 1
        else:
            E.cy -= 1
        # Set horizontal position to where the two lines joined
        E.cx = new_cx
        if E.cx >= E.screencols:
            shift = (E.cx - E.screencols) + 1
            E.coloff += shift
            E.cx -= shift
    else:
        row_del_char(row, filecol - 1)
        if E.cx == 0 and E.coloff:
            E.coloff -= 1
        else:
            E.cx -= 1

# ---------------------------------------------------------------------------
# File I/O
# ---------------------------------------------------------------------------
def editor_open(filename: str) -> None:
    E.filename = filename
    E.dirty = 0
    E.rows.clear()
    select_syntax_highlight(filename)
    try:
        with open(filename, 'r', errors='replace') as f:
            lines = f.read().splitlines()
        # Bulk-construct all rows in one pass — avoids O(n²) index renumbering
        # that would occur if we called insert_row() per line.
        E.rows = [Row(idx=i, chars=line) for i, line in enumerate(lines)]
        for row in E.rows:
            # Build render (tab expansion) without triggering dirty or index fixup.
            row.render = _build_render(row.chars)
        # Single forward syntax pass: each row may depend on previous open-comment state.
        for row in E.rows:
            update_syntax(row)
    except FileNotFoundError:
        pass   # New file — start empty
    E.dirty = 0

def editor_save() -> None:
    if E.filename is None:
        set_status_message("No filename — cannot save.")
        return
    content = rows_to_string()
    try:
        with open(E.filename, 'w') as f:
            f.write(content)
        E.dirty = 0
        set_status_message(f"{len(content)} bytes written to disk.")
    except OSError as exc:
        set_status_message(f"Can't save! I/O error: {exc.strerror}")

# ---------------------------------------------------------------------------
# Status message
# ---------------------------------------------------------------------------
def set_status_message(msg: str) -> None:
    E.statusmsg = msg
    E.statusmsg_time = time.time()

# ---------------------------------------------------------------------------
# Rendering / Screen refresh
# ---------------------------------------------------------------------------
def _cx_to_render_x(row: Row, cx: int) -> int:
    """Convert a file-column (cx) to a rendered column accounting for tabs."""
    rx = 0
    for j, ch in enumerate(row.chars):
        if j >= cx:
            break
        if ch == '\t':
            rx += TAB_STOP - (rx % TAB_STOP)
        else:
            rx += 1
    return rx

def _screen_col_for_cursor(row: Row) -> int:
    """Return the screen column (0-based) for the cursor in a single pass.

    Equivalent to _cx_to_render_x(row, coloff+cx) - _cx_to_render_x(row, coloff)
    but traverses row.chars only once instead of twice.
    """
    coloff = E.coloff
    target = coloff + E.cx
    rx = 0          # render column from start of line
    screen_rx = 0   # render column relative to left edge of screen
    for j, ch in enumerate(row.chars):
        if j >= target:
            break
        step = TAB_STOP - (rx % TAB_STOP) if ch == '\t' else 1
        rx += step
        if j >= coloff:
            screen_rx += step
    return screen_rx

def refresh_screen() -> None:
    """Redraw the entire screen with a single write to avoid flicker."""
    buf = []

    buf.append("\x1b[?25l")   # hide cursor
    buf.append("\x1b[H")      # go home

    for y in range(E.screenrows):
        filerow = E.rowoff + y

        if filerow >= E.numrows:
            if E.numrows == 0 and y == E.screenrows // 3:
                welcome = f"Kilo editor -- version {KILO_VERSION}"
                padding = (E.screencols - len(welcome)) // 2
                if padding:
                    buf.append("~")
                    buf.append(" " * (padding - 1))
                buf.append(welcome[:E.screencols])
            else:
                buf.append("~")
            buf.append("\x1b[0K\r\n")
            continue

        row = E.rows[filerow]
        render = row.render
        hl     = row.hl

        start = E.coloff
        end   = start + E.screencols
        render_slice = render[start:end]
        hl_slice = hl[start:start + len(render_slice)] if hl else []

        current_color = -1
        for j, ch in enumerate(render_slice):
            h = hl_slice[j] if j < len(hl_slice) else HL_NORMAL
            if h == HL_NONPRINT:
                buf.append("\x1b[7m")
                sym = chr(ord('@') + ord(ch)) if ord(ch) <= 26 else '?'
                buf.append(sym)
                buf.append("\x1b[0m")
                current_color = -1
            elif h == HL_NORMAL:
                if current_color != -1:
                    buf.append("\x1b[39m")
                    current_color = -1
                buf.append(ch)
            else:
                color = _syntax_to_color(h)
                if color != current_color:
                    buf.append(f"\x1b[{color}m")
                    current_color = color
                buf.append(ch)

        buf.append("\x1b[39m")
        buf.append("\x1b[0K")
        buf.append("\r\n")

    # --- Status bar ---
    buf.append("\x1b[0K")
    buf.append("\x1b[7m")
    fname = E.filename or "[No Name]"
    modified = " (modified)" if E.dirty else ""
    status  = f"{fname:.20s} - {E.numrows} lines{modified}"
    rstatus = f"{E.rowoff + E.cy + 1}/{E.numrows}"
    if len(status) > E.screencols:
        status = status[:E.screencols]
    buf.append(status)
    total_pad = E.screencols - len(status)
    rlen = len(rstatus)
    if total_pad >= rlen:
        # Fill with spaces up to where rstatus starts, then rstatus
        buf.append(" " * (total_pad - rlen))
        buf.append(rstatus)
    else:
        buf.append(" " * total_pad)
    buf.append("\x1b[0m\r\n")

    # --- Message line ---
    buf.append("\x1b[0K")
    if E.statusmsg and (time.time() - E.statusmsg_time < 5):
        buf.append(E.statusmsg[:E.screencols])

    # --- Reposition cursor (single traversal of row.chars) ---
    filerow = E.rowoff + E.cy
    rx = 0
    if filerow < E.numrows:
        rx = _screen_col_for_cursor(E.rows[filerow])

    buf.append(f"\x1b[{E.cy + 1};{rx + 1}H")
    buf.append("\x1b[?25h")

    out = ''.join(buf)
    os.write(sys.stdout.fileno(), out.encode('utf-8', errors='replace'))

# ---------------------------------------------------------------------------
# Incremental search
# ---------------------------------------------------------------------------
def editor_find() -> None:
    saved_cx     = E.cx
    saved_cy     = E.cy
    saved_coloff = E.coloff
    saved_rowoff = E.rowoff

    query = ""
    last_match   = -1
    find_next    = 0
    saved_hl_row = None
    saved_hl     = None

    def restore_hl():
        nonlocal saved_hl, saved_hl_row
        if saved_hl is not None and saved_hl_row is not None:
            E.rows[saved_hl_row].hl = saved_hl[:]
            saved_hl = None
            saved_hl_row = None

    while True:
        set_status_message(f"Search: {query} (ESC/Arrows/Enter)")
        refresh_screen()

        c = read_key()

        if c in (DEL_KEY, CTRL_H, BACKSPACE):
            if query:
                query = query[:-1]
            last_match = -1
        elif c in (ESC, ENTER):
            if c == ESC:
                E.cx     = saved_cx
                E.cy     = saved_cy
                E.coloff = saved_coloff
                E.rowoff = saved_rowoff
            restore_hl()
            set_status_message("")
            return
        elif c in (ARROW_RIGHT, ARROW_DOWN):
            find_next = 1
        elif c in (ARROW_LEFT, ARROW_UP):
            find_next = -1
        elif 32 <= c <= 126:
            if len(query) < KILO_QUERY_LEN:
                query += chr(c)
            last_match = -1
        else:
            find_next = 0
            continue

        if last_match == -1:
            find_next = 1

        if find_next and query:
            current = last_match
            for _ in range(E.numrows):
                current += find_next
                if current == -1:
                    current = E.numrows - 1
                elif current == E.numrows:
                    current = 0
                row = E.rows[current]
                match_pos = row.render.find(query)
                if match_pos != -1:
                    restore_hl()
                    last_match    = current
                    saved_hl_row  = current
                    saved_hl      = row.hl[:]
                    mlen = len(query)
                    for k in range(match_pos, match_pos + mlen):
                        if k < len(row.hl):
                            row.hl[k] = HL_MATCH
                    E.cy     = 0
                    E.cx     = match_pos
                    E.rowoff = current
                    E.coloff = 0
                    if E.cx >= E.screencols:
                        diff = E.cx - E.screencols + 1
                        E.cx     -= diff
                        E.coloff += diff
                    break
            find_next = 0

# ---------------------------------------------------------------------------
# Cursor movement
# ---------------------------------------------------------------------------
def move_cursor(key: int) -> None:
    filerow = E.rowoff + E.cy
    filecol = E.coloff + E.cx
    row = E.rows[filerow] if filerow < E.numrows else None

    if key == ARROW_LEFT:
        if E.cx == 0:
            if E.coloff:
                E.coloff -= 1
            elif filerow > 0:
                E.cy -= 1
                prev = E.rows[filerow - 1]
                E.cx = len(prev.chars)
                if E.cx > E.screencols - 1:
                    E.coloff = E.cx - E.screencols + 1
                    E.cx = E.screencols - 1
        else:
            E.cx -= 1

    elif key == ARROW_RIGHT:
        if row and filecol < len(row.chars):
            if E.cx == E.screencols - 1:
                E.coloff += 1
            else:
                E.cx += 1
        elif row and filecol == len(row.chars):
            E.cx = 0
            E.coloff = 0
            if E.cy == E.screenrows - 1:
                E.rowoff += 1
            else:
                E.cy += 1

    elif key == ARROW_UP:
        if E.cy == 0:
            if E.rowoff:
                E.rowoff -= 1
        else:
            E.cy -= 1

    elif key == ARROW_DOWN:
        if filerow < E.numrows:
            if E.cy == E.screenrows - 1:
                E.rowoff += 1
            else:
                E.cy += 1

    # Clamp cx to row length
    filerow = E.rowoff + E.cy
    row = E.rows[filerow] if filerow < E.numrows else None
    rowlen = len(row.chars) if row else 0
    filecol = E.coloff + E.cx
    if filecol > rowlen:
        E.cx -= filecol - rowlen
        if E.cx < 0:
            E.coloff += E.cx
            E.cx = 0

# ---------------------------------------------------------------------------
# Key processing (main event loop body)
# ---------------------------------------------------------------------------
def process_keypress() -> None:
    c = read_key()

    if c == ENTER:
        insert_newline()

    elif c == CTRL_Q:
        # Require KILO_QUIT_TIMES extra presses when there are unsaved changes.
        # quit_times starts at KILO_QUIT_TIMES; each press decrements it;
        # any other keypress resets it (at bottom of this function).
        if E.dirty and E.quit_times > 0:
            set_status_message(
                f"WARNING!!! File has unsaved changes. "
                f"Press Ctrl-Q {E.quit_times} more times to quit."
            )
            E.quit_times -= 1
            return   # do NOT reset quit_times — we returned early
        at_exit_cleanup()
        sys.exit(0)

    elif c == CTRL_S:
        editor_save()

    elif c == CTRL_F:
        editor_find()

    elif c in (BACKSPACE, CTRL_H, DEL_KEY):
        if c == DEL_KEY:
            # Delete-key: move right then backspace (same as C)
            move_cursor(ARROW_RIGHT)
        delete_char()

    elif c in (PAGE_UP, PAGE_DOWN):
        # C: first snap cy to top/bottom of screen, then scroll a full page
        if c == PAGE_UP:
            E.cy = 0
        else:
            E.cy = E.screenrows - 1
        times = E.screenrows
        direction = ARROW_UP if c == PAGE_UP else ARROW_DOWN
        for _ in range(times):
            move_cursor(direction)

    elif c == HOME_KEY:
        E.cx = 0
        E.coloff = 0

    elif c == END_KEY:
        filerow = E.rowoff + E.cy
        if filerow < E.numrows:
            row = E.rows[filerow]
            rlen = len(row.chars)
            if rlen <= E.screencols - 1:
                E.cx = rlen
                E.coloff = 0
            else:
                E.cx = E.screencols - 1
                E.coloff = rlen - E.screencols + 1

    elif c in (ARROW_UP, ARROW_DOWN, ARROW_LEFT, ARROW_RIGHT):
        move_cursor(c)

    elif c in (CTRL_C, CTRL_L, ESC, KEY_NULL):
        pass   # ignored (Ctrl-C is intentionally swallowed like in C)

    else:
        if 32 <= c <= 126 or c == TAB:
            insert_char(chr(c))

    # Reset quit counter on every non-Ctrl-Q keypress (or clean Ctrl-Q exit)
    E.quit_times = KILO_QUIT_TIMES

# ---------------------------------------------------------------------------
# Window resize signal handler
# ---------------------------------------------------------------------------
def handle_sigwinch(signum, frame) -> None:
    rows, cols = get_window_size()
    E.screenrows = rows - 2
    E.screencols = cols
    E.cy = min(E.cy, E.screenrows - 1)
    E.cx = min(E.cx, E.screencols - 1)
    refresh_screen()

# ---------------------------------------------------------------------------
# Initialisation
# ---------------------------------------------------------------------------
def init_editor() -> None:
    rows, cols = get_window_size()
    E.screenrows = rows - 2
    E.screencols = cols
    signal.signal(signal.SIGWINCH, handle_sigwinch)

# ---------------------------------------------------------------------------
# Entry point
# ---------------------------------------------------------------------------
def main() -> None:
    if len(sys.argv) != 2:
        print("Usage: python3 kilo.py <filename>", file=sys.stderr)
        sys.exit(1)

    filename = sys.argv[1]

    init_editor()
    editor_open(filename)

    enable_raw_mode()
    import atexit
    atexit.register(at_exit_cleanup)

    set_status_message("HELP: Ctrl-S = save | Ctrl-Q = quit | Ctrl-F = find")

    while True:
        refresh_screen()
        process_keypress()


if __name__ == '__main__':
    main()
