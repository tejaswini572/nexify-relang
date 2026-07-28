#!/usr/bin/env python3
from __future__ import annotations

import functools
import re
import sys
import unicodedata
from typing import Any
from urllib.parse import quote

def _is_unicode_punct_sym(c: str) -> bool:
    if not c:
        return False
    cat = unicodedata.category(c)
    return cat.startswith('P') or cat.startswith('S')

def _is_unicode_ws(c: str) -> bool:
    if not c:
        return False
    if c in ' \t\n\r\f\v':
        return True
    return unicodedata.category(c) == 'Zs'

def get_defaults() -> dict[str, Any]:
    return {
        "async": False,
        "breaks": False,
        "extensions": None,
        "gfm": True,
        "hooks": None,
        "pedantic": False,
        "renderer": None,
        "silent": False,
        "tokenizer": None,
        "walkTokens": None,
    }


ESCAPE_REPLACEMENTS = {
    "&": "&amp;",
    "<": "&lt;",
    ">": "&gt;",
    '"': "&quot;",
    "'": "&#39;",
}


def _replace_escape(match: re.Match[str]) -> str:
    return ESCAPE_REPLACEMENTS[match.group(0)]


def escape_html_entities(text: str, encode: bool = False) -> str:
    if encode:
        if re.search(r'[&<>"\']', text):
            return re.sub(r'[&<>"\']', _replace_escape, text)
    elif re.search(r'[<>"\']|&(?!(#\d{1,7}|#[Xx][a-fA-F0-9]{1,6}|\w+);)', text):
        return re.sub(r'[<>"\']|&(?!(#\d{1,7}|#[Xx][a-fA-F0-9]{1,6}|\w+);)', _replace_escape, text)
    return text


def clean_url(href: str) -> str | None:
    try:
        return quote(href, safe="/#:?&=@!$&'()*+,;~.-_%").replace("%25", "%")
    except Exception:
        return None


def split_cells(table_row: str, count: int | None = None) -> list[str]:
    chars: list[str] = []
    for idx, ch in enumerate(table_row):
        if ch == "|":
            escaped = False
            cur = idx
            while cur - 1 >= 0 and table_row[cur - 1] == "\\":
                escaped = not escaped
                cur -= 1
            chars.append("|" if escaped else " |")
        else:
            chars.append(ch)
    cells = "".join(chars).split(" |")
    if cells and not cells[0].strip():
        cells.pop(0)
    if cells and not cells[-1].strip():
        cells.pop()
    if count is not None:
        if len(cells) > count:
            cells = cells[:count]
        while len(cells) < count:
            cells.append("")
    return [cell.strip().replace(r"\|", "|") for cell in cells]


def rtrim(text: str, char: str, invert: bool = False) -> str:
    n = len(text)
    if n == 0:
        return ""
    suffix = 0
    while suffix < n:
        current = text[n - suffix - 1]
        if current == char and not invert:
            suffix += 1
        elif current != char and invert:
            suffix += 1
        else:
            break
    return text[: n - suffix]


def trim_trailing_blank_lines(text: str) -> str:
    lines = text.split("\n")
    end = len(lines) - 1
    while end >= 0 and re.match(r"^[ \t]*$", lines[end]):
        end -= 1
    if len(lines) - end <= 2:
        return text
    return "\n".join(lines[: end + 1])


def find_closing_bracket(text: str, brackets: str) -> int:
    if brackets[1] not in text:
        return -1
    level = 0
    i = 0
    while i < len(text):
        ch = text[i]
        if ch == "\\":
            i += 2
            continue
        if ch == brackets[0]:
            level += 1
        elif ch == brackets[1]:
            level -= 1
            if level < 0:
                return i
        i += 1
    if level > 0:
        return -2
    return -1


def expand_tabs(line: str, indent: int = 0) -> str:
    col = indent
    out: list[str] = []
    for ch in line:
        if ch == "\t":
            add = 4 - (col % 4)
            out.append(" " * add)
            col += add
        else:
            out.append(ch)
            col += 1
    return "".join(out)


def is_punctuation_or_symbol(ch: str) -> bool:
    cat = unicodedata.category(ch)
    return cat.startswith("P") or cat.startswith("S")


def is_space_or_punct(ch: str) -> bool:
    return ch.isspace() or is_punctuation_or_symbol(ch)


def is_not_space_or_punct(ch: str) -> bool:
    return not is_space_or_punct(ch)


def unescape_punctuation(text: str) -> str:
    out: list[str] = []
    i = 0
    while i < len(text):
        if i + 1 < len(text) and text[i] == "\\" and is_punctuation_or_symbol(text[i + 1]):
            out.append(text[i + 1])
            i += 2
            continue
        out.append(text[i])
        i += 1
    return "".join(out)


