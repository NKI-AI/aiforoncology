# Copyright (c) Meta Platforms, Inc. and affiliates.
#
# This source code is licensed under the Apache License, Version 2.0
# found in the LICENSE file in the root directory of dinov2 in third_party.
#
# Modified by NKI-AI4Oncology, 04-2025

import logging
from functools import partial
from typing import Any

import torch
import torch.nn as nn

from fomo.transforms.random_resize_gpu import random_resize_gpu
from fomo.transforms.transforms import (
    Normalize,
    ValueClipper,
    MinMaxNormalizer,
    RandomGaussianBlur3d,
    RandomGamma,
)

logger = logging.getLogger("dinov2")


class DataAugmentationCTBatch3d(object):
    """
    A class that performs batch-wise data augmentation on 3D CT images.

    Parameters
    ----------
    global_crops_scale : list of float
        Scale range for global crops.
    local_crops_scale : list of float
        Scale range for local crops.
    local_crops_number : int
        Number of local crops to generate.
    mean : tuple of float, optional
        Mean values for normalization, default (0.503383,).
    std : tuple of float, optional
        Standard deviation values for normalization, default (0.176303,).
    gaussian_blur_prob : float, optional
        Probability of applying Gaussian blur, default 0.5.
    gamma_prob : float, optional
        Probability of applying gamma adjustment, default 0.7.
    global_crops_size : tuple of int, optional
        Size of global crops (depth, height, width), default (16, 224, 224).
    local_crops_size : tuple of int, optional
        Size of local crops (depth, height, width), default (16, 96, 96).

    Notes
    -----
    This class implements a data augmentation pipeline that applies various kornia and
    custom transformations to batches of 3D CT images, including a number of random resized crops,
    normalization, value clipping, Gaussian blur, and gamma adjustments.
    """

    def __init__(
        self,
        global_crops_scale: list[float],
        local_crops_scale: list[float],
        local_crops_number: int,
        mean: tuple[float, ...] = (0.503383,),
        std: tuple[float, ...] = (0.176303,),
        gaussian_blur_prob: float = 0.5,
        gamma_prob: float = 0.7,
        global_crops_size: tuple[int, int, int] = (16, 224, 224),
        local_crops_size: tuple[int, int, int] = (16, 96, 96),
    ):
        self._global_crops_scale = global_crops_scale
        self._local_crops_scale = local_crops_scale
        self._local_crops_number = local_crops_number
        self._global_crops_size = global_crops_size
        self._local_crops_size = local_crops_size

        self._mean = torch.tensor(mean)
        self._std = torch.tensor(std)
        self._gaussian_blur_prob = gaussian_blur_prob
        self._gamma_prob = gamma_prob
        self._dtype = torch.half

        logger.info("###################################")
        logger.info("Using data augmentation parameters:")
        logger.info("global_crops_scale: %s", global_crops_scale)
        logger.info("local_crops_scale: %s", local_crops_scale)
        logger.info("local_crops_number: %s", local_crops_number)
        logger.info("global_crops_size: %s", global_crops_size)
        logger.info("local_crops_size: %s", local_crops_size)
        logger.info("###################################")

        crop_z_dim = global_crops_size[0]
        output_size_global = (global_crops_size[1], global_crops_size[2])
        output_size_local = (local_crops_size[1], local_crops_size[2])

        # Separate compiles for global and local crops
        compiled_random_resize_gpu_global = torch.compile(
            partial(random_resize_gpu, out_z_dim=crop_z_dim, output_size=output_size_global), backend="inductor"
        )
        compiled_random_resize_gpu_local = torch.compile(
            partial(random_resize_gpu, out_z_dim=crop_z_dim, output_size=output_size_local), backend="inductor"
        )

        # Geometric augmentations
        self._random_resize_crop_global = compiled_random_resize_gpu_global
        self._random_resize_crop_local = compiled_random_resize_gpu_local

        self._global_transfo1_extra = RandomGaussianBlur3d(kernel_size=(9, 9, 9), sigma=(0.1, 2.0), p=1.0)
        self._global_transfo2_extra = RandomGaussianBlur3d(kernel_size=(9, 9, 9), sigma=(0.1, 2.0), p=0.1)
        self._local_transfo_extra = RandomGaussianBlur3d(kernel_size=(9, 9, 9), sigma=(0.1, 2.0), p=gaussian_blur_prob)

        # Normalization
        self._minmax_norm = MinMaxNormalizer()
        self._normalize = Normalize(mean=mean, std=std)
        # 0.5th and 99.5th percentile calculated through nnUNet-v2
        self._value_clipping = ValueClipper(min_val=-1008.0, max_val=822.0)

        # Intesity adjustments
        self._gamma = RandomGamma(gain_range=(0.8, 1.2), p=gamma_prob)

        # Assemble transformations
        global_transform1_list = [
            self._value_clipping,
            self._minmax_norm,
            self._global_transfo1_extra,
            self._normalize,
        ]
        global_transform2_list = [
            self._value_clipping,
            self._minmax_norm,
            self._gamma,
            self._global_transfo2_extra,
            self._normalize,
        ]
        local_transform_list = [
            self._value_clipping,
            self._minmax_norm,
            self._gamma,
            self._local_transfo_extra,
            self._normalize,
        ]

        self._global_transform1 = nn.Sequential(*global_transform1_list)
        self._global_transform2 = nn.Sequential(*global_transform2_list)
        self._local_transform = nn.Sequential(*local_transform_list)

    def __call__(self, pad_and_output_dict: dict[str, torch.Tensor], data: dict[str, Any]):
        padded_tensor = pad_and_output_dict["padded_tensor"]
        global_crop_coords = pad_and_output_dict["global_crop_coords"]
        local_crop_coords = pad_and_output_dict["local_crop_coords"]

        # During the first batch, the augmentations are compiled and this takes a while
        global_crops = self._random_resize_crop_global(padded_tensor, global_crop_coords).unsqueeze(1)
        global_crops = self._random_resize_crop_global(padded_tensor, global_crop_coords).unsqueeze(1)
        local_crops = self._random_resize_crop_local(padded_tensor, local_crop_coords).unsqueeze(1)

        # Get transformed crops
        local_crops_transformed = self._local_transform(local_crops)

        B = global_crops.shape[0] // 2
        global_crop_1 = global_crops[:B]
        global_crop_2 = global_crops[B:]

        global_crop_1_transformed = self._global_transform1(global_crop_1)
        global_crop_2_transformed = self._global_transform2(global_crop_2)

        # Combine transformed crops back
        global_crops_transformed = torch.cat([global_crop_1_transformed, global_crop_2_transformed])

        data["collated_global_crops"] = global_crops_transformed.to(self._dtype)
        data["collated_local_crops"] = local_crops_transformed.to(self._dtype)
