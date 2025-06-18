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

import abc
from pathlib import Path
from typing import Any, Callable

import SimpleITK as sitk
import torch
from torchvision import tv_tensors
from typing_extensions import override
from eva.vision.data import tv_tensors as eva_tv_tensors
from eva.vision.data.datasets import vision

from fef.utils.typing import Orientation


class Base(vision.VisionDataset[eva_tv_tensors.Volume, torch.Tensor], abc.ABC):
    _fix_orientation: bool = True

    def __init__(self, root, split, transforms: Callable[..., Any] | None = None) -> None:
        super().__init__(transforms)
        self._root = Path(root)
        self._split = split
        self._indices = []

    @override
    def configure(self):
        self._make_indices()

    @override
    def __len__(self) -> int:
        return len(self._indices)

    @override
    def load_data(self, index: int) -> eva_tv_tensors.Volume:
        """
        Loads the volume data at the given index.

        Parameters
        ----------
        index : int
            The index of the sample.

        Returns
        -------
        eva_tv_tensors.Volume
            The loaded volume tensor with dimensions (C, D, W, H), where:
            - C: Channel dimension (1 for CT volumes)
            - D: Depth dimension (number of slices)
            - W: Width dimension
            - H: Height dimension
        """
        volume_path = self._root / self._indices[index]["filename"]
        volume_tensor = self._load_volume_tensor(volume_path)
        # Unsqueeze the channel dimension
        return eva_tv_tensors.Volume(volume_tensor.unsqueeze(0))

    @override
    def load_target(self, index: int) -> tv_tensors.Mask:
        """
        Loads the segmentation mask at the given index.

        Parameters
        ----------
        index : int
            The index of the sample.

        Returns
        -------
        tv_tensors.Mask
            The loaded segmentation mask tensor with dimensions (D, W, H), where:
            - D: Depth dimension (number of slices)
            - W: Width dimension
            - H: Height dimension
        """
        segmentation_path = self._root / "segmentations" / self._indices[index]["filename"]
        segmentation_tensor = self._load_volume_tensor(segmentation_path)
        return tv_tensors.Mask(segmentation_tensor)

    @override
    def filename(self, index: int) -> str:
        return self._indices[index]["filename"]

    def _load_volume_tensor(self, volume_path: str) -> torch.Tensor:
        """
        Loads volume array and caches it in a LRU cache.

        Parameters
        ----------
        volume_path : str
            Path to the volume file

        Returns
        -------
        torch.Tensor
            The loaded volume array converted to torch.Tensor
        """
        image = sitk.ReadImage(volume_path)
        if self._fix_orientation:
            image = self._orientation_to_lps(image)

        return torch.from_numpy(sitk.GetArrayFromImage(image))

    @staticmethod
    def _orientation_to_lps(image: sitk.Image):
        """
        Adjusts the orientation of the image to LPS format.

        Converts the orientation of the SimpleITK image to the LPS
        (Left-Posterior-Superior) coordinate system.

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

    def _make_indices(self):
        volume_names = []
        with open(self._root / f"{self._split}.txt", "r") as f:
            volume_names = f.read().splitlines()

        for volume_name in volume_names:
            self._indices.append({"filename": volume_name})
