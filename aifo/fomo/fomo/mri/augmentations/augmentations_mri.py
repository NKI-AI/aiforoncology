# Copyright (c) Meta Platforms, Inc. and affiliates.
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

from dinov2.data.transforms import GaussianBlur
from fomo.transforms.transforms import ChannelDuplicator, MinMaxNormalizer, PercentileClipper
from torchvision import transforms

logger = logging.getLogger("dinov2")


class DataAugmentationMRICPU(object):
    def __init__(
        self,
        global_crops_scale,
        local_crops_scale,
        local_crops_number,
        global_crops_size=224,
        local_crops_size=96,
        mean: tuple[float, float, float] = (0.1008),
        std: tuple[float, float, float] = (0.193391),
        model_in_chans: int = 3,
    ):
        self.global_crops_scale = global_crops_scale
        self.local_crops_scale = local_crops_scale
        self.local_crops_number = local_crops_number
        self.global_crops_size = global_crops_size
        self.local_crops_size = local_crops_size

        logger.info("###################################")
        logger.info("Using data augmentation parameters:")
        logger.info(f"global_crops_scale: {global_crops_scale}")
        logger.info(f"local_crops_scale: {local_crops_scale}")
        logger.info(f"local_crops_number: {local_crops_number}")
        logger.info(f"global_crops_size: {global_crops_size}")
        logger.info(f"local_crops_size: {local_crops_size}")
        logger.info("###################################")

        # random resized crop and flip
        self.geometric_augmentation_global = transforms.Compose(
            [
                transforms.RandomResizedCrop(
                    global_crops_size, scale=global_crops_scale, interpolation=transforms.InterpolationMode.BICUBIC
                ),
                transforms.RandomHorizontalFlip(p=0.5),
            ]
        )

        self.geometric_augmentation_local = transforms.Compose(
            [
                transforms.RandomResizedCrop(
                    local_crops_size, scale=local_crops_scale, interpolation=transforms.InterpolationMode.BICUBIC
                ),
                transforms.RandomHorizontalFlip(p=0.5),
            ]
        )

        # color distorsions / blurring
        color_jittering = transforms.Compose(
            [
                transforms.RandomApply(
                    [transforms.ColorJitter(brightness=0.4, contrast=0.4, saturation=0.0, hue=0.0)],
                    p=0.8,
                ),
            ]
        )

        global_transfo1_extra = GaussianBlur(p=1.0)

        global_transfo2_extra = transforms.Compose(
            [
                GaussianBlur(p=0.1),
            ]
        )

        local_transfo_extra = GaussianBlur(p=0.5)

        # normalization
        self.normalize = transforms.Compose(
            [
                transforms.Normalize(mean=mean, std=std),
            ]
        )

        self.initial_normalization = transforms.Compose(
            [
                PercentileClipper(percentile=99.9),
                MinMaxNormalizer(),
            ]
        )

        duplicator = ChannelDuplicator(num_channels=3)

        global_transfo1_list = [self.initial_normalization, color_jittering, global_transfo1_extra, self.normalize]
        global_transfo2_list = [self.initial_normalization, color_jittering, global_transfo2_extra, self.normalize]
        local_transfo_list = [self.initial_normalization, color_jittering, local_transfo_extra, self.normalize]
        if model_in_chans == 3:
            global_transfo1_list.append(duplicator)
            global_transfo2_list.append(duplicator)
            local_transfo_list.append(duplicator)
        elif model_in_chans == 1:
            pass
        else:
            raise ValueError(f"Invalid model_in_chans value: {model_in_chans}")

        self.global_transfo1 = transforms.Compose(global_transfo1_list)
        self.global_transfo2 = transforms.Compose(global_transfo2_list)
        self.local_transfo = transforms.Compose(local_transfo_list)
        # self.global_transfo1 = transforms.Compose([self.initial_normalization, color_jittering, global_transfo1_extra, self.normalize, duplicator])
        # self.global_transfo2 = transforms.Compose([self.initial_normalization, color_jittering, global_transfo2_extra, self.normalize, duplicator])
        # self.local_transfo = transforms.Compose([self.initial_normalization, color_jittering, local_transfo_extra, self.normalize, duplicator])

    def __call__(self, image):
        output = {}

        # global crops:
        im1_base = self.geometric_augmentation_global(image)
        global_crop_1 = self.global_transfo1(im1_base)

        im2_base = self.geometric_augmentation_global(image)
        global_crop_2 = self.global_transfo2(im2_base)

        output["global_crops"] = [global_crop_1, global_crop_2]

        # global crops for teacher:
        output["global_crops_teacher"] = [global_crop_1, global_crop_2]

        # local crops:
        local_crops = [
            self.local_transfo(self.geometric_augmentation_local(image)) for _ in range(self.local_crops_number)
        ]

        output["local_crops"] = local_crops
        output["offsets"] = ()

        return output
