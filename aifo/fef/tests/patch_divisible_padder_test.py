import pytest
import torch
from torchvision import tv_tensors
from fef.transforms.cropping import PatchDivisiblePadder


@pytest.fixture
def sample_2d_image():
    img = torch.rand((1, 30, 45))
    return tv_tensors.Image(img)


@pytest.fixture
def sample_2d_mask():
    mask = torch.zeros((1, 30, 45)).int()
    mask[:, 10:20, 10:20] = 1
    return tv_tensors.Mask(mask)


@pytest.fixture
def sample_3d_image():
    img = torch.rand((1, 7, 30, 45))
    return tv_tensors.Image(img)


@pytest.fixture
def sample_3d_mask():
    mask = torch.zeros((1, 7, 30, 45)).int()
    mask[:, 2:5, 10:20, 10:20] = 1
    return tv_tensors.Mask(mask)


def test_patch_divisible_padder_2d(sample_2d_image, sample_2d_mask):
    padder = PatchDivisiblePadder(patch_size=16)
    padded_img, padded_mask = padder(sample_2d_image, sample_2d_mask)
    # Output shape should be divisible by patch_size
    assert padded_img.shape[1] % 16 == 0
    assert padded_img.shape[2] % 16 == 0
    assert padded_mask.shape[1] % 16 == 0
    assert padded_mask.shape[2] % 16 == 0
    # The original image should be in the bottom, while padding is split equally to the left and right
    height = sample_2d_image.shape[1]
    width = sample_2d_image.shape[2]
    assert torch.allclose(
        padded_img[:, -height:, padded_img.shape[2] // 2 - width + width // 2 : padded_img.shape[2] // 2 + width // 2],
        sample_2d_image,
    )


def test_patch_divisible_padder_3d(sample_3d_image, sample_3d_mask):
    padder = PatchDivisiblePadder(patch_size=8)
    padded_img, padded_mask = padder(sample_3d_image, sample_3d_mask)
    # Output shape should be divisible by patch_size
    assert padded_img.shape[1] % 8 == 0
    assert padded_img.shape[2] % 8 == 0
    assert padded_img.shape[3] % 8 == 0
    assert padded_mask.shape[1] % 8 == 0
    assert padded_mask.shape[2] % 8 == 0
    assert padded_mask.shape[3] % 8 == 0

    # The original volume should be in top, back and padding is distributed to the left and right
    depth = sample_3d_image.shape[1]
    height = sample_3d_image.shape[2]
    width = sample_3d_image.shape[3]
    assert torch.allclose(
        padded_img[
            :, :depth, -height:, padded_img.shape[3] // 2 - width + width // 2 : padded_img.shape[3] // 2 + width // 2
        ],
        sample_3d_image,
    )
