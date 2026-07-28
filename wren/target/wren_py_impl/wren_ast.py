from dataclasses import dataclass
from typing import Any, Optional
from tokens import Token

class Expr: pass
class Stmt: pass

# --- Expressions ---

@dataclass
class Literal(Expr):
    value: Any
    token: Token

@dataclass
class StringInterpolation(Expr):
    # e.g., "a %(b) c" compiles to concatenated StringInterpolations or Binary '+'
    # Keeping it as a list of Expr makes evaluation trivial.
    expressions: list[Expr]

@dataclass
class ListLiteral(Expr):
    elements: list[Expr]

@dataclass
class MapLiteral(Expr):
    keys: list[Expr]
    values: list[Expr]

@dataclass
class Variable(Expr):
    name: Token
    
@dataclass
class VarAssign(Expr):
    name: Token
    value: Expr

@dataclass
class Field(Expr):
    name: Token
    
@dataclass
class FieldAssign(Expr):
    name: Token
    value: Expr

@dataclass
class Logical(Expr):
    # Kept distinct from 'Call' because '&&' and '||' must short-circuit evaluators
    left: Expr
    operator: Token
    right: Expr

@dataclass
class Conditional(Expr):
    condition: Expr
    then_expr: Expr
    else_expr: Expr

@dataclass
class Call(Expr):
    # Represents method calls `obj.method(args)`, property gets `obj.prop` (0-args),
    # property sets `obj.prop = val` (1-arg `prop=`), array indexing `obj[idx]` (`[_]`),
    # AND binary/unary operators `left + right` -> `left.+(right)`.
    receiver: Optional[Expr] # None implies 'this' implicitly in current scope
    method: Token
    arguments: list[Expr]

@dataclass
class Super(Expr):
    method: Token
    arguments: list[Expr]

@dataclass
class Function(Expr):
    # Wren's blocks `{ |x| print(x) }` are true closures/functions that evaluate as expressions
    parameters: list[Token]
    body: Stmt # typically a Block statement

@dataclass
class This(Expr):
    keyword: Token


# --- Statements ---

@dataclass
class ExprStmt(Stmt):
    expression: Expr

@dataclass
class VarDecl(Stmt):
    name: Token
    initializer: Optional[Expr]

@dataclass
class Block(Stmt):
    statements: list[Stmt]

@dataclass
class IfStmt(Stmt):
    condition: Expr
    then_branch: Stmt
    else_branch: Optional[Stmt]

@dataclass
class WhileStmt(Stmt):
    condition: Expr
    body: Stmt

@dataclass
class ForStmt(Stmt):
    variable: Token
    iterable: Expr
    body: Stmt

@dataclass
class ReturnStmt(Stmt):
    keyword: Token
    value: Optional[Expr]

@dataclass
class BreakStmt(Stmt):
    keyword: Token
    
@dataclass
class ContinueStmt(Stmt):
    keyword: Token

@dataclass
class MethodDecl(Stmt):
    is_static: bool
    is_construct: bool
    name: Token
    parameters: list[Token]
    body: Stmt

@dataclass
class ClassDecl(Stmt):
    name: Token
    superclass: Optional[Expr]
    methods: list[MethodDecl]
