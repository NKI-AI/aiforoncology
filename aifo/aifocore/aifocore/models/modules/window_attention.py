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
from abc import ABC
from typing import Optional, Sequence

import torch
import torch.nn as nn
from torch.nn.init import trunc_normal_


class WindowAttentionBase(nn.Module, ABC):
    """Window based multi-head self attention module with relative position bias.

    Based on [1].

    This class implements window-based self-attention with relative position bias,
    serving as a base class for both 2D and 3D window attention implementations.

    Parameters
    ----------
    dim : int
        Number of feature channels.
    num_heads : int
        Number of attention heads.
    window_size : sequence of int
        Local window size.
    relative_coords : torch.Tensor
        Relative coordinate tensor for computing relative position indices.
    relative_position_bias_table : nn.Parameter
        Learnable parameter table for relative position bias.
    qkv_bias : bool, optional
        Whether to add a learnable bias to query, key, value, by default False
    attn_drop : float, optional
        Attention dropout rate, by default 0.0
    proj_drop : float, optional
        Dropout rate of output, by default 0.0

    References
    ----------
    [1] Liu et al., Swin Transformer: Hierarchical Vision Transformer using Shifted Windows
        <https://arxiv.org/abs/2103.14030>
        https://github.com/microsoft/Swin-Transformer
    """

    def __init__(
        self,
        dim: int,
        num_heads: int,
        window_size: Sequence[int],
        relative_coords: torch.Tensor,
        relative_position_bias_table: nn.Parameter,
        qkv_bias: bool = False,
        attn_drop: float = 0.0,
        proj_drop: float = 0.0,
    ) -> None:
        """Inits :class:`WindowAttentionBase`.

        Parameters
        ----------
        dim : int
            Number of feature channels.
        num_heads : int
            Number of attention heads.
        window_size : sequence of int
            Local window size.
        relative_coords : torch.Tensor
            Relative coordinate tensor for computing relative position indices.
        relative_position_bias_table : nn.Parameter
            Learnable parameter table for relative position bias.
        qkv_bias : bool, optional
            Whether to add a learnable bias to query, key, value, by default False
        attn_drop : float, optional
            Attention dropout rate, by default 0.0
        proj_drop : float, optional
            Dropout rate of output, by default 0.0
        """
        super().__init__()
        self._dim = dim
        self._window_size = window_size
        self._num_heads = num_heads
        head_dim = dim // num_heads
        self._scale = head_dim**-0.5

        relative_position_index = relative_coords.sum(-1)
        self.relative_position_index: torch.Tensor
        self.register_buffer("relative_position_index", relative_position_index)
        self.qkv = nn.Linear(dim, dim * 3, bias=qkv_bias)
        self.attn_drop = nn.Dropout(attn_drop)
        self.proj = nn.Linear(dim, dim)
        self.proj_drop = nn.Dropout(proj_drop)
        self.relative_position_bias_table = relative_position_bias_table
        trunc_normal_(self.relative_position_bias_table, std=0.02)
        self.softmax = nn.Softmax(dim=-1)

    def forward(self, x: torch.Tensor, mask: Optional[torch.Tensor] = None) -> torch.Tensor:
        """Forward pass of window attention.

        Computes self-attention within each window using query, key, value projections.
        Includes relative position bias and optional attention mask.

        Parameters
        ----------
        x : torch.Tensor
            Input tensor of shape (B, N, C) where:
                B is batch size
                N is number of patches in window
                C is channel dimension
        mask : Optional[torch.Tensor], optional
            Attention mask tensor, by default None. If provided, should be of shape
            (nW, N, N) where:
                nW is number of windows
                N is number of patches in window

        Returns
        -------
        torch.Tensor
            Output tensor of shape (B, N, C)
        """
        B, N, C = x.shape
        qkv = self.qkv.forward(x).reshape(B, N, 3, self._num_heads, C // self._num_heads).permute(2, 0, 3, 1, 4)
        Q, K, V = qkv[0], qkv[1], qkv[2]
        Q = Q * self._scale
        attn: torch.Tensor = Q @ K.transpose(-2, -1)
        relative_position_bias = self.relative_position_bias_table[
            self.relative_position_index.clone()[:N, :N].reshape(-1)
        ].reshape(N, N, -1)
        relative_position_bias = relative_position_bias.permute(2, 0, 1).contiguous()
        attn = attn + relative_position_bias.unsqueeze(0)
        if mask is not None:
            nw = mask.shape[0]
            attn = attn.view(B // nw, nw, self._num_heads, N, N) + mask.unsqueeze(1).unsqueeze(0)
            attn = attn.view(-1, self._num_heads, N, N)
            attn = self.softmax(attn)
        else:
            attn = self.softmax(attn)

        attn = self.attn_drop.forward(attn).to(V.dtype)
        x = (attn @ V).transpose(1, 2).reshape(B, N, C)
        x = self.proj(x)
        x = self.proj_drop(x)
        return x


class WindowAttention2d(WindowAttentionBase):
    """2D Window based multi-head self attention module with relative position bias.

    Based on [1].

    This class implements window-based self-attention for 2D inputs, extending the base
    WindowAttentionBase class with 2D-specific relative position bias computation.

    Parameters
    ----------
    dim : int
        Number of feature channels.
    num_heads : int
        Number of attention heads.
    window_size : sequence of int
        Local window size for 2D (height, width).
    qkv_bias : bool, optional
        Whether to add a learnable bias to query, key, value, by default False
    attn_drop : float, optional
        Attention dropout rate, by default 0.0
    proj_drop : float, optional
        Dropout rate of output, by default 0.0

    References
    ----------
    [1] Liu et al., Swin Transformer: Hierarchical Vision Transformer using Shifted Windows
        <https://arxiv.org/abs/2103.14030>
        https://github.com/microsoft/Swin-Transformer
    """

    def __init__(
        self,
        dim: int,
        num_heads: int,
        window_size: Sequence[int],
        qkv_bias: bool = False,
        attn_drop: float = 0,
        proj_drop: float = 0,
    ) -> None:
        """Inits :class:`WindowAttention2d`.

        Parameters
        ----------
        dim : int
            Number of feature channels.
        num_heads : int
            Number of attention heads.
        window_size : sequence of int
            Local window size for 2D (height, width).
        qkv_bias : bool, optional
            Whether to add a learnable bias to query, key, value, by default False
        attn_drop : float, optional
            Attention dropout rate, by default 0.0
        proj_drop : float, optional
            Dropout rate of output, by default 0.0
        """
        relative_position_bias_table = nn.Parameter(
            torch.zeros((2 * window_size[0] - 1) * (2 * window_size[1] - 1), num_heads)
        )
        coords_h = torch.arange(window_size[0])
        coords_w = torch.arange(window_size[1])
        coords = torch.stack(torch.meshgrid(coords_h, coords_w, indexing="ij"))
        coords_flatten = torch.flatten(coords, 1)
        relative_coords = coords_flatten[:, :, None] - coords_flatten[:, None, :]
        relative_coords = relative_coords.permute(1, 2, 0).contiguous()
        relative_coords[:, :, 0] += window_size[0] - 1
        relative_coords[:, :, 1] += window_size[1] - 1
        relative_coords[:, :, 0] *= 2 * window_size[1] - 1
        super().__init__(
            dim, num_heads, window_size, relative_coords, relative_position_bias_table, qkv_bias, attn_drop, proj_drop
        )


class WindowAttention3d(WindowAttentionBase):
    """3D Window based multi-head self attention module with relative position bias.

    Based on [1].

    This class implements window-based self-attention for 3D inputs, extending the base
    WindowAttentionBase class with 3D-specific relative position bias computation.

    Parameters
    ----------
    dim : int
        Number of feature channels.
    num_heads : int
        Number of attention heads.
    window_size : sequence of int
        Local window size for 3D (depth, height, width).
    qkv_bias : bool, optional
        Whether to add a learnable bias to query, key, value, by default False
    attn_drop : float, optional
        Attention dropout rate, by default 0.0
    proj_drop : float, optional
        Dropout rate of output, by default 0.0

    References
    ----------
    [1] Liu et al., Swin Transformer: Hierarchical Vision Transformer using Shifted Windows
        <https://arxiv.org/abs/2103.14030>
        https://github.com/microsoft/Swin-Transformer
    """

    def __init__(
        self,
        dim: int,
        num_heads: int,
        window_size: Sequence[int],
        qkv_bias: bool = False,
        attn_drop: float = 0,
        proj_drop: float = 0,
    ) -> None:
        """Inits :class:`WindowAttention3d`.

        Parameters
        ----------
        dim : int
            Number of feature channels.
        num_heads : int
            Number of attention heads.
        window_size : sequence of int
            Local window size for 3D (depth, height, width).
        qkv_bias : bool, optional
            Whether to add a learnable bias to query, key, value, by default False
        attn_drop : float, optional
            Attention dropout rate, by default 0.0
        proj_drop : float, optional
            Dropout rate of output, by default 0.0
        """
        relative_position_bias_table = nn.Parameter(
            torch.zeros(
                (2 * window_size[0] - 1) * (2 * window_size[1] - 1) * (2 * window_size[2] - 1),
                num_heads,
            )
        )
        coords_d = torch.arange(window_size[0])
        coords_h = torch.arange(window_size[1])
        coords_w = torch.arange(window_size[2])
        coords = torch.stack(torch.meshgrid(coords_d, coords_h, coords_w, indexing="ij"))
        coords_flatten = torch.flatten(coords, 1)
        relative_coords = coords_flatten[:, :, None] - coords_flatten[:, None, :]
        relative_coords = relative_coords.permute(1, 2, 0).contiguous()
        relative_coords[:, :, 0] += window_size[0] - 1
        relative_coords[:, :, 1] += window_size[1] - 1
        relative_coords[:, :, 2] += window_size[2] - 1
        relative_coords[:, :, 0] *= (2 * window_size[1] - 1) * (2 * window_size[2] - 1)
        relative_coords[:, :, 1] *= 2 * window_size[2] - 1
        super().__init__(
            dim, num_heads, window_size, relative_coords, relative_position_bias_table, qkv_bias, attn_drop, proj_drop
        )
