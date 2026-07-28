import os

class PathType:
    ABSOLUTE = 1
    RELATIVE = 2
    SIMPLE = 3

def classify_path(path: str) -> int:
    """
    Classifies a path snippet strictly by its prefix to maintain classification stability.
    - ABSOLUTE: starts with `/` or `C:` etc.
    - RELATIVE: starts with `./` or `../`
    - SIMPLE: anything else (logical paths)
    This guarantees that internal or trailing segments do not affect classification.
    """
    if path.startswith("/"):
        return PathType.ABSOLUTE
    # Windows drive letter logic
    if len(path) >= 2 and path[0].isalpha() and path[1] == ':':
        return PathType.ABSOLUTE

    if path.startswith("./") or path.startswith("../"):
        return PathType.RELATIVE

    return PathType.SIMPLE

def directory_name(path: str) -> str:
    """
    Strips trailing characters following and including the final slash.
    Returns empty string if no separator exists.
    """
    idx = path.rfind('/')
    if idx == -1:
        return ""
    return path[:idx]

def remove_extension(path: str) -> str:
    """
    Scans backwards for a dot before a slash. If found, truncates from the dot.
    Does nothing if there is no dot, or if the dot is part of a directory name (before final slash).
    """
    dot_idx = path.rfind('.')
    slash_idx = path.rfind('/')
    
    if dot_idx > slash_idx:
        return path[:dot_idx]
    return path

def normalize_path(path: str) -> str:
    """
    Normalizes a path array:
    - Splits by slash.
    - Discards `.` components.
    - Consumes `..` components by eliminating the preceding structural path component.
    - Recompiles using exactly one standard separator `/`.
    
    Invariants satisfied:
    - Preserves absolute root identity.
    - Preserves relative syntax prefix (`./` or `../`) if it existed to ensure classification stability.
    - Aborts program completely if structural limits (e.g. >2048 components) are exceeded.
    """
    MAX_COMPONENTS = 2048
    
    components = path.split('/')
    if len(components) > MAX_COMPONENTS:
        import sys
        print(f"Path exceeds maximum structural depth of {MAX_COMPONENTS} components.", file=sys.stderr)
        sys.exit(1)
        
    is_absolute = classify_path(path) == PathType.ABSOLUTE
    starts_with_current = path.startswith("./")
    
    stack = []
    
    for comp in components:
        if comp == '' or comp == '.':
            continue
        elif comp == '..':
            if stack and stack[-1] != '..':
                stack.pop()
            elif not is_absolute:
                # Can't back out further, but since it's not absolute, preserve the '..'
                stack.append('..')
        else:
            stack.append(comp)

    # Reconstruct path
    result = '/'.join(stack)
    
    if is_absolute:
        # Guarantee absolute prefix
        if path.startswith("/"):
            return "/" + result
        else:
            # Drive letter e.g. "C:"
            prefix = path[:2]
            return prefix + ("/" + result if result else "")
            
    if not result:
        return "."

    # Preserve explicit relative prefix for stability
    if starts_with_current and not result.startswith(".."):
        return "./" + result

    return result
