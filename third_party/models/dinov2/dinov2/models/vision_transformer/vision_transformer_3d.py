# Copyright 2025 AI for Oncology Research Group. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

from functools import partial
from typing_extensions import override

import torch
from torch import nn

from dinov2.layers import MemEffAttention
from dinov2.layers import NestedTensorBlock as Block
from dinov2.models.vision_transformer.vision_transformer_base import (
    DinoVisionTransformerBase,
    DinoVisionTransformerDim,
    DinoVisionTransformerFFNLayer,
)


class DinoVisionTransformer3d(DinoVisionTransformerBase):
    """DinoVisionTransformer for 3D images.

    Parameters
    ----------
    img_size : int or tuple[int, int, int]
        Input image size, either a single integer or a tuple of three integers (depth, height, width).
    patch_size : int or tuple[int, int, int]
        Patch size, either a single integer or a tuple of three integers (depth, height, width).
    in_chans : int
        Number of input channels, default is 3.
    embed_dim : int
        Embedding dimension.
    depth : int
        Depth of transformer.
    num_heads : int
        Number of attention heads.
    mlp_ratio : int
        Ratio of mlp hidden dim to embedding dim.
    qkv_bias : bool
        Enable bias for qkv if True.
    proj_bias : bool
        Enable bias for proj in attn if True.
    ffn_bias : bool
        Enable bias for ffn if True.
    drop_path_rate : float
        Stochastic depth rate.
    drop_path_uniform : bool
        Apply uniform drop rate across blocks.
    weight_init : str
        Weight init scheme.
    init_values : float
        Layer-scale init values.
    act_layer : nn.Module
        MLP activation layer.
    block_fn : nn.Module
        Transformer block class.
    ffn_layer : DinoVisionTransformerFFNLayer
        Type of FFN layer to use, can be DinoVisionTransformerFFNLayer.MLP,
        DinoVisionTransformerFFNLayer.SWIGLU, DinoVisionTransformerFFNLayer.SWIGLU_FUSED,
        or DinoVisionTransformerFFNLayer.IDENTITY. Default is DinoVisionTransformerFFNLayer.MLP.
    block_chunks : int
        Split block sequence into block_chunks units for FSDP wrap.
    num_register_tokens : int
        Number of extra cls tokens (so-called "registers").
    interpolate_antialias : str
        Flag to apply anti-aliasing when interpolating positional embeddings.
    interpolate_offset : float
        Work-around offset to apply when interpolating positional embeddings.
    """

    def __init__(
        self,
        img_size: int | tuple[int, int, int] = 224,
        patch_size: int | tuple[int, int, int] = 16,
        in_chans: int = 3,
        embed_dim: int = 768,
        depth: int = 12,
        num_heads: int = 12,
        mlp_ratio: float = 4.0,
        qkv_bias: bool = True,
        ffn_bias: bool = True,
        proj_bias: bool = True,
        drop_path_rate: float = 0.0,
        drop_path_uniform: bool = False,
        init_values: float | None = None,  # for layerscale: None or 0 => no layerscale
        act_layer: nn.Module = nn.GELU,
        block_fn: nn.Module = Block,
        ffn_layer: DinoVisionTransformerFFNLayer = DinoVisionTransformerFFNLayer.MLP,
        block_chunks: int = 1,
        num_register_tokens: int = 0,
        interpolate_antialias: bool = False,
        interpolate_offset: float = 0.1,
    ) -> None:
        """Inits :class:`DinoVisionTransformer3d`.

        Parameters
        ----------
        img_size : int or tuple[int, int, int]
            Input image size, either a single integer or a tuple of three integers (depth, height, width).
        patch_size : int or tuple[int, int, int]
            Patch size, either a single integer or a tuple of three integers (depth, height, width).
        in_chans : int
            Number of input channels, default is 3.
        embed_dim : int
            Embedding dimension.
        depth : int
            Depth of transformer.
        num_heads : int
            Number of attention heads.
        mlp_ratio : int
            Ratio of mlp hidden dim to embedding dim.
        qkv_bias : bool
            Enable bias for qkv if True.
        proj_bias : bool
            Enable bias for proj in attn if True.
        ffn_bias : bool
            Enable bias for ffn if True.
        drop_path_rate : float
            Stochastic depth rate.
        drop_path_uniform : bool
            Apply uniform drop rate across blocks.
        weight_init : str
            Weight init scheme.
        init_values : float
            Layer-scale init values.
        act_layer : nn.Module
            MLP activation layer.
        block_fn : nn.Module
            Transformer block class.
        ffn_layer : DinoVisionTransformerFFNLayer
            Type of FFN layer to use, can be DinoVisionTransformerFFNLayer.MLP,
            DinoVisionTransformerFFNLayer.SWIGLU, DinoVisionTransformerFFNLayer.SWIGLU_FUSED,
            or DinoVisionTransformerFFNLayer.IDENTITY. Default is DinoVisionTransformerFFNLayer.MLP.
        block_chunks : int
            Split block sequence into block_chunks units for FSDP wrap.
        num_register_tokens : int
            Number of extra cls tokens (so-called "registers").
        interpolate_antialias : str
            Flag to apply anti-aliasing when interpolating positional embeddings.
        interpolate_offset : float
            Work-around offset to apply when interpolating positional embeddings.
        """
        super().__init__(
            dim=DinoVisionTransformerDim.THREE_D,
            img_size=img_size,
            patch_size=patch_size,
            in_chans=in_chans,
            embed_dim=embed_dim,
            depth=depth,
            num_heads=num_heads,
            mlp_ratio=mlp_ratio,
            qkv_bias=qkv_bias,
            ffn_bias=ffn_bias,
            proj_bias=proj_bias,
            drop_path_rate=drop_path_rate,
            drop_path_uniform=drop_path_uniform,
            init_values=init_values,
            act_layer=act_layer,
            block_fn=block_fn,
            ffn_layer=ffn_layer,
            block_chunks=block_chunks,
            num_register_tokens=num_register_tokens,
            interpolate_antialias=interpolate_antialias,
            interpolate_offset=interpolate_offset,
        )

    @override
    def _interpolate_and_reshape_pos_embed(
        self,
        patch_pos_embed: torch.Tensor,
        patches_resolution: tuple[int, int, int],
        dim: int,
        interpolation_kwargs: dict,
    ) -> torch.Tensor:
        """Interpolate and reshape 3D patch positional embeddings.

        Parameters
        ----------
        patch_pos_embed : torch.Tensor
            Positional embedding tensor of shape (1, N, C).
        patches_resolution : tuple of ints
            Number of patches along each spatial dimension.
        dim : int
            Embedding dimension.
        interpolation_kwargs : dict
            Arguments passed to `F.interpolate`.

        Returns
        -------
        torch.Tensor
            Reshaped and interpolated tensor of shape (1, D, H, W, C),
            where D, H, W are the number of patches along depth, height, and width.
        """
        patch_pos_embed = patch_pos_embed.reshape(1, *patches_resolution, dim).permute(0, 4, 1, 2, 3)
        patch_pos_embed = nn.functional.interpolate(
            patch_pos_embed,
            mode="trilinear",
            antialias=self.interpolate_antialias,
            **interpolation_kwargs,
        )
        return patch_pos_embed.permute(0, 2, 3, 4, 1)


