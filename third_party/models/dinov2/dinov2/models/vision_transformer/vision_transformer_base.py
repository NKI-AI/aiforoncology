# Copyright (c) Meta Platforms, Inc. and affiliates.
#
# This source code is licensed under the Apache License, Version 2.0
# found in the LICENSE file in the root directory of this source tree.

# References:
#   https://github.com/facebookresearch/dino/blob/main/vision_transformer.py
#   https://github.com/rwightman/pytorch-image-models/tree/master/timm/models/vision_transformer.py
#
# Modified by NKI-AI4Oncology, 04-2025

from abc import abstractmethod
from enum import Enum
import logging
import math
from functools import partial
from typing import Callable, Sequence

import torch
import torch.nn as nn
import torch.utils.checkpoint
from dinov2.layers import Mlp
from dinov2.layers import NestedTensorBlock as Block
from dinov2.layers import PatchEmbed, PatchEmbed3d, SwiGLUFFNFused
from torch.nn.init import trunc_normal_
from dinov2.layers.utils import make_2tuple, make_3tuple

logger = logging.getLogger("dinov2")


def named_apply(fn: Callable, module: nn.Module, name="", depth_first=True, include_root=False) -> nn.Module:
    if not depth_first and include_root:
        fn(module=module, name=name)
    for child_name, child_module in module.named_children():
        child_name = ".".join((name, child_name)) if name else child_name
        named_apply(fn=fn, module=child_module, name=child_name, depth_first=depth_first, include_root=True)
    if depth_first and include_root:
        fn(module=module, name=name)
    return module


class BlockChunk(nn.ModuleList):
    """Block chunk for FSDP wrap."""

    def forward(self, x: torch.Tensor) -> torch.Tensor:
        """Forward pass through the block chunk.

        Parameters
        ----------
        x : torch.Tensor
            Input tensor.

        Returns
        -------
        torch.Tensor
            Output tensor.
        """
        for b in self:
            x = b(x)
        return x


class DinoVisionTransformerDim(str, Enum):
    """Dimension type for DinoVisionTransformer."""

    TWO_D = "2d"
    THREE_D = "3d"


class DinoVisionTransformerFFNLayer(str, Enum):
    """FFN layer type for DinoVisionTransformer."""

    MLP = "mlp"
    SWIGLU = "swiglu"
    SWIGLU_FUSED = "swiglufused"
    IDENTITY = "identity"

    @classmethod
    def _missing_(cls, value):
        if isinstance(value, str):
            value = value.lower()
            for member in cls:
                if member.value == value:
                    return member
        raise ValueError(f"{value!r} is not a valid {cls.__name__}")


