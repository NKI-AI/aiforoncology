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
import math
import os
from functools import partial

import dinov2.distributed as distributed
import torch
import torch.multiprocessing as mp
from dinov2.data import MaskingGenerator, SamplerType, collate_data_and_cast, make_data_loader, make_dataset
from dinov2.fsdp import FSDPCheckpointer
from dinov2.logging import MetricLogger
from dinov2.train.ssl_meta_arch import SSLMetaArch
from dinov2.utils.config import setup
from fvcore.common.checkpoint import PeriodicCheckpointer
from fomo.dataset.vector_dataset.vector_dataset import SharedVectorDataset
from fomo.mri.augmentations.augmentations_mri_batch import DataAugmentationMRIBatch
from fomo.mri.augmentations.augmentations_mri_single import DataAugmentationMRISingle
from fomo.mri.dataset.debug_dataset import DebugMRIDataset
from fomo.utils.training import (
    apply_optim_scheduler,
    build_optimizer,
    build_schedulers,
    compute_gradients_statistics,
    get_args_parser,
    setup_tensorboard,
)
from torch.utils.data import DataLoader

torch.backends.cuda.matmul.allow_tf32 = True  # PyTorch 1.12 sets this to False by default
mp.set_start_method("spawn", force=True)
logger = logging.getLogger("dinov2")


def do_test(cfg, model, iteration):
    new_state_dict = model.teacher.state_dict()

    if distributed.is_main_process():
        iterstring = str(iteration)
        eval_dir = os.path.join(cfg.train.output_dir, "eval", iterstring)
        os.makedirs(eval_dir, exist_ok=True)
        # save teacher checkpoint
        teacher_ckp_path = os.path.join(eval_dir, "teacher_checkpoint.pth")
        torch.save({"teacher": new_state_dict}, teacher_ckp_path)


