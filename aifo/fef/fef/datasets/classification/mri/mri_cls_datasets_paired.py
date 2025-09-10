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

""" "Breast MRI Classification datasets for paired pre/post contrast scans."""

import functools
import os
from typing import Any, Callable, Literal

import numpy as np
import pandas as pd
from torchvision import tv_tensors
from typing_extensions import override
import logging

from .base_mri_cls_dataset import BaseMRIClassificationDataset
from fef.utils.helpers.lrucache import LRUCache

logger = logging.getLogger(__name__)


class PairedMRIClassificationDataset(BaseMRIClassificationDataset):
    """MRI Classification dataset for paired pre/post contrast scans (single-slice mode)."""

    def __init__(
        self,
        label_file: str,
        split: Literal["train", "val", "test"] | None = None,
        transforms: Callable | None = None,
        image_transforms: Callable | None = None,
        split_ratios: tuple[float, float, float] = (0.7, 0.15, 0.15),
        orientation: str = "LPS",
        orientation_plane: Literal["axial", "sagittal", "all"] = "axial",
        contrast_names: list[str] = ["pre", "post_1"],
        slice_mode: Literal["all", "middle"] = "all",
        crop_percentile: float | None = None,
        visualize_preprocessing: bool = False,
        seed: int = 42,
        tumor_ratio: float = 0.5,
        cache_size: int = 10,
        label_name: str = "Overall_pcr",
        min_class_percentage: float | None = None,
    ) -> None:
        assert len(contrast_names) > 1, "At least two contrast names must be specified for paired datasets"
        self._contrast_names = contrast_names

        super().__init__(
            label_file=label_file,
            split=split,
            transforms=transforms,
            image_transforms=image_transforms,
            split_ratios=split_ratios,
            orientation=orientation,
            orientation_plane=orientation_plane,
            slice_mode=slice_mode,
            crop_percentile=crop_percentile,
            visualize_preprocessing=visualize_preprocessing,
            seed=seed,
            tumor_ratio=tumor_ratio,
            cache_size=cache_size,
            label_name=label_name,
            min_class_percentage=min_class_percentage,
        )
        self._paired_cropped_cache = LRUCache(cache_size)

    @override
    def filename(self, index: int) -> str:
        """Return filename for paired scan including pre/post information."""
        sample_index, slice_index = self._indices[index]
        volume_path = self._volume_files[sample_index]

        label_row = self._scan_label_pairs[self._scan_label_pairs["file_path"] == volume_path]
        contrast = label_row["contrast"].values[0] if "contrast" in label_row.columns else "unknown"

        base_filename = os.path.basename(volume_path)
        name, ext = os.path.splitext(base_filename)
        return f"{name}_{contrast}_slice_{slice_index}{ext}"

    @override
    def load_metadata(self, index: int) -> dict[str, Any]:
        """Return metadata for paired scans including all paired scan info."""
        sample_index, slice_index = self._indices[index]
        volume_path = self._volume_files[sample_index]

        label_row = self._scan_label_pairs[self._scan_label_pairs["file_path"] == volume_path]

        pair = self._get_volume_pair(volume_path)
        paired_paths = pair if pair else (volume_path,)

        metadata = {
            "patient_id": label_row["patient_id"].values[0] if "patient_id" in label_row.columns else "unknown",
            "dataset": label_row["dataset"].values[0] if "dataset" in label_row.columns else "unknown",
            "acquisition_date": label_row["acquisition_date"].values[0]
            if "acquisition_date" in label_row.columns
            else "unknown",
            "sample_index": int(sample_index),
            "slice_index": int(slice_index),
            "contrast": label_row["contrast"].values[0] if "contrast" in label_row.columns else "unknown",
            "contrast_names": self._contrast_names,
        }

        for i, contrast_name in enumerate(self._contrast_names):
            if i < len(paired_paths):
                metadata[f"paired_{contrast_name}_path"] = paired_paths[i]

        return metadata

    @functools.cached_property
    def _scan_label_pairs(self) -> pd.DataFrame:
        """Load and filter paired scan-label pairs from manifest file."""
        try:
            df = pd.read_csv(self._label_file)

            required_columns = ["file_path", self._label_name, "contrast", "patient_id"]
            missing_columns = [col for col in required_columns if col not in df.columns]
            if missing_columns:
                raise ValueError(f"Missing required columns for paired dataset: {missing_columns}")

            df_filtered = df.dropna(subset=[self._label_name])

            if len(df_filtered) == 0:
                raise ValueError("No valid scan-label pairs found!")

            df_filtered = self._ensure_paired_scans(df_filtered)

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

            logger.info(f"Determined {len(df_filtered)} valid paired scan-label pairs")
            return df_filtered

        except Exception as e:
            raise RuntimeError(f"Error reading paired manifest file: {str(e)}")

    def _ensure_paired_scans(self, df: pd.DataFrame) -> pd.DataFrame:
        """Ensure that every scan has all required contrast types for pairing."""
        grouped = df.groupby("patient_id")
        valid_patients = []

        logger.info(f"Checking pairing for {len(grouped)} patients with contrast names: {self._contrast_names}")

        for patient_id, group in grouped:
            contrast_values = set(group["contrast"].values)

            has_required_contrasts = all(contrast in contrast_values for contrast in self._contrast_names)

            if has_required_contrasts:
                valid_patients.append(patient_id)
                logger.debug(f"Patient {patient_id}: Valid pair found with contrasts {contrast_values}")
            else:
                missing_contrasts = set(self._contrast_names) - contrast_values
                logger.warning(
                    f"Patient {patient_id}: Missing contrasts {missing_contrasts} - available: {contrast_values}"
                )

        if not valid_patients:
            available_contrasts = df["contrast"].unique()
            raise ValueError(
                f"No valid pairs found for contrast names {self._contrast_names}! "
                f"Available contrasts: {available_contrasts.tolist()}"
            )

        df_paired = df[(df["patient_id"].isin(valid_patients)) & (df["contrast"].isin(self._contrast_names))]

        logger.info(f"Found {len(valid_patients)} valid paired patients")
        logger.info(f"Total scans after pairing and contrast filtering: {len(df_paired)}")

        return df_paired

    @functools.cached_property
    def _volume_files(self) -> list[str]:
        """Get list of all volume file paths that are part of valid pairs."""
        paired_files = self._scan_label_pairs["file_path"].tolist()

        logger.info(f"Found {len(paired_files)} volumes with specified contrasts")
        return paired_files

    def _get_volume_pair(self, volume_path: str) -> tuple[str, ...] | None:
        """Get the paired volumes for a given volume path."""
        volume_row = self._scan_label_pairs[self._scan_label_pairs["file_path"] == volume_path]
        if volume_row.empty:
            return None

        patient_id = volume_row["patient_id"].values[0]

        patient_files = self._scan_label_pairs[self._scan_label_pairs["patient_id"] == patient_id]

        paired_paths = []
        for contrast_name in self._contrast_names:
            contrast_files = patient_files[patient_files["contrast"] == contrast_name]
            if contrast_files.empty:
                return None
            paired_paths.append(contrast_files["file_path"].values[0])

        return tuple(paired_paths)

    @override
    def load_data(self, index: int) -> tv_tensors.Image:
        """Load image data for paired scans with cropping applied."""
        try:
            sample_index, slice_index = self._indices[index]
            volume_path = self._volume_files[sample_index]

            pair = self._get_volume_pair(volume_path)
            if pair is None:
                raise ValueError(f"Could not find pair for volume {volume_path}")

            cache_key = "|".join(pair)
            cached_data = self._paired_cropped_cache.get(cache_key)

            if cached_data is None:
                volume_arrays = self._load_volume_pair(*pair)
                self._paired_cropped_cache.put(cache_key, volume_arrays)
            else:
                volume_arrays = cached_data

            volume_row = self._scan_label_pairs[self._scan_label_pairs["file_path"] == volume_path]
            current_contrast = volume_row["contrast"].values[0]

            try:
                volume_index = self._contrast_names.index(current_contrast)
            except ValueError:
                raise ValueError(f"Contrast {current_contrast} not found in contrast_names {self._contrast_names}")

            volume = volume_arrays[volume_index]

            if slice_index >= volume.shape[0]:
                raise ValueError(f"Slice index {slice_index} out of bounds for volume shape {volume.shape}")

            image_array = volume[slice_index].astype(np.float32)
            image_array = np.expand_dims(image_array, axis=0)
            image_tensor = tv_tensors.Image(image_array)

            return self._apply_image_transforms(image_tensor)

        except Exception as e:
            raise RuntimeError(f"Error loading paired image at index {index}: {str(e)}")

    def _load_volume_pair(self, *volume_paths: str) -> tuple[np.ndarray, ...]:
        """Load multiple paired volumes with cropping applied."""
        try:
            arrays = []
            for path in volume_paths:
                array = self._get_cropped_volume(path)
                arrays.append(array)

            return tuple(arrays)

        except Exception as e:
            raise RuntimeError(f"Error loading volume pair {volume_paths}: {str(e)}")


