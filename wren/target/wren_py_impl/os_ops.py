import os
import sys

# Cache the starting arguments matching `osSetArguments` behaviour
_PROCESS_ARGS = sys.argv.copy()

def get_process_arguments() -> list[str]:
    """Caches and exposes the command line arguments provided upon shell execution."""
    return _PROCESS_ARGS

def get_home_path() -> str:
    """Requests the underlying OS location of the active executing User's personal storage directory."""
    return os.path.expanduser("~")

def get_cwd() -> str:
    """Returns the absolute path where the console shell was stationed when the executable was launched."""
    return os.getcwd()

def get_pid() -> int:
    """Returns the numeric OS process identifiers for the interpreter process."""
    return os.getpid()

def get_ppid() -> int:
    """Returns the parent process ID."""
    return os.getppid()

def get_platform_name() -> str:
    """Probes compiler macros/runtime details to return a static string."""
    platform = sys.platform
    if platform == "win32":
        return "Windows"
    elif platform == "darwin":
        return "OS X"
    elif platform.startswith("linux"):
        return "Linux"
    elif os.name == "posix":
        return "POSIX"
    return "Unknown"

def is_posix() -> bool:
    """Computes a strict boolean value on whether standard unix-like architectures are natively supported."""
    return os.name == "posix"

def get_version() -> str:
    """Returns the baked-in version string of the interpreter."""
    return "0.4.0"
