import wren_ast as ast
from wren_objects import WrenClass, WrenInstance, WrenFunction, WrenFiber, wren_type_name
from core_lib import get_native_method
from greenlet import greenlet, getcurrent
from tokens import TokenType

class Environment:
    def __init__(self, enclosing=None):
        self.enclosing = enclosing
        self.values = {}
        
    def define(self, name: str, value):
        self.values[name] = value
        
    def assign(self, name: str, value):
        if name in self.values:
            self.values[name] = value
            return
        if self.enclosing:
            self.enclosing.assign(name, value)
            return
        raise RuntimeError(f"Undefined variable '{name}'.")
        
    def get(self, name: str):
        if name in self.values:
            return self.values[name]
        if self.enclosing:
            return self.enclosing.get(name)
        raise RuntimeError(f"Undefined variable '{name}'.")


class ReturnException(Exception):
    def __init__(self, value):
        self.value = value

class BreakException(Exception):
    pass

class ContinueException(Exception):
    pass


class Interpreter:
    """
    Tree-walking AST evaluator executing parsed logic.
    Provides Method Dispatch via `call_method` and native Python integration.
    """
    def __init__(self, host):
        self.host = host
        self.globals = Environment()
        self.environment = self.globals
        
        # Inject standard core classes directly into global scope.
        for name in [
            "Object", "Class", "System", "Num", "String", "Bool", "Null",
            "List", "Map", "Fiber", "Fn", "Range", "Sequence", "MapEntry",
            "MapKeySequence", "MapValueSequence",
        ]:
            self.globals.define(name, WrenClass(name, None, {}))

    def is_truthy(self, value):
        return value is not False and value is not None

    def evaluate(self, expr):
        method_name = f"visit_{type(expr).__name__}"
        visitor = getattr(self, method_name)
        return visitor(expr)
        
    def execute(self, stmt):
        method_name = f"visit_{type(stmt).__name__}"
        visitor = getattr(self, method_name)
        visitor(stmt)
        
    def execute_block(self, statements, environment):
        previous = self.environment
        try:
            self.environment = environment
            for stmt in statements:
                self.execute(stmt)
        finally:
            self.environment = previous

    # --- Method / Function Dispatch ---
    
    def call_method(self, receiver, method_name: str, args: list):
        # 1. User defined (Instance methods)
        if isinstance(receiver, WrenInstance):
            if method_name in receiver.klass.methods:
                method = receiver.klass.methods[method_name]
                return self.execute_function(method, receiver, args)
                
        # 2. Native Core Library
        cls_name = wren_type_name(receiver)
        
        # Build strict signature
        if method_name.startswith("["):
            sig = method_name
        elif len(args) > 0:
            sig = f"{method_name}({','.join(['_']*len(args))})"
        elif " " not in method_name and not method_name.startswith("[") and not method_name.endswith("="):
            sig = f"{method_name}()"
        else:
            sig = method_name
            
        native_fn = get_native_method(cls_name, sig)
        
        # Fallback to no-paren signatures for getters 
        if not native_fn and len(args) == 0:
            native_fn = get_native_method(cls_name, method_name)
            
        if native_fn:
            return native_fn(self, receiver, args)
            
        # Error handling triggers runtime exit in spec
        raise RuntimeError(f"Method '{sig}' not found on {cls_name}")

    def execute_function(self, closure: WrenFunction, receiver, args: list):
        env = Environment(closure.closure)
        if receiver is not None:
            env.define("this", receiver)
            
        for i, param in enumerate(closure.declaration.parameters):
            if i < len(args):
                env.define(param.lexeme, args[i])
            else:
                env.define(param.lexeme, None)
                
        try:
            self.execute_block(closure.declaration.body.statements, env)
        except ReturnException as ret:
            return ret.value
        return None
        
    # --- Fiber Integration ---
    
    def create_fiber(self, closure: WrenFunction):
        def _fiber_runner(arg):
            try:
                # First argument to fiber yield/call is mapped linearly to first param
                # if the block requested parameters.
                params = [arg] if arg is not None else []
                self.execute_function(closure, None, params)
            except Exception as e:
                # Capture unescaped errors to stop loop correctly (Wren exits 70)
                raise
        green = greenlet(_fiber_runner)
        return WrenFiber(green)

    def resume_fiber(self, fiber: WrenFiber, arg):
        if fiber.is_done:
            raise RuntimeError("Cannot call a finished fiber.")
            
        # .switch() inherently saves caller state as parent, creating symmetric suspend/resume
        res = fiber.greenlet_obj.switch(arg)
        if fiber.greenlet_obj.dead:
            fiber.is_done = True
        return res
        
    def suspend_fiber(self, yield_value):
        parent = getattr(getcurrent(), 'parent', None)
        if parent is None:
            raise RuntimeError("Cannot yield from the root fiber.")
        return parent.switch(yield_value)

    # --- statement visitors ---
    
    def visit_ExprStmt(self, stmt: ast.ExprStmt):
        self.evaluate(stmt.expression)
        
    def visit_VarDecl(self, stmt: ast.VarDecl):
        value = None
        if stmt.initializer:
            value = self.evaluate(stmt.initializer)
        self.environment.define(stmt.name.lexeme, value)
        
    def visit_Block(self, stmt: ast.Block):
        self.execute_block(stmt.statements, Environment(self.environment))

    def visit_IfStmt(self, stmt: ast.IfStmt):
        cond = self.evaluate(stmt.condition)
        if self.is_truthy(cond):
            self.execute(stmt.then_branch)
        elif stmt.else_branch:
            self.execute(stmt.else_branch)
            
    def visit_WhileStmt(self, stmt: ast.WhileStmt):
        while True:
            cond = self.evaluate(stmt.condition)
            if not self.is_truthy(cond):
                break
            try:
                self.execute(stmt.body)
            except ContinueException:
                continue
            except BreakException:
                break
            
    def visit_ForStmt(self, stmt: ast.ForStmt):
        iterable = self.evaluate(stmt.iterable)
        # Assuming iterable is mapped cleanly to a Python iter
        try:
            it = iter(iterable)
        except TypeError:
            raise RuntimeError("Object is not iterable.")
            
        for val in it:
            env = Environment(self.environment)
            env.define(stmt.variable.lexeme, val)
            try:
                if isinstance(stmt.body, ast.Block):
                    self.execute_block(stmt.body.statements, env)
                else:
                    previous = self.environment
                    try:
                        self.environment = env
                        self.execute(stmt.body)
                    finally:
                        self.environment = previous
            except ContinueException:
                continue
            except BreakException:
                break
            
    def visit_ReturnStmt(self, stmt: ast.ReturnStmt):
        value = None
        if stmt.value:
            value = self.evaluate(stmt.value)
        raise ReturnException(value)

    def visit_BreakStmt(self, stmt: ast.BreakStmt):
        raise BreakException()

    def visit_ContinueStmt(self, stmt: ast.ContinueStmt):
        raise ContinueException()
        
    def visit_ClassDecl(self, stmt: ast.ClassDecl):
        # We declare empty then configure to allow self references within methods
        self.environment.define(stmt.name.lexeme, None)
        
        methods = {}
        for method in stmt.methods:
            func = WrenFunction(method, self.environment)
            # using signature mapping simplified
            methods[method.name.lexeme] = func
            
        wren_class = WrenClass(stmt.name.lexeme, None, methods)
        self.environment.assign(stmt.name.lexeme, wren_class)

    def visit_MethodDecl(self, stmt: ast.MethodDecl):
        pass # Unused explicitly outside class body eval

    # --- expression visitors ---
    
    def visit_Literal(self, expr: ast.Literal):
        return expr.value
        
    def visit_Variable(self, expr: ast.Variable):
        return self.environment.get(expr.name.lexeme)
        
    def visit_VarAssign(self, expr: ast.VarAssign):
        val = self.evaluate(expr.value)
        self.environment.assign(expr.name.lexeme, val)
        return val
        
    def visit_Logical(self, expr: ast.Logical):
        left = self.evaluate(expr.left)
        if expr.operator.type == TokenType.PIPEPIPE:
            if self.is_truthy(left): return left
        else: # AMPAMP
            if not self.is_truthy(left): return left
        return self.evaluate(expr.right)

    def visit_Conditional(self, expr: ast.Conditional):
        if self.is_truthy(self.evaluate(expr.condition)):
            return self.evaluate(expr.then_expr)
        return self.evaluate(expr.else_expr)
        
    def visit_Call(self, expr: ast.Call):
        receiver = None
        if expr.receiver is not None:
            receiver = self.evaluate(expr.receiver)
        else:
            # implicit 'this' or class creation
            # If the variable maps to a class directly, we spawn instance
            try:
                receiver = self.environment.get("this")
            except RuntimeError:
                # must be a core class instantiation or variable closure?
                val = self.environment.get(expr.method.lexeme)
                if isinstance(val, WrenClass):
                    return WrenInstance(val)
                return val

        args = [self.evaluate(a) for a in expr.arguments]
        
        # intercept construction calls
        if isinstance(receiver, WrenClass) and expr.method.lexeme == "new":
            if get_native_method(wren_type_name(receiver), f"new({','.join(['_']*len(args))})"):
                return self.call_method(receiver, expr.method.lexeme, args)
            # Instantiation directly
            inst = WrenInstance(receiver)
            return inst
            
        return self.call_method(receiver, expr.method.lexeme, args)
        
    def visit_Function(self, expr: ast.Function):
        # returns the closure
        return WrenFunction(expr, self.environment)
        
    def visit_This(self, expr: ast.This):
        return self.environment.get("this")
        
    def visit_StringInterpolation(self, expr: ast.StringInterpolation):
        # evaluated parts concatenated safely through method calls or direct toString
        res = ""
        for p in expr.expressions:
            val = self.evaluate(p)
            str_val = self.call_method(val, "toString", [])
            res += str(str_val)
        return res

    def visit_ListLiteral(self, expr: ast.ListLiteral):
        return [self.evaluate(el) for el in expr.elements]
        
    def visit_MapLiteral(self, expr: ast.MapLiteral):
        keys = [self.evaluate(k) for k in expr.keys]
        vals = [self.evaluate(v) for v in expr.values]
        return dict(zip(keys, vals))
