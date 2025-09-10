# Copyright (c) Meta Platforms, Inc. and affiliates.
#
# This source code is licensed under the Apache License, Version 2.0
# found in the LICENSE file in the root directory of this source tree.

# References:
#   https://github.com/facebookresearch/dino/blob/master/vision_transformer.py
#   https://github.com/rwightman/pytorch-image-models/tree/master/timm/layers/drop.py


import torch
from torch import nn


def drop_path(x: torch.Tensor, drop_prob: float = 0.0, training: bool = False) -> torch.Tensor:
    """Drop paths (Stochastic Depth) per sample (when applied in main path of residual blocks).

    Parameters
    ----------
    x : torch.Tensor
        Input tensor of shape (B, *) where B is the batch size and * is any number of additional dimensions.
    drop_prob : float, optional
        Probability of dropping a path, by default 0.0
    training : bool, optional
        Whether the model is in training mode, by default False. If False, no paths are dropped.

    Returns
    -------
    torch.Tensor
        Output tensor with the same shape as input x, with paths dropped according to drop_prob.
    """
    if drop_prob == 0.0 or not training:
        return x
    keep_prob = 1 - drop_prob
    shape = (x.shape[0],) + (1,) * (x.ndim - 1)  # work with diff dim tensors, not just 2D ConvNets
    random_tensor = x.new_empty(shape).bernoulli_(keep_prob)
    if keep_prob > 0.0:
        random_tensor.div_(keep_prob)
    output = x * random_tensor
    return output


class DropPath(nn.Module):
    """Drop paths (Stochastic Depth) per sample (when applied in main path of residual blocks).

    Parameters
    ----------
    drop_prob : float, optional
        Probability of dropping a path, by default None. If None, no paths are dropped.
        If set to 0.0, it behaves like an identity function.
    """

    def __init__(self, drop_prob: float = 0.0) -> None:
        """Inits :class:`DropPath`.

        Parameters
        ----------
        drop_prob : float, optional
            Probability of dropping a path, by default 0.0. If None, no paths are dropped.
            If set to 0.0, it behaves like an identity function.
        """
        super().__init__()
        self.drop_prob = drop_prob

    def forward(self, x: torch.Tensor) -> torch.Tensor:
        """Forward pass of :class:`DropPath`.

        Parameters
        ----------
        x : torch.Tensor
            Input tensor of shape (B, *) where B is the batch size and * is any number of additional dimensions.

        Returns
        -------
        torch.Tensor
            Output tensor with the same shape as input x, with paths dropped according to drop_prob.
        """
        return drop_path(x, self.drop_prob, self.training)