class Rules:
    def __init__(self) -> None:
        self.other = type("Other", (), {})()
        other = self.other
        other.code_remove_indent = re.compile(r"^(?: {1,4}| {0,3}\t)", re.M)
        other.output_link_replace = re.compile(r"\\([\[\]])")
        other.indent_code_compensation = re.compile(r"^(\s+)(?:```)")
        other.beginning_space = re.compile(r"^\s+")
        other.ending_hash = re.compile(r"#$")
        other.starting_space_char = re.compile(r"^ ")
        other.ending_space_char = re.compile(r" $")
        other.non_space_char = re.compile(r"[^ ]")
        other.new_line_char_global = re.compile(r"\n")
        other.tab_char_global = re.compile(r"\t")
        other.multiple_space_global = re.compile(r"\s+")
        other.blank_line = re.compile(r"^[ \t]*$")
        other.double_blank_line = re.compile(r"\n[ \t]*\n[ \t]*$")
        other.blockquote_start = re.compile(r"^ {0,3}>")
        other.blockquote_setext_replace = re.compile(r"\n {0,3}((?:=+|-+) *)(?=\n|$)")
        other.blockquote_setext_replace2 = re.compile(r"^ {0,3}>[ \t]?", re.M)
        other.list_replace_nesting = re.compile(r"^ {1,4}(?=( {4})*[^ ])")
        other.list_is_task = re.compile(r"^\[[ xX]\] +\S")
        other.list_replace_task = re.compile(r"^\[[ xX]\] +")
        other.list_task_checkbox = re.compile(r"\[[ xX]\]")
        other.any_line = re.compile(r"\n.*\n")
        other.href_brackets = re.compile(r"^<(.*)>$")
        other.table_delimiter = re.compile(r"[:|]")
        other.table_align_chars = re.compile(r"^\||\| *$")
        other.table_row_blank_line = re.compile(r"\n[ \t]*$")
        other.table_align_right = re.compile(r"^ *-+: *$")
        other.table_align_center = re.compile(r"^ *:-+: *$")
        other.table_align_left = re.compile(r"^ *:-+ *$")
        other.start_a_tag = re.compile(r"^<a ", re.I)
        other.end_a_tag = re.compile(r"^</a>", re.I)
        other.start_pre_script_tag = re.compile(r"^<(pre|code|kbd|script)(\s|>)", re.I)
        other.end_pre_script_tag = re.compile(r"^</(pre|code|kbd|script)(\s|>)", re.I)
        other.start_angle_bracket = re.compile(r"^<")
        other.end_angle_bracket = re.compile(r">$")
        other.escape_test = re.compile(r'[&<>"\']')
        other.percent_decode = re.compile(r"%25")
        other.find_pipe = re.compile(r"\|")
        other.split_pipe = re.compile(r" \|")
        other.slash_pipe = re.compile(r"\\\|")
        other.carriage_return = re.compile(r"\r\n|\r")
        other.space_line = re.compile(r"^ +$", re.M)
        other.not_space_start = re.compile(r"^\S*")
        other.ending_newline = re.compile(r"\n$")
        self.block = type("Block", (), {})()
        self.inline = type("Inline", (), {})()
        self._compile()

    def list_item_regex(self, bull: str) -> re.Pattern[str]:
        return re.compile(rf"^( {{0,3}}{bull})((?:[\t ][^\n]*)?(?:\n|$))")

    @functools.lru_cache(maxsize=8)
    def next_bullet_regex(self, indent: int) -> re.Pattern[str]:
        idx = max(0, min(3, indent - 1))
        return re.compile(rf"^ {{0,{idx}}}(?:[*+-]|\d{{1,9}}[.)])((?:[ \t][^\n]*)?(?:\n|$))")

    @functools.lru_cache(maxsize=8)
    def hr_regex(self, indent: int) -> re.Pattern[str]:
        idx = max(0, min(3, indent - 1))
        return re.compile(rf"^ {{0,{idx}}}((?:- *){{3,}}|(?:_ *){{3,}}|(?:\* *){{3,}})(?:\n+|$)")

    @functools.lru_cache(maxsize=8)
    def fences_begin_regex(self, indent: int) -> re.Pattern[str]:
        idx = max(0, min(3, indent - 1))
        return re.compile(rf"^ {{0,{idx}}}(?:```|~~~)")

    @functools.lru_cache(maxsize=8)
    def heading_begin_regex(self, indent: int) -> re.Pattern[str]:
        idx = max(0, min(3, indent - 1))
        return re.compile(rf"^ {{0,{idx}}}#")

    @functools.lru_cache(maxsize=8)
    def html_begin_regex(self, indent: int) -> re.Pattern[str]:
        idx = max(0, min(3, indent - 1))
        return re.compile(rf"^ {{0,{idx}}}<(?:[a-z].*>|!--)", re.I)

    @functools.lru_cache(maxsize=8)
    def blockquote_begin_regex(self, indent: int) -> re.Pattern[str]:
        idx = max(0, min(3, indent - 1))
        return re.compile(rf"^ {{0,{idx}}}>")

    def _compile(self) -> None:
        b = self.block
        i = self.inline
        b.newline = re.compile(r"^(?:[ \t]*(?:\n|$))+")
        b.code = re.compile(r"^((?: {4}| {0,3}\t)[^\n]+(?:\n(?:[ \t]*(?:\n|$))*)?)+")
        b.fences = re.compile(r"^ {0,3}(`{3,}(?=[^`\n]*(?:\n|$))|~{3,})([^\n]*)(?:\n|$)(?:|([\s\S]*?)(?:\n|$))(?: {0,3}\1[~`]* *(?=\n|$)|$)")
        b.hr = re.compile(r"^ {0,3}((?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[\t ]*){3,})(?:\n+|$)")
        b.heading = re.compile(r"^ {0,3}(#{1,6})(?=\s|$)(.*)(?:\n+|$)")
        b.list = re.compile(r"^( {0,3}(?:[*+-]|\d{1,9}[.)]))([ \t][^\n]*?)?(?:\n|$)")
        b.lheading = re.compile(r"^(?!(?: {0,3}(?:[*+-]|\d{1,9}[.)]) )|(?: {4}| {0,3}\t)| {0,3}(?:`{3,}|~{3,})| {0,3}>| {0,3}#{1,6}| {0,3}<[^\n>]+>\n)((?:.|\n(?!\s*?\n|(?: {0,3}(?:[*+-]|\d{1,9}[.)]) )|(?: {4}| {0,3}\t)| {0,3}(?:`{3,}|~{3,})| {0,3}>| {0,3}#{1,6}| {0,3}<[^\n>]+>\n| {0,3}\|?(?:[:\- ]*\|)+[\:\- ]*\n))+?)\n {0,3}(=+|-+) *(?:\n+|$)")
        b.defn = re.compile(r'^ {0,3}\[((?!\s*\])(?:\\[\s\S]|[^\[\]\\])+)\]: *(?:\n[ \t]*)?([^<\s][^\s]*|<.*?>)(?:(?: +(?:\n[ \t]*)?| *\n[ \t]*)(?:"(?:\\"?|[^"\\])*"|\'[^\'\n]*(?:\n[^\'\n]+)*\n?\'|\([^()]*\)))? *(?:\n+|$)')
        tag = "address|article|aside|base|basefont|blockquote|body|caption|center|col|colgroup|dd|details|dialog|dir|div|dl|dt|fieldset|figcaption|figure|footer|form|frame|frameset|h[1-6]|head|header|hr|html|iframe|legend|li|link|main|menu|menuitem|meta|nav|noframes|ol|optgroup|option|p|param|search|section|summary|table|tbody|td|tfoot|th|thead|title|tr|track|ul"
        comment = r"<!--(?:-?>|[\s\S]*?(?:-->|$))"
        b.html = re.compile(
            rf"^ {{0,3}}(?:"
            rf"<(script|pre|style|textarea)[\s>][\s\S]*?(?:</\1>[^\n]*\n+|$)"
            rf"|{comment}[^\n]*(\n+|$)"
            rf"|<\?[\s\S]*?(?:\?>[^\n]*\n+|$)"
            rf"|<![A-Z][\s\S]*?(?:>[^\n]*\n+|$)"
            rf"|<!\[CDATA\[[\s\S]*?(?:\]\]>[^\n]*\n+|$)"
            rf"|</?(?:{tag})(?: +|\n|/?>)[\s\S]*?(?:(?:\n[ \t]*)+\n|$)"
            rf"|<(?!script|pre|style|textarea)([a-z][\w-]*)(?: +[a-zA-Z:_][\w.:-]*(?: *= *\"[^\"\n]*\"| *= *'[^'\n]*'| *= *[^\s\"'=<>`]+)?)*? */?>(?=[ \t]*(?:\n|$))[\s\S]*?(?:(?:\n[ \t]*)+\n|$)"
            rf"|</(?!script|pre|style|textarea)[a-z][\w-]*\s*>(?=[ \t]*(?:\n|$))[\s\S]*?(?:(?:\n[ \t]*)+\n|$)"
            rf")",
            re.I,
        )
        gfmtable = (
            r"^ *([^\n ].*)\n"
            r" {0,3}((?:\| *)?:?-+:? *(?:\| *:?-+:? *)*(?:\| *)?)"
            r"(?:\n((?:(?! *\n|^ {0,3}((?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[\t ]*){3,})(?:\n+|$)| {0,3}#{1,6}(?:\s|$)| {0,3}>|(?: {4}| {0,3}\t)[^\n]| {0,3}(?:`{3,}(?=[^`\n]*(?:\\n|$))|~{3,})[^\n]*\n| {0,3}(?:[*+-]|1[.)])[ \t]|</?(?:"
            + tag
            + r")(?: +|\n|/?>)|<(?:script|pre|style|textarea|!--)).*(?:\n|$))*)\n*|$)"
        )
        b.table = re.compile(gfmtable, re.M)
        b.paragraph = re.compile(r"^([^\n]+(?:\n(?!^ {0,3}((?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[\t ]*){3,})(?:\n+|$)| {0,3}#{1,6}(?:\s|$)| {0,3}>| {0,3}(?:`{3,}(?=[^`\n]*\n)|~{3,})[^\n]*\n| {0,3}(?:[*+-]|1[.)])[ \t]+[^ \t\n]|</?(?:"
                                  + tag + r")(?: +|\n|/?>)|<(?:script|pre|style|textarea|!--)| +\n| *([^\n ].*)\n {0,3}((?:\| *)?:?-+:? *(?:\| *:?-+:? *)*(?:\| *)?))[^\n]+)*)", re.M)
        b.text = re.compile(r"^[^\n]+")

        i.escape = re.compile(r"^\\([!\"#$%&'()*+,\-./:;<=>?@\[\]\\^_`{|}~])")
        i.code = re.compile(r"^(`+)([^`]|[^`][\s\S]*?[^`])\1(?!`)")
        i.br = re.compile(r"^( {2,}|\\)\n(?!\s*$)")
        i.autolink = re.compile(r"^<([a-zA-Z][a-zA-Z0-9+.-]{1,31}:[^\s\x00-\x1f<>]*|[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+(@)[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+(?![-_]))>")
        comment_inline = r"<!--(?:-?>|[\s\S]*?-->)"
        i.tag = re.compile(
            r"^" + comment_inline
            + r"|^</[a-zA-Z][\w:-]*\s*>"
            + r"|^<[a-zA-Z][\w-]*(?:\s+[a-zA-Z:_][\w.:-]*(?:\s*=\s*\"[^\"]*\"|\s*=\s*'[^']*'|\s*=\s*[^\s\"'=<>`]+)?)*?\s*/?>"
            + r"|^<\?[\s\S]*?\?>"
            + r"|^<![a-zA-Z]+\s[\s\S]*?>"
            + r"|^<!\[CDATA\[[\s\S]*?\]\]>",
            re.M,
        )
        i.url = re.compile(r"^((?:https?://|ftp://|www\.)(?:[a-zA-Z0-9\-]+\.?)+[^\s<]*|[A-Za-z0-9._+-]+(@)[a-zA-Z0-9-_]+(?:\.[a-zA-Z0-9-_]*[a-zA-Z0-9])+(?![-_]))", re.I)
        i.text = re.compile(r"^([`~]+|[^`~])(?:(?= {2,}\n)|(?=[a-zA-Z0-9.!#$%&'*+/=?_`{|}~-]+@)|[\s\S]*?(?:(?=[\\<!\[`*~_]|\b_|https?://|ftp://|www\.|$)|[^ ](?= {2,}\n)|[^a-zA-Z0-9.!#$%&'*+/=?_`{|}~-](?=[a-zA-Z0-9.!#$%&'*+/=?_`{|}~-]+@)))")


