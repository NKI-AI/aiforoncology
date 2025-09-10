# Copyright (c) Meta Platforms, Inc. and affiliates.
#
# This source code is licensed under the Apache License, Version 2.0
# found in the LICENSE file in the root directory of this source tree.

# References:
#   https://github.com/facebookresearch/dino/blob/master/vision_transformer.py
#   https://github.com/rwightman/pytorch-image-models/tree/master/timm/layers/patch_embed.py

import logging
import os
import warnings
from typing import Any, Callable, Dict, Optional, Tuple

import torch
from torch import nn

from .attention import Attention, MemEffAttention
from .drop_path import DropPath
from .layer_scale import LayerScale
from .mlp import Mlp

logger = logging.getLogger("dinov2")


XFORMERS_ENABLED = os.environ.get("XFORMERS_DISABLED") is None
try:
    if XFORMERS_ENABLED:
        from xformers.ops import fmha, index_select_cat, scaled_index_add

        XFORMERS_AVAILABLE = True
        warnings.warn("xFormers is available (Block)")
    else:
        warnings.warn("xFormers is disabled (Block)")
        raise ImportError
except ImportError:
    XFORMERS_AVAILABLE = False

    warnings.warn("xFormers is not available (Block)")


class Block(nn.Module):
    """Transformer block with multi-head self-attention and MLP.

    Parameters
    ----------
    dim : int
        Dimension of the input features.
    num_heads : int
        Number of attention heads, by default 8.
    mlp_ratio : float, optional
        Ratio of the hidden dimension in the MLP to the input dimension, by default 4.0.
    qkv_bias : bool, optional
        Whether to add a bias to the query, key, and value projections, by default False.
    proj_bias : bool, optional
        Whether to add a bias to the output projection, by default True.
    ffn_bias : bool, optional
        Whether to add a bias to the MLP layers, by default True.
    drop : float, optional
        Dropout rate for the MLP layers, by default 0.0.
    attn_drop : float, optional
        Dropout rate for the attention weights, by default 0.0.
    init_values : float or torch.Tensor, optional
        Initial values for the layer scale, by default None. If a tensor is provided, it should have shape (dim,).
    drop_path : float, optional
        Drop path rate for stochastic depth, by default 0.0.
    act_layer : Callable[..., nn.Module], optional
        Activation layer for the MLP, by default nn.GELU.
    norm_layer : Callable[..., nn.Module], optional
        Normalization layer, by default nn.LayerNorm.
    attn_class : Callable[..., nn.Module], optional
        Attention class to use, by default Attention. Can be replaced with :class:`MemEffAttention` for memory-efficient
        attention.
    ffn_layer : Callable[..., nn.Module], optional
        MLP class to use, by default Mlp.

    Raises
    ------
    ValueError
        If `dim` is not divisible by `num_heads`.
    """

    def __init__(
        self,
        dim: int,
        num_heads: int,
        mlp_ratio: float = 4.0,
        qkv_bias: bool = False,
        proj_bias: bool = True,
        ffn_bias: bool = True,
        drop: float = 0.0,
        attn_drop: float = 0.0,
        init_values=None,
        drop_path: float = 0.0,
        act_layer: Callable[..., nn.Module] = nn.GELU,
        norm_layer: Callable[..., nn.Module] = nn.LayerNorm,
        attn_class: Callable[..., nn.Module] = Attention,
        ffn_layer: Callable[..., nn.Module] = Mlp,
    ) -> None:
        """Inits :class:`Block`.

        Parameters
        ----------
        dim : int
            Dimension of the input features.
        num_heads : int
            Number of attention heads, by default 8.
        mlp_ratio : float, optional
            Ratio of the hidden dimension in the MLP to the input dimension, by default 4.0.
        qkv_bias : bool, optional
            Whether to add a bias to the query, key, and value projections, by default False.
        proj_bias : bool, optional
            Whether to add a bias to the output projection, by default True.
        ffn_bias : bool, optional
            Whether to add a bias to the MLP layers, by default True.
        drop : float, optional
            Dropout rate for the MLP layers, by default 0.0.
        attn_drop : float, optional
            Dropout rate for the attention weights, by default 0.0.
        init_values : float or torch.Tensor, optional
            Initial values for the layer scale, by default None. If a tensor is provided, it should have shape (dim,).
        drop_path : float, optional
            Drop path rate for stochastic depth, by default 0.0.
        act_layer : Callable[..., nn.Module], optional
            Activation layer for the MLP, by default nn.GELU.
        norm_layer : Callable[..., nn.Module], optional
            Normalization layer, by default nn.LayerNorm.
        attn_class : Callable[..., nn.Module], optional
            Attention class to use, by default Attention. Can be replaced with :class:`MemEffAttention` for
            memory-efficient attention.
        ffn_layer : Callable[..., nn.Module], optional
            MLP class to use, by default Mlp.

        Raises
        ------
        ValueError
            If `dim` is not divisible by `num_heads`.
        """
        super().__init__()
        if dim % num_heads != 0:
            raise ValueError(f"dim {dim} should be divisible by num_heads {num_heads}.")
        # print(f"biases: qkv: {qkv_bias}, proj: {proj_bias}, ffn: {ffn_bias}")
        self.norm1 = norm_layer(dim)
        self.attn = attn_class(
            dim,
            num_heads=num_heads,
            qkv_bias=qkv_bias,
            proj_bias=proj_bias,
            attn_drop=attn_drop,
            proj_drop=drop,
        )
        self.ls1 = LayerScale(dim, init_values=init_values) if init_values else nn.Identity()
        self.drop_path1 = DropPath(drop_path) if drop_path > 0.0 else nn.Identity()

        self.norm2 = norm_layer(dim)
        self.mlp = ffn_layer(
            in_features=dim,
            hidden_features=int(dim * mlp_ratio),
            act_layer=act_layer,
            drop=drop,
            bias=ffn_bias,
        )
        self.ls2 = LayerScale(dim, init_values=init_values) if init_values else nn.Identity()
        self.drop_path2 = DropPath(drop_path) if drop_path > 0.0 else nn.Identity()

        self.sample_drop_ratio = drop_path

    def forward(self, x: torch.Tensor) -> torch.Tensor:
        """Forward pass of :class:`Block`.

        Parameters
        ----------
        x : torch.Tensor
            Input tensor of shape (B, N, C) where B is the batch size, N is the sequence length, and C is
            the feature dimension.

        Returns
        -------
        torch.Tensor
            Output tensor of shape (B, N, C) after applying the transformer block.
        """

        def attn_residual_func(x: torch.Tensor) -> torch.Tensor:
            return self.ls1(self.attn(self.norm1(x)))

        def ffn_residual_func(x: torch.Tensor) -> torch.Tensor:
            return self.ls2(self.mlp(self.norm2(x)))

        if self.training and self.sample_drop_ratio > 0.1:
            # the overhead is compensated only for a drop path rate larger than 0.1
            x = drop_add_residual_stochastic_depth(
                x,
                residual_func=attn_residual_func,
                sample_drop_ratio=self.sample_drop_ratio,
            )
            x = drop_add_residual_stochastic_depth(
                x,
                residual_func=ffn_residual_func,
                sample_drop_ratio=self.sample_drop_ratio,
            )
        elif self.training and self.sample_drop_ratio > 0.0:
            x = x + self.drop_path1(attn_residual_func(x))
            x = x + self.drop_path1(ffn_residual_func(x))  # FIXME: drop_path2
        else:
            x = x + attn_residual_func(x)
            x = x + ffn_residual_func(x)
        return x


