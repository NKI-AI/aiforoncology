#!/usr/bin/env python3
"""
TCGA Mapping-Based Import Script for SlideInsight

This script uses the TCGA identifier mapping JSON file to import data into SlideInsight.
It reads the mapping file to determine the actual file paths and organizes them by
the manifest categories.

Usage:
    python tcga_mapping_import.py --url http://localhost:3000 --mapping /data/groups/public/archive/TCGA/identifier_mapping.json --data-dir /data/groups/public/archive/TCGA
"""

import argparse
import asyncio
import json
import logging
import re
from pathlib import Path
from typing import Dict, List, Set, Tuple

from slideinsight_sdk import SlideInsightClient, AuthenticationError
from rich.console import Console
from rich.progress import Progress, SpinnerColumn, TextColumn, BarColumn, TaskProgressColumn
from rich.prompt import Prompt, Confirm
from rich.table import Table


console = Console()


def setup_logging(verbose: bool = False) -> None:
    """Setup logging configuration."""
    level = logging.DEBUG if verbose else logging.INFO
    logging.basicConfig(level=level, format="%(asctime)s - %(levelname)s - %(message)s", datefmt="%Y-%m-%d %H:%M:%S")


def humanize_study_name(manifest_name: str) -> str:
    """Convert manifest name to human-readable study name.

    Args:
        manifest_name: Manifest name like "gdc_manifest.2021-11-01_diagnostic_breast.txt"

    Returns:
        Human-readable name like "TCGA Diagnostic Breast"
    """
    # Extract the anatomical part from manifest name
    match = re.search(r"gdc_manifest\.2021-11-01_(diagnostic|tissue)_(.+)\.txt", manifest_name)
    if not match:
        return f"TCGA {manifest_name}"

    category, anatomy = match.groups()

    # Split anatomy into parts and capitalize
    parts = anatomy.split("_")
    capitalized_parts = []

    for part in parts:
        if part.lower() == "nos":
            capitalized_parts.append("NOS")
        elif part.lower() in ["and", "of", "in", "the", "or"]:
            capitalized_parts.append(part.lower())
        else:
            capitalized_parts.append(part.capitalize())

    # Combine category and anatomy
    human_anatomy = " ".join(capitalized_parts)
    human_category = category.capitalize()

    return f"TCGA {human_category} {human_anatomy}"


def extract_case_id(slide_identifier: str) -> str:
    """Extract case ID from TCGA slide identifier.

    Args:
        slide_identifier: TCGA identifier like "TCGA-4K-AA1H-01A-01-TS1.7601120A-3053-4B7F-8589-463BD4F9D045"

    Returns:
        Case ID like "TCGA-4K-AA1H"
    """
    # TCGA case IDs follow pattern: TCGA-XX-XXXX
    match = re.match(r"(TCGA-[A-Z0-9]{2}-[A-Z0-9]{4})", slide_identifier)
    if match:
        return match.group(1)

    # Fallback: try to construct from first three parts
    parts = slide_identifier.split("-")
    if len(parts) >= 3:
        return "-".join(parts[:3])  # TCGA-XX-XXXX format

    # Final fallback - return the identifier as-is
    return slide_identifier


def parse_mapping_file(
    mapping_file: Path, data_dir: Path, check_files: bool = True
) -> Dict[str, List[Tuple[str, str, Path]]]:
    """Parse the TCGA identifier mapping file and organize by studies.

    Args:
        mapping_file: Path to identifier_mapping.json
        data_dir: Base data directory
        check_files: Whether to check if files exist (can be slow for large datasets)

    Returns:
        Dictionary mapping study names to list of (case_id, slide_id, file_path) tuples
    """
    console.print(f"📖 Reading mapping file: {mapping_file}")

    with open(mapping_file, "r") as f:
        mapping_data = json.load(f)

    console.print(f"Found {len(mapping_data)} entries in mapping file")

    studies_data = {}
    files_checked = 0
    files_found = 0
    files_missing = 0

    # Create progress bar for large files
    with Progress(
        SpinnerColumn(),
        TextColumn("[progress.description]{task.description}"),
        BarColumn(),
        TaskProgressColumn(),
        console=console,
    ) as progress:
        task = progress.add_task("Processing mapping entries...", total=len(mapping_data))

        for slide_identifier, relative_path in mapping_data.items():
            # Extract manifest name from relative path
            path_parts = relative_path.split("/")
            if len(path_parts) < 1:
                progress.advance(task)
                continue

            manifest_name = path_parts[0]

            # Generate study name from manifest
            study_name = humanize_study_name(manifest_name)

            if study_name not in studies_data:
                studies_data[study_name] = []

            # Extract case ID
            case_id = extract_case_id(slide_identifier)

            # Build full file path
            full_path = data_dir / relative_path

            # Optionally check if file exists
            if check_files:
                if full_path.exists():
                    studies_data[study_name].append((case_id, slide_identifier, full_path))
                    files_found += 1
                else:
                    files_missing += 1
                    if files_missing <= 10:  # Only show first 10 missing files
                        console.print(f"⚠️  File not found: {relative_path}", style="yellow")
                    elif files_missing == 11:
                        console.print("⚠️  ... (suppressing further missing file warnings)", style="yellow")
                files_checked += 1
            else:
                # Add all entries without checking existence
                studies_data[study_name].append((case_id, slide_identifier, full_path))
                files_found += 1

            progress.advance(task)

    if check_files:
        console.print(f"File check results: {files_found} found, {files_missing} missing out of {files_checked} total")
    else:
        console.print(f"Added {files_found} entries (file existence not checked)")

    return studies_data