def vit_3d_small(
    patch_size: int | tuple[int, int, int] = (16, 16, 16),
    num_register_tokens: int = 4,
    **kwargs,
) -> DinoVisionTransformer3d:
    """Builds a small 3d vision transformer with 384-dimensional embeddings, 12 layers, 6 heads, and 4x MLP ratio.

    Parameters
    ----------
    patch_size : int or tuple[int, int, int]
        Patch size, either a single integer or a tuple of three integers (depth, height, width).
        Default is (16, 16, 16).
    num_register_tokens : int
        Number of extra cls tokens (so-called "registers"). Default is 4.
    kwargs : dict
        Additional keyword arguments to pass to the :class:`DinoVisionTransformer3d` constructor.

    Returns
    -------
    DinoVisionTransformer3d
        A small 3d vision transformer.
    """
    model = DinoVisionTransformer3d(
        patch_size=patch_size,
        embed_dim=384,
        depth=12,
        num_heads=6,
        mlp_ratio=4,
        block_fn=partial(Block, attn_class=MemEffAttention),
        num_register_tokens=num_register_tokens,
        **kwargs,
    )
    return model


def vit_3d_base(
    patch_size: int | tuple[int, int, int] = (16, 16, 16),
    num_register_tokens: int = 4,
    **kwargs,
) -> DinoVisionTransformer3d:
    """Builds a base 3d vision transformer with 768-dimensional embeddings, 12 layers, 12 heads, and 4x MLP ratio.

    Parameters
    ----------
    patch_size : int or tuple[int, int, int]
        Patch size, either a single integer or a tuple of three integers (depth, height, width).
        Default is (16, 16, 16).
    num_register_tokens : int
        Number of extra cls tokens (so-called "registers"). Default is 4.
    kwargs : dict
        Additional keyword arguments to pass to the :class:`DinoVisionTransformer3d` constructor.

    Returns
    -------
    DinoVisionTransformer3d
        A base 3d vision transformer.
    """
    model = DinoVisionTransformer3d(
        patch_size=patch_size,
        embed_dim=768,
        depth=12,
        num_heads=12,
        mlp_ratio=4,
        block_fn=partial(Block, attn_class=MemEffAttention),
        num_register_tokens=num_register_tokens,
        **kwargs,
    )
    return model


def vit_3d_large(
    patch_size: int | tuple[int, int, int] = (16, 16, 16),
    num_register_tokens: int = 4,
    **kwargs,
) -> DinoVisionTransformer3d:
    """Builds a large 3d vision transformer with 1024-dimensional embeddings, 24 layers, 16 heads, and 4x MLP ratio.

    Parameters
    ----------
    patch_size : int or tuple[int, int, int]
        Patch size, either a single integer or a tuple of three integers (depth, height, width).
        Default is (16, 16, 16).
    num_register_tokens : int
        Number of extra cls tokens (so-called "registers"). Default is 4.
    kwargs : dict
        Additional keyword arguments to pass to the :class:`DinoVisionTransformer3d` constructor.

    Returns
    -------
    DinoVisionTransformer3d
        A large 3d vision transformer.
    """
    model = DinoVisionTransformer3d(
        patch_size=patch_size,
        embed_dim=1024,
        depth=24,
        num_heads=16,
        mlp_ratio=4,
        block_fn=partial(Block, attn_class=MemEffAttention),
        num_register_tokens=num_register_tokens,
        **kwargs,
    )
    return model
