"""This fixes the problem `{ should almost always be at the end of the previous line` cpplint generates after clang reformat"""

import re
import sys
from pathlib import Path


def fix_requires_brace(file_path: str):
    path = Path(file_path)
    text = path.read_text()

    # Regex: Find any line ending with "requires something" and a newline + {
    # Works for both "requires(...)" and "requires concept"
    fixed_text = re.sub(
        r"^(.*\brequires\b[^\n]*)\n\s*{",
        r"\1 {",
        text,
        flags=re.MULTILINE,
    )

    if fixed_text != text:
        path.write_text(fixed_text)
        print(f"Fixed: {file_path}")
    else:
        print(f"No changes needed: {file_path}")


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python fix_requires_brace.py <file1> [<file2> ...]")
        sys.exit(1)

    for file in sys.argv[1:]:
        fix_requires_brace(file)
