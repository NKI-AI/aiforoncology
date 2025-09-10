#!/usr/bin/env python3
"""
TCGA Bulk Import Script for SlideInsight

This script processes TCGA data folders and imports them into SlideInsight:
1. Creates studies from folder names (human-readable, fixes "nos" -> "NOS")
2. Extracts case IDs from TCGA filenames (e.g., TCGA-02-0001 from full filename)
3. Creates slides and assigns them to cases
4. Organizes cases within studies

Usage:
    python tcga_bulk_import.py --url http://localhost:3000 --data-dir /data/groups/public/archive/TCGA/images
"""

import argparse
import asyncio
import json
import logging
import re
from pathlib import Path
from typing import Dict, List, Set, Tuple

from slideinsight_sdk import SlideInsightClient, AuthenticationError, ValidationError
from rich.console import Console
from rich.progress import Progress, SpinnerColumn, TextColumn, BarColumn, TaskProgressColumn
from rich.prompt import Prompt, Confirm
from rich.table import Table
from rich.panel import Panel


console = Console()


def setup_logging(verbose: bool = False) -> None:
    """Setup logging configuration."""
    level = logging.DEBUG if verbose else logging.INFO
    logging.basicConfig(level=level, format="%(asctime)s - %(levelname)s - %(message)s", datefmt="%Y-%m-%d %H:%M:%S")


def humanize_study_name(folder_name: str) -> str:
    """Convert folder name to human-readable study name.

    Args:
        folder_name: Raw folder name like "diagnostic_breast" or "tissue_nos"

    Returns:
        Human-readable name like "TCGA Diagnostic Breast" or "TCGA Tissue NOS"
    """
    # Remove common prefixes
    name = folder_name.replace("gdc_manifest.2021-11-01_", "")
    name = name.replace(".txt", "")

    # Split into parts
    parts = name.split("_")

    # Capitalize each part
    capitalized_parts = []
    for part in parts:
        if part.lower() == "nos":
            capitalized_parts.append("NOS")
        elif part.lower() in ["and", "of", "in", "the"]:
            capitalized_parts.append(part.lower())
        else:
            capitalized_parts.append(part.capitalize())

    # Join with spaces and add TCGA prefix
    human_name = " ".join(capitalized_parts)
    return f"TCGA {human_name}"


def extract_case_id(filename: str) -> str:
    """Extract case ID from TCGA filename.

    Args:
        filename: TCGA filename like "TCGA-4K-AA1H-01A-01-TS1.7601120A-3053-4B7F-8589-463BD4F9D045.svs"

    Returns:
        Case ID like "TCGA-4K-AA1H"
    """
    # TCGA case IDs follow pattern: TCGA-XX-XXXX
    match = re.match(r"(TCGA-[A-Z0-9]{2}-[A-Z0-9]{4})", filename)
    if match:
        return match.group(1)

    # Fallback: try to extract anything that looks like TCGA-XX-XXXX
    match = re.search(r"TCGA-[A-Z0-9]{2}-[A-Z0-9]{4}", filename)
    if match:
        return match.group(0)

    # If no standard pattern found, use the part before first dot
    base_name = filename.split(".")[0]
    return base_name


def scan_tcga_folders(data_dir: Path) -> Dict[str, List[Tuple[str, Path]]]:
    """Scan TCGA data directory and organize files by study.

    Args:
        data_dir: Path to TCGA data directory

    Returns:
        Dictionary mapping study names to list of (case_id, file_path) tuples
    """
    studies_data = {}

    # Look for manifest files to determine studies
    manifest_files = list(data_dir.glob("gdc_manifest.2021-11-01_diagnostic_*.txt"))

    if not manifest_files:
        # Fallback: scan all subdirectories
        console.print("No manifest files found, scanning subdirectories...")
        subdirs = [d for d in data_dir.iterdir() if d.is_dir()]

        with Progress(
            SpinnerColumn(),
            TextColumn("[progress.description]{task.description}"),
            BarColumn(),
            TaskProgressColumn(),
            console=console,
        ) as progress:
            task = progress.add_task("Scanning directories...", total=len(subdirs))

            for subdir in subdirs:
                progress.update(task, description=f"Scanning: {subdir.name}")
                study_name = humanize_study_name(subdir.name)
                studies_data[study_name] = []

                # Find all .svs files in subdirectory
                svs_files = list(subdir.rglob("*.svs"))
                console.print(f"Found {len(svs_files)} .svs files in {subdir.name}")

                for svs_file in svs_files:
                    case_id = extract_case_id(svs_file.name)
                    studies_data[study_name].append((case_id, svs_file))

                progress.advance(task)
    else:
        # Use manifest files to determine study structure
        console.print(f"Found {len(manifest_files)} manifest files")

        with Progress(
            SpinnerColumn(),
            TextColumn("[progress.description]{task.description}"),
            BarColumn(),
            TaskProgressColumn(),
            console=console,
        ) as progress:
            task = progress.add_task("Processing manifest directories...", total=len(manifest_files))

            for manifest_file in manifest_files:
                progress.update(task, description=f"Processing: {manifest_file.name}")
                study_name = humanize_study_name(manifest_file.name)
                studies_data[study_name] = []

                # For now, we'll assume files are organized in subdirectories
                # matching the manifest names. In a real scenario, you might
                # need to parse the manifest files to get the actual file paths.
                study_dir = data_dir / manifest_file.stem
                if study_dir.exists():
                    svs_files = list(study_dir.rglob("*.svs"))
                    console.print(f"Found {len(svs_files)} .svs files in {study_dir.name}")

                    for svs_file in svs_files:
                        case_id = extract_case_id(svs_file.name)
                        studies_data[study_name].append((case_id, svs_file))
                else:
                    console.print(f"⚠️  Directory not found: {study_dir}", style="yellow")

                progress.advance(task)

    return studies_data


