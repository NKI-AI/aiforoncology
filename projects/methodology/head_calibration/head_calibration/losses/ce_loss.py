from typing import Any

import torch
from head_calibration.losses.loss_factory import Loss
from torch.nn.functional import cross_entropy


class CrossEntropyLoss(Loss):
    def __init__(self, weight: float = 1.0):
        super().__init__(weight)
        self.cross_entropy = cross_entropy

    def __call__(self, input: torch.Tensor, batch: dict[str, Any]):
        loss = self.cross_entropy(input, batch["labels"])
        return loss

    @property
    def name(self) -> str:
        return "cross_entropy_loss"
