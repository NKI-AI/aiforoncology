# Copyright 2025 AI for Oncology Research Group. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
import importlib
import functools
from typing import Any, Callable, Literal, Sequence
from typing_extensions import override

import torch
import torch.nn.functional as F
from torchvision.transforms import v2
from eva.vision.data import tv_tensors as eva_tv_tensors
from torchvision import tv_tensors
import monai.transforms.spatial.array as monai_spatial
import monai.transforms.croppad.array as monai_pad
from monai.utils import (
    Method,
    PytorchPadMode,
)


class ChannelUnsqueeze(v2.Transform):
    """
    Transform to add a singleton dimension at a given axis.
    """

    def __init__(self, dim: int = 0) -> None:
        super().__init__()
        self._dim = dim

    @functools.singledispatchmethod
    @override
    def transform(self, inpt: torch.Tensor, _: dict[str, Any]) -> torch.Tensor:
        return inpt.unsqueeze(self._dim)

    @transform.register(tv_tensors.Image)
    @transform.register(tv_tensors.Mask)
    @transform.register(eva_tv_tensors.Volume)
    def _(self, inpt: tv_tensors.TVTensor, _: dict[str, Any]) -> tv_tensors.TVTensor:
        """
        Transform a tensor by adding a dimension.

        Parameters
        ----------
        inpt : tv_tensors.TVTensor
            Input tensor
        _ : dict[str, Any]
            Unused params dictionary

        Returns
        -------
        tv_tensors.TVTensor
            Tensor with an added dimension
        """
        unsqueezed_inpt = inpt.unsqueeze(self._dim)
        return tv_tensors.wrap(unsqueezed_inpt, like=inpt)


class ChannelSqueeze(v2.Transform):
    """
    Transform to remove a singleton dimension at a given axis.
    """

    def __init__(self, dim: int = 0) -> None:
        super().__init__()
        self._dim = dim

    @functools.singledispatchmethod
    @override
    def transform(self, inpt: torch.Tensor, _: dict[str, Any]) -> torch.Tensor:
        return inpt.squeeze(self._dim)

    @transform.register(tv_tensors.Image)
    @transform.register(tv_tensors.Mask)
    @transform.register(eva_tv_tensors.Volume)
    def _(self, inpt: tv_tensors.TVTensor, _: dict[str, Any]) -> tv_tensors.TVTensor:
        """
        Transform a tensor by removing a dimension.

        Parameters
        ----------
        inpt : tv_tensors.TVTensor
            Input tensor
        _ : dict[str, Any]
            Unused params dictionary

        Returns
        -------
        tv_tensors.TVTensor
            Tensor with the dimension removed
        """
        squeezed_inpt = inpt.squeeze(self._dim)
        return tv_tensors.wrap(squeezed_inpt, like=inpt)


class Interpolate(v2.Transform):
    """
    Transform to interpolate a tensor to a given size.
    """

    def __init__(
        self,
        size: int | Sequence[int] = 224,
        mode: Literal["nearest", "nearest-exact", "linear", "bilinear", "bicubic", "trilinear", "area"] = "bilinear",
        mask_mode: Literal["nearest", "nearest-exact", "linear", "bilinear", "bicubic", "trilinear", "area"]
        | None = None,
    ) -> None:
        """
        Parameters
        ----------
        size : int or sequence of int, default=224
            Desired output size
        mode : str, default="bilinear"
            Interpolation mode to use
        mask_mode : str or None, default=None
            Interpolation mode to use for masks, if None masks won't be interpolated
        """
        super().__init__()
        self.size = size
        self.mode = mode
        self._mask_mode = mask_mode

    @functools.singledispatchmethod
    @override
    def transform(self, inpt: Any, _: dict[str, Any]) -> Any:
        return inpt

    @transform.register(tv_tensors.Image)
    @transform.register(eva_tv_tensors.Volume)
    def _(self, inpt: tv_tensors.TVTensor, _: dict[str, Any]) -> tv_tensors.TVTensor:
        """
        Transform a tensor by interpolating to the target size.

        Parameters
        ----------
        inpt : tv_tensors.TVTensor
            Input tensor
        _ : dict[str, Any]
            Unused params dictionary

        Returns
        -------
        tv_tensors.TVTensor
            Interpolated tensor
        """
        resized_inpt = F.interpolate(inpt.unsqueeze(0), self.size, mode=self.mode).squeeze(0)
        return tv_tensors.wrap(resized_inpt, like=inpt)

    @transform.register(tv_tensors.Mask)
    def _(self, inpt: tv_tensors.Mask, _: dict[str, Any]) -> tv_tensors.Mask:
        if self._mask_mode is None:
            return inpt

        resized_inpt = F.interpolate(inpt.unsqueeze(0), self.size, mode=self._mask_mode).squeeze(0)
        return tv_tensors.wrap(resized_inpt, like=inpt)


