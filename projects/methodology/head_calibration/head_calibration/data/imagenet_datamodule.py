from typing import Optional

import torch
from lightning.pytorch import LightningDataModule
from torch.utils.data import DataLoader, Dataset, random_split
from torchvision import transforms as tf
from torchvision.datasets import ImageNet


class ImagenetDataModule(LightningDataModule):
    def __init__(
        self,
        data_dir: str = "data",
        batch_size: int = 10,
        num_workers: int = 0,
        finetune: bool = False,
        freecalibrate: bool = False,
        img_size: int = 128,
        *args,
        **kwargs,
    ):
        super().__init__()

        self.save_hyperparameters(logger=False)

        self.data_dir = data_dir
        self.batch_size = batch_size
        self.num_workers = num_workers

        self.finetune = finetune
        self.freecalibrate = freecalibrate

        self.data_train: Optional[Dataset] = None
        self.data_val: Optional[Dataset] = None
        self.data_test: Optional[Dataset] = None

        # self.train_transforms = tf.Compose(
        #     [
        #         tf.RandomResizedCrop(img_size),
        #         # tf.RandomResizedCrop(224),
        #         tf.RandomHorizontalFlip(),
        #         tf.ToTensor(),
        #         tf.Normalize((0.5, 0.5, 0.5), (0.5, 0.5, 0.5)),
        #     ]
        # )

        self.train_transforms = tf.Compose(
            [
                tf.RandomResizedCrop(img_size),
                tf.RandomHorizontalFlip(),
                tf.ColorJitter(brightness=0.4, contrast=0.4, saturation=0.4, hue=0.2),
                tf.RandomRotation(15),
                tf.ToTensor(),
                # tf.Normalize((0.5, 0.5, 0.5), (0.5, 0.5, 0.5)),
                tf.Normalize((0.485, 0.456, 0.406), (0.229, 0.224, 0.225)),
            ]
        )

        print(f"Using train transforms: {self.train_transforms}")

        self.val_transforms = tf.Compose(
            [
                tf.Resize(img_size + 16),
                tf.CenterCrop(img_size),
                tf.ToTensor(),
                # tf.Normalize((0.5, 0.5, 0.5), (0.5, 0.5, 0.5)),
                tf.Normalize((0.485, 0.456, 0.406), (0.229, 0.224, 0.225)),
            ]
        )

    def prepare_data(self) -> None:
        pass

    def setup(self, stage: Optional[str] = None) -> None:
        if not self.data_train and not self.data_val and not self.data_test:
            trainset = ImageNet(root=self.data_dir, split="train", transform=self.train_transforms)
            trainset_no_transform = ImageNet(root=self.data_dir, split="train", transform=self.val_transforms)
            testset = ImageNet(root=self.data_dir, split="val", transform=self.val_transforms)

            self.data_train, self.data_finetune, _ = random_split(
                trainset,  # type: ignore
                [0.9, 0.0, 0.1],
                # [0.1, 0.2, 0.7],
                generator=torch.Generator().manual_seed(42),
            )

            _, _, self.data_val = random_split(
                trainset_no_transform,
                [0.7, 0.2, 0.1],
                generator=torch.Generator().manual_seed(42),
            )
            self.data_test = testset

            if self.finetune:
                if self.freecalibrate:
                    print("Using freecalibration set (same as training)")
                else:
                    print("Using finetuning set")
                    self.data_train = self.data_finetune

    def train_dataloader(self) -> DataLoader:
        return DataLoader(
            self.data_train,  # type: ignore
            batch_size=self.batch_size,
            num_workers=self.num_workers,
            shuffle=True,
            persistent_workers=self.num_workers > 0,
        )

    def val_dataloader(self) -> DataLoader:
        return DataLoader(
            self.data_val,  # type: ignore
            batch_size=self.batch_size,
            num_workers=self.num_workers,
            persistent_workers=self.num_workers > 0,
        )

    def test_dataloader(self) -> DataLoader:
        return DataLoader(
            self.data_test,  # type: ignore
            batch_size=self.batch_size,
            num_workers=self.num_workers,
            persistent_workers=self.num_workers > 0,
        )
