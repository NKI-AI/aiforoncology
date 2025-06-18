# Copyright (c) Meta Platforms, Inc. and affiliates.
#
# This source code is licensed under the Apache License, Version 2.0
# found in the LICENSE file in the root directory of dinov2 in third_party.
#
# Modified by NKI-AI4Oncology, 04-2025

import random
from typing import Callable, Optional, Any

import torch


def collate_data_and_cast_3d(
    samples_list: list[torch.Tensor],
    mask_ratio_tuple: tuple[float, float],
    mask_probability: float,
    dtype: torch.dtype,
    batch_size: int,
    n_global_crops: int,
    n_local_crops: int,
    n_tokens: Optional[int] = None,
    mask_generator: Optional[Callable[[int], list[bool]]] = None,
) -> dict[str, Any]:
    """
    Collate samples and create masks for 3D data.

    Parameters
    ----------
    samples_list : list[torch.Tensor]
        List of sample tensors to be collated
    mask_ratio_tuple : tuple[float, float]
        Start and end values for mask ratio range
    mask_probability : float
        Probability of masking a sample
    dtype : torch.dtype
        Data type for the tensors
    batch_size : int
        Number of samples in a batch
    n_global_crops : int
        Number of global crops per sample
    n_local_crops : int
        Number of local crops per sample
    n_tokens : int, optional
        Number of tokens
    mask_generator : Callable[[int], list[bool]], optional
        Function that generates masks given a number of tokens to mask

    Returns
    -------
    dict[str, Any]
        Dictionary containing:
        - samples_list: original list of samples
        - collated_masks: boolean tensor of masks
        - mask_indices_list: indices of masked elements
        - masks_weight: weights for the masks
        - upperbound: maximum number of masked tokens
        - n_masked_patches: tensor with count of masked patches

    Notes
    -------
    This is an altered version of dinov2's collate_data_and_cast func.
    """
    # returns list tensors, and the mask stuff as usual
    B = batch_size * n_global_crops
    N = n_tokens
    n_samples_masked = int(B * mask_probability)
    probs = torch.linspace(*mask_ratio_tuple, n_samples_masked + 1)
    upperbound = 0
    masks_list = []
    for i in range(0, n_samples_masked):
        prob_min = probs[i]
        prob_max = probs[i + 1]
        masks_list.append(torch.BoolTensor(mask_generator(int(N * random.uniform(prob_min, prob_max)))))
        upperbound += int(N * prob_max)
    for i in range(n_samples_masked, B):
        masks_list.append(torch.BoolTensor(mask_generator(0)))

    random.shuffle(masks_list)

    collated_masks = torch.stack(masks_list).flatten(1)
    mask_indices_list = collated_masks.flatten().nonzero().flatten()

    masks_weight = (1 / collated_masks.sum(-1).clamp(min=1.0)).unsqueeze(-1).expand_as(collated_masks)[collated_masks]

    return {
        "samples_list": samples_list,
        "collated_masks": collated_masks,
        "mask_indices_list": mask_indices_list,
        "masks_weight": masks_weight,
        "upperbound": upperbound,
        "n_masked_patches": torch.full((1,), fill_value=mask_indices_list.shape[0], dtype=torch.long),
    }
