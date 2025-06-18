from typing import Any

import numpy as np
import relplot as rp
import torch
from lightning.pytorch import Callback, LightningModule, Trainer


class CalibrationMetricsCallback(Callback):
    def __init__(self):
        """Initializes the callback with a path to save reliability diagrams.

        Parameters:
        - save_path: str, the path where the plots will be saved.
        """
        super().__init__()
        self.probs = []
        self.targets = []

    def on_validation_batch_end(
        self,
        trainer: Trainer,
        pl_module: LightningModule,
        outputs: Any,
        batch: Any,
        batch_idx: int,
    ) -> None:
        """Store the probabilities and targets for each batch during validation."""
        self.probs.append(outputs["probs"].detach().cpu())
        self.targets.append(outputs["targets"].detach().cpu())

    def on_validation_epoch_end(self, trainer: Trainer, pl_module: LightningModule) -> None:
        probs = torch.cat(self.probs, dim=0)
        targets = torch.cat(self.targets, dim=0)

        self.log_metrics(probs, targets, "val", trainer, pl_module)

        self.probs = []
        self.targets = []

    def on_test_batch_end(
        self,
        trainer: Trainer,
        pl_module: LightningModule,
        outputs: Any,
        batch: Any,
        batch_idx: int,
    ) -> None:
        """Store the probabilities and targets for each batch during test."""
        self.probs.append(outputs["probs"].detach().cpu())
        self.targets.append(outputs["targets"].detach().cpu())

    def on_test_epoch_end(self, trainer: Trainer, pl_module: LightningModule) -> None:
        probs = torch.cat(self.probs, dim=0)
        targets = torch.cat(self.targets, dim=0)

        self.log_metrics(probs, targets, "test", trainer, pl_module)

        self.probs = []
        self.targets = []

    def log_metrics(self, probs, targets, phase, trainer: Trainer, pl_module: LightningModule) -> None:
        conf, acc = rp.multiclass_logits_to_confidences(probs, targets, probs=True)
        sm_ece = rp.smECE(conf, acc)
        b10_ece = rp.metrics.binnedECE(conf, acc, nbins=10)
        b20_ece = rp.metrics.binnedECE(conf, acc, nbins=20)
        try:
            int_ece = rp.metrics.intCE_rand(conf, acc)
        except Exception:
            int_ece = np.nan

        try:
            laplace_ece = rp.metrics.laplace_calibration_approx(conf, acc)
        except Exception:
            laplace_ece = np.nan

        pl_module.log(f"{phase}/cal_smECE", sm_ece, on_step=False, on_epoch=True, prog_bar=False)  # type: ignore
        pl_module.log(f"{phase}/cal_b10ECE", b10_ece, on_step=False, on_epoch=True, prog_bar=False)
        pl_module.log(f"{phase}/cal_b20ECE", b20_ece, on_step=False, on_epoch=True, prog_bar=False)
        pl_module.log(f"{phase}/cal_intECE", int_ece, on_step=False, on_epoch=True, prog_bar=False)
        pl_module.log(f"{phase}/cal_lapECE", laplace_ece, on_step=False, on_epoch=True, prog_bar=False)
