from typing import List, Optional, Tuple, Union

import torch
from torchvision.transforms.v2 import RandomResizedCrop
from kornia.enhance import (
    adjust_brightness_accumulative,
    adjust_contrast_with_mean_subtraction,
    normalize,
)
from kornia.filters import gaussian_blur2d, filter3d
from kornia.geometry import elastic_transform2d
from kornia.filters.kernels import get_gaussian_kernel3d
from torch import Tensor, nn


class PercentileClipper(nn.Module):
    def __init__(self, percentile: float) -> None:
        super().__init__()
        self.percentile = percentile / 100.0

    def forward(self, x: torch.Tensor) -> torch.Tensor:
        batch_size = x.size(0)
        x_flat = x.view(batch_size, -1)
        upper_bound = torch.quantile(x_flat, self.percentile, dim=1, keepdim=True)
        upper_bound = upper_bound.view(batch_size, *([1] * (x.dim() - 1)))
        x = torch.min(x, upper_bound)
        return x


class ValueClipper(nn.Module):
    def __init__(self, min_val: float, max_val: float) -> None:
        super().__init__()
        self.min_val = min_val
        self.max_val = max_val

    def forward(self, x: torch.Tensor) -> torch.Tensor:
        torch.clamp_(x, self.min_val, self.max_val)
        return x


class MinMaxNormalizer(nn.Module):
    def __init__(self):
        super().__init__()

    def forward(self, x: torch.Tensor) -> torch.Tensor:
        dims = tuple(range(1, x.dim()))
        min_val = x.amin(dim=dims, keepdim=True)
        max_val = x.amax(dim=dims, keepdim=True)
        denom = (max_val - min_val).clamp(min=1e-7)
        x = (x - min_val) / denom
        return x


class ChannelDuplicator(nn.Module):
    def __init__(self, num_channels: int) -> None:
        super().__init__()
        self.num_channels = num_channels

    def forward(self, x: torch.Tensor) -> torch.Tensor:
        x = x.repeat(self.num_channels, 1, 1)
        return x


class RelativeGaussianNoise(nn.Module):
    def __init__(self, std_factor: Union[float, Tuple[float, float]] = 0.05, p: float = 1.0):
        super(RelativeGaussianNoise, self).__init__()
        self.std_factor = self._check_input(std_factor, "std_factor")
        self.p = p

    @staticmethod
    def _check_input(value, name):
        if isinstance(value, float):
            if value < 0:
                raise ValueError(f"{name} should be non-negative")
            return [value, value]  # Single value converted to a range with identical bounds
        elif isinstance(value, (tuple, list)) and len(value) == 2:
            if value[0] > value[1]:
                raise ValueError(f"{name} range should be in the format (min, max)")
            return list(value)
        else:
            raise TypeError(f"{name} should be a float or a tuple of two floats")

    def forward(self, x):
        batch_size = x.shape[0]
        device = x.device
        dtype = x.dtype

        apply_noise = torch.rand(batch_size, device=device) < self.p
        if not apply_noise.any():
            return x

        std_factor = torch.empty(batch_size, device=device).uniform_(*self.std_factor)
        std_factor = std_factor.view(-1, 1, 1, 1)

        x_std = x.view(batch_size, -1).std(dim=1).view(batch_size, 1, 1, 1)
        std = x_std * std_factor

        noise = torch.randn_like(x) * std

        mask = apply_noise.view(-1, 1, 1, 1).to(dtype)
        x = x + noise * mask

        return x


class RandomResizedCrop3d(nn.Module):
    def __init__(
        self,
        size: Tuple[int, int, int],
        scale: Tuple[float, float] = (0.08, 1.0),
        ratio: Tuple[float, float] = (3 / 4, 4 / 3),
        interpolation_mode: str = "trilinear",
    ) -> None:
        """
        Random resized crop operation for 3D volumes.

        Parameters
        ----------
        size : tuple of int
            Expected output size of the crop (depth, height, width)
        scale : tuple of float
            Range (min, max) of size of the origin crop relative to original image size
        ratio : tuple of float
            Range (min, max) of aspect ratio (width/height) of the origin crop
        interpolation_mode : str
            Desired interpolation mode for resizing ('trilinear', etc.)
        """
        super().__init__()
        self.size = size
        self.scale = scale
        self.ratio = ratio
        self.interpolation_mode = interpolation_mode

    def forward(self, x: torch.Tensor) -> torch.Tensor:
        # Get random crop parameters for height and width
        # Torch takes 2 last dims of input, so we can just input x
        h_crop_loc, w_crop_loc, crop_h, crop_w = RandomResizedCrop.get_params(x, self.scale, self.ratio)

        # Handle depth dimension
        _, _, D, _, _ = x.shape
        # We can solve having less slices in the depth than the requested number of slices in
        # several ways. In this case, we stretch the depth of the crop with the interpolation below.
        # TODO: Maybe reconsider the handling of depth at a later point in time.
        crop_d = min(self.size[0], D)

        # If crop depth is less than full depth, choose random location
        if crop_d < D:
            d_crop_loc = torch.randint(0, D - crop_d + 1, size=(1,)).item()
        else:
            d_crop_loc = 0

        crop = x[
            :, :, d_crop_loc : d_crop_loc + crop_d, h_crop_loc : h_crop_loc + crop_h, w_crop_loc : w_crop_loc + crop_w
        ]

        interpolated = torch.nn.functional.interpolate(
            crop,
            size=self.size,
            mode=self.interpolation_mode,
        )

        return interpolated


