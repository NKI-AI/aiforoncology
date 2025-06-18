from pathlib import Path
from typing import Callable

import numpy as np
import SimpleITK as sitk
import torch
from fomo.utils.types import Dimensionality, Orientation
from torch.utils.data import Dataset


class DebugMRIDataset(Dataset):
    def __init__(
        self,
        root_dir: str,
        orientation: str,
        mode: str,
        transform: None | Callable = None,
        target_transform: None | Callable = None,
    ):
        self.root_dir = root_dir
        self._file_paths = self._find_files_with_extension(root_dir, "nrrd")

        self.orientation = Orientation.from_value(orientation)
        self.transform = transform
        self.target_transform = target_transform
        self.mode = Dimensionality.from_value(mode)

        self.slices = []

        if self.mode == Dimensionality.D2:
            for file_path in self._file_paths:
                volume = self._load_and_orient_volume(file_path)
                slices = self._extract_slices_from_volume(volume)
                self.slices.extend(slices)
        else:
            raise NotImplementedError("3D mode is not implemented yet")

    @staticmethod
    def _find_files_with_extension(root_dir: str | Path, extension: str) -> list[Path]:
        root_dir = Path(root_dir)
        files = list(root_dir.glob(f"**/*.{extension}"))
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

    def _extract_slices_from_volume(self, volume_array):
        slices = []
        # to do check if z is third
        size_z = volume_array.shape[0]
        for i in range(size_z):
            slice_2d = volume_array[i, :, :]
            slice_tensor = torch.tensor(slice_2d, dtype=torch.float32)

            slice_tensor = slice_tensor.unsqueeze(0)
            slices.append(slice_tensor)
        return slices

    def __getitem__(self, idx):
        target = self._get_target(idx)  # dummy target
        if self.mode == Dimensionality.D2:
            slice_tensor = self.slices[idx]

            if self.transform:
                slice_tensor = self.transform(slice_tensor)

            return slice_tensor, target

    def __len__(self):
        return len(self.slices)


if __name__ == "__main__":
    from torch.utils.data import DataLoader

    root_dir = "/processing/e.marcus/mri_fomo_data"
    dataset = DebugMRIDataset(root_dir, orientation="LPS", mode="2D")
    dataloader = DataLoader(dataset, batch_size=1, shuffle=False)
    for entry in dataloader:
        print(entry.shape)
