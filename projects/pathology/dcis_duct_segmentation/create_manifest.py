"""This script creates a manifest for the DCIS duct segmentation project."""

import json
import random
import argparse
import csv
from pathlib import Path
import re
from rich.progress import Progress, TextColumn, BarColumn, TaskProgressColumn, TimeRemainingColumn
from rich.console import Console
from rich.table import Table
from dlup import SlideImage
from dlup.utils.backends import ImageBackend  # type: ignore
from ahcore.utils.database_models import (
    CategoryEnum,
    Image,
    ImageAnnotations,
    ImageLabels,
    Manifest,
    Mask,
    Patient,
    PatientLabels,
    Split,
    SplitDefinitions,
)
from ahcore.utils.manifest import open_db

# Validation files list (patient codes)
VALIDATION_PATIENT_CODES = []


def get_patient_code_from_filename(filename: str) -> str:
    """Extract patient code from filename.

    Examples:
    - T14_62435_A1_HE.mrxs -> T14-62435
    - DCIS_CL_T16-68183_A2.svs -> T16-68183
    - CFMPB000_DCIS_T91-05157_B2_HE_MA.svs -> T91-05157
    """
    filename = Path(filename).name  # Get just the filename without path

    # Look for T followed by digits and then either - or _ and more digits
    match = re.search(r"T(\d+)[_-](\d+)", filename)
    if match:
        year, number = match.groups()
        return f"T{year}-{number}"

    return Path(filename).stem  # Fallback to the stem as patient code


def is_validation_file(filename: str) -> bool:
    """Check if a file should be in the validation set."""

    # Extract patient code from the filename
    patient_code = get_patient_code_from_filename(filename)

    # Check if the patient code is in the validation set
    return patient_code in VALIDATION_PATIENT_CODES


def populate_from_csv(
    session,
    image_folder_root: Path,
    annotations_root: Path,
):
    """Populate the manifest from a CSV file."""
    # Automatically use the annotations.csv file in annotations_root
    csv_file = annotations_root / "annotations.csv"

    # First create and commit the manifest
    manifest = Manifest(name="v1.0")
    session.add(manifest)
    session.commit()  # Commit instead of flush to prevent ID conflicts

    split_definition = SplitDefinitions(version="v1.0", description="Validation split 4 images")
    session.add(split_definition)
    session.commit()  # Commit instead of flush

    count = 0
    patient_dict = {}  # Cache patients to avoid duplicate lookups/inserts

    # Read the CSV file and get total rows for progress bar
    with open(csv_file, "r") as f:
        total_rows = sum(1 for _ in csv.DictReader(f))

    # Reset file pointer
    with open(csv_file, "r") as f:
        reader = csv.DictReader(f)

        # Create a rich progress bar
        with Progress(
            TextColumn("[bold blue]{task.description}"),
            BarColumn(),
            TaskProgressColumn(),
            TimeRemainingColumn(),
        ) as progress:
            task = progress.add_task("[green]Processing images...", total=total_rows)

            for row in reader:
                progress.update(task, advance=1)

                image_path = row["Relative Path"]
                mask_path = row["Tissue mask filename"]
                annotation_path = row["Annotation filename"]

                # Extract filename from path
                filename = Path(image_path).name

                # Get patient code from filename
                patient_code = get_patient_code_from_filename(filename)

                # Check if patient already exists (first in our cache, then in DB)
                if patient_code in patient_dict:
                    patient = patient_dict[patient_code]
                else:
                    existing_patient = session.query(Patient).filter_by(patient_code=patient_code).first()  # type: ignore
                    if existing_patient:
                        patient = existing_patient
                    else:
                        patient = Patient(patient_code=patient_code, manifest=manifest)
                        session.add(patient)
                        session.commit()  # Commit to get a valid ID before using it

                        # Determine if this file should be in validation set
                        split_category = CategoryEnum.FIT if not is_validation_file(filename) else CategoryEnum.VALIDATE

                        split = Split(
                            category=split_category,
                            patient=patient,
                            split_definition=split_definition,
                        )
                        session.add(split)
                        session.commit()  # Commit to avoid integrity issues

                    # Store in cache to avoid repeated lookups
                    patient_dict[patient_code] = patient

                # Add the image to the database
                try:
                    with SlideImage.from_file_path(
                        image_folder_root / image_path,
                        backend=ImageBackend.PYVIPS,
                    ) as slide:
                        mpp = slide.mpp
                        width, height = slide.size
                        image = Image(
                            filename=image_path,
                            mpp=mpp,
                            height=height,
                            width=width,
                            reader="PYVIPS",
                            patient=patient,
                        )
                    session.add(image)
                    session.commit()  # Commit to get a valid image ID

                    # Add annotation if available
                    if annotation_path:
                        image_annotation = ImageAnnotations(filename=annotation_path, reader="DLUP_XML")
                        session.add(image_annotation)
                        session.commit()
                        # Update the image with the annotation ID
                        image.annotation_id = image_annotation.id
                        session.add(image)
                        session.commit()

                    # Add mask if available
                    if mask_path:
                        # Create the mask with direct reference to the image_id
                        mask = Mask(filename=mask_path, reader="IMAGEIO", image_id=image.id)
                        session.add(mask)
                        session.commit()

                    count += 1
                except Exception as e:
                    progress.console.print(f"[red]Error processing {image_path}: {e}")
                    session.rollback()

    progress.console.print(f"[bold green]Added {count} images to the manifest.")

    # Print image-to-patient mapping for verification
    print_image_patient_mapping(session)

    # Print mask-to-image mapping for verification
    print_mask_image_mapping(session)

    # Print annotation-to-image mapping for verification
    print_annotation_image_mapping(session)


