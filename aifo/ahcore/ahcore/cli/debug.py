"""Debug utilities for ahcore CLI."""

import argparse
import os
import torch
import zipfile
from pathlib import Path
from ahcore.models.dummy_model import DummyModel
from ahcore.cli import file_path


def create_dummy_jit(args: argparse.Namespace) -> None:
    """Creates a very basic JIT for a segmentation which always returns a segmentation map with center = 1 and border = 2.
    The output is a .pack file that contains both model and configuration.

    Args:
        args: Command line arguments
    """
    model = DummyModel(border_width=args.border_width)
    model_configuration = """
<AifoModelConfiguration>
  <ModelName>Dummy segmentation model</ModelName>
  <Version>1.0</Version>
  <Task>
    <Type>Segmentation</Type>
    <MergeMethod>crop</MergeMethod>
  </Task>
  <Mpp>1.0</Mpp>
  <TileSize>
    <Width>256</Width>
    <Height>256</Height>
  </TileSize>
  <TileOverlap>
    <Width>0</Width>
    <Height>0</Height>
  </TileOverlap>
  <Labels>
    <Label>
      <n>Center of patch</n>
      <HexColor>#0000ff</HexColor>
      <Index>1</Index>
    </Label>
    <Label>
      <n>Border of patch</n>
      <HexColor>#ff0000</HexColor>
      <Index>2</Index>
    </Label>
  </Labels>
  <Normalization>
    <Mean>
      <Channel0>0.0</Channel0>
      <Channel1>0.0</Channel1>
      <Channel2>0.0</Channel2>
    </Mean>
    <Std>
      <Channel0>1.0</Channel0>
      <Channel1>1.0</Channel1>
      <Channel2>1.0</Channel2>
    </Std>
  </Normalization>
</AifoModelConfiguration>
"""

    # Create the output directory if it doesn't exist
    output_path = Path(args.output_path)
    output_path.parent.mkdir(parents=True, exist_ok=True)

    # Save model temporarily
    temp_model_path = "model.pt"
    torch.jit.script(model).save(temp_model_path)

    # Create a .pack file which zips both the model and the configuration file
    with zipfile.ZipFile(output_path, "w") as zipf:
        # Add the configuration as model_config.xml
        zipf.writestr("model_config.xml", model_configuration)
        zipf.write(temp_model_path)

    # Remove temporary model file
    os.remove(temp_model_path)

    print(f"Dummy JIT model created and saved to {output_path}")


def register_parser(
    parser: argparse._SubParsersAction,  # pylint: disable=unsubscriptable-object
) -> None:  # pylint: disable=E1136
    """Register debug commands to a root parser."""
    debug_parser = parser.add_parser("debug", help="Debug utilities")
    debug_subparsers = debug_parser.add_subparsers(help="Debug subparser")
    debug_subparsers.required = True
    debug_subparsers.dest = "subcommand"

    # Add create-dummy-jit command
    _parser: argparse.ArgumentParser = debug_subparsers.add_parser(
        "create-dummy-jit",
        help="Create a dummy JIT model that always returns a segmentation map with center = 1 and border = 2",
    )

    _parser.add_argument(
        "--border-width",
        type=int,
        default=4,
        help="Width of the border in the segmentation output",
    )

    _parser.add_argument(
        "--output-path",
        type=file_path,
        default=Path(__file__).parent / "dummy_model.pack",
        help="Path where the model pack file will be saved",
    )

    _parser.set_defaults(subcommand=create_dummy_jit)
