# from torchvision import transforms as tf
import json
from pathlib import Path
from typing import Union

import nibabel
import torch.utils.data as data


class CTDataset(data.Dataset):
    """CT scan dataset"""

    def __init__(
        self,
        data_dir: Union[str, Path],
        splits_fn: str,
        split: str = "train",
        transforms=None,
    ):
        self.data_dir = Path(data_dir)
        self.ct_dir = self.data_dir / "data"

        self.transforms = transforms

        self.manifest = []
        self._load_manifest()
        self._load_splits(Path(splits_fn))

        if not split == "all":
            if split in ["train", "val", "test"]:
                split = self.splits[split]
                self.manifest = [self.manifest[i] for i in split]  # type: ignore
            elif split == "combined":
                splits = self.splits["train"] + self.splits["val"] + self.splits["test"]
                self.manifest = [self.manifest[i] for i in splits]

        physicians = [3, 12, 15, 16, 25, 32, 34, 41, 53]
        self.physicians = physicians
        self.physicians_map = {p: i for i, p in enumerate(self.physicians)}

        rest_label = len(self.physicians)
        all_phys = list(set([_["physician"] for _ in self.manifest]))
        for p in all_phys:
            if p not in self.physicians:
                self.physicians_map[p] = rest_label

        # top_physicians = [3, 12, 16, 32, 34, 41, 53]
        # self.physicians_map = {p: 0 for p in top_physicians}
        # all_phys = list(set([_["physician"] for _ in self.manifest]))
        # for p in all_phys:
        #     if p not in top_physicians:
        #         self.physicians_map[p] = 1

    def __len__(self):
        return len(self.manifest)

    def _load_manifest(self):
        with open(self.data_dir / "manifest.json", "r") as f:
            data = json.load(f)
            self.manifest = data

    def _load_splits(self, splits_fn: Path):
        with open(self.data_dir / "splits" / splits_fn, "r") as f:
            data = json.load(f)
            self.splits = data

    def __getitem__(self, index):
        item = self.manifest[index]

        patient_dir = self.ct_dir / f"patient_{item['AnonID']}"
        scan_fn = patient_dir / f"{item['scan_label']}.nii.gz"
        mask_fn = patient_dir / f"{item['scan_label']}.seg.nii.gz"

        # load the scan in float16
        scan = nibabel.load(scan_fn).get_fdata().astype("float16")  # type: ignore

        # load the mask in uint8
        mask = nibabel.load(mask_fn).get_fdata().astype("uint8")  # type: ignore

        # add a channel dimension
        scan = scan[None, :, :, :]
        mask = mask[None, :, :, :]

        # the scan ranges from -1024 to approximately 2000, normalize to 0-1
        scan = (scan + 1024) / 2524

        target = self.physicians_map[item["physician"]]

        out = {"scan": scan, "mask": mask, "target": target}
        if self.transforms:
            out = self.transforms(out)

        return out


if __name__ == "__main__":
    # test the dataset

    data_dir = "/projects/physician_gtv_fingerprint_data/m42_nii/phys_fp_crop/"
    splits_fn = "default/0.json"

    dataset = CTDataset(data_dir, splits_fn, split="train")
    dataset = CTDataset(data_dir, splits_fn, split="val")
