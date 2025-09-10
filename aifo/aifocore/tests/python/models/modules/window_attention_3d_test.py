from typing import Sequence

import pytest
import torch

from aifocore.models.modules import WindowAttention3d


@pytest.mark.parametrize("window_size", [(3, 3, 3), (2, 6, 8), (8, 6, 2)])
def test_relative_position_shapes_different_window_sizes(window_size: Sequence[int]):
    # arrange & act
    num_heads = 3
    window_attn = WindowAttention3d(
        dim=64,
        num_heads=num_heads,
        window_size=window_size,
        qkv_bias=False,
        attn_drop=0,
        proj_drop=0,
    )

    # assert
    number_of_biases = ((2 * window_size[0]) - 1) * ((2 * window_size[1]) - 1) * ((2 * window_size[2]) - 1)
    assert window_attn.relative_position_bias_table.shape == (number_of_biases, num_heads)
    assert window_attn.relative_position_index.shape == (
        window_size[0] * window_size[1] * window_size[2],
        window_size[0] * window_size[1] * window_size[2],
    )


@pytest.mark.parametrize("window_size", [(3, 3, 3), (2, 6, 8), (8, 6, 2)])
def test_forward_different_window_sizes(window_size: Sequence[int]):
    # arrange
    num_heads = 4
    dim = 64
    window_attn = WindowAttention3d(
        dim=dim,
        num_heads=num_heads,
        window_size=window_size,
        qkv_bias=False,
        attn_drop=0,
        proj_drop=0,
    )
    x = torch.randn((1, window_size[0] * window_size[1] * window_size[2], dim))

    # act
    attention_values = window_attn.forward(x)

    # assert
    assert attention_values.shape == (1, window_size[0] * window_size[1] * window_size[2], dim)
