"""This is an example on how to populate an ahcore manifest database using the TCGA dataset."""

import json
import random
import argparse
from pathlib import Path

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


def get_patient_from_tcga_id(tcga_filename: str) -> str:
    return tcga_filename[:12]


def populate_from_annotated_images(
    session,
    image_folder_root: Path,
    annotation_folder: Path,
    path_to_tcga_mapping: Path,
    path_to_camelyon16_mapping: Path,
):
    console = Console()
    console.print("[bold]Loading mapping files...[/bold]")

    with open(path_to_tcga_mapping, "r") as f:
        tcga_mapping = json.load(f)
    with open(path_to_camelyon16_mapping, "r") as f:
        camelyon16_mapping = json.load(f)

    console.print("[bold green]✓[/bold green] Mapping files loaded successfully")

    manifest = Manifest(name="v1.0")
    session.add(manifest)
    session.flush()

    split_definition = SplitDefinitions(version="v1.0", description="Initial split")
    session.add(split_definition)
    session.flush()

    tcga_annotation_folder = annotation_folder / "xmls/TCGA/"
    camelyon16_annotation_folder = annotation_folder / "xmls/Camelyon16/"

    # Count files for progress bar
    tcga_files = list(tcga_annotation_folder.glob("*.xml"))
    camelyon16_files = list(camelyon16_annotation_folder.glob("*.xml"))

    # Process TCGA files with progress bar
    with Progress(
        TextColumn("[bold blue]{task.description}"),
        BarColumn(),
        TaskProgressColumn(),
        TimeRemainingColumn(),
    ) as progress:
        tcga_task = progress.add_task("[green]Processing TCGA images...", total=len(tcga_files))
        tcga_count = 0

        for file in tcga_files:
            progress.update(tcga_task, advance=1)

            try:
                patient_code = get_patient_from_tcga_id(file.name)
                annotation_path = tcga_annotation_folder / f"{file.name}"

                existing_patient = session.query(Patient).filter_by(patient_code=patient_code).first()  # type: ignore
                if existing_patient:
                    patient = existing_patient
                else:
                    patient = Patient(patient_code=patient_code, manifest=manifest)
                    session.add(patient)
                    session.flush()

                    split_category = random.choices(
                        [CategoryEnum.FIT, CategoryEnum.VALIDATE, CategoryEnum.TEST],
                        [90, 10, 0],
                    )[0]

                    split = Split(
                        category=split_category,
                        patient=patient,
                        split_definition=split_definition,
                    )
                    session.add(split)
                    session.flush()

                # Add only the label if it does not exist yet.
                existing_label = session.query(PatientLabels).filter_by(key="study", patient_id=patient.id).first()
                if not existing_label:
                    patient_label = PatientLabels(key="study", value="BRCA", patient=patient)
                    session.add(patient_label)
                    session.flush()

                filename = tcga_mapping[file.stem]

                kwargs = {}
                if (
                    "TCGA-OL-A5RY-01Z-00-DX1.AE4E9D74-FC1C-4C1E-AE6D-5DF38899BBA6.svs" in filename
                    or "TCGA-OL-A5RW-01Z-00-DX1.E16DE8EE-31AF-4EAF-A85F-DB3E3E2C3BFF.svs" in filename
                ):
                    kwargs["overwrite_mpp"] = (0.25, 0.25)

                with SlideImage.from_file_path(
                    image_folder_root / f"TCGA/images/{filename}",
                    backend=ImageBackend.PYVIPS,
                    **kwargs,
                ) as slide:
                    mpp = slide.mpp
                    width, height = slide.size
                    image = Image(
                        filename=f"TCGA/images/{filename}",
                        mpp=mpp,
                        height=height,
                        width=width,
                        reader="PYVIPS",
                        patient=patient,
                    )
                session.add(image)
                session.flush()
                annotation_file = f"xmls/TCGA/{file.stem}.xml"
                image_annotation = ImageAnnotations(filename=str(annotation_file), reader="DLUP_XML", images=[image])
                image.annotation_id = image_annotation.id
                session.add(image_annotation)

                tcga_count += 1
                session.commit()
            except Exception as e:
                progress.console.print(f"[red]Error processing {file.name}: {e}")
                session.rollback()

        progress.console.print(f"[bold green]Added {tcga_count} images from TCGA.[/bold green]")

    # Process Camelyon16 files with progress bar
    with Progress(
        TextColumn("[bold blue]{task.description}"),
        BarColumn(),
        TaskProgressColumn(),
        TimeRemainingColumn(),
    ) as progress:
        camelyon_task = progress.add_task("[green]Processing Camelyon16 images...", total=len(camelyon16_files))
        camelyon_count = 0

        for file in camelyon16_files:
            progress.update(camelyon_task, advance=1)

            try:
                image_code = (
                    file.stem
                )  # Here, we treat image name as the patient code since Camelyon16 does not have a patient code.

                existing_patient = session.query(Patient).filter_by(patient_code=image_code).first()  # type: ignore
                if existing_patient:
                    patient = existing_patient
                else:
                    patient = Patient(patient_code=image_code, manifest=manifest)
                    session.add(patient)
                    session.flush()

                split_category = random.choices(
                    [CategoryEnum.FIT, CategoryEnum.VALIDATE, CategoryEnum.TEST],
                    [80, 10, 10],
                )[0]

                split = Split(
                    category=split_category,
                    patient=patient,
                    split_definition=split_definition,
                )
                session.add(split)
                session.flush()

                filename = camelyon16_mapping[file.stem]

                with SlideImage.from_file_path(
                    image_folder_root / f"Camelyon16/images/{filename}",
                    backend=ImageBackend.PYVIPS,
                    **{},
                ) as slide:
                    mpp = slide.mpp
                    width, height = slide.size
                    image = Image(
                        filename=f"Camelyon16/images/{filename}",
                        mpp=mpp,
                        height=height,
                        width=width,
                        reader="PYVIPS",
                        patient=patient,
                    )
                session.add(image)
                session.flush()

                annotation_file = f"xmls/Camelyon16/{file.stem}.xml"
                image_annotation = ImageAnnotations(filename=str(annotation_file), reader="DLUP_XML", images=[image])
                image.annotation_id = image_annotation.id
                session.add(image_annotation)

                camelyon_count += 1
                session.commit()
            except Exception as e:
                progress.console.print(f"[red]Error processing {file.name}: {e}")
                session.rollback()

        progress.console.print(f"[bold green]Added {camelyon_count} images from Camelyon16.[/bold green]")

    # Print summary
    console.print("[bold green]✓[/bold green] Database population completed successfully")
    console.print(f"[bold]Total images added:[/bold] {tcga_count + camelyon_count}")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Populate manifest database with TCGA and Camelyon16 data")
    parser.add_argument(
        "--annotation-folder",
        type=Path,
        required=True,
        help="Path to the annotation folder containing TCGA and Camelyon16 DLUP_XML files",
    )
    parser.add_argument(
        "--image-folder-root",
        type=Path,
        required=True,
        help="Path to the root folder containing TCGA and Camelyon16 image data",
    )
    parser.add_argument(
        "--tcga-mapping",
        type=Path,
        required=True,
        help="Path to the TCGA identifier mapping JSON file",
    )
    parser.add_argument(
        "--camelyon16-mapping",
        type=Path,
        required=True,
        help="Path to the Camelyon16 identifier mapping JSON file",
    )
    parser.add_argument(
        "--manifest-path",
        type=str,
        required=True,
        help="Path to the SQLite manifest database file",
    )

    args = parser.parse_args()

    with open_db(f"sqlite:///{args.manifest_path}", ensure_exists=False) as session:
        populate_from_annotated_images(
            session,
            args.image_folder_root,
            args.annotation_folder,
            args.tcga_mapping,
            args.camelyon16_mapping,
        )
