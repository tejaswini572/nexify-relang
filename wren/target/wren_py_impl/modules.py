import os
import sys
from paths import classify_path, directory_name, normalize_path, PathType

# Hardcoded standard library stubs 
# A complete implementation would inject Wren source code strings here.
BUILTIN_MODULES = {
    "io": "// io stub",
    "os": "// os stub",
    "scheduler": "// scheduler stub",
    "timer": "// timer stub",
    "repl": "// repl stub" 
}

def resolve_module(importer: str, requested: str) -> str:
    """
    Transforms import statements into standardized module targets.
    
    Algorithm:
    1. Classify requested path type.
    2. If Logical (SIMPLE): Return exactly as-is. Logical paths are opaque.
    3. If Relative: Append requested path to the importer's directory and normalize.
    """
    path_type = classify_path(requested)
    
    if path_type == PathType.SIMPLE:
        # Logistic path: opaque during resolution
        return requested
    elif path_type == PathType.RELATIVE:
        importer_dir = directory_name(importer)
        if importer_dir:
            joined = importer_dir + "/" + requested
        else:
            joined = requested
        return normalize_path(joined)
    else:
        # For sanity, treat absolute as just returning the requested string directly
        return requested

def load_module(root_directory: str, resolved_name: str) -> str | None:
    """
    Maps a resolved name to source strings or file paths.
    
    Algorithm:
    1. If Logical: Walk up directory tree looking for `wren_modules`.
       If found, lookup inside.
       If not found on disk, fallback to internal built-in modules.
    2. If File Path: Append `.wren`, read from disk.
    
    Missing modules return None (handled by the VM as compilation error).
    Readable failures (e.g. permission denied) exit with code 74.
    """
    path_type = classify_path(resolved_name)
    
    if path_type == PathType.SIMPLE:
        # Walk up tree to find wren_modules
        current_dir = os.path.abspath(root_dir) if root_dir else os.path.abspath(".")
        found_source = None
        
        while True:
            modules_dir = os.path.join(current_dir, "wren_modules")
            if os.path.isdir(modules_dir):
                # If module name lacks a slash (e.g. 'foo'), duplicate it to 'foo/foo'
                if "/" not in resolved_name:
                    module_subpath = f"{resolved_name}/{resolved_name}"
                else:
                    module_subpath = resolved_name
                    
                full_path = os.path.join(modules_dir, module_subpath + ".wren")
                if os.path.isfile(full_path):
                    try:
                        with open(full_path, "r") as f:
                            found_source = f.read()
                        break
                    except OSError:
                        print(f"Could not read file \"{full_path}\".", file=sys.stderr)
                        sys.exit(74)
            
            parent = os.path.dirname(current_dir)
            if parent == current_dir:  # hit root safely
                break
            current_dir = parent
            
        if found_source is not None:
            return found_source
            
        # Fallback to internal/built-in modules
        if resolved_name in BUILTIN_MODULES:
            return BUILTIN_MODULES[resolved_name]
            
        # Logical module not found at all
        return None

    else:
        # File Path resolution
        full_path = resolved_name + ".wren"
        if not os.path.isfile(full_path):
            return None
            
        try:
            with open(full_path, "r") as f:
                return f.read()
        except OSError:
            print(f"Could not read file \"{full_path}\".", file=sys.stderr)
            sys.exit(74)