class ValueClipper(v2.Transform):
    """
    Transform to clip tensor values to a specified range.
    """

    def __init__(self, min_val: float, max_val: float) -> None:
        """
        Parameters
        ----------
        min_val : float
            Minimum value to clip to
        max_val : float
            Maximum value to clip to
        """
        super().__init__()
        self.min_val = min_val
        self.max_val = max_val

    @functools.singledispatchmethod
    @override
    def transform(self, inpt: Any, _: dict[str, Any]) -> Any:
        return inpt

    @transform.register(tv_tensors.Image)
    @transform.register(eva_tv_tensors.Volume)
    def _(self, inpt: tv_tensors.TVTensor, _: dict[str, Any]) -> tv_tensors.TVTensor:
        """
        Transform a tensor by clipping its values.

        Parameters
        ----------
        inpt : tv_tensors.TVTensor
            Input tensor
        _ : dict[str, Any]
            Unused params dictionary

        Returns
        -------
        tv_tensors.TVTensor
            Tensor with values clipped to [min_val, max_val]
        """
        clipped_inpt = torch.clamp(inpt, self.min_val, self.max_val)
        return tv_tensors.wrap(clipped_inpt, like=inpt)


class MinMaxNormalizer(v2.Transform):
    """
    Transform to normalize tensor values to [0, 1] range using min-max scaling.
    """

    def __init__(self):
        super().__init__()

    @functools.singledispatchmethod
    @override
    def transform(self, inpt: Any, _: dict[str, Any]) -> Any:
        return inpt

    @transform.register(tv_tensors.Image)
    @transform.register(eva_tv_tensors.Volume)
    def _(self, inpt: tv_tensors.TVTensor, _: dict[str, Any]) -> tv_tensors.TVTensor:
        """
        Transform a tensor by normalizing to [0, 1] range.

        Parameters
        ----------
        inpt : tv_tensors.TVTensor
            Input tensor
        _ : dict[str, Any]
            Unused params dictionary

        Returns
        -------
        tv_tensors.TVTensor
            Normalized tensor with values in [0, 1]
        """
        dims = tuple(range(1, inpt.dim()))
        min_val = inpt.amin(dim=dims, keepdim=True)
        max_val = inpt.amax(dim=dims, keepdim=True)
        denom = (max_val - min_val).clamp(min=1e-7)
        normalized_inpt = (inpt - min_val) / denom
        return tv_tensors.wrap(normalized_inpt, like=inpt)


class Resize(v2.Transform):
    def __init__(
        self,
        spatial_size: Sequence[int] | int,
        size_mode: str = "all",
        mode: Literal["nearest", "nearest-exact", "linear", "bilinear", "bicubic", "trilinear", "area"] = "area",
        mask_mode: Literal["nearest", "nearest-exact", "linear", "bilinear", "bicubic", "trilinear", "area"]
        | None = None,
        align_corners: bool | None = None,
        anti_aliasing: bool = False,
        anti_aliasing_sigma: Sequence[float] | float | None = None,
        dtype: torch.dtype = torch.float32,
        lazy: bool = False,
    ) -> None:
        super().__init__()

        self._resize = monai_spatial.Resize(
            spatial_size, size_mode, mode, align_corners, anti_aliasing, anti_aliasing_sigma, dtype, lazy
        )
        self._mask_mode = mask_mode

    @functools.singledispatchmethod
    @override
    def transform(self, inpt: Any, _: dict[str, Any]) -> Any:
        return inpt

    @transform.register(tv_tensors.Image)
    @transform.register(eva_tv_tensors.Volume)
    def _(self, inpt: tv_tensors.TVTensor, _: dict[str, Any]) -> tv_tensors.TVTensor:
        resized_inpt = self._resize(inpt)
        return tv_tensors.wrap(resized_inpt, like=inpt)

    @transform.register(tv_tensors.Mask)
    def _(self, inpt: tv_tensors.Mask, _: dict[str, Any]) -> tv_tensors.Mask:
        if self._mask_mode is None:
            return inpt

        added_channel_dimension = False
        if len(inpt.shape) == 3:
            inpt.unsqueeze_(0)
            added_channel_dimension = True

        resized_inpt = self._resize(inpt, mode=self._mask_mode)

        if added_channel_dimension:
            resized_inpt.squeeze_(0)

        return tv_tensors.wrap(resized_inpt, like=inpt)


