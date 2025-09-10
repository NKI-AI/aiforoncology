import pytest
import numpy as np
from fef.transforms.cropping import VolumeBoundingBoxCropper


@pytest.fixture
def sample_3d_volume():
    # Create a 3D volume with a white cube in the center
    vol = np.zeros((30, 100, 100), dtype=np.float32)
    vol[10:20, 25:75, 25:75] = 1.0
    return vol


@pytest.fixture
def sample_3d_segmentation():
    # Create a 3D segmentation mask with a white cube in the center
    seg = np.zeros((30, 100, 100), dtype=np.int32)
    seg[10:20, 25:75, 25:75] = 1
    return seg


def test_volume_bounding_box_cropper_basic(sample_3d_volume):
    cropper = VolumeBoundingBoxCropper(lower_percentile=1.0)
    cropped_vol, _ = cropper(sample_3d_volume)
    # Should crop to the cube region
    assert cropped_vol.shape == (30, 50, 50)
    assert np.allclose(cropped_vol[:, :, :], sample_3d_volume[:, 25:75, 25:75])


def test_volume_bounding_box_cropper_with_segmentation(sample_3d_volume, sample_3d_segmentation):
    cropper = VolumeBoundingBoxCropper(lower_percentile=1.0)
    cropped_vol, cropped_seg = cropper(sample_3d_volume, sample_3d_segmentation)
    assert cropped_vol.shape == (30, 50, 50)
    assert cropped_seg.shape == (30, 50, 50)
    assert np.allclose(cropped_vol, sample_3d_volume[:, 25:75, 25:75])
    assert np.allclose(cropped_seg, sample_3d_segmentation[:, 25:75, 25:75])


def test_volume_bounding_box_cropper_high_percentile(sample_3d_volume):
    cropper = VolumeBoundingBoxCropper(lower_percentile=90.0)
    cropped_vol, _ = cropper(sample_3d_volume)
    # At high percentile, the threshold is 1.0, so only the cube is above threshold
    assert cropped_vol.shape == (30, 50, 50)
    assert np.allclose(cropped_vol, sample_3d_volume[:, 25:75, 25:75])


def test_volume_bounding_box_cropper_all_background():
    vol = np.zeros((30, 100, 100), dtype=np.float32)
    cropper = VolumeBoundingBoxCropper(lower_percentile=1.0)
    cropped_vol, _ = cropper(vol)
    # Should return the original volume if all background
    assert cropped_vol.shape == vol.shape
    assert np.allclose(cropped_vol, vol)


def test_volume_bounding_box_cropper_partial_cube():
    vol = np.zeros((30, 100, 100), dtype=np.float32)
    vol[5:15, 10:60, 40:90] = 2.0
    cropper = VolumeBoundingBoxCropper(lower_percentile=10.0)
    cropped_vol, _ = cropper(vol)
    # Should crop to the region containing the partial cube
    assert cropped_vol.shape == (30, 50, 50)
    assert np.allclose(cropped_vol, vol[:, 10:60, 40:90])
