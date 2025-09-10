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
import torchvision.transforms as T

logger = logging.getLogger("dinov2")


class DataAugmentationMRISingle(object):
    def __init__(
        self,
        global_crops_scale,
        local_crops_scale,
        local_crops_number,
        global_crops_size=224,
        local_crops_size=96,
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

        # Geometric augmentations
        self.geometric_augmentation_global = T.RandomResizedCrop(
            size=(global_crops_size, global_crops_size),
            scale=global_crops_scale,
            interpolation=T.InterpolationMode.BILINEAR,
        )
        self.geometric_augmentation_local = T.RandomResizedCrop(
            size=(local_crops_size, local_crops_size),
            scale=local_crops_scale,
            interpolation=T.InterpolationMode.BILINEAR,
        )

    def __call__(self, image):
        output = {}

        im1_base = self.geometric_augmentation_global(image)
        global_crop_1 = im1_base.squeeze(1)  # Squeeze the channel dimension

        im2_base = self.geometric_augmentation_global(image)
        global_crop_2 = im2_base.squeeze(1)  # Squeeze the channel dimension

        output["global_crops"] = [global_crop_1, global_crop_2]

        output["global_crops_teacher"] = [global_crop_1, global_crop_2]

        local_crops = []
        for _ in range(self.local_crops_number):
            local_crop = self.geometric_augmentation_local(image)
            local_crop = local_crop.squeeze(1)  # Squeeze the channel dimension
            local_crops.append(local_crop)

        output["local_crops"] = torch.stack(local_crops)
        output["offsets"] = ()

        return output
