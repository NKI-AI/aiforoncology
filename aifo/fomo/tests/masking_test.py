import torch
import pytest
from fomo.utils.masking.generator_2d import MaskingGenerator2d
from fomo.utils.masking.generator_3d import MaskingGenerator3d


class TestMaskingGenerator2d:
    @pytest.mark.parametrize(
        "input_size,min_patches,max_patches",
        [
            (8, 3, 10),
            (10, 5, 20),
            (16, 8, 30),
        ],
    )
    def test_mask_applies_masking(self, input_size, min_patches, max_patches):
        # Arrange
        mask = torch.zeros((input_size, input_size), dtype=torch.bool)
        generator = MaskingGenerator2d(
            input_size=input_size,
            min_num_patches=min_patches,
            max_num_patches=max_patches,
        )

        # Act
        delta = generator._mask(mask, max_mask_patches=max_patches)

        # Assert
        assert delta > 0
        assert mask.sum().item() == delta
        assert min_patches <= delta <= max_patches

    @pytest.mark.parametrize(
        "input_size,min_patches,max_patches,test_max",
        [
            (8, 3, 15, 7),
            (10, 5, 20, 10),
            (16, 8, 30, 15),
        ],
    )
    def test_mask_respects_max_patches(self, input_size, min_patches, max_patches, test_max):
        # Arrange
        mask = torch.zeros((input_size, input_size), dtype=torch.bool)
        generator = MaskingGenerator2d(
            input_size=input_size,
            min_num_patches=min_patches,
            max_num_patches=max_patches,
        )

        # Act
        delta = generator._mask(mask, max_mask_patches=test_max)

        # Assert
        assert delta <= test_max

    @pytest.mark.parametrize(
        "input_size,min_patches,max_patches",
        [
            (8, 3, 10),
            (10, 5, 20),
            (16, 8, 30),
        ],
    )
    def test_mask_handles_overlap(self, input_size, min_patches, max_patches):
        # Arrange
        mask = torch.zeros((input_size, input_size), dtype=torch.bool)
        # Pre-mask some positions (scale based on input_size)
        mask_start = input_size // 4
        mask_end = input_size // 2
        mask[mask_start:mask_end, mask_start:mask_end] = True
        initial_masked = mask.sum().item()
        generator = MaskingGenerator2d(
            input_size=input_size,
            min_num_patches=min_patches,
            max_num_patches=max_patches,
        )

        # Act
        delta = generator._mask(mask, max_mask_patches=max_patches)

        # Assert
        assert mask.sum().item() == initial_masked + delta


class TestMaskingGenerator3d:
    @pytest.mark.parametrize(
        "input_size,min_patches,max_patches",
        [
            (8, 3, 10),
            (10, 5, 20),
            (16, 8, 30),
        ],
    )
    def test_mask_applies_masking(self, input_size, min_patches, max_patches):
        # Arrange
        mask = torch.zeros((input_size, input_size, input_size), dtype=torch.bool)
        generator = MaskingGenerator3d(
            input_size=input_size,
            min_num_patches=min_patches,
            max_num_patches=max_patches,
        )

        # Act
        delta = generator._mask(mask, max_mask_patches=max_patches)

        # Assert
        assert delta > 0
        assert mask.sum().item() == delta
        assert min_patches <= delta <= max_patches

    @pytest.mark.parametrize(
        "input_size,min_patches,max_patches,test_max",
        [
            (8, 3, 15, 7),
            (10, 5, 20, 10),
            (16, 8, 30, 15),
        ],
    )
    def test_mask_respects_max_patches(self, input_size, min_patches, max_patches, test_max):
        # Arrange
        mask = torch.zeros((input_size, input_size, input_size), dtype=torch.bool)
        generator = MaskingGenerator3d(
            input_size=input_size,
            min_num_patches=min_patches,
            max_num_patches=max_patches,
        )

        # Act
        delta = generator._mask(mask, max_mask_patches=test_max)

        # Assert
        assert delta <= test_max

    @pytest.mark.parametrize(
        "input_size,min_patches,max_patches",
        [
            (8, 3, 10),
            (10, 5, 20),
            (16, 8, 30),
        ],
    )
    def test_mask_handles_overlap(self, input_size, min_patches, max_patches):
        # Arrange
        mask = torch.zeros((input_size, input_size, input_size), dtype=torch.bool)
        # Pre-mask some positions (scale based on input_size)
        mask_start = input_size // 4
        mask_end = input_size // 2
        mask[mask_start:mask_end, mask_start:mask_end, mask_start:mask_end] = True
        initial_masked = mask.sum().item()
        generator = MaskingGenerator3d(
            input_size=input_size,
            min_num_patches=min_patches,
            max_num_patches=max_patches,
        )

        # Act
        delta = generator._mask(mask, max_mask_patches=max_patches)

        # Assert
        assert mask.sum().item() == initial_masked + delta

    @pytest.mark.parametrize(
        "input_size,min_patches,max_patches",
        [
            (8, 3, 10),
            (10, 5, 20),
            (16, 8, 30),
        ],
    )
    def test_mask_with_custom_aspect_ratios(self, input_size, min_patches, max_patches):
        # Arrange
        mask = torch.zeros((input_size, input_size, input_size), dtype=torch.bool)
        generator = MaskingGenerator3d(
            input_size=input_size,
            min_num_patches=min_patches,
            max_num_patches=max_patches,
            min_aspect_height_width=0.5,
            max_aspect_height_width=2.0,
            min_aspect_height_depth=0.5,
            max_aspect_height_depth=2.0,
        )

        # Act
        delta = generator._mask(mask, max_mask_patches=max_patches)

        # Assert
        assert delta > 0
        assert mask.sum().item() == delta
