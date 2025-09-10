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

import logging
from typing import Callable, Optional

import numpy as np
import SimpleITK as sitk
from fomo.database_models import Image
from fomo.dataset.vector_dataset.data_producer import DataProducer
from fomo.mri.dataset.cropping import crop_to_bbox_percentile

logger = logging.getLogger(__name__)


class MRIDataProducer(DataProducer):
    def __init__(
        self,
        name: str,
        chunk_size_bytes: int,
        max_memory_size_bytes: int,
        database_url: str,
        num_workers: int,
        preprocess_func: Optional[Callable] = None,
        min_mask_size: int = 50,
    ):
        super().__init__(name, chunk_size_bytes, max_memory_size_bytes, database_url, num_workers)
        self.preprocess_func = self._default_preprocess_mask if preprocess_func is None else preprocess_func
        self.min_mask_size = min_mask_size

    # override, method-annotation only available in python 3.12
    def _load_and_preprocess_image(self, db_entry: Image) -> list[np.ndarray]:
        """Process an image and append its slices to the shared vector."""
        image_data, mask_data = self._read_image_and_mask(db_entry)
        return self.preprocess_func(image_data, mask_data)

    def _read_image_and_mask(self, image_record: Image) -> tuple[sitk.Image, Optional[sitk.Image]]:
        """Read an image file and its mask (if any), and reorient them to LPS."""
        image = sitk.ReadImage(image_record.filename)
        image = self._reorient_image_to_lps(image)

        mask_data = None
        if image_record.mask and image_record.mask.filename:
            try:
                mask = sitk.ReadImage(image_record.mask.filename)
                mask = self._reorient_image_to_lps(mask)
                mask_data = mask
            except Exception as e:
                logger.error(f"Error reading mask for image ID {image_record.id}: {e}")

        return image, mask_data

    def _default_preprocess(self, image: sitk.Image, mask: Optional[sitk.Image]) -> list[np.ndarray]:
        """Default preprocessing function for images."""
        size = image.GetSize()[-1]
        spacing = image.GetSpacing()[-1]
        center_slice = size // 2
        slices = int(75 / spacing)  # 15 cm range
        start_slice = max(0, center_slice - slices)
        end_slice = min(size, center_slice + slices)

        image_array = sitk.GetArrayFromImage(image)
        processed_slices = []
        for idx in range(start_slice, end_slice):
            slice_data = image_array[idx].astype(np.float32)
            cropped_slice = crop_to_bbox_percentile(slice_data, lower_percentile=30.0)
            contiguous_slice = np.ascontiguousarray(cropped_slice)
            processed_slices.append(contiguous_slice)

        return processed_slices

    def _default_preprocess_mask(self, image: sitk.Image, mask: Optional[sitk.Image]) -> list[np.ndarray]:
        """Default preprocessing function for images, using mask to crop if available."""
        try:
            image_array = sitk.GetArrayFromImage(image)
            processed_slices = []

            if mask is not None:
                mask_array = sitk.GetArrayFromImage(mask)
                try:
                    min_coords, max_coords = self.get_bounding_box(mask_array)
                    size = max_coords - min_coords + 1

                    if np.any(size < self.min_mask_size):
                        logger.warning(f"Skipping mask with dimensions {size} (minimum size: {self.min_mask_size})")
                        return []
                except Exception as e:
                    logger.error(f"Mask is invalid: {e}")
                    return []

                minZ, minY, minX = min_coords
                maxZ, maxY, maxX = max_coords

                # Ensure indices are within bounds
                minZ = max(0, minZ)
                maxZ = min(image_array.shape[0] - 1, maxZ)
                minY = max(0, minY)
                maxY = min(image_array.shape[1] - 1, maxY)
                minX = max(0, minX)
                maxX = min(image_array.shape[2] - 1, maxX)
            else:
                # No mask provided, use default center slices
                size = image_array.shape[0]
                spacing = image.GetSpacing()[-1]
                center_slice = size // 2
                slices = int(75 / spacing)  # 15 cm range
                minZ = max(0, center_slice - slices)
                maxZ = min(size - 1, center_slice + slices)
                minY = 0
                maxY = image_array.shape[1] - 1
                minX = 0
                maxX = image_array.shape[2] - 1

            for idx in range(int(minZ), int(maxZ) + 1):
                slice_data = image_array[idx].astype(np.float32)
                cropped_slice = slice_data[int(minY) : int(maxY) + 1, int(minX) : int(maxX) + 1]
                contiguous_slice = np.ascontiguousarray(cropped_slice)
                processed_slices.append(contiguous_slice)

            return processed_slices

        except Exception as e:
            logger.error(f"Error during preprocessing: {e}")
            return []  # Skips this image

    def get_bounding_box(self, array: np.ndarray) -> tuple[tuple[int, int, int], tuple[int, int, int]]:
        """
        Compute the bounding box of a binary mask in an array.
        Returns min_coords and max_coords (inclusive indices).
        """
        non_zero_indices = np.argwhere(array > 0)
        if non_zero_indices.size == 0:
            raise ValueError("No non-zero pixels found in the mask.")

        min_coords = non_zero_indices.min(axis=0)
        max_coords = non_zero_indices.max(axis=0)
        return min_coords, max_coords
