"""Load parameters from TOML configuration files."""

import toml
from pathlib import Path


def load_parameters(filepath):
    """
    Load parameters from a TOML configuration file.

    Parameters
    ----------
    filepath : str or Path
        Path to the TOML file

    Returns
    -------
    dict
        Dictionary containing the parameters loaded from the TOML file
    """
    config_path = Path(filepath)
    with open(config_path, "r") as file:
        return toml.load(file)
