#!/usr/bin/env python3
"""
SlideInsight CLI Tool

Modern command-line interface for SlideInsight operations using the Python SDK.
Provides bulk import functionality and other common operations.

Environment Variables:
    SLIDEINSIGHT_API_URL: Default SlideInsight server URL
    SLIDEINSIGHT_USERNAME: Default email (optional)
    SLIDEINSIGHT_PASSWORD: Default password (optional, not recommended for security)
"""

import asyncio
import csv
import getpass
import logging
import os
from pathlib import Path
from typing import Optional
import functools

import click
from rich.console import Console
from rich.logging import RichHandler
from rich.progress import Progress, SpinnerColumn, TextColumn, BarColumn, TaskProgressColumn
from rich.table import Table
from rich.prompt import Prompt, Confirm
from rich.panel import Panel
from rich.text import Text

from slideinsight_sdk import SlideInsightClient, AuthenticationError, ValidationError


# Setup rich console and logging
console = Console()


def setup_logging(verbose: bool = False) -> None:
    """Setup logging with rich formatting."""
    level = logging.DEBUG if verbose else logging.INFO
    logging.basicConfig(
        level=level, format="%(message)s", datefmt="[%X]", handlers=[RichHandler(console=console, rich_tracebacks=True)]
    )


def get_credentials(
    url: Optional[str] = None, email: Optional[str] = None, password: Optional[str] = None
) -> tuple[str, str, str]:
    """Get credentials from arguments, environment variables, or interactive prompts.

    Args:
        url: URL from command line arguments
        email: Username from command line arguments
        password: Password from command line arguments

    Returns:
        Tuple of (url, email, password)
    """
    # Get URL - check argument, then env var, then prompt
    if not url:
        url = os.getenv("SLIDEINSIGHT_API_URL")
        if url:
            console.print(f"🌐 Using API URL from environment: {url}")
        else:
            url = Prompt.ask("SlideInsight server URL")

    # Get email - check argument, then env var, then prompt
    if not email:
        email = os.getenv("SLIDEINSIGHT_USERNAME")
        if email:
            console.print(f"👤 Using email from environment: {email}")
        else:
            email = Prompt.ask("Username")

    # Get password - check argument, then env var, then prompt
    if not password:
        password = os.getenv("SLIDEINSIGHT_PASSWORD")
        if password:
            console.print("🔑 Using password from environment")
        else:
            password = Prompt.ask("Password", password=True)

    return url, email, password


def _is_vector_annotation(file_uri: str) -> bool:
    """Check if the file is a vector annotation based on extension.

    Args:
        file_uri: Path or URI to the annotation file

    Returns:
        True if it's a vector annotation, False if raster
    """
    if not file_uri:
        return False

    file_path = Path(file_uri)
    extension = file_path.suffix.lower()

    # Vector annotation formats
    vector_extensions = {".json", ".geojson"}

    return extension in vector_extensions


def _get_annotation_type_display(file_uri: Optional[str]) -> str:
    """Get display string for annotation type.

    Args:
        file_uri: Path or URI to the annotation file

    Returns:
        Display string for the annotation type
    """
    if not file_uri:
        return "[dim]None[/dim]"

    if _is_vector_annotation(file_uri):
        return f"[cyan]Vector[/cyan]: {file_uri}"
    else:
        return f"[yellow]Raster[/yellow]: {file_uri}"


def async_command(f):
    """Decorator to make click commands work with async functions."""

    @functools.wraps(f)
    def wrapper(*args, **kwargs):
        return asyncio.run(f(*args, **kwargs))

    return wrapper


