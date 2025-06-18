# Copyright 2025 Kaiko. All Rights Reserved.
# Copyright 2025 AI for Oncology Research Group. All Rights Reserved.
#
# Based on EVA SemanticSegmentationModule:
#    https://github.com/kaiko-ai/eva/blob/main/src/eva/vision/models/modules/semantic_segmentation.py
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
from typing import Any, Callable

import torch
from eva.vision.models.modules.semantic_segmentation import SemanticSegmentationModule
from eva.core.metrics import structs as metrics_lib
from eva.core.models.modules.utils import batch_postprocess
from eva.core.models.modules.typings import INPUT_TENSOR_BATCH
from eva.vision.models.networks import decoders
from eva.vision.models.networks.decoders import segmentation
from eva.vision.models.networks.decoders.segmentation.typings import DecoderInputs
from lightning.pytorch.cli import LRSchedulerCallable, OptimizerCallable
from lightning.pytorch.utilities.types import STEP_OUTPUT
from monai.inferers.inferer import Inferer
from torch.optim.adamw import AdamW as AdamW
from torch import nn, optim
from torch.optim import lr_scheduler
from typing_extensions import override


class SemanticSegmentation3dModule(SemanticSegmentationModule):
    def __init__(
        self,
        decoder: decoders.Decoder | nn.Module,
        criterion: Callable[..., torch.Tensor],
        encoder: dict[str, Any] | Callable[[torch.Tensor], list[torch.Tensor]] | None = None,
        lr_multiplier_encoder: float = 0.0,
        inferer: Inferer | None = None,
        optimizer: OptimizerCallable = optim.AdamW,
        lr_scheduler: LRSchedulerCallable = lr_scheduler.ConstantLR,
        metrics: metrics_lib.MetricsSchema | None = None,
        postprocess: batch_postprocess.BatchPostProcess | None = None,
        save_decoder_only: bool = True,
    ) -> None:
        super().__init__(
            decoder,
            criterion,
            encoder,
            lr_multiplier_encoder,
            inferer,
            optimizer,
            lr_scheduler,
            metrics,
            postprocess,
            save_decoder_only,
        )
        self._image_size: tuple = (-1, -1, -1)

    @override
    def _batch_step(self, batch: INPUT_TENSOR_BATCH) -> STEP_OUTPUT:
        """Performs a model forward step and calculates the loss.

        Parameters
        ----------
        batch : INPUT_TENSOR_BATCH
            The desired batch to process.

        Returns
        -------
        STEP_OUTPUT
            The batch step output.
        """
        data, targets, metadata = INPUT_TENSOR_BATCH(*batch)
        # TODO: This is a hacky solution. The Decoder class expects an image size to interpolate the
        # predicted segmentation mask to the size of the target mask. However, when applying offline
        # learning the image_size is only represented within the target which is not available in the forward.
        # This has been addressed to Kaiko, the parent class is subject to change to integrate offline
        # training into the pipeline again.
        self._image_size = tuple(targets.shape[-3:])
        predictions = self(data)
        loss = self.criterion(predictions, targets)
        return {
            "loss": loss,
            "targets": targets,
            "predictions": predictions,
            "metadata": metadata,
        }

    @override
    def _forward_networks(self, tensor: torch.Tensor | list[torch.Tensor]) -> torch.Tensor:
        """Passes the input tensor through the encoder and decoder."""
        features = self.encoder(tensor) if self.encoder else tensor

        if not isinstance(tensor, list):
            tensor = [tensor]

        if isinstance(self.decoder, segmentation.Decoder):
            if not isinstance(features, list):
                raise ValueError(f"Expected a list of feature map tensors, got {type(features)}.")
            return self.decoder(DecoderInputs(features, self._image_size, tensor))
        return self.decoder(features)
