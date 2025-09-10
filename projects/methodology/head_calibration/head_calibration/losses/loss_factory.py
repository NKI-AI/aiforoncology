from abc import ABC, abstractmethod
from typing import Any

import torch
import torch.nn as nn


class Loss(ABC, nn.Module):
    def __init__(self, weight: float = 1.0):
        super().__init__()
        self.weight = weight

    @abstractmethod
    def __call__(self, input: torch.Tensor, batch: dict[str, Any]):
        pass

    @property
    @abstractmethod
    def name(self) -> str:
        pass


class LossFactory(nn.Module):
    """Loss factory to construct the total loss."""

    def __init__(
        self,
        losses: list[dict[str, "Loss"]],
    ):
        """
        ----------
        Parameters
        ----------
        losses : list
            List of losses which are functions which accept `(input, batch, weight)`. batch will be a dict(str,Any) containing
            for instance the labels and any other needed data. The weight will be applied per loss.
        """
        super().__init__()
        self.losses = nn.ModuleDict()
        for loss_dict in losses:
            for name, loss_obj in loss_dict.items():
                self.losses[name] = loss_obj

    def forward(self, input: torch.Tensor, batch: dict[str, Any]) -> dict[str, torch.Tensor]:
        total_loss = torch.tensor(0.0, device=input.device)
        detailed_losses = {}

        for name, loss in self.losses.items():
            current_loss = loss(input, batch) * loss.weight
            detailed_losses[name] = current_loss
            total_loss += current_loss

        detailed_losses["total_loss"] = total_loss
        return detailed_losses

    def __str__(self) -> str:
        loss_names = ", ".join([loss.name for loss in self.losses])
        return super().__str__() + f" with losses: {loss_names}"
