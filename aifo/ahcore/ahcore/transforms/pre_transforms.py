# Copyright 2025 AI for Oncology Research Group. All Rights Reserved.
# Copyright 2025 Jonas Teuwen. All Rights Reserved.
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
"""
Module for the pre-transforms, which are the transforms that are applied before samples are outputted in a
dataset.
"""

from __future__ import annotations

from typing import Any, Callable
import warnings
import numpy as np
import numpy.typing as npt
import pyvips
import torch
from ahcore.utils.io import get_logger
from ahcore.utils.types import DlupDatasetSample

PreTransformCallable = Callable[[Any], Any]

logger = get_logger(__name__)


class PreTransformFactory:
    def __init__(self, transforms: list[PreTransformCallable], *args, **kwargs) -> None:
        """
        Pre-transforms are transforms that are applied to the samples directly originating from the dataset.
        These transforms are typically the same for the specific tasks (e.g., segmentation,
        detection or whole-slide classification).
        Each of these tasks has a specific constructor. In all cases, the final transforms convert the PIL image
        (as the image key of the output sample) to a tensor, and ensure that the sample dictionary can be collated.
        In ahcore, the augmentations are done separately and are part of the model in the forward function.

        Parameters
        ----------
        transforms : list
            List of transforms to be used.
        """
        self._transforms = transforms

    def __call__(self, data: DlupDatasetSample) -> DlupDatasetSample:
        for transform in self._transforms:
            data = transform(data)
        return data


class ToDict:
    """Converts the input to a dictionary. Should be used as the first transform in the pipeline if the input is not yet of dict type.
    Parameters
    ----------
    keys : list[str]
        List of keys to use for the dictionary.
    """

    def __init__(self, keys: list[str]) -> None:
        self.keys = keys

    def __call__(self, sample: DlupDatasetSample) -> dict[str, Any]:
        return {key: getattr(sample, key) for key in self.keys}


class SampleNFeatures:
    """Sample N features from the image. Sampling is done with replacement if there are not enough tiles.
    Parameters
    ----------
    n : int
        Number of features to sample.
    """

    def __init__(self, n: int = 1000) -> None:
        self.n = n
        logger.warning(
            f"Sampling {n} features from the image. Sampling WITH replacement is done if there are not enough tiles."
        )

    def __call__(self, sample: dict) -> dict:
        features = sample["image"]

        # Get the dimensions of the image
        h = features.height  # Height
        w = features.width  # Width

        if h != 1:
            raise ValueError(f"Expected features to have a width dimension of 1, got {h}.")

        n_random_indices = (
            np.random.choice(w, self.n, replace=False) if w > self.n else np.random.choice(w, self.n, replace=True)
        )

        # Extract the selected columns (indices) from the image
        # Create a new image from the selected indices
        # todo: this can probably be done without a for-loop quicker
        selected_columns = [features.crop(idx, 0, 1, h) for idx in n_random_indices]

        # Combine the selected columns back into a single image
        sample["image"] = pyvips.Image.arrayjoin(selected_columns, across=1)

        return sample


class LabelToClassIndex:
    """
    Maps label values to class indices according to the index_map specified in the data description.
    Example:
        If there are two tasks:
            - Task1 with classes {A, B, C}
            - Task2 with classes {X, Y}
        Then an input sample could look like: {{"labels": {"Task1": "C", "Task2: "Y"}, ...}
        If the index map is: {"A": 0, "B": 1, "C": 2, "X": 0, "Y": 1}
        The returned sample will look like: {"labels": {"task1": 2, "task2": 1}, ...}
    """

    def __init__(self, index_map: dict[str, int]):
        self._index_map = index_map

    def __call__(self, sample: dict) -> dict:
        sample["labels"] = {
            label_name: self._index_map[label_value] for label_name, label_value in sample["labels"].items()
        }

        return sample


class SelectSpecificLabels:
    """Removes labels that are not in the list of keys.
    Parameters
    ----------
    keys : list[str] | str
        List of keys to retain.
    """

    def __init__(self, keys: list[str] | str):
        if isinstance(keys, str):
            keys = [keys]
        self._keys = keys

    def __call__(self, sample: dict) -> dict:
        sample["labels"] = {
            label_key: label_value for label_key, label_value in sample["labels"].items() if label_key in self._keys
        }
        return sample


