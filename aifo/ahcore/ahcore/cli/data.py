"""Module to write copy manifests files over to SCRATCH directory"""

import argparse
import hashlib
import os
import shutil
import sys
from pathlib import Path
from typing import List, Tuple

from ahcore.cli import dir_path
from ahcore.utils.manifest import DataManager
from rich.progress import Progress, TextColumn, BarColumn, TaskProgressColumn, TimeRemainingColumn, FileSizeColumn


def _quick_hash(file_path: Path, max_bytes: int = 10**6) -> str:
    hasher = hashlib.sha256()
    with open(file_path, "rb") as f:
        block = f.read(max_bytes)
        hasher.update(block)
    return hasher.hexdigest()


def copy_data(args: argparse.Namespace) -> None:
    manifest_url = args.manifest_uri
    base_dir = args.base_dir
    dataset_name = args.dataset_name
    target_dir = os.environ.get("SCRATCH", None)

    # Initialize a counter for the total data size
    total_size = 0

    if target_dir is None or not os.access(target_dir, os.W_OK):
        print("Please set the SCRATCH environment variable to a writable directory.")
        sys.exit(1)

    with DataManager(manifest_url) as dm:
        # First collect all files to copy and count them
        files_to_copy: List[Tuple[Path, Path, bool]] = []  # (source, destination, is_mrxs)

        all_records = dm.get_records_by_split(args.manifest_name, args.split_name, split_category=None)
        for patient in all_records:
            for image in patient.images:
                image_fn = image.filename
                get_from = base_dir / image_fn
                write_to = Path(target_dir) / dataset_name / image_fn

                is_mrxs = get_from.suffix.lower() == ".mrxs"
                files_to_copy.append((get_from, write_to, is_mrxs))

        # Now copy with proper progress tracking
        with Progress(
            TextColumn("[bold blue]{task.description}"),
            BarColumn(),
            TaskProgressColumn(),
            TimeRemainingColumn(),
            FileSizeColumn(),
        ) as progress:
            # Create the main task with the correct total count
            task = progress.add_task(f"[cyan]Copying {len(files_to_copy)} files...", total=len(files_to_copy))

            for get_from, write_to, is_mrxs in files_to_copy:
                write_to.parent.mkdir(parents=True, exist_ok=True)

                if write_to.exists():
                    # compute the hash of previous and new file
                    try:
                        old_hash = _quick_hash(write_to)
                        new_hash = _quick_hash(get_from)
                        if old_hash == new_hash:
                            # Skip if they are the same
                            progress.console.log(f"[yellow]Skipping (already exists): {get_from.name}")
                            progress.update(task, advance=1)
                            continue
                    except (FileNotFoundError, PermissionError) as e:
                        progress.console.log(f"[red]Error checking hash: {e}")

                file_size = get_from.stat().st_size
                total_size += file_size

                # Copy file from get_from to write_to
                progress.console.log(f"[green]Copying ({file_size / 1024**2:.1f} MB): {get_from.name}")
                try:
                    shutil.copy(get_from, write_to)

                    # Handle MRXS associated files
                    if is_mrxs:
                        # Copy the folder with the same name as the mrxs file
                        source_folder = get_from.parent / get_from.stem
                        dest_folder = write_to.parent / write_to.stem

                        if source_folder.exists():
                            progress.console.log(f"[cyan]Copying mrxs data folder: {source_folder.name}")
                            if dest_folder.exists():
                                shutil.rmtree(dest_folder)
                            shutil.copytree(source_folder, dest_folder)

                            # Calculate folder size - the .mrxs file itself is already counted separately above
                            folder_size = sum(f.stat().st_size for f in source_folder.rglob("*") if f.is_file())
                            total_size += folder_size
                            progress.console.log(
                                f"[cyan]Added {folder_size / 1024**2:.1f} MB from mrxs (total: {(file_size + folder_size) / 1024**2:.1f} MB)"
                            )

                    progress.update(task, advance=1)
                except Exception as e:
                    progress.console.log(f"[red]Error copying {get_from}: {e}")
                    progress.update(task, advance=1)  # Still advance to show progress

    progress.console.log(f"[bold green]Total data copied: {total_size / 1024**3:.2f} GB")


def register_parser(
    parser: argparse._SubParsersAction,  # pylint: disable=unsubscriptable-object
) -> None:  # pylint: disable=E1136
    """Register inspect commands to a root parser."""
    data_parser = parser.add_parser("data", help="Data utilities")
    data_subparsers = data_parser.add_subparsers(help="Data subparser")
    data_subparsers.required = True
    data_subparsers.dest = "subcommand"

    _parser: argparse.ArgumentParser = data_subparsers.add_parser(
        "copy-data-from-manifest",
        help="Copy the data to a different drive based on the manifest. "
        "The data will be copied over to $SCRATCH / DATASET_NAME",
    )

    _parser.add_argument(
        "manifest_uri",
        type=str,
        help="URI that refers to the sqlalchemy supported database path.",
    )
    _parser.add_argument(
        "manifest_name",
        type=str,
        help="Name of the manifest to copy the data from.",
    )
    _parser.add_argument(
        "split_name",
        type=str,
        help="Name of the split in the database to copy the data from.",
    )
    _parser.add_argument(
        "base_dir",
        type=dir_path(require_writable=False),
        help="Directory to which the images paths defined in the manifest are relative to.",
    )
    _parser.add_argument(
        "dataset_name",
        type=str,
        help="Name of the dataset to copy the data to. The data will be copied over to $SCRATCH / DATASET_NAME",
    )
    _parser.set_defaults(subcommand=copy_data)
