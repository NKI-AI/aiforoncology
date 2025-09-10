"""Path constants for the project."""

from pathlib import Path


PROJECT_ROOT = Path("projects/radiology/adaptive_mri_bounds")

# Config paths
CONFIG_DIR = PROJECT_ROOT / "configs"
CHANGING_PARAMETERS_PATH = CONFIG_DIR / "changing_parameters.toml"
PERFORMANCE_PARAMETERS_PATH = CONFIG_DIR / "performance_parameters.toml"
PLOT_PARAMETERS_PATH = CONFIG_DIR / "plot_parameters.toml"
