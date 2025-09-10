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

"""Custom decoders for segmentation."""

from eva.vision.models.networks.decoders.segmentation.semantic.common import ConvDecoder1x1


class CustomConvDecoder1x1(ConvDecoder1x1):
    """A custom convolutional decoder with a single 1x1 convolutional layer."""

    def __init__(self, in_features: int, num_classes: int) -> None:
        """Initializes the decoder.

        Parameters
        ----------
        in_features : int
            The hidden dimension size of the embeddings.
        num_classes : int
            Number of output classes as channels.
        """
        super().__init__(in_features=in_features, num_classes=num_classes)
        self._combine_features = False
