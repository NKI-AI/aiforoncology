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

"""Volume-level dice score metric for 3D medical image segmentation."""

from typing import Any, Optional

import torch
from torch import Tensor
from torchmetrics import Metric


class VolumeDiceMetric(Metric):
    """Computes dice score over 3D volumes by aggregating 2D slice predictions.

    This metric computes the dice score at the volume level by first aggregating
    all pixels across slices, then computing the dice score on the entire volume.
    """

    is_differentiable: bool = False
    higher_is_better: bool = True
    full_state_update: bool = False

    def __init__(
        self,
        num_classes: int,
        include_background: bool = True,
        reduction: str = "mean",
        **kwargs: Any,
    ) -> None:
        """Init :class:`VolumeDiceMetric`.

        Parameters
        ----------
        num_classes : int
            Number of classes to compute dice score for
        include_background : bool, default=True
            Whether to include background class in computation
        reduction : str, default="mean"
            How to reduce the dice scores across volumes. Options are:
            - "mean": Average dice scores across volumes
            - "none": Return dice scores for each volume
        """
        super().__init__(**kwargs)
        self.num_classes = num_classes
        self.include_background = include_background
        self.reduction = reduction

        self.volumes = {}  # patient_id -> {slice_idx -> (pred, target)}

        if reduction == "mean":
            for i in range(num_classes):
                if not include_background and i == 0:
                    continue
                self.add_state(f"volume_dice_class_{i}", default=torch.tensor(0.0), dist_reduce_fx="mean")

    def update(
        self,
        preds: Optional[Tensor] = None,
        target: Optional[Tensor] = None,
        metadata: Optional[dict[str, Any]] = None,
        **kwargs: Any,
    ) -> None:
        """Gather slices for each volume in a dictionary.

        Parameters
        ----------
        preds : Optional[Tensor]
            Predictions from the model, expected shape [B, C, H, W] or [B, H, W]
        target : Optional[Tensor]
            Ground truth targets, expected shape [B, C, H, W] or [B, H, W]
        metadata : Optional[dict[str, Any]]
            Metadata containing patient IDs and slice indices, expected to be a list of dictionaries
        **kwargs : Any
            Additional keyword arguments, can include 'predictions' and 'targets' for compatibility
        """
        # Handle both parameter names for compatibility
        predictions = preds if preds is not None else kwargs.get("predictions")
        targets = target if target is not None else kwargs.get("targets")

        if predictions is None or targets is None:
            raise ValueError("No predictions or targets found in input")

        if metadata is None:
            raise ValueError("No metadata found in input")

        batch_size = predictions.size(0)

        # Handle both single dictionary and list of dictionaries
        if isinstance(metadata, dict):
            # If metadata is a single dictionary, use it for all items in the batch
            metadata = [metadata] * batch_size

        for i in range(batch_size):
            item_metadata = metadata[i]
            if not item_metadata:
                raise ValueError(f"No metadata found for item {i}")

            patient_id = item_metadata["patient_id"]
            if isinstance(patient_id, list):
                patient_id = patient_id[0]

            slice_idx = item_metadata["slice_index"]
            if isinstance(slice_idx, Tensor):
                slice_idx = slice_idx.item()

            pred = predictions[i : i + 1]
            targ = targets[i : i + 1]

            # Only process if needed - check if input is logits/probabilities
            if pred.dim() == 4:  # If shape is [B, C, H, W]
                pred = pred.argmax(dim=1)
            if targ.dim() == 4:  # If shape is [B, C, H, W]
                targ = targ.argmax(dim=1)

            if patient_id not in self.volumes:
                self.volumes[patient_id] = {}

            self.volumes[patient_id][slice_idx] = (pred, targ)

    def compute(self) -> dict[str, Tensor]:
        """Compute dice scores for complete volumes."""
        if not self.volumes:
            # Return dummy values if no volumes have been processed
            return {
                f"volume_dice_class_{i}": torch.tensor(0.0)
                for i in range(self.num_classes)
                if self.include_background or i > 0
            }

        results = {}

        for vol_id, slices in self.volumes.items():
            try:
                sorted_indices = sorted(slices.keys())
                sorted_preds = [slices[idx][0] for idx in sorted_indices]
                sorted_targets = [slices[idx][1] for idx in sorted_indices]

                volume_preds = torch.stack(sorted_preds, dim=1).squeeze(0)
                volume_targets = torch.stack(sorted_targets, dim=1).squeeze(0)

                for class_idx in range(self.num_classes):
                    if not self.include_background and class_idx == 0:
                        continue

                    pred_mask = (volume_preds == class_idx).float()
                    target_mask = (volume_targets == class_idx).float()

                    intersection = (pred_mask * target_mask).sum()
                    total = pred_mask.sum() + target_mask.sum()
                    dice_score = 2.0 * intersection / (total + 1e-8)

                    if self.reduction == "none":
                        results[f"volume_{vol_id}_class_{class_idx}"] = dice_score
                    else:
                        if f"class_{class_idx}" not in results:
                            results[f"class_{class_idx}"] = []
                        results[f"class_{class_idx}"].append(dice_score)

            except RuntimeError as e:
                raise e

        final_results = {}
        if self.reduction == "mean":
            for class_idx in range(self.num_classes):
                if not self.include_background and class_idx == 0:
                    continue
                if f"class_{class_idx}" in results:
                    class_scores = torch.stack(results[f"class_{class_idx}"])
                    final_results[f"volume_dice_class_{class_idx}"] = class_scores.mean()
        else:
            final_results = results

        return final_results

    def reset(self) -> None:
        """Reset the metric state."""
        super().reset()
        self.volumes.clear()

    def forward(self, *args: Any, **kwargs: Any) -> Any:
        """First forward all input through cast logic."""
        self.update(*args, **kwargs)  # Calls our update method
        return self.compute() if self.compute_on_step else None