def do_train(cfg, model, resume=False):
    model.train()
    inputs_dtype = torch.half
    fp16_scaler = model.fp16_scaler  # for mixed precision training

    if cfg.get("track_gradients", False):
        track_gradients = True
    else:
        track_gradients = False

    # setup optimizer

    optimizer = build_optimizer(cfg, model.get_params_groups())
    (
        lr_schedule,
        wd_schedule,
        momentum_schedule,
        teacher_temp_schedule,
        last_layer_lr_schedule,
    ) = build_schedulers(cfg)

    # checkpointer
    checkpointer = FSDPCheckpointer(model, cfg.train.output_dir, optimizer=optimizer, save_to_disk=True)

    start_iter = checkpointer.resume_or_load(cfg.MODEL.WEIGHTS, resume=resume).get("iteration", -1) + 1

    OFFICIAL_EPOCH_LENGTH = cfg.train.OFFICIAL_EPOCH_LENGTH
    max_iter = cfg.optim.epochs * OFFICIAL_EPOCH_LENGTH

    periodic_checkpointer = PeriodicCheckpointer(
        checkpointer,
        period=3 * OFFICIAL_EPOCH_LENGTH,
        max_iter=max_iter,
        max_to_keep=3,
    )

    if distributed.is_main_process():
        writer = setup_tensorboard(cfg.train.output_dir)

    # setup data preprocessing

    img_size = cfg.crops.global_crops_size
    patch_size = cfg.student.patch_size
    n_tokens = (img_size // patch_size) ** 2
    mask_generator = MaskingGenerator(
        input_size=(img_size // patch_size, img_size // patch_size),
        max_num_patches=0.5 * img_size // patch_size * img_size // patch_size,
    )

    data_transform_single = DataAugmentationMRISingle(
        cfg.crops.global_crops_scale,
        cfg.crops.local_crops_scale,
        cfg.crops.local_crops_number,
        global_crops_size=cfg.crops.global_crops_size,
        local_crops_size=cfg.crops.local_crops_size,
    )

    data_transform_batch = DataAugmentationMRIBatch(
        model_in_chans=cfg.student.in_chans if hasattr(cfg.student, "in_chans") else 3,
        dtype=inputs_dtype,
    )

    collate_fn = partial(
        collate_data_and_cast,
        mask_ratio_tuple=cfg.ibot.mask_ratio_min_max,
        mask_probability=cfg.ibot.mask_sample_probability,
        n_tokens=n_tokens,
        mask_generator=mask_generator,
        dtype=inputs_dtype,
    )

    # setup data loader

    if cfg.get("custom_dataset", None) is not None:
        logger.info("Using custom dataset")
        if cfg.custom_dataset.name == "DebugMRIDataset":
            dataset = DebugMRIDataset(
                root_dir=cfg.custom_dataset.root_dir,
                transform=data_transform_single,
                target_transform=lambda _: (),
                orientation=cfg.custom_dataset.orientation,
                mode=cfg.custom_dataset.mode,
            )
            dataloader_type = 0
        elif cfg.custom_dataset.name == "MRIVectorDataset":
            name = cfg.custom_dataset.vector_name
            max_memory_size_bytes = cfg.custom_dataset.queue_size_gb * (1024**3)
            chunk_size_bytes = cfg.custom_dataset.chunk_size_mb * (1024**2)
            dataset = SharedVectorDataset(
                name=name,
                max_memory_size_bytes=max_memory_size_bytes,
                chunk_size_bytes=chunk_size_bytes,
                transforms=data_transform_single,
            )
            dataloader_type = 1
        else:
            raise NotImplementedError("Custom dataset not implemented")
    else:
        dataset = make_dataset(
            dataset_str=cfg.train.dataset_path,
            transform=data_transform_single,
            target_transform=lambda _: (),
        )
        dataloader_type = 0

    # sampler_type = SamplerType.INFINITE
    if dataloader_type == 0:
        sampler_type = SamplerType.SHARDED_INFINITE
        data_loader = make_data_loader(
            dataset=dataset,
            batch_size=cfg.train.batch_size_per_gpu,
            num_workers=cfg.train.num_workers,
            shuffle=True,
            seed=start_iter,  # TODO: Fix this -- cfg.train.seed
            sampler_type=sampler_type,
            sampler_advance=0,  # TODO(qas): fix this -- start_iter * cfg.train.batch_size_per_gpu,
            drop_last=True,
            collate_fn=collate_fn,
        )
    elif dataloader_type == 1:
        data_loader = DataLoader(
            batch_size=cfg.train.batch_size_per_gpu,
            num_workers=cfg.train.num_workers,
            dataset=dataset,
            collate_fn=collate_fn,
            pin_memory=True,
            prefetch_factor=4,
        )
    else:
        raise NotImplementedError("Dataloader type not implemented")

    # training loop

    iteration = start_iter

    logger.info("Starting training from iteration {}".format(start_iter))
    metrics_file = os.path.join(cfg.train.output_dir, "training_metrics.json")
    metric_logger = MetricLogger(delimiter="  ", output_file=metrics_file)
    header = "Training"
    for data in metric_logger.log_every(
        data_loader,
        10,
        header,
        max_iter,
        start_iter,
    ):
        data["collated_global_crops"] = data["collated_global_crops"].to("cuda", non_blocking=False)
        data["collated_local_crops"] = data["collated_local_crops"].to("cuda", non_blocking=False)

        data = data_transform_batch(data)  # batched augmentations

        current_batch_size = data["collated_global_crops"].shape[0] / 2
        if iteration > max_iter:
            return

        # apply schedules

        lr = lr_schedule[iteration]
        wd = wd_schedule[iteration]
        mom = momentum_schedule[iteration]
        teacher_temp = teacher_temp_schedule[iteration]
        last_layer_lr = last_layer_lr_schedule[iteration]
        apply_optim_scheduler(optimizer, lr, wd, last_layer_lr)

        # compute losses

        optimizer.zero_grad(set_to_none=True)
        loss_dict = model.forward_backward(data, teacher_temp=teacher_temp)

        # compute gradients statistics
        if track_gradients and distributed.is_main_process():
            grad_norm, avg_grad, max_grad = compute_gradients_statistics(model.student)
            writer.add_scalar("Gradients/Gradient Norm", grad_norm, iteration)
            writer.add_scalar("Gradients/Average Gradient", avg_grad, iteration)
            writer.add_scalar("Gradients/Maximum Gradient", max_grad, iteration)

        # clip gradients

        if fp16_scaler is not None:
            if cfg.optim.clip_grad:
                fp16_scaler.unscale_(optimizer)
                for v in model.student.values():
                    v.clip_grad_norm_(cfg.optim.clip_grad)
            fp16_scaler.step(optimizer)
            fp16_scaler.update()
        else:
            if cfg.optim.clip_grad:
                for v in model.student.values():
                    v.clip_grad_norm_(cfg.optim.clip_grad)
            optimizer.step()

        # log gradient after clipping
        if track_gradients and distributed.is_main_process():
            grad_norm, avg_grad, max_grad = compute_gradients_statistics(model.student)
            writer.add_scalar("Gradients/Gradient Norm Post Clip", grad_norm, iteration)
            writer.add_scalar("Gradients/Average Gradient Post Clip", avg_grad, iteration)
            writer.add_scalar("Gradients/Maximum Gradient Post Clip", max_grad, iteration)

        # perform teacher EMA update

        model.update_teacher(mom)

        # logging

        if distributed.get_global_size() > 1:
            for v in loss_dict.values():
                torch.distributed.all_reduce(v)
        loss_dict_reduced = {k: v.item() / distributed.get_global_size() for k, v in loss_dict.items()}

        if math.isnan(sum(loss_dict_reduced.values())):
            logger.info("NaN detected")
            raise AssertionError
        losses_reduced = sum(loss for loss in loss_dict_reduced.values())

        metric_logger.update(lr=lr)
        metric_logger.update(wd=wd)
        metric_logger.update(mom=mom)
        metric_logger.update(last_layer_lr=last_layer_lr)
        metric_logger.update(current_batch_size=current_batch_size)
        metric_logger.update(total_loss=losses_reduced, **loss_dict_reduced)

        # TensorBoard logging
        if distributed.is_main_process():
            writer.add_scalar("Learning Rate", lr, iteration)
            writer.add_scalar("Weight Decay", wd, iteration)
            writer.add_scalar("Momentum", mom, iteration)
            writer.add_scalar("Last Layer LR", last_layer_lr, iteration)
            writer.add_scalar("Batch Size", current_batch_size, iteration)
            writer.add_scalar("Total Loss", losses_reduced, iteration)

            for loss_name, loss_value in loss_dict_reduced.items():
                writer.add_scalar(f"Losses/{loss_name}", loss_value, iteration)

        # checkpointing and testing

        if cfg.evaluation.eval_period_iterations > 0 and (iteration + 1) % cfg.evaluation.eval_period_iterations == 0:
            do_test(cfg, model, f"training_{iteration}")
            torch.cuda.synchronize()
        periodic_checkpointer.step(iteration)

        iteration = iteration + 1
    metric_logger.synchronize_between_processes()
    writer.close()
    return {k: meter.global_avg for k, meter in metric_logger.meters.items()}


def main(args):
    cfg = setup(args)

    model = SSLMetaArch(cfg).to(torch.device("cuda"))
    model.prepare_for_distributed_training()

    logger.info("Model:\n{}".format(model))
    if args.eval_only:
        iteration = (
            FSDPCheckpointer(model, save_dir=cfg.train.output_dir)
            .resume_or_load(cfg.MODEL.WEIGHTS, resume=not args.no_resume)
            .get("iteration", -1)
            + 1
        )
        return do_test(cfg, model, f"manual_{iteration}")

    do_train(cfg, model, resume=not args.no_resume)


if __name__ == "__main__":
    args = get_args_parser(add_help=True).parse_args()
    main(args)