def print_image_patient_mapping(session):
    """Print a table showing the mapping between images and patients."""
    console = Console()

    # Create a table for display
    table = Table(title="Image to Patient ID Mapping")
    table.add_column("Image ID", style="cyan")
    table.add_column("Image Filename", style="green")
    table.add_column("Patient ID", style="magenta")
    table.add_column("Patient Code", style="yellow")

    # Query all images with their patient information
    results = (
        session.query(Image.id, Image.filename, Patient.id, Patient.patient_code)
        .join(Patient, Image.patient_id == Patient.id)
        .order_by(Patient.patient_code)
        .all()
    )

    # Display results in the table
    for image_id, filename, patient_id, patient_code in results:
        table.add_row(str(image_id), str(filename), str(patient_id), patient_code)

    # Print the table
    console.print(table)
    console.print(f"Total images: {len(results)}")


def print_mask_image_mapping(session):
    """Print a table showing the mapping between masks and images."""
    console = Console()

    # Create a table for display
    table = Table(title="Mask to Image Mapping")
    table.add_column("Mask ID", style="cyan")
    table.add_column("Mask Filename", style="green")
    table.add_column("Image ID", style="magenta")
    table.add_column("Image Filename", style="yellow")

    # Query all masks with their image information
    results = (
        session.query(Mask.id, Mask.filename, Image.id, Image.filename)
        .join(Image, Mask.image_id == Image.id)
        .order_by(Mask.id)
        .all()
    )

    # Display results in the table
    for mask_id, mask_filename, image_id, image_filename in results:
        table.add_row(str(mask_id), str(mask_filename), str(image_id), str(image_filename))

    # Print the table
    console.print(table)
    console.print(f"Total masks: {len(results)}")


def print_annotation_image_mapping(session):
    """Print a table showing the mapping between annotations and images."""
    console = Console()

    # Create a table for display
    table = Table(title="Annotation to Image Mapping")
    table.add_column("Annotation ID", style="cyan")
    table.add_column("Annotation Filename", style="green")
    table.add_column("Image ID", style="magenta")
    table.add_column("Image Filename", style="yellow")

    # Query all annotations with their image information
    results = (
        session.query(ImageAnnotations.id, ImageAnnotations.filename, Image.id, Image.filename)
        .join(Image, Image.annotation_id == ImageAnnotations.id)
        .order_by(ImageAnnotations.id)
        .all()
    )

    # Display results in the table
    for annotation_id, annotation_filename, image_id, image_filename in results:
        table.add_row(str(annotation_id), str(annotation_filename), str(image_id), str(image_filename))

    # Print the table
    console.print(table)
    console.print(f"Total annotations: {len(results)}")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Populate manifest database with DCIS duct segmentation data")
    parser.add_argument(
        "--image-root",
        type=Path,
        required=True,
        help="Path to the root folder containing image data",
    )
    parser.add_argument(
        "--annotations-root",
        type=Path,
        required=True,
        help="Path to the root folder containing annotation files and annotations.csv",
    )
    parser.add_argument(
        "--manifest-path",
        type=str,
        required=True,
        help="Path to the SQLite manifest database file",
    )
    parser.add_argument(
        "--validation-codes",
        type=Path,
        default=Path(__file__).parent / "validation_codes.txt",
        help="Path to the validation codes file (default: validation_codes.txt in script directory)",
    )

    args = parser.parse_args()

    # Read validation codes from file
    if args.validation_codes.exists():
        with open(args.validation_codes, "r") as f:
            VALIDATION_PATIENT_CODES = [line.strip() for line in f if line.strip()]
    else:
        print(f"Warning: Validation codes file not found at {args.validation_codes}")

    with open_db(f"sqlite:///{args.manifest_path}", ensure_exists=False) as session:
        populate_from_csv(
            session,
            args.image_root,
            args.annotations_root,
        )
