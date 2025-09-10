# Copyright (c) Meta Platforms, Inc. and affiliates.
#
# This source code is licensed under the Apache License, Version 2.0
# found in the LICENSE file in the root directory of this source tree.

import os
import warnings
from typing import Callable, Optional

import torch
import torch.nn.functional as F
from torch import nn


class SwiGLUFFN(nn.Module):
    r"""SwiGLU Feed-Forward Network (FFN) layer.

    SwiGLU Feed-Forward Network (FFN) layer.

    This module applies a two-layer position-wise feed-forward transformation with a SwiGLU activation: 
    a gated unit combining the SiLU nonlinearity with an elementwise multiplication.

    Given input tensor ``x`` of shape ``(B, d)``, the computation is:

    .. math::

        [z_1, z_2] = x W_{12} + b_{12} \\\\
        h = \mathrm{SiLU}(z_1) \odot z_2 \\\\
        y = h W_3 + b_3

    where:
        - :math:`W_{12} \in \mathbb{R}^{d \times 2h}`, :math:`b_{12} \in \mathbb{R}^{2h}`
        - :math:`W_3 \in \mathbb{R}^{h \times d_{\text{out}}}`, :math:`b_3 \in \mathbb{R}^{d_{\text{out}}}`
        - :math:`\mathrm{SiLU}(x) = x \cdot \sigma(x)` is the Sigmoid Linear Unit
        - :math:`\odot` denotes elementwise multiplication

    Parameters
    ----------
    in_features : int
        Input feature dimensionality (d).
    hidden_features : int, optional
        Hidden layer dimensionality (h). Defaults to in_features.
    out_features : int, optional
        Output feature dimensionality (d_out). Defaults to in_features.
    act_layer : Callable[..., nn.Module], optional
        Unused. Included for compatibility.
    drop : float, optional
        Dropout rate (unused).
    bias : bool, optional
        Whether to include bias terms in linear layers.
    """

    def __init__(
        self,
        in_features: int,
        hidden_features: Optional[int] = None,
        out_features: Optional[int] = None,
        act_layer: Callable[..., nn.Module] = None,
        drop: float = 0.0,
        bias: bool = True,
    ) -> None:
        """Inits :class:`SwiGLUFFN`.

        Parameters
        ----------
        in_features : int
            Input feature dimensionality (d).
        hidden_features : int, optional
            Hidden layer dimensionality (h). Defaults to in_features.
        out_features : int, optional
            Output feature dimensionality (d_out). Defaults to in_features.
        act_layer : Callable[..., nn.Module], optional
            Unused. Included for compatibility.
        drop : float, optional
            Dropout rate (unused).
        bias : bool, optional
            Whether to include bias terms in linear layers.
        """
        super().__init__()
        out_features = out_features or in_features
        hidden_features = hidden_features or in_features
        self.w12 = nn.Linear(in_features, 2 * hidden_features, bias=bias)
        self.w3 = nn.Linear(hidden_features, out_features, bias=bias)

    def forward(self, x: torch.Tensor) -> torch.Tensor:
        """Forward pass of :class:`SwiGLUFFN`.

        Parameters
        ----------
        x : torch.Tensor
            Input tensor of shape (B, N, C) where B is the batch size, N is the sequence length, and C is
            the input feature dimension.

        Returns
        -------
        torch.Tensor
            Output tensor of shape (B, N, out_features) after applying the SwiGLU feed-forward network.
        """
        x12 = self.w12(x)
        x1, x2 = x12.chunk(2, dim=-1)
        hidden = F.silu(x1) * x2
        return self.w3(hidden)


XFORMERS_ENABLED = os.environ.get("XFORMERS_DISABLED") is None
try:
    if XFORMERS_ENABLED:
        from xformers.ops import SwiGLU

        XFORMERS_AVAILABLE = True
        warnings.warn("xFormers is available (SwiGLU)")
    else:
        warnings.warn("xFormers is disabled (SwiGLU)")
        raise ImportError
except ImportError:
    SwiGLU = SwiGLUFFN
    XFORMERS_AVAILABLE = False

    warnings.warn("xFormers is not available (SwiGLU)")


class SwiGLUFFNFused(SwiGLU):
    """Fused SwiGLU Feed-Forward Network (FFN) layer.

    Fused SwiGLU Feed-Forward Network (FFN) layer that uses xFormers' fused implementation if available.
    This layer combines the linear transformations and activation into a single operation for improved performance.

    Parameters
    ----------
    in_features : int
        Input feature dimensionality (d).
    hidden_features : int, optional
        Hidden layer dimensionality (h). Defaults to in_features.
    out_features : int, optional
        Output feature dimensionality (d_out). Defaults to in_features.
    act_layer : Callable[..., nn.Module], optional
        Unused. Included for compatibility.
    drop : float, optional
        Dropout rate (unused).
    bias : bool, optional
        Whether to include bias terms in linear layers.
    """

    def __init__(
        self,
        in_features: int,
        hidden_features: Optional[int] = None,
        out_features: Optional[int] = None,
        act_layer: Callable[..., nn.Module] = None,
        drop: float = 0.0,
        bias: bool = True,
    ) -> None:
        """Inits :class:`SwiGLUFFNF

        Parameters
        ----------
        in_features : int
            Input feature dimensionality (d).
        hidden_features : int, optional
            Hidden layer dimensionality (h). Defaults to in_features.
        out_features : int, optional
            Output feature dimensionality (d_out). Defaults to in_features.
        act_layer : Callable[..., nn.Module], optional
            Unused. Included for compatibility.
        drop : float, optional
            Dropout rate (unused).
        bias : bool, optional
            Whether to include bias terms in linear layers.
        """
        out_features = out_features or in_features
        hidden_features = hidden_features or in_features
        hidden_features = (int(hidden_features * 2 / 3) + 7) // 8 * 8
        super().__init__(
            in_features=in_features,
            hidden_features=hidden_features,
            out_features=out_features,
            bias=bias,
        )
