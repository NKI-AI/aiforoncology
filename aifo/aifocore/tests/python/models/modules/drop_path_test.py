from typing import Sequence

import torch
import pytest

from aifocore.models.modules import DropPath


def test_forward_returns_identical_on_not_training() -> None:
    # arrange
    input = torch.randn((4, 16, 16, 3))
    drop_path = DropPath(drop_prob=0.2)
    drop_path.eval()

    # act
    output = drop_path.forward(input)

    # assert
    assert torch.equal(input, output)


def test_forward_returns_identical_on_no_drop_path() -> None:
    # arrange
    input = torch.randn((4, 16, 16, 3))
    drop_path = DropPath(drop_prob=0.0)

    # act
    output = drop_path.forward(input)

    # assert
    assert torch.equal(input, output)


@pytest.mark.parametrize("shape", [(4, 16, 16, 3), (4, 16, 16, 16, 3)])
def test_forward_returns_same_shape_on_dropped_paths(shape: Sequence[int]) -> None:
    # arrange
    input = torch.randn(shape)
    drop_path = DropPath(drop_prob=0.5)

    # act
    output = drop_path.forward(input)

    # assert
    assert output.shape == shape
