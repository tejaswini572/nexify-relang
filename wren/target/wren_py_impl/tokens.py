from enum import Enum, auto
from dataclasses import dataclass
from typing import Any

class TokenType(Enum):
    # Single-character punctuation.
    LPAREN = auto()
    RPAREN = auto()
    LBRACKET = auto()
    RBRACKET = auto()
    LBRACE = auto()
    RBRACE = auto()
    COLON = auto()
    QUESTION = auto()
    DOT = auto()
    COMMA = auto()
    STAR = auto()
    SLASH = auto()
    PERCENT = auto()
    PLUS = auto()
    MINUS = auto()
    TILDE = auto()
    CARET = auto()
    PIPE = auto()
    AMP = auto()
    BANG = auto()
    EQ = auto()
    LT = auto()
    GT = auto()
    NEWLINE = auto()
    
    # Two or more characters.
    DOTDOT = auto()
    DOTDOTDOT = auto()
    BANGEQ = auto()
    EQEQ = auto()
    GTEQ = auto()
    LTEQ = auto()
    PIPEPIPE = auto()
    AMPAMP = auto()
    
    # Literals
    NAME = auto()
    NUMBER = auto()
    STRING = auto()
    INTERPOLATION = auto()
    
    # Keywords
    AS = auto()
    BREAK = auto()
    CLASS = auto()
    CONSTRUCT = auto()
    CONTINUE = auto()
    ELSE = auto()
    FALSE = auto()
    FOR = auto()
    FOREIGN = auto()
    IF = auto()
    IMPORT = auto()
    IN = auto()
    IS = auto()
    NULL = auto()
    RETURN = auto()
    STATIC = auto()
    SUPER = auto()
    THIS = auto()
    TRUE = auto()
    VAR = auto()
    WHILE = auto()
    
    EOF = auto()


@dataclass
class Token:
    type: TokenType
    lexeme: str
    line: int
    value: Any = None

    def __repr__(self) -> str:
        return f"{self.type.name} '{self.lexeme}' {self.value if self.value is not None else ''} @ {self.line}"
