import math
from typing import Sequence

import pytest
import torch

from aifocore.models.blocks.swin_transformer_block import (
    SwinTransformerBlock,
    _pad_to_window_size,
    _get_window_size,
    _window_partition,
    _window_reverse,
)
from aifocore.models.modules import WindowAttention2d, WindowAttention3d, WindowAttentionBase


@pytest.mark.parametrize(
    ["spatial_dimensions", "window_size", "shift_size", "resulting_window_shape"],
    [
        ((16, 16), (7, 7), (0, 0), (7, 7)),
        ((4, 16), (7, 7), (0, 0), (4, 7)),
        ((16, 16, 16), (7, 7, 7), (0, 0, 0), (7, 7, 7)),
        ((16, 16, 4), (7, 7, 7), (0, 0, 0), (7, 7, 4)),
    ],
)
def test_get_window_size_without_shift(
    spatial_dimensions: Sequence[int],
    window_size: Sequence[int],
    shift_size: Sequence[int],
    resulting_window_shape: Sequence[int],
) -> None:
    # arrange & act
    window_size, _ = _get_window_size(spatial_dimensions, window_size, shift_size=shift_size)

    # assert
    assert resulting_window_shape == window_size


@pytest.mark.parametrize(
    ["spatial_dimensions", "window_size", "shift_size", "resulting_window_shape"],
    [
        ((4, 16), (7, 7), (3, 0), (4, 7)),
        ((16, 16, 4), (7, 7, 7), (0, 0, 3), (7, 7, 4)),
    ],
)
def test_get_window_size_with_shift(
    spatial_dimensions: Sequence[int],
    window_size: Sequence[int],
    shift_size: Sequence[int],
    resulting_window_shape: Sequence[int],
) -> None:
    # arrange & act
    window_size, shift_size = _get_window_size(spatial_dimensions, window_size, shift_size=shift_size)

    # assert
    assert resulting_window_shape == window_size
    assert all([x == 0 for x in shift_size])


@pytest.mark.parametrize(
    ["input_shape", "window_size", "resulting_input_shape"],
    [
        ((4, 64, 128, 16), (7, 7), (4, 70, 133, 16)),
        ((4, 32, 64, 128, 16), (7, 7, 7), (4, 35, 70, 133, 16)),
        ((4, 16, 32, 64, 128, 16), (7, 7, 7, 7), (4, 21, 35, 70, 133, 16)),
    ],
)
def test_pad_to_window_size_different_input_sizes(
    input_shape: Sequence[int], window_size: Sequence[int], resulting_input_shape: Sequence[int]
) -> None:
    # arrange
    input = torch.randn(input_shape)

    # act
    padded_input = _pad_to_window_size(input, window_size)

    # assert
    assert padded_input.shape == resulting_input_shape


@pytest.mark.parametrize(
    ["input", "window_size", "expected_output"],
    [
        ((4, 64, 128, 3), (8, 8), (512, 64, 3)),
        ((4, 32, 64, 128, 3), (8, 8, 8), (2048, 512, 3)),
        ((1, 32, 32, 3), (32, 32), (1, 1024, 3)),
        ((1, 32, 32, 3), (1, 1), (1024, 1, 3)),
    ],
)
def test_window_partition_different_input_sizes(
    input: Sequence[int], window_size: Sequence[int], expected_output: Sequence[int]
) -> None:
    # arrange
    input = torch.randn(input)

    # act
    windows = _window_partition(input, window_size)

    # assert
    assert windows.shape == expected_output


@pytest.mark.parametrize(
    ["input_shape", "window_size", "expected_windows"],
    [
        (
            (1, 4, 4, 1),
            (2, 2),
            ([[[0], [1], [4], [5]], [[2], [3], [6], [7]], [[8], [9], [12], [13]], [[10], [11], [14], [15]]]),
        ),
        (
            (1, 4, 4, 4, 1),
            (2, 2, 2),
            (
                [
                    [[0], [1], [4], [5], [16], [17], [20], [21]],
                    [[2], [3], [6], [7], [18], [19], [22], [23]],
                    [[8], [9], [12], [13], [24], [25], [28], [29]],
                    [[10], [11], [14], [15], [26], [27], [30], [31]],
                    [[32], [33], [36], [37], [48], [49], [52], [53]],
                    [[34], [35], [38], [39], [50], [51], [54], [55]],
                    [[40], [41], [44], [45], [56], [57], [60], [61]],
                    [[42], [43], [46], [47], [58], [59], [62], [63]],
                ]
            ),
        ),
    ],
)
def test_window_partition_creates_correct_windows(
    input_shape: Sequence[int], window_size: Sequence[int], expected_windows: Sequence[int]
) -> None:
    # arrange
    input = torch.arange(0, math.prod(input_shape)).view(input_shape)
    expected_windows_tensor = torch.tensor(expected_windows)

    # act
    windows = _window_partition(input, window_size)

    # assert
    assert torch.equal(windows, expected_windows_tensor)