class PairedMRIMILClassificationDataset(BaseMRIClassificationDataset):
    """MRI Classification dataset for paired pre/post contrast scans (MIL mode)."""

    def __init__(
        self,
        label_file: str,
        split: Literal["train", "val", "test"] | None = None,
        transforms: Callable | None = None,
        image_transforms: Callable | None = None,
        split_ratios: tuple[float, float, float] = (0.7, 0.15, 0.15),
        orientation: str = "LPS",
        orientation_plane: Literal["axial", "sagittal", "all"] = "axial",
        contrast_names: list[str] = ["pre", "post_1"],
        slice_mode: Literal["all", "middle"] = "all",
        crop_percentile: float | None = None,
        visualize_preprocessing: bool = False,
        seed: int = 42,
        tumor_ratio: float = 0.5,
        cache_size: int = 10,
        label_name: str = "Overall_pcr",
        min_class_percentage: float | None = None,
    ) -> None:
        assert len(contrast_names) > 1, "At least two contrast names must be specified for paired datasets"
        self._contrast_names = contrast_names

        super().__init__(
            label_file=label_file,
            split=split,
            transforms=transforms,
            image_transforms=image_transforms,
            split_ratios=split_ratios,
            orientation=orientation,
            orientation_plane=orientation_plane,
            slice_mode=slice_mode,
            crop_percentile=crop_percentile,
            visualize_preprocessing=visualize_preprocessing,
            seed=seed,
            tumor_ratio=tumor_ratio,
            cache_size=cache_size,
            label_name=label_name,
            min_class_percentage=min_class_percentage,
        )
        self._paired_cropped_cache = LRUCache(cache_size)

    @override
    def filename(self, index: int) -> str:
        """Return filename for MIL paired scans (volume-level identifier)."""
        sample_index, _ = self._indices[index]
        volume_path = self._volume_files[sample_index]

        pair = self._get_volume_pair(volume_path)
        if pair:
            pre_path, post_path = pair
            label_row = self._scan_label_pairs[self._scan_label_pairs["file_path"] == volume_path]
            patient_id = label_row["patient_id"].values[0] if "patient_id" in label_row.columns else "unknown"
            contrast = label_row["contrast"].values[0] if "contrast" in label_row.columns else "unknown"
            return f"patient_{patient_id}_{contrast}"
        else:
            # Just get the actual filename, not the path
            return os.path.basename(volume_path)

    @override
    def load_metadata(self, index: int) -> dict[str, Any]:
        """Return metadata for MIL paired scans."""
        sample_index, slice_index = self._indices[index]
        volume_path = self._volume_files[sample_index]

        label_row = self._scan_label_pairs[self._scan_label_pairs["file_path"] == volume_path]

        pair = self._get_volume_pair(volume_path)
        paired_paths = pair if pair else (volume_path,)

        num_slices = self._get_number_of_slices_per_volume(sample_index)

        metadata = {
            "patient_id": label_row["patient_id"].values[0] if "patient_id" in label_row.columns else "unknown",
            "dataset": label_row["dataset"].values[0] if "dataset" in label_row.columns else "unknown",
            "acquisition_date": label_row["acquisition_date"].values[0]
            if "acquisition_date" in label_row.columns
            else "unknown",
            "sample_index": int(sample_index),
            "slice_index": int(slice_index),
            "contrast": label_row["contrast"].values[0] if "contrast" in label_row.columns else "unknown",
            "contrast_names": self._contrast_names,
            "multi_id": int(sample_index),  # Volume identifier for MIL
            "middle_slice": slice_index == num_slices // 2,
        }

        for i, contrast_name in enumerate(self._contrast_names):
            if i < len(paired_paths):
                metadata[f"paired_{contrast_name}_path"] = paired_paths[i]

        return metadata

    @functools.cached_property
    def _scan_label_pairs(self) -> pd.DataFrame:
        """Load and filter paired scan-label pairs from manifest file."""
        try:
            df = pd.read_csv(self._label_file)

            required_columns = ["file_path", self._label_name, "contrast", "patient_id"]
            missing_columns = [col for col in required_columns if col not in df.columns]
            if missing_columns:
                raise ValueError(f"Missing required columns for paired dataset: {missing_columns}")

            df_filtered = df.dropna(subset=[self._label_name])

            if len(df_filtered) == 0:
                raise ValueError("No valid scan-label pairs found!")

            df_filtered = self._ensure_paired_scans(df_filtered)

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

            logger.info(f"Determined {len(df_filtered)} valid paired scan-label pairs")
            return df_filtered

        except Exception as e:
            raise RuntimeError(f"Error reading paired manifest file: {str(e)}")

    def _ensure_paired_scans(self, df: pd.DataFrame) -> pd.DataFrame:
        """Ensure that every scan has all required contrast types for pairing."""
        grouped = df.groupby("patient_id")
        valid_patients = []

        logger.info(f"Checking pairing for {len(grouped)} patients with contrast names: {self._contrast_names}")

        for patient_id, group in grouped:
            contrast_values = set(group["contrast"].values)

            has_required_contrasts = all(contrast in contrast_values for contrast in self._contrast_names)

            if has_required_contrasts:
                valid_patients.append(patient_id)
                logger.debug(f"Patient {patient_id}: Valid pair found with contrasts {contrast_values}")
            else:
                missing_contrasts = set(self._contrast_names) - contrast_values
                logger.warning(
                    f"Patient {patient_id}: Missing contrasts {missing_contrasts} - available: {contrast_values}"
                )

        if not valid_patients:
            available_contrasts = df["contrast"].unique()
            raise ValueError(
                f"No valid pairs found for contrast names {self._contrast_names}! "
                f"Available contrasts: {available_contrasts.tolist()}"
            )

        df_paired = df[(df["patient_id"].isin(valid_patients)) & (df["contrast"].isin(self._contrast_names))]

        logger.info(f"Found {len(valid_patients)} valid paired patients")
        logger.info(f"Total scans after pairing and contrast filtering: {len(df_paired)}")

        return df_paired

    @functools.cached_property
    def _volume_files(self) -> list[str]:
        """Get list of all volume file paths that are part of valid pairs."""

        paired_files = self._scan_label_pairs["file_path"].tolist()

        logger.info(f"Found {len(paired_files)} volumes with specified contrasts")
        return paired_files

    def _get_volume_pair(self, volume_path: str) -> tuple[str, ...] | None:
        """Get the paired volumes for a given volume path."""
        volume_row = self._scan_label_pairs[self._scan_label_pairs["file_path"] == volume_path]
        if volume_row.empty:
            return None

        patient_id = volume_row["patient_id"].values[0]

        patient_files = self._scan_label_pairs[self._scan_label_pairs["patient_id"] == patient_id]

        paired_paths = []
        for contrast_name in self._contrast_names:
            contrast_files = patient_files[patient_files["contrast"] == contrast_name]
            if contrast_files.empty:
                return None
            paired_paths.append(contrast_files["file_path"].values[0])

        return tuple(paired_paths)

    @override
    def load_data(self, index: int) -> tv_tensors.Image:
        """Load image data for paired scans with cropping applied."""
        try:
            sample_index, slice_index = self._indices[index]
            volume_path = self._volume_files[sample_index]

            pair = self._get_volume_pair(volume_path)
            if pair is None:
                raise ValueError(f"Could not find pair for volume {volume_path}")

            cache_key = "|".join(pair)
            cached_data = self._paired_cropped_cache.get(cache_key)

            if cached_data is None:
                volume_arrays = self._load_volume_pair(*pair)
                self._paired_cropped_cache.put(cache_key, volume_arrays)
            else:
                volume_arrays = cached_data

            volume_row = self._scan_label_pairs[self._scan_label_pairs["file_path"] == volume_path]
            current_contrast = volume_row["contrast"].values[0]
            try:
                volume_index = self._contrast_names.index(current_contrast)
            except ValueError:
                raise ValueError(f"Contrast {current_contrast} not found in contrast_names {self._contrast_names}")

            volume = volume_arrays[volume_index]

            if slice_index >= volume.shape[0]:
                raise ValueError(f"Slice index {slice_index} out of bounds for volume shape {volume.shape}")

            image_array = volume[slice_index].astype(np.float32)
            image_array = np.expand_dims(image_array, axis=0)
            image_tensor = tv_tensors.Image(image_array)

            return self._apply_image_transforms(image_tensor)

        except Exception as e:
            raise RuntimeError(f"Error loading paired image at index {index}: {str(e)}")

    def _load_volume_pair(self, *volume_paths: str) -> tuple[np.ndarray, ...]:
        """Load multiple paired volumes with cropping applied."""
        try:
            arrays = []
            for path in volume_paths:
                array = self._get_cropped_volume(path)
                arrays.append(array)

            return tuple(arrays)

        except Exception as e:
            raise RuntimeError(f"Error loading volume pair {volume_paths}: {str(e)}")