@click.group()
@click.option("--verbose", "-v", is_flag=True, help="Enable verbose logging")
@click.option("--debug", is_flag=True, help="Enable debug logging with HTTP request details")
@click.pass_context
def cli(ctx, verbose: bool, debug: bool):
    """SlideInsight CLI - Modern command-line interface for SlideInsight operations.

    Environment Variables:
        SLIDEINSIGHT_API_URL: Default SlideInsight server URL
        SLIDEINSIGHT_USERNAME: Default email (optional)
        SLIDEINSIGHT_PASSWORD: Default password (optional, not recommended)
    """
    # Enable debug logging if debug flag is set (this includes verbose logging)
    setup_logging(verbose or debug)
    ctx.ensure_object(dict)
    ctx.obj["verbose"] = verbose
    ctx.obj["debug"] = debug


@cli.command()
@click.option("--url", help="SlideInsight server URL (or set SLIDEINSIGHT_API_URL)")
@click.option("--email", help="Username for authentication (or set SLIDEINSIGHT_USERNAME)")
@click.option("--password", help="Password for authentication (or set SLIDEINSIGHT_PASSWORD)")
@async_command
async def login(ctx, url: Optional[str], email: Optional[str], password: Optional[str]):
    """Test login to SlideInsight server."""
    try:
        url, email, password = get_credentials(url, email, password)
        debug = ctx.obj.get("debug", False)

        async with SlideInsightClient(url, debug=debug) as client:
            await client.login(email, password)
            user = await client.get_current_user()

            console.print(f"✅ Successfully authenticated as [bold green]{user.email}[/bold green]")
            console.print(f"Email: {user.email}")
            console.print(f"Active: {user.is_active}")

    except AuthenticationError as e:
        console.print(f"❌ Authentication failed: {e.message}", style="red")
        raise click.Abort()
    except Exception as e:
        console.print(f"❌ Error: {e}", style="red")
        raise click.Abort()


