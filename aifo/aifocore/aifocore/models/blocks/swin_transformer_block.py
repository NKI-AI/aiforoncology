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
import math
from typing import Sequence

import torch
import torch.nn as nn
import torch.nn.functional as F
import torch.utils.checkpoint as checkpoint
from torch.nn import LayerNorm

from aifocore.models.blocks.mlp_block import MLPBlock
from aifocore.models.modules import DropPath, WindowAttentionBase


def _get_window_size(
    input_dimensions: Sequence[int],
    window_size: Sequence[int],
    shift_size: Sequence[int],
) -> tuple[tuple[int], tuple[int]]:
    """Computing window size based on: [1]

    Avoids windows larger than input size.

    Parameters
    ----------
    input_dimensions : Sequence[int]
        Input size (D1, ..., Dn).
    window_size : Sequence[int]
        Local window size (D1, ..., Dn).
    shift_size : Sequence[int]
        Window shifting size (D1, ..., Dn).

    Returns
    -------
    tuple[tuple[int], tuple[int]]
        Tuple containing the computed window size and shift size.

    References
    ----------
    [1] Liu et al., Swin Transformer: Hierarchical Vision Transformer using Shifted Windows
        <https://arxiv.org/abs/2103.14030>
        https://github.com/microsoft/Swin-Transformer
    """

    use_window_size = list(window_size)
    use_shift_size = list(shift_size)
    for i in range(len(input_dimensions)):
        if input_dimensions[i] <= window_size[i]:
            use_window_size[i] = input_dimensions[i]
            use_shift_size[i] = 0

    return tuple(use_window_size), tuple(use_shift_size)


def _pad_to_window_size(input: torch.Tensor, window_size: Sequence[int]) -> torch.Tensor:
    """Pad input tensor to match window size.

    Parameters
    ----------
    input : torch.Tensor
        Input tensor in format of B, D1, ..., Dn, C.
    window_size : Sequence[int]
        Window size of shape D1, ..., Dn.

    Returns
    -------
    torch.Tensor
        Padded input tensor.
    """
    dimensions = input.shape[1:-1]
    padding_start = [0 for _ in range(len(dimensions))]

    required_padding = []
    for input_size, window_divisor in zip(dimensions, window_size):
        required_padding.append((window_divisor - input_size % window_divisor) % window_divisor)
    # padding starts from last dimension to first in torch functional padding
    required_padding = reversed(required_padding)

    # interleaves the start of padding for each dimension with amount of padding that is needed for each dimension
    padding_sequence = [
        pad_val for start_end_pair in zip(padding_start, required_padding) for pad_val in start_end_pair
    ]

    # do not pad the channel dimension
    input_padded = F.pad(input, (0, 0, *padding_sequence))
    return input_padded


