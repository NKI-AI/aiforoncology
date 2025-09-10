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

"""Embeddings classification for paired pre/post contrast datasets."""

import os
from typing import Any, Callable, Literal

import pandas as pd
import torch
from eva.core.data.datasets import embeddings as embeddings_base
from typing_extensions import override


class PrePostClassificationDataset(embeddings_base.EmbeddingsDataset[torch.Tensor]):
    """Embeddings dataset class for paired pre/post contrast classification tasks."""

    def __init__(
        self,
        root: str,
        manifest_file: str,
        split: Literal["train", "val", "test"],
        contrast_names: list[str] = ["pre", "post_1"],
        column_mapping: dict[str, str] = embeddings_base.default_column_mapping,
        embeddings_transforms: Callable | None = None,
        target_transforms: Callable | None = None,
    ):
        """Initialize the dataset."""
        if len(contrast_names) < 2:
            raise ValueError("At least two contrast names must be specified for paired datasets")

        self._contrast_names = contrast_names

        # Add required column mappings for paired datasets
        required_mappings = {
            "patient_id": "patient_id",
            "sample_index": "sample_index",
            "slice_index": "slice_index",
            "contrast": "contrast",
        }
        column_mapping = column_mapping | required_mappings

        super().__init__(
            manifest_file=manifest_file,
            root=root,
            split=split,
            column_mapping=column_mapping,
            embeddings_transforms=embeddings_transforms,
            target_transforms=target_transforms,
        )

    @override
    def _load_manifest(self) -> pd.DataFrame:
        """Load manifest and filter for valid paired data."""
        data = super()._load_manifest()

        def has_all_contrasts(group):
            contrasts = set(group[self._column_mapping["contrast"]])
            return all(contrast in contrasts for contrast in self._contrast_names)

        # Filter to only include patient-slice combinations that have all required contrasts
        paired_data = data.groupby([self._column_mapping["patient_id"], self._column_mapping["slice_index"]]).filter(
            has_all_contrasts
        )

        return paired_data.reset_index(drop=True)

    def _load_embedding_from_path(self, path: str) -> torch.Tensor:
        """Load a single embedding from file."""
        embedding_path = os.path.join(self._root, path)
        tensor = torch.load(embedding_path, map_location="cpu")

        # Handle case where embedding is stored as a list containing a single tensor
        if isinstance(tensor, list):
            if len(tensor) == 1:
                return tensor[0]
            else:
                return torch.stack(tensor, dim=0)
        else:
            if tensor.size(0) == 1:
                return tensor.squeeze(0)
            return tensor

    def _get_contrast_rows(self, patient_id: str, slice_index: int = None) -> tuple[pd.Series, ...]:
        """Get rows for all required contrasts for a given patient_id and optionally slice_index."""
        base_filter = self._data[self._column_mapping["patient_id"]] == patient_id

        if slice_index is not None:
            base_filter = base_filter & (self._data[self._column_mapping["slice_index"]] == slice_index)

        contrast_rows = []
        for contrast_name in self._contrast_names:
            contrast_row = self._data[base_filter & (self._data[self._column_mapping["contrast"]] == contrast_name)]

            if contrast_row.empty:
                slice_info = f", slice {slice_index}" if slice_index is not None else ""
                raise ValueError(f"Missing {contrast_name} contrast data for patient_id {patient_id}{slice_info}.")

            contrast_rows.append(contrast_row.iloc[0])

        return tuple(contrast_rows)

    @override
    def load_embeddings(self, index: int) -> torch.Tensor:
        """Load and stack all contrast embeddings for the given index."""
        patient_id = self._data.at[index, self._column_mapping["patient_id"]]
        slice_index = self._data.at[index, self._column_mapping["slice_index"]]

        contrast_rows = self._get_contrast_rows(patient_id, slice_index)

        embeddings = []
        for row in contrast_rows:
            embedding = self._load_embedding_from_path(row[self._column_mapping["path"]])
            embeddings.append(embedding)

        # Stack embeddings to shape [2, embed_dim]
        # NOTE: Could be changed to torch.cat(embeddings, dim=0) for [2*embed_dim] format
        # if using standard MLP instead of PairedMLP in the future
        return torch.stack(embeddings, dim=0)

    @override
    def load_target(self, index: int) -> torch.Tensor:
        """Load target for the given index."""
        target = self._data.at[index, self._column_mapping["target"]]
        return torch.tensor(target, dtype=torch.int64)

    def load_metadata(self, index: int) -> dict[str, Any]:
        """Load metadata for the given index."""
        metadata = {}

        for key, column_name in self._column_mapping.items():
            if key in ["path", "target", "split"]:
                continue

            if column_name in self._data.columns:
                metadata[key] = self._data.at[index, column_name]

        return metadata

    @override
    def __getitem__(self, index: int) -> tuple:
        """Return embeddings, target, and metadata."""
        embeddings, target = super().__getitem__(index)
        metadata = self.load_metadata(index)
        return embeddings, target, metadata

    @override
    def __len__(self) -> int:
        """Return the number of rows in the filtered paired data."""
        return len(self._data)