@cli.command()
@click.option("--url", help="SlideInsight server URL (or set SLIDEINSIGHT_API_URL)")
@click.option("--email", help="Username for authentication (or set SLIDEINSIGHT_USERNAME)")
@click.option("--password", help="Password for authentication (or set SLIDEINSIGHT_PASSWORD)")
@click.option("--study-id", help="Study ID to add cases to")
@click.option("--csv", "csv_file", type=click.Path(exists=True), help="CSV file with case and slide data")
@click.option("--dry-run", is_flag=True, help="Show what would be done without making changes")
@click.pass_context
@async_command
async def bulk_import(
    ctx,
    url: Optional[str],
    email: Optional[str],
    password: Optional[str],
    study_id: Optional[str],
    csv_file: Optional[str],
    dry_run: bool,
):
    """Bulk import cases and slides from CSV file.

    The CSV file should contain 2-3 columns: casename,slide_uri[,annotation_uri]
    - casename: Human-readable name for the case/slide
    - slide_uri: Path or URI to the slide file
    - annotation_uri: (Optional) Path or URI to the annotation file
      * .tiff, .png files will be imported as raster annotations (masks)
      * .json, .geojson files will be imported as vector annotations
    """
    # Get credentials from env vars or prompts
    url, email, password = get_credentials(url, email, password)
    debug = ctx.obj.get("debug", False)

    if not study_id:
        study_id = Prompt.ask("Study ID to add cases to")

    if not csv_file:
        csv_file = Prompt.ask("CSV file path")

    csv_path = Path(csv_file)
    if not csv_path.exists():
        console.print(f"❌ CSV file not found: {csv_path}", style="red")
        raise click.Abort()

    # Read and validate CSV
    console.print(f"📖 Reading CSV file: {csv_path}")
    rows = read_csv_file(csv_path)

    if not rows:
        console.print("❌ No valid rows found in CSV file", style="red")
        raise click.Abort()

    console.print(f"Found {len(rows)} rows to process")

    # Show preview
    table = Table(title="Import Preview")
    table.add_column("Row", style="cyan", no_wrap=True)
    table.add_column("Case Name", style="green")
    table.add_column("Slide URI", style="blue")
    table.add_column("Annotation", style="magenta", max_width=40)

    for i, (slide_uri, case_name, annotation_uri) in enumerate(rows[:5], 1):
        table.add_row(str(i), case_name, slide_uri, _get_annotation_type_display(annotation_uri))

    if len(rows) > 5:
        table.add_row("...", "...", "...", "...")

    console.print(table)

    if dry_run:
        console.print("🧪 [bold yellow]DRY RUN MODE[/bold yellow] - No changes will be made")
        return

    if not Confirm.ask(f"Continue with import of {len(rows)} rows?"):
        console.print("❌ Operation cancelled")
        raise click.Abort()

    # Perform the import
    try:
        async with SlideInsightClient(url, debug=debug) as client:
            await client.login(email, password)

            # Verify study exists
            console.print(f"🔍 Verifying study: {study_id}")
            study = await client.studies.get(study_id)
            console.print(f"✅ Found study: [bold green]{study.name}[/bold green]")

            # Process rows with progress bar
            success_count = 0
            error_count = 0

            with Progress(
                SpinnerColumn(),
                TextColumn("[progress.description]{task.description}"),
                BarColumn(),
                TaskProgressColumn(),
                console=console,
            ) as progress:
                task = progress.add_task("Processing rows...", total=len(rows))

                for i, (slide_uri, case_name, annotation_uri) in enumerate(rows, 1):
                    progress.update(task, description=f"Processing: {case_name}")

                    try:
                        await process_row(client, study_id, slide_uri, case_name, annotation_uri)
                        success_count += 1
                        progress.console.print(f"✅ {i:3d}. {case_name}", style="green")

                    except Exception as e:
                        error_count += 1
                        progress.console.print(f"❌ {i:3d}. {case_name}: {e}", style="red")

                        # Ask to continue on errors
                        if error_count < 3:
                            if not Confirm.ask("Continue processing remaining rows?"):
                                break

                    progress.advance(task)

            # Final summary
            console.print("\n" + "=" * 50)
            console.print("[bold]📊 FINAL SUMMARY[/bold]")
            console.print("=" * 50)
            console.print(f"Total rows: {len(rows)}")
            console.print(f"✅ Successful: {success_count}")
            console.print(f"❌ Failed: {error_count}")

            if error_count > 0:
                console.print(f"\n⚠️ {error_count} row(s) failed to process", style="yellow")
            else:
                console.print("\n🎉 All rows processed successfully!", style="green")

    except AuthenticationError as e:
        console.print(f"❌ Authentication failed: {e.message}", style="red")
        raise click.Abort()
    except Exception as e:
        console.print(f"❌ Error: {e}", style="red")
        raise click.Abort()


