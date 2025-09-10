from typing import Any, Dict, Tuple

import torch
from lightning.pytorch import LightningModule
from torch.optim.optimizer import Optimizer
from torchmetrics import MaxMetric
from torchmetrics.classification.accuracy import Accuracy


class MulticlassClassifierLitModule(LightningModule):
    def __init__(
        self,
        num_classes: int,
        net: torch.nn.Module,
        optimizer: torch.optim.Optimizer,
        loss: torch.nn.Module,
        scheduler: Any,
        compile: bool,
    ) -> None:
        super().__init__()

        self.save_hyperparameters(logger=False, ignore="net")

        self.strict_loading = False

        self.num_classes = num_classes

        self.net = net

        self.loss = loss

        self.train_acc = Accuracy(task="multiclass", num_classes=num_classes)
        self.val_acc = Accuracy(task="multiclass", num_classes=num_classes)
        self.test_acc = Accuracy(task="multiclass", num_classes=num_classes)

        self.val_acc_best = MaxMetric()
        self.signature = None

    def forward(self, x: torch.Tensor) -> dict[str, torch.Tensor]:
        return {"output": self.net(x)}

    def on_train_start(self) -> None:
        self.val_acc.reset()
        self.val_acc_best.reset()

    def model_step(self, batch: Tuple[torch.Tensor, torch.Tensor]) -> Tuple[dict, torch.Tensor, torch.Tensor]:
        img, class_idx = batch

        output_dict = self.forward(img)
        logits = output_dict["output"]
        probs = torch.softmax(logits, dim=1)

        # one-hot encode the class_idx
        loss_input_dict = {"labels": class_idx}

        loss_dict = self.loss(logits, loss_input_dict)
        return loss_dict, probs, class_idx

    def training_step(self, batch: Tuple[torch.Tensor, torch.Tensor], batch_idx: int) -> torch.Tensor:
        loss_dict, probs, targets = self.model_step(batch)

        self.train_acc(probs, targets)
        self.log_dict(loss_dict, "train", on_step=True, on_epoch=True, prog_bar=True)
        self.log("train/acc", self.train_acc, on_step=True, on_epoch=True, prog_bar=True)

        return loss_dict["total_loss"]

    def on_train_epoch_end(self) -> None:
        pass

    def validation_step(self, batch: Tuple[torch.Tensor, int], batch_idx: int) -> dict:
        loss_dict, probs, targets = self.model_step(batch)  # type: ignore

        self.val_acc(probs, targets)
        self.log_dict(loss_dict, "val", on_step=False, on_epoch=True, prog_bar=False)  # type: ignore
        self.log("val/acc", self.val_acc, on_step=False, on_epoch=True, prog_bar=True)

        # if self.signature is None:
        #     img, class_idx = batch
        #     output_dict = self.forward(img)
        #     self.signature = infer_signature(img.cpu().numpy(), output_dict["output"].detach().cpu().numpy())

        return {"probs": probs, "targets": targets}

    def on_validation_epoch_end(self) -> None:
        acc = self.val_acc.compute()  # get current val acc
        self.val_acc_best(acc)  # update best so far val acc

        self.log("val/acc_best", self.val_acc_best.compute(), prog_bar=True)

        # necessary for proper epoch graphs in mlflow somehow. stupid
        self.log("current_epoch", self.current_epoch, on_epoch=True)

    def test_step(self, batch: Tuple[torch.Tensor, int], batch_idx: int) -> dict:
        loss_dict, probs, targets = self.model_step(batch)  # type: ignore

        preds = torch.argmax(probs, dim=1)
        self.test_acc(preds, targets)
        self.log_dict(loss_dict, "test", on_step=False, on_epoch=True, prog_bar=False)  # type: ignore
        self.log("test/acc", self.test_acc, on_step=False, on_epoch=True, prog_bar=True)

        return {"probs": probs, "targets": targets}

    def log_dict(self, dict_to_log: dict, prefix: str, on_step: bool, on_epoch: bool, prog_bar: bool = False) -> None:
        for key, value in dict_to_log.items():
            self.log(f"{prefix}/{key}", value, on_step=on_step, on_epoch=on_epoch, prog_bar=prog_bar)

    def on_test_epoch_end(self) -> None:
        pass

    def on_before_optimizer_step(self, optimizer: Optimizer) -> None:
        pass

    def on_load_checkpoint(self, checkpoint: Dict[str, Any]) -> None:
        # adjust the learning rate
        # lr_ckpt = checkpoint["optimizer_states"][0]["param_groups"][0]["lr"]
        # checkpoint["optimizer_states"][0]["param_groups"][0]["lr"] = lr_ckpt * 0.01
        # checkpoint["optimizer_states"][0]["param_groups"][0]["weight_decay"] = 0.001
        return super().on_load_checkpoint(checkpoint)

    def setup(self, stage: str) -> None:
        """
        :param stage: Either `"fit"`, `"validate"`, `"test"`, or `"predict"`.
        """
        if self.hparams.compile and stage == "fit":  # type: ignore
            self.net = torch.compile(self.net)

    def configure_optimizers(self) -> Dict[str, Any]:
        optimizer = self.hparams.optimizer(params=self.parameters())  # type: ignore

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
