import argparse
from concurrent.futures import ThreadPoolExecutor, as_completed
from pathlib import Path
from typing import Generator, NamedTuple, Optional

import numpy as np
import numpy.typing as npt
import SimpleITK as sitk
from fomo.database_models import Base, Image, Mask
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker
from tqdm import tqdm


class ImageMetadata(NamedTuple):
    filename: str
    origin: list[float]
    spacing: list[float]
    direction: list[float]
    shape: list[int]


class MaskMetadata(NamedTuple):
    filename: str
    origin: list[float]
    spacing: list[float]
    direction: list[float]
    shape: list[int]
    bbox: tuple[tuple[int, ...], tuple[int, ...]]


def read_image_metadata(image_fn: str) -> ImageMetadata:
    """
    Reads only the metadata (origin, spacing, direction) of an image file without loading the pixel data.
    """
    reader = sitk.ImageFileReader()
    reader.SetFileName(image_fn)
    reader.ReadImageInformation()

    origin = list(reader.GetOrigin())
    spacing = list(reader.GetSpacing())
    direction = list(reader.GetDirection())
    shape = list(reader.GetSize())[::-1]

    return ImageMetadata(image_fn, origin, spacing, direction, shape)


def get_bounding_box(
    array: npt.NDArray[np.int_],
) -> tuple[tuple[int, ...], tuple[int, ...]]:
    """
    Compute the bounding box of a binary mask in an array.
    """
    non_zero_indices = np.argwhere(array > 0)

    if non_zero_indices.size == 0:
        raise ValueError("No non-zero pixels found in the image.")

    min_coords = non_zero_indices.min(axis=0)
    max_coords = non_zero_indices.max(axis=0)
    size = max_coords - min_coords + 1

    return min_coords.tolist(), size.tolist()


def process_image_and_mask(image_fn: str, mask_fn: Optional[str]) -> tuple[ImageMetadata, Optional[MaskMetadata]]:
    """
    Extract metadata from the image and mask and compute the bounding box.
    """
    image_metadata = read_image_metadata(image_fn)
    mask_metadata = None

    if mask_fn:
        mask_meta = read_image_metadata(mask_fn)
        itk_mask = sitk.ReadImage(mask_fn)
        min_coords, size = get_bounding_box(sitk.GetArrayFromImage(itk_mask))

        bbox = (min_coords, size)

        mask_metadata = MaskMetadata(
            mask_fn,
            mask_meta.origin,
            mask_meta.spacing,
            mask_meta.direction,
            mask_meta.shape,
            bbox,
        )

    return image_metadata, mask_metadata


def store_metadata_in_db(image_data: ImageMetadata, mask_data: Optional[MaskMetadata], session):
    """
    Stores the extracted metadata for the image and mask in the database.
    """
    mask = None
    if mask_data:
        existing_mask = session.query(Mask).filter_by(filename=str(mask_data.filename)).first()
        if existing_mask:
            mask = existing_mask
        else:
            mask = Mask(
                filename=mask_data.filename,
                origin=mask_data.origin,
                spacing=mask_data.spacing,
                direction=mask_data.direction,
                shape=mask_data.shape,
                bbox=mask_data.bbox,
            )
            session.add(mask)

    existing_image = session.query(Image).filter_by(filename=str(image_data.filename)).first()
    if existing_image:
        print(f"Image {image_data.filename} already exists in the database. Skipping...")
        return

    image = Image(
        filename=str(image_data.filename),
        origin=image_data.origin,
        spacing=image_data.spacing,
        direction=image_data.direction,
        shape=image_data.shape,
        mask=mask,  # This can be None if there is no mask
    )
    session.add(image)
    session.commit()


def batch_process_image_and_masks_parallel(data_pairs, session, batch_size=100, num_workers=4):
    """
    Processes image and mask pairs in batches, using parallel threads to obtain the metadata and bounding boxes.
    Commits to the database after every batch.
    """
    counter = 0
    with ThreadPoolExecutor(max_workers=num_workers) as executor:
        futures = {
            executor.submit(process_image_and_mask, str(image_fn), mask_fn if mask_fn else None): (image_fn, mask_fn)
            for image_fn, mask_fn in data_pairs
        }

        batch_results = []
        with tqdm(desc="Processing", unit="pair") as pbar:
            for future in as_completed(futures):
                try:
                    image_data, mask_data = future.result()
                    batch_results.append((image_data, mask_data))
                    counter += 1
                    pbar.update(1)

                    if counter % batch_size == 0:
                        for image_data, mask_data in batch_results:
                            store_metadata_in_db(image_data, mask_data, session)
                        session.commit()
                        batch_results.clear()

                except Exception as e:
                    print(f"Error processing {futures[future]}: {e}")

        if batch_results:
            for image_data, mask_data in batch_results:
                store_metadata_in_db(image_data, mask_data, session)
            session.commit()
            print(f"Processed {counter} records in total.")


def get_data_pairs(base_path: Path) -> Generator[tuple[str, Optional[str]], None, None]:
    """
    Recursively find image and mask pairs in the dataset path.
    """
    for folder in base_path.glob("*"):
        if not folder.is_dir():
            continue

        image_fn_pre = folder / f"{folder.name}_0000.nrrd"
        image_fn_post = folder / f"{folder.name}_0001.nrrd"
        mask_fn = folder / f"{folder.name}.nrrd"

        if image_fn_pre.exists():
            mask_str = str(mask_fn) if mask_fn.exists() else None
            yield str(image_fn_pre), mask_str
        if image_fn_post.exists():
            mask_str = str(mask_fn) if mask_fn.exists() else None
            yield str(image_fn_post), mask_str


def main():
    parser = argparse.ArgumentParser(
        description="Create an SQL database from the images and masks in the dataset. Will recursively search in dataset_path for images and masks."
    )
    parser.add_argument(
        "--dataset-path",
        type=Path,
        default=Path("/projects/mri_fomo/nki_breast_mri_nrrd/nrrd/"),
        help="Path to the dataset.",
    )
    parser.add_argument(
        "--database-path",
        type=str,
        default="database.sqlite",
        help="Path to the database.",
    )
    parser.add_argument(
        "--batch-size",
        type=int,
        default=100,
        help="Number of records to batch before committing to the database.",
    )
    parser.add_argument(
        "--num-workers",
        type=int,
        default=4,
        help="Number of threads to use for parallel processing.",
    )
    args = parser.parse_args()

    engine = create_engine(f"sqlite:///{args.database_path}")
    Base.metadata.create_all(engine)

    Session = sessionmaker(bind=engine)
    session = Session()

    base_path = args.dataset_path
    for subpath in ["fs", "nfs"]:
        print(f"Processing images and masks in '{subpath}'...")
        data_pairs = get_data_pairs(base_path / subpath)

        batch_process_image_and_masks_parallel(
            data_pairs,
            session,
            batch_size=args.batch_size,
            num_workers=args.num_workers,
        )

    print("All images and masks have been processed.")


if __name__ == "__main__":
    main()
