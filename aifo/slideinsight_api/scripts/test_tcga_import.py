#!/usr/bin/env python3
"""
Test script for TCGA import functionality.

This script demonstrates how to use the TCGA import scripts and validates
the functionality with sample data.
"""

import asyncio
import json
import tempfile
from pathlib import Path

from slideinsight_sdk import SlideInsightClient
from rich.console import Console

console = Console()


async def test_study_name_generation():
    """Test the study name humanization function."""
    from tcga_bulk_import import humanize_study_name

    test_cases = [
        ("diagnostic_breast", "TCGA Diagnostic Breast"),
        (
            "diagnostic_bones_joints_and_articular_cartilage_of_limbs",
            "TCGA Diagnostic Bones Joints and Articular Cartilage of Limbs",
        ),
        ("tissue_nos", "TCGA Tissue NOS"),
        ("diagnostic_other_and_unspecified_parts_of_mouth", "TCGA Diagnostic Other and Unspecified Parts of Mouth"),
        ("gdc_manifest.2021-11-01_diagnostic_brain.txt", "TCGA Diagnostic Brain"),
    ]

    console.print("🧪 Testing study name generation...")

    for input_name, expected in test_cases:
        result = humanize_study_name(input_name)
        status = "✅" if result == expected else "❌"
        console.print(f"{status} {input_name} → {result}")
        if result != expected:
            console.print(f"   Expected: {expected}")


async def test_case_id_extraction():
    """Test the case ID extraction function."""
    from tcga_bulk_import import extract_case_id

    test_cases = [
        ("TCGA-4K-AA1H-01A-01-TS1.7601120A-3053-4B7F-8589-463BD4F9D045.svs", "TCGA-4K-AA1H"),
        ("TCGA-2X-A9D5-01A-01-TS1.C19455E8-D6A5-4BC3-9EC5-E8F22DA7B6BC.svs", "TCGA-2X-A9D5"),
        ("TCGA-02-0001-01A-01-TS1.svs", "TCGA-02-0001"),
        ("TCGA-AA-3502-01A-01-TS1.svs", "TCGA-AA-3502"),
    ]

    console.print("\n🧪 Testing case ID extraction...")

    for filename, expected in test_cases:
        result = extract_case_id(filename)
        status = "✅" if result == expected else "❌"
        console.print(f"{status} {filename} → {result}")
        if result != expected:
            console.print(f"   Expected: {expected}")


async def create_sample_data():
    """Create sample TCGA data structure for testing."""
    with tempfile.TemporaryDirectory() as temp_dir:
        base_dir = Path(temp_dir)

        # Create sample directory structure
        breast_dir = base_dir / "gdc_manifest.2021-11-01_diagnostic_breast.txt"
        brain_dir = base_dir / "gdc_manifest.2021-11-01_diagnostic_brain.txt"

        breast_dir.mkdir()
        brain_dir.mkdir()

        # Create sample files (empty for testing)
        sample_files = [
            breast_dir / "TCGA-A1-A0SB-01A-01-TS1.svs",
            breast_dir / "TCGA-A1-A0SD-01A-01-TS1.svs",
            brain_dir / "TCGA-02-0001-01A-01-TS1.svs",
            brain_dir / "TCGA-02-0003-01A-01-TS1.svs",
        ]

        for file_path in sample_files:
            file_path.write_text("# Sample TCGA file for testing")

        # Create sample mapping file
        mapping_data = {
            "TCGA-A1-A0SB-01A-01-TS1.ABC123": "gdc_manifest.2021-11-01_diagnostic_breast.txt/TCGA-A1-A0SB-01A-01-TS1.svs",
            "TCGA-A1-A0SD-01A-01-TS1.DEF456": "gdc_manifest.2021-11-01_diagnostic_breast.txt/TCGA-A1-A0SD-01A-01-TS1.svs",
            "TCGA-02-0001-01A-01-TS1.GHI789": "gdc_manifest.2021-11-01_diagnostic_brain.txt/TCGA-02-0001-01A-01-TS1.svs",
            "TCGA-02-0003-01A-01-TS1.JKL012": "gdc_manifest.2021-11-01_diagnostic_brain.txt/TCGA-02-0003-01A-01-TS1.svs",
        }

        mapping_file = base_dir / "identifier_mapping.json"
        mapping_file.write_text(json.dumps(mapping_data, indent=2))

        console.print(f"\n📁 Created sample data in: {base_dir}")
        console.print("Files created:")
        for file_path in sample_files:
            console.print(f"  - {file_path.relative_to(base_dir)}")
        console.print(f"  - {mapping_file.name}")

        return base_dir, mapping_file