async def create_or_get_study(client: SlideInsightClient, study_name: str) -> str:
    """Create a study or get existing one by name.

    Args:
        client: SlideInsight client
        study_name: Name of the study

    Returns:
        Study UID
    """
    # Try to find existing study first
    try:
        studies = await client.studies.list(search=study_name, limit=10)
        for study in studies.items:
            if study.name == study_name:
                console.print(f"Found existing study: {study_name}")
                return study.study_uid
    except Exception as e:
        console.print(f"Error searching for existing study: {e}")

    # Create new study
    try:
        study = await client.studies.create(
            name=study_name, description=f"TCGA study for {study_name.replace('TCGA ', '')} samples", is_published=False
        )
        console.print(f"Created new study: {study_name}")
        return study.study_uid
    except Exception as e:
        console.print(f"Error creating study {study_name}: {e}")
        raise


async def create_or_get_case(client: SlideInsightClient, case_id: str) -> str:
    """Create a case or get existing one by name.

    Args:
        client: SlideInsight client
        case_id: Case identifier

    Returns:
        Case UID
    """
    # Try to find existing case first
    try:
        cases = await client.cases.list(search=case_id, limit=10)
        for case in cases.items:
            if case.name == case_id:
                return case.case_uid
    except Exception as e:
        console.print(f"Error searching for existing case: {e}")

    # Create new case
    try:
        case = await client.cases.create(name=case_id, metadata=json.dumps({"tcga_case_id": case_id, "source": "TCGA"}))
        return case.case_uid
    except Exception as e:
        console.print(f"Error creating case {case_id}: {e}")
        raise


async def process_study(
    client: SlideInsightClient, study_name: str, study_files: List[Tuple[str, Path]], dry_run: bool = False
) -> Tuple[int, int]:
    """Process a single study and its files.

    Args:
        client: SlideInsight client
        study_name: Name of the study
        study_files: List of (case_id, file_path) tuples
        dry_run: If True, don't make actual changes

    Returns:
        Tuple of (success_count, error_count)
    """
    if dry_run:
        console.print(f"[DRY RUN] Would process study: {study_name} with {len(study_files)} files")
        return len(study_files), 0

    success_count = 0
    error_count = 0

    try:
        # Create study
        study_uid = await create_or_get_study(client, study_name)

        # Group files by case ID
        cases_files: Dict[str, List[Path]] = {}
        for case_id, file_path in study_files:
            if case_id not in cases_files:
                cases_files[case_id] = []
            cases_files[case_id].append(file_path)

        # Process each case
        for case_id, case_files in cases_files.items():
            try:
                # Create case
                case_uid = await create_or_get_case(client, case_id)

                # Add case to study
                try:
                    await client.studies.add_case(study_uid, case_uid)
                except Exception as e:
                    # Case might already be in study, which is fine
                    if "already exists" not in str(e).lower():
                        console.print(f"Warning: Could not add case {case_id} to study: {e}")

                # Process each slide in the case
                for slide_file in case_files:
                    try:
                        slide_name = slide_file.stem  # Filename without extension
                        slide_uri = str(slide_file.absolute())

                        # Add slide to case
                        slide = await client.cases.add_slide(
                            case_uid=case_uid, slide_uri=slide_uri, slide_name=slide_name
                        )

                        success_count += 1
                        console.print(f"✅ Added slide {slide_name} to case {case_id}")

                    except Exception as e:
                        error_count += 1
                        console.print(f"❌ Error adding slide {slide_file.name}: {e}")

            except Exception as e:
                error_count += len(case_files)
                console.print(f"❌ Error processing case {case_id}: {e}")

    except Exception as e:
        error_count += len(study_files)
        console.print(f"❌ Error processing study {study_name}: {e}")

    return success_count, error_count


