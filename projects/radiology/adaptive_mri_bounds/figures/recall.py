import matplotlib.pyplot as plt
from matplotlib.ticker import MultipleLocator, FuncFormatter
from utils.parameter_loader import load_parameters
from utils.paths import PLOT_PARAMETERS_PATH


def standard_recall_figure(data, changing_param, param_dict, performance_params):
    """
    Create a figure showing recall rates for different protocols.

    Parameters
    ----------
    data : dict
        Dictionary containing the data to plot with keys:
        'recall_ai_best_case', 'recall_ai_worst_case', 'recall_abbr',
        'recall_full', and the changing parameter name
    changing_param : str
        Name of the parameter being varied in the analysis
    param_dict : dict
        Dictionary containing parameter information including 'name' for axis label
    performance_params : dict
        Dictionary containing performance parameters (not directly used in plotting)

    Returns
    -------
    dict
        Dictionary containing:
        'fig' : matplotlib.figure.Figure
            The generated figure object
        'name' : str
            The filename for saving the figure

    Notes
    -----
    The figure shows recall rates for Adaptive (with best/worst case range),
    Abbreviated, and Full protocols. The plot includes a shaded region between
    best and worst case scenarios for the Adaptive protocol.
    """
    # load params
    params = load_parameters(PLOT_PARAMETERS_PATH)
    params["percentage_formatter"] = FuncFormatter(lambda x, pos: f"{x:.0%}")

    # Create figure
    fig, ax = plt.subplots(1, 1, figsize=(12, 6))

    # Recall rate plot
    (line_adaptive,) = ax.plot(
        data[changing_param],
        data["recall_ai_best_case"],
        label="Adaptive",
        linewidth=params["linewidth"],
        color=params["colors"]["ai"],
    )
    (line_adaptive2,) = ax.plot(
        data[changing_param],
        data["recall_ai_worst_case"],
        label="Adaptive2",
        linewidth=params["linewidth"],
        color=params["colors"]["ai"],
    )
    (line_abbr,) = ax.plot(
        data[changing_param],
        data["recall_abbr"],
        label="Abbreviated",
        linewidth=params["linewidth"],
        color=params["colors"]["abbr"],
    )
    (line_full,) = ax.plot(
        data[changing_param],
        [x + y for x, y in zip(data["recall_full"], [0.0002] * len(data["recall_full"]))],
        label="Full",
        linewidth=params["linewidth"],
        color=params["colors"]["full"],
    )

    # Fill between the best and worst cases for 'Adaptive'
    ax.fill_between(
        data[changing_param],
        data["recall_ai_best_case"],
        data["recall_ai_worst_case"],
        color=params["colors"]["ai"],
        alpha=params["colors"]["ai_alpha"],
    )

    # Labels and legend
    ax.set_xlabel(param_dict["name"], fontsize=params["label_size"])
    ax.set_ylabel("Recall Rate (%)", fontsize=params["label_size"])
    ax.legend(
        handles=[line_abbr, line_full, line_adaptive],
        labels=["Abbreviated", "Full", "Adaptive"],
        loc="lower left",
        fontsize=params["legend_size"],
    )

    # Axis and grid
    ax.xaxis.set_major_formatter(params["percentage_formatter"])
    ax.yaxis.set_major_formatter(params["percentage_formatter"])
    ax.margins(x=0)
    ax.set_ylim(0, params["recall_ylim"])
    ax.tick_params(axis="both", labelsize=params["tick_size"])
    ax.minorticks_on()
    ax.xaxis.set_major_locator(MultipleLocator(0.2))
    ax.xaxis.set_minor_locator(MultipleLocator(0.1))
    ax.yaxis.set_major_locator(MultipleLocator(0.05))
    ax.yaxis.set_minor_locator(MultipleLocator(0.025))
    ax.grid(which="major", linestyle="-", linewidth=0.7)
    ax.grid(which="minor", linestyle=":", linewidth=0.5)

    return {"fig": fig, "name": f"{changing_param}_recall.{params['format']}"}
