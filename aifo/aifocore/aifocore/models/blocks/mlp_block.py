# Copyright (c) MONAI Consortium
# Copyright 2025 AI for Oncology Research Group. All Rights Reserved.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#     http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

from __future__ import annotations
from typing import Any

import torch
import torch.nn as nn


class MLPBlock(nn.Module):
    """A multi-layer perceptron block, based on: "Dosovitskiy et al.,
    An Image is Worth 16x16 Words: Transformers for Image Recognition at Scale <https://arxiv.org/abs/2010.11929>"

    Parameters
    ----------
    hidden_size : int
        Dimension of hidden layer.
    mlp_dim : int
        Dimension of feedforward layer. If 0, `hidden_size` will be used.
    dropout_rate : float, optional
        Fraction of the input units to drop, by default 0.0
    activation_function : nn.Module, optional
        Activation type and arguments, by default nn.GELU
    """

    def __init__(
        self,
        hidden_size: int,
        mlp_dim: int,
        dropout_rate: float = 0.0,
        activation_function: nn.Module = nn.GELU,
        **activation_args: dict[str, Any],
    ) -> None:
        """Inits :class:`MLPBlock`

        Parameters
        ----------
        hidden_size : int
            Dimension of hidden layer.
        mlp_dim : int
            Dimension of feedforward layer. If 0, `hidden_size` will be used.
        dropout_rate : float, optional
            Fraction of the input units to drop, by default 0.0
        activation_function : nn.Module, optional
            Activation type and arguments, by default nn.GELU
        """

        super().__init__()

        if not (0 <= dropout_rate <= 1):
            raise ValueError(f"dropout_rate should be between 0 and 1, but is set to '{dropout_rate}'.")
        mlp_dim = mlp_dim or hidden_size
        self.linear1 = nn.Linear(hidden_size, mlp_dim)
        self.linear2 = nn.Linear(mlp_dim, hidden_size)
        self.fn = activation_function(**activation_args)
        self.drop = nn.Dropout(dropout_rate)

    def forward(self, x: torch.Tensor) -> torch.Tensor:
        """Forward pass of MLPBlock.

        Parameters
        ----------
        x : torch.Tensor
            Input tensor of shape (..., hidden_size)

        Returns
        -------
        torch.Tensor
            Output tensor of shape (..., hidden_size) after applying two linear
            transformations with activation and dropout in between
        """
        x = self.fn(self.linear1(x))
        x = self.drop(x)
        x = self.linear2(x)
        x = self.drop(x)
        return x