def drop_add_residual_stochastic_depth(
    x: torch.Tensor,
    residual_func: Callable[[torch.Tensor], torch.Tensor],
    sample_drop_ratio: float = 0.0,
) -> torch.Tensor:
    """Applies stochastic depth by dropping a subset of samples in the batch and adding a residual.

    This function extracts a random subset of the batch, applies a residual function to it, and adds the result back
    to the original tensor, scaling the residual appropriately.

    Parameters
    ----------
    x : torch.Tensor
        Input tensor of shape (B, N, D) where B is the batch size, N is the sequence length, and D is the
        feature dimension.
    residual_func : Callable[[torch.Tensor], torch.Tensor]
        Function that takes a tensor of shape (B', N, D) and returns a tensor of the same shape, representing the
        residual.
    sample_drop_ratio : float, optional
        Ratio of samples to drop from the batch, by default 0.0. If set to 0.0, no samples are dropped.

    Returns
    -------
    torch.Tensor
        Output tensor of the same shape as input x, with the residual added back to the original tensor.
    """
    # 1) extract subset using permutation
    B = x.shape[0]
    sample_subset_size = max(int(B * (1 - sample_drop_ratio)), 1)
    brange = (torch.randperm(B, device=x.device))[:sample_subset_size]
    x_subset = x[brange]

    # 2) apply residual_func to get residual
    residual = residual_func(x_subset)

    x_flat = x.flatten(1)
    residual = residual.flatten(1)

    residual_scale_factor = B / sample_subset_size

    # 3) add the residual
    x_plus_residual = torch.index_add(x_flat, 0, brange, residual.to(dtype=x.dtype), alpha=residual_scale_factor)
    return x_plus_residual.view_as(x)


