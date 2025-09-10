from typing import Sequence, Union

import pytest
import torch

from aifocore.models import SwinTransformer
from aifocore.models.modules import WindowAttention2d, WindowAttention3d, WindowAttentionBase, PatchEmbed, PatchEmbed3d


@pytest.mark.parametrize(
    ["input_shape", "embed_dim", "patch_size", "window_size", "window_attn", "patch_embed", "expected_output_shapes"],
    [
        # 2d: standard case
        (
            (4, 3, 64, 64),
            48,
            (2, 2),
            (4, 4),
            WindowAttention2d,
            PatchEmbed,
            [(4, 48, 32, 32), (4, 48, 32, 32), (4, 96, 16, 16), (4, 192, 8, 8), (4, 384, 4, 4)],
        ),
        # 2d: uneven window size
        (
            (4, 3, 64, 64),
            48,
            (2, 2),
            (3, 3),
            WindowAttention2d,
            PatchEmbed,
            [(4, 48, 32, 32), (4, 48, 32, 32), (4, 96, 16, 16), (4, 192, 8, 8), (4, 384, 4, 4)],
        ),
        # 3d: standard case
        (
            (4, 3, 64, 64, 64),
            48,
            (2, 2, 2),
            (4, 4, 4),
            WindowAttention3d,
            PatchEmbed3d,
            [(4, 48, 32, 32, 32), (4, 48, 32, 32, 32), (4, 96, 16, 16, 16), (4, 192, 8, 8, 8), (4, 384, 4, 4, 4)],
        ),
        # 3d: uneven window size
        (
            (4, 3, 64, 64, 64),
            48,
            (2, 2, 2),
            (3, 3, 3),
            WindowAttention3d,
            PatchEmbed3d,
            [(4, 48, 32, 32, 32), (4, 48, 32, 32, 32), (4, 96, 16, 16, 16), (4, 192, 8, 8, 8), (4, 384, 4, 4, 4)],
        ),
    ],
)
def test_swin_transformer_forward_all_stages_correct_shapes(
    input_shape: Sequence[int],
    embed_dim: int,
    patch_size: Sequence[int],
    window_size: Sequence[int],
    window_attn: type[WindowAttentionBase],
    patch_embed: Union[type[PatchEmbed], type[PatchEmbed3d]],
    expected_output_shapes: list[Sequence[int]],
) -> None:
    # arrange
    input = torch.randn(input_shape)
    _, C, *spatial_dims = input.shape
    swin_transformer = SwinTransformer(
        in_chans=C,
        embed_dim=embed_dim,
        img_size=tuple(spatial_dims),
        window_size=window_size,
        patch_size=patch_size,
        depths=(2, 2, 2, 2),
        num_heads=(3, 6, 12, 24),
        window_attn=window_attn,
        patch_embed=patch_embed,
    )

    # act
    output_list = swin_transformer.forward(input)

    # assert
    output_shapes = [tuple(tensor.shape) for tensor in output_list]
    assert output_shapes == expected_output_shapes