async def main():
    """Main function."""
    parser = argparse.ArgumentParser(
        description="Bulk import TCGA data into SlideInsight", formatter_class=argparse.RawDescriptionHelpFormatter
    )

    parser.add_argument("--url", required=True, help="SlideInsight server URL")
    parser.add_argument("--email", help="Username for authentication")
    parser.add_argument("--password", help="Password for authentication")
    parser.add_argument("--data-dir", required=True, type=Path, help="Path to TCGA data directory")
    parser.add_argument("--dry-run", action="store_true", help="Show what would be done without making changes")
    parser.add_argument("--verbose", "-v", action="store_true", help="Enable verbose logging")
    parser.add_argument("--limit-studies", type=int, help="Limit number of studies to process (for testing)")

    args = parser.parse_args()

    setup_logging(args.verbose)

    # Validate data directory
    if not args.data_dir.exists():
        console.print(f"❌ Data directory not found: {args.data_dir}", style="red")
        return 1

    # Get credentials
    email = args.email or Prompt.ask("Username")
    password = args.password or Prompt.ask("Password", password=True)

    console.print("🔬 TCGA Bulk Import Tool")
    console.print("=" * 50)

    # Scan TCGA data
    console.print(f"📂 Scanning TCGA data in: {args.data_dir}")
    studies_data = scan_tcga_folders(args.data_dir)

    if not studies_data:
        console.print("❌ No TCGA data found", style="red")
        return 1

    # Limit studies if requested
    if args.limit_studies:
        studies_items = list(studies_data.items())[: args.limit_studies]
        studies_data = dict(studies_items)

    # Show preview
    table = Table(title="Studies to Process")
    table.add_column("Study Name", style="green")
    table.add_column("Cases", style="cyan", justify="right")
    table.add_column("Slides", style="yellow", justify="right")

    total_slides = 0
    for study_name, files in studies_data.items():
        case_ids = set(case_id for case_id, _ in files)
        slide_count = len(files)
        total_slides += slide_count

        table.add_row(study_name, str(len(case_ids)), str(slide_count))

    console.print(table)
    console.print(f"\nTotal: {len(studies_data)} studies, {total_slides} slides")

    if args.dry_run:
        console.print("🧪 [bold yellow]DRY RUN MODE[/bold yellow] - No changes will be made")
        return 0

    if not Confirm.ask(f"Continue with import?"):
        console.print("❌ Operation cancelled")
        return 0

    # Process studies
    try:
        async with SlideInsightClient(args.url) as client:
            await client.login(email, password)
            console.print("✅ Successfully authenticated")

            total_success = 0
            total_errors = 0

            with Progress(
                SpinnerColumn(),
                TextColumn("[progress.description]{task.description}"),
                BarColumn(),
                TaskProgressColumn(),
                console=console,
            ) as progress:
                task = progress.add_task("Processing studies...", total=len(studies_data))

                for study_name, study_files in studies_data.items():
                    progress.update(task, description=f"Processing: {study_name}")

                    success_count, error_count = await process_study(client, study_name, study_files, args.dry_run)

                    total_success += success_count
                    total_errors += error_count

                    progress.advance(task)

            # Final summary
            console.print("\n" + "=" * 50)
            console.print("[bold]📊 FINAL SUMMARY[/bold]")
            console.print("=" * 50)
            console.print(f"Studies processed: {len(studies_data)}")
            console.print(f"✅ Slides imported successfully: {total_success}")
            console.print(f"❌ Slides failed: {total_errors}")

            if total_errors > 0:
                console.print(f"\n⚠️ {total_errors} slides failed to import", style="yellow")
                return 1
            else:
                console.print("\n🎉 All slides imported successfully!", style="green")
                return 0

    except AuthenticationError as e:
        console.print(f"❌ Authentication failed: {e.message}", style="red")
        return 1
    except Exception as e:
        console.print(f"❌ Error: {e}", style="red")
        return 1


if __name__ == "__main__":
    import sys

    sys.exit(asyncio.run(main()))
