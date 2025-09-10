# Copyright 2025 AI for Oncology Research Group. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import argparse
import re
from concurrent.futures import ThreadPoolExecutor, as_completed
from pathlib import Path
from typing import List, NamedTuple, Optional, Pattern

import SimpleITK as sitk
import sqlalchemy
from omegaconf import OmegaConf
from fomo.database_models import Base, Image
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker
from tqdm import tqdm


class ImageMetadata(NamedTuple):
    filename: str
    origin: list[float]
    spacing: list[float]
    direction: list[float]
    shape: list[int]


class Dataset:
    path: Path
    extension: str
    exclude: Optional[Pattern[str]]

    def __init__(self, path: str, extension: str, exclude: str = ""):
        self.path = Path(path)
        self.extension = extension
        self.exclude = None

        if exclude:
            self.exclude = re.compile(exclude)


def process_image(image_fn: str, apply_ct_filters: bool = False, exclude_scouts: bool = False) -> ImageMetadata | None:
    """
    Reads only the metadata (origin, spacing, direction) of an image file without loading the pixel data and applies optional filtering.

    Args:
        image_fn: Path to the image file
        apply_ct_filters: If True, applies CT-specific filtering (512x512, single channel, etc.)
        exclude_scouts: If True, excludes scout/localizer images based on slice count and naming patterns
    """
    try:
        reader = sitk.ImageFileReader()
        reader.SetFileName(image_fn)
        reader.ReadImageInformation()

        shape = list(reader.GetSize())[::-1]
        channels = reader.GetNumberOfComponents()

        # Check for scout images if exclusion is enabled
        if exclude_scouts:
            filename_lower = image_fn.lower()
            # Check for scout/localizer keywords in filename
            scout_keywords = ["scout", "localizer", "survey", "planning", "silico"]
            if any(keyword in filename_lower for keyword in scout_keywords):
                print(f"Excluding scout image based on filename: '{image_fn}'")
                return None

            # Check for scout characteristics: very few slices (typically ≤ 10)
            if len(shape) == 3 and shape[0] <= 5:
                print(f"Excluding potential scout image with {shape[0]} slices: '{image_fn}'")
                return None

        # For CT data, enforce single channel requirement
        if apply_ct_filters and channels != 1:
            print(
                f"Volume with filename '{image_fn}' does not have a single channel, actual no. of channels: '{channels}'."
            )
            return None

        # For MRI data, allow multi-channel but warn
        if not apply_ct_filters and channels != 1:
            print(
                f"Warning: Volume with filename '{image_fn}' has {channels} channels. Processing anyway for MRI data."
            )

        # Apply CT-specific filters only if requested
        if apply_ct_filters:
            # Discard if the CT-scan does not have the proper dimensionalities, i.e. the CT-scan is a scout.
            if len(shape) != 3 or shape[1] != 512 or shape[2] != 512 or shape[0] <= 10:
                print(
                    f"Volume with filename '{image_fn}' does not have shape '[x, 512, 512]' with x > 10, actual shape: '{shape}'."
                )
                return None
        else:
            # For MRI/generic data, just check that it's 3D
            if len(shape) != 3:
                print(f"Volume with filename '{image_fn}' is not 3D, actual shape: '{shape}'.")
                return None

        spacing = list(reader.GetSpacing())
        origin = list(reader.GetOrigin())
        direction = list(reader.GetDirection())

        return ImageMetadata(image_fn, origin, spacing, direction, shape)

    except Exception as e:
        print(f"Error processing image '{image_fn}': {e}")
        return None


def store_metadata_in_db(image_data: ImageMetadata, session: sqlalchemy.orm.session.Session):
    """
    Stores the extracted metadata for the image in the session to the database.
    """

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
    )

    session.add(image)


def batch_process_image_parallel(
    image_filenames, session, batch_size=100, num_workers=4, apply_ct_filters=False, exclude_scouts=False
):
    """
    Processes image in batches, using parallel threads to obtain the metadata.
    Commits to the database after every batch.

    Args:
        image_filenames: List of image file paths
        session: Database session
        batch_size: Number of images to process before committing
        num_workers: Number of parallel workers
        apply_ct_filters: If True, applies CT-specific filtering
        exclude_scouts: If True, excludes scout/localizer images
    """
    counter = 0
    with ThreadPoolExecutor(max_workers=num_workers) as executor:
        futures = {
            executor.submit(process_image, str(image_fn), apply_ct_filters, exclude_scouts): image_fn
            for image_fn in image_filenames
        }

        batch_results = []
        with tqdm(desc="Processing", unit="image") as pbar:
            for future in as_completed(futures):
                try:
                    image_data = future.result()

                    if not image_data:
                        continue

                    batch_results.append(image_data)
                    counter += 1
                    pbar.update(1)

                    if counter % batch_size == 0:
                        for image_data in batch_results:
                            store_metadata_in_db(image_data, session)
                        session.commit()
                        batch_results.clear()

                except Exception as e:
                    print(f"Error processing {futures[future]}: {e}")

        if batch_results:
            for image_data in batch_results:
                store_metadata_in_db(image_data, session)
            session.commit()
            print(f"Processed {counter} records in total.")


