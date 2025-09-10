# Copyright (c) Meta Platforms, Inc. and affiliates.
#
# This source code is licensed under the Apache License, Version 2.0
# found in the LICENSE file in the root directory of dinov2 in third_party.
#
# Modified by NKI-AI4Oncology, 04-2025

import logging
import math
import os
from functools import partial

import dinov2.distributed as distributed
import torch
import torch.multiprocessing as mp
from dinov2.fsdp import FSDPCheckpointer
from dinov2.logging import MetricLogger
from dinov2.train.ssl_meta_arch import SSLMetaArch
from dinov2.utils.config import setup
from fvcore.common.checkpoint import PeriodicCheckpointer
from fomo.ct.augmentations import DataAugmentationCTBatch3d
from fomo.dataset.vector_dataset.vector_dataset import SharedVectorDataset
from fomo.dataset.prefetch import PinnedDataPrefetcher
from dinov2.models import build_model_from_cfg
from fomo.transforms import collate_data_and_cast_3d, pad_tensor_to_gpu
from fomo.utils.masking import MaskingGenerator3d
from fomo.utils.training import (
    apply_optim_scheduler,
    build_optimizer,
    build_schedulers,
    compute_gradients_statistics,
    get_args_parser,
    setup_tensorboard,
)
from torch.utils.data import DataLoader
import concurrent
from itertools import count

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

    img_size = (
        (cfg.crops.global_crops_size,) * 3
        if isinstance(cfg.crops.global_crops_size, int)
        else tuple(cfg.crops.global_crops_size)
    )
    patch_size = (
        (cfg.student.patch_size,) * 3 if isinstance(cfg.student.patch_size, int) else tuple(cfg.student.patch_size)
    )
    n_depth_patches = img_size[0] // patch_size[0]
    n_height_patches = img_size[1] // patch_size[1]
    n_width_patches = img_size[2] // patch_size[2]
    n_tokens = n_depth_patches * n_height_patches * n_width_patches
    mask_generator = MaskingGenerator3d(
        input_size=(n_depth_patches, n_height_patches, n_width_patches),
        min_num_patches=4,
        max_num_patches=0.5 * n_tokens,
    )

    data_transform_batch = DataAugmentationCTBatch3d(
        cfg.crops.global_crops_scale,
        cfg.crops.local_crops_scale,
        cfg.crops.local_crops_number,
        global_crops_size=cfg.crops.global_crops_size,
        local_crops_size=cfg.crops.local_crops_size,
    )

    collate_fn = partial(
        collate_data_and_cast_3d,
        mask_ratio_tuple=cfg.ibot.mask_ratio_min_max,
        mask_probability=cfg.ibot.mask_sample_probability,
        n_tokens=n_tokens,
        mask_generator=mask_generator,
        dtype=inputs_dtype,
        batch_size=cfg.train.batch_size_per_gpu,
        n_global_crops=2,
        n_local_crops=cfg.crops.local_crops_number,
    )

    pad_tensor_gpu_fn = partial(
        pad_tensor_to_gpu,
        pad_size=cfg.custom_pad.pad_size,
        n_global_crops=2,
        n_local_crops=cfg.crops.local_crops_number,
        global_scale_range=cfg.crops.global_crops_scale,
        local_scale_range=cfg.crops.local_crops_scale,
    )

    # setup data loader

    logger.info("Using custom dataset")
    if cfg.custom_dataset.name == "CTVectorDataset":
        name = cfg.custom_dataset.vector_name
        max_memory_size_bytes = cfg.custom_dataset.queue_size_gb * (1024**3)
        chunk_size_bytes = cfg.custom_dataset.chunk_size_mb * (1024**2)
        dataset = SharedVectorDataset(
            name=name, max_memory_size_bytes=max_memory_size_bytes, chunk_size_bytes=chunk_size_bytes, is_3d=True
        )
    else:
        raise NotImplementedError("Custom dataset not implemented")

    data_loader = DataLoader(
        batch_size=cfg.train.batch_size_per_gpu,
        num_workers=cfg.train.num_workers,
        dataset=dataset,
        collate_fn=collate_fn,
        pin_memory=False,  # We do this in our prefetcher
    )

    # training loop
    iteration = start_iter
    logger.info("Starting training from iteration {}".format(start_iter))
    logger.info("First batch needs to compile the augmentations and may take a while...")
    metrics_file = os.path.join(cfg.train.output_dir, "training_metrics.json")
    metric_logger = MetricLogger(delimiter="  ", output_file=metrics_file)
    header = "Training"

    # -------------------- Data Setup ----------------------------
    copy_stream = torch.cuda.Stream(device=torch.cuda.current_device())
    executor = concurrent.futures.ThreadPoolExecutor(max_workers=1)
    prefetcher = PinnedDataPrefetcher(data_loader, max_prefetch=2)

    data = prefetcher.next()
    # Launch pad_tensor_gpu_fn asynchronously for the first buffer.
    future_buffer = executor.submit(
        partial(pad_tensor_gpu_fn, custom_stream=copy_stream, device=f"cuda:{torch.cuda.current_device()}"),
        data["samples_list"],
    )
    # Wait for the first pad to finish to initialize our double buffer.
    current_buffer = future_buffer.result()

    accumulation_steps = int(cfg.train.accumulation_steps)
    counter = count(
        start=0, step=1 / accumulation_steps
    )  # Purely to leave the din loop intact (we don't plug dataloader into the loop here anymore)

    for accumulation_step in metric_logger.log_every(counter, 10, header, max_iter, start_iter, accumulation_steps):
        # async pad the next batch
        data_next = prefetcher.next()

        future_buffer = executor.submit(
            partial(pad_tensor_gpu_fn, custom_stream=copy_stream, device=f"cuda:{torch.cuda.current_device()}"),
            data_next["samples_list"],
        )

        # batched augmentations for current batch
        data_transform_batch(current_buffer, data)

        current_batch_size = (data["collated_global_crops"].shape[0] / 2) * accumulation_steps
        if iteration > max_iter:
            return

        if float(accumulation_step).is_integer():
            # apply schedules
            lr = lr_schedule[iteration]
            wd = wd_schedule[iteration]
            mom = momentum_schedule[iteration]
            teacher_temp = teacher_temp_schedule[iteration]
            last_layer_lr = last_layer_lr_schedule[iteration]
            apply_optim_scheduler(optimizer, lr, wd, last_layer_lr)

        loss_dict = model.forward_backward(data, teacher_temp=teacher_temp, accumulation_steps=accumulation_steps)

        if float(accumulation_step).is_integer():
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
            loss_dict_reduced = {k: (v.item() / distributed.get_global_size()) for k, v in loss_dict.items()}

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

            if (
                cfg.evaluation.eval_period_iterations > 0
                and (iteration + 1) % cfg.evaluation.eval_period_iterations == 0
            ):
                do_test(cfg, model, f"training_{iteration}")
                torch.cuda.synchronize()
            periodic_checkpointer.step(iteration)
            iteration = iteration + 1

            # reset gradients
            optimizer.zero_grad()

        # Swap buffers
        current_buffer = future_buffer.result()
        data = data_next

    metric_logger.synchronize_between_processes()
    writer.close()
    return {k: meter.global_avg for k, meter in metric_logger.meters.items()}


def main(args):
    cfg = setup(args)

    model = SSLMetaArch(cfg, builder_fn=build_model_from_cfg).to(torch.device("cuda"))
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
