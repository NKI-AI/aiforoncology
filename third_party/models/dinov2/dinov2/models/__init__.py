# Copyright (c) Meta Platforms, Inc. and affiliates.
#
# This source code is licensed under the Apache License, Version 2.0
# found in the LICENSE file in the root directory of this source tree.

import logging

import torch

from . import vision_transformer as vits
from dinov2.layers.patch_embed import PatchEmbed

logger = logging.getLogger("dinov2")


def build_model(args, only_teacher=False, img_size=224, custom_pretrained_path: str | None = None):
    args.arch = args.arch.removesuffix("_memeff")
    if "vit" in args.arch:
        vit_kwargs = dict(
            img_size=img_size,
            patch_size=args.patch_size,
            init_values=args.layerscale,
            embed_layer=args.embed_layer if hasattr(args, "embed_layer") else PatchEmbed,
            ffn_layer=args.ffn_layer,
            block_chunks=args.block_chunks,
            qkv_bias=args.qkv_bias,
            proj_bias=args.proj_bias,
            ffn_bias=args.ffn_bias,
            num_register_tokens=args.num_register_tokens,
            interpolate_offset=args.interpolate_offset,
            interpolate_antialias=args.interpolate_antialias,
            in_chans=args.in_chans if hasattr(args, "in_chans") else 3,
        )
        teacher = vits.__dict__[args.arch](**vit_kwargs)
        if only_teacher:
            return teacher, teacher.embed_dim
        student = vits.__dict__[args.arch](
            **vit_kwargs,
            drop_path_rate=args.drop_path_rate,
            drop_path_uniform=args.drop_path_uniform,
        )
        embed_dim = student.embed_dim
        if custom_pretrained_path:
            state_dict = torch.load(custom_pretrained_path)
            # Load weights from a pretrained pth
            try:
                student.load_state_dict(state_dict)
                teacher.load_state_dict(state_dict)
                logger.info(f"Loaded pretrained weights from {custom_pretrained_path}")
            except RuntimeError:
                logger.warning(f"Failed to load pretrained weights from {custom_pretrained_path}")
                logger.info(
                    "Attempting to remove pos_embed and retry"
                )  # The pos_embed in pretraine dinov2 models is not compatible with the current model
                state_dict["pos_embed"] = student.pos_embed
                student.load_state_dict(state_dict)
                teacher.load_state_dict(state_dict)
                logger.info(f"Loaded pretrained weights from {custom_pretrained_path}")
    return student, teacher, embed_dim


def build_model_from_cfg(cfg, only_teacher=False):
    return build_model(
        cfg.student,
        only_teacher=only_teacher,
        img_size=cfg.crops.global_crops_size,
        custom_pretrained_path=cfg.get("custom_pretrained_path", None),
    )
