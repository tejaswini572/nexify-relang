NATIVE_METHODS = {}

from wren_objects import WrenClass, WrenInstance, wren_type_name

def register(class_name: str, method_signature: str):
    """
    Decorator that registers a native python function under a Wren class and signature.
    Signature should match the literal method token format, e.g. "+(_)" or "toString".
    """
    def decorator(func):
        NATIVE_METHODS[(class_name, method_signature)] = func
        return func
    return decorator

def get_native_method(class_name: str, method_signature: str):
    # Attempt literal class map
    func = NATIVE_METHODS.get((class_name, method_signature))
    if func is not None:
        return func
        
    # Standard Object fallback (==, !=, etc)
    if class_name != "Object":
        return NATIVE_METHODS.get(("Object", method_signature))
        
    return None

def wren_to_string(ctx, value):
    return str(ctx.call_method(value, "toString", []))

def wren_index(value):
    if isinstance(value, float) and value.is_integer():
        return int(value)
    return value

def wren_class_for(ctx, name):
    return ctx.globals.get(name)

# --- Object ---
@register("Object", "==(_)")
def obj_eq(ctx, receiver, args):
    # Wren requires 0 == -0 evaluation
    if isinstance(receiver, float) and isinstance(args[0], float):
        if receiver == 0.0 and args[0] == 0.0:
            return True
    return receiver == args[0]

@register("Object", "!=(_)")
def obj_neq(ctx, receiver, args):
    return not obj_eq(ctx, receiver, args)

@register("Object", "is(_)")
def obj_is(ctx, receiver, args):
    target = args[0]
    if not isinstance(target, WrenClass):
        return False
    if target.name == "Object":
        return True
    if isinstance(receiver, WrenClass):
        return target.name == "Class"
    if isinstance(receiver, WrenInstance):
        return receiver.klass.name == target.name
    return wren_type_name(receiver) == target.name

@register("Object", "type")
def obj_type(ctx, receiver, args):
    if isinstance(receiver, WrenClass):
        if receiver.name == "Class" or receiver.name.endswith(" metaclass"):
            return wren_class_for(ctx, "Class")
        return WrenClass(f"{receiver.name} metaclass", wren_class_for(ctx, "Class"), {})
    if isinstance(receiver, WrenInstance):
        return receiver.klass
    return wren_class_for(ctx, wren_type_name(receiver))

@register("Object", "name")
def obj_name(ctx, receiver, args):
    if isinstance(receiver, WrenClass):
        return receiver.name
    return wren_type_name(receiver)

@register("Object", "toString")
def obj_tostring(ctx, receiver, args):
    return f"instance of {wren_type_name(receiver)}"

@register("Object", "!()")
def obj_not(ctx, receiver, args):
    return False

# --- Bool ---
@register("Bool", "!()")
def bool_not(ctx, receiver, args):
    return not receiver

@register("Bool", "toString")
def bool_tostring(ctx, receiver, args):
    return "true" if receiver else "false"

# --- Null ---
@register("Null", "!()")
def null_not(ctx, receiver, args):
    return True

@register("Null", "toString")
def null_tostring(ctx, receiver, args):
    return "null"

# --- Num ---
@register("Num", "-()")
def num_negate(ctx, receiver, args): return -receiver
@register("Num", "+(_)")
def num_add(ctx, receiver, args): return receiver + args[0]
@register("Num", "-(_)")
def num_sub(ctx, receiver, args): return receiver - args[0]
@register("Num", "*(_)")
def num_mul(ctx, receiver, args): return receiver * args[0]
@register("Num", "/(_)")
def num_div(ctx, receiver, args): return receiver / args[0]
@register("Num", "%(_)")
def num_mod(ctx, receiver, args): return receiver % args[0]

@register("Num", "<(_)")
def num_lt(ctx, receiver, args): return receiver < args[0]
@register("Num", "<=(_)")
def num_lte(ctx, receiver, args): return receiver <= args[0]
@register("Num", ">(_)")
def num_gt(ctx, receiver, args): return receiver > args[0]
@register("Num", ">=(_)")
def num_gte(ctx, receiver, args): return receiver >= args[0]

