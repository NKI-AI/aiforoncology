import json
from typing import Any

from lightning.pytorch import Callback, LightningModule, Trainer
from matplotlib import pyplot as plt


class DifficultSamplesCallback(Callback):
    def __init__(self):
        """Initializes the callback"""
        super().__init__()

    def on_validation_batch_end(
        self,
        trainer: Trainer,
        pl_module: LightningModule,
        outputs: Any,
        batch: Any,
        batch_idx: int,
    ) -> None:
        # outputs: {"probs": probs, "targets": targets}
        # batch: [images, targets]

        # select all samples where the model's prediction is < 0.5 and > 0.3
        probs_inrange = ((outputs["probs"] > 0.3) & (outputs["probs"] < 0.7)).any(dim=1).detach().cpu()
        targets = outputs["targets"][probs_inrange].detach().cpu()
        probsf = outputs["probs"][probs_inrange].detach().cpu()
        imgs = batch[0][probs_inrange].detach().cpu().numpy()
        imgs = imgs.transpose(0, 2, 3, 1)
        imgs = (imgs - imgs.min()) / (imgs.max() - imgs.min())

        with open("data/imagenet_class_index.json", "r") as json_file:
            class_labels = json.load(json_file)

        # get the top 3 predictions for each image
        top3 = probsf.topk(3, dim=1)
        top3_labels = top3.indices

        # make a column of images with figsize a multiple of 20
        fig, ax = plt.subplots(len(imgs), 1, figsize=(5, 5 * len(imgs)))

        for i, img in enumerate(imgs):
            ax[i].imshow(img)
            ax[i].set_title(
                f"Target: {class_labels[str(targets[i].item())][1]}\n {', '.join([class_labels[str(label.item())][1] for label in top3_labels[i]])}\n {', '.join([f'{prob:.2f}' for prob in top3.values[i]])}"
            )
            ax[i].axis("off")

        # save the figure
        fig.tight_layout()
        plt.savefig(f"/home/j.brunekreef/tmp/difficult_samples_{batch_idx}.png")

    def on_validation_epoch_end(self, trainer: Trainer, pl_module: LightningModule) -> None:
        pass

    def on_test_batch_end(
        self,
        trainer: Trainer,
        pl_module: LightningModule,
        outputs: Any,
        batch: Any,
        batch_idx: int,
    ) -> None:
        pass

    def on_test_epoch_end(self, trainer: Trainer, pl_module: LightningModule) -> None:
        pass

    def log_metrics(self, probs, targets, phase, trainer: Trainer, pl_module: LightningModule) -> None:
        pass
