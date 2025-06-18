import torch
from fomo.transforms.random_resize_gpu import bilinear_interpolate


def test_bilinear_interpolate_midpoint():
    # Arrange
    # Create a simple 2x2 single-channel image
    input_tensor = torch.zeros(1, 1, 2, 2)
    input_tensor[0, 0, 0, 0] = 1.0  # top-left
    input_tensor[0, 0, 0, 1] = 2.0  # top-right
    input_tensor[0, 0, 1, 0] = 3.0  # bottom-left
    input_tensor[0, 0, 1, 1] = 4.0  # bottom-right

    # Create a 1x1 grid that samples exactly the center point
    grid = torch.zeros(1, 1, 1, 2)  # Shape: (1, 1, 1, 2)

    # Act
    output = bilinear_interpolate(input_tensor, grid)

    # Assert
    # Center point should be average of four corners
    expected = torch.tensor([[[[2.5]]]])  # (1+2+3+4)/4 = 2.5
    torch.testing.assert_close(output, expected)


def test_bilinear_interpolate_2x_upscale():
    # Arrange
    # Create a 2x2 test image
    input_tensor = torch.tensor([[[[1.0, 2.0], [3.0, 4.0]]]])  # Shape: (1, 1, 2, 2)

    # Create a 3x3 grid for upscaling
    y, x = torch.meshgrid(torch.linspace(-1, 1, 3), torch.linspace(-1, 1, 3), indexing="ij")
    grid = torch.stack([x, y], dim=-1).unsqueeze(0)  # Shape: (1, 3, 3, 2)

    # Act
    output = bilinear_interpolate(input_tensor, grid)

    # Assert
    # Center is 2.5
    assert output[0, 0, 1, 1] == 2.5


def test_bilinear_interpolate_identity_on_aligned_corners():
    # Arrange
    # Create a simple 4x4 single-channel test image
    input_tensor = torch.zeros(1, 1, 4, 4)
    # Set the four corners to different values
    input_tensor[0, 0, 0, 0] = 1.0  # top-left
    input_tensor[0, 0, 0, 3] = 2.0  # top-right
    input_tensor[0, 0, 3, 0] = 3.0  # bottom-left
    input_tensor[0, 0, 3, 3] = 4.0  # bottom-right

    # Identity grid (sample exact same locations)
    y, x = torch.meshgrid(torch.linspace(-1, 1, 4), torch.linspace(-1, 1, 4), indexing="ij")
    grid = torch.stack([x, y], dim=-1).unsqueeze(0)  # Shape: (1, 4, 4, 2)

    # Act
    output = bilinear_interpolate(input_tensor, grid, align_corners=True)

    # Assert
    # Output should match input with identity grid
    torch.testing.assert_close(output, input_tensor)