class OneHotEncodeMask:
    def __init__(self, use_roi: bool = True):
        """Create the one-hot encoding of the mask for segmentation.
        If we have `N` classes, the result will be an `(B, N + 1, H, W)` tensor, where the first sample is the
        background.

        Parameters
        ----------
        index_map : dict[str, int]
            Index map mapping the label name to the integer value it has in the mask.
        use_roi : bool
            Whether to use the region of interest mask.

        """
        self._use_roi = use_roi

    def __call__(self, sample: dict) -> dict:
        annotations = sample["annotations"]
        mask = annotations.polygons.to_mask(default_value=0).numpy()
        max_value = mask.max()

        # TODO: This can also be done in C++
        # There is also a functional interface.
        new_mask = np.zeros((max_value + 1, *mask.shape))
        for idx in range(max_value + 1):
            new_mask[idx] = (mask == idx).astype(np.float32)

        sample["annotations"].mask = new_mask
        if self._use_roi:
            rois = annotations.rois.to_mask(default_value=0).numpy()
            sample["annotations"].rois = rois

        return sample


class PolygonsToMask:
    def __init__(self, num_classes: int):
        """
        Create the one-hot encoding of the mask for segmentation.
        """
        self._num_classes = num_classes

    def __call__(self, sample: dict) -> dict:
        annotations_as_array = sample["annotations"].polygons.to_mask().numpy()
        sample["target"] = one_hot_encoding(self._num_classes, annotations_as_array)
        return sample


def one_hot_encoding(num_classes: int, mask: npt.NDArray[np.int_ | np.float_]) -> npt.NDArray[np.float32]:
    """
    functional interface to convert labels/predictions into one-hot codes

    Parameters
    ----------
    num_classes: int
        The number of classes in the mask

    mask: npt.NDArray
        The numpy array of model predictions or ground truth labels.

    Returns
    -------
    new_mask: npt.NDArray
        One-hot encoded output
    """
    new_mask = np.zeros((num_classes, *mask.shape), dtype=np.float32)
    for idx in range(num_classes):
        new_mask[idx] = mask == idx
    return new_mask


class AllowCollate:
    """Path objects cannot be collated in the standard pytorch collate function.
    This transform converts the path to a string. Same holds for the annotations and labels
    """

    def __call__(self, sample: dict) -> dict[str, Any]:
        # Path objects cannot be collated
        output = sample.copy()
        for key in sample:
            if key == "path":
                output["path"] = str(sample["path"])
            if key == "annotations":
                del output[key]
            if key == "labels" and sample["labels"] is None:
                del output[key]

        return output


class ImageToTensor:
    """
    Transform to translate the output of a dlup dataset to data_description supported by AhCore
    """

    def __call__(self, sample: dict) -> dict:
        tile: pyvips.Image = sample["image"]
        # Flatten the image to remove the alpha channel, using white as the background color
        using_features = False

        if tile.bands > 4:
            # assuming that more than four bands/channels means that we are handling features
            using_features = True

        # Convert VIPS image to a numpy array then to a torch tensor
        np_image = tile.numpy()
        if using_features:
            # n_tiles x 1 x feature_dim --> n_tiles x feature_dim
            sample["image"] = torch.from_numpy(np_image).squeeze(1).float()
        else:
            # h x w x c --> c x h x w
            sample["image"] = torch.from_numpy(np_image).permute(2, 0, 1).float()

        if sample["image"].sum() == 0:
            warnings.warn(f"Empty tile for {sample['path']} at {sample['coordinates']} (mpp = {sample['mpp']})")

        if "labels" in sample and sample["labels"] is not None:
            # TODO: fix this
            for key, value in sample["labels"].items():
                if isinstance(value, float) or isinstance(value, int):
                    sample["labels"][key] = torch.tensor(value, dtype=torch.float32)
                    if sample["labels"][key].dim() == 0:
                        sample["labels"][key] = sample["labels"][key].unsqueeze(0)

        if "annotations" not in sample:
            return sample

        if sample["annotations"] is not None and sample["annotations"].has_rois:
            rois = sample["annotations"].rois.to_mask().numpy()
            sample["roi"] = torch.from_numpy(rois[np.newaxis, ...]).float()

        sample["mpp"] = torch.tensor(sample["mpp"], dtype=torch.float32)
        # TODO: Why is coordinates a float here? Need to fix that.
        sample["coordinates"] = torch.tensor(sample["coordinates"], dtype=torch.long)

        return sample

    def __repr__(self) -> str:
        return f"{type(self).__name__}()"