def _window_partition(input: torch.Tensor, window_size: Sequence[int]) -> torch.Tensor:
    """Window partition operation based on: [1]

    Parameters
    ----------
    input : torch.Tensor
        Input tensor.
    window_size : Sequence[int]
        Local window size.

    Returns
    -------
    torch.Tensor
        Partitioned windows tensor.

    References
    ----------
    [1] Liu et al., Swin Transformer: Hierarchical Vision Transformer using Shifted Windows
        <https://arxiv.org/abs/2103.14030>
        https://github.com/microsoft/Swin-Transformer
    """

    B, *spatial_dimensions, C = input.shape
    # transform input into windows such that we have:
    #    [B, D1 // window_size[1], window_size[1], ..., Dn // window_size[n], window_size[n]]
    input = input.view(
        B,
        *[
            item
            for i in range(len(spatial_dimensions))
            for item in [spatial_dimensions[i] // window_size[i], window_size[i]]
        ],
        C,
    )
    input_windows_length = len(input.shape)
    # permute, so that the windows and dimensions are grouped, then flatten the window dims
    windows = (
        input.permute(
            0,  # batch
            *[1 + 2 * i for i in range(len(spatial_dimensions))],  # window counts: 1, 3, 5, ...
            *[2 + 2 * i for i in range(len(spatial_dimensions))],  # window sizes: 2, 4, 6, ...
            input_windows_length - 1,  # channels
            # flatten and reshape to num_windows x windows x C
        )
        .contiguous()
        .view(-1, math.prod(window_size), C)
    )

    return windows


def _window_reverse(windows: torch.Tensor, window_size: Sequence[int], original_dims: torch.Size) -> torch.Tensor:
    """Window reverse operation based on: [1]

    Parameters
    ----------
    windows : torch.Tensor
        Windows tensor of shape num_windows x window_size_1 x ... x windows_size_n x C.
    window_size : Sequence[int]
        Local window size, sequence of shape (D1, ..., Dn).
    original_dims : torch.Size
        Dimension values of the tensor pre-partitioning, torch.Size of (B, D1, ..., Dn, C).

    Returns
    -------
    torch.Tensor
        Output tensor with spatial dimensions of original_dims.

    References
    ----------
    [1] Liu et al., Swin Transformer: Hierarchical Vision Transformer using Shifted Windows
        <https://arxiv.org/abs/2103.14030>
        https://github.com/microsoft/Swin-Transformer
    """
    B, *spatial_dims, C = original_dims
    output = windows.view(
        B,
        *[spatial_dims[i] // window_size[i] for i in range(len(spatial_dims))],
        *[window_size[i] for i in range(len(window_size))],
        C,
    )

    # number of spatial dimensions
    N = len(spatial_dims)

    # reorder the windows to be of shape (B, #windows_d1, windows_d1, ..., #windows_dn, windows_dn, C)
    permute_sequence = [0] + sum([[1 + i, 1 + N + i] for i in range(N)], []) + [2 * N + 1]

    output = output.permute(*permute_sequence).contiguous()
    return output.view(B, *spatial_dims, C)


class SwinTransformerBlock(nn.Module):
    """
    Swin Transformer block based on: [1]

    Parameters
    ----------
    dim : int
        Number of feature channels.
    window_attention : type[WindowAttentionBase]
        Window attention module type to use.
    num_heads : int
        Number of attention heads.
    window_size : Sequence[int]
        Local window size.
    shift_size : Sequence[int]
        Window shift size.
    mlp_ratio : float, optional
        Ratio of mlp hidden dim to embedding dim, by default 4.0
    qkv_bias : bool, optional
        Whether to add learnable bias to query, key, value, by default True
    drop : float, optional
        Dropout rate, by default 0.0
    attn_drop : float, optional
        Attention dropout rate, by default 0.0
    drop_path : float, optional
        Stochastic depth rate, by default 0.0
    activation_function : type[nn.Module], optional
        Activation layer type, by default nn.GELU
    norm_layer : type[LayerNorm], optional
        Normalization layer type, by default nn.LayerNorm
    use_checkpoint : bool, optional
        Whether to use gradient checkpointing for reduced memory usage, by default False

    References
    ----------
    [1] Liu et al., Swin Transformer: Hierarchical Vision Transformer using Shifted Windows
        <https://arxiv.org/abs/2103.14030>
        https://github.com/microsoft/Swin-Transformer
    """

    def __init__(
        self,
        dim: int,
        window_attention: type[WindowAttentionBase],
        num_heads: int,
        window_size: Sequence[int],
        shift_size: Sequence[int],
        mlp_ratio: float = 4.0,
        qkv_bias: bool = True,
        dropout: float = 0.0,
        attn_drop: float = 0.0,
        drop_path: float = 0.0,
        activation_function: type[nn.Module] = nn.GELU,
        norm_layer: type[LayerNorm] = nn.LayerNorm,
        use_checkpoint: bool = False,
    ) -> None:
        """Inits :class:`SwinTransformerBlock`

        Parameters
        ----------
        dim : int
            Number of feature channels.
        window_attention : type[WindowAttentionBase]
            Window attention module type to use.
        num_heads : int
            Number of attention heads.
        window_size : Sequence[int]
            Local window size.
        shift_size : Sequence[int]
            Window shift size.
        mlp_ratio : float, optional
            Ratio of mlp hidden dim to embedding dim, by default 4.0
        qkv_bias : bool, optional
            Whether to add learnable bias to query, key, value, by default True
        drop : float, optional
            Dropout rate, by default 0.0
        attn_drop : float, optional
            Attention dropout rate, by default 0.0
        drop_path : float, optional
            Stochastic depth rate, by default 0.0
        activation_function : type[nn.Module], optional
            Activation layer type, by default nn.GELU
        norm_layer : type[LayerNorm], optional
            Normalization layer type, by default nn.LayerNorm
        use_checkpoint : bool, optional
            Whether to use gradient checkpointing for reduced memory usage, by default False
        """

        super().__init__()
        self.num_heads = num_heads
        self.window_size = window_size
        self.shift_size = shift_size
        self.mlp_ratio = mlp_ratio
        self.use_checkpoint = use_checkpoint
        self.norm1 = norm_layer(dim)
        self.attn = window_attention(
            dim=dim,
            window_size=self.window_size,
            num_heads=num_heads,
            qkv_bias=qkv_bias,
            attn_drop=attn_drop,
            proj_drop=dropout,
        )

        self.drop_path = DropPath(drop_path) if drop_path > 0.0 else nn.Identity()
        self.norm2 = norm_layer(dim)
        mlp_hidden_dim = int(dim * mlp_ratio)
        self.mlp = MLPBlock(
            hidden_size=dim,
            mlp_dim=mlp_hidden_dim,
            dropout_rate=dropout,
            activation_function=activation_function,
        )

    def sliding_window_attn_forward(self, x: torch.Tensor, mask_matrix: torch.Tensor) -> torch.Tensor:
        """Applies sliding window attention to input tensor.

        Handles padding when window size does not evenly divide input dimensions.
        Includes window shifting, partitioning, attention computation and reversal
        of these operations.

        Parameters
        ----------
        x : torch.Tensor
            Input tensor of shape (B, D1, D2, ..., Dn, C) where B is batch size,
            D1...Dn are spatial dimensions, and C is number of channels
        mask_matrix : torch.Tensor
            Attention mask matrix for shifted windows

        Returns
        -------
        torch.Tensor
            Output tensor after applying sliding window attention,
            same shape as input (B, D1, D2, ..., Dn, C)
        """
        _, *spatial_dimensions, C = x.shape
        x = self.norm1(x)
        window_size, shift_size = _get_window_size(spatial_dimensions, self.window_size, self.shift_size)
        # allow for dynamic image sizing
        x = _pad_to_window_size(x, window_size)
        shape_after_pad = x.shape

        # shift windows if block is part of shifted window layer
        dimensions_to_shift = tuple(torch.arange(1, len(spatial_dimensions) + 1))
        if any(i > 0 for i in shift_size):
            shifted_x = torch.roll(x, shifts=tuple(-torch.tensor(shift_size)), dims=dimensions_to_shift)
            attn_mask = mask_matrix
        else:
            shifted_x = x
            attn_mask = None

        # partition windows, perform SW-MSA, and reverse partitioning
        x_windows = _window_partition(shifted_x, window_size)
        attn_windows = self.attn.forward(x_windows, mask=attn_mask)
        attn_windows = attn_windows.view(-1, *(window_size + (C,)))
        shifted_x = _window_reverse(attn_windows, window_size, shape_after_pad)

        # reverse the cyclic shift if applied earlier
        if any(i > 0 for i in shift_size):
            x = torch.roll(shifted_x, shifts=shift_size, dims=dimensions_to_shift)
        else:
            x = shifted_x

        # remove padding for each dimension if padding was applied
        if any(
            padded_dim != original_dim for padded_dim, original_dim in zip(shape_after_pad[1:-1], spatial_dimensions)
        ):
            for i, dim_size in enumerate(spatial_dimensions):
                x = x.narrow(dim=i + 1, start=0, length=dim_size)

        return x.contiguous()

    def projection_forward(self, x: torch.Tensor) -> torch.Tensor:
        """Project input through normalization, MLP and dropout layers.

        Parameters
        ----------
        x : torch.Tensor
            Input tensor of shape (B, D1, D2, ..., Dn, C) where B is batch size,
            D1...Dn are spatial dimensions, and C is number of channels.

        Returns
        -------
        torch.Tensor
            Output tensor after applying layer normalization, MLP projection and dropout,
            same shape as input (B, D1, D2, ..., Dn, C).
        """
        return self.drop_path(self.mlp(self.norm2(x)))

    def load_from(self, weights: dict[str, torch.Tensor], n_block: int, layer: str) -> None:
        root = f"module.{layer}.0.blocks.{n_block}."
        block_names = [
            "norm1.weight",
            "norm1.bias",
            "attn.relative_position_bias_table",
            "attn.relative_position_index",
            "attn.qkv.weight",
            "attn.qkv.bias",
            "attn.proj.weight",
            "attn.proj.bias",
            "norm2.weight",
            "norm2.bias",
            "mlp.fc1.weight",
            "mlp.fc1.bias",
            "mlp.fc2.weight",
            "mlp.fc2.bias",
        ]
        with torch.no_grad():
            self.norm1.weight.copy_(weights["state_dict"][root + block_names[0]])
            self.norm1.bias.copy_(weights["state_dict"][root + block_names[1]])
            self.attn.relative_position_bias_table.copy_(weights["state_dict"][root + block_names[2]])
            self.attn.relative_position_index.copy_(weights["state_dict"][root + block_names[3]])  # type: ignore[operator]
            self.attn.qkv.weight.copy_(weights["state_dict"][root + block_names[4]])
            self.attn.qkv.bias.copy_(weights["state_dict"][root + block_names[5]])
            self.attn.proj.weight.copy_(weights["state_dict"][root + block_names[6]])
            self.attn.proj.bias.copy_(weights["state_dict"][root + block_names[7]])
            self.norm2.weight.copy_(weights["state_dict"][root + block_names[8]])
            self.norm2.bias.copy_(weights["state_dict"][root + block_names[9]])
            self.mlp.linear1.weight.copy_(weights["state_dict"][root + block_names[10]])
            self.mlp.linear1.bias.copy_(weights["state_dict"][root + block_names[11]])
            self.mlp.linear2.weight.copy_(weights["state_dict"][root + block_names[12]])
            self.mlp.linear2.bias.copy_(weights["state_dict"][root + block_names[13]])

    def forward(self, x: torch.Tensor, mask_matrix: torch.Tensor) -> torch.Tensor:
        """Forward pass of Swin Transformer block.

        Parameters
        ----------
        x : torch.Tensor
            Input tensor of shape (B, D1, D2, ..., Dn, C) where B is batch size,
            D1...Dn are spatial dimensions, and C is number of channels.
        mask_matrix : torch.Tensor
            Attention mask matrix used for shifted window partitioning.
            Shape depends on input spatial dimensions and window size.

        Returns
        -------
        torch.Tensor
            Output tensor after applying self-attention and MLP blocks with residual
            connections. Has same shape as input (B, D1, D2, ..., Dn, C).
        """
        shortcut = x
        if self.use_checkpoint:
            x = checkpoint.checkpoint(self.sliding_window_attn_forward, x, mask_matrix, use_reentrant=False)
        else:
            x = self.sliding_window_attn_forward(x, mask_matrix)
        x = shortcut + self.drop_path(x)
        if self.use_checkpoint:
            x = x + checkpoint.checkpoint(self.projection_forward, x, use_reentrant=False)
        else:
            x = x + self.projection_forward(x)
        return x