async def test_directory_scanning():
    """Test the directory scanning functionality."""
    from tcga_bulk_import import scan_tcga_folders

    console.print("\n🧪 Testing directory scanning...")

    # Create temporary test data
    base_dir, _ = await create_sample_data()

    try:
        studies_data = scan_tcga_folders(base_dir)

        console.print(f"Found {len(studies_data)} studies:")
        for study_name, files in studies_data.items():
            case_count = len(set(case_id for case_id, _ in files))
            console.print(f"  - {study_name}: {case_count} cases, {len(files)} slides")

        # Validate expected structure
        expected_studies = {"TCGA Diagnostic Breast", "TCGA Diagnostic Brain"}
        found_studies = set(studies_data.keys())

        if expected_studies.issubset(found_studies):
            console.print("✅ Directory scanning successful")
        else:
            console.print("❌ Directory scanning failed")
            console.print(f"Expected: {expected_studies}")
            console.print(f"Found: {found_studies}")

    except Exception as e:
        console.print(f"❌ Directory scanning failed: {e}")


async def test_mapping_parsing():
    """Test the mapping file parsing functionality."""
    from tcga_mapping_import import parse_mapping_file

    console.print("\n🧪 Testing mapping file parsing...")

    # Create temporary test data
    base_dir, mapping_file = await create_sample_data()

    try:
        studies_data = parse_mapping_file(mapping_file, base_dir)

        console.print(f"Found {len(studies_data)} studies:")
        for study_name, slides in studies_data.items():
            case_count = len(set(case_id for case_id, _, _ in slides))
            console.print(f"  - {study_name}: {case_count} cases, {len(slides)} slides")

        # Validate expected structure
        expected_studies = {"TCGA Diagnostic Breast", "TCGA Diagnostic Brain"}
        found_studies = set(studies_data.keys())

        if expected_studies.issubset(found_studies):
            console.print("✅ Mapping file parsing successful")
        else:
            console.print("❌ Mapping file parsing failed")
            console.print(f"Expected: {expected_studies}")
            console.print(f"Found: {found_studies}")

    except Exception as e:
        console.print(f"❌ Mapping file parsing failed: {e}")


async def demonstrate_dry_run():
    """Demonstrate how to use the scripts in dry-run mode."""
    console.print("\n🚀 Example dry-run commands:")
    console.print("=" * 50)

    console.print("\n1. Directory-based import (dry-run):")
    console.print("python tcga_bulk_import.py \\")
    console.print("    --url http://localhost:3000 \\")
    console.print("    --data-dir /data/groups/public/archive/TCGA/images \\")
    console.print("    --limit-studies 3 \\")
    console.print("    --dry-run")

    console.print("\n2. Mapping-based import (dry-run):")
    console.print("python tcga_mapping_import.py \\")
    console.print("    --url http://localhost:3000 \\")
    console.print("    --mapping /data/groups/public/archive/TCGA/identifier_mapping.json \\")
    console.print("    --data-dir /data/groups/public/archive/TCGA \\")
    console.print("    --study-filter breast \\")
    console.print("    --dry-run")

    console.print("\n3. Test with limited data:")
    console.print("python tcga_bulk_import.py \\")
    console.print("    --url http://localhost:3000 \\")
    console.print("    --data-dir /path/to/sample/data \\")
    console.print("    --limit-studies 1 \\")
    console.print("    --verbose \\")
    console.print("    --dry-run")


async def main():
    """Run all tests."""
    console.print("🔬 TCGA Import Script Tests")
    console.print("=" * 50)

    # Test individual functions
    await test_study_name_generation()
    await test_case_id_extraction()

    # Test data processing
    await test_directory_scanning()
    await test_mapping_parsing()

    # Show usage examples
    await demonstrate_dry_run()

    console.print("\n" + "=" * 50)
    console.print("✅ All tests completed!")
    console.print("\nNext steps:")
    console.print("1. Run the actual import scripts with --dry-run first")
    console.print("2. Verify the preview output looks correct")
    console.print("3. Remove --dry-run to perform the actual import")


if __name__ == "__main__":
    asyncio.run(main())
