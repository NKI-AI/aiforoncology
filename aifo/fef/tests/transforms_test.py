import torch
from fef.transforms.transforms import PercentileClipper
import pytest
from torchvision import tv_tensors


@pytest.fixture
def sample_2d_image():
    img = torch.arange(0, 16, dtype=torch.float32).reshape(1, 4, 4)
    return tv_tensors.Image(img)


def test_percentile_clipper_clips_above_percentile(sample_2d_image):
    # Arrange: create a PercentileClipper for the 75th percentile
    clipper = PercentileClipper(percentile=75)
    img = sample_2d_image
    # Act: apply the transform
    clipped = clipper(img)
    # Compute the expected upper bound (75th percentile)
    # The input is shape (1, 4, 4), values 0..15
    flat = img.view(1, -1)
    expected_upper = torch.quantile(flat, 0.75, dim=1, keepdim=True)
    # All values above expected_upper should be clipped
    assert torch.all(clipped <= expected_upper)
    # All values below or equal to expected_upper should be unchanged
    below_mask = img <= expected_upper.view(1, 1, 1)
    assert torch.all(clipped[below_mask] == img[below_mask])