@pytest.mark.parametrize(
    ["windows", "output_dims"],
    [
        ((32, 8, 8, 3), (8, 16, 16, 3)),
        ((32, 8, 8, 8, 3), (4, 16, 16, 16, 3)),
    ],
)
def test_window_reverse_different_window_sizes(windows: Sequence[int], output_dims: Sequence[int]) -> None:
    # arrange
    windows: torch.Tensor = torch.randn(windows)
    window_size = windows.shape[1:-1]
    output_dims = torch.Size(output_dims)

    # act
    reversal = _window_reverse(windows, window_size, output_dims)

    # assert
    assert reversal.shape == output_dims


@pytest.mark.parametrize(
    ["input_shape", "window_size"],
    [
        ((4, 64, 128, 3), (8, 8)),
        ((4, 32, 64, 128, 3), (8, 8, 8)),
        ((1, 32, 32, 3), (32, 32)),
        ((1, 32, 32, 3), (1, 1)),
    ],
)
def test_window_partition_and_reversal_e2e(
    input_shape: Sequence[int],
    window_size: Sequence[int],
) -> None:
    # arrange
    input = torch.randint(low=-1000, high=1000, size=input_shape)

    # act
    windows = _window_partition(input, window_size)
    reversal = _window_reverse(windows, window_size, input_shape)

    # assert
    assert torch.equal(input, reversal)


@pytest.mark.parametrize(
    ["input_shape", "window_size", "shift_size", "window_attention"],
    [
        # ------ 2D test cases ------
        # standard 2d case
        ((4, 64, 128, 24), (8, 8), (0, 0), WindowAttention2d),
        # window fills entire spatial dims
        ((1, 32, 32, 24), (32, 32), (0, 0), WindowAttention2d),
        # 1 by 1 windows
        ((1, 32, 32, 24), (1, 1), (0, 0), WindowAttention2d),
        # padding
        ((1, 32, 32, 24), (7, 7), (0, 0), WindowAttention2d),
        # shift
        ((1, 32, 32, 24), (8, 8), (4, 4), WindowAttention2d),
        # padding + shift
        ((1, 32, 32, 24), (8, 8), (4, 4), WindowAttention2d),
        # ------ 3D test cases ------
        # standard 3d case
        ((4, 32, 64, 128, 24), (8, 8, 8), (0, 0, 0), WindowAttention3d),
        # 1 by 1 by 1 windows
        ((4, 32, 32, 32, 24), (1, 1, 1), (0, 0, 0), WindowAttention3d),
        # window fills entire spatial dims
        ((4, 8, 8, 8, 24), (8, 8, 8), (0, 0, 0), WindowAttention3d),
        # padding
        ((4, 32, 32, 32, 24), (7, 7, 7), (0, 0, 0), WindowAttention3d),
        # shift
        ((4, 32, 32, 32, 24), (8, 8, 8), (4, 4, 4), WindowAttention3d),
        # padding + shift
        ((4, 32, 32, 32, 24), (7, 7, 7), (4, 4, 4), WindowAttention3d),
    ],
)
def test_sliding_window_attention_forward_correct_shapes(
    input_shape: Sequence[int],
    window_size: Sequence[int],
    shift_size: Sequence[int],
    window_attention: type[WindowAttentionBase],
) -> None:
    # arrange
    swin_block = SwinTransformerBlock(
        dim=24, window_attention=window_attention, num_heads=3, window_size=window_size, shift_size=shift_size
    )
    input = torch.randn(input_shape)

    # act
    fwd_output = swin_block.sliding_window_attn_forward(input, mask_matrix=None)

    # assert
    assert fwd_output.shape == input.shape
