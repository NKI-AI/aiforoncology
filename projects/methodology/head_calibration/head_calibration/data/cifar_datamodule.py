from typing import Optional

import torch
from lightning.pytorch import LightningDataModule
from torch.utils.data import DataLoader, Dataset, random_split
from torchvision import transforms as tf
from torchvision.datasets import CIFAR10, CIFAR100


class CIFARDataModule(LightningDataModule):
    def __init__(
        self,
        data_dir: str = "data",
        batch_size: int = 10,
        num_workers: int = 0,
        cifar_version: int = 10,
        finetune: bool = False,
        freecalibrate: bool = False,
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

        if cifar_version == 10:
            self.dataset_prototype = CIFAR10
        elif cifar_version == 100:
            self.dataset_prototype = CIFAR100
        else:
            raise ValueError("CIFAR version must be 10 or 100")

        self.train_transforms = tf.Compose(
            [
                tf.RandomCrop(32, padding=4),
                tf.RandomHorizontalFlip(),
                # tf.ColorJitter(brightness=0.1, contrast=0.1, saturation=0.1, hue=0.1),
                # tf.RandomRotation(15),
                tf.ToTensor(),
                tf.Normalize((0.4914, 0.4822, 0.4465), (0.2023, 0.1994, 0.2010)),
            ]
        )

        self.val_transforms = tf.Compose(
            [tf.ToTensor(), tf.Normalize((0.4914, 0.4822, 0.4465), (0.2023, 0.1994, 0.2010))]
        )

    def prepare_data(self):
        self.dataset_prototype(root=self.data_dir, train=True, download=True)
        self.dataset_prototype(root=self.data_dir, train=False, download=True)

    def setup(self, stage: Optional[str] = None) -> None:
        if not self.data_train and not self.data_val and not self.data_test:
            trainset = self.dataset_prototype(
                root=self.data_dir, train=True, download=False, transform=self.train_transforms
            )
            trainset_no_transform = self.dataset_prototype(
                root=self.data_dir, train=True, download=False, transform=self.val_transforms
            )

            testset = self.dataset_prototype(
                root=self.data_dir, train=False, download=False, transform=self.val_transforms
            )

            self.data_train, data_finetune, _ = random_split(
                trainset,  # type: ignore
                [0.7, 0.2, 0.1],
                generator=torch.Generator().manual_seed(42),
            )

            _, _, self.data_val = random_split(
                trainset_no_transform,  # type: ignore
                [0.7, 0.2, 0.1],
                generator=torch.Generator().manual_seed(42),
            )

            self.data_test = testset

            if self.finetune:
                if self.freecalibrate:
                    print("Using freecalibration set (same as training).")
                else:
                    print(f"Using finetuning set ({len(data_finetune)} samples).")
                    self.data_train = data_finetune

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
