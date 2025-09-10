import numpy as np
from pathlib import Path
from calculations.process_parameters import process_parameter_set
from utils.parameter_loader import load_parameters
from utils.paths import PERFORMANCE_PARAMETERS_PATH
from utils.save_figure import save_and_close_figure
from figures.recall import standard_recall_figure
from figures.time import standard_time_figure, time_diff_time_figure


def create_figure(changing_param, param_dict, save_dir):
    """
    Create figures for recall rate and average protocol time based on changing parameters.

    Parameters
    ----------
    changing_param : str
        Name of the parameter being varied in the analysis
    param_dict : dict
        Dictionary containing parameter range information with keys:
        'parameter_range': dict with 'start', 'end', and 'step' values
    save_dir : Path or str
        Directory where the generated figures will be saved

    Returns
    -------
    None
        Figures are saved to the specified directory

    Notes
    -----
    The function creates different types of figures based on the changing parameter:
    - For time-related parameters ('full_time', 'abbr_time'): creates time difference plots
    - For other parameters: creates standard recall and time plots
    """
    # Import standard performance parameters
    performance_params = load_parameters(PERFORMANCE_PARAMETERS_PATH)

    # Calculate bounds for recall rate and average protocol time at every step
    param_range = param_dict["parameter_range"]
    data = {
        "recall_ai_best_case": [],
        "recall_ai_worst_case": [],
        "recall_abbr": [],
        "recall_full": [],
        "avg_time_ai_best_case": [],
        "avg_time_ai_worst_case": [],
        "avg_time_abbr": [],
        "avg_time_full": [],
        changing_param: [],
    }
    for value in np.linspace(param_range["start"], param_range["end"], param_range["step"]):
        # Update the changing parameter in parameter set
        performance_params[changing_param] = value
        # Calculate the recall and time for parameter set
        new_data = process_parameter_set(**performance_params)
        # Append the results to the data dict
        for key in data.keys():
            if key == changing_param:
                data[key].append(value)
            else:
                data[key].append(new_data[key])

    # Create figures
    fig_list = []
    if changing_param in ["full_time", "abbr_time"]:
        fig_list.append(time_diff_time_figure(data, changing_param, param_dict, performance_params))
    else:
        fig_list.append(standard_recall_figure(data, changing_param, param_dict, performance_params))
        fig_list.append(standard_time_figure(data, changing_param, param_dict, performance_params))

    for fig_dict in fig_list:
        fig = fig_dict["fig"]
        name = fig_dict["name"]
        save_and_close_figure(fig=fig, save_name=name, save_dir=save_dir)
