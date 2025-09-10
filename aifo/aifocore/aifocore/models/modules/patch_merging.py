# Copyright (c) MONAI Consortium
# Copyright 2025 AI for Oncology Research Group. All Rights Reserved.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#     http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
import itertools
import logging

import torch
import torch.nn as nn
import torch.nn.functional as F
from torch.nn import LayerNorm


class PatchMerging(nn.Module):
    """Patch merging layer based on [1].

    Parameters
    ----------
    dim : int
        Number of feature channels.
    norm_layer : type[LayerNorm], optional
        Normalization layer, by default nn.LayerNorm
    spatial_dims : int, optional
        Number of spatial dims, by default 3

    References
    ----------
    [1] Liu et al., Swin Transformer: Hierarchical Vision Transformer using Shifted Windows
        <https://arxiv.org/abs/2103.14030>
        https://github.com/microsoft/Swin-Transformer
    """

    def __init__(self, dim: int, norm_layer: type[LayerNorm] = nn.LayerNorm, spatial_dims: int = 3) -> None:
        """Inits :class:`PatchMerging`

        Parameters
        ----------
        dim : int
            Number of feature channels.
        norm_layer : type[LayerNorm], optional
            Normalization layer, by default nn.LayerNorm
        spatial_dims : int, optional
            Number of spatial dims, by default 3
        """

        super().__init__()
        self.reduction = nn.Linear(2**spatial_dims * dim, 2 * dim, bias=False)
        self.norm = norm_layer(2**spatial_dims * dim)

    def forward(self, x: torch.Tensor) -> torch.Tensor:
        """Forward pass of patch merging layer.

        Expects input to be in (B, D1, ..., Dn, C) shape. Merges patches based on their proximity. Divides dimensions
        by 2 by concatenating feature maps of adjacent patches; channel will increase by 2^{num_dimensions} before
        reduction, 2 * num_dimensions after reduction.

        Parameters
        ----------
        x : torch.Tensor
            Input tensor in shape (B, D1, ..., Dn, C)

        Returns
        -------
        torch.Tensor
            Merged patches tensor
        """
        _, *spatial_dims, _ = x.shape

        should_pad_input = any([dim % 2 == 1 for dim in spatial_dims])
        if should_pad_input:
            logging.warning("Detected input of shape '%s', required to pad before merging.", x.shape)
            padding = []
            # pad each uneven dimension with 1
            for dim in reversed(spatial_dims):
                padding.extend([0, dim % 2])
            x = F.pad(x, (0, 0, *padding))

        # This merges adjacent patches by creating a grid of overlapping slices:
        # For 2D example with 4x4 input:
        # slice(0,None,2) gives positions [0,2]
        # slice(1,None,2) gives positions [1,3]
        # Combining these creates all adjacent pairs:
        # [(0,0),(0,1),(1,0),(1,1)] -> captures 2x2 neighborhoods
        # Each neighborhood's features are concatenated, effectively merging
        # local 2x2 (2D) or 2x2x2 (3D) regions into single tokens
        # Output is (B, D1/2, ..., Dn/2, 2^len(spatial_dims) * C)
        slice_ranges = list(itertools.product(range(2), repeat=len(spatial_dims)))
        slice_combinations = [[slice(index, None, 2) for index in slice_range] for slice_range in slice_ranges]
        x = torch.cat([x[:, *slice_combination, :] for slice_combination in slice_combinations], dim=-1)

        x = self.norm(x)
        x = self.reduction(x)
        return x
