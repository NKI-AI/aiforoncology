import pytest
import numpy as np
import SimpleITK as sitk
from unittest.mock import MagicMock
from fomo.ct.dataset.ct_data_producer import CTDataProducer
from fomo.database_models import Image


@pytest.fixture
def mock_image():
    # Create a mock image with a specific size
    image_size = (150, 256, 256)  # (depth, height, width)
    image_array = np.random.rand(*image_size).astype(np.float32)
    image = sitk.GetImageFromArray(image_array)
    return image


@pytest.fixture
def mock_db_entry():
    # Create a mock database entry
    return Image(filename="mock_image_path")


@pytest.fixture
def mock_ct_data_producer() -> CTDataProducer:
    # Create a mock CTDataProducer
    return CTDataProducer(
        name="test", chunk_size_bytes=1024, max_memory_size_bytes=1024 * 1024, database_url="mock_db_url", num_workers=1
    )


def test_load_and_preprocess_image_whole_array(mock_ct_data_producer, mock_db_entry, mock_image):
    # Arrange
    sitk.ReadImage = MagicMock(return_value=mock_image)
    mock_ct_data_producer._reorient_image_to_lps = MagicMock(return_value=mock_image)
    # Select the whole array by setting slices_per_chunk to 0
    mock_ct_data_producer._slices_per_chunk = 0

    # Act
    chunks = mock_ct_data_producer._load_and_preprocess_image(mock_db_entry)

    # Assert
    assert len(chunks) == 1
    assert chunks[0].shape == (150, 256, 256)


def test_load_and_preprocess_image_single_slices(mock_ct_data_producer, mock_db_entry, mock_image):
    # Arrange
    sitk.ReadImage = MagicMock(return_value=mock_image)
    mock_ct_data_producer._reorient_image_to_lps = MagicMock(return_value=mock_image)
    mock_ct_data_producer._slices_per_chunk = 1

    # Act
    chunks = mock_ct_data_producer._load_and_preprocess_image(mock_db_entry)

    # Assert
    assert len(chunks) == 150
    for chunk in chunks:
        assert chunk.shape == (1, 256, 256)


def test_load_and_preprocess_image_chunks_of_64(mock_ct_data_producer, mock_db_entry, mock_image):
    # Arrange
    sitk.ReadImage = MagicMock(return_value=mock_image)
    mock_ct_data_producer._reorient_image_to_lps = MagicMock(return_value=mock_image)
    # This will throw away some slices because depth is not divisible by slices_per_chunk
    mock_ct_data_producer._slices_per_chunk = 64

    # Act
    chunks = mock_ct_data_producer._load_and_preprocess_image(mock_db_entry)

    # Assert
    assert len(chunks) == 2
    for chunk in chunks:
        assert chunk.shape == (64, 256, 256)
