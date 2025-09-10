from typing import Any, Dict, Tuple

import numpy as np
import torch
from lightning import LightningModule
from sklearn.metrics import roc_auc_score
from sklearn.preprocessing import label_binarize
from torchmetrics import MaxMetric, MeanMetric
from torchmetrics.classification.accuracy import Accuracy


class ClassifierLitModule(LightningModule):
    def __init__(
        self,
        net: torch.nn.Module,
        optimizer: torch.optim.Optimizer,
        compile: bool,
        scheduler: torch.optim.lr_scheduler = None,  # type: ignore
    ) -> None:
        super().__init__()

        self.save_hyperparameters(logger=False)

        self.net = net

        self.criterion = torch.nn.CrossEntropyLoss()

        self.num_classes = 10
        self.train_acc = Accuracy(task="multiclass", num_classes=self.num_classes)
        self.val_acc = Accuracy(task="multiclass", num_classes=self.num_classes)
        self.test_acc = Accuracy(task="multiclass", num_classes=self.num_classes)

        self.train_loss = MeanMetric()
        self.val_loss = MeanMetric()
        self.test_loss = MeanMetric()

        self.val_acc_best = MaxMetric()

    def forward(self, x: torch.Tensor) -> torch.Tensor:
        return self.net(x)

    def on_train_start(self) -> None:
        self.val_loss.reset()
        self.val_acc.reset()
        self.val_acc_best.reset()

    def on_train_epoch_start(self) -> None:
        self.logits = []
        self.targets = []

    def model_step(
        self, batch: Dict[str, torch.Tensor]
    ) -> Tuple[torch.Tensor, torch.Tensor, torch.Tensor, torch.Tensor]:
        scan = batch["scan"]
        mask = batch["mask"]
        y = batch["target"]
        x = torch.cat([scan, mask], dim=1)
        logits = self.forward(x)
        loss = self.criterion(logits, y)
        preds = torch.argmax(logits, dim=1)
        return loss, preds, logits, y

    def training_step(self, batch: Dict[str, torch.Tensor], batch_idx: int) -> torch.Tensor:
        loss, preds, logits, targets = self.model_step(batch)

        # update and log metrics
        self.train_loss(loss)
        self.train_acc(preds, targets)
        self.log("train/loss", self.train_loss, on_step=False, on_epoch=True, prog_bar=True)
        self.log("train/acc", self.train_acc, on_step=False, on_epoch=True, prog_bar=True)

        self.logits.append(logits.detach())  # type: ignore
        self.targets.append(targets.detach())  # type: ignore

        # return loss or backpropagation will fail
        return loss

    def on_train_epoch_end(self) -> None:
        # check if self.logits is a list
        if not isinstance(self.logits, list):
            return

        self.logits = torch.cat(self.logits).cpu()  # type: ignore
        probs = torch.nn.functional.softmax(self.logits, dim=1).numpy()
        self.targets = torch.cat(self.targets).cpu().numpy()  # type: ignore

        if self.num_classes > 2:
            targets_bin = label_binarize(self.targets, classes=range(self.num_classes))
        # if there are only two classes, LabelBinarizer will not return a one-hot encoding
        elif self.num_classes == 2:
            targets_bin = np.array([[1, 0] if t == 0 else [0, 1] for t in self.targets])

        auc_scores = {}
        for i in range(self.num_classes):
            if targets_bin[:, i].sum() == 0:
                continue
            auc = roc_auc_score(targets_bin[:, i], probs[:, i])
            auc_scores[i] = f"{auc:.4f}"

            self.log(f"train/auc_{i}", auc, sync_dist=True, prog_bar=True)

        pass

    def on_validation_epoch_start(self) -> None:
        self.logits = []
        self.targets = []

    def validation_step(self, batch: Dict[str, torch.Tensor], batch_idx: int) -> None:
        loss, preds, logits, targets = self.model_step(batch)

        self.val_loss(loss)
        self.val_acc(preds, targets)
        self.log("val/loss", self.val_loss, on_step=False, on_epoch=True, prog_bar=True)
        self.log("val/acc", self.val_acc, on_step=False, on_epoch=True, prog_bar=True)

        # add logits and targets to compute AUC later
        self.logits.append(logits)  # type: ignore
        self.targets.append(targets)  # type: ignore

    def on_validation_epoch_end(self) -> None:
        acc = self.val_acc.compute()  # get current val acc
        self.val_acc_best(acc)  # update best so far val acc
        self.log("val/acc_best", self.val_acc_best.compute(), sync_dist=True, prog_bar=True)

        self.logits = torch.cat(self.logits).cpu()  # type: ignore
        probs = torch.nn.functional.softmax(self.logits, dim=1).numpy()
        self.targets = torch.cat(self.targets).cpu().numpy()  # type: ignore

        if self.num_classes > 2:
            targets_bin = label_binarize(self.targets, classes=range(self.num_classes))
        # if there are only two classes, LabelBinarizer will not return a one-hot encoding
        elif self.num_classes == 2:
            targets_bin = np.array([[1, 0] if t == 0 else [0, 1] for t in self.targets])

        auc_scores = {}
        for i in range(self.num_classes):
            if targets_bin[:, i].sum() == 0:  # type: ignore
                continue
            auc = roc_auc_score(targets_bin[:, i], probs[:, i])  # type: ignore
            auc_scores[i] = f"{auc:.4f}"

            self.log(f"val/auc_{i}", auc, sync_dist=True, prog_bar=True)  # type: ignore

        print(auc_scores)

    def test_step(self, batch: Dict[str, torch.Tensor], batch_idx: int) -> None:
        loss, preds, logits, targets = self.model_step(batch)

        self.test_loss(loss)
        self.test_acc(preds, targets)
        self.log("test/loss", self.test_loss, on_step=False, on_epoch=True, prog_bar=True)
        self.log("test/acc", self.test_acc, on_step=False, on_epoch=True, prog_bar=True)

    def on_test_epoch_end(self) -> None:
        pass

    def setup(self, stage: str) -> None:
        if self.hparams.compile and stage == "fit":  # type: ignore
            self.net = torch.compile(self.net)

    def configure_optimizers(self) -> Dict[str, Any]:
        optimizer = self.hparams.optimizer(params=self.trainer.model.parameters())  # type: ignore
        if self.hparams.scheduler is not None:  # type: ignore
            scheduler = self.hparams.scheduler(optimizer=optimizer)  # type: ignore
            return {
                "optimizer": optimizer,
                "lr_scheduler": {
                    "scheduler": scheduler,
                    "monitor": "val/loss",
                    "interval": "epoch",
                    "frequency": 1,
                },
            }
        return {"optimizer": optimizer}


if __name__ == "__main__":
    pass
