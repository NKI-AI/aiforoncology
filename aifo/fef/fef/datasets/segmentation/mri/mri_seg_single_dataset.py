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

"""Breast MRI Segmentation dataset."""

import functools
import os
import warnings
import matplotlib.pyplot as plt
import numpy as np
import pandas as pd
import SimpleITK as sitk
import torch
from typing import Any, Callable, Literal
from torchvision import tv_tensors
from typing_extensions import override
from eva.vision.data.datasets import _utils as data_utils
from eva.vision.data.datasets.vision import VisionDataset
from fef.utils.typing.orientation import Orientation
from fef.utils.helpers.lrucache import LRUCache
import logging

# I'm adding the volume cropping here, but ideally we would define a volume based vs slice based transform in eva somehow
from fef.transforms.cropping import VolumeBoundingBoxCropper

logger = logging.getLogger(__name__)


class MRITumorSegDataset(VisionDataset[tv_tensors.Image, tv_tensors.Mask]):
    def __init__(
        self,
        label_file: str,
        split: Literal["train", "val", "test"] | None = None,
        transforms: Callable | None = None,
        image_transforms: Callable | None = None,
        split_ratios: tuple[float, float, float] = (0.7, 0.15, 0.15),
        visualize_preprocessing: bool = False,
        orientation: str = "LPS",
        orientation_plane: Literal["axial", "sagittal", "all"] = "axial",
        contrast_type: Literal["post", "pre", "subtraction", "all"] = "post",
        crop_percentile: float | None = None,
        seed: int = 42,
        tumor_ratio: float = 0.5,
        cache_size: int = 10,
    ) -> None:
        """Initialize dataset.

        Args:
            label_file: Path to the CSV file containing scan-segmentation pairs.
            split: Dataset split to use.
            transforms: A function/transform for joint image-target transforms
            image_transforms: A function/transform for image-only transforms
            split_ratios: Tuple of (train, val, test) ratios that should sum to 1.0
            visualize_preprocessing: Whether to visualize preprocessing steps.
            orientation: Orientation to use for the dataset (e.g., "LPS").
            orientation_plane: Filter dataset by orientation plane ("axial", "sagittal", or "all").
            contrast_type: Filter dataset by contrast type ("post", "pre", "subtraction", or "all").
            crop_percentile: Percentile below which pixels are considered background.
            seed: Seed for reproducible random sampling.
            tumor_ratio: Ratio of tumor slices in dataset.
                        0.5 means equal number of tumor and non-tumor slices.
                        Values between 0 and 1 are valid.
            cache_size: Size of the LRU cache for volume and segmentation loading
        """
        super().__init__(transforms=transforms)

        # Validate split ratios
        if not isinstance(split_ratios, (tuple, list)) or len(split_ratios) != 3:
            raise ValueError("split_ratios must be a tuple/list of length 3")
        if abs(sum(split_ratios) - 1.0) > 1e-6:
            raise ValueError("split_ratios must sum to 1.0")
        if not all(0 <= ratio <= 1 for ratio in split_ratios):
            raise ValueError("All split ratios must be between 0 and 1")

        self._label_file = label_file
        self._split = split
        self._split_ratios = split_ratios
        self._image_transforms = image_transforms
        self._transforms = transforms
        self._visualize_preprocessing = visualize_preprocessing
        self._orientation = Orientation.from_value(orientation)
        self._orientation_plane = orientation_plane
        self._contrast_type = contrast_type
        self._crop_percentile = crop_percentile
        self._seed = seed
        self._tumor_ratio = tumor_ratio

        # Initialize cropper if needed
        self._volume_cropper = None
        if crop_percentile is not None:
            self._volume_cropper = VolumeBoundingBoxCropper(lower_percentile=crop_percentile)

        # Initialize LRU caches with specified size
        self._volume_cache = LRUCache(cache_size)
        self._segmentation_cache = LRUCache(cache_size)
        self._cropped_volumes = LRUCache(cache_size)
        self._cropped_segmentations = LRUCache(cache_size)
        self._slice_indices_cache = LRUCache(cache_size)

    @property
    @override
    def classes(self) -> list[str]:
        return ["breast", "tumor"]

    @functools.cached_property
    @override
    def class_to_idx(self) -> dict[str, int]:
        return {label: index for index, label in enumerate(self.classes)}

    @override
    def filename(self, index: int) -> str:
        sample_index, _ = self._indices[index]
        volume_file_path = self._volume_files[sample_index]
        return os.path.basename(volume_file_path)

    @override
    def configure(self) -> None:
        """Configure dataset for the specified split."""
        logger.info(f"Configuring dataset for split {self._split} ...")
        self._indices = self._create_indices()

    @override
    def validate(self) -> None:
        """Validate that all files exist and are properly paired."""
        if len(self._scan_seg_pairs) == 0:
            raise ValueError("No valid scan-segmentation pairs found after validation.")

        if len(self._volume_files) != len(self._segmentation_files):
            raise ValueError("The number of volume files does not match the number of the segmentation ones.")

    @override
    def load_data(self, index: int) -> tv_tensors.Image:
        """Load the image data."""
        try:
            sample_index, slice_index = self._indices[index]
            volume_path = self._volume_files[sample_index]
            volume, _ = self._get_cropped_volume_and_segmentation(volume_path)

            image_array = volume[slice_index].astype(np.float32)
            # Add channel dimension for single-channel image
            image_array = np.expand_dims(image_array, axis=0)
            image_tensor = tv_tensors.Image(image_array)

            # Apply image-only transforms if specified
            if self._image_transforms is not None:
                image_tensor = self._image_transforms(image_tensor)

            return image_tensor

        except Exception as e:
            raise RuntimeError(f"Error loading image at index {index}: {str(e)}")

    @override
    def load_target(self, index: int) -> tv_tensors.Mask:
        """Load the target mask."""
        sample_index, slice_index = self._indices[index]
        volume_path = self._volume_files[sample_index]
        _, segmentation = self._get_cropped_volume_and_segmentation(volume_path)
        semantic_labels = segmentation[slice_index].astype(np.float32)

        mask_tensor = tv_tensors.Mask(semantic_labels, dtype=torch.int64)

        return mask_tensor

    @override
    def load_metadata(self, index: int) -> dict[str, Any]:
        """Load metadata for the sample."""
        sample_index, slice_index = self._indices[index]
        volume_path = self._volume_files[sample_index]

        label_row = self._scan_seg_pairs[self._scan_seg_pairs["file_path"] == volume_path]

        # Start with basic metadata
        metadata = {
            "sample_index": int(sample_index),
            "slice_index": int(slice_index),
            "patient_id": label_row["patient_id"].values[0] if "patient_id" in label_row.columns else "",
        }

        # Add all columns from the label file
        for column in label_row.columns:
            # Skip already added metadata
            if column in metadata:
                continue

            # Handle different data types appropriately
            value = label_row[column].values[0]

            # Skip None/NaN values - this prevents collation errors
            if pd.isna(value):
                continue

            # Convert numpy types to Python native types for JSON serialization
            if isinstance(value, (np.int64, np.int32, np.int16, np.int8)):
                metadata[column] = int(value)
            elif isinstance(value, (np.float64, np.float32, np.float16)):
                metadata[column] = float(value)
            elif isinstance(value, str):
                metadata[column] = value  # Keep strings as they are
            else:
                # Try to convert to a standard Python type if possible
                try:
                    metadata[column] = value.item() if hasattr(value, "item") else value
                except (AttributeError, ValueError):
                    metadata[column] = str(value)  # Fall back to string representation

        return metadata

    @override
    def __len__(self) -> int:
        return len(self._indices)

    def _get_number_of_slices_per_volume(self, sample_index: int) -> int:
        file_path = self._volume_files[sample_index]
        volume = self._load_volume(file_path)  # No unpacking needed
        return volume.shape[0]

    @functools.cached_property
    def _scan_seg_pairs(self) -> pd.DataFrame:
        """Get valid scan-segmentation pairs from the manifest file."""
        try:
            # Read the manifest file
            df = pd.read_csv(self._label_file)

            # Verify required columns exist
            required_columns = ["file_path", "seg_path", "contrast"]
            missing_columns = [col for col in required_columns if col not in df.columns]
            if missing_columns:
                raise ValueError(f"Missing required columns in manifest: {missing_columns}")

            # Filter by contrast type if specified
            if self._contrast_type != "all":
                if self._contrast_type == "post":
                    # Filter for post-contrast scans (post_1, post_2, etc.)
                    contrast_filter = df["contrast"].str.contains("post", case=False, na=False)
                    df_filtered = df[contrast_filter]

                    # Further filter to only get post_1 scans
                    post_1_filter = df_filtered["contrast"].str.contains("post_1", case=False, na=False)
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
                    raise ValueError(
                        f"No scans found with contrast type '{self._contrast_type}'. "
                        f"Available contrasts: {df['contrast'].unique().tolist()}"
                    )
            else:
                df_filtered = df

            if len(df_filtered) == 0:
                raise ValueError("No valid scan-segmentation pairs found!")

            print(f"Determined {len(df_filtered)} valid scan-segmentation pairs")

            return df_filtered
        except Exception as e:
            raise RuntimeError(f"Error reading manifest file: {str(e)}")

    @functools.cached_property
    def _volume_files(self) -> list[str]:
        """Get list of volume files."""
        return self._scan_seg_pairs["file_path"].tolist()

    @functools.cached_property
    def _segmentation_files(self) -> list[str]:
        """Get list of corresponding segmentation files."""
        return self._scan_seg_pairs["seg_path"].tolist()

    def _get_valid_slice_indices(
        self, segmentation: np.ndarray, file_path: str = "", volume: np.ndarray = None
    ) -> tuple[list[int], list[int]]:
        """Get indices of slices that should be included in the dataset.

        Identifies breast-only and breast-with-tumor slices separately for balanced sampling.

        Args:
            segmentation: 3D numpy array of shape (D, H, W)
            file_path: Path to the segmentation file (for logging)
            volume: Volume data corresponding to the segmentation

        Returns:
            Tuple of (breast_only_indices, breast_with_tumor_indices)
        """
        if np.sum(segmentation) == 0:
            warnings.warn(f"Segmentation is completely empty in file {file_path}")
            return [], []

        breast_filter = np.sum(volume, axis=(1, 2)) > 0
        tumor_filter = np.sum(segmentation > 0, axis=(1, 2)) > 0

        breast_only_filter = breast_filter & ~tumor_filter
        breast_with_tumor_filter = breast_filter & tumor_filter

        breast_only_indices = list(np.where(breast_only_filter)[0])
        breast_with_tumor_indices = list(np.where(breast_with_tumor_filter)[0])

        return breast_only_indices, breast_with_tumor_indices

    def _create_indices(self) -> list[tuple[int, int]]:
        """Builds the dataset indices for the specified split using balanced sampling."""
        indices = []
        empty_segmentations = 0
        volumes_processed = 0

        tumor_percent = int(self._tumor_ratio * 100)
        non_tumor_percent = 100 - tumor_percent
        logger.info(
            f"Creating dataset with {tumor_percent}% tumor / {non_tumor_percent}% non-tumor ratio for '{self._split}' split"
        )

        random_generator = np.random.default_rng(seed=self._seed)

        for sample_idx in self._get_split_indices():
            vol_path = self._volume_files[sample_idx]
            seg_path = self._segmentation_files[sample_idx]

            volume = self._load_volume(vol_path)
            segmentation = self._load_segmentation(seg_path)
            volumes_processed += 1

            if self._visualize_preprocessing:
                viz_dir = "visualizations/segmentation"
                self.visualize_preprocessing_steps(vol_path, viz_dir, volume, segmentation)

            if np.sum(segmentation) == 0:
                empty_segmentations += 1
                continue

            cached_data = self._slice_indices_cache.get(seg_path)
            if cached_data is None:
                slice_data = self._get_valid_slice_indices(segmentation, seg_path, volume)
                self._slice_indices_cache.put(seg_path, slice_data)
            else:
                slice_data = cached_data

            breast_only_indices, breast_with_tumor_indices = slice_data

            if not breast_with_tumor_indices:
                continue

            indices.extend([(sample_idx, int(slice_idx)) for slice_idx in breast_with_tumor_indices])

            if breast_only_indices and self._tumor_ratio < 1.0:
                n_tumor_samples = len(breast_with_tumor_indices)
                n_non_tumor_needed = int(n_tumor_samples * (1 - self._tumor_ratio) / self._tumor_ratio)
                n_non_tumor_samples = min(n_non_tumor_needed, len(breast_only_indices))

                if n_non_tumor_samples > 0:
                    sampled_breast_only = random_generator.choice(
                        breast_only_indices, size=n_non_tumor_samples, replace=False
                    )
                    indices.extend([(sample_idx, int(slice_idx)) for slice_idx in sampled_breast_only])

        logger.info(f"Total slices: {len(indices)}")
        logger.info(f"Volumes processed: {volumes_processed}")
        if empty_segmentations > 0:
            logger.info(f"Skipped {empty_segmentations} empty segmentations")

        if not indices:
            raise ValueError("No valid slices found in any segmentation. Check your segmentation files.")

        return indices

    def _get_split_indices(self) -> list[int]:
        """Returns the sample indices for the specified dataset split."""
        _train_index_ranges, _val_index_ranges, _test_index_ranges = self._create_dynamic_ranges()
        split_index_ranges = {
            "train": _train_index_ranges,
            "val": _val_index_ranges,
            "test": _test_index_ranges,
            None: [(0, len(self._volume_files))],
        }
        index_ranges = split_index_ranges.get(self._split)
        if index_ranges is None:
            raise ValueError("Invalid data split. Use 'train', 'val', 'test' or `None`.")

        return data_utils.ranges_to_indices(index_ranges)

    def _create_dynamic_ranges(self) -> tuple[list[tuple[int, int]], list[tuple[int, int]], list[tuple[int, int]]]:
        """Dynamically create train, val, and test index ranges."""
        total_volumes = len(self._volume_files)

        train_ratio, val_ratio, _ = self._split_ratios
        train_len = int(train_ratio * total_volumes)
        val_len = int(val_ratio * total_volumes)

        _train_index_ranges = [(0, train_len)]
        _val_index_ranges = [(train_len, train_len + val_len)]
        _test_index_ranges = [(train_len + val_len, total_volumes)]

        return _train_index_ranges, _val_index_ranges, _test_index_ranges

    def _apply_image_transforms(self, image: tv_tensors.Image) -> tv_tensors.Image:
        """Apply image-only transforms."""
        if self._image_transforms is not None:
            image = self._image_transforms(image)
        return image

    @override
    def _apply_transforms(self, image: tv_tensors.Image, target: torch.Tensor) -> tuple[tv_tensors.Image, torch.Tensor]:
        """Apply joint transforms from config."""
        if self._transforms is not None:
            image, target = self._transforms(image, target)
        return image, target

    def _load_volume(self, file_path: str) -> np.ndarray:
        """Load volume and orient correctly based on acquisition plane."""
        cached = self._volume_cache.get(file_path)
        if cached is not None:
            return cached

        try:
            # 1. Read image
            image = sitk.ReadImage(str(file_path))

            # 2. Get acquisition plane
            acquisition_plane = self._get_acquisition_plane(file_path)

            # 3. Orient to standard orientation
            oriented_image = self._orient_image(image, acquisition_plane)

            # 4. Convert to array
            array = self._sitk_to_array(oriented_image)

            # Store in cache
            self._volume_cache.put(file_path, array)
            return array
        except Exception as e:
            logger.error(f"Error loading volume {file_path}: {str(e)}")
            logger.error(f"Available columns in manifest: {self._scan_seg_pairs.columns.tolist()}")
            raise RuntimeError(f"Error loading volume {file_path}: {str(e)}")

    def _load_segmentation(self, file_path: str) -> np.ndarray:
        """Load segmentation following exactly the same steps as volume loading."""
        cached = self._segmentation_cache.get(file_path)
        if cached is not None:
            return cached

        try:
            # 1. Read image
            seg_image = sitk.ReadImage(str(file_path))

            # Get corresponding volume path
            volume_idx = self._segmentation_files.index(file_path)
            volume_path = self._volume_files[volume_idx]
            volume_image = sitk.ReadImage(str(volume_path))

            # Check and fix geometry if needed
            if not self._check_geometry_match(volume_image, seg_image):
                seg_image = self._resample_to_reference(seg_image, volume_image)

            # 2. Get acquisition plane (use volume's plane)
            acquisition_plane = self._get_acquisition_plane(volume_path)

            # 3. Orient to standard orientation
            oriented_seg = self._orient_image(seg_image, acquisition_plane)

            # 4. Convert to array
            array = self._sitk_to_array(oriented_seg)

            # Store in cache
            self._segmentation_cache.put(file_path, array)
            return array
        except Exception as e:
            raise RuntimeError(f"Error loading segmentation {file_path}: {str(e)}")

    def _get_cropped_volume_and_segmentation(self, volume_path: str) -> tuple[np.ndarray, np.ndarray]:
        """Get the cropped (or original) volume and segmentation."""
        volume_idx = self._volume_files.index(volume_path)
        seg_path = self._segmentation_files[volume_idx]

        # Check cache first
        cached_volume = self._cropped_volumes.get(volume_path)
        cached_segmentation = self._cropped_segmentations.get(seg_path)

        if cached_volume is not None and cached_segmentation is not None:
            return cached_volume, cached_segmentation

        # If not in cache, load and crop
        volume = self._load_volume(volume_path)  # No unpacking needed
        segmentation = self._load_segmentation(seg_path)

        if volume.shape != segmentation.shape:
            raise ValueError(f"Volume shape {volume.shape} doesn't match segmentation shape {segmentation.shape}")

        if self._volume_cropper is not None:
            cropped_volume, cropped_segmentation = self._volume_cropper(volume, segmentation)
        else:
            cropped_volume, cropped_segmentation = volume, segmentation

        # Store in cache
        self._cropped_volumes.put(volume_path, cropped_volume)
        self._cropped_segmentations.put(seg_path, cropped_segmentation)

        return cropped_volume, cropped_segmentation

    def _check_geometry_match(self, image1: sitk.Image, image2: sitk.Image, tolerance: float = 1e-6) -> bool:
        """Check if two images have matching geometry."""
        # Check size
        if image1.GetSize() != image2.GetSize():
            return False

        # Check spacing
        spacing1 = image1.GetSpacing()
        spacing2 = image2.GetSpacing()
        if any(abs(s1 - s2) > tolerance for s1, s2 in zip(spacing1, spacing2)):
            return False

        # Check origin
        origin1 = image1.GetOrigin()
        origin2 = image2.GetOrigin()
        if any(abs(o1 - o2) > tolerance for o1, o2 in zip(origin1, origin2)):
            return False

        # Check direction
        direction1 = image1.GetDirection()
        direction2 = image2.GetDirection()
        if any(abs(d1 - d2) > tolerance for d1, d2 in zip(direction1, direction2)):
            return False

        return True

    def _get_geometry_info(self, image: sitk.Image) -> dict:
        """Get geometry information for an image."""
        return {
            "size": image.GetSize(),
            "spacing": image.GetSpacing(),
            "origin": image.GetOrigin(),
            "direction": image.GetDirection(),
        }

    def _resample_to_reference(self, image: sitk.Image, reference: sitk.Image) -> sitk.Image:
        """Resample an image to match the geometry of a reference image."""
        resampler = sitk.ResampleImageFilter()
        resampler.SetReferenceImage(reference)
        resampler.SetDefaultPixelValue(0)
        resampler.SetInterpolator(sitk.sitkNearestNeighbor)  # Use nearest neighbor for segmentations
        return resampler.Execute(image)

    def _get_acquisition_plane(self, file_path: str) -> str:
        """Determine acquisition plane from metadata or filename."""
        # First check if we have this information in our dataframe
        for idx, row in self._scan_seg_pairs.iterrows():
            if row["file_path"] == file_path or row["seg_path"] == file_path:
                if "orientation" in row and pd.notna(row["orientation"]):
                    orientation = row["orientation"].lower()
                    if "axial" in orientation:
                        return "axial"
                    elif "sagittal" in orientation:
                        return "sagittal"
                    elif "coronal" in orientation:
                        return "coronal"

        # If not found in dataframe, try to infer from the image itself
        try:
            image = sitk.ReadImage(str(file_path))
            # Check image metadata for acquisition plane information
            for key in image.GetMetaDataKeys():
                if "plane" in key.lower() or "orientation" in key.lower():
                    value = image.GetMetaData(key).lower()
                    if "axial" in value:
                        return "axial"
                    elif "sagittal" in value:
                        return "sagittal"
                    elif "coronal" in value:
                        return "coronal"

            # If no metadata, infer from image dimensions
            spacing = image.GetSpacing()

            # Typically, the slice dimension has the largest spacing
            max_spacing_dim = np.argmax(spacing)
            if max_spacing_dim == 0:
                return "sagittal"
            elif max_spacing_dim == 1:
                return "coronal"
            else:
                return "axial"

        except Exception as e:
            warnings.warn(f"Could not determine acquisition plane for {file_path}: {e}")
            return "axial"  # Default to axial

    def _orient_image(self, image: sitk.Image, acquisition_plane: str) -> sitk.Image:
        """Orient a SimpleITK image based on the acquisition plane.

        Args:
            image: Input SimpleITK image
            acquisition_plane: Acquisition plane ("axial", "sagittal", "coronal")

        Returns:
            Oriented SimpleITK image
        """
        # Get appropriate orientation for the acquisition plane
        target_orientation = Orientation.get_orientation_for_plane(acquisition_plane, self._orientation.value)

        # Apply orientation using DICOMOrient
        oriented_image = sitk.DICOMOrient(image, target_orientation.value)

        if len(sitk.GetArrayFromImage(oriented_image).shape) > 3:
            raise ValueError("Multivolume cases are not supported.")

        return oriented_image

    def _sitk_to_array(self, image: sitk.Image) -> np.ndarray:
        """Convert SimpleITK image to properly oriented numpy array.

        Args:
            image: Input SimpleITK image

        Returns:
            Numpy array with depth as first dimension (d,h,w)
        """
        # Get array from SimpleITK image
        array = sitk.GetArrayFromImage(image)

        # Find the depth dimension (the one that differs from the other two)
        dims = array.shape
        if dims[1] == dims[2]:  # depth is in position 0
            return array
        elif dims[0] == dims[1]:  # depth is in position 2
            return array.transpose(2, 0, 1)
        else:  # depth is in position 1
            return array.transpose(1, 0, 2)

    def visualize_preprocessing_steps(
        self, file_path: str, output_dir: str, volume: np.ndarray, segmentation: np.ndarray
    ) -> None:
        """Visualize transforms with segmentation overlay."""
        if self._visualize_preprocessing:
            logger.info("\nVisualization details:")
            logger.info(f"Input volume shape: {volume.shape}")
            logger.info(f"Input segmentation shape: {segmentation.shape}")

        os.makedirs(output_dir, exist_ok=True)

        # Find the slice to visualize (same logic as before)
        if np.sum(segmentation) == 0:
            warnings.warn("Segmentation is completely empty (all zeros)")
            slice_to_show = volume.shape[0] // 2
        else:
            tumor_slices = []
            tumor_voxel_counts = []
            for i in range(segmentation.shape[0]):
                tumor_voxels = np.sum(segmentation[i] > 0)
                if tumor_voxels > 0:
                    tumor_slices.append(i)
                    tumor_voxel_counts.append(tumor_voxels)
            if tumor_slices:
                centroid_slice = int(np.average(tumor_slices, weights=tumor_voxel_counts))
                slice_to_show = centroid_slice
            else:
                slice_to_show = volume.shape[0] // 2

        # Get original image and mask for the selected slice
        original_image = volume[slice_to_show].astype(np.float32)

        # Use the same cropping function as in actual loading
        cropped_volume, cropped_segmentation = self._get_cropped_volume_and_segmentation(file_path)

        # Adjust slice index if needed after cropping
        cropped_slice = min(slice_to_show, cropped_volume.shape[0] - 1)
        cropped_image = cropped_volume[cropped_slice].astype(np.float32)
        cropped_mask = cropped_segmentation[cropped_slice].astype(np.float32)

        if self._visualize_preprocessing:
            logger.info("\nVisualization details:")
            logger.info(f"Original volume shape: {volume.shape}")
            logger.info(f"Cropped volume shape: {cropped_volume.shape}")
            logger.info(f"Original slice index: {slice_to_show}")
            logger.info(f"Adjusted slice index: {cropped_slice}")
            logger.info(f"Original slice shape: {original_image.shape}")
            logger.info(f"Cropped slice shape: {cropped_image.shape}")

        # Convert to tensors for transforms
        current_image = tv_tensors.Image(cropped_image)
        current_mask = tv_tensors.Mask(cropped_mask)

        def normalize_for_display(img_array):
            """Normalize image for display."""
            if isinstance(img_array, torch.Tensor):
                img_array = img_array.detach().cpu().numpy()

            img_min, img_max = np.percentile(img_array, (1, 99))
            normalized = np.clip(img_array, img_min, img_max)
            normalized = (normalized - normalized.min()) / (normalized.max() - normalized.min() + 1e-8)
            return normalized

        def create_rgb_image(image):
            """Convert grayscale to RGB."""
            if isinstance(image, (tv_tensors.Image, torch.Tensor)):
                # Handle channel dimension if present
                if image.ndim > 2:
                    image = image.squeeze(0)  # Remove channel dim if present
                image = image.detach().cpu().numpy()

            # Normalize to [0,1]
            img_normalized = normalize_for_display(image)

            # Create RGB image
            if img_normalized.ndim == 2:
                return np.stack([img_normalized] * 3, axis=-1)
            else:
                # Already RGB
                return img_normalized

        def overlay_segmentation(image, segmentation, alpha=0.5):
            """Create RGB overlay of segmentation on image with transparency."""
            # Convert image to RGB
            rgb = create_rgb_image(image)

            # Convert segmentation to binary mask
            if isinstance(segmentation, (tv_tensors.Mask, torch.Tensor)):
                segmentation = segmentation.detach().cpu().numpy()

            # Ensure binary mask
            segmentation = segmentation > 0.5

            # Add semi-transparent red overlay for segmentation
            if np.any(segmentation):
                # Create a copy of the original image for the segmentation area
                seg_area = rgb.copy()
                seg_area[segmentation] = [1, 0, 0]  # Pure red

                # Blend with original using alpha
                rgb = rgb * (1 - alpha * segmentation[:, :, np.newaxis]) + seg_area * (
                    alpha * segmentation[:, :, np.newaxis]
                )

            return rgb

        # Apply transforms
        if self._transforms is not None:
            transformed_image, transformed_mask = self._apply_transforms(current_image, current_mask)
        else:
            transformed_image = current_image
            transformed_mask = current_mask

        # Apply image-only transforms
        if self._image_transforms is not None:
            transformed_image = transformed_image.contiguous()
            normalized_image = self._image_transforms(transformed_image)
            final_image = normalized_image
        else:
            final_image = transformed_image

        # Create visualization steps
        transform_steps = [
            ("Original", create_rgb_image(original_image)),
            ("After Cropping", create_rgb_image(cropped_image)),
            ("Final with Segmentation", overlay_segmentation(final_image, transformed_mask)),
        ]

        # Create figure
        n_steps = len(transform_steps)
        fig, axes = plt.subplots(1, n_steps, figsize=(5 * n_steps, 5))
        if n_steps == 1:
            axes = [axes]
        fig.suptitle(f"Transform Steps for {os.path.basename(file_path)}")

        for ax, (title, img) in zip(axes, transform_steps):
            ax.imshow(img)
            ax.axis("off")
            ax.set_title(title)

        plt.tight_layout()
        output_path = os.path.join(
            output_dir, f"{os.path.splitext(os.path.basename(file_path))[0]}_transform_steps.png"
        )
        plt.savefig(output_path, bbox_inches="tight", dpi=150)
        plt.close()

        if self._visualize_preprocessing:
            print(f"\nVisualization saved to: {output_path}")
            print(f"Showing slice {slice_to_show} out of {volume.shape[0]} slices")
            print(f"Original image shape: {original_image.shape}")
            print(f"Cropped image shape: {cropped_image.shape}")

    def _save_slice_visualization(self, array: np.ndarray, sample_index: int, slice_index: int, type_str: str) -> None:
        """Save a slice visualization as PNG.

        Args:
            array: The image or mask array to save
            sample_index: The sample index for folder naming
            slice_index: The slice index for file naming
            type_str: Either 'img' or 'mask' for file naming
        """
        # Create directories if they don't exist
        viz_dir = os.path.join("visualizations", f"sample_{sample_index}")
        os.makedirs(viz_dir, exist_ok=True)

        # Create the figure
        plt.figure(figsize=(8, 8))
        if type_str == "img":
            # Normalize image for better visualization
            vmin, vmax = np.percentile(array, (1, 99))
            plt.imshow(array, cmap="gray", vmin=vmin, vmax=vmax)
        else:  # mask
            plt.imshow(array, cmap="tab20")

        plt.axis("off")
        plt.title(f"Sample {sample_index}, Slice {slice_index}")

        # Save the figure
        output_path = os.path.join(viz_dir, f"slice_{slice_index}_{type_str}.png")
        plt.savefig(output_path, bbox_inches="tight", pad_inches=0, dpi=150)
        plt.close()

        if self._visualize_preprocessing:
            print(f"Saved visualization to: {output_path}")


# Example usage / test code
if __name__ == "__main__":
    import torchvision.transforms as T

    # Set up test paths and transforms
    data_dir = "/data/groups/aiforoncology/derived/radiology/mskcc-breast/nrrd_compressed"
    output_dir = "visualizations"

    # Basic transforms pipeline
    transforms = T.Compose([T.RandomHorizontalFlip(p=0.5), T.RandomRotation(10)])

    image_transforms = T.Compose([T.Normalize(mean=[0.5], std=[0.5])])

    # Create dataset instance
    dataset = MRITumorSegDataset(
        label_file="/data/groups/aiforoncology/derived/radiology/mskcc-breast/meta/600_post_contrast_withseg.csv",
        orientation="LPS",
        orientation_plane="axial",
        contrast_type="post",
        crop_percentile=80.0,  # Specify the cropping threshold here
        split="train",
        visualize_preprocessing=False,
        transforms=transforms,
        image_transforms=image_transforms,
    )

    # Test loading a few samples
    for i in range(min(3, len(dataset))):
        sample = dataset[i]
        print(f"Sample {i}:")
        print(f"Image shape: {sample['image'].shape}")
        print(f"Mask shape: {sample['mask'].shape}")
        print("-" * 50)
