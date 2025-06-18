# Copyright (c) Meta Platforms, Inc. and affiliates.
#
# This source code is licensed under the Apache License, Version 2.0
# found in the LICENSE file in the root directory of this source tree.
#
# Modified by NKI-AI4Oncology, 04-2025

import logging

import torch
import torch.nn as nn
from fomo.transforms.transforms import (
    ChannelDuplicator,
    Normalize,
    RandomElasticTransform,
    RandomGaussianBlur,
    ValueClipper,
)

logger = logging.getLogger("dinov2")


class DataAugmentationCTBatch2d(object):
    def __init__(
        self,
        mean: tuple[float, ...] = (-86.808624,),
        std: tuple[float, ...] = (322.634704,),
        model_in_chans: int = 1,
        gaussian_blur_prob: float = 0.5,
        elastic_transform_prob: float = 0.7,
        dtype=torch.float32,
    ):
        self.gaussian_blur_prob = gaussian_blur_prob
        self.dtype = dtype
        self.mean = torch.tensor(mean)
        self.std = torch.tensor(std)
        self.device = torch.device("cuda" if torch.cuda.is_available() else "cpu")

        logger.info("###################################")
        logger.info("Using data augmentation parameters:")
        logger.info("normalization_mean: %s", mean)
        logger.info("normalization_std: %s", std)
        logger.info("gaussian_blur_prob: %s", gaussian_blur_prob)
        logger.info("elastic_transform_prob: %s", elastic_transform_prob)
        logger.info("###################################")

        self.global_transfo1_extra = RandomGaussianBlur(kernel_size=(9, 9), sigma=(0.1, 2.0), p=1.0)
        self.global_transfo2_extra = RandomGaussianBlur(kernel_size=(9, 9), sigma=(0.1, 2.0), p=0.1)
        self.local_transfo_extra = RandomGaussianBlur(kernel_size=(9, 9), sigma=(0.1, 1.0), p=gaussian_blur_prob)

        # Normalization
        self.normalize = Normalize(mean=mean, std=std)
        self.value_clipping = ValueClipper(min_val=-1008.0, max_val=822.0)
        self.duplicator = ChannelDuplicator(num_channels=model_in_chans)
        self.elastic_transform = RandomElasticTransform(
            sigma=(5.0, 5.0), alpha=(0.75, 0.75), padding_mode="border", p=0.8
        )

        # Assemble transformations
        global_transfo1_list = [
            self.value_clipping,
            self.global_transfo1_extra,
            self.normalize,
        ]
        global_transfo2_list = [
            self.value_clipping,
            self.global_transfo2_extra,
            self.elastic_transform,
            self.normalize,
        ]
        local_transfo_list = [
            self.value_clipping,
            self.local_transfo_extra,
            self.normalize,
        ]

        if model_in_chans == 3:
            global_transfo1_list.append(self.duplicator)
            global_transfo2_list.append(self.duplicator)
            local_transfo_list.append(self.duplicator)
        elif model_in_chans == 1:
            pass
        else:
            raise ValueError(f"Invalid model_in_chans value: {model_in_chans}")

        self.global_transfo1 = nn.Sequential(*global_transfo1_list)
        self.global_transfo2 = nn.Sequential(*global_transfo2_list)
        self.local_transfo = nn.Sequential(*local_transfo_list)

    def __call__(self, data):
        collated_global_crops = data["collated_global_crops"]
        B = collated_global_crops.shape[0] // 2  # Assuming two global crops per sample

        global_crop_1 = collated_global_crops[:B]
        global_crop_2 = collated_global_crops[B:]

        global_crop_1 = self.global_transfo1(global_crop_1)
        global_crop_2 = self.global_transfo2(global_crop_2)

        data["collated_global_crops"] = torch.cat([global_crop_1, global_crop_2], dim=0)

        data["collated_local_crops"] = self.local_transfo(data["collated_local_crops"])

        data["collated_global_crops"] = data["collated_global_crops"].to(self.dtype)
        data["collated_local_crops"] = data["collated_local_crops"].to(self.dtype)

        return data
