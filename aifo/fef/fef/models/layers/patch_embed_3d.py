# Copyright 2025 AI for Oncology Research Group. All Rights Reserved.
#
# Modified from: https://github.com/facebookresearch/dinov2/blob/main/dinov2/layers/patch_embed.py
# Copyright (c) Meta Platforms, Inc. and affiliates.
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
from fef.utils.helpers import make_3tuple
from typing import Callable, Optional, Union
import torch.nn as nn
from torch import Tensor


class PatchEmbed3d(nn.Module):
    """
    3D image to patch embedding.

    Transforms a 3D image to patch embedding: (B,C,D,H,W) -> (B,N,D)

    Parameters
    ----------
    img_size : Union[int, tuple[int, int, int]]
        Image size, either int for cube or tuple for specifying dimensions.
    patch_size : Union[int, tuple[int, int, int]]
        Patch token size, either int for cube or tuple for specifying dimensions.
    in_chans : int
        Number of input image channels.
    embed_dim : int
        Number of linear projection output channels.
    norm_layer : Optional[Callable]
        Normalization layer.
    strict_img_size : bool
        Determines whether to flatten the embedding output. If the image size is dynamic,
        it's useful to retain the original dimensions as information for later steps.

    Returns
    -------
    Tensor
        Patch embedding tensor.
    """

    def __init__(
        self,
        img_size: Union[int, tuple[int, int, int]] = (16, 224, 224),
        patch_size: Union[int, tuple[int, int, int]] = 16,
        in_chans: int = 1,
        embed_dim: int = 768,
        norm_layer: Optional[Callable] = None,
        strict_img_size: bool = True,
        **_,
    ) -> None:
        super().__init__()

        image_DHW = make_3tuple(img_size)
        patch_DHW = make_3tuple(patch_size)
        self._img_size = image_DHW
        self._patch_size = patch_DHW
        self._in_chans = in_chans
        self._embed_dim = embed_dim
        self._strict_img_size = strict_img_size

        self._grid_size = tuple([i // p for i, p in zip(self._img_size, self._patch_size)])
        self._num_patches = self._grid_size[0] * self._grid_size[1] * self._grid_size[2]

        self.proj = nn.Conv3d(in_chans, embed_dim, kernel_size=patch_DHW, stride=patch_DHW)
        self.norm = norm_layer(embed_dim) if norm_layer else nn.Identity()

    @property
    def grid_size(self):
        return self._grid_size

    @property
    def num_patches(self):
        return self._num_patches

    @property
    def patch_size(self):
        return self._patch_size

    def forward(self, x: Tensor) -> Tensor:
        _, _, D, H, W = x.shape
        patch_D, patch_H, patch_W = self._patch_size

        if D % patch_D != 0:
            raise ValueError(f"Input image depth {D} is not a multiple of patch depth {patch_D}")

        if H % patch_H != 0:
            raise ValueError(f"Input image height {H} is not a multiple of patch height {patch_H}")

        if W % patch_W != 0:
            raise ValueError(f"Input image width {W} is not a multiple of patch width {patch_W}")

        x = self.proj(x)
        D, H, W = x.size(2), x.size(3), x.size(4)
        x = x.flatten(2).transpose(1, 2)
        x = self.norm(x)

        if not self._strict_img_size:
            x = x.reshape(-1, D, H, W, self._embed_dim)

        return x
