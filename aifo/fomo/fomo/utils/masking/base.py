# Copyright (c) Meta Platforms, Inc. and affiliates.
#
# This source code is licensed under the Apache License, Version 2.0
# found in the LICENSE file in the root directory of dinov2 in third_party.
#
# Modified by NKI-AI4Oncology, 04-2025
import math
import torch
from abc import ABC, abstractmethod


class MaskingGenerator(ABC):
    """
    Abstract base class for generating masks for vision transformer patches.

    This class creates boolean masks for transformer patches, allowing for
    rectangular or square masked regions with configurable sizes and aspect ratios.

    Parameters
    ----------
    input_size : int or list of int
        Size of the input grid. If int, assumes a square grid. If list, specifies dimensions.
    min_num_patches : int
        Minimum number of patches to mask in a single mask region.
    max_num_patches : int
        Maximum number of patches to mask in a single mask region.
    min_aspect_height_width : float, default=0.3
        Minimum aspect ratio (height/width) for the masked regions.
        Values < 1 allow for wider-than-tall rectangles.
    max_aspect_height_width : float, optional
        Maximum aspect ratio (height/width) for the masked regions.
        If None, it's set to the reciprocal of min_aspect_height_width,
        allowing for taller-than-wide rectangles.

    Notes
    -----
    The aspect ratios are log-transformed for uniform sampling between
    the minimum and maximum values.
    """

    def __init__(
        self,
        input_size: int | tuple[int],
        min_num_patches: int,
        max_num_patches: int,
        min_aspect_height_width: int = 0.3,
        max_aspect_height_width: int | None = None,
    ) -> None:
        self._input_size = input_size
        self._min_num_patches = min_num_patches
        self._max_num_patches = max_num_patches
        self._min_aspect_height_width = min_aspect_height_width
        self._max_aspect_height_width = (
            max_aspect_height_width if max_aspect_height_width else 1 / min_aspect_height_width
        )

        # take the log of the aspect ratios for fair uniform sampling
        self._log_aspect_ratio_height_width = (
            math.log(self._min_aspect_height_width),
            math.log(self._max_aspect_height_width),
        )

    @abstractmethod
    def _get_shape(self) -> tuple[int]:
        """
        Get the shape of the mask based on input size.

        Returns
        -------
        tuple[int]
            The shape of the mask as a list of integers.
            For 2D masks, returns [height, width].
            For 3D masks, returns [depth, height, width].

        Notes
        -----
        This method should interpret self._input_size and convert it
        to the appropriate shape list. If self._input_size is an integer,
        it typically represents a square grid. If it's a list, it directly
        specifies the dimensions.
        """
        pass

    @abstractmethod
    def _mask(self, mask: torch.Tensor, max_mask_patches: int) -> int:
        """
        Apply masking to the input tensor.

        Parameters
        ----------
        mask : torch.Tensor
            Boolean tensor to be modified with new masked regions.
        max_mask_patches : int
            Maximum number of patches that can be masked in this call.

        Returns
        -------
        int
            The number of patches that were actually masked.

        Notes
        -----
        This method modifies the input mask tensor in-place by setting
        selected regions to True (masked).
        """
        pass

    def __call__(self, num_masking_patches: int):
        mask = torch.zeros(self._get_shape(), dtype=bool)
        mask_count = 0
        while mask_count < num_masking_patches:
            current_max_mask_patches = num_masking_patches - mask_count
            current_max_mask_patches = min(current_max_mask_patches, self._max_num_patches)

            delta = self._mask(mask, current_max_mask_patches)
            if delta == 0:
                break
            else:
                mask_count += delta

        return mask
