"""GoFre - Build system for Python packages with Go extensions."""

import os
import stat
import subprocess
import sys

__version__ = "0.1.0"


def get_binary_path():
    """Return the path to the bundled gofre binary."""
    pkg_dir = os.path.dirname(os.path.abspath(__file__))
    binary = os.path.join(pkg_dir, "bin", "gofre")
    if sys.platform == "win32":
        binary += ".exe"
    return binary


def main():
    """Execute the bundled gofre binary."""
    binary = get_binary_path()

    if not os.path.exists(binary):
        print(
            f"Error: gofre binary not found at {binary}\n"
            "This may happen if gofre was installed from source without building.\n"
            "Try: pip install gofre --force-reinstall",
            file=sys.stderr,
        )
        sys.exit(1)

    if sys.platform != "win32":
        current_mode = os.stat(binary).st_mode
        if not (current_mode & stat.S_IXUSR):
            os.chmod(binary, current_mode | stat.S_IRWXU | stat.S_IXGRP | stat.S_IXOTH)
        os.execvp(binary, [binary] + sys.argv[1:])
    else:
        sys.exit(subprocess.call([binary] + sys.argv[1:]))
