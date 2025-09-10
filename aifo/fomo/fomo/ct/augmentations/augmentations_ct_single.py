# Copyright (c) Meta Platforms, Inc. and affiliates.
#
# This source code is licensed under the Apache License, Version 2.0
# found in the LICENSE file in the root directory of dinov2 in third_party.
#
# Modified by NKI-AI4Oncology, 04-2025

import logging

import torch
import torchvision.transforms as T

logger = logging.getLogger("dinov2")


class DataAugmentationCTSingle2d(object):
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
        logger.info("global_crops_scale: %s", global_crops_scale)
        logger.info("local_crops_scale: %s", local_crops_scale)
        logger.info("local_crops_number: %s", local_crops_number)
        logger.info("global_crops_size: %s", global_crops_size)
        logger.info("local_crops_size: %s", local_crops_size)
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

        if global_crop_1.shape[0] != 1 or global_crop_1.shape[1] != 224 or global_crop_1.shape[2] != 224:
            logger.error("Image found with shape %s and crop became %s", image.shape, global_crop_1.shape)

        if global_crop_2.shape[0] != 1 or global_crop_2.shape[1] != 224 or global_crop_2.shape[2] != 224:
            logger.error("Image found with shape %s and crop became %s", image.shape, global_crop_2.shape)

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
