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

import csv
import logging
from multiprocessing import Process
from typing import Callable, Optional

import numpy as np
import pandas as pd
import SimpleITK as sitk
from fomo.database_models import Image
from fomo.dataset.vector_dataset.data_manager import DataManager
from fomo.mri.dataset.cropping import crop_to_bbox_percentile

logger = logging.getLogger(__name__)


class ImageProcessor:
    def __init__(
        self,
        database_url: str,
        output_csv_prefix: str,
        num_workers: int,
        preprocess_func: Optional[Callable] = None,
    ):
        """
        Initialize the ImageProcessor for processing images with multiple workers.

        Parameters
        ----------
        database_url : str
            Database connection URL.
        output_csv_prefix : str
            Prefix for the output CSV files.
        num_workers : int
            Number of worker processes.
        preprocess_func : callable, optional
            Function to preprocess the images.
        """
        self.database_url = database_url
        self.output_csv_prefix = output_csv_prefix
        self.num_workers = num_workers
        self.preprocess_func = self._default_preprocess if preprocess_func is None else preprocess_func
        self.processes = []

    def process_images(self) -> None:
        """Process images using multiple workers."""
        for worker_id in range(self.num_workers):
            process = Process(target=self._worker_process, args=(worker_id,))
            process.start()
            self.processes.append(process)
        # Wait for all processes to finish
        for process in self.processes:
            process.join()

    def _compute_mean_std_incrementally(self, slices):
        """Compute mean and std over slices incrementally to save memory."""
        n_total = 0
        mean_total = 0.0
        m2_total = 0.0  # Sum of squares of differences from the current mean

        for slice_data in slices:
            pixels = slice_data.flatten()
            n = pixels.size
            mean = np.mean(pixels)
            m2 = np.var(pixels) * n  # Variance times n

            # Update totals
            delta = mean - mean_total
            n_total += n
            mean_total += delta * n / n_total
            m2_total += m2 + delta**2 * n * (n_total - n) / n_total

        variance = m2_total / n_total
        std_total = np.sqrt(variance)
        return mean_total, std_total

    def _worker_process(self, worker_id: int) -> None:
        """Worker process to process a subset of images."""
        data_manager = DataManager(self.database_url)
        # Get images with offset and step to divide the work among workers
        images = data_manager.get_images_with_offset(offset=worker_id, step=self.num_workers)
        output_csv = f"{self.output_csv_prefix}_worker_{worker_id}.csv"
        with open(output_csv, mode="w", newline="") as csv_file:
            fieldnames = ["filename", "mean", "std"]
            writer = csv.DictWriter(csv_file, fieldnames=fieldnames)
            writer.writeheader()
            for idx, image_record in enumerate(images):
                try:
                    if idx % 100 == 0:
                        print(f"Worker {worker_id}: Processing image {idx}")
                    image_data = self._read_image(image_record)
                    slices = self.preprocess_func(image_data)
                    mean_value, std_value = self._compute_mean_std_incrementally(slices)
                    filename = image_record.filename
                    # Record the filename, mean_value, std_value
                    self._record_statistics(writer, filename, mean_value, std_value)
                    csv_file.flush()
                except Exception as e:
                    logger.error(f"Worker {worker_id}: Error processing image ID {image_record.id}: {e}")

    def _read_image(self, image_record: Image) -> sitk.Image:
        """Read an image file from the image record and reorient it to LPS."""
        image = sitk.ReadImage(image_record.filename)
        return self._reorient_image_to_lps(image)

    def _reorient_image_to_lps(self, image: sitk.Image) -> sitk.Image:
        """Reorient an image to LPS orientation."""
        return sitk.DICOMOrient(image, "LPS")

    def _default_preprocess(self, image: sitk.Image) -> list[np.ndarray]:
        """Default preprocessing function for images."""
        size = image.GetSize()[-1]
        spacing = image.GetSpacing()[-1]
        center_slice = size // 2
        slices = int(75 / spacing)  # 15 cm range
        start_slice = max(0, center_slice - slices)
        end_slice = min(size, center_slice + slices)

        image_array = sitk.GetArrayViewFromImage(image)
        processed_slices = []
        for idx in range(start_slice, end_slice):
            slice_data = image_array[idx].astype(np.float32)
            cropped_slice = crop_to_bbox_percentile(slice_data, lower_percentile=30.0)
            processed_slices.append(cropped_slice)

        return processed_slices

    def _record_statistics(self, writer, filename: str, mean_value: float, std_value: float):
        """Record the filename, mean, and std to the CSV file."""
        writer.writerow({"filename": filename, "mean": mean_value, "std": std_value})


def combine_csv_files(output_csv_prefix, num_workers, final_output_csv):
    csv_files = [f"{output_csv_prefix}_worker_{i}.csv" for i in range(num_workers)]
    df_list = [pd.read_csv(csv_file) for csv_file in csv_files]
    combined_df = pd.concat(df_list, ignore_index=True)
    combined_df.to_csv(final_output_csv, index=False)


if __name__ == "__main__":
    print("Running ImageProcessor")
    database_url = "sqlite:////projects/mri_fomo/database/nki_breast_mri_nrrd_t1_pre_post.db"
    output_csv_prefix = "output_statistics"
    num_workers = 6  # Adjust based on your system's capabilities
    processor = ImageProcessor(database_url, output_csv_prefix, num_workers)
    processor.process_images()

    final_output_csv = "combined_output_statistics.csv"
    combine_csv_files(output_csv_prefix, num_workers, final_output_csv)
    logger.info(f"Combined output CSV file saved to {final_output_csv}")
