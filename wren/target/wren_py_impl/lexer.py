from tokens import Token, TokenType

KEYWORDS = {
    "as": TokenType.AS,
    "break": TokenType.BREAK,
    "class": TokenType.CLASS,
    "construct": TokenType.CONSTRUCT,
    "continue": TokenType.CONTINUE,
    "else": TokenType.ELSE,
    "false": TokenType.FALSE,
    "for": TokenType.FOR,
    "foreign": TokenType.FOREIGN,
    "if": TokenType.IF,
    "import": TokenType.IMPORT,
    "in": TokenType.IN,
    "is": TokenType.IS,
    "null": TokenType.NULL,
    "return": TokenType.RETURN,
    "static": TokenType.STATIC,
    "super": TokenType.SUPER,
    "this": TokenType.THIS,
    "true": TokenType.TRUE,
    "var": TokenType.VAR,
    "while": TokenType.WHILE,
}

class Lexer:
    """
    Lexical scanner taking raw source strings and yielding Token objects.
    Uses explicit stack for robust nested interpolation handling.
    """
    def __init__(self, source: str):
        self.source = source
        self.start = 0
        self.current = 0
        self.line = 1
        self.tokens: list[Token] = []
        
        # Explicit stack to track paren nesting for string interpolation
        self.interpolation_depths: list[int] = []
        self.current_parens = 0

    def is_at_end(self) -> bool:
        return self.current >= len(self.source)

    def advance(self) -> str:
        c = self.source[self.current]
        self.current += 1
        return c

    def match(self, expected: str) -> bool:
        if self.is_at_end(): return False
        if self.source[self.current] != expected: return False
        self.current += 1
        return True
        
    def peek(self) -> str:
        if self.is_at_end(): return '\0'
        return self.source[self.current]

    def peek_next(self) -> str:
        if self.current + 1 >= len(self.source): return '\0'
        return self.source[self.current + 1]

    def add_token(self, type: TokenType, value=None):
        text = self.source[self.start:self.current]
        self.tokens.append(Token(type, text, self.line, value))

    def scan_tokens(self) -> list[Token]:
        while not self.is_at_end():
            self.start = self.current
            self.scan_token()
            
        self.tokens.append(Token(TokenType.EOF, "", self.line))
        return self.tokens

    def scan_token(self):
        c = self.advance()
        
        # Fast exit skipping
        if c in ' \r\t':
            return
        
        # Structural 
        if c == '\n':
            self.add_token(TokenType.NEWLINE)
            self.line += 1
        elif c == '(':
            self.current_parens += 1
            self.add_token(TokenType.LPAREN)
        elif c == ')':
            # This is the crux of the nested interpolation logic:
            # If we're closing a paren expression strictly tied to an active interpolation expression...
            if self.interpolation_depths and self.current_parens == self.interpolation_depths[-1]:
                self.interpolation_depths.pop()
                self.consume_string() # resume string reading instantly!
            else:
                self.current_parens -= 1
                self.add_token(TokenType.RPAREN)
                
        elif c == '{': self.add_token(TokenType.LBRACE)
        elif c == '}': self.add_token(TokenType.RBRACE)
        elif c == '[': self.add_token(TokenType.LBRACKET)
        elif c == ']': self.add_token(TokenType.RBRACKET)
        elif c == ':': self.add_token(TokenType.COLON)
        elif c == '?': self.add_token(TokenType.QUESTION)
        elif c == ',': self.add_token(TokenType.COMMA)
        elif c == '*': self.add_token(TokenType.STAR)
        elif c == '%': self.add_token(TokenType.PERCENT)
        elif c == '+': self.add_token(TokenType.PLUS)
        elif c == '-': self.add_token(TokenType.MINUS)
        elif c == '~': self.add_token(TokenType.TILDE)
        elif c == '^': self.add_token(TokenType.CARET)
        elif c == '|': self.add_token(TokenType.PIPEPIPE if self.match('|') else TokenType.PIPE)
        elif c == '&': self.add_token(TokenType.AMPAMP if self.match('&') else TokenType.AMP)
        elif c == '!': self.add_token(TokenType.BANGEQ if self.match('=') else TokenType.BANG)
        elif c == '=': self.add_token(TokenType.EQEQ if self.match('=') else TokenType.EQ)
        elif c == '<': self.add_token(TokenType.LTEQ if self.match('=') else TokenType.LT)
        elif c == '>': self.add_token(TokenType.GTEQ if self.match('=') else TokenType.GT)
        elif c == '.':
            if self.match('.'):
                if self.match('.'):
                    self.add_token(TokenType.DOTDOTDOT)
                else:
                    self.add_token(TokenType.DOTDOT)
            else:
                self.add_token(TokenType.DOT)
                
        elif c == '/':
            if self.match('/'):
                # Line comment
                while self.peek() != '\n' and not self.is_at_end():
                    self.advance()
            elif self.match('*'):
                # Block comment
                self.consume_block_comment()
            else:
                self.add_token(TokenType.SLASH)
                
        elif c == '"':
            if self.match('"'):
                if self.match('"'):
                    self.consume_raw_string()
                else:
                    self.add_token(TokenType.STRING, "")
            else:
                self.consume_string()
                
        elif c.isdigit():
            self.consume_number()
            
        elif self.is_alpha(c):
            self.consume_identifier()
            
        else:
            raise RuntimeError(f"Unexpected character {c} at line {self.line}")

    def consume_block_comment(self):
        nesting = 1
        while nesting > 0:
            if self.is_at_end():
                raise RuntimeError(f"Unterminated block comment at line {self.line}")
            
            if self.peek() == '/' and self.peek_next() == '*':
                self.advance()
                self.advance()
                nesting += 1
            elif self.peek() == '*' and self.peek_next() == '/':
                self.advance()
                self.advance()
                nesting -= 1
            elif self.peek() == '\n':
                self.line += 1
                self.advance()
            else:
                self.advance()

    def consume_string(self):
        value = []
        # Run until we hit an unescaped end-quote or the %( interpolation trigger
        while not self.is_at_end():
            if self.peek() == '"':
                break
            if self.peek() == '%' and self.peek_next() == '(':
                break
                
            c = self.advance()
            if c == '\n':
                self.line += 1
                
            if c == '\\':
                nc = self.advance()
                if nc == 'n': value.append('\n')
                elif nc == '"': value.append('"')
                elif nc == '\\': value.append('\\')
                elif nc == 'r': value.append('\r')
                elif nc == 't': value.append('\t')
                elif nc == 'b': value.append('\b')
                elif nc == '0': value.append('\0')
                elif nc == 'a': value.append('\a')
                elif nc == 'v': value.append('\v')
                elif nc == '%': value.append('%')
                # Optional Unicode / Hex parsing could be expanded here
                else:
                    value.append(nc)
            else:
                value.append(c)

        if self.is_at_end():
            raise RuntimeError(f"Unterminated string at line {self.line}")

        if self.peek() == '"':
            self.advance()
            self.add_token(TokenType.STRING, "".join(value))
        elif self.peek() == '%' and self.peek_next() == '(':
            self.advance() # %
            self.advance() # (
            self.add_token(TokenType.INTERPOLATION, "".join(value))
            # Track the paren depth we expect to hit when this interpolation exits
            self.interpolation_depths.append(self.current_parens)
            
    def consume_raw_string(self):
        value = []
        while not self.is_at_end():
            if self.peek() == '"' and self.peek_next() == '"':
                # Check for triple
                self.advance()
                self.advance()
                if self.peek() == '"':
                    self.advance()
                    break
                else:
                    value.append('"')
                    value.append('"')
                    continue
            
            c = self.advance()
            if c == '\n':
                self.line += 1
            value.append(c)
            
        if self.is_at_end():
            raise RuntimeError(f"Unterminated raw string at line {self.line}")
            
        self.add_token(TokenType.STRING, "".join(value))

    def is_alpha(self, c: str) -> bool:
        return c.isalpha() or c == '_'
    
    def is_alphanumeric(self, c: str) -> bool:
        return self.is_alpha(c) or c.isdigit()

    def consume_identifier(self):
        while self.is_alphanumeric(self.peek()):
            self.advance()
        
        text = self.source[self.start:self.current]
        token_type = KEYWORDS.get(text, TokenType.NAME)
        self.add_token(token_type)

    def consume_number(self):
        while self.peek().isdigit():
            self.advance()

        # Fractional part
        if self.peek() == '.' and self.peek_next().isdigit():
            self.advance() # consume dot
            while self.peek().isdigit():
                self.advance()
                
        # Exponent (scientific notation)
        if self.peek() in 'eE':
            self.advance()
            if self.peek() in '+-':
                self.advance()
            while self.peek().isdigit():
                self.advance()

        val = float(self.source[self.start:self.current])
        self.add_token(TokenType.NUMBER, val)

