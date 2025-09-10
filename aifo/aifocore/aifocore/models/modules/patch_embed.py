# Copyright (c) Meta Platforms, Inc. and affiliates.
# Copyright 2020 Ross Wightman.
# Copyright 2025 AI for Oncology Research Group. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#     http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# References:
#   https://github.com/facebookresearch/dino/blob/master/vision_transformer.py
#   https://github.com/rwightman/pytorch-image-models/tree/master/timm/layers/patch_embed.py
from typing import Callable, Optional

import torch
from torch import nn

from aifocore.models.utils.collections import ensure_tuple


class PatchEmbed(nn.Module):
    """Patch embedding layer for Vision Transformers for 2D images.

    This layer divides the input image into patches and projects them into a higher-dimensional space.

    Parameters
    ----------
    img_size : int or tuple[int, int], optional
        Size of the input image. If an integer is provided, it is assumed to be square (img_size, img_size).
        If a tuple is provided, it should be of the form (height, width), by default 224.
    patch_size : int or tuple[int, int], optional
        Size of the patches to be extracted from the input image. If an integer is provided, it is assumed to be square
        (patch_size, patch_size). If a tuple is provided, it should be of the form (height, width), by default 16.
    in_chans : int, optional
        Number of input channels in the image, by default 3 (for RGB images).
    embed_dim : int, optional
        Dimension of the embedding space to which the patches will be projected, by default 768.
    norm_layer : Callable, optional
        Normalization layer to apply to the embeddings, by default None. If None, no normalization is applied.
    flatten_embedding : bool, optional
        Whether to flatten the embedding output, by default True.
    """

    def __init__(
        self,
        img_size: int | tuple[int, int] = 224,
        patch_size: int | tuple[int, int] = 16,
        in_chans: int = 3,
        embed_dim: int = 768,
        norm_layer: Optional[Callable] = None,
        flatten_embedding: bool = True,
    ) -> None:
        """Inits :class:`PatchEmbed`.

        Parameters
        ----------
        img_size : int or tuple[int, int], optional
            Size of the input image. If an integer is provided, it is assumed to be square (img_size, img_size).
            If a tuple is provided, it should be of the form (height, width), by default 224.
        patch_size : int or tuple[int, int], optional
            Size of the patches to be extracted from the input image. If an integer is provided, it is assumed to be square
            (patch_size, patch_size). If a tuple is provided, it should be of the form (height, width), by default 16.
        in_chans : int, optional
            Number of input channels in the image, by default 3 (for RGB images).
        embed_dim : int, optional
            Dimension of the embedding space to which the patches will be projected, by default 768.
        norm_layer : Callable, optional
            Normalization layer to apply to the embeddings, by default None. If None, no normalization is applied.
        flatten_embedding : bool, optional
            Whether to flatten the embedding output, by default True.
        """
        super().__init__()

        image_HW = ensure_tuple(img_size, n=2)
        patch_HW = ensure_tuple(patch_size, n=2)
        patch_grid_size = (
            image_HW[0] // patch_HW[0],
            image_HW[1] // patch_HW[1],
        )

        self.img_size = image_HW
        self.patch_size = patch_HW
        self.patches_resolution = patch_grid_size
        self.num_patches = patch_grid_size[0] * patch_grid_size[1]

        self.in_chans = in_chans
        self.embed_dim = embed_dim

        self.flatten_embedding = flatten_embedding

        self.proj = nn.Conv2d(in_chans, embed_dim, kernel_size=patch_HW, stride=patch_HW)
        self.norm = norm_layer(embed_dim) if norm_layer else nn.Identity()

    def forward(self, x: torch.Tensor) -> torch.Tensor:
        """Forward pass of :class:`PatchEmbed`.

        Parameters
        ----------
        x : torch.Tensor
            Input tensor of shape (B, C, H, W) where B is the batch size, C is the number of channels,
            H is the height, and W is the width of the input image.

        Raises
        ------
        ValueError
            If the input image dimensions are not compatible with the patch size.
        """
        _, _, H, W = x.shape
        patch_H, patch_W = self.patch_size
        if H % patch_H != 0:
            raise ValueError(f"Input image height {H} is not a multiple of patch height {patch_H}")
        if W % patch_W != 0:
            raise ValueError(f"Input image width {W} is not a multiple of patch width: {patch_W}")

        x = self.proj(x)  # B C H W
        H, W = x.size(2), x.size(3)
        x = x.flatten(2).transpose(1, 2)  # B HW C

        x = self.norm(x)
        if not self.flatten_embedding:
            x = x.reshape(-1, H, W, self.embed_dim)  # B H W C
        return x

    def flops(self) -> float:
        """Calculate the number of floating point operations (FLOPs) for the patch embedding layer.

        Returns
        -------
        float
            The number of FLOPs for the patch embedding layer.
        """
        Ho, Wo = self.patches_resolution
        flops = Ho * Wo * self.embed_dim * self.in_chans * (self.patch_size[0] * self.patch_size[1])
        if not isinstance(self.norm, nn.Identity):
            flops += Ho * Wo * self.embed_dim
        return flops


