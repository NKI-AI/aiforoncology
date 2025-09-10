import matplotlib.pyplot as plt
from pathlib import Path
from utils.paths import PROJECT_ROOT


def save_and_close_figure(fig, save_name, save_dir=PROJECT_ROOT):
    """
    Save the given matplotlib figure to the specified path and close it.

    Parameters
    ----------
    fig : matplotlib.figure.Figure
        The matplotlib figure to be saved
    save_name : str
        The name of the file to save the figure as
    save_dir : Path or str, optional
        The directory where the figure will be saved,
        by default PROJECT_ROOT

    Returns
    -------
    None
        The figure is saved to disk and closed
    """
    save_path = Path(save_dir) / save_name
    save_path.parent.mkdir(parents=True, exist_ok=True)  # Ensure directory exists
    fig.savefig(str(save_path))  # Convert to string for matplotlib compatibility
    plt.close(fig)
