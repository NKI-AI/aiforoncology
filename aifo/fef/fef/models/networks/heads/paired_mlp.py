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

from typing import Type

import torch
import torch.nn as nn


class PairedMLP(nn.Module):
    """A Multi-layer Perceptron (MLP) network that takes two inputs and outputs a single prediction."""

    def __init__(
        self,
        input_size: int,
        output_size: int,
        hidden_layer_sizes: tuple[int, ...] | None = None,
        hidden_activation_fn: Type[torch.nn.Module] | None = nn.ReLU,
        output_activation_fn: Type[torch.nn.Module] | None = None,
        dropout: float = 0.0,
    ) -> None:
        """Initializes the MLP.

        Parameters
        ----------
        input_size : int
            The number of input features.
        output_size : int
            The number of output features.
        hidden_layer_sizes : tuple[int, ...], optional
            A list specifying the number of units in each hidden layer.
        dropout : float, default=0.0
            Dropout probability for hidden layers.
        hidden_activation_fn : Type[torch.nn.Module], optional
            Activation function to use for hidden layers. Default is ReLU.
        output_activation_fn : Type[torch.nn.Module], optional
            Activation function to use for the output layer. Default is None.
        """
        super().__init__()

        self.flatten = nn.Flatten()
        self.input_size = input_size
        self.output_size = output_size
        self.hidden_layer_sizes = hidden_layer_sizes if hidden_layer_sizes is not None else ()
        self.hidden_activation_fn = hidden_activation_fn
        self.output_activation_fn = output_activation_fn
        self.dropout = dropout

        self._network = self._build_network()

    def _build_network(self) -> nn.Sequential:
        """Builds the neural network's layers and returns a nn.Sequential container."""
        layers = []
        prev_size = self.input_size
        for size in self.hidden_layer_sizes:
            layers.append(nn.Linear(prev_size, size))
            if self.hidden_activation_fn is not None:
                layers.append(self.hidden_activation_fn())
            if self.dropout > 0:
                layers.append(nn.Dropout(self.dropout))
            prev_size = size

        layers.append(nn.Linear(prev_size, self.output_size))
        if self.output_activation_fn is not None:
            layers.append(self.output_activation_fn())

        return nn.Sequential(*layers)

    def forward(self, x: torch.Tensor) -> torch.Tensor:
        """Defines the forward pass of the MLP.

        Parameters
        ----------
        x : torch.Tensor
            The input tensor.

        Returns
        -------
        torch.Tensor
            The output of the network.
        """
        x = self.flatten(x)
        for layer in self._network:
            x = layer(x)
        return x