async def create_or_get_study(client: SlideInsightClient, study_name: str) -> str:
    """Create a study or get existing one by name."""
    try:
        studies = await client.studies.list(search=study_name, limit=10)
        for study in studies.items:
            if study.name == study_name:
                console.print(f"Found existing study: {study_name}")
                return study.study_uid
    except Exception as e:
        console.print(f"Error searching for existing study: {e}")

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
    """Create a case or get existing one by name."""
    try:
        cases = await client.cases.list(search=case_id, limit=10)
        for case in cases.items:
            if case.name == case_id:
                return case.case_uid
    except Exception as e:
        console.print(f"Error searching for existing case: {e}")

    try:
        case = await client.cases.create(name=case_id, metadata=json.dumps({"tcga_case_id": case_id, "source": "TCGA"}))
        return case.case_uid
    except Exception as e:
        console.print(f"Error creating case {case_id}: {e}")
        raise


async def process_study(
    client: SlideInsightClient, study_name: str, study_slides: List[Tuple[str, str, Path]], dry_run: bool = False
) -> Tuple[int, int]:
    """Process a single study and its slides."""
    if dry_run:
        console.print(f"[DRY RUN] Would process study: {study_name} with {len(study_slides)} slides")
        return len(study_slides), 0

    success_count = 0
    error_count = 0

    try:
        # Create study
        study_uid = await create_or_get_study(client, study_name)

        # Group slides by case ID
        cases_slides: Dict[str, List[Tuple[str, Path]]] = {}
        for case_id, slide_id, file_path in study_slides:
            if case_id not in cases_slides:
                cases_slides[case_id] = []
            cases_slides[case_id].append((slide_id, file_path))

        # Process each case
        for case_id, case_slides in cases_slides.items():
            try:
                # Create case
                case_uid = await create_or_get_case(client, case_id)

                # Add case to study
                try:
                    await client.studies.add_case(study_uid, case_uid)
                except Exception as e:
                    if "already exists" not in str(e).lower():
                        console.print(f"Warning: Could not add case {case_id} to study: {e}")

                # Process each slide in the case
                for slide_id, slide_file in case_slides:
                    try:
                        slide_name = slide_id  # Use the full TCGA identifier as slide name
                        slide_uri = str(slide_file.absolute())

                        # Add slide to case
                        slide = await client.cases.add_slide(
                            case_uid=case_uid, slide_uri=slide_uri, slide_name=slide_name, slide_id=slide_id
                        )

                        success_count += 1
                        console.print(f"✅ Added slide {slide_id} to case {case_id}")

                    except Exception as e:
                        error_count += 1
                        console.print(f"❌ Error adding slide {slide_id}: {e}")

            except Exception as e:
                error_count += len(case_slides)
                console.print(f"❌ Error processing case {case_id}: {e}")

    except Exception as e:
        error_count += len(study_slides)
        console.print(f"❌ Error processing study {study_name}: {e}")

    return success_count, error_count


async def main():
    """Main function."""
    parser = argparse.ArgumentParser(
        description="Import TCGA data using identifier mapping file",
        formatter_class=argparse.RawDescriptionHelpFormatter,
    )

    parser.add_argument("--url", required=True, help="SlideInsight server URL")
    parser.add_argument("--email", help="Username for authentication")
    parser.add_argument("--password", help="Password for authentication")
    parser.add_argument("--mapping", required=True, type=Path, help="Path to identifier_mapping.json file")
    parser.add_argument("--data-dir", required=True, type=Path, help="Base path to TCGA data directory")
    parser.add_argument("--dry-run", action="store_true", help="Show what would be done without making changes")
    parser.add_argument("--verbose", "-v", action="store_true", help="Enable verbose logging")
    parser.add_argument("--limit-studies", type=int, help="Limit number of studies to process (for testing)")
    parser.add_argument("--study-filter", help="Only process studies matching this pattern")
    parser.add_argument(
        "--skip-file-check",
        action="store_true",
        help="Skip checking if files exist (faster but may include missing files)",
    )

    args = parser.parse_args()

    setup_logging(args.verbose)

    # Validate files
    if not args.mapping.exists():
        console.print(f"❌ Mapping file not found: {args.mapping}", style="red")
        return 1

    if not args.data_dir.exists():
        console.print(f"❌ Data directory not found: {args.data_dir}", style="red")
        return 1

    # Get credentials
    email = args.email or Prompt.ask("Username")
    password = args.password or Prompt.ask("Password", password=True)

    console.print("🔬 TCGA Mapping-Based Import Tool")
    console.print("=" * 50)

    # Parse mapping file
    studies_data = parse_mapping_file(args.mapping, args.data_dir, check_files=not args.skip_file_check)

    if not studies_data:
        console.print("❌ No TCGA data found in mapping file", style="red")
        return 1

    # Apply study filter if provided
    if args.study_filter:
        filtered_studies = {}
        for study_name, slides in studies_data.items():
            if args.study_filter.lower() in study_name.lower():
                filtered_studies[study_name] = slides
        studies_data = filtered_studies
        console.print(f"Filtered to {len(studies_data)} studies matching '{args.study_filter}'")

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
    for study_name, slides in studies_data.items():
        case_ids = set(case_id for case_id, _, _ in slides)
        slide_count = len(slides)
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

                for study_name, study_slides in studies_data.items():
                    progress.update(task, description=f"Processing: {study_name}")

                    success_count, error_count = await process_study(client, study_name, study_slides, args.dry_run)

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
