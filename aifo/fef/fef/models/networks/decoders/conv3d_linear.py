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
import torch.nn as nn

from fef.models.networks.decoders.decoder3d import Decoder3d


class Conv3dLinear(Decoder3d):
    def __init__(self, in_features: int, num_classes: int) -> None:
        super().__init__(
            layers=nn.Conv3d(
                in_channels=in_features,
                out_channels=num_classes,
                kernel_size=(1, 1, 1),
            )
        )
