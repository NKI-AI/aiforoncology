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

import torch
from torch.nn.functional import pad as torch_pad
from torchvision import tv_tensors
from torchvision.transforms import functional as F
from typing import Optional
import numpy as np


def _calculate_padding_2d(patch_size: int | None, height: int, width: int) -> tuple[int, int, int, int]:
    """Calculate padding needed to make dimensions divisible by patch_size for 2D images.

    Padding is added to top and equally to both sides.

    Parameters
    ----------
    patch_size : int, optional
        The patch size to make dimensions divisible by.
    height : int
        Current height of the image
    width : int
        Current width of the image

    Returns
    -------
    tuple[int, int, int, int]
        Padding values (left, top, right, bottom) as expected by torchvision.transforms.functional.pad.
    """
    if patch_size is None:
        return (0, 0, 0, 0)

    # Calculate how much padding is needed for each dimension
    pad_height = (patch_size - (height % patch_size)) % patch_size
    pad_width = (patch_size - (width % patch_size)) % patch_size

    # Add padding to top and distribute width padding equally
    pad_left = pad_width // 2
    pad_right = pad_width - pad_left  # Handle odd padding
    pad_top = pad_height
    pad_bottom = 0

    return (pad_left, pad_top, pad_right, pad_bottom)


def _calculate_padding_3d(patch_size: int | None, depth: int, height: int, width: int) -> tuple[int, ...]:
    """Calculate padding needed to make dimensions divisible by patch_size for 3D images.

    Padding is added to depth to front and distributed width padding equally.

    Parameters
    ----------
    patch_size : int, optional
        The patch size to make dimensions divisible by.
    depth : int
        Current depth of the image
    height : int
        Current height of the image
    width : int
        Current width of the image

    Returns
    -------
    tuple[int, ...]
        Padding values (left, right, top, bottom, front, back) as expected by torch.nn.functional.pad.
    """
    if patch_size is None:
        return (0,) * 6
    pad_depth = (patch_size - (depth % patch_size)) % patch_size
    pad_height = (patch_size - height % patch_size) % patch_size
    pad_width = (patch_size - width % patch_size) % patch_size

    # We use torch.nn.functional.pad instead of torchvision.transforms.functional.pad because the latter
    # does not support 3D padding.
    # It expects (pad_left, pad_right, pad_top, pad_bottom, pad_front, pad_back)

    pad_left, pad_right = pad_width // 2, pad_width - (pad_width // 2)
    pad_top, pad_bottom = pad_height, 0
    pad_front, pad_back = 0, pad_depth

    return (pad_left, pad_right, pad_top, pad_bottom, pad_front, pad_back)


