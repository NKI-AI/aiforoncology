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
import torch
import torch.nn as nn


class DropPath(nn.Module):
    """Stochastic drop paths per sample for residual blocks based on [1].

    Parameters
    ----------
    drop_prob : float, optional
        Drop path probability, by default 0.0
    scale_by_keep : bool, optional
        Whether to scale outputs by non-dropped probability, by default True

    Raises
    ------
    ValueError
        If drop probability is not between 0 and 1

    References
    ----------
    [1] https://github.com/rwightman/pytorch-image-models
    """

    def __init__(self, drop_prob: float = 0.0, scale_by_keep: bool = True) -> None:
        """Inits :class:`DropPath`.

        Parameters
        ----------
        drop_prob : float, optional
            Drop path probability, by default 0.0
        scale_by_keep : bool, optional
            Whether to scale outputs by non-dropped probability, by default True

        Raises
        ------
        ValueError
            If drop probability is not between 0 and 1
        """
        super().__init__()
        self.drop_prob = drop_prob
        self.scale_by_keep = scale_by_keep

        if not (0 <= drop_prob <= 1):
            raise ValueError("Drop path probability should be between 0 and 1.")

    @staticmethod
    def drop_path(x: torch.Tensor, drop_prob: float = 0.0, training: bool = False, scale_by_keep: bool = True):
        """Apply drop path to input tensor.

        Parameters
        ----------
        x : torch.Tensor
            Input tensor
        drop_prob : float, optional
            Drop path probability, by default 0.0
        training : bool, optional
            Whether in training mode, by default False
        scale_by_keep : bool, optional
            Whether to scale by non-dropped probability, by default True

        Returns
        -------
        torch.Tensor
            Output tensor after applying drop path
        """
        if drop_prob == 0.0 or not training:
            return x
        keep_prob = 1 - drop_prob
        shape = (x.shape[0],) + (1,) * (x.ndim - 1)
        random_tensor = x.new_empty(shape).bernoulli_(keep_prob)
        if keep_prob > 0.0 and scale_by_keep:
            random_tensor.div_(keep_prob)
        return x * random_tensor

    def forward(self, x):
        return self.drop_path(x, self.drop_prob, self.training, self.scale_by_keep)
