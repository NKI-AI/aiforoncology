import argparse
from pathlib import Path
from utils.parameter_loader import load_parameters
from utils.paths import CHANGING_PARAMETERS_PATH
from figures.create_figures import create_figure


def main(save_dir):
    """
    Calculate all bounds for recall rate and average protocol time, and create all plots.

    Parameters
    ----------
    save_dir : str or Path
        Directory path where the generated figures will be saved.

    Returns
    -------
    None
        Figures are saved to the specified directory.

    Notes
    -----
    This function performs the following steps:
    1. Creates the save directory if it doesn't exist
    2. Loads parameters from the changing_parameters.toml configuration file
    3. Generates and saves figures for each changing parameter
    """
    # Ensure save directory exists
    save_path = Path(save_dir)
    save_path.mkdir(parents=True, exist_ok=True)

    # Load parameters for changing parameters
    changing_params = load_parameters(CHANGING_PARAMETERS_PATH)

    # Create and save figures for each changing parameter
    for changing_param, param_dict in changing_params.items():
        create_figure(changing_param, param_dict, save_path)

    print(f"All figures generated and saved to {save_path}.")


if __name__ == "__main__":
    # Parse command line arguments inside the __main__ block
    parser = argparse.ArgumentParser(description="Generate and save figures.")
    parser.add_argument("--save_dir", type=str, required=True, help="Directory to save the figures")
    args = parser.parse_args()

    # Call the main function with the parsed save directory
    main(args.save_dir)
