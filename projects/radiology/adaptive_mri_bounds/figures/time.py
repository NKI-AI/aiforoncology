import matplotlib.pyplot as plt
from matplotlib.ticker import MultipleLocator, FuncFormatter
import numpy as np
from utils.parameter_loader import load_parameters
from utils.paths import PLOT_PARAMETERS_PATH


def standard_time_figure(data, changing_param, param_dict, performance_params):
    """
    Create a figure showing average protocol times for different protocols.

    Parameters
    ----------
    data : dict
        Dictionary containing the data to plot with keys:
        'avg_time_ai_best_case', 'avg_time_ai_worst_case', 'avg_time_abbr',
        'avg_time_full', and the changing parameter name
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
    The figure shows average protocol times for Adaptive (with best/worst case range),
    Abbreviated, and Full protocols. The plot includes a shaded region between
    best and worst case scenarios for the Adaptive protocol and uses a cut-off y-axis
    to better visualize the differences.
    """
    # load params
    params = load_parameters(PLOT_PARAMETERS_PATH)
    params["percentage_formatter"] = FuncFormatter(lambda x, pos: f"{x:.0%}")

    # Create figure
    fig, ax = plt.subplots(1, 1, figsize=(12, 6))

    # Recall rate plot
    (line_adaptive,) = ax.plot(
        data[changing_param],
        data["avg_time_ai_best_case"],
        label="Adaptive",
        linewidth=params["linewidth"],
        color=params["colors"]["ai"],
    )
    (line_adaptive2,) = ax.plot(
        data[changing_param],
        data["avg_time_ai_worst_case"],
        label="Adaptive2",
        linewidth=params["linewidth"],
        color=params["colors"]["ai"],
    )
    (line_abbr,) = ax.plot(
        data[changing_param],
        data["avg_time_abbr"],
        label="Abbreviated",
        linewidth=params["linewidth"],
        color=params["colors"]["abbr"],
    )
    (line_full,) = ax.plot(
        data[changing_param],
        data["avg_time_full"],
        label="Full",
        linewidth=params["linewidth"],
        color=params["colors"]["full"],
    )

    # Fill between the best and worst cases for 'Adaptive'
    ax.fill_between(
        data[changing_param],
        data["avg_time_ai_best_case"],
        data["avg_time_ai_worst_case"],
        color=params["colors"]["ai"],
        alpha=params["colors"]["ai_alpha"],
    )

    # Labels and legend
    ax.set_xlabel(param_dict["name"], fontsize=params["label_size"])
    ax.set_ylabel("Average time (s)", fontsize=params["label_size"])
    ax.legend(
        handles=[line_abbr, line_full, line_adaptive],
        labels=["Abbreviated", "Full", "Adaptive"],
        loc="lower left",
        fontsize=params["legend_size"],
    )

    # Axis and grid
    ax.xaxis.set_major_formatter(params["percentage_formatter"])
    ax.margins(x=0)
    ax.set_ylim(0, params["time_ylim"])
    ax.tick_params(axis="both", labelsize=params["tick_size"])
    ax.minorticks_on()
    ax.xaxis.set_major_locator(MultipleLocator(0.2))
    ax.xaxis.set_minor_locator(MultipleLocator(0.1))
    ax.yaxis.set_major_locator(MultipleLocator(100))
    ax.yaxis.set_minor_locator(MultipleLocator(50))
    ax.grid(which="major", linestyle="-", linewidth=0.7)
    ax.grid(which="minor", linestyle=":", linewidth=0.5)

    # Cut of line y-axis
    false_zero = 300
    ylim = [false_zero, 850]
    fz_factor = 10
    yticks = np.arange(false_zero // 100 * 100 + 100, ylim[1] // 100 * 100 + 100, 100)
    ax.set_ylim(ylim[0], ylim[1])
    ax.set_yticks(np.append(false_zero, yticks))
    ax.set_yticklabels(np.append(0, yticks))
    ax.spines["left"].set_visible(False)
    if false_zero > 0:
        ax.add_line(
            plt.Line2D(
                ydata=[
                    false_zero,
                    false_zero + 1 * fz_factor,
                    false_zero + 2 * fz_factor,
                    false_zero + 3 * fz_factor,
                    false_zero + 4 * fz_factor,
                    ylim[1],
                ],
                xdata=[0, 0, 0.015, -0.015, 0, 0],
                color=ax.spines["left"].get_edgecolor(),
                lw=ax.spines["left"].get_linewidth(),
                clip_on=False,
                transform=ax.get_yaxis_transform(),
            )
        )

    return {"fig": fig, "name": f"{changing_param}_time.{params['format']}"}


def time_diff_time_figure(data, changing_param, param_dict, performance_params):
    """
    Create a figure showing relative protocol times as percentages of full protocol duration.

    Parameters
    ----------
    data : dict
        Dictionary containing the data to plot with keys:
        'avg_time_ai_best_case', 'avg_time_ai_worst_case', 'avg_time_abbr',
        'avg_time_full', and the changing parameter name
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
    The figure shows protocol times normalized as percentages of the full protocol duration.
    Times for Adaptive (with best/worst case range), Abbreviated, and Full protocols are shown.
    The plot includes a shaded region between best and worst case scenarios for the Adaptive
    protocol and uses a cut-off y-axis starting at 15% to better visualize the differences.
    """
    # load params
    params = load_parameters(PLOT_PARAMETERS_PATH)
    params["percentage_formatter"] = FuncFormatter(lambda x, pos: f"{x:.0%}")

    # Change to percentage
    data["avg_time_ai_best_case"] = [x / y for x, y in zip(data["avg_time_ai_best_case"], data["avg_time_full"])]
    data["avg_time_ai_worst_case"] = [x / y for x, y in zip(data["avg_time_ai_worst_case"], data["avg_time_full"])]
    data["avg_time_abbr"] = [x / y for x, y in zip(data["avg_time_abbr"], data["avg_time_full"])]
    data["avg_time_full"] = [x / y for x, y in zip(data["avg_time_full"], data["avg_time_full"])]

    # Create figure
    fig, ax = plt.subplots(1, 1, figsize=(12, 6))

    # Recall rate plot
    (line_adaptive,) = ax.plot(
        data[changing_param],
        data["avg_time_ai_best_case"],
        label="Adaptive",
        linewidth=params["linewidth"],
        color=params["colors"]["ai"],
    )
    (line_adaptive2,) = ax.plot(
        data[changing_param],
        data["avg_time_ai_worst_case"],
        label="Adaptive2",
        linewidth=params["linewidth"],
        color=params["colors"]["ai"],
    )
    (line_abbr,) = ax.plot(
        data[changing_param],
        data["avg_time_abbr"],
        label="Abbreviated",
        linewidth=params["linewidth"],
        color=params["colors"]["abbr"],
    )
    (line_full,) = ax.plot(
        data[changing_param],
        data["avg_time_full"],
        label="Full",
        linewidth=params["linewidth"],
        color=params["colors"]["full"],
    )

    # Fill between the best and worst cases for 'Adaptive'
    ax.fill_between(
        data[changing_param],
        data["avg_time_ai_best_case"],
        data["avg_time_ai_worst_case"],
        color=params["colors"]["ai"],
        alpha=params["colors"]["ai_alpha"],
    )

    # Labels and legend
    ax.set_xlabel(f"{param_dict['name']} (s)", fontsize=params["label_size"])
    ax.set_ylabel("Average time (% of full protocol duration)", fontsize=params["label_size"])
    ax.legend(
        handles=[line_abbr, line_full, line_adaptive],
        labels=["Abbreviated", "Full", "Adaptive"],
        loc="lower left",
        fontsize=params["legend_size"],
    )

    # Axis and grid
    ax.yaxis.set_major_formatter(params["percentage_formatter"])
    ax.margins(x=0)
    ax.set_ylim(0, 1.05)
    ax.tick_params(axis="both", labelsize=params["tick_size"])
    ax.minorticks_on()
    ax.xaxis.set_major_locator(MultipleLocator(100))
    ax.xaxis.set_minor_locator(MultipleLocator(25))
    ax.yaxis.set_major_locator(MultipleLocator(0.1))
    ax.yaxis.set_minor_locator(MultipleLocator(0.05))
    ax.grid(which="major", linestyle="-", linewidth=0.7)
    ax.grid(which="minor", linestyle=":", linewidth=0.5)

    # Cut of line y-axis
    false_zero = 0.15  # Custom baseline
    ylim = [false_zero, 1.05]  # y-axis limits
    fz_factor = 0.01  # Factor for offset visualization
    yticks = np.arange(0.2, ylim[1], 0.1)  # Ticks above the false zero

    # Apply y-axis limits and ticks
    ax.set_ylim(ylim[0], ylim[1])
    ax.set_yticks(np.concatenate(([false_zero], yticks)))  # Add false_zero as the first tick
    ax.set_yticklabels([0] + [f"{ytick:.1f}" for ytick in yticks])  # Adjust labels with "0" for false_zero

    # Hide the left spine
    ax.spines["left"].set_visible(False)

    # Add a cut-off line on the y-axis to indicate a false zero
    if false_zero > 0:
        ax.add_line(
            plt.Line2D(
                ydata=[
                    false_zero,
                    false_zero + 1 * fz_factor,
                    false_zero + 2 * fz_factor,
                    false_zero + 3 * fz_factor,
                    false_zero + 4 * fz_factor,
                    ylim[1],
                ],
                xdata=[0, 0, 0.015, -0.015, 0, 0],
                color=ax.spines["left"].get_edgecolor(),
                lw=ax.spines["left"].get_linewidth(),
                clip_on=False,
                transform=ax.get_yaxis_transform(),
            )
        )

    return {"fig": fig, "name": f"{changing_param}_time.{params['format']}"}
