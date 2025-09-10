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

from enum import Enum
import torch
import torch.nn as nn


class ViTEmbedMode(str, Enum):
    """WORK IN PROGRESS
    Embedding modes for Vision Transformers"""

    CLS_ONLY = "cls_only"
    PATCH_ONLY = "patch_only"
    MEAN = "mean"
    CONCAT_MEAN = "concat_mean"
    CONCAT = "concat"


class ViTEmbed(nn.Module):
    def __init__(self, embedding_mode: ViTEmbedMode) -> None:
        super().__init__()

        self.__embedding_mode = embedding_mode

        if self.__embedding_mode == ViTEmbedMode.CLS_ONLY:
            embed_fn = self.cls_only
        elif self.__embedding_mode == ViTEmbedMode.PATCH_ONLY:
            embed_fn = self.patch_only
        elif self.__embedding_mode == ViTEmbedMode.MEAN:
            embed_fn = self.mean
        elif self.__embedding_mode == ViTEmbedMode.CONCAT_MEAN:
            embed_fn = self.concat_mean
        elif self.__embedding_mode == ViTEmbedMode.CONCAT:
            embed_fn = self.concat

        self.embed_fn = embed_fn

    def forward(self, cls_token: torch.Tensor, patch_tokens: torch.Tensor) -> torch.Tensor:
        """
        Returns the final vision transformer embedding.

        Parameters
        ----------
        cls_token :
            class token of shape (B, feature_dim)
        patch_tokens :
            patch tokens of (B, num_patches, feature_dim)

        Returns
        -------
        output: torch.Tensor
            Embedding of shape [B, feature_dim]

        """
        output = self.embed_fn(cls_token=cls_token, patch_tokens=patch_tokens)
        return output

    @property
    def embedding_mode(self) -> str:
        return self.__embedding_mode.value

    @property
    def dim_factor(self) -> int:
        """
        Returns the scaling factor by which the output feature dimensionality will increase
        when using a certain embedding method.
        E.g. the concat method will make the output dimensionality twice as big.
        """
        if self.embedding_mode == ViTEmbedMode.CONCAT:
            return 2
        else:
            return 1

    @staticmethod
    def concat(pre_embeddings: torch.Tensor, post_embeddings: torch.Tensor) -> torch.Tensor:
        output = torch.cat([pre_embeddings, post_embeddings], dim=1)
        return output