def crawl_dataset_paths(datasets: List[Dataset]) -> List[str]:
    filepaths = []

    for dataset in datasets:
        for filepath in dataset.path.rglob(f"**/*{dataset.extension}"):
            filepath_str = str(filepath)
            if dataset.exclude and dataset.exclude.findall(filepath_str):
                continue

            filepaths.append(filepath_str)

    return filepaths


def print_database_sample(database_path: str, num_rows: int = 10) -> None:
    """
    Print a sample of the database contents for verification.

    Args:
        database_path: Path to the SQLite database
        num_rows: Number of rows to display
    """
    try:
        engine = create_engine(f"sqlite:///{database_path}")
        Session = sessionmaker(bind=engine)
        session = Session()

        # Query the first num_rows images
        images = session.query(Image).limit(num_rows).all()

        if not images:
            print("No images found in the database.")
            return

        print(f"\n=== Database Sample: First {len(images)} rows ===")
        print(f"{'ID':<5} {'Filename':<80} {'Shape':<20} {'Spacing':<30}")
        print("-" * 135)

        for image in images:
            # Truncate filename if too long
            filename = image.filename
            if len(filename) > 77:
                filename = "..." + filename[-74:]

            # Format shape and spacing for display
            shape_str = str(image.shape)
            spacing_str = f"[{', '.join([f'{x:.2f}' for x in image.spacing[:3]])}]"  # Show first 3 values

            print(f"{image.id:<5} {filename:<80} {shape_str:<20} {spacing_str:<30}")

        print(f"\nTotal images in database: {session.query(Image).count()}")
        session.close()

    except Exception as e:
        print(f"Error reading database sample: {e}")


def main():
    parser = argparse.ArgumentParser(
        description="Create an SQL database from the images in the dataset. In dataset_config, you can declare a list of datasets to crawl for image data."
    )
    parser.add_argument(
        "--dataset-config",
        type=Path,
        default=Path("src/python/research/fomo/ct/configs/dataset/dataset_config.yaml"),
        help="Path to yaml-configuration which contains a list of tuples outlining dataset path and filetype extension. Ex: (/path/to/crawl, .nrrd).",
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
        default=3,
        help="Number of threads to use for parallel processing.",
    )
    parser.add_argument(
        "--apply-ct-filters",
        action="store_true",
        help="Apply CT-specific filters (512x512 shape requirement, etc.). Use for CT data only.",
    )
    parser.add_argument(
        "--modality",
        type=str,
        choices=["ct", "mri", "generic"],
        default="generic",
        help="Imaging modality. 'ct' applies CT filters, 'mri' and 'generic' do not.",
    )
    parser.add_argument(
        "--exclude-scouts",
        action="store_true",
        help="Exclude scout/localizer images based on filename keywords and low slice count (≤10 slices).",
    )
    args = parser.parse_args()

    # Determine whether to apply CT filters
    apply_ct_filters = args.apply_ct_filters or args.modality == "ct"

    config = OmegaConf.load(args.dataset_config)
    datasets = [Dataset(**dataset) for dataset in config]
    filepaths = crawl_dataset_paths(datasets)

    modality_str = f" ({args.modality})" if args.modality != "generic" else ""
    filter_str = " with CT filters" if apply_ct_filters else " without CT filters"
    scout_str = " excluding scouts" if args.exclude_scouts else ""
    print(f"Found {len(filepaths)} suitable images to process{modality_str}{filter_str}{scout_str}.")

    engine = create_engine(f"sqlite:///{args.database_path}")
    Base.metadata.create_all(engine)
    Session = sessionmaker(bind=engine)
    session = Session()

    batch_process_image_parallel(
        filepaths,
        session,
        batch_size=args.batch_size,
        num_workers=args.num_workers,
        apply_ct_filters=apply_ct_filters,
        exclude_scouts=args.exclude_scouts,
    )

    session.close()
    print("All images have been processed.")

    # Print database sample
    print_database_sample(args.database_path)


if __name__ == "__main__":
    main()
