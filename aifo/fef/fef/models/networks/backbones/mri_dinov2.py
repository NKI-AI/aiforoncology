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

import os
import functools
from typing import Callable, Literal

import timm
import torch
import torch.nn as nn
from eva.core.models import wrappers
from timm.models import vision_transformer
from timm.models.registry import register_model


@register_model
def mri_vit_base_patch14_224(pretrained=False, **kwargs):
    """Custom ViT model with DINOv2 architecture, patch size 14 and image size 224x224."""
    # Remove arguments not used by VisionTransformer.__init__
    kwargs.pop("pretrained_cfg", None)
    kwargs.pop("pretrained_cfg_overlay", None)
    kwargs.pop("features_only", None)
    kwargs.pop("out_indices", None)
    kwargs.pop("cache_dir", None)  # Remove cache_dir if present

    # Default parameters for DINOv2 base model with patch size 14
    img_size = kwargs.pop("img_size", 224)
    patch_size = kwargs.pop("patch_size", 14)
    in_chans = kwargs.pop("in_chans", 1)  # MRI images are single channel
    num_classes = kwargs.pop("num_classes", 0)  # No classification head
    embed_dim = kwargs.pop("embed_dim", 768)
    depth = kwargs.pop("depth", 12)
    num_heads = kwargs.pop("num_heads", 12)
    mlp_ratio = kwargs.pop("mlp_ratio", 4.0)
    qkv_bias = kwargs.pop("qkv_bias", True)
    norm_layer = kwargs.pop("norm_layer", functools.partial(nn.LayerNorm, eps=1e-6))
    init_values = kwargs.pop("init_values", 1e-5)
    no_embed_class = kwargs.pop("no_embed_class", True)  # DINOv2 doesn't use class token
    dynamic_img_size = kwargs.pop("dynamic_img_size", True)

    model = vision_transformer.VisionTransformer(
        img_size=img_size,
        patch_size=patch_size,
        in_chans=in_chans,
        num_classes=num_classes,
        embed_dim=embed_dim,
        depth=depth,
        num_heads=num_heads,
        mlp_ratio=mlp_ratio,
        qkv_bias=qkv_bias,
        norm_layer=norm_layer,
        init_values=init_values,
        no_embed_class=no_embed_class,
        global_pool="",  # No global pooling for encoder-only mode
        **kwargs,
    )
    model.dynamic_img_size = dynamic_img_size

    return model


