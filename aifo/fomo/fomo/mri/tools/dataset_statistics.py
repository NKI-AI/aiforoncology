import logging
import sys
from pathlib import Path
from typing import List, Tuple, Union

import numpy as np
import SimpleITK as sitk
from fomo.utils.io import find_files_by_extension

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


def load_volume(file_path: Path) -> np.ndarray:
    image = sitk.ReadImage(str(file_path))
    array = sitk.GetArrayViewFromImage(image)
    return array


def extract_middle_slices(volume: np.ndarray, spacing: float, cm_range: float = 75.0) -> np.ndarray:
    center_slice = volume.shape[0] // 2
    slices = int(cm_range / spacing)
    start_slice = max(0, center_slice - slices)
    end_slice = min(volume.shape[0], center_slice + slices)
    return volume[start_slice:end_slice, :, :]


def remove_highest_percentile(volume: np.ndarray, percentile: float = 99.9) -> np.ndarray:
    threshold = np.percentile(volume, percentile)
    volume_clipped = np.clip(volume, None, threshold)
    return volume_clipped


def minmax_normalize(volume: np.ndarray) -> np.ndarray:
    min_val = np.min(volume)
    max_val = np.max(volume)
    diff = max_val - min_val
    if diff == 0:
        logger.warning("Found max and min equal, aborting normalize")
        return volume
    return (volume - min_val) / (max_val - min_val)


def update_mean_std(
    current_mean: float, current_var: float, current_count: int, new_data: np.ndarray
) -> Tuple[float, float, int]:
    n = current_count
    new_count = n + new_data.size
    new_mean = current_mean + (np.sum(new_data) - current_mean * new_data.size) / new_count
    new_var = current_var + np.sum((new_data - new_mean) * (new_data - current_mean))

    return new_mean, new_var, new_count


def process_directory_incrementally(directory: Path, extensions: Union[str, List[str]]) -> Tuple[float, float]:
    files = find_files_by_extension(directory, extensions)
    mean = 0.0
    var = 0.0
    count = 0

    for file_path in files:
        volume = load_volume(file_path)
        spacing = get_spacing(file_path)
        slices = extract_middle_slices(volume, spacing)
        slices = remove_highest_percentile(slices)
        slices = minmax_normalize(slices)

        mean, var, count = update_mean_std(mean, var, count, slices)

    std = np.sqrt(var / count)
    logger.info(f"Overall mean: {mean}, Overall std: {std}")
    return mean, std


def get_spacing(file_path: Path) -> float:
    image = sitk.ReadImage(str(file_path))
    spacing = image.GetSpacing()[0]
    return spacing


# Example usage
if __name__ == "__main__":
    if not sys.argv[1]:
        sys.exit("Please provide a path to a dataset.")
    directory = Path(sys.argv[1])
    extensions = [".nrrd", ".nii"]
    mean, std = process_directory_incrementally(directory, extensions)