@cli.command()
@click.option("--url", help="SlideInsight server URL (or set SLIDEINSIGHT_API_URL)")
@click.option("--email", help="Username for authentication (or set SLIDEINSIGHT_USERNAME)")
@click.option("--password", help="Password for authentication (or set SLIDEINSIGHT_PASSWORD)")
@click.option("--page-size", default=10, help="Number of items per page")
@click.option("--search", help="Search query")
@async_command
async def list_studies(
    url: Optional[str], email: Optional[str], password: Optional[str], page_size: int, search: Optional[str]
):
    """List studies with interactive pagination."""
    url, email, password = get_credentials(url, email, password)
    debug = ctx.obj.get("debug", False)

    try:
        async with SlideInsightClient(url, debug=debug) as client:
            await client.login(email, password)

            current_page = 1

            while True:
                # Fetch current page
                response = await client.studies.list(page=current_page, limit=page_size, search=search)

                if not response.items:
                    console.print("No studies found")
                    return

                # Display studies table
                table = Table(title=f"Studies (Page {response.pagination.page} of {response.pagination.total_pages})")
                table.add_column("UID", style="cyan", no_wrap=True, max_width=12)
                table.add_column("Name", style="green", max_width=30)
                table.add_column("Description", style="blue", max_width=40)
                table.add_column("Published", style="yellow", justify="center")
                table.add_column("Created", style="magenta", max_width=20)

                for study in response.items:
                    table.add_row(
                        study.study_uid[:10] + "..." if len(study.study_uid) > 10 else study.study_uid,
                        study.name,
                        (study.description[:37] + "...") if len(study.description) > 40 else study.description,
                        "✅" if study.is_published else "❌",
                        study.created_at.split("T")[0] if study.created_at else "N/A",
                    )

                console.print(table)

                # Display pagination info
                pagination_info = Panel(
                    f"Total: {response.pagination.total} studies | "
                    f"Page {response.pagination.page} of {response.pagination.total_pages} | "
                    f"Showing {len(response.items)} items"
                )
                console.print(pagination_info)

                # Navigation options
                options = []
                if response.pagination.has_prev:
                    options.append("p")
                if response.pagination.has_next:
                    options.append("n")
                options.extend(["r", "q"])

                nav_text = []
                if response.pagination.has_prev:
                    nav_text.append("[bold cyan]p[/bold cyan]revious")
                if response.pagination.has_next:
                    nav_text.append("[bold cyan]n[/bold cyan]ext")
                nav_text.extend(["[bold cyan]r[/bold cyan]efresh", "[bold cyan]q[/bold cyan]uit"])

                console.print(f"\nNavigation: {' | '.join(nav_text)}")

                # Get user choice
                choice = Prompt.ask("Choose an option", choices=options, default="q")

                if choice == "n" and response.pagination.has_next:
                    current_page += 1
                elif choice == "p" and response.pagination.has_prev:
                    current_page -= 1
                elif choice == "r":
                    # Refresh current page
                    pass
                elif choice == "q":
                    break

                console.clear()

    except Exception as e:
        console.print(f"❌ Error: {e}", style="red")
        raise click.Abort()


@cli.command()
@click.option("--url", help="SlideInsight server URL (or set SLIDEINSIGHT_API_URL)")
@click.option("--email", help="Username for authentication (or set SLIDEINSIGHT_USERNAME)")
@click.option("--password", help="Password for authentication (or set SLIDEINSIGHT_PASSWORD)")
@click.option("--page-size", default=10, help="Number of items per page")
@click.option("--search", help="Search query")
@async_command
async def list_users(
    url: Optional[str], email: Optional[str], password: Optional[str], page_size: int, search: Optional[str]
):
    """List users with interactive pagination."""
    url, email, password = get_credentials(url, email, password)
    debug = ctx.obj.get("debug", False)

    try:
        async with SlideInsightClient(url, debug=debug) as client:
            await client.login(email, password)

            current_page = 1

            while True:
                # Fetch current page
                response = await client.users.list(page=current_page, limit=page_size, search=search)

                if not response.items:
                    console.print("No users found")
                    return

                # Display users table
                table = Table(title=f"Users (Page {response.pagination.page} of {response.pagination.total_pages})")
                table.add_column("UID", style="cyan", no_wrap=True, max_width=12)
                table.add_column("Username", style="green", max_width=20)
                table.add_column("Email", style="blue", max_width=30)
                table.add_column("Name", style="magenta", max_width=25)
                table.add_column("Active", style="yellow", justify="center")
                table.add_column("Verified", style="yellow", justify="center")
                table.add_column("Created", style="dim", max_width=20)

                for user in response.items:
                    full_name = []
                    if user.first_name:
                        full_name.append(user.first_name)
                    if user.last_name:
                        full_name.append(user.last_name)
                    display_name = " ".join(full_name) if full_name else "[dim]N/A[/dim]"

                    table.add_row(
                        user.user_uid[:10] + "..." if len(user.user_uid) > 10 else user.user_uid,
                        user.email,
                        user.email,
                        display_name,
                        "✅" if user.is_active else "❌",
                        "✅" if user.email_verified else "❌",
                        user.created_at.split("T")[0] if user.created_at else "N/A",
                    )

                console.print(table)

                # Display pagination info
                pagination_info = Panel(
                    f"Total: {response.pagination.total} users | "
                    f"Page {response.pagination.page} of {response.pagination.total_pages} | "
                    f"Showing {len(response.items)} items"
                )
                console.print(pagination_info)

                # Navigation options
                options = []
                if response.pagination.has_prev:
                    options.append("p")
                if response.pagination.has_next:
                    options.append("n")
                options.extend(["r", "q"])

                nav_text = []
                if response.pagination.has_prev:
                    nav_text.append("[bold cyan]p[/bold cyan]revious")
                if response.pagination.has_next:
                    nav_text.append("[bold cyan]n[/bold cyan]ext")
                nav_text.extend(["[bold cyan]r[/bold cyan]efresh", "[bold cyan]q[/bold cyan]uit"])

                console.print(f"\nNavigation: {' | '.join(nav_text)}")

                # Get user choice
                choice = Prompt.ask("Choose an option", choices=options, default="q")

                if choice == "n" and response.pagination.has_next:
                    current_page += 1
                elif choice == "p" and response.pagination.has_prev:
                    current_page -= 1
                elif choice == "r":
                    # Refresh current page
                    pass
                elif choice == "q":
                    break

                console.clear()

    except Exception as e:
        console.print(f"❌ Error: {e}", style="red")
        raise click.Abort()