class DINOv2VIT(wrappers.BaseModel):
    """Model wrapper for DINOv2 Vision Transformer with dynamic image sizing and MRI support."""

    def __init__(
        self,
        tensor_transforms: Callable | None = None,
        checkpoint_path: str | None = None,
        timm_ckpt: str | None = None,
        model_name: str = "mri_vit_base_patch14_224",
        pretrained: bool = False,
        out_indices: int | None = None,
        model_kwargs: dict | None = None,
        output_tokens: Literal["patch", "cls", "all"] = "patch",
    ) -> None:
        """Initialize the DINOv2 ViT encoder.

        Parameters
        ----------
        tensor_transforms : Callable, optional
            The transforms to apply to the output tensor.
        checkpoint_path : str, optional
            Path to a DINOv2 checkpoint to load.
        timm_ckpt : str, optional
            Path to save the converted TIMM checkpoint.
        model_name : str, default="mri_vit_base_patch14_224"
            Name of the model to instantiate.
        pretrained : bool, default=False
            If set to True, load pretrained weights.
        out_indices : int, optional
            Returns specific block outputs if provided.
        model_kwargs : dict, optional
            Extra model arguments.
        output_tokens : {"patch", "cls", "all"}, default="patch"
            Type of tokens to output.
        """
        super().__init__(tensor_transforms=tensor_transforms)

        self._checkpoint_path = checkpoint_path
        self._timm_ckpt = timm_ckpt
        self._model_name = model_name
        self._pretrained = pretrained
        self._out_indices = out_indices
        self._model_kwargs = model_kwargs or {}
        self._output_tokens = output_tokens

        if self._checkpoint_path and not self._timm_ckpt:
            raise ValueError("timm_ckpt cannot be None if checkpoint_path is provided")

        self._forward_fn = self._select_forward_function()

        self.load_model()

    def load_model(self) -> None:
        """Builds and loads the DINOv2 model as a feature extractor."""
        model_kwargs = self._model_kwargs.copy()
        model_kwargs.pop("dynamic_img_size", None)
        model_kwargs.pop("cache_dir", None)  # Remove cache_dir if present

        self._feature_extractor = timm.create_model(
            model_name=self._model_name,
            pretrained=self._pretrained,
            out_indices=self._out_indices,
            features_only=self._out_indices is not None,
            **model_kwargs,
        )

        self._feature_extractor.dynamic_img_size = True

        if hasattr(self._feature_extractor, "patch_embed"):
            self._feature_extractor.patch_embed.img_size = None
            self._feature_extractor.patch_embed.flatten = False
            self._feature_extractor.patch_embed.output_fmt = "NHWC"
            self._feature_extractor.patch_embed.last_grid_size = None

        if self._checkpoint_path:
            state_dict = self._load_dinov2_checkpoint(self._checkpoint_path)
            if self._timm_ckpt:
                os.makedirs(os.path.dirname(self._timm_ckpt), exist_ok=True)
                torch.save(state_dict, self._timm_ckpt)
            self._feature_extractor.load_state_dict(state_dict, strict=False)
            print(f"Checkpoint loaded from {self._checkpoint_path}")
        else:
            print("No checkpoint path provided.")

        self._orig_forward_features = self._feature_extractor.forward_features

        def forward_features_custom(x):
            """Custom forward_features that fixes device mismatch issue and stores grid size for dynamic (non-square) sizing
            Example:
            Input Image (224x280)     →    Patch Grid (16x20)    →    Sequence (320 patches)    →    2D Output (16x20)
            [224x280 pixels]          →    [16 rows x 20 cols]   →    [1 x 320 x embed_dim]     →    [16 x 20 x embed_dim]
                                           (with 14x14 patches)
            """
            device = x.device
            if next(self._feature_extractor.parameters()).device != device:
                self._feature_extractor = self._feature_extractor.to(device)

            # Store input shape to convert transformer sequence back to 2d features later
            _, _, h, w = x.shape
            patch_size = self._feature_extractor.patch_embed.patch_size
            if isinstance(patch_size, tuple):
                patch_h, patch_w = patch_size
            else:
                patch_h = patch_w = patch_size

            grid_h = (h + patch_h - 1) // patch_h
            grid_w = (w + patch_w - 1) // patch_w

            if hasattr(self._feature_extractor.patch_embed, "last_grid_size"):
                self._feature_extractor.patch_embed.last_grid_size = (grid_h, grid_w)

            return self._orig_forward_features(x)

        self._feature_extractor.forward_features = forward_features_custom

    def _load_dinov2_checkpoint(self, path: str) -> dict[str, torch.Tensor]:
        """Load and convert DINOv2 checkpoint."""
        checkpoint = torch.load(path, map_location=torch.device("cpu"))
        if "teacher" in checkpoint:
            state_dict = checkpoint["teacher"]
        elif "state_dict" in checkpoint:
            state_dict = checkpoint["state_dict"]
        else:
            state_dict = checkpoint

        return self._convert_dinov2_to_timm(state_dict)

    def _convert_dinov2_to_timm(self, state_dict: dict[str, torch.Tensor]) -> dict[str, torch.Tensor]:
        """Convert DINOv2 state dict to TIMM format."""
        timm_state_dict = {}
        for key, value in state_dict.items():
            if key.startswith("dino_head.") or key.startswith("head."):
                continue
            if key.startswith("backbone."):
                key = key.replace("backbone.", "")
            timm_state_dict[key] = value

        return vision_transformer._convert_dinov2(timm_state_dict, self._feature_extractor)

    def _select_forward_function(self) -> Callable:
        """Select the appropriate forward function based on output_tokens setting."""
        if self._output_tokens == "cls":
            return self._forward_cls
        elif self._output_tokens == "patch":
            return self._forward_patch
        else:  # 'all'
            return self._forward_all

    def _forward_cls(self, features: torch.Tensor) -> torch.Tensor:
        """Forward function for CLS token output."""
        return features[:, 0, :]

    def _forward_patch(self, features: torch.Tensor) -> torch.Tensor:
        """Forward function for patch token output."""
        batch_size, _, embed_dim = features.shape
        start_idx = 0
        if hasattr(self._feature_extractor, "cls_token") and self._feature_extractor.cls_token is not None:
            start_idx += 1
        if hasattr(self._feature_extractor, "reg_token") and self._feature_extractor.reg_token is not None:
            start_idx += self._feature_extractor.reg_token.shape[1]

        patch_tokens = features[:, start_idx:, :]
        num_patches = patch_tokens.size(1)

        if hasattr(self._feature_extractor, "patch_embed") and hasattr(
            self._feature_extractor.patch_embed, "last_grid_size"
        ):
            grid_h, grid_w = self._feature_extractor.patch_embed.last_grid_size
        else:
            # if no last_grid_size, fall back to square grid estimation
            grid_h = int(num_patches**0.5)
            grid_w = num_patches // grid_h
            while grid_h * grid_w != num_patches:
                grid_h -= 1
                grid_w = num_patches // grid_h
                if grid_h <= 0:
                    raise ValueError(f"Cannot determine grid dimensions for {num_patches} patches")

        return patch_tokens.transpose(1, 2).reshape(batch_size, embed_dim, grid_h, grid_w)

    def _forward_all(self, features: torch.Tensor) -> torch.Tensor:
        """Forward function for all tokens output."""
        return features.transpose(1, 2)

    def model_forward(self, tensor: torch.Tensor) -> torch.Tensor:
        """Forward pass through the model with device handling.

        Parameters
        ----------
        tensor : torch.Tensor
            Input tensor of shape (batch_size, channels, height, width)

        Returns
        -------
        torch.Tensor
            Output features with shape depending on selected output_tokens mode
        """
        device = tensor.device
        if next(self._feature_extractor.parameters()).device != device:
            self._feature_extractor = self._feature_extractor.to(device)

        # Calculate grid size before forward pass
        _, _, h, w = tensor.shape
        patch_size = self._feature_extractor.patch_embed.patch_size
        if isinstance(patch_size, tuple):
            patch_h, patch_w = patch_size
        else:
            patch_h = patch_w = patch_size

        grid_h = (h + patch_h - 1) // patch_h
        grid_w = (w + patch_w - 1) // patch_w
        grid_size = (grid_h, grid_w)

        features = self._feature_extractor.forward_features(tensor)

        if self._output_tokens == "patch":
            return self._forward_patch_with_grid(features, grid_size)

        return self._forward_fn(features)

    def _forward_patch_with_grid(self, features: torch.Tensor, grid_size: tuple) -> torch.Tensor:
        """Forward function for patch token output with explicit grid dimensions.

        Parameters
        ----------
        features : torch.Tensor
            Input tensor of shape (batch_size, seq_length, embed_dim)
        grid_size : tuple
            Tuple of (height, width) for the patch grid

        Returns
        -------
        torch.Tensor
            Patch token features of shape (batch_size, embed_dim, H, W)
        """
        batch_size, _, embed_dim = features.shape
        grid_h, grid_w = grid_size

        start_idx = 0
        if hasattr(self._feature_extractor, "cls_token") and self._feature_extractor.cls_token is not None:
            start_idx += 1
        if hasattr(self._feature_extractor, "reg_token") and self._feature_extractor.reg_token is not None:
            start_idx += self._feature_extractor.reg_token.shape[1]

        patch_tokens = features[:, start_idx:, :]

        # Verify that we have the expected number of patches
        num_patches = patch_tokens.size(1)
        expected_patches = grid_h * grid_w

        if num_patches != expected_patches:
            print(
                f"Warning: Number of patches ({num_patches}) doesn't match grid size ({grid_h}x{grid_w}={expected_patches})"
            )
            # Recalculate grid size based on actual number of patches
            grid_h = int(num_patches**0.5)
            grid_w = num_patches // grid_h
            while grid_h * grid_w != num_patches:
                grid_h -= 1
                grid_w = num_patches // grid_h
                if grid_h <= 0:
                    raise ValueError(f"Cannot reshape {num_patches} patches to a grid")

        return patch_tokens.transpose(1, 2).reshape(batch_size, embed_dim, grid_h, grid_w)

    def _ensure_same_device(self, tensor: torch.Tensor) -> None:
        """Ensure the model is on the same device as the input tensor.

        Parameters
        ----------
        tensor : torch.Tensor
            Input tensor to match device with
        """
        model_device = next(self._feature_extractor.parameters()).device
        tensor_device = tensor.device

        if model_device != tensor_device:
            print(f"Moving model from {model_device} to {tensor_device}")
            self._feature_extractor = self._feature_extractor.to(tensor_device)