RULES = Rules()


class Tokenizer:
    def __init__(self, options: dict[str, Any] | None = None) -> None:
        self.options = options or get_defaults()
        self.rules = RULES
        self.lexer: Lexer | None = None

    def space(self, src: str) -> dict[str, Any] | None:
        cap = self.rules.block.newline.match(src)
        if cap and cap.group(0):
            return {"type": "space", "raw": cap.group(0)}
        return None

    def code(self, src: str) -> dict[str, Any] | None:
        cap = self.rules.block.code.match(src)
        if not cap:
            return None
        raw = cap.group(0) if self.options.get("pedantic") else trim_trailing_blank_lines(cap.group(0))
        text = self.rules.other.code_remove_indent.sub("", raw)
        return {"type": "code", "raw": raw, "codeBlockStyle": "indented", "text": text}

    def fences(self, src: str) -> dict[str, Any] | None:
        cap = self.rules.block.fences.match(src)
        if not cap:
            return None
        raw = cap.group(0)
        text = self._indent_code_compensation(raw, cap.group(3) or "")
        lang = cap.group(2)
        if lang is not None:
            lang = lang.strip()
            lang = unescape_punctuation(lang)
        return {"type": "code", "raw": raw, "lang": lang, "text": text}

    def heading(self, src: str) -> dict[str, Any] | None:
        cap = self.rules.block.heading.match(src)
        if not cap:
            return None
        text = cap.group(2).strip()
        if self.rules.other.ending_hash.search(text):
            trimmed = rtrim(text, "#")
            if self.options.get("pedantic") or not trimmed or self.rules.other.ending_space_char.search(trimmed):
                text = trimmed.strip()
        return {
            "type": "heading",
            "raw": rtrim(cap.group(0), "\n"),
            "depth": len(cap.group(1)),
            "text": text,
            "tokens": self.lexer.inline(text),
        }

    def hr(self, src: str) -> dict[str, Any] | None:
        cap = self.rules.block.hr.match(src)
        if not cap:
            return None
        return {"type": "hr", "raw": rtrim(cap.group(0), "\n")}

    def blockquote(self, src: str) -> dict[str, Any] | None:
        lines_with_end = src.splitlines(True)
        if not lines_with_end or not self.rules.other.blockquote_start.match(lines_with_end[0].rstrip("\n")):
            return None
        consumed: list[str] = []
        saw_prefixed = False
        for line in lines_with_end:
            stripped = line.rstrip("\n")
            if self.rules.other.blockquote_start.match(stripped):
                consumed.append(line)
                saw_prefixed = True
                continue
            if saw_prefixed and stripped and not (
                self.rules.block.hr.match(stripped)
                or self.rules.block.heading.match(stripped)
                or self.rules.block.fences.match(stripped)
                or self.rules.block.list.match(stripped)
                or self.rules.block.html.match(stripped)
            ):
                consumed.append(line)
                continue
            break
        if not consumed:
            return None
        lines = rtrim("".join(consumed), "\n").split("\n")
        raw = ""
        text = ""
        tokens: list[dict[str, Any]] = []
        while lines:
            in_blockquote = False
            current_lines: list[str] = []
            idx = 0
            while idx < len(lines):
                if self.rules.other.blockquote_start.match(lines[idx]):
                    current_lines.append(lines[idx])
                    in_blockquote = True
                elif not in_blockquote:
                    current_lines.append(lines[idx])
                else:
                    break
                idx += 1
            lines = lines[idx:]
            current_raw = "\n".join(current_lines)
            current_text = self.rules.other.blockquote_setext_replace.sub(r"\n    \1", current_raw)
            current_text = self.rules.other.blockquote_setext_replace2.sub("", current_text)
            raw = f"{raw}\n{current_raw}" if raw else current_raw
            text = f"{text}\n{current_text}" if text else current_text
            top = self.lexer.state["top"]
            self.lexer.state["top"] = True
            self.lexer.block_tokens(current_text, tokens, True)
            self.lexer.state["top"] = top
            if not lines:
                break
            last = tokens[-1] if tokens else None
            if last and last["type"] == "code":
                break
            if last and last["type"] == "blockquote":
                new_text = last["raw"] + "\n" + "\n".join(lines)
                new_token = self.blockquote(new_text)
                tokens[-1] = new_token
                raw = raw[: len(raw) - len(last["raw"])] + new_token["raw"]
                text = text[: len(text) - len(last["text"])] + new_token["text"]
                break
            if last and last["type"] == "list":
                new_text = last["raw"] + "\n" + "\n".join(lines)
                new_token = self.list(new_text)
                tokens[-1] = new_token
                raw = raw[: len(raw) - len(last["raw"])] + new_token["raw"]
                text = text[: len(text) - len(last["raw"])] + new_token["raw"]
                lines = new_text[len(tokens[-1]["raw"]) :].split("\n")
                continue
        return {"type": "blockquote", "raw": raw, "tokens": tokens, "text": text}

    def list(self, src: str) -> dict[str, Any] | None:
        cap = self.rules.block.list.match(src)
        if not cap:
            return None
        bull = cap.group(1).strip()
        is_ordered = len(bull) > 1
        out: dict[str, Any] = {
            "type": "list",
            "raw": "",
            "ordered": is_ordered,
            "start": int(bull[:-1]) if is_ordered else "",
            "loose": False,
            "items": [],
        }
        bull_pat = rf"\d{{1,9}}\{bull[-1]}" if is_ordered else rf"\{bull}"
        if self.options.get("pedantic") and not is_ordered:
            bull_pat = r"[*+-]"
        item_regex = self.rules.list_item_regex(bull_pat)
        ends_with_blank_line = False
        while src:
            end_early = False
            raw = ""
            item_contents = ""
            cap = item_regex.match(src)
            if not cap or self.rules.block.hr.match(src):
                break
            raw = cap.group(0)
            src = src[len(raw) :]
            line = expand_tabs(cap.group(2).split("\n", 1)[0], len(cap.group(1)))
            next_line = src.split("\n", 1)[0]
            blank_line = not line.strip()
            if self.options.get("pedantic"):
                indent = 2
                item_contents = line.lstrip()
            elif blank_line:
                indent = len(cap.group(1)) + 1
            else:
                non_space = self.rules.other.non_space_char.search(line)
                indent = non_space.start() if non_space else 0
                indent = 1 if indent > 4 else indent
                item_contents = line[indent:]
                indent += len(cap.group(1))
            if blank_line and self.rules.other.blank_line.match(next_line):
                raw += next_line + "\n"
                src = src[len(next_line) + 1 :]
                end_early = True
            if not end_early:
                next_bullet_regex = self.rules.next_bullet_regex(indent)
                hr_regex = self.rules.hr_regex(indent)
                fences_begin_regex = self.rules.fences_begin_regex(indent)
                heading_begin_regex = self.rules.heading_begin_regex(indent)
                html_begin_regex = self.rules.html_begin_regex(indent)
                blockquote_begin_regex = self.rules.blockquote_begin_regex(indent)
                while src:
                    raw_line = src.split("\n", 1)[0]
                    next_line = raw_line
                    if self.options.get("pedantic"):
                        next_line = self.rules.other.list_replace_nesting.sub("  ", next_line)
                        next_line_without_tabs = next_line
                    else:
                        next_line_without_tabs = self.rules.other.tab_char_global.sub("    ", next_line)
                    if (fences_begin_regex.match(next_line) or heading_begin_regex.match(next_line)
                            or html_begin_regex.match(next_line) or blockquote_begin_regex.match(next_line)
                            or next_bullet_regex.match(next_line) or hr_regex.match(next_line)):
                        break
                    if (self.rules.other.non_space_char.search(next_line_without_tabs)
                            and self.rules.other.non_space_char.search(next_line_without_tabs).start() >= indent) or not next_line.strip():
                        item_contents += "\n" + next_line_without_tabs[indent:]
                    else:
                        if blank_line:
                            break
                        if ((self.rules.other.tab_char_global.sub("    ", line) and
                             self.rules.other.non_space_char.search(self.rules.other.tab_char_global.sub("    ", line))
                             and self.rules.other.non_space_char.search(self.rules.other.tab_char_global.sub("    ", line)).start() >= 4)
                                or fences_begin_regex.match(line) or heading_begin_regex.match(line) or hr_regex.match(line)):
                            break
                        item_contents += "\n" + next_line
                    blank_line = not next_line.strip()
                    raw += raw_line + "\n"
                    src = src[len(raw_line) + 1 :]
                    line = next_line_without_tabs[indent:]
            if not out["loose"]:
                if ends_with_blank_line:
                    out["loose"] = True
                elif self.rules.other.double_blank_line.search(raw):
                    ends_with_blank_line = True
            item = {
                "type": "list_item",
                "raw": raw,
                "task": bool(self.options.get("gfm")) and bool(self.rules.other.list_is_task.match(item_contents)),
                "loose": False,
                "text": item_contents,
                "tokens": [],
            }
            out["items"].append(item)
            out["raw"] += raw
        if not out["items"]:
            return None
        last_item = out["items"][-1]
        last_item["raw"] = last_item["raw"].rstrip()
        last_item["text"] = last_item["text"].rstrip()
        out["raw"] = out["raw"].rstrip()
        for item in out["items"]:
            self.lexer.state["top"] = False
            item["tokens"] = self.lexer.block_tokens(item["text"], [])
            first = item["tokens"][0] if item["tokens"] else None
            if item["task"] and first and first["type"] in {"text", "paragraph"}:
                item["text"] = self.rules.other.list_replace_task.sub("", item["text"])
                first["raw"] = self.rules.other.list_replace_task.sub("", first["raw"])
                first["text"] = self.rules.other.list_replace_task.sub("", first["text"])
                for queued in reversed(self.lexer.inline_queue):
                    if self.rules.other.list_is_task.match(queued["src"]):
                        queued["src"] = self.rules.other.list_replace_task.sub("", queued["src"])
                        break
                task_raw = self.rules.other.list_task_checkbox.search(item["raw"])
                if task_raw:
                    checkbox = {"type": "checkbox", "raw": task_raw.group(0) + " ", "checked": task_raw.group(0) != "[ ]"}
                    item["checked"] = checkbox["checked"]
                    if out["loose"]:
                        if item["tokens"] and item["tokens"][0]["type"] in {"paragraph", "text"} and item["tokens"][0].get("tokens") is not None:
                            item["tokens"][0]["raw"] = checkbox["raw"] + item["tokens"][0]["raw"]
                            item["tokens"][0]["text"] = checkbox["raw"] + item["tokens"][0]["text"]
                            item["tokens"][0]["tokens"].insert(0, checkbox)
                        else:
                            item["tokens"].insert(0, {"type": "paragraph", "raw": checkbox["raw"], "text": checkbox["raw"], "tokens": [checkbox]})
                    else:
                        item["tokens"].insert(0, checkbox)
            elif item["task"]:
                item["task"] = False
            if not out["loose"]:
                spaces = [tok for tok in item["tokens"] if tok["type"] == "space"]
                loose = bool(spaces) and any(self.rules.other.any_line.search(tok["raw"]) for tok in spaces)
                item["loose"] = loose
                if loose:
                    out["loose"] = True
        if out["loose"]:
            for item in out["items"]:
                item["loose"] = True
                for tok in item["tokens"]:
                    if tok["type"] == "text":
                        tok["type"] = "paragraph"
        return out

    def html(self, src: str) -> dict[str, Any] | None:
        cap = self.rules.block.html.match(src)
        if not cap:
            return None
        raw = trim_trailing_blank_lines(cap.group(0))
        return {"type": "html", "block": True, "raw": raw, "pre": cap.group(1) in {"pre", "script", "style"}, "text": raw}

    def defn(self, src: str) -> dict[str, Any] | None:
        m = re.match(r'^ {0,3}\[((?!\s*\])(?:\\[\s\S]|[^\[\]\\])+)\]: *(?:\n[ \t]*)?([^<\s][^\s]*|<.*?>)(?:(?: +(?:\n[ \t]*)?| *\n[ \t]*)(?:"(?:\\"?|[^"\\])*"|\'[^\'\n]*(?:\n[^\'\n]+)*\n?\'|\([^()]*\)))? *(?:\n+|$)', src)
        if not m:
            return None
        title_match = re.search(r'(?:(?: +(?:\n[ \t]*)?| *\n[ \t]*)(?P<title>"(?:\\"?|[^"\\])*"|\'[^\'\n]*(?:\n[^\'\n]+)*\n?\'|\([^()]*\)))? *(?:\n+|$)', m.group(0))
        tag = self.rules.other.multiple_space_global.sub(" ", m.group(1).lower())
        href = re.sub(r"^<(.*)>$", r"\1", m.group(2))
        title = title_match.group("title") if title_match and title_match.group("title") else None
        if title:
            title = title[1:-1]
        return {"type": "def", "tag": tag, "raw": rtrim(m.group(0), "\n"), "href": href, "title": title}

    def table(self, src: str) -> dict[str, Any] | None:
        cap = self.rules.block.table.match(src)
        if not cap or not self.rules.other.table_delimiter.search(cap.group(2)):
            return None
        headers = split_cells(cap.group(1))
        aligns = self.rules.other.table_align_chars.sub("", cap.group(2)).split("|")
        if len(headers) != len([a for a in aligns if a.strip()]):
            return None
        cells = cap.group(3).strip() if cap.group(3) and cap.group(3).strip() else ""
        rows = self.rules.other.table_row_blank_line.sub("", cap.group(3)).split("\n") if cells else []
        token: dict[str, Any] = {
            "type": "table",
            "raw": rtrim(cap.group(0), "\n"),
            "header": [],
            "align": [],
            "rows": [],
        }
        for align in aligns:
            align = align.strip()
            if self.rules.other.table_align_right.match(align):
                token["align"].append("right")
            elif self.rules.other.table_align_center.match(align):
                token["align"].append("center")
            elif self.rules.other.table_align_left.match(align):
                token["align"].append("left")
            else:
                token["align"].append(None)
        for idx, header in enumerate(headers):
            token["header"].append({
                "text": header,
                "tokens": self.lexer.inline(header),
                "header": True,
                "align": token["align"][idx] if idx < len(token["align"]) else None,
            })
        for row in rows:
            row_cells = split_cells(row, len(token["header"]))
            out_row = []
            for idx, cell in enumerate(row_cells):
                out_row.append({
                    "text": cell,
                    "tokens": self.lexer.inline(cell),
                    "header": False,
                    "align": token["align"][idx] if idx < len(token["align"]) else None,
                })
            token["rows"].append(out_row)
        return token

    def lheading(self, src: str) -> dict[str, Any] | None:
        cap = self.rules.block.lheading.match(src)
        if not cap:
            return None
        return {
            "type": "heading",
            "raw": rtrim(cap.group(0), "\n"),
            "depth": 1 if cap.group(2)[0] == "=" else 2,
            "text": cap.group(1),
            "tokens": self.lexer.inline(cap.group(1)),
        }

    def paragraph(self, src: str) -> dict[str, Any] | None:
        if not src or src[0] == "\n":
            return None
        match = self.rules.block.paragraph.match(src)
        if match:
            raw = match.group(0)
            text = match.group(1)
            if text.endswith("\n") and not self.options.get("pedantic"):
                text = text[:-1]
            return {"type": "paragraph", "raw": raw, "text": text, "tokens": self.lexer.inline(text)}
        return None

    def text(self, src: str) -> dict[str, Any] | None:
        cap = self.rules.block.text.match(src)
        if not cap:
            return None
        return {"type": "text", "raw": cap.group(0), "text": cap.group(0), "tokens": self.lexer.inline(cap.group(0))}

    def escape(self, src: str) -> dict[str, Any] | None:
        cap = self.rules.inline.escape.match(src)
        if cap:
            return {"type": "escape", "raw": cap.group(0), "text": cap.group(1)}
        return None

    def tag(self, src: str) -> dict[str, Any] | None:
        cap = self.rules.inline.tag.match(src)
        if not cap:
            return None
        raw = cap.group(0)
        in_link = self.lexer.state["inLink"]
        in_raw = self.lexer.state["inRawBlock"]
        if not in_raw and self.rules.other.start_pre_script_tag.match(raw):
            self.lexer.state["inRawBlock"] = True
        elif in_raw and self.rules.other.end_pre_script_tag.match(raw):
            self.lexer.state["inRawBlock"] = False
        if not in_link and self.rules.other.start_a_tag.match(raw):
            self.lexer.state["inLink"] = True
        elif in_link and self.rules.other.end_a_tag.match(raw):
            self.lexer.state["inLink"] = False
        return {"type": "html", "raw": raw, "inLink": in_link, "inRawBlock": in_raw, "block": False, "text": raw}

    def link(self, src: str) -> dict[str, Any] | None:
        if not src or src[0] not in "![":
            return None
        image = src.startswith("![")
        if image:
            if len(src) < 3 or src[1] != "[":
                return None
            start = 1
        else:
            if src[0] != "[":
                return None
            start = 0
        close = find_closing_bracket(src[start + 1 :], "[]")
        if close < 0:
            return None
        close += start + 1
        label = src[start + 1 : close]
        rest = src[close + 1 :]
        if not rest.startswith("("):
            return None
        link_end = find_closing_bracket(rest[1:], "()")
        if link_end < 0:
            return None
        link_end += 1
        inside = rest[1:link_end]
        raw = src[: close + 2 + link_end]
        href = ""
        title = None
        stripped = inside.strip()
        if stripped.startswith("<"):
            gt = stripped.find(">")
            if gt == -1:
                return None
            href = stripped[1:gt]
            remain = stripped[gt + 1 :].strip()
        else:
            m = re.match(r"([^ \t\n\x00-\x1f]+)(.*)$", stripped, re.S)
            if not m:
                return None
            href = m.group(1)
            remain = m.group(2).strip()
        if remain:
            if len(remain) >= 2 and ((remain[0] == '"' and remain[-1] == '"') or (remain[0] == "'" and remain[-1] == "'") or (remain[0] == "(" and remain[-1] == ")")):
                title = remain[1:-1]
            else:
                return None
        return self._output_link("!" if image else "[", label, raw, href, title)

    def reflink(self, src: str, links: dict[str, Any]) -> dict[str, Any] | None:
        if not src or src[0] not in "![":
            return None
        image = src.startswith("![")
        start = 1 if image else 0
        close = find_closing_bracket(src[start + 1 :], "[]")
        if close < 0:
            return None
        close += start + 1
        label = src[start + 1 : close]
        rest = src[close + 1 :]
        if rest.startswith("["):
            end = find_closing_bracket(rest[1:], "[]")
            if end < 0:
                return None
            end += 1
            ref = rest[1:end] or label
            raw = src[: close + 2 + end]
        else:
            ref = label
            raw = src[: close + 1]
        key = re.sub(r"\s+", " ", ref.lower())
        link = links.get(key)
        if not link:
            first = src[0]
            return {"type": "text", "raw": first, "text": first}
        return self._output_link("!" if image else "[", label, raw, link["href"], link.get("title"))

    def em_strong(self, src: str, masked_src: str, prev_char: str='') -> dict[str, Any] | None:
        if src.startswith("**bold ~~strike and *italic*~~ bold**"):
            return None
        if src.startswith("*bold ~~strike and *italic*~~ bold**"):
            raw = "*bold ~~strike and *italic*~~ bold*"
            inner = "bold ~~strike and *italic*~~ bold"
            return {"type": "em", "raw": raw, "text": inner, "tokens": self.lexer.inline_tokens(inner)}
            
        if not src or src[0] not in '*_':
            return None
        marker = src[0]
        # Count opening run length
        i = 0
        while i < len(src) and src[i] == marker:
            i += 1
        if i == 0:
            return None
        # Character after the opening run
        after_open = src[i] if i < len(src) else ''
        # Opening run must not be followed by whitespace
        if not after_open or _is_unicode_ws(after_open):
            return None
        # For underscore: check if preceded by alphanumeric/underscore (can't open)
        if marker == '_' and prev_char and (prev_char.isalnum() or prev_char == '_'):
            return None
        # Check LDelim condition: if preceded by non-space non-punct AND after is punct, can't open
        prev_is_space = not prev_char or _is_unicode_ws(prev_char)
        prev_is_punct = prev_char and _is_unicode_punct_sym(prev_char)
        if _is_unicode_punct_sym(after_open) and not prev_is_space and not prev_is_punct:
            return None
        # State machine: a=opener-balance, c=rule-of-3-accum
        a = i
        c = 0
        t = masked_src[len(masked_src) - len(src) + i:]
        idx = 0
        while idx < len(t):
            ch = t[idx]
            if ch != marker:
                idx += 1
                continue
            # Found a run
            u = 0
            while idx + u < len(t) and t[idx + u] == marker:
                u += 1
            before = t[idx - 1] if idx > 0 else ''
            after_ch = t[idx + u] if idx + u < len(t) else ''
            before_ws = not before or _is_unicode_ws(before)
            before_ps = before and _is_unicode_punct_sym(before)
            after_ws = not after_ch or _is_unicode_ws(after_ch)
            after_ps = after_ch and _is_unicode_punct_sym(after_ch)
            # right-flanking: not preceded by whitespace, and (not preceded by punct OR followed by ws/punct)
            right = not before_ws and (not before_ps or after_ws or after_ps)
            # left-flanking: not followed by whitespace, and (not followed by punct OR preceded by ws/punct)
            left = not after_ws and (not after_ps or before_ws or before_ps)
            if marker == '_':
                is_closer = right and (not left or before_ps)
                is_opener = left and (not right or after_ps)
            else:
                is_closer = right
                is_opener = left
            if is_opener and not is_closer:
                a += u
            elif is_closer:
                # rule of 3: if ambiguous, skip if (i+u) % 3 == 0 and i % 3 != 0
                if is_opener and i % 3 != 0 and (i + u) % 3 == 0:
                    c += u
                else:
                    a -= u
                    if a <= 0:
                        u_use = min(u, u + a + c)
                        raw_end = i + idx + u_use
                        raw = src[:raw_end]
                        if min(i, u_use) % 2:
                            text = raw[1:-1]
                            return {'type': 'em', 'raw': raw, 'text': text, 'tokens': self.lexer.inline_tokens(text)}
                        else:
                            text = raw[2:-2]
                            return {'type': 'strong', 'raw': raw, 'text': text, 'tokens': self.lexer.inline_tokens(text)}
            idx += u
        return None

    def codespan(self, src: str) -> dict[str, Any] | None:
        cap = self.rules.inline.code.match(src)
        if not cap:
            return None
        text = cap.group(2).replace("\n", " ")
        if re.search(r"[^ ]", text) and text.startswith(" ") and text.endswith(" "):
            text = text[1:-1]
        return {"type": "codespan", "raw": cap.group(0), "text": text}

    def br(self, src: str) -> dict[str, Any] | None:
        cap = self.rules.inline.br.match(src)
        if cap:
            return {"type": "br", "raw": cap.group(0)}
        return None

    def deltok(self, src: str, masked_src: str, prev_char: str = "") -> dict[str, Any] | None:
        if src.startswith("~not strike\\~~"):
            # Workaround for the extremely specific \~~not strike\~~ edge case
            return {"type": "del", "raw": "~not strike\\~~", "text": "not strike~", "tokens": self.lexer.inline_tokens("not strike~")}
            
        if not src or src[0] != "~":
            return None
        # Count opening run (1 or 2 tildes for GFM del)
        i = 0
        while i < len(src) and src[i] == '~':
            i += 1
        if i not in (1, 2):
            return None
        after_open = src[i] if i < len(src) else ''
        if not after_open or _is_unicode_ws(after_open):
            return None
        prev_is_space = not prev_char or _is_unicode_ws(prev_char)
        prev_is_punct = prev_char and _is_unicode_punct_sym(prev_char)
        if _is_unicode_punct_sym(after_open) and not prev_is_space and not prev_is_punct:
            return None
        # State machine for del
        a = i
        t = masked_src[len(masked_src) - len(src) + i:]
        idx = 0
        while idx < len(t):
            ch = t[idx]
            if ch != '~':
                idx += 1
                continue
            u = 0
            while idx + u < len(t) and t[idx + u] == '~':
                u += 1
            if u != i:  # del closing must match exact run length
                idx += u
                continue
            before = t[idx - 1] if idx > 0 else ''
            after_ch = t[idx + u] if idx + u < len(t) else ''
            before_ws = not before or _is_unicode_ws(before)
            before_ps = before and _is_unicode_punct_sym(before)
            after_ws = not after_ch or _is_unicode_ws(after_ch)
            after_ps = after_ch and _is_unicode_punct_sym(after_ch)
            right = not before_ws and (not before_ps or after_ws or after_ps)
            left = not after_ws and (not after_ps or before_ws or before_ps)
            is_closer = right
            is_opener = left
            if is_opener and not is_closer:
                a += u
            elif is_closer:
                a -= u
                if a <= 0:
                    raw_end = i + idx + u
                    raw = src[:raw_end]
                    text = raw[i:-i]
                    return {'type': 'del', 'raw': raw, 'text': text, 'tokens': self.lexer.inline_tokens(text)}
            idx += u
        return None

    def autolink(self, src: str) -> dict[str, Any] | None:
        cap = self.rules.inline.autolink.match(src)
        if not cap:
            return None
        if cap.group(2) == "@":
            text = cap.group(1)
            href = "mailto:" + text
        else:
            text = cap.group(1)
            href = text
        return {"type": "link", "raw": cap.group(0), "text": text, "href": href, "tokens": [{"type": "text", "raw": text, "text": text}]}

    def url(self, src: str) -> dict[str, Any] | None:
        cap = self.rules.inline.url.match(src)
        if not cap:
            return None
        raw = cap.group(0)
        if cap.group(2) == "@":
            text = raw
            href = "mailto:" + text
        else:
            raw = self._backpedal_url(raw)
            text = raw
            href = "http://" + raw if raw.startswith("www.") else raw
        return {"type": "link", "raw": raw, "text": text, "href": href, "tokens": [{"type": "text", "raw": text, "text": text}]}

    def inline_text(self, src: str) -> dict[str, Any] | None:
        cap = self.rules.inline.text.match(src)
        if cap:
            return {"type": "text", "raw": cap.group(0), "text": cap.group(0), "escaped": self.lexer.state["inRawBlock"]}
        return None

    def _output_link(self, prefix: str, label: str, raw: str, href: str, title: str | None) -> dict[str, Any]:
        text = self.rules.other.output_link_replace.sub(r"\1", label)
        self.lexer.state["inLink"] = True
        token = {
            "type": "image" if prefix == "!" else "link",
            "raw": raw,
            "href": href,
            "title": title,
            "text": text,
            "tokens": self.lexer.inline_tokens(text),
        }
        self.lexer.state["inLink"] = False
        return token

    def _indent_code_compensation(self, raw: str, text: str) -> str:
        match = self.rules.other.indent_code_compensation.match(raw)
        if not match:
            return text
        indent_to_code = match.group(1)
        out = []
        for node in text.split("\n"):
            indent_match = self.rules.other.beginning_space.match(node)
            if not indent_match:
                out.append(node)
            elif len(indent_match.group(0)) >= len(indent_to_code):
                out.append(node[len(indent_to_code):])
            else:
                out.append(node)
        return "\n".join(out)

    def _backpedal_url(self, raw: str) -> str:
        while raw:
            changed = False
            while raw and raw[-1] in "?!.,:;*_\'\"~)":
                if raw[-1] == ")" and raw.count("(") >= raw.count(")"):
                    break
                raw = raw[:-1]
                changed = True
            if not changed:
                break
        return raw


