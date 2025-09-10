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

from pathlib import Path
from typing import Callable
import re
import numpy as np
import SimpleITK as sitk
import torch
from PIL import Image
from fomo.utils.types import Dimensionality, Orientation
from torch.utils.data import Dataset
import logging

logger = logging.getLogger(__name__)

# Try to import sqlalchemy, but make it optional
try:
    from sqlalchemy import create_engine, text

    SQLALCHEMY_AVAILABLE = True
except ImportError:
    SQLALCHEMY_AVAILABLE = False
    logger.warning("sqlalchemy not available. Database functionality will be limited.")


class DebugMRIDataset(Dataset):
    def __init__(
        self,
        data_source: str,  # Can be either directory path or database URL
        orientation: str,
        mode: str,
        transform: None | Callable = None,
        target_transform: None | Callable = None,
        save_crops: bool = False,
        crops_output_dir: str = None,
        max_crops_per_patient: int = 50,
        max_nrrds_per_patient: int = None,  # None means process all NRRDs
        max_slices_per_nrrd: int = 1,  # Only take middle slice
        max_random_studies: int = None,  # If set, randomly select this many studies
    ):
        self.data_source = data_source

        # Determine if data_source is a database URL or directory path
        if self._is_database_url(data_source):
            self._file_paths = self._get_file_paths_from_database(data_source)
        else:
            # Treat as directory path (backward compatibility)
            self._file_paths = self._find_files_with_extension(data_source, "nrrd")

        # Group files by patient
        patient_files = {}
        for file_path in self._file_paths:
            patient_id = self._extract_patient_id(str(file_path))
            if patient_id not in patient_files:
                patient_files[patient_id] = []
            patient_files[patient_id].append(file_path)

        # Randomly select patients if max_random_studies is set
        # (Note: keeping the parameter name for backward compatibility)
        if max_random_studies is not None and max_random_studies < len(patient_files):
            import random

            # Sort patients by number to divide into early/middle/late groups
            sorted_patients = sorted(list(patient_files.keys()))
            n_patients = len(sorted_patients)

            # Calculate how many patients to take from each group
            patients_per_group = max_random_studies // 3
            remainder = max_random_studies % 3

            # Define the three groups (early, middle, late)
            early_patients = sorted_patients[: n_patients // 3]
            middle_patients = sorted_patients[n_patients // 3 : 2 * n_patients // 3]
            late_patients = sorted_patients[2 * n_patients // 3 :]

            # Select patients from each group
            selected_patients = []
            selected_patients.extend(random.sample(early_patients, patients_per_group))
            selected_patients.extend(random.sample(middle_patients, patients_per_group))
            selected_patients.extend(random.sample(late_patients, patients_per_group))

            # Add any remainder patients randomly from any group
            if remainder > 0:
                remaining_patients = list(set(sorted_patients) - set(selected_patients))
                selected_patients.extend(random.sample(remaining_patients, remainder))

            logger.info(f"Randomly selected {max_random_studies} patients out of {len(patient_files)} total patients:")
            logger.info("Early patients:")
            for patient in sorted([p for p in selected_patients if p in early_patients]):
                logger.info(f"  - {patient} ({len(patient_files[patient])} files)")
            logger.info("Middle patients:")
            for patient in sorted([p for p in selected_patients if p in middle_patients]):
                logger.info(f"  - {patient} ({len(patient_files[patient])} files)")
            logger.info("Late patients:")
            for patient in sorted([p for p in selected_patients if p in late_patients]):
                logger.info(f"  - {patient} ({len(patient_files[patient])} files)")
            # Filter file paths to only include selected patients
            selected_files = []
            for patient in selected_patients:
                selected_files.extend(patient_files[patient])
            self._file_paths = selected_files

            # Update patient_files to only include selected patients
            patient_files = {patient: patient_files[patient] for patient in selected_patients}

            logger.info(f"Selected {len(self._file_paths)} files from {max_random_studies} patients")

        self.orientation = Orientation.from_value(orientation)
        self.transform = transform
        self.target_transform = target_transform
        self.mode = Dimensionality.from_value(mode)

        # Crop saving functionality
        self.save_crops = save_crops
        self.crops_output_dir = crops_output_dir
        self.max_crops_per_patient = max_crops_per_patient
        self.max_nrrds_per_patient = max_nrrds_per_patient
        self.max_slices_per_nrrd = max_slices_per_nrrd
        self.patient_crop_counts = {}

        if save_crops and crops_output_dir:
            # Extract database name from path
            if isinstance(data_source, str):
                db_name = Path(data_source).stem  # Gets filename without extension
            else:
                db_name = Path(str(data_source)).stem

            # Create database-specific output directory
            output_path = Path(crops_output_dir) / db_name
            output_path.mkdir(parents=True, exist_ok=True)
            (output_path / "global_crops").mkdir(exist_ok=True)
            (output_path / "local_crops").mkdir(exist_ok=True)
            self.crops_output_dir = str(output_path)  # Update output dir to include database name
            logger.info(f"Debug dataset: Will save crops to {self.crops_output_dir}")
            logger.info(f"Limits: {max_nrrds_per_patient} NRRDs per patient, {max_slices_per_nrrd} slices per NRRD")
            if max_random_studies:
                logger.info(f"Will process {max_random_studies} random patients")

        self.slices = []
        self.slice_metadata = []  # Store filename info for each slice

        if self.mode == Dimensionality.D2:
            # Process files with limits
            processed_patients = set()  # Keep track of which patients we've processed
            for patient_id, files in patient_files.items():
                # Skip if we've already processed enough patients
                if max_random_studies is not None and len(processed_patients) >= max_random_studies:
                    break

                processed_patients.add(patient_id)

                # Process all NRRD files per patient (or limit if specified)
                if self.max_nrrds_per_patient is not None:
                    selected_files = files[: self.max_nrrds_per_patient]
                    logger.info(f"Patient {patient_id}: Processing {len(selected_files)}/{len(files)} NRRD files")
                else:
                    selected_files = files
                    logger.info(f"Patient {patient_id}: Processing all {len(files)} NRRD files")

                for file_path in selected_files:
                    volume = self._load_and_orient_volume(file_path)
                    # Extract only middle slice from each NRRD
                    middle_slice, middle_idx = self._extract_middle_slice_from_volume(volume)
                    if middle_slice is not None:
                        self.slices.append(middle_slice)

                        # Store metadata for the middle slice
                        scanner_type = self._extract_scanner_type(str(file_path))
                        nrrd_name = self._extract_nrrd_name(str(file_path))
                        self.slice_metadata.append(
                            {
                                "filename": str(file_path),
                                "patient_id": patient_id,
                                "scanner_type": scanner_type,
                                "nrrd_name": nrrd_name,
                                "slice_idx": middle_idx,
                            }
                        )
        else:
            raise NotImplementedError("3D mode is not implemented yet")

    def _is_database_url(self, data_source: str) -> bool:
        """Check if data_source is a database URL or directory path."""
        return data_source.endswith(".sqlite") or data_source.endswith(".db") or data_source.startswith("sqlite://")

    def _get_file_paths_from_database(self, database_url: str) -> list[Path]:
        """Get file paths from database."""
        if not SQLALCHEMY_AVAILABLE:
            raise ImportError("sqlalchemy is required for database functionality but is not available")

        try:
            # Handle different URL formats
            if not database_url.startswith("sqlite://"):
                if database_url.startswith("/"):
                    database_url = f"sqlite:///{database_url}"
                else:
                    database_url = f"sqlite:///{database_url}"

            logger.info(f"Connecting to database: {database_url}")
            engine = create_engine(database_url)

            with engine.connect() as connection:
                # Query to get all file paths from the database
                query = text("SELECT filename FROM images")
                result = connection.execute(query)

                file_paths = []
                for row in result:
                    file_path = Path(row[0])
                    if file_path.exists():
                        file_paths.append(file_path)
                    else:
                        logger.warning(f"File not found: {file_path}")

                logger.info(f"Found {len(file_paths)} files from database")
                return file_paths

        except Exception as e:
            logger.error(f"Error reading database {database_url}: {e}")
            raise RuntimeError(f"Failed to read from database: {database_url}. Error: {e}")

    @staticmethod
    def _find_files_with_extension(root_dir: str | Path, extension: str) -> list[Path]:
        """Find all files with given extension in directory (recursive)."""
        root_dir = Path(root_dir)
        files = list(root_dir.glob(f"**/*.{extension}"))
        logger.info(f"Found {len(files)} .{extension} files in {root_dir}")
        return files

    def _load_and_orient_volume(self, file_path: Path):
        # Read the image using ITK
        volume = sitk.ReadImage(str(file_path))
        oriented_volume = sitk.DICOMOrient(volume, self.orientation.value)

        np_array = sitk.GetArrayFromImage(oriented_volume)
        oriented_volume_array = np_array.astype(np.float32)
        return oriented_volume_array

    def _get_target(self, idx: int):
        dummy_target = torch.zeros(1)
        return dummy_target

    def _extract_slices_from_volume(self, volume_array, max_slices: int = 10):
        slices = []
        # to do check if z is third
        size_z = volume_array.shape[0]
        for i in range(size_z):
            slice_2d = volume_array[i, :, :]
            slice_tensor = torch.tensor(slice_2d, dtype=torch.float32)

            slice_tensor = slice_tensor.unsqueeze(0)
            slices.append(slice_tensor)
            if len(slices) >= max_slices:
                break
        return slices

    def _extract_middle_slice_from_volume(self, volume_array):
        """Extract only the middle slice from a volume."""
        size_z = volume_array.shape[0]
        if size_z == 0:
            return None, None

        # Get middle slice index
        middle_idx = size_z // 2

        # Extract middle slice
        slice_2d = volume_array[middle_idx, :, :]
        slice_tensor = torch.tensor(slice_2d, dtype=torch.float32)
        slice_tensor = slice_tensor.unsqueeze(0)

        return slice_tensor, middle_idx

    def _extract_patient_id(self, filename: str) -> str:
        """Extract patient ID from filename."""
        # Pattern for NKI breast data: B000000001, B000003629, etc.
        match = re.search(r"B\d{9}", filename)
        if match:
            return match.group(0)

        # Fallback: use directory name
        path_parts = Path(filename).parts
        for part in path_parts:
            if part.startswith("B") and len(part) == 10:
                return part

        # Last resort: use first part of filename
        return Path(filename).stem.split("-")[0]

    def _extract_scanner_type(self, filename: str) -> str:
        """Extract scanner type from filename."""
        patient_id = self._extract_patient_id(filename)
        siemens_patients = ["B000000001", "B000000044", "B000000053", "B000000073", "B000000110"]
        return "siemens" if patient_id in siemens_patients else "philips"

    def _extract_nrrd_name(self, filename: str) -> str:
        """Extract NRRD name from filename."""
        return Path(filename).stem

    def _save_crops_from_transform_output(self, crops, metadata):
        """Save crops generated by the transform."""
        if not self.save_crops or not self.crops_output_dir:
            return

        patient_id = metadata["patient_id"]
        scanner_type = metadata["scanner_type"]
        nrrd_name = metadata["nrrd_name"]
        slice_idx = metadata["slice_idx"]

        # Check if we've reached the limit for this patient
        if self.patient_crop_counts.get(patient_id, 0) >= self.max_crops_per_patient:
            return

        # Initialize crop count for this patient
        if patient_id not in self.patient_crop_counts:
            self.patient_crop_counts[patient_id] = 0

        # Import the individual transforms used in the batch augmentation
        from fomo.transforms.transforms import (
            PercentileClipper,
            MinMaxNormalizer,
            ColorJitter,
            RandomGaussianBlur,
            RandomGamma,
            RelativeGaussianNoise,
            Normalize,
        )
        import torch.nn as nn

        # Create the same intensity transforms as used in training
        initial_normalization = nn.Sequential(
            PercentileClipper(percentile=99.9),
            MinMaxNormalizer(),
        )

        intensity_augmentations = nn.Sequential(
            RandomGamma(gamma_range=(0.9, 1.1), gain_range=(1.0, 1.0), p=0.7),
            RelativeGaussianNoise(std_factor=(0.01, 0.2), p=0.3),
        )

        color_jittering = ColorJitter(brightness=0.3, contrast=0.3, p=0.8)
        normalize = Normalize(
            mean=(0.1008,), std=(0.193391,), device=torch.device("cuda" if torch.cuda.is_available() else "cpu")
        )

        # Handle DINOv2 transform output (dictionary format)
        if isinstance(crops, dict):
            # Extract and process global crops
            if "global_crops" in crops:
                global_crops = crops["global_crops"]
                if isinstance(global_crops, (list, tuple)):
                    for crop_idx, crop in enumerate(global_crops):
                        if self.patient_crop_counts[patient_id] >= self.max_crops_per_patient:
                            break

                        # Apply intensity augmentations to individual crop
                        crop = crop.unsqueeze(0)  # Add batch dimension
                        crop = initial_normalization(crop)
                        crop = intensity_augmentations(crop)
                        crop = color_jittering(crop)

                        # Apply different blur for global crops
                        if crop_idx == 0:
                            blur = RandomGaussianBlur(kernel_size=(9, 9), sigma=(0.1, 2.0), p=1.0)
                        else:
                            blur = RandomGaussianBlur(kernel_size=(9, 9), sigma=(0.1, 2.0), p=0.1)
                        crop = blur(crop)

                        crop = normalize(crop)
                        crop = crop.squeeze(0)  # Remove batch dimension

                        self._save_single_crop(crop, patient_id, scanner_type, nrrd_name, slice_idx, "global", crop_idx)

            # Extract and process local crops
            if "local_crops" in crops:
                local_crops = crops["local_crops"]
                if isinstance(local_crops, torch.Tensor):
                    # local_crops is usually a stacked tensor of shape (N, C, H, W)
                    for crop_idx in range(local_crops.shape[0]):
                        if self.patient_crop_counts[patient_id] >= self.max_crops_per_patient:
                            break
                        crop = local_crops[crop_idx]

                        # Apply intensity augmentations to individual crop
                        crop = crop.unsqueeze(0)  # Add batch dimension
                        crop = initial_normalization(crop)
                        crop = intensity_augmentations(crop)
                        crop = color_jittering(crop)

                        # Apply local blur
                        blur = RandomGaussianBlur(kernel_size=(9, 9), sigma=(0.1, 1.0), p=0.5)
                        crop = blur(crop)

                        crop = normalize(crop)
                        crop = crop.squeeze(0)  # Remove batch dimension

                        self._save_single_crop(crop, patient_id, scanner_type, nrrd_name, slice_idx, "local", crop_idx)
                elif isinstance(local_crops, (list, tuple)):
                    for crop_idx, crop in enumerate(local_crops):
                        if self.patient_crop_counts[patient_id] >= self.max_crops_per_patient:
                            break

                        # Apply intensity augmentations to individual crop
                        crop = crop.unsqueeze(0)  # Add batch dimension
                        crop = initial_normalization(crop)
                        crop = intensity_augmentations(crop)
                        crop = color_jittering(crop)

                        # Apply local blur
                        blur = RandomGaussianBlur(kernel_size=(9, 9), sigma=(0.1, 1.0), p=0.5)
                        crop = blur(crop)

                        crop = normalize(crop)
                        crop = crop.squeeze(0)  # Remove batch dimension

                        self._save_single_crop(crop, patient_id, scanner_type, nrrd_name, slice_idx, "local", crop_idx)
        else:
            # Handle legacy format (list of tensors or single tensor)
            if isinstance(crops, (list, tuple)):
                crop_list = crops
            else:
                crop_list = [crops]

            for crop_idx, crop in enumerate(crop_list):
                if self.patient_crop_counts[patient_id] >= self.max_crops_per_patient:
                    break

                # Apply intensity augmentations to individual crop
                crop = crop.unsqueeze(0)  # Add batch dimension
                crop = initial_normalization(crop)
                crop = intensity_augmentations(crop)
                crop = color_jittering(crop)

                # Determine crop type and apply appropriate blur
                crop_type = "global" if crop_idx < 2 else "local"
                if crop_type == "global":
                    if crop_idx == 0:
                        blur = RandomGaussianBlur(kernel_size=(9, 9), sigma=(0.1, 2.0), p=1.0)
                    else:
                        blur = RandomGaussianBlur(kernel_size=(9, 9), sigma=(0.1, 2.0), p=0.1)
                else:
                    blur = RandomGaussianBlur(kernel_size=(9, 9), sigma=(0.1, 1.0), p=0.5)
                crop = blur(crop)

                crop = normalize(crop)
                crop = crop.squeeze(0)  # Remove batch dimension

                self._save_single_crop(crop, patient_id, scanner_type, nrrd_name, slice_idx, crop_type, crop_idx)

    def _save_single_crop(self, crop, patient_id, scanner_type, nrrd_name, slice_idx, crop_type, crop_idx):
        """Save a single crop tensor as an image."""
        # Convert tensor to numpy
        if isinstance(crop, torch.Tensor):
            if crop.dim() == 3:  # CHW format
                crop_array = crop.permute(1, 2, 0).numpy()
            else:  # HW format
                crop_array = crop.numpy()
        else:
            crop_array = crop

        # Handle grayscale
        if len(crop_array.shape) == 3 and crop_array.shape[-1] == 1:
            crop_array = crop_array.squeeze(-1)
        elif len(crop_array.shape) == 3 and crop_array.shape[-1] == 3:
            # Convert RGB to grayscale
            crop_array = np.dot(crop_array, [0.299, 0.587, 0.114])

        # Normalize to 0-255
        crop_array = (crop_array - crop_array.min()) / (crop_array.max() - crop_array.min() + 1e-8)
        crop_array = (crop_array * 255).astype(np.uint8)

        # Extract scan type and date from path for aprep111 dataset
        scan_type = ""
        study_date = ""
        if "aprep111" in str(self.data_source):
            # Get the full path from the metadata
            metadata_idx = next(
                (i for i, meta in enumerate(self.slice_metadata) if meta["patient_id"] == patient_id), None
            )
            if metadata_idx is not None:
                full_path = str(self.slice_metadata[metadata_idx]["filename"])

                # Find the scan type and date in the path components
                path_parts = full_path.split("/")
                for part in path_parts:
                    if part in ["dce", "t1", "t2", "adc", "sinwas", "suitwas", "mip"]:
                        scan_type = f"_{part}_"
                    # Look for date format YYYYMMDD
                    if len(part) == 8 and part.isdigit():
                        study_date = f"_{part}_"

        # Create filename with NRRD name, study date, slice info, and scan type for aprep111
        # Format: {patient}_{manufacturer}_{study_date}_{nrrd_name}{scan_type}{slice_idx}_{crop_type}_{crop_idx}.png
        filename = f"{patient_id}_{scanner_type}{study_date}{Path(nrrd_name).stem}{scan_type}{slice_idx}_{crop_type}_{crop_idx}.png"

        output_path = Path(self.crops_output_dir)
        if crop_type == "global":
            save_path = output_path / "global_crops" / filename
        else:
            save_path = output_path / "local_crops" / filename

        # Save image
        Image.fromarray(crop_array, mode="L").save(save_path)
        self.patient_crop_counts[patient_id] += 1

    def __getitem__(self, idx):
        target = self._get_target(idx)  # dummy target
        if self.mode == Dimensionality.D2:
            slice_tensor = self.slices[idx]
            metadata = self.slice_metadata[idx]

            if self.transform:
                transformed = self.transform(slice_tensor)

                # Save crops if enabled
                if self.save_crops:
                    self._save_crops_from_transform_output(transformed, metadata)

                return transformed, target
            else:
                return slice_tensor, target

    def __len__(self):
        return len(self.slices)

    def get_crop_summary(self):
        """Get a summary of saved crops."""
        if not self.save_crops:
            return "Crop saving is disabled"

        total_crops = sum(self.patient_crop_counts.values())
        summary = f"Total crops saved: {total_crops}\n"

        # Group by patient
        patient_counts = {}
        for metadata in self.slice_metadata:
            patient_id = metadata["patient_id"]
            if patient_id not in patient_counts:
                patient_counts[patient_id] = {"scanner_type": metadata["scanner_type"], "crop_count": 0}
            if patient_id in self.patient_crop_counts:
                patient_counts[patient_id]["crop_count"] = self.patient_crop_counts[patient_id]

        summary += "\nCrops per patient:\n"
        for patient_id, info in sorted(patient_counts.items()):
            summary += f"  {patient_id} ({info['scanner_type']}): {info['crop_count']} crops\n"

        return summary


if __name__ == "__main__":
    from torch.utils.data import DataLoader

    root_dir = "/processing/e.marcus/mri_fomo_data"
    dataset = DebugMRIDataset(root_dir, orientation="LPS", mode="2D")
    dataloader = DataLoader(dataset, batch_size=1, shuffle=False)
    for entry in dataloader:
        logger.info(f"Entry shape: {entry.shape}")
