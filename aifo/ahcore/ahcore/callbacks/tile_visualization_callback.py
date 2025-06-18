import abc
import pathlib
from pathlib import Path
from typing import Any

import numpy as np
import lightning.pytorch as pl
from matplotlib import pyplot as plt
from lightning.pytorch import Callback
from ahcore.utils.io import get_logger

logger = get_logger(__name__)


class TileVisualizationCallback(abc.ABC, Callback):
    def __init__(
        self,
        dump_dir: pathlib.Path | str,
    ) -> None:
        self._dump_dir = Path(dump_dir)
        self._image_dir = self._dump_dir / "images"
        self._image_dir.mkdir(parents=True, exist_ok=True)

    @property
    def dump_dir(self) -> pathlib.Path:
        return self._dump_dir

    def setup(self, trainer: pl.Trainer, pl_module: pl.LightningModule, stage: str) -> None:
        if trainer.global_rank != 0:
            return

    def on_train_batch_end(
        self,
        trainer: "pl.Trainer",
        pl_module: "pl.LightningModule",
        outputs: Any,
        batch: Any,
        batch_idx: int,
        dataloader_idx: int = 0,
    ) -> None:
        # self._batch_end(trainer, pl_module, outputs, batch, batch_idx, "fit", dataloader_idx)
        pass

    def on_validation_batch_end(
        self,
        trainer: "pl.Trainer",
        pl_module: "pl.LightningModule",
        outputs: Any,
        batch: Any,
        batch_idx: int,
        dataloader_idx: int = 0,
    ) -> None:
        self._batch_end(trainer, pl_module, outputs, batch, batch_idx, "validate", dataloader_idx)

    def _batch_end(
        self,
        trainer: "pl.Trainer",
        pl_module: "pl.LightningModule",
        outputs: Any,
        batch: Any,
        batch_idx: int,
        stage: str,
        dataloader_idx: int = 0,
    ) -> None:
        # if stage == "fit":
        #     return

        if trainer.global_rank != 0:
            return

        current_epoch = trainer.current_epoch

        if current_epoch % 5 != 0 or batch_idx > 10:
            return

        batch_size = batch["image"].shape[0]

        factor = 16
        fig, axs = plt.subplots(batch_size, 3, figsize=(3 * factor, batch_size * factor))

        for i in range(batch_size):
            # the image is a shape (3, 1024, 1024) tensor
            image = batch["image"][i].cpu().numpy().transpose(1, 2, 0)
            # clip the image to [0, 1]
            image = np.clip(image, 0, 1)
            # the mask is a one-hot encoded tensor of shape (3, 1024, 1024)
            target_mask = batch["target"][i].cpu().numpy().transpose(1, 2, 0)
            lympho_mask = target_mask[:, :, 1] > 0.5
            mono_mask = target_mask[:, :, 2] > 0.5

            # the prediction is a tensor of probabilities of shape (3, 1024, 1024)
            prediction_mask = outputs["prediction"][i].detach().cpu().numpy().transpose(1, 2, 0)
            # take the argmax to get the class with the highest probability
            prediction_mask = np.argmax(prediction_mask, axis=-1)

            axs[i, 0].imshow(image)
            axs[i, 1].imshow(image)
            axs[i, 2].imshow(image)

            lympho_rgba = np.zeros((*image.shape[:2], 4))
            lympho_rgba[lympho_mask] = [1.0, 0, 0, 0.4]
            axs[i, 1].imshow(lympho_rgba)

            mono_rgba = np.zeros((*image.shape[:2], 4))
            mono_rgba[mono_mask] = [0, 1.0, 0, 0.4]
            axs[i, 1].imshow(mono_rgba)

            prediction_rgba = np.zeros((*image.shape[:2], 4))
            prediction_rgba[prediction_mask == 1] = [1.0, 0, 0, 0.4]
            prediction_rgba[prediction_mask == 2] = [0, 1.0, 0, 0.4]
            axs[i, 2].imshow(prediction_rgba)

            axs[i, 0].axis("off")
            axs[i, 1].axis("off")
            axs[i, 2].axis("off")

        plt.tight_layout()
        plt.savefig(self._image_dir / f"{stage}_{current_epoch}_{batch_idx}.png")

        # close the figure to avoid memory leak
        plt.close(fig)