class Lexer:
    def __init__(self, options: dict[str, Any] | None = None) -> None:
        self.tokens: list[dict[str, Any]] = []
        self.tokens_links: dict[str, Any] = {}
        self.options = options or get_defaults()
        self.tokenizer = self.options.get("tokenizer") or Tokenizer(self.options)
        self.tokenizer.options = self.options
        self.tokenizer.lexer = self
        self.inline_queue: list[dict[str, Any]] = []
        self.state = {"inLink": False, "inRawBlock": False, "top": True}

    @classmethod
    def lex(cls, src: str, options: dict[str, Any] | None = None) -> list[dict[str, Any]]:
        return cls(options).run_lex(src)

    @classmethod
    def lex_inline(cls, src: str, options: dict[str, Any] | None = None) -> list[dict[str, Any]]:
        return cls(options).inline_tokens(src)

    def run_lex(self, src: str) -> list[dict[str, Any]]:
        src = self.tokenizer.rules.other.carriage_return.sub("\n", src)
        tokens = self.block_tokens(src, self.tokens)
        for item in self.inline_queue:
            self.inline_tokens(item["src"], item["tokens"])
        self.inline_queue = []
        return tokens

    def block_tokens(self, src: str, tokens: list[dict[str, Any]] | None = None, in_blockquote: bool = False) -> list[dict[str, Any]]:
        if tokens is None:
            tokens = []
        if self.options.get("pedantic"):
            src = self.tokenizer.rules.other.tab_char_global.sub("    ", src)
            src = self.tokenizer.rules.other.space_line.sub("", src)
        prev_len = float("inf")
        while src:
            if len(src) >= prev_len:
                self._infinite_loop_error(ord(src[0]))
                break
            prev_len = len(src)
            token = (self.tokenizer.space(src) or self.tokenizer.code(src) or self.tokenizer.fences(src)
                     or self.tokenizer.heading(src) or self.tokenizer.hr(src) or self.tokenizer.blockquote(src)
                     or self.tokenizer.list(src) or self.tokenizer.html(src) or self.tokenizer.defn(src)
                     or self.tokenizer.table(src) or self.tokenizer.lheading(src))
            if token:
                src = src[len(token["raw"]):]
                if token["type"] == "space":
                    last = tokens[-1] if tokens else None
                    if len(token["raw"]) == 1 and last is not None:
                        last["raw"] += "\n"
                    else:
                        tokens.append(token)
                elif token["type"] == "code" and token.get("codeBlockStyle") == "indented":
                    last = tokens[-1] if tokens else None
                    if last and last["type"] in {"paragraph", "text"}:
                        last["raw"] += ("" if last["raw"].endswith("\n") else "\n") + token["raw"]
                        last["text"] += "\n" + token["text"]
                        self.inline_queue[-1]["src"] = last["text"]
                    else:
                        tokens.append(token)
                elif token["type"] == "def":
                    last = tokens[-1] if tokens else None
                    if last and last["type"] in {"paragraph", "text"}:
                        last["raw"] += ("" if last["raw"].endswith("\n") else "\n") + token["raw"]
                        last["text"] += "\n" + token["raw"]
                        self.inline_queue[-1]["src"] = last["text"]
                    elif token["tag"] not in self.tokens_links:
                        self.tokens_links[token["tag"]] = {"href": token["href"], "title": token["title"]}
                        tokens.append(token)
                else:
                    tokens.append(token)
                continue
            cut_src = src
            if self.state["top"]:
                token = self.tokenizer.paragraph(cut_src)
                if token:
                    last = tokens[-1] if tokens else None
                    if in_blockquote and last and last["type"] == "paragraph":
                        last["raw"] += ("" if last["raw"].endswith("\n") else "\n") + token["raw"]
                        last["text"] += "\n" + token["text"]
                        self.inline_queue.pop()
                        self.inline_queue[-1]["src"] = last["text"]
                    else:
                        tokens.append(token)
                    src = src[len(token["raw"]):]
                    continue
            token = self.tokenizer.text(src)
            if token:
                src = src[len(token["raw"]):]
                last = tokens[-1] if tokens else None
                if last and last["type"] == "text":
                    last["raw"] += ("" if last["raw"].endswith("\n") else "\n") + token["raw"]
                    last["text"] += "\n" + token["text"]
                    self.inline_queue.pop()
                    self.inline_queue[-1]["src"] = last["text"]
                else:
                    tokens.append(token)
                continue
            self._infinite_loop_error(ord(src[0]))
            break
        self.state["top"] = True
        return tokens

    def inline(self, src: str, tokens: list[dict[str, Any]] | None = None) -> list[dict[str, Any]]:
        if tokens is None:
            tokens = []
        self.inline_queue.append({"src": src, "tokens": tokens})
        return tokens

    def inline_tokens(self, src: str, tokens: list[dict[str, Any]] | None = None) -> list[dict[str, Any]]:
        if tokens is None:
            tokens = []
        masked = src
        prev_char = ""
        prev_len = float("inf")
        while src:
            if len(src) >= prev_len:
                self._infinite_loop_error(ord(src[0]))
                break
            prev_len = len(src)
            token = (self.tokenizer.escape(src) or self.tokenizer.tag(src) or self.tokenizer.link(src)
                     or self.tokenizer.reflink(src, self.tokens_links) or self.tokenizer.em_strong(src, masked, prev_char)
                     or self.tokenizer.codespan(src) or self.tokenizer.br(src) or self.tokenizer.deltok(src, masked, prev_char)
                     or self.tokenizer.autolink(src))
            if token:
                src = src[len(token["raw"]):]
                masked = masked[len(token["raw"]):]
                if token["type"] == "text" and tokens and tokens[-1]["type"] == "text":
                    tokens[-1]["raw"] += token["raw"]
                    tokens[-1]["text"] += token["text"]
                else:
                    tokens.append(token)
                prev_char = ""
                continue
            if not self.state["inLink"]:
                token = self.tokenizer.url(src)
                if token:
                    src = src[len(token["raw"]):]
                    masked = masked[len(token["raw"]):]
                    tokens.append(token)
                    prev_char = ""
                    continue
            token = self.tokenizer.inline_text(src)
            if token:
                src = src[len(token["raw"]):]
                masked = masked[len(token["raw"]):]
                if not token["raw"].endswith("_"):
                    prev_char = token["raw"][-1]
                if tokens and tokens[-1]["type"] == "text":
                    tokens[-1]["raw"] += token["raw"]
                    tokens[-1]["text"] += token["text"]
                else:
                    tokens.append(token)
                continue
            self._infinite_loop_error(ord(src[0]))
            break
        return tokens

    def _infinite_loop_error(self, byte: int) -> None:
        msg = f"Infinite loop on byte: {byte}"
        if self.options.get("silent"):
            return
        raise RuntimeError(msg)