class MultiPrePostClassificationDataset(PrePostClassificationDataset):
    """Multi-instance learning version of paired pre/post contrast classification dataset."""

    def __init__(
        self,
        root: str,
        manifest_file: str,
        split: Literal["train", "val", "test"],
        contrast_names: list[str] = ["pre", "post_1"],
        column_mapping: dict[str, str] = embeddings_base.default_column_mapping,
        embeddings_transforms: Callable | None = None,
        target_transforms: Callable | None = None,
    ):
        """Initialize the MIL dataset."""
        super().__init__(
            root=root,
            manifest_file=manifest_file,
            split=split,
            contrast_names=contrast_names,
            column_mapping=column_mapping,
            embeddings_transforms=embeddings_transforms,
            target_transforms=target_transforms,
        )

        self._patient_ids: list[str] = []

    @override
    def setup(self):
        """Setup the dataset and extract unique patient IDs."""
        super().setup()
        self._patient_ids = list(self._data[self._column_mapping["patient_id"]].unique())

    @override
    def load_embeddings(self, index: int) -> torch.Tensor:
        """Load and concatenate all embeddings for the given patient."""
        patient_id = self._patient_ids[index]

        contrast_rows = self._get_contrast_rows(patient_id)

        all_embeddings = []
        for row in contrast_rows:
            embedding = self._load_embedding_from_path(row[self._column_mapping["path"]])
            all_embeddings.append(embedding)

        concatenated_embedding = torch.cat(all_embeddings, dim=0)

        if not concatenated_embedding.ndim == 2:
            raise ValueError(f"Expected 2D tensor, got {concatenated_embedding.ndim} for patient_id {patient_id}.")

        return concatenated_embedding

    @override
    def load_target(self, index: int) -> torch.Tensor:
        """Load target for the given patient."""
        patient_id = self._patient_ids[index]
        targets = self._data.loc[
            self._data[self._column_mapping["patient_id"]] == patient_id, self._column_mapping["target"]
        ]

        if not targets.nunique() == 1:
            raise ValueError(f"Multiple targets found for patient_id {patient_id}.")

        return torch.tensor(targets.iloc[0], dtype=torch.int64)

    @override
    def load_metadata(self, index: int) -> dict[str, Any]:
        """Load metadata for the given patient."""
        patient_id = self._patient_ids[index]

        patient_rows = self._data[self._data[self._column_mapping["patient_id"]] == patient_id]
        representative_row = patient_rows.iloc[0]

        metadata = {}
        for key, column_name in self._column_mapping.items():
            if key in ["path", "target", "split"]:
                continue

            if column_name in self._data.columns:
                metadata[key] = representative_row[column_name]

        return metadata

    @override
    def __len__(self) -> int:
        """Return the number of unique patients."""
        return len(self._patient_ids)
