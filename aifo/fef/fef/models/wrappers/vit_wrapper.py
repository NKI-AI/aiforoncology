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
from typing import Any, Callable, Literal
from eva.vision.models import wrappers
from monai.inferers.inferer import Inferer
from torch import Tensor
from typing_extensions import override
from fef.utils.typing import OutputType
from timm.models import VisionTransformer


class ViTWrapper(wrappers.TimmModel):
    """Vision Transformer (ViT) wrapper for timm models.

    A wrapper for timm's Vision Transformer models that allows for flexible
    output configurations based on the specified output type.

    Parameters
    ----------
    model_name : str
        Name of the timm model to load.
    output_type : Literal[OutputType.CLS, OutputType.PATCH]
        Type of output to produce:
        - CLS: Return the classification token output (first token)
        - PATCH: Return patch embeddings for segmentation tasks
    pretrained : bool, default=False
        Whether to load ImageNet pretrained weights.
    checkpoint_path : str, default=""
        Path to a model checkpoint to load.
    out_indices : int or tuple[int, ...] or None, default=None
        Indices of transformer blocks to return features from when using
        segmentation output type.
    norm : bool, default=True
        Whether to apply layer normalization to outputs.
    output_fmt : str, default="NCHW"
        Output tensor format for segmentation outputs.
    intermediates_only : bool, default=True
        If True, return only intermediate features without the final layer.
    model_kwargs : dict[str, Any] or None, default=None
        Additional keyword arguments passed to the underlying model.
    tensor_transforms : Callable[..., Any] or None, default=None
        Optional transforms to apply to the input tensor.

    Attributes
    ----------
    _model : VisionTransformer, private
        The underlying timm VisionTransformer model.
    _forward : Callable, private
        The selected forward function based on the output_type.
    """

    def __init__(
        self,
        model_name: str,
        output_type: Literal[OutputType.CLS, OutputType.PATCH],
        pretrained: bool = False,
        checkpoint_path: str = "",
        out_indices: int | tuple[int, ...] | None = 1,
        norm: bool = True,
        output_fmt: str = "NCHW",
        intermediates_only=True,
        model_kwargs: dict[str, Any] | None = None,
        tensor_transforms: Callable[..., Any] | None = None,
        inferer: Inferer | None = None,
    ) -> None:
        super().__init__(
            model_name=model_name,
            pretrained=pretrained,
            checkpoint_path=checkpoint_path,
            out_indices=None,
            model_kwargs=model_kwargs,
            tensor_transforms=tensor_transforms,
        )

        self._norm = norm
        self._output_fmt = output_fmt
        self._intermediates_only = intermediates_only
        self._out_indices = out_indices
        self._inferer = inferer
        self._model: VisionTransformer

        match output_type:
            case OutputType.PATCH:
                self._forward = self._segmentation_fwd
            case OutputType.CLS:
                self._forward = self._classification_fwd
            case _:
                raise ValueError(
                    f"The {OutputType.__name__}: '{output_type}' is not supported for {ViTWrapper.__name__}."
                )

    def _segmentation_fwd(self, tensor: Tensor) -> Tensor:
        return self._model.forward_intermediates(
            tensor,
            indices=self._out_indices,
            norm=self._norm,
            output_fmt=self._output_fmt,
            intermediates_only=self._intermediates_only,
        )

    def _classification_fwd(self, tensor: Tensor) -> Tensor:
        return self._model.forward_features(tensor)[:, 0, :]

    @override
    def model_forward(self, tensor: Tensor) -> Tensor:
        return self._inferer(tensor, self._forward) if self._inferer and not self.training else self._forward(tensor)
