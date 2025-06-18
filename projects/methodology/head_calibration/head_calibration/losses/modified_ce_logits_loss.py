from typing import Any

import torch
from head_calibration.losses.loss_factory import Loss
from torch.nn.functional import binary_cross_entropy_with_logits as bce_with_logits


class ModifiedCEWithLogitsLoss(Loss):
    def __init__(self, weight: float = 1.0, beta: float = 1.0):
        super().__init__(weight)
        self.beta = beta

    def __call__(self, input: torch.Tensor, batch: dict[str, Any]):
        scaled_logits = input * self.beta
        loss = bce_with_logits(scaled_logits, batch["labels"])
        return loss

    @property
    def name(self) -> str:
        return "modified_ce_with_logits_loss"
