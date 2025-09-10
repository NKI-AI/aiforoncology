import argparse
import os
import sys


def main() -> None:
    """
    Main entrypoint for the CLI command of ahcore.
    """
    # Perhaps we should detect if we are running under Bazel
    # and use the correct program name.
    prog_name = "bazelisk run //aifo/llm_serveahcore:cli --"

    # From https://stackoverflow.com/questions/17073688/how-to-use-argparse-subparsers-correctly
    root_parser = argparse.ArgumentParser(formatter_class=argparse.ArgumentDefaultsHelpFormatter, prog=prog_name)

    root_subparsers = root_parser.add_subparsers(help="Possible ahcore CLI utils to run.")
    root_subparsers.required = True
    root_subparsers.dest = "subcommand"

    # Prevent circular import
    from ahcore.cli.data import register_parser as register_data_subcommand

    # Data related commands.
    register_data_subcommand(root_subparsers)

    # Prevent circular import
    from ahcore.cli.tiling import register_parser as register_tiling_subcommand

    # Tiling related commands
    register_tiling_subcommand(root_subparsers)

    # Prevent circular import
    from ahcore.cli.debug import register_parser as register_debug_subcommand

    # Debug related commands
    register_debug_subcommand(root_subparsers)

    # Prevent circular import
    from ahcore.cli.download import register_parser as register_download_subcommand

    # Download related commands
    register_download_subcommand(root_subparsers)

    # Prevent circular import
    from ahcore.cli.inference import register_parser as register_inference_subcommand

    # Inference related commands
    register_inference_subcommand(root_subparsers)

    args = root_parser.parse_args()
    args.subcommand(args)


if __name__ == "__main__":
    main()
