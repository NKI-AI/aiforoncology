#!/usr/bin/env python3
"""Bazel workspace status program.

Invoke bazel with --workspace_status_command=bazel/build-version.py
to get this invoked and populate bazel-out/volatile-status.txt
"""

import os

from subprocess import Popen, PIPE


def run(*cmd):
    process = Popen(cmd, stdout=PIPE)
    output, _ = process.communicate()
    return output.strip().decode()


def main():
    try:
        date = run("git", "log", "-n1", "--date=short", "--format=%H")
    except:
        date = ""

    try:
        version = run("git", "describe", "--abbrev=0", "--tags")
    except:
        version = ""

    if not date:
        try:
            date = os.environ["GIT_DATE"]
        except:
            date = "<unknown-git-date>"

    if not version:
        try:
            version = os.environ["GIT_VERSION"]
        except:
            version = "<unknown-git-version>"

    print("GIT_DATE", '"{}"'.format(date))
    print("GIT_DESCRIBE", '"{}"'.format(version))


if __name__ == "__main__":
    main()
