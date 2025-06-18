# Copyright (c) Meta Platforms, Inc. and affiliates.
#
# This source code is licensed under the Apache License, Version 2.0
# found in the LICENSE file in the root directory of this source tree.
#
# Modified by NKI-AI4Oncology, 04-2025
from functools import partial
import torch
import torch.nn as nn
from dinov2.layers import NestedTensorBlock as Block, PatchEmbed3d, MemEffAttention
from typing_extensions import override
from dinov2.models.vision_transformer import DinoVisionTransformer


class DinoVisionTransformer3d(DinoVisionTransformer):
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
        init_values: float | None = None,
        embed_layer: nn.Module = PatchEmbed3d,
        act_layer: nn.Module = nn.GELU,
        block_fn: nn.Module = Block,
        ffn_layer: str = "mlp",
        block_chunks: int = 1,
        num_register_tokens: int = 0,
        interpolate_antialias: bool = False,
        interpolate_offset: float = 0.1,
    ) -> None:
        super().__init__(
            img_size,
            patch_size,
            in_chans,
            embed_dim,
            depth,
            num_heads,
            mlp_ratio,
            qkv_bias,
            ffn_bias,
            proj_bias,
            drop_path_rate,
            drop_path_uniform,
            init_values,
            embed_layer,
            act_layer,
            block_fn,
            ffn_layer,
            block_chunks,
            num_register_tokens,
            interpolate_antialias,
            interpolate_offset,
        )
        self._patch_depth, self._patch_width, self._patch_height = (
            (self.patch_size,) * 3 if isinstance(self.patch_size, int) else tuple(self.patch_size)
        )
        img_size = (img_size,) * 3 if isinstance(img_size, int) else tuple(img_size)

        self._num_patches_d, self._num_patches_h, self._num_patches_w = (
            img_size[0] // self._patch_depth,
            img_size[1] // self._patch_height,
            img_size[2] // self._patch_width,
        )

    @override
    def _interpolate_pos_encoding(
        self, embeddings: torch.Tensor, img_width: int, img_height: int, img_depth: int
    ) -> torch.Tensor:
        """
        Interpolates the positional embedding to allow for different image volume sizes.

        Adapted from:
          https://github.com/huggingface/transformers/blob/main/src/transformers/models/dinov2/modeling_dinov2.py

        Parameters
        ----------
        embeddings : torch.Tensor
            The input embeddings tensor.
        img_width : int
            Width of the input image volume.
        img_height : int
            Height of the input image volume.
        img_depth : int
            Depth of the input image volume.

        Returns
        -------
        torch.Tensor
            The interpolated positional embeddings.
        """
        num_patches = embeddings.shape[1] - 1
        num_positions = self.pos_embed.shape[1] - 1

        if num_patches == num_positions and img_width == img_height and img_depth == self._patch_depth:
            return self.pos_embed

        class_pos_embed = self.pos_embed[:, :1]
        patch_pos_embed = self.pos_embed[:, 1:]

        new_depth = img_depth // self._patch_depth
        new_height = img_height // self._patch_height
        new_width = img_width // self._patch_width

        patch_pos_embed = patch_pos_embed.reshape(
            1, self._num_patches_d, self._num_patches_h, self._num_patches_w, self.embed_dim
        )
        # Permute to B x embed_dim x depth x height x width
        patch_pos_embed = patch_pos_embed.permute(0, 4, 1, 2, 3)
        target_dtype = patch_pos_embed.dtype
        patch_pos_embed: torch.Tensor = nn.functional.interpolate(
            patch_pos_embed.to(torch.float32),
            size=(new_depth, new_height, new_width),
            mode="trilinear",
            align_corners=False,
        ).to(dtype=target_dtype)

        patch_pos_embed = patch_pos_embed.permute(0, 2, 3, 4, 1).view(1, -1, self.embed_dim)

        return torch.cat((class_pos_embed, patch_pos_embed), dim=1)

    @override
    def _prepare_tokens_with_masks(self, x: torch.Tensor, masks: torch.Tensor | None = None) -> torch.Tensor:
        """Prepare tokens for the Transformer by applying masks and positional encodings.

        This method processes the input tensor by:
        1. Converting the input volume into patch embeddings
        2. Applying masks if provided (replacing masked patches with a learnable mask token)
        3. Prepending the class token
        4. Adding positional encodings (interpolated to match the input dimensions)
        5. Adding register tokens if configured

        Parameters
        ----------
        x : torch.Tensor
            Input tensor with shape (batch_size, channels, depth, height, width)
        masks : torch.Tensor | None, optional
            Binary mask tensor where True indicates positions to be masked,
            by default None

        Returns
        -------
        torch.Tensor
            Processed token sequence ready for the transformer encoder,
            with shape (batch_size, num_tokens, embed_dim)
        """
        img_depth, img_width, img_height = x.shape[-3:]
        x = self.patch_embed(x)
        if masks is not None:
            x = torch.where(masks.unsqueeze(-1), self.mask_token.to(x.dtype).unsqueeze(0), x)

        x = torch.cat((self.cls_token.expand(x.shape[0], -1, -1), x), dim=1)
        x = x + self._interpolate_pos_encoding(x, img_width, img_height, img_depth)

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


def vit_small_3d(
    patch_size: int | tuple[int] = [16, 16, 16], num_register_tokens: int = 4, **kwargs
) -> DinoVisionTransformer3d:
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


def vit_base_3d(
    patch_size: int | tuple[int] = [16, 16, 16], num_register_tokens: int = 4, **kwargs
) -> DinoVisionTransformer3d:
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


def vit_large_3d(
    patch_size: int | tuple[int] = [16, 16, 16], num_register_tokens: int = 4, **kwargs
) -> DinoVisionTransformer3d:
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
