"""Download utilities for ahcore CLI."""

import argparse
from ahcore.cli.fomo_jits.download_fomo_jits import register_parser as register_fomo_jits_subcommand


def register_parser(
    parser: argparse._SubParsersAction,  # pylint: disable=unsubscriptable-object
) -> None:  # pylint: disable=E1136
    """Register download commands to a root parser."""
    download_parser = parser.add_parser("download", help="Download utilities")
    download_subparsers = download_parser.add_subparsers(help="Download subparser")
    download_subparsers.required = True
    download_subparsers.dest = "subcommand"

    # Add fomo-jits command
    register_fomo_jits_subcommand(download_subparsers)
