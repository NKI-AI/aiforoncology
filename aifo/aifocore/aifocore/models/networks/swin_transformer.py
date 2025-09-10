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
from typing import Sequence, Union

import torch
import torch.nn as nn
import torch.nn.functional as F
from torch.nn import LayerNorm

from aifocore.models.layers import SwinTransformerStage
from aifocore.models.modules import PatchMerging, WindowAttentionBase, PatchEmbed, PatchEmbed3d


class SwinTransformer(nn.Module):
    """Swin Transformer based on: "Liu et al.,
    Swin Transformer: Hierarchical Vision Transformer using Shifted Windows
    <https://arxiv.org/abs/2103.14030>"
    https://github.com/microsoft/Swin-Transformer

    Parameters
    ----------
    in_chans : int
        Dimension of input channels
    embed_dim : int
        Number of linear projection output channels
    img_size : Sequence[int]
        Input image size
    window_size : Sequence[int]
        Local window size
    patch_size : Sequence[int]
        Patch size
    depths : Sequence[int]
        Number of layers in each stage
    num_heads : Sequence[int]
        Number of attention heads
    window_attn : type[WindowAttentionBase]
        Window attention module type
    patch_embed : Union[type[PatchEmbed], type[PatchEmbed3d]]
        Patch embedding module type
    mlp_ratio : float, optional
        Ratio of mlp hidden dim to embedding dim, by default 4.0
    qkv_bias : bool, optional
        Add learnable bias to query, key, value, by default True
    drop_rate : float, optional
        Dropout rate, by default 0.0
    attn_drop_rate : float, optional
        Attention dropout rate, by default 0.0
    drop_path_rate : float, optional
        Stochastic depth rate, by default 0.0
    norm_layer : type[LayerNorm], optional
        Normalization layer, by default nn.LayerNorm
    patch_norm : bool, optional
        Add normalization after patch embedding, by default False
    use_checkpoint : bool, optional
        Use gradient checkpointing for reduced memory usage, by default False
    """

    _NUM_STAGES = 4

    def __init__(
        self,
        in_chans: int,
        embed_dim: int,
        img_size: Sequence[int],
        window_size: Sequence[int],
        patch_size: Sequence[int],
        depths: tuple[int, int, int, int],
        num_heads: tuple[int, int, int, int],
        window_attn: type[WindowAttentionBase],
        patch_embed: Union[type[PatchEmbed], type[PatchEmbed3d]],
        mlp_ratio: float = 4.0,
        qkv_bias: bool = True,
        drop_rate: float = 0.0,
        attn_drop_rate: float = 0.0,
        drop_path_rate: float = 0.0,
        norm_layer: type[LayerNorm] = nn.LayerNorm,
        patch_norm: bool = False,
        use_checkpoint: bool = False,
    ) -> None:
        """Inits :class:`SwinTransformer`.

        Parameters
        ----------
        in_chans : int
            Dimension of input channels
        embed_dim : int
            Number of linear projection output channels
        img_size : Sequence[int]
            Input image size
        window_size : Sequence[int]
            Local window size
        patch_size : Sequence[int]
            Patch size
        depths : tuple[int, int, int, int]
            Number of layers in each of the four stages
        num_heads : tuple[int, int, int, int]
            Number of attention heads per stage
        window_attn : type[WindowAttentionBase]
            Window attention module type
        patch_embed : Union[type[PatchEmbed], type[PatchEmbed3d]]
            Patch embedding module type
        mlp_ratio : float, optional
            Ratio of mlp hidden dim to embedding dim, by default 4.0
        qkv_bias : bool, optional
            Add learnable bias to query, key, value, by default True
        drop_rate : float, optional
            Dropout rate, by default 0.0
        attn_drop_rate : float, optional
            Attention dropout rate, by default 0.0
        drop_path_rate : float, optional
            Stochastic depth rate, by default 0.0
        norm_layer : type[LayerNorm], optional
            Normalization layer, by default nn.LayerNorm
        patch_norm : bool, optional
            Add normalization after patch embedding, by default False
        use_checkpoint : bool, optional
            Use gradient checkpointing for reduced memory usage, by default False
        """

        if len(depths) != self._NUM_STAGES:
            raise ValueError(
                f"Depths has to be the length of number of stages: stages = `{self._NUM_STAGES}`, depths = `{depths}`"
            )

        if len(num_heads) != self._NUM_STAGES:
            raise ValueError(
                f"Number of heads has to be the length of number of stages: \
                    stages = `{self._NUM_STAGES}`, num_heads = `{num_heads}`"
            )

        super().__init__()
        self.embed_dim = embed_dim
        self.patch_norm = patch_norm
        self.window_size = window_size
        self.patch_size = patch_size
        self.patch_embed = patch_embed(
            img_size=img_size,
            patch_size=self.patch_size,
            in_chans=in_chans,
            embed_dim=embed_dim,
            norm_layer=(norm_layer if self.patch_norm else None),
            flatten_embedding=False,
        )
        self.pos_drop = nn.Dropout(p=drop_rate)
        dpr = [x.item() for x in torch.linspace(0, drop_path_rate, sum(depths))]
        self.stages: nn.ModuleList[SwinTransformerStage] = nn.ModuleList()
        for stage_idx in range(self._NUM_STAGES):
            self.stages.append(
                SwinTransformerStage(
                    dim=int(embed_dim * 2**stage_idx),
                    window_attention=window_attn,
                    depth=depths[stage_idx],
                    num_heads=num_heads[stage_idx],
                    window_size=self.window_size,
                    drop_path=dpr[sum(depths[:stage_idx]) : sum(depths[: stage_idx + 1])],
                    mlp_ratio=mlp_ratio,
                    qkv_bias=qkv_bias,
                    dropout=drop_rate,
                    attn_drop=attn_drop_rate,
                    norm_layer=norm_layer,
                    downsample=(PatchMerging if stage_idx < self._NUM_STAGES - 1 else None),
                    use_checkpoint=use_checkpoint,
                )
            )

        self.num_features = int(embed_dim * 2 ** (self._NUM_STAGES - 1))

    def proj_out(self, x: torch.Tensor, normalize: bool = False) -> torch.Tensor:
        """Normalize output features if specified.

        Parameters
        ----------
        x : torch.Tensor
            Input tensor of shape (B, C, D1, ..., Dn) where:
                B is batch size
                C is number of channels
                D1,...,Dn are spatial dimensions
        normalize : bool, optional
            Whether to apply layer normalization, by default False

        Returns
        -------
        torch.Tensor
            Output tensor of same shape as input, normalized if specified
        """
        if normalize:
            channels = int(x.shape[1])
            # B, C, D1, ..., Dn -> B, D1, ..., Dn, C
            x = x.movedim(1, -1)
            x = F.layer_norm(x, [channels])
            # B, D1, ..., Dn, C -> B, C, D1, ..., Dn
            x = x.movedim(-1, 1)
        return x

    def forward(self, x: torch.Tensor, normalize: bool = True) -> list[torch.Tensor]:
        """Forward pass of the Swin Transformer.

        Parameters
        ----------
        x : torch.Tensor
            Input tensor of shape (B, C, D1, ..., Dn) where:
                B is batch size
                C is number of input channels
                D1,...,Dn are spatial dimensions
        normalize : bool, optional
            Whether to apply layer normalization to outputs, by default True

        Returns
        -------
        list[torch.Tensor]
            List of feature maps at different scales:
            - Patch embedding output, shape (B, embed_dim, D1, ..., Dn)
            - Stage 1 output, shape (B, embed_dim, D1, ..., Dn)
            - Stage 2 output, shape (B, embed_dim*2, D1 / 2, ..., Dn / 2)
            - Stage 3 output, shape (B, embed_dim*4, D1 / 4, ..., Dn / 4)
            - Stage 4 output, shape (B, embed_dim*8, D1 / 8, ..., Dn / 8)
        """
        # patch embedding returns B, D1, ..., Dn, C
        x = self.patch_embed(x).movedim(-1, 1)
        x: torch.Tensor = self.pos_drop(x)
        outs = [self.proj_out(x, normalize)]

        for stage in self.stages:
            x, x_pre_downsample = stage(x.contiguous())
            outs.append(self.proj_out(x_pre_downsample))

        return outs
