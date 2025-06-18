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
"""Dataset for lesion classification."""

from functools import lru_cache
import json
from pathlib import Path
import numpy.typing as npt
import SimpleITK as sitk
import torch
from typing import Callable, Literal
from torchvision import tv_tensors
from typing_extensions import override
from eva.vision.data.datasets import vision

from fef.utils.typing import Orientation


class Lesion(vision.VisionDataset[tv_tensors.Image, torch.Tensor]):
    _fix_orientation: bool = True
    """Whether to fix the orientation of the images to match the default for radiologists."""

    """Dataset for lesion classification."""

    def __init__(
        self,
        root: str,
        split: Literal["train", "val"],
        transforms: Callable,
        chunk_size: int = 1,
        seed: int = 42,
    ) -> None:
        """
        Parameters
        ----------
        root : str
            Root directory of the dataset
        split : {'train', 'val'}
            Dataset split to use
        transforms : callable
            Transform to apply to the data
        chunk_size : int, default=1
            Size of chunk for slices
        seed : int, default=42
            Random seed for reproducibility
        """
        self._root = Path(root)
        self._split = split
        self._transforms = transforms
        self._seed = seed
        self._chunk_size = chunk_size
        self._indices = []

    @property
    @override
    def classes(self) -> list[str]:
        return ["negative", "positive"]

    @property
    @override
    def class_to_idx(self) -> dict[str, int]:
        return {name: index for index, name in enumerate(self.classes)}

    @override
    def configure(self) -> None:
        """
        Configures the dataset.
        """
        self._make_indices()

    @override
    def validate(self) -> None:
        """
        TODO: Validates the dataset.
        """
        pass

    @override
    def __len__(self) -> int:
        """
        Returns the length of the dataset.

        Returns
        -------
        int
            The length of the dataset.
        """
        return len(self._indices)

    @override
    def filename(self, index: int) -> str:
        """
        Returns the filename for the given index.

        Parameters
        ----------
        index : int
            The index of the sample.

        Returns
        -------
        str
            The filename for the given index.
        """
        sample = self._indices[index]
        return f"{sample['start_idx']}_{sample['end_idx']}_{sample['filename']}"

    @lru_cache(maxsize=5)
    def _load_volume_array(self, volume_path: str) -> npt.NDArray:
        """
        Loads volume array and caches it in a LRU cache.

        Parameters
        ----------
        volume_path : str
            Path to the volume file

        Returns
        -------
        ndarray
            The loaded volume array
        """
        image = sitk.ReadImage(volume_path)
        if self._fix_orientation:
            image = self._orientation_to_lps(image)

        return sitk.GetArrayFromImage(image)

    @override
    def load_data(self, index: int) -> tv_tensors.Image:
        """
        Loads the image at the given index with caching for better performance.

        Parameters
        ----------
        index : int
            The index of the sample

        Returns
        -------
        tv_tensors.Image
            The loaded image tensor
        """
        chunk = self._indices[index]
        filename = chunk["filename"]
        full_path = self._root / filename
        volume_array = self._load_volume_array(full_path)
        chunk_array = volume_array[chunk["start_idx"] : chunk["end_idx"]]
        image = tv_tensors.Image(chunk_array, dtype=torch.float32)

        if self._chunk_size > 1:
            # unsqueeze batch dimension
            image.unsqueeze_(0)

        return image

    @override
    def load_target(self, index: int) -> torch.Tensor:
        """
        Loads the target (label) at the given index.

        Parameters
        ----------
        index : int
            The index of the sample.

        Returns
        -------
        Tensor
            The classification label (maximum label in the chunk).
        """
        return torch.tensor(self._indices[index]["label"], dtype=torch.uint8)

    @staticmethod
    def _orientation_to_lps(image: sitk.Image):
        """
        Adjusts the orientation of the image to LPS format.

        Converts the orientation of the SimpleITK image to the LPS
        (Left-Posterior-Superior) coordinate system, which is the standard
        for radiological viewing.

        Parameters
        ----------
        image : sitk.Image
            A SimpleITK image.

        Returns
        -------
        sitk.Image
            The image with adjusted orientation.
        """
        return sitk.DICOMOrient(image, Orientation.LPS.value)

    def _make_indices(self) -> None:
        """
        Creates indices for the dataset by chunking the slices.

        Reads the slice labels from the JSON file, splits them into chunks
        of size self._chunk_size, and stores relevant information for each chunk.
        Only complete chunks (of exactly chunk_size) are kept; incomplete chunks
        at the end of volumes are discarded.
        """
        with open(self._root / f"{self._split}.json") as f:
            train_dict: dict = json.loads(f.read())

        for filename, slice_labels in train_dict.items():
            # Split slice_labels into chunks of chunk_size
            num_slices = len(slice_labels)
            for start_idx in range(0, num_slices, self._chunk_size):
                end_idx = start_idx + self._chunk_size

                # Skip if this chunk would be incomplete
                if end_idx > num_slices:
                    continue

                chunk = slice_labels[start_idx:end_idx]

                # Get the maximum label in this chunk
                chunk_label = max(chunk)

                # Store the filename, chunk start index, and label
                self._indices.append(
                    {"filename": filename, "start_idx": start_idx, "end_idx": end_idx, "label": chunk_label}
                )
