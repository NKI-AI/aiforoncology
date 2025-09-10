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

from eva.vision.data.datasets.segmentation.embeddings import EmbeddingsSegmentationDataset
from typing_extensions import override
from typing import Any


class MRIEmbeddingsSegmentationDataset(EmbeddingsSegmentationDataset):
    """Embeddings segmentation dataset that adds metadata handling.

    This class extends the standard EVA EmbeddingsSegmentationDataset to return
    metadata along with embeddings and targets, enabling volume-based metrics
    that require patient_id and slice_index information.
    """

    @override
    def load_metadata(self, index: int) -> dict[str, Any]:
        """Load metadata from manifest based on column mapping."""
        metadata = {}

        # Load all metadata keys that are present in the column mapping
        # This includes sample_index, slice_index, patient_id, and any other
        # metadata columns that were specified in the configuration
        for key, column_name in self._column_mapping.items():
            # Skip the standard data columns (path, target, split)
            if key in ["path", "target", "split"]:
                continue

            if column_name in self._data.columns:
                metadata[key] = self._data.at[index, column_name]

        return metadata

    @override
    def __getitem__(self, index: int) -> tuple:
        """Return embeddings, target, and metadata.

        This extends the parent class to return metadata as the third element,
        which is required for metrics that need access to patient_id and slice_index.
        """
        embeddings, target = super().__getitem__(index)
        metadata = self.load_metadata(index)

        return embeddings, target, metadata