# Below are transforms that use kornia functional APIs
# Unfortunately, while using the original Kornia classes it inevitably ends up moving tensors at runtime
# (called _detach_tensor_to_cpu in the ImageModuleMixIn class)
# which is a huge performance hit.


class Normalize(nn.Module):
    def __init__(
        self,
        mean: Union[Tensor, Tuple[float, ...], List[float], float],
        std: Union[Tensor, Tuple[float, ...], List[float], float],
        device: Optional[torch.device] = None,
    ):
        super().__init__()
        self.mean = torch.tensor(mean, device=device)
        self.std = torch.tensor(std, device=device)

    def forward(self, x: Tensor) -> Tensor:
        return normalize(x, self.mean, self.std)


class RandomGaussianBlur(nn.Module):
    def __init__(self, kernel_size=(9, 9), sigma=(0.1, 2.0), p=0.5):
        super().__init__()
        self.kernel_size = kernel_size
        self.sigma_range = sigma
        self.p = p

    def forward(self, x):
        batch_size = x.shape[0]
        device = x.device
        apply_blur = torch.rand(batch_size, device=device) < self.p
        if not apply_blur.any():
            return x

        sigma = torch.empty(batch_size, device=device).uniform_(*self.sigma_range)
        sigma = sigma.unsqueeze(1).expand(-1, 2)

        x_blurred = gaussian_blur2d(x, self.kernel_size, sigma)
        x = torch.where(apply_blur.view(-1, 1, 1, 1), x_blurred, x)
        return x


class RandomGamma(nn.Module):
    def __init__(
        self,
        gain_range: Tuple[float, float] = (1.0, 1.0),
        gamma_range: Tuple[float, float] = (1.0, 1.0),
        p: float = 1.0,
    ):
        """
        Creates a transform for dimensionality-agnostic random gamma adjustment.

        Parameters
        ----------
        gain_range : tuple of float
            The range from which to sample the gain factor.
            The gain factor is multiplied with the input tensor after gamma correction.
            Default is (1.0, 1.0), meaning no gain adjustment.
        gamma_range : tuple of float
            The range from which to sample the gamma value.
            The gamma value is used to adjust the intensity of the input tensor.
            Default is (1.0, 1.0), meaning no gamma adjustment.
        p : float
            Probability of applying the gamma adjustment to each sample in the batch.
            Default is 1.0, meaning the adjustment is always applied.
        """
        super().__init__()
        self.gamma_range = gamma_range
        self.gain_range = gain_range
        self.p = p

    def forward(self, x: torch.Tensor) -> torch.Tensor:
        batch_size = x.shape[0]
        device = x.device
        apply_gamma = torch.rand(batch_size, device=device) < self.p
        if not apply_gamma.any():
            return x

        gamma = torch.empty(batch_size, device=device).uniform_(*self.gamma_range)
        gain = torch.empty(batch_size, device=device).uniform_(*self.gain_range)
        gamma = gamma.view(-1, *([1] * (x.dim() - 1)))
        gain = gain.view(-1, *([1] * (x.dim() - 1)))

        x_adjusted = gain * torch.pow(x, gamma)  # Apply gamma correction
        x_adjusted = torch.clamp(x_adjusted, 0.0, 1.0)

        x = torch.where(apply_gamma.view(-1, *([1] * (x.dim() - 1))), x_adjusted, x)
        return x