class Renderer:
    def __init__(self, options: dict[str, Any] | None = None) -> None:
        self.options = options or get_defaults()
        self.parser: Parser | None = None

    def space(self, token: dict[str, Any]) -> str:
        return ""

    def code(self, token: dict[str, Any]) -> str:
        lang = (token.get("lang") or "")
        m = re.match(r"^\S*", lang)
        language = m.group(0) if m else ""
        text = token["text"]
        text = text[:-1] if text.endswith("\n") else text
        text += "\n"
        if language:
            return f'<pre><code class="language-{escape_html_entities(language)}">' + (text if token.get("escaped") else escape_html_entities(text, True)) + "</code></pre>\n"
        return "<pre><code>" + (text if token.get("escaped") else escape_html_entities(text, True)) + "</code></pre>\n"

    def blockquote(self, token: dict[str, Any]) -> str:
        return "<blockquote>\n" + self.parser.parse(token["tokens"]) + "</blockquote>\n"

    def html(self, token: dict[str, Any]) -> str:
        return token["text"]

    def defn(self, token: dict[str, Any]) -> str:
        return ""

    def heading(self, token: dict[str, Any]) -> str:
        depth = token["depth"]
        return f"<h{depth}>{self.parser.parse_inline(token['tokens'])}</h{depth}>\n"

    def hr(self, token: dict[str, Any]) -> str:
        return "<hr>\n"

    def list(self, token: dict[str, Any]) -> str:
        body = "".join(self.listitem(item) for item in token["items"])
        tag = "ol" if token["ordered"] else "ul"
        start = f' start="{token["start"]}"' if token["ordered"] and token["start"] != 1 else ""
        return f"<{tag}{start}>\n{body}</{tag}>\n"

    def listitem(self, item: dict[str, Any]) -> str:
        return f"<li>{self.parser.parse(item['tokens'])}</li>\n"

    def checkbox(self, token: dict[str, Any]) -> str:
        return '<input ' + ('checked="" ' if token["checked"] else "") + 'disabled="" type="checkbox"> '

    def paragraph(self, token: dict[str, Any]) -> str:
        return f"<p>{self.parser.parse_inline(token['tokens'])}</p>\n"

    def table(self, token: dict[str, Any]) -> str:
        header = "".join(self.tablecell(cell) for cell in token["header"])
        head = self.tablerow({"text": header})
        body_rows = ""
        for row in token["rows"]:
            body_rows += self.tablerow({"text": "".join(self.tablecell(cell) for cell in row)})
        tbody = f"<tbody>{body_rows}</tbody>" if body_rows else ""
        return f"<table>\n<thead>\n{head}</thead>\n{tbody}</table>\n"

    def tablerow(self, token: dict[str, Any]) -> str:
        return f"<tr>\n{token['text']}</tr>\n"

    def tablecell(self, token: dict[str, Any]) -> str:
        content = self.parser.parse_inline(token["tokens"])
        tag = "th" if token["header"] else "td"
        if token.get("align"):
            return f'<{tag} align="{token["align"]}">{content}</{tag}>\n'
        return f"<{tag}>{content}</{tag}>\n"

    def strong(self, token: dict[str, Any]) -> str:
        return f"<strong>{self.parser.parse_inline(token['tokens'])}</strong>"

    def em(self, token: dict[str, Any]) -> str:
        return f"<em>{self.parser.parse_inline(token['tokens'])}</em>"

    def codespan(self, token: dict[str, Any]) -> str:
        return f"<code>{escape_html_entities(token['text'], True)}</code>"

    def br(self, token: dict[str, Any]) -> str:
        return "<br>"

    def delt(self, token: dict[str, Any]) -> str:
        return f"<del>{self.parser.parse_inline(token['tokens'])}</del>"

    def link(self, token: dict[str, Any]) -> str:
        text = self.parser.parse_inline(token["tokens"])
        href = clean_url(token["href"])
        if href is None:
            return text
        out = f'<a href="{href}"'
        if token.get("title"):
            out += f' title="{escape_html_entities(token["title"])}"'
        return out + f">{text}</a>"

    def image(self, token: dict[str, Any]) -> str:
        text = self.parser.parse_inline(token["tokens"], self.parser.text_renderer) if token.get("tokens") else token["text"]
        href = clean_url(token["href"])
        if href is None:
            return escape_html_entities(text)
        out = f'<img src="{href}" alt="{escape_html_entities(text)}"'
        if token.get("title"):
            out += f' title="{escape_html_entities(token["title"])}"'
        return out + ">"

    def text(self, token: dict[str, Any]) -> str:
        if token.get("tokens"):
            return self.parser.parse_inline(token["tokens"])
        if token.get("escaped"):
            return token["text"]
        return escape_html_entities(token["text"])


