from typing import Sequence

import torch
import pytest

from aifocore.models.modules import PatchMerging


@pytest.mark.parametrize(
    ["input_shape", "expected_output_shape"],
    [
        ((4, 16, 16, 3), (4, 8, 8, 6)),
        # padding for uneven input dimensions
        ((4, 15, 15, 3), (4, 8, 8, 6)),
        ((4, 16, 16, 16, 3), (4, 8, 8, 8, 6)),
        # padding for uneven input dimensions
        ((4, 15, 15, 15, 3), (4, 8, 8, 8, 6)),
    ],
)
def test_forward_returns_correct_shape(input_shape: Sequence[int], expected_output_shape: Sequence[int]) -> None:
    # arrange
    input = torch.randn(input_shape)
    merging_module = PatchMerging(input_shape[-1], spatial_dims=len(input_shape[1:-1]))

    # act
    merged_input = merging_module.forward(input)

    # assert
    assert merged_input.shape == expected_output_shape