@cli.command()
@click.option("--url", help="SlideInsight server URL (or set SLIDEINSIGHT_API_URL)")
@click.option("--email", help="Username for authentication (or set SLIDEINSIGHT_USERNAME)")
@click.option("--password", help="Password for authentication (or set SLIDEINSIGHT_PASSWORD)")
@click.option("--page-size", default=10, help="Number of items per page")
@click.option("--search", help="Search query")
@async_command
async def list_cases(
    url: Optional[str], email: Optional[str], password: Optional[str], page_size: int, search: Optional[str]
):
    """List cases with interactive pagination."""
    url, email, password = get_credentials(url, email, password)
    debug = ctx.obj.get("debug", False)

    try:
        async with SlideInsightClient(url, debug=debug) as client:
            await client.login(email, password)

            current_page = 1

            while True:
                # Fetch current page
                response = await client.cases.list(page=current_page, limit=page_size, search=search)

                if not response.items:
                    console.print("No cases found")
                    return

                # Display cases table
                table = Table(title=f"Cases (Page {response.pagination.page} of {response.pagination.total_pages})")
                table.add_column("UID", style="cyan", no_wrap=True, max_width=12)
                table.add_column("Name", style="green", max_width=30)
                table.add_column("Has Vector", style="yellow", justify="center")
                table.add_column("Has Raster", style="yellow", justify="center")
                table.add_column("Deleted", style="red", justify="center")
                table.add_column("Created", style="magenta", max_width=20)

                for case in response.items:
                    table.add_row(
                        case.case_uid[:10] + "..." if len(case.case_uid) > 10 else case.case_uid,
                        case.name,
                        "✅"
                        if case.has_vector_annotations
                        else ("❌" if case.has_vector_annotations is False else "❓"),
                        "✅"
                        if case.has_raster_annotations
                        else ("❌" if case.has_raster_annotations is False else "❓"),
                        "🗑️" if case.deleted_at else "❌",
                        case.created_at.split("T")[0] if case.created_at else "N/A",
                    )

                console.print(table)

                # Display pagination info
                pagination_info = Panel(
                    f"Total: {response.pagination.total} cases | "
                    f"Page {response.pagination.page} of {response.pagination.total_pages} | "
                    f"Showing {len(response.items)} items"
                )
                console.print(pagination_info)

                # Navigation options
                options = []
                if response.pagination.has_prev:
                    options.append("p")
                if response.pagination.has_next:
                    options.append("n")
                options.extend(["r", "q"])

                nav_text = []
                if response.pagination.has_prev:
                    nav_text.append("[bold cyan]p[/bold cyan]revious")
                if response.pagination.has_next:
                    nav_text.append("[bold cyan]n[/bold cyan]ext")
                nav_text.extend(["[bold cyan]r[/bold cyan]efresh", "[bold cyan]q[/bold cyan]uit"])

                console.print(f"\nNavigation: {' | '.join(nav_text)}")

                # Get user choice
                choice = Prompt.ask("Choose an option", choices=options, default="q")

                if choice == "n" and response.pagination.has_next:
                    current_page += 1
                elif choice == "p" and response.pagination.has_prev:
                    current_page -= 1
                elif choice == "r":
                    # Refresh current page
                    pass
                elif choice == "q":
                    break

                console.clear()

    except Exception as e:
        console.print(f"❌ Error: {e}", style="red")
        raise click.Abort()


