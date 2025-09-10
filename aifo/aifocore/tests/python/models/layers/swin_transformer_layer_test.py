from typing import Sequence

import pytest
import torch

from aifocore.models.modules import PatchMerging, WindowAttention2d, WindowAttention3d, WindowAttentionBase
from aifocore.models.layers import SwinTransformerStage
from aifocore.models.layers.swin_transformer_layer import _compute_mask


@pytest.mark.parametrize(
    ["dimensions", "window_size", "shift_size", "expected_shape"],
    [((14, 14), (7, 7), (3, 3), (4, 49, 49)), ((14, 14, 14), (7, 7, 7), (3, 3, 3), (8, 343, 343))],
)
def test_compute_mask_outputs_correct_mask_shapes(
    dimensions: Sequence[int], window_size: Sequence[int], shift_size: Sequence[int], expected_shape: Sequence[int]
):
    # arrange & act
    mask = _compute_mask(dimensions, window_size, shift_size, torch.device("cpu"))

    # assert
    assert mask.shape == expected_shape


@pytest.mark.parametrize(
    ["dimensions", "window_size", "shift_size", "expected_mask"],
    [
        (
            (4, 4),
            (2, 2),
            (1, 1),
            torch.Tensor(
                [
                    # first window
                    [
                        # attention not masked for any value
                        [0, 0, 0, 0],
                        [0, 0, 0, 0],
                        [0, 0, 0, 0],
                        [0, 0, 0, 0],
                    ],
                    # second window
                    [
                        # index 0 pays attention to 2, 1 to 3, all to self.
                        # segments are two 2 by 1 vertical bars in this window
                        [0, -torch.inf, 0, -torch.inf],
                        [-torch.inf, 0, -torch.inf, 0],
                        [0, -torch.inf, 0, -torch.inf],
                        [-torch.inf, 0, -torch.inf, 0],
                    ],
                    # third window
                    [
                        # index 0 pays attention to 1, 2 to 3, all to self.
                        # segments are two 2 by 1 horizontal bars in this window
                        [0, 0, -torch.inf, -torch.inf],
                        [0, 0, -torch.inf, -torch.inf],
                        [-torch.inf, -torch.inf, 0, 0],
                        [-torch.inf, -torch.inf, 0, 0],
                    ],
                    # fourth window
                    [
                        # consists of four single partitions, only pay attention to self.
                        [0, -torch.inf, -torch.inf, -torch.inf],
                        [-torch.inf, 0, -torch.inf, -torch.inf],
                        [-torch.inf, -torch.inf, 0, -torch.inf],
                        [-torch.inf, -torch.inf, -torch.inf, 0],
                    ],
                ]
            ),
        ),
        (
            (4, 4, 4),
            (2, 2, 2),
            (1, 1, 1),
            torch.tensor(
                [
                    # first window - no shifting, all elements can attend to each other
                    # This is a complete 2x2x2 cube, no masking needed
                    [
                        [0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0],
                        [0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0],
                        [0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0],
                        [0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0],
                        [0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0],
                        [0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0],
                        [0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0],
                        [0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0],
                    ],
                    # second window - alternating pattern along x-axis after shift
                    # Elements alternate between two x-segments: indices 0,2,4,6 vs 1,3,5,7
                    # Only elements within the same x-segment can attend to each other
                    [
                        [0.0, -torch.inf, 0.0, -torch.inf, 0.0, -torch.inf, 0.0, -torch.inf],
                        [-torch.inf, 0.0, -torch.inf, 0.0, -torch.inf, 0.0, -torch.inf, 0.0],
                        [0.0, -torch.inf, 0.0, -torch.inf, 0.0, -torch.inf, 0.0, -torch.inf],
                        [-torch.inf, 0.0, -torch.inf, 0.0, -torch.inf, 0.0, -torch.inf, 0.0],
                        [0.0, -torch.inf, 0.0, -torch.inf, 0.0, -torch.inf, 0.0, -torch.inf],
                        [-torch.inf, 0.0, -torch.inf, 0.0, -torch.inf, 0.0, -torch.inf, 0.0],
                        [0.0, -torch.inf, 0.0, -torch.inf, 0.0, -torch.inf, 0.0, -torch.inf],
                        [-torch.inf, 0.0, -torch.inf, 0.0, -torch.inf, 0.0, -torch.inf, 0.0],
                    ],
                    # third window - block pattern along y-axis after shift
                    # Elements form two y-segments: indices 0,1,4,5 vs 2,3,6,7
                    # Only elements within the same y-segment can attend to each other
                    [
                        [0.0, 0.0, -torch.inf, -torch.inf, 0.0, 0.0, -torch.inf, -torch.inf],
                        [0.0, 0.0, -torch.inf, -torch.inf, 0.0, 0.0, -torch.inf, -torch.inf],
                        [-torch.inf, -torch.inf, 0.0, 0.0, -torch.inf, -torch.inf, 0.0, 0.0],
                        [-torch.inf, -torch.inf, 0.0, 0.0, -torch.inf, -torch.inf, 0.0, 0.0],
                        [0.0, 0.0, -torch.inf, -torch.inf, 0.0, 0.0, -torch.inf, -torch.inf],
                        [0.0, 0.0, -torch.inf, -torch.inf, 0.0, 0.0, -torch.inf, -torch.inf],
                        [-torch.inf, -torch.inf, 0.0, 0.0, -torch.inf, -torch.inf, 0.0, 0.0],
                        [-torch.inf, -torch.inf, 0.0, 0.0, -torch.inf, -torch.inf, 0.0, 0.0],
                    ],
                    # fourth window - mixed x-y shift creates 4 isolated groups
                    # Groups: {0,4}, {1,5}, {2,6}, {3,7} - each element can only attend to its partner
                    # This creates a sparse attention pattern with only self and one other
                    [
                        [0.0, -torch.inf, -torch.inf, -torch.inf, 0.0, -torch.inf, -torch.inf, -torch.inf],
                        [-torch.inf, 0.0, -torch.inf, -torch.inf, -torch.inf, 0.0, -torch.inf, -torch.inf],
                        [-torch.inf, -torch.inf, 0.0, -torch.inf, -torch.inf, -torch.inf, 0.0, -torch.inf],
                        [-torch.inf, -torch.inf, -torch.inf, 0.0, -torch.inf, -torch.inf, -torch.inf, 0.0],
                        [0.0, -torch.inf, -torch.inf, -torch.inf, 0.0, -torch.inf, -torch.inf, -torch.inf],
                        [-torch.inf, 0.0, -torch.inf, -torch.inf, -torch.inf, 0.0, -torch.inf, -torch.inf],
                        [-torch.inf, -torch.inf, 0.0, -torch.inf, -torch.inf, -torch.inf, 0.0, -torch.inf],
                        [-torch.inf, -torch.inf, -torch.inf, 0.0, -torch.inf, -torch.inf, -torch.inf, 0.0],
                    ],
                    # fifth window - z-axis separation after shift
                    # Elements form two z-layers: front layer {0,1,2,3} vs back layer {4,5,6,7}
                    # Only elements within the same z-layer can attend to each other
                    [
                        [0.0, 0.0, 0.0, 0.0, -torch.inf, -torch.inf, -torch.inf, -torch.inf],
                        [0.0, 0.0, 0.0, 0.0, -torch.inf, -torch.inf, -torch.inf, -torch.inf],
                        [0.0, 0.0, 0.0, 0.0, -torch.inf, -torch.inf, -torch.inf, -torch.inf],
                        [0.0, 0.0, 0.0, 0.0, -torch.inf, -torch.inf, -torch.inf, -torch.inf],
                        [-torch.inf, -torch.inf, -torch.inf, -torch.inf, 0.0, 0.0, 0.0, 0.0],
                        [-torch.inf, -torch.inf, -torch.inf, -torch.inf, 0.0, 0.0, 0.0, 0.0],
                        [-torch.inf, -torch.inf, -torch.inf, -torch.inf, 0.0, 0.0, 0.0, 0.0],
                        [-torch.inf, -torch.inf, -torch.inf, -torch.inf, 0.0, 0.0, 0.0, 0.0],
                    ],
                    # sixth window - x-z shift creates alternating pattern across z-layers
                    # Front layer has x-alternation: {0,2} vs {1,3}, back layer: {4,6} vs {5,7}
                    # No cross-layer attention allowed
                    [
                        [0.0, -torch.inf, 0.0, -torch.inf, -torch.inf, -torch.inf, -torch.inf, -torch.inf],
                        [-torch.inf, 0.0, -torch.inf, 0.0, -torch.inf, -torch.inf, -torch.inf, -torch.inf],
                        [0.0, -torch.inf, 0.0, -torch.inf, -torch.inf, -torch.inf, -torch.inf, -torch.inf],
                        [-torch.inf, 0.0, -torch.inf, 0.0, -torch.inf, -torch.inf, -torch.inf, -torch.inf],
                        [-torch.inf, -torch.inf, -torch.inf, -torch.inf, 0.0, -torch.inf, 0.0, -torch.inf],
                        [-torch.inf, -torch.inf, -torch.inf, -torch.inf, -torch.inf, 0.0, -torch.inf, 0.0],
                        [-torch.inf, -torch.inf, -torch.inf, -torch.inf, 0.0, -torch.inf, 0.0, -torch.inf],
                        [-torch.inf, -torch.inf, -torch.inf, -torch.inf, -torch.inf, 0.0, -torch.inf, 0.0],
                    ],
                    # seventh window - y-z shift creates block pattern across z-layers
                    # Front layer has y-blocks: {0,1} vs {2,3}, back layer: {4,5} vs {6,7}
                    # No cross-layer attention, only within-block attention per layer
                    [
                        [0.0, 0.0, -torch.inf, -torch.inf, -torch.inf, -torch.inf, -torch.inf, -torch.inf],
                        [0.0, 0.0, -torch.inf, -torch.inf, -torch.inf, -torch.inf, -torch.inf, -torch.inf],
                        [-torch.inf, -torch.inf, 0.0, 0.0, -torch.inf, -torch.inf, -torch.inf, -torch.inf],
                        [-torch.inf, -torch.inf, 0.0, 0.0, -torch.inf, -torch.inf, -torch.inf, -torch.inf],
                        [-torch.inf, -torch.inf, -torch.inf, -torch.inf, 0.0, 0.0, -torch.inf, -torch.inf],
                        [-torch.inf, -torch.inf, -torch.inf, -torch.inf, 0.0, 0.0, -torch.inf, -torch.inf],
                        [-torch.inf, -torch.inf, -torch.inf, -torch.inf, -torch.inf, -torch.inf, 0.0, 0.0],
                        [-torch.inf, -torch.inf, -torch.inf, -torch.inf, -torch.inf, -torch.inf, 0.0, 0.0],
                    ],
                    # eighth window - triple shift (x,y,z) creates 8 isolated elements
                    # Each of the 8 positions becomes its own partition, only self-attention allowed
                    # This is the most restrictive masking pattern with minimal connectivity
                    [
                        [0.0, -torch.inf, -torch.inf, -torch.inf, -torch.inf, -torch.inf, -torch.inf, -torch.inf],
                        [-torch.inf, 0.0, -torch.inf, -torch.inf, -torch.inf, -torch.inf, -torch.inf, -torch.inf],
                        [-torch.inf, -torch.inf, 0.0, -torch.inf, -torch.inf, -torch.inf, -torch.inf, -torch.inf],
                        [-torch.inf, -torch.inf, -torch.inf, 0.0, -torch.inf, -torch.inf, -torch.inf, -torch.inf],
                        [-torch.inf, -torch.inf, -torch.inf, -torch.inf, 0.0, -torch.inf, -torch.inf, -torch.inf],
                        [-torch.inf, -torch.inf, -torch.inf, -torch.inf, -torch.inf, 0.0, -torch.inf, -torch.inf],
                        [-torch.inf, -torch.inf, -torch.inf, -torch.inf, -torch.inf, -torch.inf, 0.0, -torch.inf],
                        [-torch.inf, -torch.inf, -torch.inf, -torch.inf, -torch.inf, -torch.inf, -torch.inf, 0.0],
                    ],
                ]
            ),
        ),
    ],
)
def test_compute_mask_outputs_correct_masking(
    dimensions: Sequence[int], window_size: Sequence[int], shift_size: Sequence[int], expected_mask: Sequence[int]
):
    # arrange & act
    mask = _compute_mask(dimensions, window_size, shift_size, torch.device("cpu"))

    # assert
    assert torch.equal(mask, expected_mask)


