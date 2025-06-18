from typing import Any

import torch
from head_calibration.losses.loss_factory import Loss
from torch.nn.functional import cross_entropy


class FocalLoss(Loss):
    def __init__(self, weight: float = 1.0, gamma=2.0, reduction="mean"):
        super().__init__(weight)
        self.gamma = gamma
        self.reduction = reduction
        self.cross_entropy = cross_entropy

    def __call__(self, input: torch.Tensor, batch: dict[str, Any]):
        ce_loss = self.cross_entropy(input, batch["labels"], reduction="none")
        pt = torch.exp(-ce_loss)  # Prevents nans when probability 0
        F_loss = (1 - pt) ** self.gamma * ce_loss

        if self.reduction == "mean":
            return torch.mean(F_loss)
        elif self.reduction == "sum":
            return torch.sum(F_loss)
        else:
            return F_loss

    @property
    def name(self) -> str:
        return "focal_loss"


if __name__ == "__main__":
    inputs = torch.randn(10, requires_grad=True)
    targets = torch.empty(10).random_(2)
    batch = {"labels": targets}

    criterion_f = FocalLoss(weight=1.0, gamma=0.0, reduction="mean")
    loss = criterion_f(inputs, batch)

    from head_calibration.losses.ce_loss import CrossEntropyLoss

    criterion_c = CrossEntropyLoss()
    loss_c = criterion_c(inputs, batch)

    print(f"Focal Loss: {loss}")
    print(f"Cross Entropy Loss: {loss_c}")  # these should be equal