class DivisiblePad(v2.Transform):
    def __init__(
        self,
        k: Sequence[int] | int,
        mode: str = PytorchPadMode.CONSTANT,
        method: str = Method.SYMMETRIC,
        lazy: bool = False,
        mask_mode: str | None = None,
        mask_constant_value: int = 0,
        transform_kwargs: dict[str, Any] | None = None,
    ) -> None:
        super().__init__()

        self._pad = monai_pad.DivisiblePad(k, mode, method, lazy, **transform_kwargs)
        self._mask_mode = mask_mode
        self._mask_constant_value = mask_constant_value

    @functools.singledispatchmethod
    @override
    def transform(self, inpt: Any, _: dict[str, Any]) -> Any:
        return inpt

    @transform.register(tv_tensors.Image)
    @transform.register(eva_tv_tensors.Volume)
    def _(self, inpt: tv_tensors.TVTensor, _: dict[str, Any]) -> tv_tensors.TVTensor:
        resized_inpt = self._pad(inpt)
        return tv_tensors.wrap(resized_inpt, like=inpt)

    @transform.register(tv_tensors.Mask)
    def _(self, inpt: tv_tensors.Mask, _: dict[str, Any]) -> tv_tensors.Mask:
        if self._mask_mode is None:
            return inpt

        if self._mask_mode == PytorchPadMode.CONSTANT:
            resized_inpt = self._pad(inpt, mode=self._mask_mode, constant_values=self._mask_constant_value)
            return tv_tensors.wrap(resized_inpt, like=inpt)

        resized_inpt = self._pad(inpt, mode=self._mask_mode)
        return tv_tensors.wrap(resized_inpt, like=inpt)


class ToDtype(v2.ToDtype):
    def __init__(self, dtype: dict[str, torch.dtype] | torch.dtype, scale: bool = False) -> None:
        """
        Wrapper for torchvision's ToDtype that handles string-based type specifications.

        Parameters
        ----------
        dtype : dict[str, torch.dtype] | torch.dtype
            The dtype to convert to. If a dict is provided, keys should be string representations
            of tv_tensor types (e.g., "torchvision.tv_tensors.Image") which will be converted
            to their actual types.
        scale : bool, optional
            Whether to scale the values for images or videos, by default False

        Notes
        -----
        Torch's v2.ToDtype maps a tv_tensor type to a dtype. This class provides a wrapper to instantiate
        dictionary keys to their respective types.
        """
        if isinstance(dtype, torch.dtype):
            super().__init__(dtype, scale)
            return

        typed_dict = {}
        for key, val in dtype.items():
            mapped_type = self._get_tensor_type(key)
            typed_dict[mapped_type] = val

        super().__init__(typed_dict, scale)

    @staticmethod
    def _get_tensor_type(object_name: str) -> tv_tensors.TVTensor | str:
        """
        Convert a string representation of a type to the actual type.

        Parameters
        ----------
        object_name : str
            String representation of a type (e.g., "torchvision.tv_tensors.Image")

        Returns
        -------
        tv_tensors.TVTensor | str
            The actual type object or "others" if that was the input

        Raises
        ------
        ModuleNotFoundError
            When the module specified in the string does not exist
        AttributeError
            When the module does not have a type associated with the provided string
        """
        if object_name == "others":
            return "others"

        module_name, _, type_name = object_name.rpartition(".")
        module = importlib.import_module(module_name)
        return getattr(module, type_name)


class DinoPretrainTransforms(v2.Compose):
    """Transforms for a pretrained DINO-CT model"""

    def __init__(
        self,
        size: int | Sequence[int] = 224,
        mode: str = "bilinear",
        min_val: float = -1008.0,
        max_val: float = 822.0,
        mean: Sequence[float] = (0.0,),
        std: Sequence[float] = (1.0,),
        apply_feature_scaling=False,
    ) -> None:
        """
        Parameters
        ----------
        size : int or sequence of int, default=224
            Desired output size of the crop. If size is an int instead
            of sequence like (h, w), a square crop (size, size) is made.
        mode : str, default="bilinear"
            Interpolation mode to use
        min_val : float, default=-1008.0
            Minimum value to clip to
        max_val : float, default=822.0
            Maximum value to clip to
        mean : sequence of float, default=(0.0,)
            Sequence of means for each image channel
        std : sequence of float, default=(1.0,)
            Sequence of standard deviations for each image channel
        """
        self._size = size
        self._mode = mode
        self._mean = mean
        self._std = std
        self._min_val = min_val
        self._max_val = max_val
        self._apply_feature_scaling = apply_feature_scaling

        super().__init__(transforms=self._build_transforms())

    def _build_transforms(self) -> Sequence[Callable]:
        """
        Builds and returns the list of transforms.

        Returns
        -------
        Sequence[Callable]
            List of transforms to apply
        """
        transforms = [
            ChannelUnsqueeze(),
            Interpolate(self._size, self._mode),
            ChannelSqueeze(),
            ValueClipper(self._min_val, self._max_val),
        ]

        if self._apply_feature_scaling:
            transforms.append(MinMaxNormalizer())

        transforms.append(
            v2.Normalize(
                mean=self._mean,
                std=self._std,
            )
        )
        return transforms