class DinoVisionTransformerBase(nn.Module):
    """Base class for DinoVisionTransformer, supporting both 2D and 3D vision transformers.

    Parameters
    ----------
    dim : DinoVisionTransformerDim
        Dimension type, either DinoVisionTransformerDim.TWO_D or DinoVisionTransformerDim.THREE_D.
    img_size : int, tuple[int, int] or tuple[int, int, int]
        Input image size, either a single integer or a tuple.
        For 2D, it should be a tuple of two integers (height, width).
        For 3D, it should be a tuple of three integers (depth, height, width).
    patch_size : int, tuple[int, int] or tuple[int, int, int]
        Patch size, either a single integer or a tuple.
        For 2D, it should be a tuple of two integers (height, width).
        For 3D, it should be a tuple of three integers (depth, height, width).
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
        dim: DinoVisionTransformerDim,
        img_size: int | tuple[int, int] | tuple[int, int, int] = 224,
        patch_size: int | tuple[int, int] | tuple[int, int, int] = 16,
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
        """Inits :class:`DinoVisionTransformerBase`.

        Parameters
        ----------
        dim : DinoVisionTransformerDim
            Dimension type, either DinoVisionTransformerDim.TWO_D or DinoVisionTransformerDim.THREE_D.
        img_size : int, tuple[int, int] or tuple[int, int, int]
            Input image size, either a single integer or a tuple.
            For 2D, it should be a tuple of two integers (height, width).
            For 3D, it should be a tuple of three integers (depth, height, width).
        patch_size : int, tuple[int, int] or tuple[int, int, int]
            Patch size, either a single integer or a tuple.
            For 2D, it should be a tuple of two integers (height, width).
            For 3D, it should be a tuple of three integers (depth, height, width).
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

        super().__init__()
        self.logger = logging.getLogger(type(self).__name__)
        self.dim = dim

        norm_layer = partial(nn.LayerNorm, eps=1e-6)

        self.num_features = self.embed_dim = embed_dim  # num_features for consistency with other models
        self.num_tokens = 1
        self.n_blocks = depth
        self.num_heads = num_heads

        self.patch_size = make_2tuple(patch_size) if dim == DinoVisionTransformerDim.TWO_D else make_3tuple(patch_size)
        self.img_size = make_2tuple(img_size) if dim == DinoVisionTransformerDim.TWO_D else make_3tuple(img_size)

        if len(self.patch_size) != len(self.img_size):
            raise ValueError("Patch size and image size must have the same number of dimensions")

        self.num_register_tokens = num_register_tokens
        self.interpolate_antialias = interpolate_antialias
        self.interpolate_offset = interpolate_offset

        self.patch_embed = (PatchEmbed if dim == DinoVisionTransformerDim.TWO_D else PatchEmbed3d)(
            img_size=self.img_size,
            patch_size=self.patch_size,
            in_chans=in_chans,
            embed_dim=embed_dim,
        )
        num_patches = self.patch_embed.num_patches

        self.cls_token = nn.Parameter(torch.zeros(1, 1, embed_dim))
        self.pos_embed = nn.Parameter(torch.zeros(1, num_patches + self.num_tokens, embed_dim))
        assert num_register_tokens >= 0
        self.register_tokens = (
            nn.Parameter(torch.zeros(1, num_register_tokens, embed_dim)) if num_register_tokens else None
        )

        if drop_path_uniform is True:
            dpr = [drop_path_rate] * depth
        else:
            dpr = [x.item() for x in torch.linspace(0, drop_path_rate, depth)]  # stochastic depth decay rule

        if ffn_layer == DinoVisionTransformerFFNLayer.MLP:
            self.logger.info("Using MLP layer as FFN")
            ffn_layer = Mlp
        elif (
            ffn_layer == DinoVisionTransformerFFNLayer.SWIGLU or ffn_layer == DinoVisionTransformerFFNLayer.SWIGLU_FUSED
        ):
            self.logger.info("Using SwiGLU layer as FFN")
            ffn_layer = SwiGLUFFNFused
        else:  # ffn_layer == DinoVisionTransformerFFNLayer.IDENTITY:
            self.logger.info("Using Identity layer as FFN")
            ffn_layer = nn.Identity

        blocks_list = [
            block_fn(
                dim=embed_dim,
                num_heads=num_heads,
                mlp_ratio=mlp_ratio,
                qkv_bias=qkv_bias,
                proj_bias=proj_bias,
                ffn_bias=ffn_bias,
                drop_path=dpr[i],
                norm_layer=norm_layer,
                act_layer=act_layer,
                ffn_layer=ffn_layer,
                init_values=init_values,
            )
            for i in range(depth)
        ]
        if block_chunks > 0:
            self.chunked_blocks = True
            chunked_blocks = []
            chunksize = depth // block_chunks
            for i in range(0, depth, chunksize):
                # this is to keep the block index consistent if we chunk the block list
                chunked_blocks.append([nn.Identity()] * i + blocks_list[i : i + chunksize])
            self.blocks = nn.ModuleList([BlockChunk(p) for p in chunked_blocks])
        else:
            self.chunked_blocks = False
            self.blocks = nn.ModuleList(blocks_list)

        self.norm = norm_layer(embed_dim)
        self.head = nn.Identity()

        self.mask_token = nn.Parameter(torch.zeros(1, embed_dim))

        self.init_weights()

    def init_weights(self) -> None:
        """Initialize weights of the model."""
        trunc_normal_(self.pos_embed, std=0.02)
        nn.init.normal_(self.cls_token, std=1e-6)
        if self.register_tokens is not None:
            nn.init.normal_(self.register_tokens, std=1e-6)
        named_apply(init_weights_vit_timm, self)

    def _interpolate_pos_encoding(
        self, x: torch.Tensor, img_shape: tuple[int, int] | tuple[int, int, int]
    ) -> torch.Tensor:
        """Interpolate the positional encoding to match the input image shape.

        This method resizes the positional encoding tensor to match the spatial dimensions of the input tensor.

        Parameters
        ----------
        x : torch.Tensor
            Input tensor of shape (B, N, C) where B is the batch size, N is the number of patches + tokens,
            and C is the embedding dimension.
        img_shape : tuple[int, int] | tuple[int, int, int]
            Spatial dimensions of the input image. For 2D, it should be a tuple of two integers (height, width).
            For 3D, it should be a tuple of three integers (depth, height, width).

        Returns
        -------
        torch.Tensor
            Interpolated positional encoding tensor of shape (1, N, C), where N is the number of patches + tokens
        """
        previous_dtype = x.dtype
        num_image_patches = x.shape[1] - 1

        N = self.pos_embed.shape[1] - 1

        if num_image_patches == N and all(img_shape[i] == img_shape[i + 1] for i in range(len(img_shape) - 1)):
            return self.pos_embed

        pos_embed = self.pos_embed.float()

        class_pos_embed = pos_embed[:, 0]
        patch_pos_embed = pos_embed[:, 1:]
        dim = x.shape[-1]

        img_shape0 = [img_shape[i] // self.patch_size[i] for i in range(len(img_shape))]

        patches_resolution = self.patch_embed.patches_resolution  # Recover the number of patches in each dimension

        if N != math.prod(patches_resolution):
            raise ValueError(
                f"Mismatch: learned pos_embed has {N} tokens, but expected {math.prod(patches_resolution)} patches "
                f"corresponding to {patches_resolution} resolution."
            )

        interpolation_kwargs = {}
        if self.interpolate_offset:
            scale_factor = [float(s + self.interpolate_offset) / m for (s, m) in zip(img_shape0, patches_resolution)]
            interpolation_kwargs["scale_factor"] = scale_factor
        else:
            # Simply specify an output size instead of a scale factor
            interpolation_kwargs["size"] = img_shape0

        patch_pos_embed = self._interpolate_and_reshape_pos_embed(
            patch_pos_embed, patches_resolution, dim, interpolation_kwargs
        )

        if tuple(img_shape0) != patch_pos_embed.shape[1:-1]:
            raise ValueError(
                f"Positional embedding shape mismatch: expected {img_shape0}, got {patch_pos_embed.shape[1:-1]}. "
                "This may lead to unexpected behavior."
            )

        patch_pos_embed = patch_pos_embed.view(1, -1, dim)

        return torch.cat((class_pos_embed.unsqueeze(0), patch_pos_embed), dim=1).to(previous_dtype)

    @abstractmethod
    def _interpolate_and_reshape_pos_embed(
        self, patch_pos_embed: torch.Tensor, patches_resolution: tuple[int, ...], dim: int, interpolation_kwargs: dict
    ) -> torch.Tensor:
        """Subclasses should implement interpolation and reshaping appropriate for 2D or 3D positional embeddings.

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
            Reshaped and interpolated tensor of shape (1, ..., ..., C).
        """
        raise NotImplementedError("Subclasses must implement `_interpolate_and_reshape_pos_embed` method.")

    def _prepare_tokens_with_masks(self, x: torch.Tensor, masks: torch.Tensor | None = None) -> torch.Tensor:
        """Prepare tokens with masks for the input tensor.

        This method applies patch embedding, adds class tokens, and interpolates positional encodings.
        If masks are provided, it replaces the corresponding patches with a mask token.

        Parameters
        ----------
        x : torch.Tensor
            Input tensor of shape (B, C, H, W) for 2D or (B, C, D, H, W) for 3D,
            where B is the batch size, C is the number of channels, and H, W (or D, H, W) are the spatial dimensions.
        masks : torch.Tensor, optional
            Optional mask tensor of shape (B, N) where B is the batch size and N is the number of patches.
            Default is None.

        Returns
        -------
        torch.Tensor
            Prepared tensor of shape (B, N, C) where B is the batch size, N is the number of patches + tokens,
            and C is the embedding dimension.
        """
        x_shape = x.shape[2:]
        x = self.patch_embed(x)
        if masks is not None:
            x = torch.where(masks.unsqueeze(-1), self.mask_token.to(x.dtype).unsqueeze(0), x)

        x = torch.cat((self.cls_token.expand(x.shape[0], -1, -1), x), dim=1)
        x = x + self._interpolate_pos_encoding(x, x_shape)
        if self.register_tokens is not None:
            x = torch.cat(
                (
                    x[:, :1],
                    self.register_tokens.expand(x.shape[0], -1, -1),
                    x[:, 1:],
                ),
                dim=1,
            )

        return x

    def forward_features_list(
        self, x_list: list[torch.Tensor], masks_list: list[torch.Tensor]
    ) -> list[dict[str, torch.Tensor]]:
        """Forward pass for a list of input tensors with corresponding masks.

        Parameters
        ----------
        x_list : list[torch.Tensor]
            List of input tensors, each of shape (B, C, H, W) for 2D or (B, C, D, H, W) for 3D,
            where B is the batch size, C is the number of channels, and H, W (or D, H, W) are the spatial dimensions.
        masks_list : list[torch.Tensor]
            List of mask tensors, each of shape (B, N) where B is the batch size and N is the number of patches.

        Returns
        -------
        list[dict[str, torch.Tensor]]
            List of dictionaries containing the normalized outputs and masks for each input tensor.
        """
        x = [self._prepare_tokens_with_masks(x, masks) for x, masks in zip(x_list, masks_list)]
        for blk in self.blocks:
            x = blk(x)

        all_x = x
        output = []
        for x, masks in zip(all_x, masks_list):
            x_norm = self.norm(x)
            output.append(
                {
                    "x_norm_clstoken": x_norm[:, 0],
                    "x_norm_regtokens": x_norm[:, 1 : self.num_register_tokens + 1],
                    "x_norm_patchtokens": x_norm[:, self.num_register_tokens + 1 :],
                    "x_prenorm": x,
                    "masks": masks,
                }
            )
        return output

    def forward_features(
        self,
        x: torch.Tensor | list[torch.Tensor],
        masks: torch.Tensor | list[torch.Tensor] = None,
    ) -> dict[str, torch.Tensor]:
        """Return features from the input.

        Parameters
        ----------
        x : torch.Tensor | list[torch.Tensor]
            Input tensor or list of input tensors.
        masks : torch.Tensor | list[torch.Tensor], optional
            Mask tensor or list of mask tensors.

        Returns
        -------
        dict[str, torch.Tensor]
            Dictionary containing the normalized outputs and masks.
        """
        if isinstance(x, list):
            return self.forward_features_list(x, masks)

        x = self._prepare_tokens_with_masks(x, masks)

        for blk in self.blocks:
            x = blk(x)

        x_norm = self.norm(x)
        return {
            "x_norm_clstoken": x_norm[:, 0],
            "x_norm_regtokens": x_norm[:, 1 : self.num_register_tokens + 1],
            "x_norm_patchtokens": x_norm[:, self.num_register_tokens + 1 :],
            "x_prenorm": x,
            "masks": masks,
        }

    def _get_intermediate_layers_not_chunked(self, x: torch.Tensor, n: int | list[int] = 1) -> list[torch.Tensor]:
        """Get intermediate layers from the transformer blocks."""
        x = self._prepare_tokens_with_masks(x)
        # If n is an int, take the n last blocks. If it's a list, take them
        output, total_block_len = [], len(self.blocks)
        blocks_to_take = range(total_block_len - n, total_block_len) if isinstance(n, int) else n
        for i, blk in enumerate(self.blocks):
            x = blk(x)
            if i in blocks_to_take:
                output.append(x)
        assert len(output) == len(blocks_to_take), f"only {len(output)} / {len(blocks_to_take)} blocks found"
        return output

    def _get_intermediate_layers_chunked(self, x: torch.Tensor, n: int | list[int] = 1) -> list[torch.Tensor]:
        """Get intermediate layers from the transformer blocks when using chunked blocks."""
        x = self._prepare_tokens_with_masks(x)
        output, i, total_block_len = [], 0, len(self.blocks[-1])
        # If n is an int, take the n last blocks. If it's a list, take them
        blocks_to_take = range(total_block_len - n, total_block_len) if isinstance(n, int) else n
        for block_chunk in self.blocks:
            for blk in block_chunk[i:]:  # Passing the nn.Identity()
                x = blk(x)
                if i in blocks_to_take:
                    output.append(x)
                i += 1
        assert len(output) == len(blocks_to_take), f"only {len(output)} / {len(blocks_to_take)} blocks found"
        return output

    def get_intermediate_layers(
        self,
        x: torch.Tensor,
        n: int | Sequence = 1,  # Layers or n last layers to take
        reshape: bool = False,
        return_class_token: bool = False,
        norm: bool = True,
    ) -> tuple[torch.Tensor | tuple[torch.Tensor]]:
        """Get intermediate layers from the transformer blocks.

        Parameters
        ----------
        x : torch.Tensor
            Input tensor.
        n : int or Sequence, optional
            Number of layers or specific layers to take.
        reshape : bool, optional
            Whether to reshape the output.
        return_class_token : bool, optional
            Whether to return the class token.
        norm : bool, optional
            Whether to apply normalization.

        Returns
        -------
        tuple[torch.Tensor | tuple[torch.Tensor]]
            Intermediate layers from the transformer blocks.
        """
        if self.chunked_blocks:
            outputs = self._get_intermediate_layers_chunked(x, n)
        else:
            outputs = self._get_intermediate_layers_not_chunked(x, n)

        if norm:
            outputs = [self.norm(out) for out in outputs]

        class_tokens = [out[:, 0] for out in outputs]
        outputs = [out[:, 1 + self.num_register_tokens :] for out in outputs]

        if reshape:
            B = x.size(0)
            spatial_dims = x.shape[2:]
            outputs = [
                out.reshape([B] + [s // p for s, p in zip(spatial_dims, self.patch_size)] + [-1])
                .permute([0] + [x.ndim - 1] + list(range(1, x.ndim - 1)))
                .contiguous()
                for out in outputs
            ]

        if return_class_token:
            return tuple(zip(outputs, class_tokens))

        return tuple(outputs)

    def forward(self, *args, is_training=False, **kwargs) -> dict[str, torch.Tensor] | torch.Tensor:
        """Forward pass of :class:`DinoVisionTransformerBase`."""
        ret = self.forward_features(*args, **kwargs)
        if is_training:
            return ret
        else:
            return self.head(ret["x_norm_clstoken"])


def init_weights_vit_timm(module: nn.Module, name: str = "") -> None:
    """ViT weight initialization, original timm impl (for reproducibility)"""
    if isinstance(module, nn.Linear):
        trunc_normal_(module.weight, std=0.02)
        if module.bias is not None:
            nn.init.zeros_(module.bias)
