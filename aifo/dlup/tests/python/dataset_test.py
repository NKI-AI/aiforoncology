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
"""Test the datasets facility classes."""

from unittest.mock import patch

import dlup
import numpy as np
import pytest
from dlup.data.dataset import ConcatDataset, Dataset, SlideDataset, TilingMode


@patch.object(dlup.SlideImage, "from_file_path")
@patch.object(SlideDataset, "slide_image")
def test_tiled_level_slide_image_dataset(mock_slide_image, mock_from_file_path, dlup_wsi):
    """Test a single image dataset."""
    mock_from_file_path.return_value = dlup_wsi
    with patch.object(SlideDataset, "slide_image", dlup_wsi):
        dataset = SlideDataset.from_standard_tiling(
            "dummy",
            mpp=1.0,
            tile_size=(32, 24),
            tile_overlap=(0, 0),
            tile_mode=TilingMode.skip,
            mask=None,
        )
        tile_data = dataset[0]
        tile = tile_data.image
        coordinates = tile_data.coordinates

        # Numpy array has height, width, channels.
        # Images have width, height, channels.
        assert np.asarray(tile).shape == (24, 32, 4)
        assert len(coordinates) == 2

        assert len(dataset) == 70

        # Let's grab a few samples instead of one
        tile_data = dataset[0:2]
        assert len(tile_data) == 2
        assert all([np.asarray(tile.image).shape == (24, 32, 4) for tile in tile_data])


class TestConcatDataset:
    #  Concatenate two or more datasets and access elements via integer index
    def test_concatenate_and_access(self):
        class SimpleDataset(Dataset):
            def __init__(self, data):
                self.data = data

            def __len__(self):
                return len(self.data)

            def __getitem__(self, idx):
                return self.data[idx]

        dataset1 = SimpleDataset([1, 2, 3])
        dataset2 = SimpleDataset([4, 5, 6])
        concat_dataset = ConcatDataset([dataset1, dataset2])

        assert concat_dataset[0] == 1
        assert concat_dataset[2] == 3
        assert concat_dataset[3] == 4
        assert concat_dataset[5] == 6
        assert concat_dataset[-6] == 1

        assert concat_dataset[0:3] == [1, 2, 3]
        assert concat_dataset[1:3] == [2, 3]
        assert concat_dataset[-1:-4:-1] == [6, 5, 4]

        with pytest.raises(ValueError) as exc_info:
            concat_dataset.index_to_dataset(-7)
        assert str(exc_info.value) == "Absolute value of index should not exceed dataset length"

    #  Initialize ConcatDataset with an empty list of datasets and catch assertion error
    def test_empty_dataset_initialization(self):
        with pytest.raises(AssertionError) as exc_info:
            _ = ConcatDataset([])
        assert str(exc_info.value) == "datasets should not be an empty iterable"

    def test_non_indexable_dataset(self):
        class SimpleDataset:
            def __init__(self, data):
                self.data = data

            def __len__(self):
                return len(self.data)

        dataset1 = SimpleDataset([1, 2, 3])
        with pytest.raises(ValueError) as exc_info:
            _ = ConcatDataset([dataset1])
        assert str(exc_info.value) == "ConcatDataset requires datasets to be indexable."
