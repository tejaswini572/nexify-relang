import sys
import os
from scheduler import Scheduler
from modules import load_module
from io_ops import IoModule

class InterpreterHost:
    """
    Virtual Machine Lifecycle & Script Execution.
    Defines how the runtime bootstraps, evaluates code, and reports output/errors.
    """
    def __init__(self):
        self.scheduler = Scheduler()
        self.io = IoModule(self.scheduler)
        self.root_directory = ""
        self.exit_code = 0

    def print_output(self, text: str):
        """For normal output, emit the string to STDOUT."""
        sys.stdout.write(text)

    def print_error(self, err_type: str, module: str, line: int, message: str):
        """
        For errors, format based on error type and write to STDERR.
        """
        if err_type == "RUNTIME":
            sys.stderr.write(f"{message}\n")
        else:
            sys.stderr.write(f"[{module} line {line}] {message}\n")

    def run_file(self, script_path: str):
        """
        Runs a script from a file.
        Terminates the process upon fatal OS read errors or uncaught script crashes.
        """
        try:
            with open(script_path, "r") as f:
                source = f.read()
        except OSError:
            sys.stderr.write(f"Could not read file \"{script_path}\".\n")
            sys.exit(66)

        # Normalize the file path (stripping extensions and resolving dots)
        script_val = os.path.normpath(script_path)
        # Determine script's directory, cache globally as root for logical imports
        self.root_directory = os.path.dirname(script_val)
        if not self.root_directory:
            self.root_directory = "."
            
        def _main_fiber():
            try:
                from lexer import Lexer
                from parser import Parser
                from interpreter import Interpreter
                from tokens import TokenType
                
                lexer = Lexer(source)
                tokens = lexer.scan_tokens()
                
                parser = Parser(tokens)
                statements = parser.parse()
                
                interpreter = Interpreter(self)
                for stmt in statements:
                    interpreter.execute(stmt)
            except Exception as e:
                # Catch parse and native runtime errors for the host loop
                sys.stderr.write(str(e) + "\n")
                sys.exit(70)

        # Add initial fiber to queue
        self.scheduler.add(_main_fiber)

        # Block on event loop until empty
        try:
            self.scheduler.run_loop()
        except SystemExit as e:
            self.exit_code = e.code

        # Cleanup triggers
        sys.exit(self.exit_code)

    def run_repl(self):
        """
        Runs the standard REPL format evaluating interactive lines.
        """
        # Set root directory to current
        self.root_directory = "."

        print("\\\\/\"-")
        print(" \\_/   wren")

        def _repl_fiber():
            # synthetic import "repl" executed here in evaluation hook
            pass

        self.scheduler.add(_repl_fiber)

        try:
            self.scheduler.run_loop()
        except SystemExit as e:
            self.exit_code = e.code

        sys.exit(self.exit_code)


if __name__ == "__main__":
    host = InterpreterHost()
    if len(sys.argv) > 1:
        host.run_file(sys.argv[1])
    else:
        host.run_repl()