def get_branges_scales(x: torch.Tensor, sample_drop_ratio: float = 0.0) -> tuple[torch.Tensor, float]:
    """Generates random indices for dropping samples in the batch and computes the scale factor for the residual.

    This function extracts a random subset of the batch and computes a scale factor based on the original batch size
    and the size of the subset. The scale factor is used to scale the residual when it is added back to the original
    tensor.

    Parameters
    ----------
    x : torch.Tensor
        Input tensor of shape (B, N, D) where B is the batch size, N is the sequence length, and D is the
        feature dimension.
    sample_drop_ratio : float, optional
        Ratio of samples to drop from the batch, by default 0.0. If set to 0.0, no samples are dropped.

    Returns
    -------
    tuple[torch.Tensor, float]
        A tuple containing:
        - brange: A tensor of indices representing the subset of the batch to keep.
        - residual_scale_factor: A float representing the scale factor for the residual.
    """

    B = x.shape[0]
    sample_subset_size = max(int(B * (1 - sample_drop_ratio)), 1)
    brange = (torch.randperm(B, device=x.device))[:sample_subset_size]
    residual_scale_factor = B / sample_subset_size
    return brange, residual_scale_factor


def add_residual(
    x: torch.Tensor,
    brange: torch.Tensor,
    residual: torch.Tensor,
    residual_scale_factor: float,
    scaling_vector: Optional[torch.Tensor] = None,
) -> torch.Tensor:
    """Adds a residual to the input tensor, scaling it appropriately.

    This function takes a tensor `x`, a set of indices `brange`, and a residual tensor, and adds the residual to the
    corresponding indices in `x`. If a scaling vector is provided, it scales the residual before adding it.

    Parameters
    ----------
    x : torch.Tensor
        Input tensor of shape (B, N, D) where B is the batch size, N is the sequence length, and D is the
        feature dimension.
    brange : torch.Tensor
        torch.Tensor of indices representing the subset of the batch to which the residual will be added.
    residual : torch.Tensor
        Residual tensor of shape (B', N, D) where B' is the size of the subset defined by `brange`.
    residual_scale_factor : float
        Scale factor for the residual, computed as the ratio of the original batch size to the subset size.
    scaling_vector : Optional[torch.Tensor], optional
        Scaling vector to scale the residual before adding it, by default None. If provided, it should have shape (D,).

    Returns
    -------
    torch.Tensor
        Output tensor of the same shape as input `x`, with the residual added back to the original tensor.
    """
    if scaling_vector is None:
        x_flat = x.flatten(1)
        residual = residual.flatten(1)
        x_plus_residual = torch.index_add(x_flat, 0, brange, residual.to(dtype=x.dtype), alpha=residual_scale_factor)
    else:
        x_plus_residual = scaled_index_add(
            x,
            brange,
            residual.to(dtype=x.dtype),
            scaling=scaling_vector,
            alpha=residual_scale_factor,
        )
    return x_plus_residual


attn_bias_cache: Dict[Tuple, Any] = {}


def get_attn_bias_and_cat(
    x_list: list[torch.Tensor], branges: Optional[list[torch.Tensor]] = None
) -> tuple[Any, torch.Tensor]:
    """Get attention bias and concatenate tensors from a list of tensors.

    This function checks if the attention bias for the given shapes is already cached. If not, it creates a new
    attention bias using the `fmha.BlockDiagonalMask` from xFormers. It then concatenates the tensors in `x_list`
    based on the provided `branges`. If `branges` is not provided, it concatenates the tensors directly.

    Parameters
    ----------
    x_list : list of torch.Tensors
        List of tensors to concatenate. Each tensor should have shape (B, N, D) where B is the batch size, N is the
        sequence length, and D is the feature dimension.
    branges : list of torch.Tensors, optional
        List of tensors containing indices for selecting samples from the batch. If provided, it will index select
        and concatenate the tensors in `x_list`. If not provided, it will concatenate the tensors directly.

    Returns
    -------
    tuple[Any, torch.Tensor]
        A tuple containing:
        - attn_bias: Attention bias tensor created using `fmha.BlockDiagonalMask` from xFormers.
        - cat_tensors: Concatenated tensor of shape (1, B', D) where B' is the total number of samples selected from
        the batch based on `branges` or the total number of samples in `x_list` if `branges` is not provided.
    If `branges` is provided, the concatenated tensor will have shape (1, sum of sizes in branges, D).
    If `branges` is not provided, the concatenated tensor will have shape (1, sum of batch sizes in x_list, D).
    """

    batch_sizes = [b.shape[0] for b in branges] if branges is not None else [x.shape[0] for x in x_list]
    all_shapes = tuple((b, x.shape[1]) for b, x in zip(batch_sizes, x_list))
    if all_shapes not in attn_bias_cache.keys():
        seqlens = []
        for b, x in zip(batch_sizes, x_list):
            for _ in range(b):
                seqlens.append(x.shape[1])
        attn_bias = fmha.BlockDiagonalMask.from_seqlens(seqlens)
        attn_bias._batch_sizes = batch_sizes
        attn_bias_cache[all_shapes] = attn_bias

    if branges is not None:
        cat_tensors = index_select_cat([x.flatten(1) for x in x_list], branges).view(1, -1, x_list[0].shape[-1])
    else:
        tensors_bs1 = tuple(x.reshape([1, -1, *x.shape[2:]]) for x in x_list)
        cat_tensors = torch.cat(tensors_bs1, dim=1)

    return attn_bias_cache[all_shapes], cat_tensors


