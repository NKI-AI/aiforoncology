# Copyright Ross Wightman 2020. All Rights Reserved.
# Copyright 2025 AI for Oncology Research Group. All Rights Reserved.
#
# Modified from https://github.com/huggingface/pytorch-image-models/blob/main/timm/models/vision_transformer.py
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
import torch
import torch.nn as nn
from torch import Tensor

from typing import Callable, Literal, Optional, Type
from typing_extensions import override
from timm.layers.mlp import Mlp
from timm.layers.typing import LayerType
from timm.models import register_model, VisionTransformer
from timm.models.features import feature_take_indices
from timm.models.vision_transformer import Block
from timm.models._builder import build_model_with_cfg

from fef.models.layers import PatchEmbed3d


class ViT3d(VisionTransformer):
    def __init__(
        self,
        img_size: int | tuple[int, int, int] = 224,
        patch_size: int | tuple[int, int, int] = 16,
        in_chans: int = 3,
        num_classes: int = 1000,
        global_pool: Literal["", "avg", "avgmax", "max", "token", "map"] = "token",
        embed_dim: int = 768,
        depth: int = 12,
        num_heads: int = 12,
        mlp_ratio: float = 4.0,
        qkv_bias: bool = True,
        qk_norm: bool = False,
        proj_bias: bool = True,
        init_values: Optional[float] = None,
        class_token: bool = True,
        pos_embed: str = "learn",
        no_embed_class: bool = False,
        reg_tokens: int = 0,
        pre_norm: bool = False,
        final_norm: bool = True,
        fc_norm: Optional[bool] = None,
        dynamic_img_size: bool = False,
        dynamic_img_pad: bool = False,
        drop_rate: float = 0.0,
        pos_drop_rate: float = 0.0,
        patch_drop_rate: float = 0.0,
        proj_drop_rate: float = 0.0,
        attn_drop_rate: float = 0.0,
        drop_path_rate: float = 0.0,
        weight_init: Literal["skip", "jax", "jax_nlhb", "moco", ""] = "",
        fix_init: bool = False,
        embed_layer: Callable = PatchEmbed3d,
        embed_norm_layer: Optional[LayerType] = None,
        norm_layer: Optional[LayerType] = None,
        act_layer: Optional[LayerType] = None,
        block_fn: Type[nn.Module] = Block,
        mlp_layer: Type[nn.Module] = Mlp,
    ) -> None:
        super().__init__(
            img_size,
            patch_size,
            in_chans,
            num_classes,
            global_pool,
            embed_dim,
            depth,
            num_heads,
            mlp_ratio,
            qkv_bias,
            qk_norm,
            proj_bias,
            init_values,
            class_token,
            pos_embed,
            no_embed_class,
            reg_tokens,
            pre_norm,
            final_norm,
            fc_norm,
            dynamic_img_size,
            dynamic_img_pad,
            drop_rate,
            pos_drop_rate,
            patch_drop_rate,
            proj_drop_rate,
            attn_drop_rate,
            drop_path_rate,
            weight_init,
            fix_init,
            embed_layer,
            embed_norm_layer,
            norm_layer,
            act_layer,
            block_fn,
            mlp_layer,
        )

    @override
    def _pos_embed(self, x: Tensor) -> Tensor:
        """
        Apply positional embedding to the input tensor.

        Parameters
        ----------
        x : Tensor
            Input tensor of shape (B, N, C) or (B, D, H, W, C) if dynamic_img_size is True.

        Returns
        -------
        Tensor
            Tensor with positional embedding applied, of shape (B, N+tokens, C) where tokens
            is the number of prefix tokens (class token, register tokens).

        Notes
        -----
        This method handles dynamic image sizes by resampling the positional embedding
        to match the current feature map size. It also handles the addition of class tokens
        and register tokens.
        """
        if self.pos_embed is None:
            return x.view(x.shape[0], -1, x.shape[-1])

        if self.dynamic_img_size:
            if len(x.shape) != 5:
                raise ValueError(
                    "Shape of embedded patches should be BDHWC on dynamic resizing. "
                    "Disable flattening in PatchEmbed layer."
                )

            _, D, H, W, _ = x.shape
            pos_embed = self.resample_positional_embedding_3d(
                self.pos_embed,
                new_size=(D, H, W),
                old_size=self.patch_embed.grid_size,
                number_of_prefix_tokens=0 if self.no_embed_class else self.num_prefix_tokens,
            )
            x = x.view(x.shape[0], -1, x.shape[-1])
        else:
            pos_embed = self.pos_embed

        concatenation_tokens = []
        if self.cls_token is not None:
            concatenation_tokens.append(self.cls_token.expand(x.shape[0], -1, -1))
        if self.reg_token is not None:
            concatenation_tokens.append(self.reg_token.expand(x.shape[0], -1, -1))

        # This assumes that the concatenation tokens are not included within the positional
        # embedding, while the base class accounts for both cases.
        # In the old timm models this is the case, but the newer timm models don't do
        # this and also fix this in the conversion of ckpt's to timm standard. For example: DINOv2
        # has an positionally embedded class token, but this pos embed is added to the CLS_token
        # during conversion.
        x = x + pos_embed
        if concatenation_tokens:
            x = torch.cat(concatenation_tokens + [x], dim=1)

        return self.pos_drop(x)

    @staticmethod
    def resample_positional_embedding_3d(
        pos_embed: nn.Parameter,
        new_size: tuple[int, int, int],
        old_size: tuple[int, int, int],
        number_of_prefix_tokens: int = 0,
        interpolation="trilinear",
        antialias=False,
    ) -> nn.Parameter:
        """
        Resample 3D positional embeddings to a new size.

        Parameters
        ----------
        pos_embed : nn.Parameter
            Original positional embedding parameter of shape (1, N, C).
        new_size : tuple[int, int, int]
            Target size (D, H, W) for the positional embedding grid.
        old_size : tuple[int, int, int]
            Original size (D, H, W) of the positional embedding grid.
        number_of_prefix_tokens : int, optional
            Number of prefix tokens (like class tokens) in the positional embedding, by default 0.
        interpolation : str, optional
            Interpolation method to use, by default 'trilinear'.
        antialias : bool, optional
            Whether to use antialiasing during interpolation, by default False.

        Returns
        -------
        nn.Parameter
            Resampled positional embedding of shape (1, N', C) where N' corresponds
            to the number of patches in the new size plus any prefix tokens.

        Notes
        -----
        This method handles the resampling of positional embeddings when the input
        image size changes. It preserves any prefix tokens (like class tokens) and
        only resamples the spatial positional embeddings.
        """
        if old_size == new_size:
            return pos_embed

        if number_of_prefix_tokens:
            posemb_prefix, pos_embed = pos_embed[:, :number_of_prefix_tokens], pos_embed[:, number_of_prefix_tokens:]
        else:
            posemb_prefix, pos_embed = None, pos_embed

        # reshape positional embedding to a grid of 1 x N_PATCHES_D x N_PATCHES_H x N_PATCHES_W x embed_dim
        pos_embed = pos_embed.reshape(1, *old_size, -1)

        # Permute to 1 x embed_dim x N_PATCHES_D x N_PATCHES_H x N_PATCHES_W
        embed_dim = pos_embed.shape[-1]
        pos_embed = pos_embed.permute(0, 4, 1, 2, 3)
        # Convert to float for interpolation and convert back to old dtype after interpolation
        target_dtype = pos_embed.dtype
        pos_embed: torch.Tensor = nn.functional.interpolate(
            pos_embed.to(torch.float32),
            size=new_size,
            mode=interpolation,
            align_corners=antialias,
        ).to(dtype=target_dtype)

        # Permute back to 1 x N_PATCHES_D x N_PATCHES_H x N_PATCHES_W x embed_dim
        # and reshape to 1 x N_PATCHES x embed_dim
        pos_embed = pos_embed.permute(0, 2, 3, 4, 1).view(1, -1, embed_dim)

        # Add back extra (class, etc) prefix tokens
        if posemb_prefix is not None:
            pos_embed = torch.cat([posemb_prefix, pos_embed], dim=1)

        return pos_embed

    @override
    def forward_intermediates(
        self,
        x: Tensor,
        indices: int | list[int] | None = None,
        return_prefix_tokens: bool = False,
        norm: bool = False,
        stop_early: bool = False,
        output_fmt: str = "NCHW",
        intermediates_only: bool = False,
    ) -> list[Tensor] | tuple[Tensor, list[Tensor]]:
        if output_fmt not in ("NCHW", "NLC"):
            raise ValueError("Output format must be one of NCHW or NLC.")

        reshape = output_fmt == "NCHW"
        intermediates = []
        take_indices, max_index = feature_take_indices(len(self.blocks), indices)

        B, _, depth, height, width = x.shape
        x = self.patch_embed(x)
        x = self._pos_embed(x)
        x = self.patch_drop(x)
        x = self.norm_pre(x)

        if torch.jit.is_scripting() or not stop_early:
            blocks = self.blocks
        else:
            blocks = self.blocks[: max_index + 1]
        for i, blk in enumerate(blocks):
            x = blk(x)
            if i in take_indices:
                intermediates.append(self.norm(x) if norm else x)

        if self.num_prefix_tokens:
            prefix_tokens = [y[:, 0 : self.num_prefix_tokens] for y in intermediates]
            intermediates = [y[:, self.num_prefix_tokens :] for y in intermediates]
        else:
            prefix_tokens = None

        if reshape:
            # In the original Timm implementation dynamic padding is done within the model
            # Padding should be decoupled from model implementation, use a pad transform
            # To ensure that the input is divisible by patch size.
            D, H, W = (
                dim // patch_size for dim, patch_size in zip((depth, height, width), self.patch_embed.patch_size)
            )
            intermediates = [y.reshape(B, D, H, W, -1).permute(0, 4, 1, 2, 3).contiguous() for y in intermediates]
        if not torch.jit.is_scripting() and return_prefix_tokens and prefix_tokens is not None:
            intermediates = list(zip(intermediates, prefix_tokens))

        if intermediates_only:
            return intermediates

        x = self.norm(x)

        return x, intermediates


