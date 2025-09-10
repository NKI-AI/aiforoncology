# Copyright (c) Meta Platforms, Inc. and affiliates.
#
# This source code is licensed under the Apache License, Version 2.0
# found in the LICENSE file in the root directory of this source tree.

# Modified from: https://github.com/huggingface/pytorch-image-models/blob/main/timm/models/vision_transformer.py#L103-L110

from typing import Union

import torch
from torch import nn


class LayerScale(nn.Module):
    """Layer scale module for scaling the output of a layer.

    Parameters
    ----------
    dim : int
        Dimension of the layer scale.
    init_values : float or torch.Tensor, optional
        Initial values for the layer scale, by default 1e-5. If a tensor is provided, it should have shape (dim,).
    inplace : bool, optional
        Whether to perform the operation in-place, by default False.
    """

    def __init__(
        self,
        dim: int,
        init_values: Union[float, torch.Tensor] = 1e-5,
        inplace: bool = False,
    ) -> None:
        """Inits :class:`LayerScale

        Parameters
        ----------
        dim : int
            Dimension of the layer scale.
        init_values : float or torch.Tensor, optional
            Initial values for the layer scale, by default 1e-5. If a tensor is provided, it should have shape (dim,).
        inplace : bool, optional
            Whether to perform the operation in-place, by default False.
        """
        super().__init__()

        self.inplace = inplace
        self.gamma = nn.Parameter(init_values * torch.ones(dim))

    def forward(self, x: torch.Tensor) -> torch.Tensor:
        """Forward pass of :class:`LayerScale`.

        Parameters
        ----------
        x : torch.Tensor
            Input tensor of shape (B, N, C) where B is the batch size, N is the sequence length, and C is
            the feature dimension.

        Returns
        -------
        torch.Tensor
            Scaled output tensor of shape (B, N, C).
        """
        return x.mul_(self.gamma) if self.inplace else x * self.gamma