class PatchDivisiblePadder(torch.nn.Module):
    """
    A transform that pads an image to make its dimensions divisible by a patch size.
    Padding is added to the top and equally to both sides.

    Can handle both single input (image) and dual input (image, target) scenarios.
    Supports both 2D (C, H, W) and 3D (C, D, H, W) tensors.
    """

    def __init__(self, patch_size: int, fill: float = 0.0) -> None:
        """
        Initialize the PatchDivisiblePadder.

        Parameters
        ----------
        patch_size : int
            The patch size to make dimensions divisible by.
        fill : float, optional
            The value to use for padding. Default is 0.0.
        """
        super().__init__()
        self.patch_size = patch_size
        self.fill = fill

    def forward(
        self, img: tv_tensors.Image, target: Optional[tv_tensors.Mask] = None
    ) -> tuple[tv_tensors.Image, tv_tensors.Mask | None]:
        """
        Apply padding to make dimensions divisible by patch_size.
        Return type adapts based on input - returns just the image if target is None,
        or returns (image, target) if target is provided.

        Parameters
        ----------
        img : Image
            Input image tensor of shape (C, H, W) or (C, D, H, W)
        target : Mask | None, optional
            Optional segmentation mask to pad using the same amounts

        Returns
        -------
        Image or Tuple[Image, Mask]
            Padded image (if target is None) or padded image and mask (if target is provided)
        """
        # Get spatial dimensions
        if img.ndim == 3:  # 2D image (C, H, W)
            height, width = img.shape[-2:]
            padding = _calculate_padding_2d(self.patch_size, height, width)
            if any(padding):  # Only pad if necessary
                img = F.pad(img, padding, fill=self.fill)
                if target is not None:
                    target = F.pad(target, padding, fill=0)  # Always use 0 for mask padding
        elif img.ndim == 4:  # 3D image (C, D, H, W)
            depth, height, width = img.shape[-3:]
            padding = _calculate_padding_3d(self.patch_size, depth, height, width)
            if any(padding):  # Only pad if necessary
                img = torch_pad(img, padding, mode="constant", value=self.fill)
                if target is not None:
                    target = torch_pad(target, padding, mode="constant", value=0)  # Always use 0 for mask padding
        else:
            raise ValueError(f"Expected 2D (C, H, W) or 3D (C, D, H, W) tensor, got shape {img.shape}")

        # Return type adapts based on inputs
        if target is None:
            return img
        else:
            return img, target

    def __repr__(self) -> str:
        return f"{self.__class__.__name__}(patch_size={self.patch_size}, fill={self.fill})"


class VolumeBoundingBoxCropper:
    """
    A transform that crops a 3D volume to the bounding box of pixels above a given lower percentile.
    Applied at the volume level before slicing.
    """

    def __init__(self, lower_percentile: float = 80.0) -> None:
        """
        Initialize the VolumeBoundingBoxCropper.

        Parameters
        ----------
        lower_percentile : float
            The percentile below which pixels are considered background. Default is 80.0.
        """
        self.lower_percentile = lower_percentile

    def __call__(
        self, volume: np.ndarray, segmentation: np.ndarray | None = None
    ) -> tuple[np.ndarray, np.ndarray | None]:
        """
        Apply the cropping transform to the input volume and optionally its segmentation.

        Parameters
        ----------
        volume : np.ndarray
            Input volume of shape (D, H, W)
        segmentation : np.ndarray | None
            Optional segmentation volume to crop using the same bounds

        Returns
        -------
        Tuple[np.ndarray, np.ndarray | None]
            Cropped volume and optionally cropped segmentation
        """
        # Calculate the pixel intensity threshold based on the lower percentile
        threshold = np.percentile(volume, self.lower_percentile)

        # Find the indices of pixels above the threshold
        intensity_mask = volume > threshold
        if not intensity_mask.any():
            return volume, segmentation

        # Get the bounds of non-zero elements for each dimension
        nonzero = np.where(intensity_mask)

        # For a (D, H, W) volume:
        # nonzero[0] is depth indices
        # nonzero[1] is height indices
        # nonzero[2] is width indices

        # Only crop in height and width, preserve depth
        min_h, max_h = np.min(nonzero[1]), np.max(nonzero[1])
        min_w, max_w = np.min(nonzero[2]), np.max(nonzero[2])

        # Crop the volume, keeping all depth slices
        cropped_volume = volume[:, min_h : max_h + 1, min_w : max_w + 1]

        # Crop the segmentation if provided
        cropped_segmentation = None
        if segmentation is not None:
            cropped_segmentation = segmentation[:, min_h : max_h + 1, min_w : max_w + 1]

        if cropped_volume.shape != (volume.shape[0], max_h - min_h + 1, max_w - min_w + 1):
            raise ValueError(
                f"Unexpected shape after cropping: got {cropped_volume.shape}, "
                f"expected ({volume.shape[0]}, {max_h - min_h + 1}, {max_w - min_w + 1})"
            )

        return cropped_volume, cropped_segmentation

    def __repr__(self) -> str:
        return f"{self.__class__.__name__}(lower_percentile={self.lower_percentile})"