class PatchEmbed3d(nn.Module):
    """Patch embedding layer for Vision Transformers for 3D images.

    This layer divides the input 3D image volume into patches and projects them into a higher-dimensional space.

    Parameters
    ----------
    img_size : int or tuple[int, int, int], optional
        Size of the input image volume. If an integer is provided, it is assumed to be cubic (img_size, img_size, img_size).
        If a tuple is provided, it should be of the form (depth, height, width), by default 224.
    patch_size : int or tuple[int, int, int], optional
        Size of the patches to be extracted from the input image volume. If an integer is provided, it is assumed to be cubic
        (patch_size, patch_size, patch_size). If a tuple is provided, it should be of the form (depth, height, width), by default 16.
    in_chans : int, optional
        Number of input channels in the image volume, by default 3 (for RGB images).
    embed_dim : int, optional
        Dimension of the embedding space to which the patches will be projected, by default 768.
    norm_layer : Callable, optional
        Normalization layer to apply to the embeddings, by default None. If None, no normalization is applied.
    flatten_embedding : bool, optional
        Whether to flatten the embedding output, by default True.
    """

    def __init__(
        self,
        img_size: int | tuple[int, int, int] = 224,
        patch_size: int | tuple[int, int, int] = 16,
        in_chans: int = 3,
        embed_dim: int = 768,
        norm_layer: Optional[Callable] = None,
        flatten_embedding: bool = True,
    ) -> None:
        """Inits :class:`PatchEmbed3d`.

        Parameters
        ----------
        img_size : int or tuple[int, int, int], optional
            Size of the input image volume. If an integer is provided, it is assumed to be cubic
            (img_size, img_size, img_size).
            If a tuple is provided, it should be of the form (depth, height, width), by default 224.
        patch_size : int or tuple[int, int, int], optional
            Size of the patches to be extracted from the input image volume. If an integer is provided, it is
            assumed to be cubic (patch_size, patch_size, patch_size). If a tuple is provided, it should be of the
            form (depth, height, width), by default 16.
        in_chans : int, optional
            Number of input channels in the image volume, by default 3 (for RGB images).
        embed_dim : int, optional
            Dimension of the embedding space to which the patches will be projected, by default 768.
        norm_layer : Callable, optional
            Normalization layer to apply to the embeddings, by default None. If None, no normalization is applied.
        flatten_embedding : bool, optional
            Whether to flatten the embedding output, by default True.
        """
        super().__init__()

        image_DHW = ensure_tuple(img_size, n=3)
        patch_DHW = ensure_tuple(patch_size, n=3)

        patch_grid_size = (
            image_DHW[0] // patch_DHW[0],
            image_DHW[1] // patch_DHW[1],
            image_DHW[2] // patch_DHW[2],
        )

        self.img_size = image_DHW
        self.patch_size = patch_DHW
        self.patches_resolution = patch_grid_size
        self.num_patches = patch_grid_size[0] * patch_grid_size[1] * patch_grid_size[2]

        self.in_chans = in_chans
        self.embed_dim = embed_dim

        self.flatten_embedding = flatten_embedding
        self.proj = nn.Conv3d(in_chans, embed_dim, kernel_size=patch_DHW, stride=patch_DHW)
        self.norm = norm_layer(embed_dim) if norm_layer else nn.Identity()

    def forward(self, x: torch.Tensor) -> torch.Tensor:
        """Forward pass of :class:`PatchEmbed3d`.

        Parameters
        ----------
        x : torch.Tensor
            Input tensor of shape (B, C, D, H, W) where B is the batch size, C is the number of channels,
            D is the depth, H is the height, and W is the width of the input volume.

        Raises
        ------
        ValueError
            If the input volume dimensions are not compatible with the patch size.
        """
        _, _, D, H, W = x.shape
        patch_D, patch_H, patch_W = self.patch_size
        if D % patch_D != 0:
            raise ValueError(f"Input volume depth {D} is not a multiple of patch depth {patch_D}")
        if H % patch_H != 0:
            raise ValueError(f"Input volume height {H} is not a multiple of patch height {patch_H}")
        if W % patch_W != 0:
            raise ValueError(f"Input volume width {W} is not a multiple of patch width {patch_W}")

        x = self.proj(x)  # B C D H W
        D, H, W = x.size(2), x.size(3), x.size(4)
        x = x.flatten(2).transpose(1, 2)  # B (DHW) C

        x = self.norm(x)
        if not self.flatten_embedding:
            x = x.reshape(-1, D, H, W, self.embed_dim)  # B D H W C
        return x

    def flops(self) -> float:
        """Calculate the number of floating point operations (FLOPs) for the patch embedding 3D layer.

        Returns
        -------
        float
            The number of FLOPs for the patch embedding layer.
        """
        Do, Ho, Wo = self.patches_resolution
        flops = (
            Do
            * Ho
            * Wo
            * self.embed_dim
            * self.in_chans
            * (self.patch_size[0] * self.patch_size[1] * self.patch_size[2])
        )
        if not isinstance(self.norm, nn.Identity):
            flops += Do * Ho * Wo * self.embed_dim
        return flops
