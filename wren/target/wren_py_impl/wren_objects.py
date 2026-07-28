class WrenClass:
    """Represents a user-defined or mock native class."""
    def __init__(self, name: str, superclass, methods: dict):
        self.name = name
        self.superclass = superclass
        self.methods = methods
        # In a strict implementation, classes are instances of metaclasses.
        # This MVP handles static calls gracefully in the interpreter.

class WrenInstance:
    """Represents an instance of a user-defined class."""
    def __init__(self, klass: WrenClass):
        self.klass = klass
        self.fields = {}
        
    def __repr__(self):
        return f"instance of {self.klass.name}"

class WrenFunction:
    """Represents a callable closure."""
    def __init__(self, declaration, closure: dict):
        self.declaration = declaration  # ast.Function or MethodDecl
        self.closure = closure          # captured environment dictionary

class WrenFiber:
    """Represents a runtime Fiber, holding its greenlet execution context."""
    def __init__(self, greenlet_obj):
        self.greenlet_obj = greenlet_obj
        self.error = None
        self.is_done = False


def wren_type_name(obj) -> str:
    """
    To maximize developer velocity and minimize object wrapping overhead,
    pure Wren primitives map strictly to underlying Python primitives.
    This function bridges Python's type system to Wren's class nomenclature.
    """
    if isinstance(obj, bool): return "Bool" # Check bool before int, as bool subclasses int in Python
    if isinstance(obj, (int, float)): return "Num"
    if isinstance(obj, str): return "String"
    if obj is None: return "Null"
    if isinstance(obj, list): return "List"
    if isinstance(obj, dict): return "Map"
    if isinstance(obj, WrenFiber): return "Fiber"
    if isinstance(obj, WrenFunction): return "Fn"
    
    if isinstance(obj, WrenClass): 
        # A static method call on a Class
        return f"{obj.name} class"
        
    if isinstance(obj, WrenInstance): 
        return obj.klass.name
        
    return "Object"