class ColorJitter(nn.Module):
    def __init__(
        self,
        brightness: Union[float, Tuple[float, float]] = 0.0,
        contrast: Union[float, Tuple[float, float]] = 0.0,
        p: float = 1.0,
    ):
        super().__init__()
        self.brightness = self._check_input(brightness, "brightness", center=1.0)
        self.contrast = self._check_input(contrast, "contrast", center=1.0)
        self.p = p

        self.transform_funcs = []
        if self.brightness[0] != 1.0 or self.brightness[1] != 1.0:
            self.transform_funcs.append(self._apply_brightness)
        if self.contrast[0] != 1.0 or self.contrast[1] != 1.0:
            self.transform_funcs.append(self._apply_contrast)

    @staticmethod
    def _check_input(value, name, center):
        if isinstance(value, float):
            if value < 0:
                raise ValueError(f"{name} value should be non-negative")
            return [center - value, center + value]
        elif isinstance(value, (tuple, list)) and len(value) == 2:
            return value
        else:
            raise TypeError(f"{name} should be a float or a tuple of two floats")

    def _apply_brightness(self, x, batch_size, apply_jitter):
        brightness_factor = torch.empty(batch_size, device=x.device).uniform_(*self.brightness)
        brightness_factor = brightness_factor.view(-1, 1, 1, 1)
        x_adjusted = adjust_brightness_accumulative(x, brightness_factor)
        x = torch.where(apply_jitter.view(-1, 1, 1, 1), x_adjusted, x)
        return x

    def _apply_contrast(self, x, batch_size, apply_jitter):
        contrast_factor = torch.empty(batch_size, device=x.device).uniform_(*self.contrast)
        contrast_factor = contrast_factor.view(-1, 1, 1, 1)
        x_adjusted = adjust_contrast_with_mean_subtraction(x, contrast_factor)
        x = torch.where(apply_jitter.view(-1, 1, 1, 1), x_adjusted, x)
        return x

    def forward(self, x: torch.Tensor) -> torch.Tensor:
        batch_size = x.shape[0]
        device = x.device
        apply_jitter = torch.rand(batch_size, device=device) < self.p
        if not apply_jitter.any():
            return x

        transform_order = torch.randperm(len(self.transform_funcs)).tolist()

        for idx in transform_order:
            x = self.transform_funcs[idx](x, batch_size, apply_jitter)

        return x


class RandomElasticTransform(nn.Module):
    def __init__(
        self,
        kernel_size: Tuple[int, int] = (63, 63),
        sigma: Tuple[float, float] = (32.0, 32.0),
        alpha: Tuple[float, float] = (1.0, 1.0),
        interpolation_mode: str = "bilinear",
        align_corners: bool = False,
        padding_mode: str = "zeros",
        p: float = 0.5,
    ) -> None:
        super().__init__()
        self.kernel_size = kernel_size
        self.sigma = sigma
        self.alpha = alpha
        self.interpolation_mode = interpolation_mode
        self.align_corners = align_corners
        self.padding_mode = padding_mode
        self.p = p

    def forward(self, x):
        B, _, H, W = x.shape
        device = x.device
        apply_transformation = torch.rand(B, device=device) < self.p
        if not apply_transformation.any():
            return x

        displacement_noise = torch.rand(B, 2, H, W, device=device) * 2 - 1
        x_transformed = elastic_transform2d(
            x,
            displacement_noise,
            self.kernel_size,
            self.sigma,
            self.alpha,
            self.align_corners,
            self.interpolation_mode,
            self.padding_mode,
        )
        x = torch.where(apply_transformation.view(-1, 1, 1, 1), x_transformed, x)
        return x


class RandomGaussianBlur3d(nn.Module):
    def __init__(
        self,
        kernel_size: Tuple[int, int, int] = (9, 9, 9),
        sigma: Tuple[float, float] = (0.1, 2.0),
        border_type: str = "reflect",
        p: float = 0.5,
    ) -> None:
        """
        Random Gaussian blur for 3D inputs.

        Parameters
        ----------
        kernel_size : tuple of int
            Size of the Gaussian kernel (depth, height, width)
        sigma : tuple of float
            Range (min, max) of standard deviation for Gaussian kernel
        border_type : str
            Border padding mode for filtering ('reflect', 'replicate', 'circular', 'constant')
        p : float
            Probability of applying the blur transform
        """
        super().__init__()
        self.kernel_size = kernel_size
        self.sigma_range = sigma
        self.border_type = border_type
        self.p = p

    def forward(self, x: torch.Tensor) -> torch.Tensor:
        batch_size = x.shape[0]
        device = x.device

        # Determine which samples in batch to blur
        apply_blur = torch.rand(batch_size, device=device) < self.p
        if not apply_blur.any():
            return x

        # Generate random sigma values for each sample
        sigma = torch.empty(batch_size, device=device).uniform_(*self.sigma_range)
        sigma = sigma.unsqueeze(1).expand(-1, 3)  # Expand to 3 dimensions

        # Apply 3D Gaussian blur
        kernel = get_gaussian_kernel3d(kernel_size=self.kernel_size, sigma=sigma, device=device)
        x_blurred = filter3d(x, kernel, border_type=self.border_type)

        # Only apply blur to selected samples
        x = torch.where(apply_blur.view(-1, 1, 1, 1, 1), x_blurred, x)
        return x