def drop_add_residual_stochastic_depth_list(
    x_list: list[torch.Tensor],
    residual_func: Callable[[torch.Tensor, Any], torch.Tensor],
    sample_drop_ratio: float = 0.0,
    scaling_vector=None,
) -> list[torch.Tensor]:
    """Applies stochastic depth to a list of tensors, dropping a subset of samples in each tensor and adding a residual.
    This function processes a list of tensors, generating random indices for dropping samples in each tensor,
    computing the attention bias, and applying a residual function to each tensor. The results are then combined
    and returned as a list of tensors.

    Parameters
    ----------
    x_list : list of torch.Tensors
        List of tensors to process. Each tensor should have shape (B, N, D) where B is the batch size, N is the sequence
        length, and D is the feature dimension.
    residual_func : Callable[[torch.Tensor, Any], torch.Tensor]
        Function that takes a tensor of shape (B', N, D) and an attention bias (if applicable) and returns a tensor of
        the same shape, representing the residual.
    sample_drop_ratio : float, optional
        Ratio of samples to drop from the batch, by default 0.0. If set to 0.0, no samples are dropped.
    scaling_vector : Optional[torch.Tensor], optional
        Scaling vector to scale the residual before adding it, by default None. If provided, it should have shape (D,).

    Returns
    -------
    list of torch.Tensors
        List of output tensors, each of the same shape as the corresponding input tensor in `x_list`, with the residual
        added back to the original tensor.
    """
    # 1) generate random set of indices for dropping samples in the batch
    branges_scales = [get_branges_scales(x, sample_drop_ratio=sample_drop_ratio) for x in x_list]
    branges = [s[0] for s in branges_scales]
    residual_scale_factors = [s[1] for s in branges_scales]

    # 2) get attention bias and index+concat the tensors
    attn_bias, x_cat = get_attn_bias_and_cat(x_list, branges)

    # 3) apply residual_func to get residual, and split the result
    residual_list = attn_bias.split(residual_func(x_cat, attn_bias=attn_bias))  # type: ignore

    outputs = []
    for x, brange, residual, residual_scale_factor in zip(x_list, branges, residual_list, residual_scale_factors):
        outputs.append(add_residual(x, brange, residual, residual_scale_factor, scaling_vector).view_as(x))
    return outputs


