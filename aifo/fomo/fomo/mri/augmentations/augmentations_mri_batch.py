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

import logging

import torch
import torch.nn as nn
from fomo.transforms.transforms import (
    ChannelDuplicator,
    ColorJitter,
    MinMaxNormalizer,
    Normalize,
    PercentileClipper,
    RandomGamma,
    RandomGaussianBlur,
    RelativeGaussianNoise,
)

logger = logging.getLogger("dinov2")


class DataAugmentationMRIBatch(object):
    def __init__(
        self,
        mean=(0.1008,),
        std=(0.193391,),
        model_in_chans=3,
        color_jitter_prob=0.8,
        gaussian_blur_prob=0.5,
        gamma_augmentation_prob=0.7,
        gaussian_noise_augmentation_prob=0.3,
        dtype=torch.float32,
    ):
        self.color_jitter_prob = color_jitter_prob
        self.gaussian_blur_prob = gaussian_blur_prob
        self.dtype = dtype
        self.mean = torch.tensor(mean)
        self.std = torch.tensor(std)
        self.device = torch.device("cuda" if torch.cuda.is_available() else "cpu")

        logger.info("###################################")
        logger.info("Using data augmentation parameters:")
        logger.info(f"color_jitter_prob: {color_jitter_prob}")
        logger.info(f"gaussian_blur_prob: {gaussian_blur_prob}")
        logger.info(f"intensity_augmentation_prob: {gamma_augmentation_prob}")
        logger.info(f"gaussian_noise_augmentation_prob: {gaussian_noise_augmentation_prob}")
        logger.info("###################################")

        # Color distortions / blurring
        self.color_jittering = ColorJitter(brightness=0.3, contrast=0.3, p=color_jitter_prob)

        self.global_transfo1_extra = RandomGaussianBlur(kernel_size=(9, 9), sigma=(0.1, 2.0), p=1.0)
        self.global_transfo2_extra = RandomGaussianBlur(kernel_size=(9, 9), sigma=(0.1, 2.0), p=0.1)
        self.local_transfo_extra = RandomGaussianBlur(kernel_size=(9, 9), sigma=(0.1, 1.0), p=gaussian_blur_prob)

        # Normalization
        self.normalize = Normalize(mean=mean, std=std, device=self.device)

        self.initial_normalization = nn.Sequential(
            PercentileClipper(percentile=99.9),
            MinMaxNormalizer(),
        )

        self.duplicator = ChannelDuplicator(num_channels=model_in_chans)

        self.intensity_augmentations = nn.Sequential(
            RandomGamma(gamma_range=(0.9, 1.1), gain_range=(1.0, 1.0), p=gamma_augmentation_prob),
            RelativeGaussianNoise(std_factor=(0.01, 0.2), p=gaussian_noise_augmentation_prob),
        )

        # Assemble transformations
        global_transfo1_list = [
            self.initial_normalization,
            self.intensity_augmentations,
            self.color_jittering,
            self.global_transfo1_extra,
            self.normalize,
        ]
        global_transfo2_list = [
            self.initial_normalization,
            self.intensity_augmentations,
            self.color_jittering,
            self.global_transfo2_extra,
            self.normalize,
        ]
        local_transfo_list = [
            self.initial_normalization,
            self.intensity_augmentations,
            self.color_jittering,
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
