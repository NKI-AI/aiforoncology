# Copyright (c) Meta Platforms, Inc. and affiliates.
#
# This source code is licensed under the Apache License, Version 2.0
# found in the LICENSE file in the root directory of this source tree.

# References:
#   https://github.com/facebookresearch/dino/blob/master/vision_transformer.py
#   https://github.com/rwightman/pytorch-image-models/tree/master/timm/models/vision_transformer.py

import logging
import os
import warnings
from typing import Optional

import torch
from torch import nn

logger = logging.getLogger("dinov2")


XFORMERS_ENABLED = os.environ.get("XFORMERS_DISABLED") is None
try:
    if XFORMERS_ENABLED:
        from xformers.ops import memory_efficient_attention, unbind

        XFORMERS_AVAILABLE = True
        warnings.warn("xFormers is available (Attention)")
    else:
        warnings.warn("xFormers is disabled (Attention)")
        raise ImportError
except ImportError:
    XFORMERS_AVAILABLE = False
    warnings.warn("xFormers is not available (Attention)")


class Attention(nn.Module):
    """Multi-head self-attention module.

    Parameters
    ----------
    dim : int
        Dimension of the input features.
    num_heads : int, optional
        Number of attention heads, by default 8.
    qkv_bias : bool, optional
        Whether to add a bias to the query, key, and value projections, by default False.
    proj_bias : bool, optional
        Whether to add a bias to the output projection, by default True.
    attn_drop : float, optional
        Dropout rate for the attention weights, by default 0.0.
    proj_drop : float, optional
        Dropout rate for the output projection, by default 0.0.

    Raises
    ------
    ValueError
        If `dim` is not divisible by `num_heads`.
    """

    def __init__(
        self,
        dim: int,
        num_heads: int = 8,
        qkv_bias: bool = False,
        proj_bias: bool = True,
        attn_drop: float = 0.0,
        proj_drop: float = 0.0,
    ) -> None:
        """Inits :class:`Attention`.

        Parameters
        ----------
        dim : int
            Dimension of the input features.
        num_heads : int, optional
            Number of attention heads, by default 8.
        qkv_bias : bool, optional
            Whether to add a bias to the query, key, and value projections, by default False.
        proj_bias : bool, optional
            Whether to add a bias to the output projection, by default True.
        attn_drop : float, optional
            Dropout rate for the attention weights, by default 0.0.
        proj_drop : float, optional
            Dropout rate for the output projection, by default 0.0.

        Raises
        ------
        ValueError
            If `dim` is not divisible by `num_heads`.
        """
        super().__init__()
        if dim % num_heads != 0:
            raise ValueError(f"dim {dim} should be divisible by num_heads {num_heads}.")
        self.num_heads = num_heads
        head_dim = dim // num_heads
        self.scale = head_dim**-0.5

        self.qkv = nn.Linear(dim, dim * 3, bias=qkv_bias)
        self.attn_drop = nn.Dropout(attn_drop)
        self.proj = nn.Linear(dim, dim, bias=proj_bias)
        self.proj_drop = nn.Dropout(proj_drop)

    def forward(self, x: torch.Tensor) -> torch.Tensor:
        """Forward pass of :class:`Attention`.

        Parameters
        ----------
        x : torch.Tensor
            Input tensor of shape (B, N, C) where B is the batch size, N is the sequence length, and C is
            the feature dimension.

        Returns
        -------
        torch.Tensor
            Output tensor of shape (B, N, C) after applying multi-head self-attention.
        """
        B, N, C = x.shape
        qkv = self.qkv(x).reshape(B, N, 3, self.num_heads, C // self.num_heads).permute(2, 0, 3, 1, 4)

        q, k, v = qkv[0] * self.scale, qkv[1], qkv[2]
        attn = q @ k.transpose(-2, -1)

        attn = attn.softmax(dim=-1)
        attn = self.attn_drop(attn)

        x = (attn @ v).transpose(1, 2).reshape(B, N, C)
        x = self.proj(x)
        x = self.proj_drop(x)
        return x


class MemEffAttention(Attention):
    """Memory-efficient multi-head self-attention module using xFormers.

    Parameters
    ----------
    dim : int
        Dimension of the input features.
    num_heads : int, optional
        Number of attention heads, by default 8.
    qkv_bias : bool, optional
        Whether to add a bias to the query, key, and value projections, by default False.
    proj_bias : bool, optional
        Whether to add a bias to the output projection, by default True.
    attn_drop : float, optional
        Dropout rate for the attention weights, by default 0.0.
    proj_drop : float, optional
        Dropout rate for the output projection, by default 0.0.

    Raises
    ------
    ValueError
        If `dim` is not divisible by `num_heads`.
    """

    def forward(self, x: torch.Tensor, attn_bias: Optional[torch.Tensor] = None) -> torch.Tensor:
        """Forward pass of :class:`MemEffAttention`.

        Parameters
        ----------
        x : torch.Tensor
            Input tensor of shape (B, N, C) where B is the batch size, N is the sequence length, and C is
            the feature dimension.
        attn_bias : Optional[torch.Tensor], optional
            Attention bias tensor for memory-efficient attention, by default None.

        Raises
        ------
        AssertionError
            If xFormers is not available and `attn_bias` is provided.

        Returns
        -------
        torch.Tensor
            Output tensor of shape (B, N, C) after applying memory-efficient multi-head self-attention.
        """
        if not XFORMERS_AVAILABLE:
            if attn_bias is not None:
                raise AssertionError("xFormers is required for using nested tensors")
            return super().forward(x)

        B, N, C = x.shape
        qkv = self.qkv(x).reshape(B, N, 3, self.num_heads, C // self.num_heads)

        q, k, v = unbind(qkv, 2)

        x = memory_efficient_attention(q, k, v, attn_bias=attn_bias)
        x = x.reshape([B, N, C])

        x = self.proj(x)
        x = self.proj_drop(x)
        return x