@register("Num", "toString")
def num_tostring(ctx, receiver, args):
    """Formats doubles using C %.14g rules exactly."""
    if receiver == int(receiver):
        # Python %g sometimes truncates whole integers nicely, but float checking is safer
        return str(int(receiver))
    return f"{receiver:.14g}"

# --- String ---
@register("String", "toString")
def str_tostring(ctx, receiver, args): return receiver

@register("String", "+(_)")
def str_add(ctx, receiver, args):
    return str(receiver) + str(args[0])

@register("String", "split(_)")
def str_split(ctx, receiver, args):
    return receiver.split(args[0])
    
@register("String", "trim()")
def str_trim(ctx, receiver, args):
    return receiver.strip()
    
@register("String", "trim(_)")
def str_trim_chars(ctx, receiver, args):
    return receiver.strip(args[0])
    
@register("String", "contains(_)")
def str_contains(ctx, receiver, args):
    return args[0] in receiver

# --- System ---
@register("System class", "print(_)")
def sys_print(ctx, receiver, args):
    """Calls toString dynamically on the argument, then outputs."""
    # We yield into the interpreter method dispatcher to stringify the object correctly
    val = ctx.call_method(args[0], "toString", [])
    ctx.host.print_output(str(val) + "\n")
    return args[0]

@register("System class", "print()")
def sys_print_empty(ctx, receiver, args):
    ctx.host.print_output("\n")
    return None

# --- List ---
@register("List class", "new()")
def list_new(ctx, receiver, args):
    return []

@register("List", "toString")
def list_tostring(ctx, receiver, args):
    return "[" + ", ".join(wren_to_string(ctx, item) for item in receiver) + "]"

@register("List", "add(_)")
def list_add(ctx, receiver, args):
    receiver.append(args[0])
    return args[0]

@register("List", "count")
def list_count(ctx, receiver, args):
    return len(receiver)
    
@register("List", "[_]")
def list_get(ctx, receiver, args):
    return receiver[wren_index(args[0])]

@register("List", "[_]=(_)")
def list_set(ctx, receiver, args):
    receiver[wren_index(args[0])] = args[1]
    return args[1]

# --- Map ---
# Standard Python 3.7+ dictionaries naturally map to insertion order
@register("Map", "toString")
def map_tostring(ctx, receiver, args):
    pairs = []
    for key, value in receiver.items():
        pairs.append(f"{wren_to_string(ctx, key)}: {wren_to_string(ctx, value)}")
    return "{" + ", ".join(pairs) + "}"

@register("Map", "count")
def map_count(ctx, receiver, args):
    return len(receiver)
    
@register("Map", "[_]")
def map_get(ctx, receiver, args):
    return receiver.get(args[0], None)
    
@register("Map", "[_]=(_)")
def map_set(ctx, receiver, args):
    receiver[args[0]] = args[1]
    return args[1]

# --- Fiber ---
@register("Fiber class", "new(_)")
def fiber_new(ctx, receiver, args):
    # args[0] is the closure Block/Fn
    return ctx.create_fiber(args[0])

# --- Fn ---
@register("Fn class", "new(_)")
def fn_new(ctx, receiver, args):
    return args[0]

@register("Fn", "call()")
def fn_call(ctx, receiver, args):
    return ctx.execute_function(receiver, None, [])

@register("Fn", "call(_)")
def fn_call_one(ctx, receiver, args):
    return ctx.execute_function(receiver, None, args)

@register("Fiber", "call()")
def fiber_call(ctx, receiver, args):
    return ctx.resume_fiber(receiver, None)

@register("Fiber", "call(_)")
def fiber_call_val(ctx, receiver, args):
    return ctx.resume_fiber(receiver, args[0])

@register("Fiber class", "yield()")
def fiber_yield(ctx, receiver, args):
    return ctx.suspend_fiber(None)

@register("Fiber class", "yield(_)")
def fiber_yield_val(ctx, receiver, args):
    return ctx.suspend_fiber(args[0])

