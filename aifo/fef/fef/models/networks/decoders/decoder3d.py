# Copyright 2025 Kaiko. All Rights Reserved.
# Copyright 2025 AI for Oncology Research Group. All Rights Reserved.
#
# Based on EVA decoder2d: https://github.com/kaiko-ai/eva/blob/main/src/eva/vision/models/networks/decoders/segmentation/decoder2d.py
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
from typing_extensions import override

import torch
import torch.nn as nn
import torch.nn.functional as F
from eva.vision.models.networks.decoders.segmentation import base
from eva.vision.models.networks.decoders.segmentation.typings import DecoderInputs


class Decoder3d(base.Decoder):
    def __init__(self, layers: nn.Module, combine_features: bool = True) -> None:
        super().__init__()

        self._layers = layers
        self._combine_features = combine_features

    @override
    def forward(self, decoder_inputs: DecoderInputs) -> torch.Tensor:
        """Maps the patch embeddings to a segmentation mask of the image size.

        Parameters
        ----------
        decoder_inputs : DecoderInputs
            Inputs required by the decoder.

        Returns
        -------
        torch.Tensor
            Tensor containing scores for all of the classes with shape
            (batch_size, n_classes, image_depth, image_height, image_width).
        """
        features, image_size, _ = DecoderInputs(*decoder_inputs)
        if self._combine_features:
            features = self._forward_features(features)
        logits = self._forward_head(features)
        return self._upscale(logits, image_size)

    def _forward_features(self, features: torch.Tensor | list[torch.Tensor]) -> torch.Tensor:
        """Forward function for multi-level feature maps to a single one.

        It will interpolate the features and concat them into a single tensor
        on the dimension axis of the hidden size.

        Parameters
        ----------
        features : torch.Tensor or list[torch.Tensor]
            List of multi-level image features of shape (batch_size,
            hidden_size, n_patches_depth, n_patches_height, n_patches_width).

        Returns
        -------
        torch.Tensor
            A tensor of shape (batch_size, hidden_size, n_patches_depth, n_patches_height, n_patches_width)
            which is feature map of the decoder head.

        Examples
        --------
        >>> features = [torch.Tensor(16, 384, 2, 14, 14), torch.Size(16, 384, 2, 14, 14)]
        >>> output = self._forward_features(features)
        >>> assert output.shape == torch.Size([16, 768, 2, 14, 14])
        """
        if isinstance(features, torch.Tensor):
            features = [features]
        if not isinstance(features, (list, tuple)) or features[0].ndim != 5:
            raise ValueError(
                "Input features should be a list of five (5) dimensional inputs of "
                "shape (batch_size, hidden_size, n_patches_depth, n_patches_height, n_patches_width)."
            )

        upsampled_features = [
            F.interpolate(
                input=embeddings,
                size=features[0].shape[2:],
                mode="trilinear",
                align_corners=False,
            )
            for embeddings in features
        ]
        return torch.cat(upsampled_features, dim=1)

    def _forward_head(self, patch_embeddings: torch.Tensor) -> torch.Tensor:
        """Forward of the decoder head.

        Parameters
        ----------
        patch_embeddings : torch.Tensor
            The patch embeddings tensor of shape
            (batch_size, hidden_size, n_patches_depth, n_patches_height, n_patches_width).

        Returns
        -------
        torch.Tensor
            The logits as a tensor (batch_size, n_classes, upscale_depth, upscale_height, upscale_width).
        """
        return self._layers(patch_embeddings)

    def _upscale(
        self,
        logits: torch.Tensor,
        image_size: tuple[int, int, int],
    ) -> torch.Tensor:
        """Upscales the calculated logits to the target image size.

        Parameters
        ----------
        logits : torch.Tensor
            The decoder outputs of shape (batch_size, n_classes, depth, height, width).
        image_size : tuple[int, int, int]
            The target image size (depth, height, width).

        Returns
        -------
        torch.Tensor
            Tensor containing scores for all of the classes with shape
            (batch_size, n_classes, image_depth, image_height, image_width).
        """
        return F.interpolate(logits, image_size, mode="trilinear")