def read_csv_file(csv_path: Path) -> list[tuple[str, str, Optional[str]]]:
    """Read CSV file and return list of (slide_uri, case_name, annotation_uri) tuples."""
    rows = []

    with open(csv_path, "r", newline="", encoding="utf-8") as csvfile:
        reader = csv.reader(csvfile)

        # Check if first row is header
        first_row = next(reader, None)
        if first_row is None:
            raise ValueError("CSV file is empty")

        # Skip header row if detected
        is_header = first_row[0].lower() in ["name", "casename", "case_name", "slide_name"] and first_row[
            1
        ].lower() in ["filename", "file", "path", "slide_path", "path_to_slide"]

        if not is_header:
            # First row is data, process it
            annotation_uri = first_row[2].strip() if len(first_row) > 2 and first_row[2].strip() else None
            rows.append((first_row[1].strip(), first_row[0].strip(), annotation_uri))

        # Process remaining rows
        for row_num, row in enumerate(reader, start=2):
            if len(row) < 2:
                console.print(f"⚠️  Row {row_num} has insufficient columns, skipping", style="yellow")
                continue

            case_name = row[0].strip()
            slide_uri = row[1].strip()
            annotation_uri = row[2].strip() if len(row) > 2 and row[2].strip() else None

            if not slide_uri or not case_name:
                console.print(f"⚠️  Row {row_num} has empty filename or name, skipping", style="yellow")
                continue

            rows.append((slide_uri, case_name, annotation_uri))

    return rows


async def process_row(
    client: SlideInsightClient,
    study_uid: str,
    slide_uri: str,
    case_name: str,
    annotation_uri: Optional[str] = None,
) -> None:
    """Process a single row from the CSV file.

    Creates a case with a slide and optionally adds annotations based on file type:
    - .tiff, .png files are added as raster annotations (masks)
    - .json, .geojson files are added as vector annotations
    """
    # Create case with slide
    case_uid, slide_uids = await client.create_case_with_slides(
        case_name=case_name,
        slides_data=[
            {
                "slide_uri": slide_uri,
                "slide_name": case_name,
            }
        ],
        study_uid=study_uid,
    )

    # Add annotation if provided
    if annotation_uri and slide_uids:
        slide_uid = slide_uids[0]

        if _is_vector_annotation(annotation_uri):
            # Add vector annotation (.json/.geojson)
            await client.slides.add_vector_annotation(
                slide_uid=slide_uid,
                file_uri=annotation_uri,
                name=f"{case_name} Vector Annotation",
                format="geojson",
            )
        else:
            # Add raster annotation (.tiff/.png)
            await client.slides.add_raster_annotation(
                slide_uid=slide_uid,
                mask_uri=annotation_uri,
                mask_name=f"{case_name} Mask",
            )


def main():
    """Main entry point for the CLI."""
    try:
        cli()
    except KeyboardInterrupt:
        console.print("\n⚠️  Operation cancelled by user", style="yellow")
    except Exception as e:
        console.print(f"❌ Unexpected error: {e}", style="red")


if __name__ == "__main__":
    main()
