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

import functools
import os
import warnings
from typing import Any, Callable, Literal
import numpy as np
import pandas as pd
import SimpleITK as sitk
import torch
import matplotlib.pyplot as plt
from eva.vision.data.datasets.vision import VisionDataset
from torchvision import tv_tensors
from typing_extensions import override
from fef.utils.typing.orientation import Orientation
from fef.utils.helpers.lrucache import LRUCache
from fef.transforms.cropping import VolumeBoundingBoxCropper
import logging

logger = logging.getLogger(__name__)


class BaseMRIClassificationDataset(VisionDataset):
    """Base class for Breast MRI Classification datasets (single-slice and MIL).
    Handles all shared logic, including split, label handling, cropping, and caching.
    Subclasses must implement filename() and load_metadata().
    """

    def __init__(
        self,
        label_file: str,
        split: Literal["train", "val", "test"] | None = None,
        transforms: Callable | None = None,
        image_transforms: Callable | None = None,
        split_ratios: tuple[float, float, float] = (0.7, 0.15, 0.15),
        orientation: str = "LPS",
        orientation_plane: Literal["axial", "sagittal", "all"] = "axial",
        contrast_type: Literal["post", "pre", "subtraction", "all"] = "post",
        slice_mode: Literal["all", "middle"] = "all",
        crop_percentile: float | None = None,
        visualize_preprocessing: bool = False,
        seed: int = 42,
        tumor_ratio: float = 0.5,
        cache_size: int = 10,
        label_name: str = "Overall_pcr",
        min_class_percentage: float | None = None,
    ) -> None:
        super().__init__(transforms=transforms)

        if abs(sum(split_ratios) - 1.0) > 1e-6:
            raise ValueError("split_ratios must sum to 1.0")
        if not all(0 <= ratio <= 1 for ratio in split_ratios):
            raise ValueError("All split ratios must be between 0 and 1")

        self._label_file = label_file
        self._split = split
        self._split_ratios = split_ratios
        self._orientation = Orientation.from_value(orientation)
        self._orientation_plane = orientation_plane
        self._contrast_type = contrast_type
        self._mode = "2D"
        self._slice_mode = slice_mode
        self._indices: list[tuple[int, int]] = []
        self._cache_size = cache_size
        self._label_name = label_name
        self._crop_percentile = crop_percentile
        self._visualize_preprocessing = visualize_preprocessing
        self._image_transforms = image_transforms
        self._transforms = transforms
        self._seed = seed
        self._tumor_ratio = tumor_ratio
        self._min_class_percentage = min_class_percentage
        self._volume_cache = LRUCache(cache_size)
        self._cropped_volumes = LRUCache(cache_size)
        self._slice_indices_cache = LRUCache(cache_size)
        self._volume_cropper = None
        if crop_percentile is not None:
            self._volume_cropper = VolumeBoundingBoxCropper(lower_percentile=crop_percentile)

    @override
    def configure(self) -> None:
        """Configure dataset by creating indices for the specified split."""
        logger.info(f"Configuring dataset for split {self._split} ...")
        self._indices = self._create_indices()

    @override
    def validate(self) -> None:
        """Validate that dataset contains valid scan-label pairs."""
        if len(self._scan_label_pairs) == 0:
            raise ValueError("No valid scans with labels found after validation.")

    @override
    def __len__(self) -> int:
        """Return the total number of samples in the dataset."""
        return len(self._indices)

    @property
    @override
    def classes(self) -> list[str]:
        """Get sorted list of unique class labels."""
        return sorted(self._scan_label_pairs[self._label_name].dropna().unique())

    @functools.cached_property
    @override
    def class_to_idx(self) -> dict[str, int]:
        """Map class labels to numerical indices."""
        return {label: index for index, label in enumerate(self.classes)}

    def filename(self, index: int) -> str:
        """Return the filename for the given index. Must be implemented by subclasses."""
        raise NotImplementedError("Subclasses must implement filename()")

    def load_metadata(self, index: int) -> dict[str, Any]:
        """Return metadata for the given index. Must be implemented by subclasses."""
        raise NotImplementedError("Subclasses must implement load_metadata()")

    @override
    def load_data(self, index: int) -> tv_tensors.Image:
        """Load the image data."""
        try:
            sample_index, slice_index = self._indices[index]
            volume_path = self._volume_files[sample_index]
            volume = self._get_cropped_volume(volume_path)

            # Extract the slice and convert to float32
            image_array = volume[slice_index].astype(np.float32)
            # Add channel dimension for single-channel image
            image_array = np.expand_dims(image_array, axis=0)
            image_tensor = tv_tensors.Image(image_array)

            # Apply image-only transforms
            return self._apply_image_transforms(image_tensor)

        except Exception as e:
            raise RuntimeError(f"Error loading image at index {index}: {str(e)}")

    @override
    def load_target(self, index: int) -> torch.Tensor:
        """Load the target label."""
        sample_index, slice_index = self._indices[index]
        volume_path = self._volume_files[sample_index]

        label_row = self._scan_label_pairs[self._scan_label_pairs["file_path"] == volume_path]
        target_label = label_row[self._label_name].values[0]
        target_idx = self.class_to_idx[target_label]
        return torch.tensor(target_idx, dtype=torch.int64)

    def _apply_image_transforms(self, image: tv_tensors.Image) -> tv_tensors.Image:
        """Apply image-only transforms to the input image."""
        if self._image_transforms is not None:
            image = self._image_transforms(image)
        return image

    @override
    def _apply_transforms(self, image: tv_tensors.Image, target: torch.Tensor) -> tuple[tv_tensors.Image, torch.Tensor]:
        """Apply joint transforms to both image and target."""
        if self._transforms is not None:
            image, target = self._transforms(image, target)
        return image, target

    def _load_volume(self, file_path: str) -> np.ndarray:
        """Load and preprocess a volume from a file path.

        Args:
            file_path: Path to the volume file

        Returns:
            Preprocessed volume as numpy array
        """
        cached = self._volume_cache.get(file_path)
        if cached is not None:
            return cached

        try:
            if not os.path.exists(file_path):
                raise FileNotFoundError(f"File not found: {file_path}")

            image = sitk.ReadImage(str(file_path))
            oriented_image = self._orient_image(image)
            array = self._sitk_to_array(oriented_image)

            self._volume_cache.put(file_path, array)
            return array

        except Exception as e:
            raise RuntimeError(f"Error loading volume {file_path}: {str(e)}")

    def _get_cropped_volume(self, volume_path: str) -> np.ndarray:
        """Get the cropped (or original) volume."""
        # Check if in cache, otherwise load and crop
        cached_volume = self._cropped_volumes.get(volume_path)
        if cached_volume is not None:
            return cached_volume
        volume = self._load_volume(volume_path)

        if self._volume_cropper is not None:
            # Little bit hacky until kaiko implements a volume cropper
            # Apply cropping by passing volume twice - once as volume and once as "segmentation"
            # This works because the cropper just needs array shapes to match
            cropped_volume, _ = self._volume_cropper(volume, volume)
        else:
            cropped_volume = volume

        self._cropped_volumes.put(volume_path, cropped_volume)

        return cropped_volume

    def _get_acquisition_plane(self, image: sitk.Image) -> str:
        """Determine the acquisition plane of a volume.

        Args:
            image: SimpleITK image

        Returns:
            Acquisition plane as string ('axial', 'sagittal', or 'coronal')
        """
        spacing = image.GetSpacing()
        if spacing[2] > spacing[0] and spacing[2] > spacing[1]:
            return "axial"
        elif spacing[1] > spacing[0] and spacing[1] > spacing[2]:
            return "sagittal"
        else:
            return "coronal"

    def _orient_image(self, image: sitk.Image) -> sitk.Image:
        """Orient volume to standard orientation based on acquisition plane.

        Args:
            image: SimpleITK image

        Returns:
            Reoriented SimpleITK image
        """
        target_orientation = Orientation.get_orientation_for_plane(
            self._get_acquisition_plane(image), self._orientation.value
        )

        oriented_image = sitk.DICOMOrient(image, target_orientation.value)

        if len(sitk.GetArrayFromImage(oriented_image).shape) > 3:
            raise ValueError("Multivolume cases are not supported.")

        return oriented_image

    def _sitk_to_array(self, image: sitk.Image) -> np.ndarray:
        """Convert SimpleITK image to numpy array.

        Args:
            image: SimpleITK image

        Returns:
            Volume as numpy array
        """
        array = sitk.GetArrayFromImage(image)

        dims = array.shape
        if dims[1] == dims[2]:
            return array
        elif dims[0] == dims[1]:
            return array.transpose(2, 0, 1)
        else:
            return array.transpose(1, 0, 2)

    @functools.cached_property
    def _scan_label_pairs(self) -> pd.DataFrame:
        """Load and filter scan-label pairs from manifest file.

        Applies filtering based on contrast type and minimum class percentage requirements.
        Returns DataFrame with valid scan-label pairs.
        """
        try:
            df = pd.read_csv(self._label_file)
            required_columns = ["file_path", self._label_name]
            missing_columns = [col for col in required_columns if col not in df.columns]
            if missing_columns:
                raise ValueError(f"Missing required columns in manifest: {missing_columns}")
            if self._contrast_type != "all" and "contrast" in df.columns:
                if self._contrast_type == "post":
                    contrast_filter = df["contrast"].str.contains("post", case=False, na=False)
                    df_filtered = df[contrast_filter]
                    post_1_filter = df_filtered["contrast"].str.contains("post_1", case=False, na=False)
                    if post_1_filter.any():
                        df_filtered = df_filtered[post_1_filter]
                elif self._contrast_type == "pre":
                    contrast_filter = df["contrast"].str.contains("pre", case=False, na=False)
                    df_filtered = df[contrast_filter]
                elif self._contrast_type == "subtraction":
                    contrast_filter = df["contrast"].str.contains("sub", case=False, na=False)
                    df_filtered = df[contrast_filter]
                else:
                    raise ValueError(f"Invalid contrast type: {self._contrast_type}")
                if len(df_filtered) == 0:
                    if "contrast" in df.columns:
                        raise ValueError(
                            f"No scans found with contrast type '{self._contrast_type}'. "
                            f"Available contrasts: {df['contrast'].unique().tolist()}"
                        )
                    else:
                        df_filtered = df
            else:
                df_filtered = df
            df_filtered = df_filtered.dropna(subset=[self._label_name])
            if len(df_filtered) == 0:
                raise ValueError("No valid scan-label pairs found!")
            if self._min_class_percentage is not None:
                class_counts = df_filtered[self._label_name].value_counts()
                total_samples = len(df_filtered)
                min_samples = total_samples * (self._min_class_percentage / 100)
                valid_classes = class_counts[class_counts >= min_samples].index
                df_filtered = df_filtered[df_filtered[self._label_name].isin(valid_classes)]
                if len(df_filtered) == 0:
                    raise ValueError(
                        f"No classes meet the minimum percentage requirement of {self._min_class_percentage}%. "
                        f"Class distribution: {class_counts.to_dict()}"
                    )
                logger.info(f"Filtered out classes with less than {self._min_class_percentage}% of samples")
                logger.info(f"Remaining class distribution: {df_filtered[self._label_name].value_counts().to_dict()}")
                logger.info(
                    f"Total amount of classes is now {len(df_filtered[self._label_name].unique())}. Remember to adjust NUM_CLASSES in the config file for fitting."
                )
            logger.info(f"Determined {len(df_filtered)} valid scan-label pairs")
            return df_filtered
        except Exception as e:
            raise RuntimeError(f"Error reading manifest file: {str(e)}")

    @functools.cached_property
    def _volume_files(self) -> list[str]:
        """Get list of all volume file paths."""
        return self._scan_label_pairs["file_path"].tolist()

    def _get_valid_slice_indices(self, volume: np.ndarray, file_path: str = "") -> tuple[list[int], list[int]]:
        """Get indices of valid slices in a volume.

        Args:
            volume: 3D volume array
            file_path: Path to volume file for label lookup

        Returns:
            Tuple of (positive_indices, negative_indices) where positive indices
            contain tumor/target and negative indices do not.
        """
        if self._slice_mode == "middle":
            middle_idx = volume.shape[0] // 2
            return [middle_idx], []
        breast_filter = np.sum(volume > 0, axis=(1, 2)) > 100
        try:
            sample_idx = self._volume_files.index(file_path)
            label = self._scan_label_pairs.iloc[sample_idx][self._label_name]
            if len(self.classes) == 2:
                positive_classes = [self.classes[-1]]
            else:
                positive_classes = self.classes[1:]
            is_positive_sample = label in positive_classes
            breast_indices = list(np.where(breast_filter)[0])
            if is_positive_sample:
                return breast_indices, []
            else:
                return [], breast_indices
        except Exception as e:
            warnings.warn(f"Error determining label for {file_path}: {e}")
            breast_indices = list(np.where(breast_filter)[0])
            return [], breast_indices

    def _create_indices(self) -> list[tuple[int, int]]:
        """Create dataset indices based on split and tumor ratio requirements.

        Returns list of (sample_index, slice_index) tuples for all valid slices.
        """
        indices = []
        volumes_processed = 0
        if self._tumor_ratio is None:
            logger.info(f"Creating dataset for '{self._split}' split with ALL slices (no tumor ratio filtering)")
        else:
            tumor_percent = int(self._tumor_ratio * 100)
            non_tumor_percent = 100 - tumor_percent
            logger.info(
                f"Creating dataset with {tumor_percent}% tumor / {non_tumor_percent}% non-tumor ratio for '{self._split}' split"
            )
        random_generator = np.random.default_rng(seed=self._seed)
        sample_indices = self._get_split_indices()
        logger.info(f"Split {self._split} has {len(sample_indices)} samples")
        for sample_idx in sample_indices:
            try:
                sample_idx = int(sample_idx)
                if sample_idx < 0 or sample_idx >= len(self._volume_files):
                    raise IndexError(f"Sample index {sample_idx} out of bounds (0-{len(self._volume_files) - 1})")
                vol_path = self._volume_files[sample_idx]
                volume = self._load_volume(vol_path)
                volumes_processed += 1
                if self._visualize_preprocessing:
                    viz_dir = "visualizations/classification"
                    self.visualize_preprocessing_steps(vol_path, viz_dir)
                cached_data = self._slice_indices_cache.get(vol_path)
                if cached_data is None:
                    slice_data = self._get_valid_slice_indices(volume, vol_path)
                    self._slice_indices_cache.put(vol_path, slice_data)
                else:
                    slice_data = cached_data
                positive_indices, negative_indices = slice_data
                if self._tumor_ratio is None:
                    if positive_indices:
                        indices.extend([(sample_idx, int(slice_idx)) for slice_idx in positive_indices])
                    if negative_indices:
                        indices.extend([(sample_idx, int(slice_idx)) for slice_idx in negative_indices])
                else:
                    if positive_indices:
                        indices.extend([(sample_idx, int(slice_idx)) for slice_idx in positive_indices])
                    if negative_indices and positive_indices and self._tumor_ratio < 1.0:
                        n_positive_samples = len(positive_indices)
                        n_negative_needed = int(n_positive_samples * (1 - self._tumor_ratio) / self._tumor_ratio)
                        n_negative_samples = min(n_negative_needed, len(negative_indices))
                        if n_negative_samples > 0:
                            sampled_negative = random_generator.choice(
                                negative_indices, size=n_negative_samples, replace=False
                            )
                            indices.extend([(sample_idx, int(slice_idx)) for slice_idx in sampled_negative])
                    elif negative_indices and not positive_indices:
                        indices.extend([(sample_idx, int(slice_idx)) for slice_idx in negative_indices])
            except Exception as e:
                logger.warning(f"Skipping sample {sample_idx}: {str(e)}")
                continue
        logger.info(f"Created {len(indices)} indices from {volumes_processed} volumes")
        if not indices:
            raise ValueError("No valid slices found. Check your scan files and processing.")
        return indices

    def _get_split_indices(self) -> list[int]:
        """Get sample indices for the current split (train/val/test)."""
        _train_indices, _val_indices, _test_indices = self._create_stratified_split()
        split_indices = {
            "train": _train_indices,
            "val": _val_indices,
            "test": _test_indices,
            None: list(range(len(self._volume_files))),
        }
        indices = split_indices.get(self._split)
        if indices is None:
            raise ValueError("Invalid data split. Use 'train', 'val', 'test' or `None`.")
        return indices

    def _create_stratified_split(self) -> tuple[list[int], list[int], list[int]]:
        """Create stratified train/val/test splits maintaining class distribution."""
        labels = self._get_labels()
        indices_by_class = {}
        for idx, label in enumerate(labels):
            if label not in indices_by_class:
                indices_by_class[label] = []
            indices_by_class[label].append(idx)
        for class_indices in indices_by_class.values():
            class_indices.sort()
        train_indices, val_indices, test_indices = [], [], []
        for class_indices in indices_by_class.values():
            n = len(class_indices)
            train_ratio, val_ratio, _ = self._split_ratios
            train_end = int(n * train_ratio)
            val_end = train_end + int(n * val_ratio)
            train_indices.extend(class_indices[:train_end])
            val_indices.extend(class_indices[train_end:val_end])
            test_indices.extend(class_indices[val_end:])
        train_indices = [int(idx) for idx in train_indices]
        val_indices = [int(idx) for idx in val_indices]
        test_indices = [int(idx) for idx in test_indices]
        self._print_class_distribution(train_indices, val_indices, test_indices, labels)
        return train_indices, val_indices, test_indices

    def _get_labels(self) -> list[int]:
        """Get numerical labels for all samples."""
        labels = self._scan_label_pairs[self._label_name].tolist()
        return [self.class_to_idx[label] for label in labels]

    def _print_class_distribution(
        self, train_indices: list[int], val_indices: list[int], test_indices: list[int], labels: list[int]
    ) -> None:
        """Print class distribution statistics for each split."""
        train_labels = [labels[i] for i in train_indices]
        val_labels = [labels[i] for i in val_indices]
        test_labels = [labels[i] for i in test_indices]
        logger.info("Class distribution in train split:")
        logger.info(pd.Series(train_labels).value_counts())
        logger.info("Class distribution in val split:")
        logger.info(pd.Series(val_labels).value_counts())
        logger.info("Class distribution in test split:")
        logger.info(pd.Series(test_labels).value_counts())

    def _get_number_of_slices_per_volume(self, sample_index: int) -> int:
        """Get the number of slices in a volume at the given sample index."""
        file_path = self._volume_files[sample_index]
        volume = self._load_volume(file_path)
        return volume.shape[0]

    def visualize_preprocessing_steps(self, file_path: str, output_dir: str) -> None:
        """Visualize preprocessing steps for debugging.

        Args:
            file_path: Path to volume file
            output_dir: Directory to save visualizations
        """
        os.makedirs(output_dir, exist_ok=True)
        volume = sitk.ReadImage(file_path)
        volume_array = sitk.GetArrayFromImage(volume)
        plt.figure(figsize=(15, 5))
        plt.subplot(131)
        plt.imshow(volume_array[volume_array.shape[0] // 2], cmap="gray")
        plt.title("Original")
        plt.axis("off")
        volume = self._orient_image(volume)
        volume_array = sitk.GetArrayFromImage(volume)
        plt.subplot(132)
        plt.imshow(volume_array[volume_array.shape[0] // 2], cmap="gray")
        plt.title("Oriented")
        plt.axis("off")
        volume = self._get_cropped_volume(volume)
        volume_array = sitk.GetArrayFromImage(volume)
        plt.subplot(133)
        plt.imshow(volume_array[volume_array.shape[0] // 2], cmap="gray")
        plt.title("Cropped")
        plt.axis("off")
        plt.tight_layout()
        plt.savefig(os.path.join(output_dir, os.path.basename(file_path) + ".png"))
        plt.close()