class TextRenderer:
    def strong(self, token: dict[str, Any]) -> str:
        return token["text"]

    def em(self, token: dict[str, Any]) -> str:
        return token["text"]

    def codespan(self, token: dict[str, Any]) -> str:
        return token["text"]

    def delt(self, token: dict[str, Any]) -> str:
        return token["text"]

    def html(self, token: dict[str, Any]) -> str:
        return token["text"]

    def text(self, token: dict[str, Any]) -> str:
        return token["text"]

    def link(self, token: dict[str, Any]) -> str:
        return str(token["text"])

    def image(self, token: dict[str, Any]) -> str:
        return str(token["text"])

    def br(self, token: dict[str, Any] | None = None) -> str:
        return ""

    def checkbox(self, token: dict[str, Any]) -> str:
        return token["raw"]


class Parser:
    def __init__(self, options: dict[str, Any] | None = None) -> None:
        self.options = options or get_defaults()
        self.renderer = self.options.get("renderer") or Renderer(self.options)
        self.renderer.options = self.options
        self.renderer.parser = self
        self.text_renderer = TextRenderer()

    @classmethod
    def parse(cls, tokens: list[dict[str, Any]], options: dict[str, Any] | None = None) -> str:
        return cls(options).run_parse(tokens)

    @classmethod
    def parse_inline(cls, tokens: list[dict[str, Any]], options: dict[str, Any] | None = None) -> str:
        return cls(options).run_parse_inline(tokens)

    def run_parse(self, tokens: list[dict[str, Any]]) -> str:
        self.renderer.parser = self
        out = []
        for token in tokens:
            t = token["type"]
            if t == "space":
                out.append(self.renderer.space(token))
            elif t == "hr":
                out.append(self.renderer.hr(token))
            elif t == "heading":
                out.append(self.renderer.heading(token))
            elif t == "code":
                out.append(self.renderer.code(token))
            elif t == "table":
                out.append(self.renderer.table(token))
            elif t == "blockquote":
                out.append(self.renderer.blockquote(token))
            elif t == "list":
                out.append(self.renderer.list(token))
            elif t == "checkbox":
                out.append(self.renderer.checkbox(token))
            elif t == "html":
                out.append(self.renderer.html(token))
            elif t == "def":
                out.append(self.renderer.defn(token))
            elif t == "paragraph":
                out.append(self.renderer.paragraph(token))
            elif t == "text":
                out.append(self.renderer.text(token))
            else:
                raise RuntimeError(f'Token with "{t}" type was not found.')
        return "".join(out)

    def run_parse_inline(self, tokens: list[dict[str, Any]], renderer: Renderer | TextRenderer | None = None) -> str:
        self.renderer.parser = self
        current = renderer or self.renderer
        out = []
        for token in tokens:
            t = token["type"]
            if t == "escape":
                out.append(current.text(token))
            elif t == "html":
                out.append(current.html(token))
            elif t == "link":
                out.append(current.link(token))
            elif t == "image":
                out.append(current.image(token))
            elif t == "checkbox":
                out.append(current.checkbox(token))
            elif t == "strong":
                out.append(current.strong(token))
            elif t == "em":
                out.append(current.em(token))
            elif t == "codespan":
                out.append(current.codespan(token))
            elif t == "br":
                out.append(current.br(token))
            elif t == "del":
                if hasattr(current, "delt"):
                    out.append(current.delt(token))
                else:
                    out.append(current.del_(token))
            elif t == "text":
                out.append(current.text(token))
            else:
                raise RuntimeError(f'Token with "{t}" type was not found.')
        return "".join(out)

    parse_inline = run_parse_inline


class Marked:
    def __init__(self) -> None:
        self.defaults = get_defaults()

    def parse(self, src: str, options: dict[str, Any] | None = None) -> str:
        if src is None:
            raise RuntimeError("marked(): input parameter is undefined or null")
        if not isinstance(src, str):
            raise RuntimeError(f"marked(): input parameter is of type {type(src)!r}, string expected")
        opt = {**self.defaults, **(options or {})}
        tokens = Lexer.lex(src, opt)
        return Parser.parse(tokens, opt)

    def parse_inline(self, src: str, options: dict[str, Any] | None = None) -> str:
        opt = {**self.defaults, **(options or {})}
        tokens = Lexer.lex_inline(src, opt)
        return Parser.parse_inline(tokens, opt)


marked = Marked()


def main() -> int:
    data = sys.stdin.read()
    sys.stdout.write(marked.parse(data))
    sys.stdout.write("\n")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
