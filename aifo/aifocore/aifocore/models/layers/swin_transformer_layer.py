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
import math
from typing import Sequence

import torch
import torch.nn as nn
from torch.nn import LayerNorm

from aifocore.models.modules import WindowAttentionBase
from aifocore.models.blocks import SwinTransformerBlock
from aifocore.models.blocks.swin_transformer_block import _get_window_size, _window_partition


def _compute_mask(
    dimensions: Sequence[int], window_size: Sequence[int], shift_size: Sequence[int], device: torch.device | str
) -> torch.Tensor:
    """Computing region masks based on: [1]

    Calculates which areas within a window are required to pay attention to other windows after a shift.

    Parameters
    ----------
    dimensions : Sequence[int]
        Dimension values (D1, ..., Dn)
    window_size : Sequence[int]
        Local window size (D1, ..., Dn)
    shift_size : Sequence[int]
        Shift size (D1, ..., Dn)
    device : torch.device
        Device for the mask

    Returns
    -------
    torch.Tensor
        Attention mask tensor

    References
    ----------
    [1] Liu et al., Swin Transformer: Hierarchical Vision Transformer using Shifted Windows
        <https://arxiv.org/abs/2103.14030>
        https://github.com/microsoft/Swin-Transformer
    """

    img_mask = torch.zeros((1, *dimensions, 1), device=device)
    mask_indices = [
        [slice(-window_size[i]), slice(-window_size[i], -shift_size[i]), slice(-shift_size[i], None)]
        for i in range(len(dimensions))
    ]

    partition = 0
    for index in itertools.product(*mask_indices):
        img_mask[:, *index, :] = partition
        partition += 1

    mask_windows = _window_partition(img_mask, window_size)
    mask_windows = mask_windows.squeeze(-1)
    attn_mask = mask_windows.unsqueeze(1) - mask_windows.unsqueeze(2)
    attn_mask = attn_mask.masked_fill(attn_mask != 0, -torch.inf).masked_fill(attn_mask == 0, float(0.0))

    return attn_mask


class SwinTransformerStage(nn.Module):
    """SwinTransformer stage based on: [1]

    Parameters
    ----------
    dim : int
        Number of feature channels
    window_attention : type[WindowAttentionBase]
        Window attention module type
    depth : int
        Number of layers in each stage
    num_heads : int
        Number of attention heads
    window_size : Sequence[int]
        Local window size
    drop_path : float | Sequence[float]
        Stochastic depth rate
    mlp_ratio : float, optional
        Ratio of mlp hidden dim to embedding dim, by default 4.0
    qkv_bias : bool, optional
        Add learnable bias to query, key, value, by default False
    dropout : float, optional
        Dropout rate, by default 0.0
    attn_drop : float, optional
        Attention dropout rate, by default 0.0
    norm_layer : type[LayerNorm], optional
        Normalization layer, by default nn.LayerNorm
    downsample : nn.Module or None, optional
        Optional downsampling layer at the end of the layer, by default None
    use_checkpoint : bool, optional
        Use gradient checkpointing for reduced memory usage, by default False

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
        depth: int,
        num_heads: int,
        window_size: Sequence[int],
        drop_path: float | Sequence[float],
        mlp_ratio: float = 4.0,
        qkv_bias: bool = False,
        dropout: float = 0.0,
        attn_drop: float = 0.0,
        norm_layer: type[LayerNorm] = nn.LayerNorm,
        downsample: nn.Module | None = None,
        use_checkpoint: bool = False,
    ) -> None:
        """Inits :class:`SwinTransformerStage`.

        Parameters
        ----------
        dim : int
            Number of feature channels
        window_attention : type[WindowAttentionBase]
            Window attention module type
        depth : int
            Number of layers in each stage
        num_heads : int
            Number of attention heads
        window_size : Sequence[int]
            Local window size
        drop_path : float | Sequence[float]
            Stochastic depth rate
        mlp_ratio : float, optional
            Ratio of mlp hidden dim to embedding dim, by default 4.0
        qkv_bias : bool, optional
            Add learnable bias to query, key, value, by default False
        dropout : float, optional
            Dropout rate, by default 0.0
        attn_drop : float, optional
            Attention dropout rate, by default 0.0
        norm_layer : type[LayerNorm], optional
            Normalization layer, by default nn.LayerNorm
        downsample : nn.Module or None, optional
            Optional downsampling layer at the end of the layer, by default None
        use_checkpoint : bool, optional
            Use gradient checkpointing for reduced memory usage, by default False
        """
        super().__init__()
        self.window_size = window_size
        self.shift_size = tuple(i // 2 for i in window_size)
        self.no_shift = tuple(0 for _ in window_size)
        self.depth = depth
        self.use_checkpoint = use_checkpoint
        self.blocks = nn.ModuleList(
            [
                SwinTransformerBlock(
                    dim=dim,
                    window_attention=window_attention,
                    num_heads=num_heads,
                    window_size=self.window_size,
                    shift_size=(self.no_shift if (i % 2 == 0) else self.shift_size),
                    mlp_ratio=mlp_ratio,
                    qkv_bias=qkv_bias,
                    dropout=dropout,
                    attn_drop=attn_drop,
                    # TODO: might be interesting to activate drop_path before applying forward rather than masking
                    # for larger/deeper models to reduce memory and computation.
                    drop_path=(drop_path[i] if isinstance(drop_path, list) else drop_path),
                    norm_layer=norm_layer,
                    use_checkpoint=use_checkpoint,
                )
                for i in range(depth)
            ]
        )
        self.downsample = downsample
        if callable(self.downsample):
            self.downsample = downsample(dim=dim, norm_layer=norm_layer, spatial_dims=len(self.window_size))

    def forward(self, x: torch.Tensor) -> tuple[torch.Tensor, torch.Tensor]:
        """Forward pass of SwinTransformerLayer.

        Parameters
        ----------
        x : torch.Tensor
            Input tensor of shape (B, C, D1, ..., Dn) where:
                B is batch size
                C is number of channels
                D1,...,Dn are spatial dimensions

        Returns
        -------
        tuple[torch.Tensor, torch.Tensor]
            - Output of the stage after downsampling (if applicable)
            - Output of the stage before downsampling
        """
        B, _, *spatial_dims = x.shape
        window_size, shift_size = _get_window_size(spatial_dims, self.window_size, self.shift_size)
        # B, C, D1, ..., Dn -> B, D1, ..., Dn, C
        x = x.movedim(1, -1)

        dimension_sizes_after_pad = [
            int(math.ceil(dim / window_size[i])) * window_size[i] for i, dim in enumerate(spatial_dims)
        ]
        attn_mask = _compute_mask(dimension_sizes_after_pad, window_size, shift_size, x.device)
        for blk in self.blocks:
            x = blk(x, attn_mask)
        x = x.view(B, *spatial_dims, -1)
        # B, D1, ..., Dn, C -> B, C, D1, ..., Dn
        x_pre_downsample = x.clone().movedim(-1, 1)
        if self.downsample is not None:
            x = self.downsample(x)
        # B, D1, ..., Dn, C -> B, C, D1, ..., Dn
        x = x.movedim(-1, 1)

        return (x, x_pre_downsample)
