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

from .base_mri_cls_dataset import BaseMRIClassificationDataset
import os


class MRIClassificationDataset(BaseMRIClassificationDataset):
    def filename(self, index: int) -> str:
        sample_index, slice_index = self._indices[index]
        volume_file_path = self._volume_files[sample_index]
        base_filename = os.path.basename(volume_file_path)
        name, ext = os.path.splitext(base_filename)
        filename_with_slice = f"{name}-{slice_index}{ext}"
        return filename_with_slice

    def load_metadata(self, index: int) -> dict:
        sample_index, slice_index = self._indices[index]
        volume_path = self._volume_files[sample_index]
        label_row = self._scan_label_pairs[self._scan_label_pairs["file_path"] == volume_path]
        return {
            "sample_index": int(sample_index),
            "slice_index": int(slice_index),
            "patient_id": label_row["patient_id"].values[0],
        }


class MRIMILClassificationDataset(BaseMRIClassificationDataset):
    def filename(self, index: int) -> str:
        sample_index, _ = self._indices[index]
        volume_file_path = self._volume_files[sample_index]
        return volume_file_path

    def load_metadata(self, index: int) -> dict:
        sample_index, slice_index = self._indices[index]
        volume_path = self._volume_files[sample_index]
        label_row = self._scan_label_pairs[self._scan_label_pairs["file_path"] == volume_path]
        return {
            "multi_id": int(sample_index),
            "slice_index": int(slice_index),
            "patient_id": label_row["patient_id"].values[0],
        }
