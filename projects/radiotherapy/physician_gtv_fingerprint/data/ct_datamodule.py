from pathlib import Path
from typing import Any, Optional

import monai.transforms as tf
import numpy as np
from lightning import LightningDataModule
from research.physician_gtv_fingerprint.data.components.ct_dataset import CTDataset
from torch.utils.data import DataLoader, Dataset


class CTDataModule(LightningDataModule):
    def __init__(
        self,
        data_dir: str = "data/",
        split_name: str = "default",
        split_label: str = "0",
        dims: tuple[int, int, int] = (128, 128, 64),
        batch_size: int = 8,
        num_workers: int = 0,
        pin_memory: bool = False,
    ) -> None:
        super().__init__()

        # this line allows to access init params with 'self.hparams' attribute
        # also ensures init params will be stored in ckpt
        self.save_hyperparameters(logger=False)

        self.data_dir = Path(data_dir)
        self.split_name = split_name
        self.split_label = split_label
        self.dims = dims
        self.batch_size = batch_size

        self.data_train: Optional[Dataset] = None
        self.data_val: Optional[Dataset] = None
        self.data_test: Optional[Dataset] = None

        pad_expand = (20, 20, 20)
        pad_dims = tuple(d + p for d, p in zip(self.dims, pad_expand))
        self.aug_transforms = tf.Compose(
            [
                tf.SpatialPadd(keys=["scan", "mask"], spatial_size=pad_dims, mode="constant"),
                tf.CenterSpatialCropd(keys=["scan", "mask"], roi_size=pad_dims),
                tf.RandFlipd(keys=["scan", "mask"], spatial_axis=0, prob=0.5),
                tf.RandFlipd(keys=["scan", "mask"], spatial_axis=1, prob=0.5),
                tf.RandFlipd(keys=["scan", "mask"], spatial_axis=2, prob=0.5),
                # tf.RandAffined(
                #     keys=["scan", "mask"], prob=0.5, spatial_size=self.dims, rotate_range=(10, 10, 10), translate_range=pad_expand
                # ),
                tf.RandAffined(keys=["scan", "mask"], prob=0.5, spatial_size=self.dims, translate_range=pad_expand),
                # tf.RandGaussianNoised(keys="scan", prob=0.5, std=0.05),
                # tf.RandScaleIntensityd(keys="scan", factors=0.1, prob=0.5),
                # tf.RandAffined(keys=["scan", "mask"], prob=0.5, translate_range=pad_expand),
                tf.CenterSpatialCropd(keys=["scan", "mask"], roi_size=self.dims),
                tf.Lambdad(func=lambda x: x.astype(np.float32), keys="scan"),
                tf.Lambdad(func=lambda x: (x > 0.5).astype(np.float32), keys="mask"),
                tf.Lambdad(func=lambda x: np.int64(x), keys="target"),
                tf.ToTensord(keys=["scan", "mask", "target"]),
            ]
        )

        self.prepr_transforms = tf.Compose(
            [
                tf.SpatialPadd(keys=["scan", "mask"], spatial_size=pad_dims, mode="constant"),
                tf.CenterSpatialCropd(keys=["scan", "mask"], roi_size=self.dims),
                tf.Lambdad(func=lambda x: x.astype(np.float32), keys=["scan", "mask"]),
                tf.Lambdad(func=lambda x: np.int64(x), keys="target"),
                tf.ToTensord(keys=["scan", "mask", "target"]),
            ]
        )

        self.batch_size = batch_size

    @property
    def num_classes(self) -> int:
        return 10

    def setup(self, stage: Optional[str] = None) -> None:
        self.data_train = CTDataset(
            data_dir=self.data_dir,
            splits_fn=f"{self.split_name}/{self.split_label}.json",
            split="train",
            transforms=self.aug_transforms,
        )

        self.data_val = CTDataset(
            data_dir=self.data_dir,
            splits_fn=f"{self.split_name}/{self.split_label}.json",
            split="val",
            transforms=self.prepr_transforms,
        )

        self.data_test = CTDataset(
            data_dir=self.data_dir,
            splits_fn=f"{self.split_name}/{self.split_label}.json",
            split="test",
            transforms=self.prepr_transforms,
        )

    def train_dataloader(self) -> DataLoader[Any]:
        return DataLoader(
            dataset=self.data_train,  # type: ignore
            batch_size=self.batch_size,
            num_workers=self.hparams.num_workers,  # type: ignore
            pin_memory=self.hparams.pin_memory,  # type: ignore
            shuffle=True,
        )

    def val_dataloader(self) -> DataLoader[Any]:
        return DataLoader(
            dataset=self.data_val,  # type: ignore
            batch_size=self.batch_size,
            num_workers=self.hparams.num_workers,  # type: ignore
            pin_memory=self.hparams.pin_memory,  # type: ignore
            shuffle=False,
        )

    def test_dataloader(self) -> DataLoader[Any]:
        return DataLoader(
            dataset=self.data_test,  # type: ignore
            batch_size=self.batch_size,
            num_workers=self.hparams.num_workers,  # type: ignore
            pin_memory=self.hparams.pin_memory,  # type: ignore
            shuffle=False,
        )
