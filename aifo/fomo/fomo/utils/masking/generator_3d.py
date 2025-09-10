# Copyright (c) Meta Platforms, Inc. and affiliates.
#
# This source code is licensed under the Apache License, Version 2.0
# found in the LICENSE file in the root directory of dinov2 in third_party.
#
# Modified by NKI-AI4Oncology, 04-2025
import math
import random
from torch import Tensor
from fomo.utils.masking.base import MaskingGenerator


class MaskingGenerator3d(MaskingGenerator):
    # number of retries to find an appropriate random mask
    _RETRY_RANGE = 10

    def __init__(
        self,
        input_size: int | tuple[int],
        min_num_patches: int,
        max_num_patches: int,
        min_aspect_height_width: int = 0.3,
        max_aspect_height_width: int | None = None,
        min_aspect_height_depth: int = 0.3,
        max_aspect_height_depth: int | None = 0.4,
    ) -> None:
        super().__init__(input_size, min_num_patches, max_num_patches, min_aspect_height_width, max_aspect_height_width)
        self._input_depth, self._input_height, self._input_width = (
            self._input_size if isinstance(input_size, tuple) else (self._input_size,) * 3
        )

        self._min_aspect_height_depth = min_aspect_height_depth
        self._max_aspect_height_depth = (
            max_aspect_height_depth if max_aspect_height_depth else 1 / min_aspect_height_depth
        )

        self._log_aspect_ratio_height_depth = (
            math.log(self._min_aspect_height_depth),
            math.log(self._max_aspect_height_depth),
        )

    def _get_shape(self) -> tuple[int, int, int]:
        return self._input_depth, self._input_height, self._input_width

    def _mask(self, mask: Tensor, max_mask_patches: int) -> int:
        """
        Apply masking to the input 3D tensor.

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
        selected regions to True (masked). It creates 3D cuboid masks
        with configurable aspect ratios for height/width and height/depth.

        Masking method adapted from the 2D version and extended to 3D.
        """
        delta = 0
        for _ in range(self._RETRY_RANGE):
            # TODO: Keeping this consistent with the implementation of dinov2 for now, but this does not
            # make sense. max_mask_patches can become smaller than self._min_num_patches, because
            # max_mask_patches is decremented in __call__ after an iteration of _mask returns a new masked
            # block of patches. There is no active constraint on the number of minimum patches, which
            # means the last/last few blocks that are masked are smaller than self._min_num_patches.
            target_volume = random.uniform(self._min_num_patches, max_mask_patches)
            aspect_ratio_hw = math.exp(random.uniform(*self._log_aspect_ratio_height_width))
            aspect_ratio_hd = math.exp(random.uniform(*self._log_aspect_ratio_height_depth))

            mask_height = int(round((target_volume / (aspect_ratio_hd * aspect_ratio_hw)) ** (1 / 3)))
            mask_depth = int(round(aspect_ratio_hd * mask_height))
            mask_width = int(round(aspect_ratio_hw * mask_height))

            # NOTE: typically, mask_width and mask_height are much smaller than the input width and height.
            # However, for depth we'll initially only have a few levels of patches meaning that it's possible
            # to take the entirety of depth-level patches as mask.
            if (
                mask_width < self._input_width
                and mask_height < self._input_height
                and 0 < mask_depth <= self._input_depth
            ):
                front = random.randint(0, self._input_depth - mask_depth)
                top = random.randint(0, self._input_height - mask_height)
                left = random.randint(0, self._input_width - mask_width)

                num_masked = mask[front : front + mask_depth, top : top + mask_height, left : left + mask_width].sum()

                # Account for overlap
                if 0 < mask_height * mask_width * mask_depth - num_masked <= max_mask_patches:
                    # Create a view of the region to be masked
                    mask_region = mask[front : front + mask_depth, top : top + mask_height, left : left + mask_width]
                    # Find positions that are currently 0
                    zero_positions = mask_region == 0
                    # Count how many positions will be changed
                    delta = zero_positions.sum().item()
                    # Set those positions to 1
                    mask_region[zero_positions] = True

                if delta > 0:
                    break

        return delta
