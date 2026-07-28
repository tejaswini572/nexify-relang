from tokens import TokenType, Token
import wren_ast as ast

class Parser:
    """
    Recursive descent parser for Wren mapping directly into the custom AST.
    Operator bridging transforms `1 + 2` -> `1.+(2)` at parse time.
    """
    def __init__(self, tokens: list[Token]):
        self.tokens = tokens
        self.current = 0

    def parse(self) -> list[ast.Stmt]:
        statements = []
        while not self.is_at_end():
            self.skip_newlines()
            if self.is_at_end():
                break
            statements.append(self.declaration())
        return statements

    def is_at_end(self) -> bool:
        return self.peek().type == TokenType.EOF

    def peek(self) -> Token:
        return self.tokens[self.current]

    def previous(self) -> Token:
        return self.tokens[self.current - 1]

    def advance(self) -> Token:
        if not self.is_at_end():
            self.current += 1
        return self.previous()

    def check(self, type: TokenType) -> bool:
        if self.is_at_end():
            return False
        return self.peek().type == type

    def match(self, *types: TokenType) -> bool:
        for t in types:
            if self.check(t):
                self.advance()
                return True
        return False

    def consume(self, type: TokenType, message: str) -> Token:
        if self.check(type):
            return self.advance()
        raise RuntimeError(f"Parse error at line {self.peek().line}: {message}")

    def skip_newlines(self):
        while self.match(TokenType.NEWLINE):
            pass

    # --- Statements ---

    def declaration(self) -> ast.Stmt:
        if self.match(TokenType.CLASS):
            return self.class_decl()
        if self.match(TokenType.VAR):
            return self.var_decl()
        return self.statement()

    def class_decl(self) -> ast.Stmt:
        name = self.consume(TokenType.NAME, "Expect class name.")
        superclass = None
        if self.match(TokenType.IS):
            # Inheritance: class Foo is Bar
            self.consume(TokenType.NAME, "Expect superclass name.")
            superclass = ast.Variable(self.previous())
            
        self.consume(TokenType.LBRACE, "Expect '{' before class body.")
        methods = []
        self.skip_newlines()
        while not self.check(TokenType.RBRACE) and not self.is_at_end():
            methods.append(self.method_decl())
            self.skip_newlines()
        
        self.consume(TokenType.RBRACE, "Expect '}' after class body.")
        return ast.ClassDecl(name, superclass, methods)

    def method_decl(self) -> ast.MethodDecl:
        is_static = self.match(TokenType.STATIC)
        is_construct = self.match(TokenType.CONSTRUCT)
        
        name = self.consume(TokenType.NAME, "Expect method name.")
        
        parameters = []
        if self.match(TokenType.LPAREN):
            if not self.check(TokenType.RPAREN):
                parameters.append(self.consume(TokenType.NAME, "Expect parameter name."))
                while self.match(TokenType.COMMA):
                    parameters.append(self.consume(TokenType.NAME, "Expect parameter name."))
            self.consume(TokenType.RPAREN, "Expect ')' after parameters.")
            
        self.consume(TokenType.LBRACE, "Expect '{' before method body.")
        body = ast.Block(self.block_body())
        return ast.MethodDecl(is_static, is_construct, name, parameters, body)

    def var_decl(self) -> ast.Stmt:
        name = self.consume(TokenType.NAME, "Expect variable name.")
        initializer = None
        if self.match(TokenType.EQ):
            self.skip_newlines()
            initializer = self.expression()
        return ast.VarDecl(name, initializer)

    def statement(self) -> ast.Stmt:
        if self.match(TokenType.FOR): return self.for_stmt()
        if self.match(TokenType.IF): return self.if_stmt()
        if self.match(TokenType.RETURN): return self.return_stmt()
        if self.match(TokenType.BREAK): return ast.BreakStmt(self.previous())
        if self.match(TokenType.CONTINUE): return ast.ContinueStmt(self.previous())
        if self.match(TokenType.WHILE): return self.while_stmt()
        if self.match(TokenType.LBRACE): return ast.Block(self.block_body())
        return self.expr_stmt()

    def block_body(self) -> list[ast.Stmt]:
        statements = []
        self.skip_newlines()
        while not self.check(TokenType.RBRACE) and not self.is_at_end():
            statements.append(self.declaration())
            self.skip_newlines()
        self.consume(TokenType.RBRACE, "Expect '}' after block.")
        return statements

    def for_stmt(self) -> ast.Stmt:
        self.consume(TokenType.LPAREN, "Expect '(' after 'for'.")
        var_name = self.consume(TokenType.NAME, "Expect loop variable.")
        self.consume(TokenType.IN, "Expect 'in' after loop variable.")
        self.skip_newlines()
        iterable = self.expression()
        self.consume(TokenType.RPAREN, "Expect ')' after loop construct.")
        body = self.statement()
        return ast.ForStmt(var_name, iterable, body)

    def if_stmt(self) -> ast.Stmt:
        self.consume(TokenType.LPAREN, "Expect '(' after 'if'.")
        condition = self.expression()
        self.consume(TokenType.RPAREN, "Expect ')' after if condition.")
        then_branch = self.statement()
        else_branch = None
        if self.match(TokenType.ELSE):
            else_branch = self.statement()
        return ast.IfStmt(condition, then_branch, else_branch)

    def while_stmt(self) -> ast.Stmt:
        self.consume(TokenType.LPAREN, "Expect '(' after 'while'.")
        condition = self.expression()
        self.consume(TokenType.RPAREN, "Expect ')' after while condition.")
        body = self.statement()
        return ast.WhileStmt(condition, body)

    def return_stmt(self) -> ast.Stmt:
        keyword = self.previous()
        value = None
        # In Wren, newlines are statement terminators.
        # If we see a newline, it's an empty return.
        if not self.check(TokenType.NEWLINE) and not self.check(TokenType.RBRACE):
            value = self.expression()
        return ast.ReturnStmt(keyword, value)

    def expr_stmt(self) -> ast.Stmt:
        expr = self.expression()
        return ast.ExprStmt(expr)

    # --- Expressions ---

    def expression(self) -> ast.Expr:
        return self.assignment()

    def assignment(self) -> ast.Expr:
        expr = self.conditional()
        
        if self.match(TokenType.EQ):
            equals = self.previous()
            value = self.assignment()
            
            if isinstance(expr, ast.Variable):
                return ast.VarAssign(expr.name, value)
            elif isinstance(expr, ast.Call) and expr.method.lexeme == "[_]":
                new_method = Token(TokenType.NAME, "[_]=(_)", expr.method.line)
                return ast.Call(expr.receiver, new_method, [*expr.arguments, value])
            elif isinstance(expr, ast.Call) and len(expr.arguments) == 0:
                # obj.prop = val -> obj.prop=(val) method call bridging
                method_name = f"{expr.method.lexeme}="
                new_method = Token(TokenType.NAME, method_name, expr.method.line)
                return ast.Call(expr.receiver, new_method, [value])
            else:
                raise RuntimeError("Invalid assignment target.")
                
        return expr

    def conditional(self) -> ast.Expr:
        expr = self.logic_or()
        if self.match(TokenType.QUESTION):
            self.skip_newlines()
            then_expr = self.expression()
            self.consume(TokenType.COLON, "Expect ':' after then branch of conditional.")
            self.skip_newlines()
            else_expr = self.conditional()
            expr = ast.Conditional(expr, then_expr, else_expr)
        return expr

    def logic_or(self) -> ast.Expr:
        expr = self.logic_and()
        while self.match(TokenType.PIPEPIPE):
            op = self.previous()
            self.skip_newlines()
            right = self.logic_and()
            expr = ast.Logical(expr, op, right)
        return expr

    def logic_and(self) -> ast.Expr:
        expr = self.is_expr()
        while self.match(TokenType.AMPAMP):
            op = self.previous()
            self.skip_newlines()
            right = self.is_expr()
            expr = ast.Logical(expr, op, right)
        return expr

    def is_expr(self) -> ast.Expr:
        expr = self.equality()
        if self.match(TokenType.IS):
            op = self.previous()
            self.skip_newlines()
            right = self.equality()
            # Compile 'is' down to a standard method call .is(other)
            # This enables dynamic dispatch overrides seamlessly.
            expr = ast.Call(expr, op, [right])
        return expr

    def parse_binary(self, next_parser, *tokens) -> ast.Expr:
        expr = next_parser()
        while self.match(*tokens):
            op = self.previous()
            self.skip_newlines()
            right = next_parser()
            # Compile standard operators down to generic method calls per Wren spec
            expr = ast.Call(expr, op, [right])
        return expr

    def equality(self) -> ast.Expr: return self.parse_binary(self.comparison, TokenType.EQEQ, TokenType.BANGEQ)
    def comparison(self) -> ast.Expr: return self.parse_binary(self.bitwise_or, TokenType.GT, TokenType.GTEQ, TokenType.LT, TokenType.LTEQ)
    def bitwise_or(self) -> ast.Expr: return self.parse_binary(self.bitwise_xor, TokenType.PIPE)
    def bitwise_xor(self) -> ast.Expr: return self.parse_binary(self.bitwise_and, TokenType.CARET)
    def bitwise_and(self) -> ast.Expr: return self.parse_binary(self.term, TokenType.AMP)
    def term(self) -> ast.Expr: return self.parse_binary(self.factor, TokenType.MINUS, TokenType.PLUS)
    def factor(self) -> ast.Expr: return self.parse_binary(self.unary, TokenType.SLASH, TokenType.STAR, TokenType.PERCENT)

    def unary(self) -> ast.Expr:
        if self.match(TokenType.BANG, TokenType.MINUS, TokenType.TILDE):
            op = self.previous()
            self.skip_newlines()
            right = self.unary()
            # Convert Unary to 0-argument method call on receiver `right`
            # Wait! Unary minus `-x` translates to `x.-()`. 
            return ast.Call(right, op, [])
        return self.call()

    def call(self) -> ast.Expr:
        expr = self.primary()
        
        while True:
            if self.match(TokenType.LPAREN):
                # Standard invocation
                args = []
                if not self.check(TokenType.RPAREN):
                    args.append(self.expression())
                    while self.match(TokenType.COMMA):
                        args.append(self.expression())
                self.consume(TokenType.RPAREN, "Expect ')' after arguments.")
                
                if isinstance(expr, ast.Variable):
                    # implicit 'this' call: foo(1) -> this.foo(1)
                    expr = ast.Call(None, expr.name, args)
                elif isinstance(expr, ast.Call) and len(expr.arguments) == 0:
                    # obj.foo(1) - we parsed obj.foo as a 0-arg call, now we populate arguments
                    expr.arguments = args
                else:
                    # e.g., (func_expr)() -> call method named "call" on object
                    call_token = Token(TokenType.NAME, "call", getattr(self.previous(), 'line', 0))
                    expr = ast.Call(expr, call_token, args)
            
            elif self.match(TokenType.DOT):
                name = self.consume(TokenType.NAME, "Expect property name after '.'.")
                expr = ast.Call(expr, name, [])
            
            elif self.match(TokenType.LBRACKET):
                args = []
                args.append(self.expression())
                while self.match(TokenType.COMMA):
                    args.append(self.expression())
                self.consume(TokenType.RBRACKET, "Expect ']' after subscript.")
                
                # subscript is a method named [_] taking args
                method = Token(TokenType.NAME, "[_]", getattr(expr, 'line', 0))
                expr = ast.Call(expr, method, args)

            elif self.match(TokenType.LBRACE):
                params = []
                if self.match(TokenType.PIPE):
                    if not self.check(TokenType.PIPE):
                        params.append(self.consume(TokenType.NAME, "Expect parameter name."))
                        while self.match(TokenType.COMMA):
                            params.append(self.consume(TokenType.NAME, "Expect parameter name."))
                    self.consume(TokenType.PIPE, "Expect '|' after block parameters.")

                block_body = self.block_body()
                fn = ast.Function(params, ast.Block(block_body))

                if isinstance(expr, ast.Variable):
                    expr = ast.Call(None, expr.name, [fn])
                elif isinstance(expr, ast.Call):
                    expr.arguments.append(fn)
                else:
                    raise RuntimeError("Cannot pass block argument to this expression.")
                
            else:
                break

        return expr

    def primary(self) -> ast.Expr:
        if self.match(TokenType.FALSE): return ast.Literal(False, self.previous())
        if self.match(TokenType.TRUE): return ast.Literal(True, self.previous())
        if self.match(TokenType.NULL): return ast.Literal(None, self.previous())
        if self.match(TokenType.NUMBER, TokenType.STRING):
            return ast.Literal(self.previous().value, self.previous())
        if self.match(TokenType.THIS): return ast.This(self.previous())
        if self.match(TokenType.NAME): return ast.Variable(self.previous())
        
        if self.match(TokenType.INTERPOLATION):
            parts = [ast.Literal(self.previous().value, self.previous())]
            
            while True:
                parts.append(self.expression())
                if self.match(TokenType.STRING):
                    parts.append(ast.Literal(self.previous().value, self.previous()))
                    break
                elif self.match(TokenType.INTERPOLATION):
                    parts.append(ast.Literal(self.previous().value, self.previous()))
                else:
                    self.error(self.peek(), "Expected end of interpolation string.")
                    break
            return ast.StringInterpolation(parts)
            
        if self.match(TokenType.LBRACKET):
            elements = []
            if not self.check(TokenType.RBRACKET):
                elements.append(self.expression())
                while self.match(TokenType.COMMA):
                    self.skip_newlines()
                    elements.append(self.expression())
            self.consume(TokenType.RBRACKET, "Expect ']' after list elements.")
            return ast.ListLiteral(elements)
            
        if self.match(TokenType.LBRACE):
            keys = []
            values = []
            if not self.check(TokenType.RBRACE):
                keys.append(self.expression())
                self.consume(TokenType.COLON, "Expect ':' after map key.")
                values.append(self.expression())
                while self.match(TokenType.COMMA):
                    keys.append(self.expression())
                    self.consume(TokenType.COLON, "Expect ':' after map key.")
                    values.append(self.expression())
            self.consume(TokenType.RBRACE, "Expect '}' after map elements.")
            return ast.MapLiteral(keys, values)

        if self.match(TokenType.LPAREN):
            expr = self.expression()
            self.consume(TokenType.RPAREN, "Expect ')' after expression.")
            return expr

        raise RuntimeError(f"Expect expression at {self.peek().lexeme}")
