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

import numpy.typing as npt
import torch
from eva.core.data.transforms.dtype.array import ArrayToTensor


class ArrayToLongTensor(ArrayToTensor):
    """Converts a numpy array or torch tensor to a long tensor."""

    def __call__(self, array: npt.ArrayLike | torch.Tensor):
        """Call method for the transformation.

        Args:
            array: The input numpy array or torch tensor.
        """
        if isinstance(array, torch.Tensor):
            return array.long()
        return super().__call__(array).long()
