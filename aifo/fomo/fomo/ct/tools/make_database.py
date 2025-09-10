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


def process_image(image_fn: str) -> ImageMetadata | None:
    """
    Reads only the metadata (origin, spacing, direction) of an image file without loading the pixel data and applies some filtering.
    """
    reader = sitk.ImageFileReader()
    reader.SetFileName(image_fn)
    reader.ReadImageInformation()

    shape = list(reader.GetSize())[::-1]
    channels = reader.GetNumberOfComponents()

    if channels != 1:
        print(f"Volume with filename '{image_fn}' does not a single channel, actual no. of channels: '{channels}'.")
        return None

    # Discard if the CT-scan does not have the proper dimensionalities, i.e. the CT-scan is a scout.
    if len(shape) != 3 or shape[1] != 512 or shape[2] != 512 or shape[0] <= 10:
        print(
            f"Volume with filename '{image_fn}' does not have shape '[x, 512, 512]' with x > 10, actual shape: '{shape}'."
        )
        return None

    spacing = list(reader.GetSpacing())
    origin = list(reader.GetOrigin())
    direction = list(reader.GetDirection())

    return ImageMetadata(image_fn, origin, spacing, direction, shape)


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


def batch_process_image_parallel(image_filenames, session, batch_size=100, num_workers=4):
    """
    Processes image in batches, using parallel threads to obtain the metadata.
    Commits to the database after every batch.
    """
    counter = 0
    with ThreadPoolExecutor(max_workers=num_workers) as executor:
        futures = {executor.submit(process_image, str(image_fn)): image_fn for image_fn in image_filenames}

        batch_results = []
        with tqdm(desc="Processing", unit="pair") as pbar:
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
    args = parser.parse_args()

    config = OmegaConf.load(args.dataset_config)
    datasets = [Dataset(**dataset) for dataset in config]
    filepaths = crawl_dataset_paths(datasets)
    print(f"Found {len(filepaths)} suitable CT-scans to process.")

    engine = create_engine(f"sqlite:///{args.database_path}")
    Base.metadata.create_all(engine)
    Session = sessionmaker(bind=engine)
    session = Session()

    batch_process_image_parallel(
        filepaths,
        session,
        batch_size=args.batch_size,
        num_workers=args.num_workers,
    )

    print("All images have been processed.")


if __name__ == "__main__":
    main()
