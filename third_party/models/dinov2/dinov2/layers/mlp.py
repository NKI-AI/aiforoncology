# Copyright (c) Meta Platforms, Inc. and affiliates.
#
# This source code is licensed under the Apache License, Version 2.0
# found in the LICENSE file in the root directory of this source tree.

# References:
#   https://github.com/facebookresearch/dino/blob/master/vision_transformer.py
#   https://github.com/rwightman/pytorch-image-models/tree/master/timm/layers/mlp.py


from typing import Callable, Optional

import torch
from torch import nn


class Mlp(nn.Module):
    """Multi-layer perceptron (MLP) module.

    Creates a simple MLP with two linear layers and an activation function in between and dropout after each layer.

    Parameters
    ----------
    in_features : int
        Number of input features.
    hidden_features : int, optional
        Number of hidden features, by default 4 * in_features.
    out_features : int, optional
        Number of output features, by default in_features.
    act_layer : Callable[..., nn.Module], optional
        Activation layer, by default nn.GELU.
    drop : float, optional
        Dropout rate, by default 0.0.
    bias : bool, optional
        Whether to use bias in the linear layers, by default True.
    """

    def __init__(
        self,
        in_features: int,
        hidden_features: Optional[int] = None,
        out_features: Optional[int] = None,
        act_layer: Callable[..., nn.Module] = nn.GELU,
        drop: float = 0.0,
        bias: bool = True,
    ) -> None:
        """Inits :class:`Mlp`.

        Parameters
        ----------

        in_features : int
            Number of input features.
        hidden_features : int, optional
            Number of hidden features, by default 4 * in_features.
        out_features : int, optional
            Number of output features, by default in_features.
        act_layer : Callable[..., nn.Module], optional
            Activation layer, by default nn.GELU.
        drop : float, optional
            Dropout rate, by default 0.0.
        bias : bool, optional
            Whether to use bias in the linear layers, by default True.
        """
        super().__init__()

        out_features = out_features or in_features
        hidden_features = hidden_features or in_features

        self.fc1 = nn.Linear(in_features, hidden_features, bias=bias)
        self.act = act_layer()
        self.fc2 = nn.Linear(hidden_features, out_features, bias=bias)
        self.drop = nn.Dropout(drop)

    def forward(self, x: torch.Tensor) -> torch.Tensor:
        """Forward pass of :class:`Mlp`.

        Parameters
        ----------
        x : torch.Tensor
            Input tensor of shape (B, N, C) where B is the batch size, N is the sequence length, and C is
            the feature dimension.

        Returns
        -------
        torch.Tensor
            Output tensor of shape (B, N, out_features) after applying the MLP.
        """
        x = self.fc1(x)
        x = self.act(x)
        x = self.drop(x)
        x = self.fc2(x)
        x = self.drop(x)
        return x
