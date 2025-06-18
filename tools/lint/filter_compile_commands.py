"""Filter compile_commands.json to only include specified files.

This script takes a list of changed files and filters the compile_commands.json
to only include entries for those files. This is useful for running clang-tidy
or other tools only on changed files.

It also patches the directory field to be the GitHub workspace, which is
necessary for clang-tidy to find the files.

Example usage:
    python filter_compile_commands.py --input compile_commands.json --changed-files file1.cc file2.h
"""
import os
import argparse
import json
import pathlib
from typing import Sequence


def filter_compile_commands(changed_files: Sequence[str], input_path: pathlib.Path) -> None:
    """Filter and patch compile_commands.json to only include specified files."""
    changed_files_abs = {str(pathlib.Path(f).resolve()) for f in changed_files}

    with open(input_path, "r", encoding="utf-8") as f:
        compile_commands = json.load(f)

    github_workspace = os.environ.get("GITHUB_WORKSPACE", "/github/workspace")

    filtered_commands = []
    for cmd in compile_commands:
        file_abs = str(pathlib.Path(cmd["file"]).resolve())
        if file_abs in changed_files_abs:
            cmd["directory"] = github_workspace  # 🔧 patch the directory
            filtered_commands.append(cmd)

    with open(input_path, "w", encoding="utf-8") as f:
        json.dump(filtered_commands, f)



def main() -> None:
    parser = argparse.ArgumentParser(description="Filter compile_commands.json to only include specified files.")
    parser.add_argument(
        "--changed-files",
        nargs="+",
        required=True,
        help="List of changed files to include in the filtered output",
    )
    parser.add_argument(
        "--input",
        type=pathlib.Path,
        required=True,
        help="Path to the compile_commands.json file that will be overwritten",
    )

    args = parser.parse_args()
    filter_compile_commands(args.changed_files, args.input)


if __name__ == "__main__":
    main()