class NestedTensorBlock(Block):
    """Transformer block with multi-head self-attention and MLP, supporting nested tensors.

    This class extends the :class:`Block` class to support nested tensors, allowing for more flexible input shapes.

    Parameters
    ----------
    dim : int
        Dimension of the input features.
    num_heads : int
        Number of attention heads, by default 8.
    mlp_ratio : float, optional
        Ratio of the hidden dimension in the MLP to the input dimension, by default 4.0.
    qkv_bias : bool, optional
        Whether to add a bias to the query, key, and value projections, by default False.
    proj_bias : bool, optional
        Whether to add a bias to the output projection, by default True.
    ffn_bias : bool, optional
        Whether to add a bias to the feed-forward network, by default True.
    drop : float, optional
        Dropout rate for the MLP layers, by default 0.0.
    attn_drop : float, optional
        Dropout rate for the attention weights, by default 0.0.
    init_values : float or torch.Tensor, optional
        Initial values for the layer scale, by default None. If a tensor is provided, it should have shape (dim,).
    drop_path : float, optional
        Drop path rate for stochastic depth, by default 0.0.
    act_layer : Callable[..., nn.Module], optional
        Activation layer for the MLP, by default nn.GELU.
    norm_layer : Callable[..., nn.Module], optional
        Normalization layer, by default nn.LayerNorm.
    attn_class : Callable[..., nn.Module], optional
        Attention class to use, by default Attention. Can be replaced with :class:`MemEffAttention` for
        memory-efficient attention.
    ffn_layer : Callable[..., nn.Module], optional
        MLP class to use, by default :class:`Mlp`.
    sample_drop_ratio : float, optional
        Drop path rate for stochastic depth, by default 0.0. This is used to control the stochastic depth
        during training.
    """

    def forward_nested(self, x_list: list[torch.Tensor]) -> list[torch.Tensor]:
        """Forward pass for list of tensors, applying attention and MLP with stochastic depth.

        This method applies the attention and MLP layers to a list of tensors, applying stochastic depth if the model is
        in training mode and `sample_drop_ratio` is greater than 0.0. It uses the :class:`MemEffAttention` class
        for memory-efficient attention. The method expects `x_list` to be a list of tensors, where each tensor has
        the same feature dimension. If the model is not in training mode or `sample_drop_ratio` is 0.0,
        it applies the attention and MLP layers without stochastic depth.

        Parameters
        ----------
        x_list : list[torch.Tensor]
            List of tensors to process. Each tensor should have shape (B, N, D) where B is the batch size, N is the
            sequence length, and D is the feature dimension.

        Returns
        -------
        list[torch.Tensor]
            List of processed tensors, each with the same shape as the corresponding input tensor in `x_list`.
        """
        assert isinstance(self.attn, MemEffAttention)

        if self.training and self.sample_drop_ratio > 0.0:

            def attn_residual_func(x: torch.Tensor, attn_bias=None) -> torch.Tensor:
                return self.attn(self.norm1(x), attn_bias=attn_bias)

            def ffn_residual_func(x: torch.Tensor, attn_bias=None) -> torch.Tensor:
                return self.mlp(self.norm2(x))

            x_list = drop_add_residual_stochastic_depth_list(
                x_list,
                residual_func=attn_residual_func,
                sample_drop_ratio=self.sample_drop_ratio,
                scaling_vector=(self.ls1.gamma if isinstance(self.ls1, LayerScale) else None),
            )
            x_list = drop_add_residual_stochastic_depth_list(
                x_list,
                residual_func=ffn_residual_func,
                sample_drop_ratio=self.sample_drop_ratio,
                scaling_vector=(self.ls2.gamma if isinstance(self.ls1, LayerScale) else None),
            )
            return x_list
        else:

            def attn_residual_func(x: torch.Tensor, attn_bias=None) -> torch.Tensor:
                return self.ls1(self.attn(self.norm1(x), attn_bias=attn_bias))

            def ffn_residual_func(x: torch.Tensor, attn_bias=None) -> torch.Tensor:
                return self.ls2(self.mlp(self.norm2(x)))

            attn_bias, x = get_attn_bias_and_cat(x_list)
            x = x + attn_residual_func(x, attn_bias=attn_bias)
            x = x + ffn_residual_func(x)
            return attn_bias.split(x)

    def forward(self, x_or_x_list: torch.Tensor | list[torch.Tensor]) -> torch.Tensor | list[torch.Tensor]:
        """Forward pass of :class:`NestedTensorBlock`.

        Parameters
        ----------
        x_or_x_list : torch.Tensor or list[torch.Tensor]
            Input tensor or list of tensors. If a tensor is provided, it should have shape (B, N, D) where B is the
            batch size, N is the sequence length, and D is the feature dimension. If a list of tensors is provided,
            each tensor should have the same shape.

        Returns
        -------
        torch.Tensor or list[torch.Tensor]
            Output tensor or list of tensors after applying the transformer block. If a tensor is provided, the output
            will be a tensor of the same shape. If a list of tensors is provided, the output will be a list of tensors,
            each with the same shape as the corresponding input tensor.

        Raises
        ------
        AssertionError
            If `xFormers` is not available.
        ValueError
            If `x_or_x_list` is neither a torch.Tensor nor a list of torch.Tensors.
        """
        if isinstance(x_or_x_list, torch.Tensor):
            return super().forward(x_or_x_list)
        elif isinstance(x_or_x_list, list):
            if not XFORMERS_AVAILABLE:
                raise AssertionError("xFormers is required for using nested tensors")
            return self.forward_nested(x_or_x_list)
        else:
            raise ValueError(
                f"Expected input to be a torch.Tensor or a list of torch.Tensors, got {type(x_or_x_list)}."
            )
