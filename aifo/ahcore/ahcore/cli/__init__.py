"""Ahcore Command-line interface. This is the file which builds the main parser."""

import argparse
import os
import pathlib
from typing import Callable


def dir_path(require_writable: bool = False) -> Callable[[str], pathlib.Path]:
    def check_dir_path(path: str) -> pathlib.Path:
        """Check if the path is a valid and (optionally) writable directory.

        Parameters
        ----------
        path : str

        Returns
        -------
        pathlib.Path
            The path as a pathlib.Path object.
        """
        _path = pathlib.Path(path)
        if _path.is_dir():
            if require_writable:
                if os.access(_path, os.W_OK):
                    return _path
                else:
                    raise argparse.ArgumentTypeError(f"{path} is not a writable directory.")
            else:
                return _path
        raise argparse.ArgumentTypeError(f"{path} is not a valid directory.")

    return check_dir_path


def file_path(path: str) -> pathlib.Path:
    """Check if the path is a valid file.

    Parameters
    ----------
    path : str

    Returns
    -------
    pathlib.Path
        The path as a pathlib.Path object.

    """
    _path = pathlib.Path(path)
    if _path.is_file():
        return _path
    raise argparse.ArgumentTypeError(f"{path} is not a valid file.")