@register_model
def vit_small_patch16_224_3d(pretrained: bool = False, **kwargs):
    """ViT-Small-3d (ViT-S-3d/16)"""
    model_args = {
        "img_size": (16, 224, 224),
        "patch_size": 16,
        "embed_dim": 384,
        "depth": 12,
        "num_heads": 6,
        "embed_layer": PatchEmbed3d,
    }
    model = build_model_with_cfg(
        ViT3d,
        "vit_small_patch16_224_3d",
        pretrained=pretrained,
        pretrained_strict=False,
        # In the default _create_vision_transformer this is set to a checkpoint_filter_fn
        # This checkpoint filter fn assumes the PatchEmbed layer is a Conv2d
        pretrained_filter_fn=None,
        **dict(model_args, **kwargs),
    )
    return model


@register_model
def vit_small_patch8x16x16_224_3d(pretrained: bool = False, **kwargs):
    """ViT-Small-3d (ViT-S-3d/8x16x16)"""
    model_args = {
        "img_size": (24, 224, 224),
        "patch_size": (8, 16, 16),
        "embed_dim": 384,
        "depth": 12,
        "num_heads": 6,
        "embed_layer": PatchEmbed3d,
    }
    model = build_model_with_cfg(
        ViT3d,
        "vit_small_patch8x16x16_224_3d",
        pretrained=pretrained,
        pretrained_strict=False,
        # In the default _create_vision_transformer this is set to a checkpoint_filter_fn
        # This checkpoint filter fn assumes the PatchEmbed layer is a Conv2d
        pretrained_filter_fn=None,
        **dict(model_args, **kwargs),
    )
    return model


@register_model
def vit_base_patch16_224_3d(pretrained: bool = False, **kwargs):
    """ViT-Base-3d (ViT-B-3d/16)"""
    model_args = {
        "img_size": (16, 224, 224),
        "patch_size": 16,
        "embed_dim": 768,
        "depth": 12,
        "num_heads": 12,
        "embed_layer": PatchEmbed3d,
    }
    model = build_model_with_cfg(
        ViT3d,
        "vit_base_patch16_224_3d",
        pretrained=pretrained,
        pretrained_strict=False,
        # In the default _create_vision_transformer this is set to a checkpoint_filter_fn
        # This checkpoint filter fn assumes the PatchEmbed layer is a Conv2d
        pretrained_filter_fn=None,
        **dict(model_args, **kwargs),
    )
    return model
