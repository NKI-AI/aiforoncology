import torch
from fomo.transforms.transforms import RandomGaussianBlur3d, RandomGamma


def test_3d_gaussian_blur_forward_all_ones_remains_constant():
    # arrange
    kernel_size = (3, 3, 3)
    sigma_range = (1.0, 1.0)
    p = 1.0
    input_matrix = torch.ones((1, 1, 3, 3, 3)).type(torch.float)

    # act
    transform = RandomGaussianBlur3d(kernel_size=kernel_size, sigma=sigma_range, p=p)
    transformed_input = transform(input_matrix)

    assert torch.allclose(input_matrix, transformed_input)


def test_3d_gaussian_blur_forward_impulse_gives_kernel():
    # arrange
    kernel_size = (3, 3, 3)
    sigma_range = (1.0, 1.0)
    p = 1.0
    input_matrix = torch.zeros((5, 5, 5)).type(torch.float)
    input_matrix[2, 2, 2] = 1
    input_matrix = input_matrix.unsqueeze(0).unsqueeze(0)
    gaussian_kernel = torch.tensor(
        [
            [[0.0206, 0.0339, 0.0206], [0.0339, 0.0560, 0.0339], [0.0206, 0.0339, 0.0206]],
            [[0.0339, 0.0560, 0.0339], [0.0560, 0.0923, 0.0560], [0.0339, 0.0560, 0.0339]],
            [[0.0206, 0.0339, 0.0206], [0.0339, 0.0560, 0.0339], [0.0206, 0.0339, 0.0206]],
        ]
    )

    # act
    transform = RandomGaussianBlur3d(kernel_size=kernel_size, sigma=sigma_range, p=p)
    transformed_input = transform(input_matrix).squeeze(0).squeeze(0)
    transformed_input = transformed_input[1:4, 1:4, 1:4]

    # assert
    assert torch.allclose(transformed_input, gaussian_kernel, rtol=1e-4, atol=1e-4)


def test_3d_gaussian_blur_forward_p_zero_no_transform():
    # arrange
    kernel_size = (3, 3, 3)
    sigma_range = (1.0, 1.0)
    p = 0.0
    input_matrix = torch.ones((1, 1, 5, 5, 5)).type(torch.float)

    # act
    transform = RandomGaussianBlur3d(kernel_size=kernel_size, sigma=sigma_range, p=p)
    transformed_input = transform(input_matrix)

    # assert
    assert torch.all(torch.eq(input_matrix, transformed_input))


def test_random_gamma_forward_no_gain_gamma_applied_same_matrix():
    # arrange
    gain_range = (1.0, 1.0)
    gamma_range = (1.0, 1.0)
    p = 1.0
    input_matrix = torch.rand((1, 1, 16, 224, 224)).type(torch.float)

    # act
    transform = RandomGamma(gain_range=gain_range, gamma_range=gamma_range, p=p)
    transformed_input = transform(input_matrix)

    # assert
    assert torch.equal(input_matrix, transformed_input)


def test_random_gamma_forward_with_gain_and_gamma_test_2d():
    # arrange
    gain_range = (1.1, 1.1)
    gamma_range = (1.1, 1.1)
    p = 1.0
    input_matrix = torch.rand((1, 1, 224, 224)).type(torch.float)

    # act
    transform = RandomGamma(gain_range=gain_range, gamma_range=gamma_range, p=p)
    transformed_input = transform(input_matrix)

    input_matrix = 1.1 * torch.pow(input_matrix, 1.1)
    input_matrix = input_matrix.clamp(0.0, 1.0)

    # assert
    assert torch.allclose(input_matrix, transformed_input)


def test_random_gamma_forward_with_gain_and_gamma_test_3d():
    # arrange
    gain_range = (1.1, 1.1)
    gamma_range = (1.1, 1.1)
    p = 1.0
    input_matrix = torch.rand((1, 1, 16, 224, 224)).type(torch.float)

    # act
    transform = RandomGamma(gain_range=gain_range, gamma_range=gamma_range, p=p)
    transformed_input = transform(input_matrix)

    input_matrix = 1.1 * torch.pow(input_matrix, 1.1)
    input_matrix = input_matrix.clamp(0.0, 1.0)

    # assert
    assert torch.allclose(input_matrix, transformed_input)