@pytest.mark.parametrize(
    ["input_shape", "window_size", "window_attn_type", "expected_output_shape"],
    [
        ((4, 48, 16, 16), (4, 4), WindowAttention2d, (4, 48, 16, 16)),
        ((4, 48, 15, 15), (4, 4), WindowAttention2d, (4, 48, 15, 15)),
        ((4, 48, 16, 16, 16), (4, 4, 4), WindowAttention3d, (4, 48, 16, 16, 16)),
        ((4, 48, 16, 16, 16), (4, 4, 4), WindowAttention3d, (4, 48, 16, 16, 16)),
    ],
)
def test_swin_transformer_layer_forward_correct_shapes(
    input_shape: Sequence[int],
    window_size: Sequence[int],
    window_attn_type: WindowAttentionBase,
    expected_output_shape: Sequence[int],
) -> None:
    # arrange
    input = torch.randn(input_shape)
    channels = input_shape[1]
    swin_layer = SwinTransformerStage(
        dim=channels,
        window_attention=window_attn_type,
        depth=2,
        num_heads=3,
        window_size=window_size,
        drop_path=0.0,
    )

    # act
    output, _ = swin_layer.forward(input)

    # assert
    assert output.shape == expected_output_shape


@pytest.mark.parametrize(
    ["input_shape", "window_size", "window_attn_type", "expected_output_shape"],
    [
        ((4, 48, 16, 16), (4, 4), WindowAttention2d, ((4, 48, 16, 16), (4, 96, 8, 8))),
        # output is padded during downsampling
        ((4, 48, 15, 15), (4, 4), WindowAttention2d, ((4, 48, 15, 15), (4, 96, 8, 8))),
        ((4, 48, 16, 16, 16), (4, 4, 4), WindowAttention3d, ((4, 48, 16, 16, 16), (4, 96, 8, 8, 8))),
        # output is padded during downsampling
        ((4, 48, 15, 15, 15), (4, 4, 4), WindowAttention3d, ((4, 48, 15, 15, 15), (4, 96, 8, 8, 8))),
    ],
)
def test_swin_transformer_layer_forward_with_downsampling_correct_shape(
    input_shape: Sequence[int],
    window_size: Sequence[int],
    window_attn_type: WindowAttentionBase,
    expected_output_shape: tuple[Sequence[int], Sequence[int]],
) -> None:
    # arrange
    input = torch.randn(input_shape)
    channels = input_shape[1]
    swin_layer = SwinTransformerStage(
        dim=channels,
        window_attention=window_attn_type,
        depth=2,
        num_heads=3,
        window_size=window_size,
        drop_path=0.0,
        downsample=PatchMerging,
    )
    expected_output_shape_before_downsample = expected_output_shape[0]
    expected_output_shape_after_downsample = expected_output_shape[1]

    # act
    output_after_downsample, output_pre_downsampling = swin_layer.forward(input)

    # assert
    assert output_after_downsample.shape == expected_output_shape_after_downsample
    assert output_pre_downsampling.shape == expected_output_shape_before_downsample
