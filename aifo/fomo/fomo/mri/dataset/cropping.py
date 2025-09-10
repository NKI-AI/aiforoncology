# Copyright 2025 AI for Oncology Research Group. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import matplotlib.pyplot as plt
import numpy as np


def crop_to_bbox_percentile(slice_data: np.ndarray, lower_percentile: float = 1.0) -> np.ndarray:
    """
    Crop a slice to the bounding box of pixels above a given lower percentile.

    Parameters
    ----------
    slice_data : np.ndarray
        The 2D slice data to be cropped.
    lower_percentile : float, optional
        The percentile below which pixels are considered 'black'. Default is 1.0 (i.e., the bottom 1%).

    Returns
    -------
    np.ndarray
        Cropped slice data within the bounding box of non-near-zero pixels.
    """
    # Calculate the pixel intensity threshold based on the lower percentile
    threshold = np.percentile(slice_data, lower_percentile)

    # Find the indices of pixels above the threshold
    non_zero_coords = np.argwhere(slice_data > threshold)

    if non_zero_coords.size == 0:
        # If all pixels are below the threshold, return the original slice
        return slice_data

    # Get the min and max bounds for the non-zero coordinates
    min_row, min_col = non_zero_coords.min(axis=0)
    max_row, max_col = non_zero_coords.max(axis=0)

    # Crop the slice based on the bounding box
    cropped_slice = slice_data[min_row : max_row + 1, min_col : max_col + 1]

    return cropped_slice


def plot_before_after_crop(original_slice: np.ndarray, cropped_slice: np.ndarray) -> None:
    """
    Plot the original slice and cropped slice side by side for comparison.

    Parameters
    ----------
    original_slice : np.ndarray
        The original 2D slice before cropping.
    cropped_slice : np.ndarray
        The cropped 2D slice after removing black borders.
    """
    fig, axs = plt.subplots(1, 2, figsize=(10, 5))

    # Plot original slice
    axs[0].imshow(original_slice, cmap="gray")
    axs[0].set_title("Original Slice")
    axs[0].axis("off")

    # Plot cropped slice
    axs[1].imshow(cropped_slice, cmap="gray")
    axs[1].set_title("Cropped Slice")
    axs[1].axis("off")

    plt.tight_layout()
    plt.savefig("cropcompare.png")
