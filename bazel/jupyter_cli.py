#!/usr/bin/env python3
"""
Jupyter CLI wrapper for Bazel that supports multiple subcommands.
"""

import sys
import subprocess
from typing import List


def main() -> None:
    """Main entry point for the Jupyter CLI wrapper."""
    # Get command line arguments (excluding the script name)
    args = sys.argv[1:]

    if not args:
        # No arguments provided, show help
        show_help()
        return

    subcommand = args[0]

    # Handle different subcommands
    if subcommand == "lab":
        launch_jupyter_lab(args[1:])
    elif subcommand == "notebook":
        launch_jupyter_notebook(args[1:])
    elif subcommand == "console":
        launch_jupyter_console(args[1:])
    elif subcommand in ["--help", "-h", "help"]:
        show_help()
    elif subcommand == "--version":
        show_version()
    else:
        print(f"Unknown subcommand: {subcommand}")
        show_help()
        sys.exit(1)


def launch_jupyter_lab(args: List[str]) -> None:
    """Launch JupyterLab with the given arguments."""
    try:
        # Add --core-mode by default to use pre-built assets
        # This avoids the need for building JavaScript assets in Bazel
        if "--core-mode" not in args and "--dev-mode" not in args:
            args = ["--core-mode"] + args

        # Use subprocess to launch JupyterLab module
        cmd = [sys.executable, "-m", "jupyterlab"] + args
        result = subprocess.run(cmd, check=False)
        sys.exit(result.returncode)

    except FileNotFoundError:
        print("Error: JupyterLab not available")
        sys.exit(1)
    except Exception as e:
        print(f"Error launching JupyterLab: {e}")
        sys.exit(1)


def launch_jupyter_notebook(args: List[str]) -> None:
    """Launch Jupyter Notebook with the given arguments."""
    try:
        # Use subprocess to launch Jupyter Notebook module
        cmd = [sys.executable, "-m", "notebook"] + args
        result = subprocess.run(cmd, check=False)
        sys.exit(result.returncode)

    except FileNotFoundError:
        print("Error: Jupyter Notebook not available")
        sys.exit(1)
    except Exception as e:
        print(f"Error launching Jupyter Notebook: {e}")
        sys.exit(1)


def launch_jupyter_console(args: List[str]) -> None:
    """Launch Jupyter Console with the given arguments."""
    try:
        # Use subprocess to launch Jupyter Console module
        cmd = [sys.executable, "-m", "jupyter_console"] + args
        result = subprocess.run(cmd, check=False)
        sys.exit(result.returncode)

    except FileNotFoundError:
        print("Error: Jupyter Console not available")
        sys.exit(1)
    except Exception as e:
        print(f"Error launching Jupyter Console: {e}")
        sys.exit(1)


def show_help() -> None:
    """Show help information."""
    print("""
Jupyter CLI Wrapper for Bazel

Usage: jupyter <subcommand> [options]

Available subcommands:
  lab         Launch JupyterLab
  notebook    Launch Jupyter Notebook
  console     Launch Jupyter Console
  --help, -h  Show this help message
  --version   Show version information

Examples:
  jupyter lab                    # Launch JupyterLab
  jupyter lab --port=8889        # Launch JupyterLab on port 8889
  jupyter notebook               # Launch Jupyter Notebook
  jupyter console                # Launch Jupyter Console

For subcommand-specific help, use:
  jupyter <subcommand> --help
""")


def show_version() -> None:
    """Show version information."""
    try:
        import jupyter_core

        print(f"jupyter-core: {jupyter_core.__version__}")
    except ImportError:
        print("jupyter-core: not available")

    try:
        import jupyterlab

        print(f"jupyterlab: {jupyterlab.__version__}")
    except ImportError:
        print("jupyterlab: not available")

    try:
        import notebook

        print(f"notebook: {notebook.__version__}")
    except ImportError:
        print("notebook: not available")


if __name__ == "__main__":
    main()
