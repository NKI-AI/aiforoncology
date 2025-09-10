import torch
import argparse
import tempfile
from pathlib import Path

import imageio.v3 as iio
import numpy as np
import PIL.Image

from dlup import SlideDataset
from dlup.tiling import Grid
from dlup.writers import TiffCompression, TifffileImageWriter
from ahcore.readers import StitchingMode, ZarrFileImageReader
from ahcore.writers import DataFormat, InferencePrecision, ZarrFileImageWriter
from tqdm import tqdm


def overlay_segmentation(image, segmentation_map):
    image_region_image = PIL.Image.fromarray(image).convert("RGBA")
    color_lut = np.array([[0, 0, 0], [0, 0, 255]], dtype=np.uint8)
    mask_array = color_lut[segmentation_map[..., 0]]

    image_mask = PIL.Image.fromarray(mask_array).convert("RGBA").resize(image_region_image.size)

    mask_array = np.array(image_mask)
    non_black_mask = np.any(mask_array[:, :, :3] != 0, axis=2)
    new_alpha = np.where(non_black_mask, 90, 0).astype(np.uint8)  # 90 is alpha value
    transparent_mask = mask_array.copy()
    transparent_mask[:, :, 3] = new_alpha
    transparent_mask_image = PIL.Image.fromarray(transparent_mask, "RGBA")
    combined_image = PIL.Image.alpha_composite(image_region_image, transparent_mask_image)
    return combined_image


class ProcessImage:
    def __init__(
        self,
        model_path: Path,
        image_file: Path,
        output_file: Path,
        tile_size=(1024, 1024),
        tile_overlap=(256, 256),
        mpp=12.0,
        create_thumbnail: bool = True,
    ):
        self._model_path = model_path
        self._image_file = image_file
        self._output_file = output_file
        self._tile_size = tile_size
        self._tile_overlap = tile_overlap
        self._mpp = mpp

        self._device = torch.device(
            "cuda" if torch.cuda.is_available() else "mps" if torch.backends.mps.is_available() else "cpu"
        )
        self._jitted_model = torch.jit.load(str(self._model_path)).to(self._device)
        self._jitted_model.eval()
        self._dataset = self.construct_dataset()
        self._create_thumbnail = create_thumbnail

    def construct_dataset(self):
        dataset = SlideDataset.from_standard_tiling(
            self._image_file,
            mpp=self._mpp,
            tile_size=self._tile_size,
            tile_overlap=self._tile_overlap,
            random_sample_in_grid=False,
            tile_mode="overflow",
            grid_order="C",
            crop=False,
            backend="PYVIPS",
            internal_handler="vips",
            limit_bounds=True,
        )
        return dataset

    def infer_and_save(self):
        with tempfile.TemporaryDirectory() as tmpdir:
            temp_file = Path(tmpdir) / "tempfile.zarr"
            self.run_inference(temp_file)
            self.process_results(temp_file)

    def run_inference(self, temp_file: Path):
        slide_image = self._dataset.slide_image
        scaling = slide_image.get_scaling(self._mpp)
        scaled_region_view = slide_image.get_scaled_view(scaling)

        writer = ZarrFileImageWriter(
            temp_file,
            size=scaled_region_view.size,
            mpp=self._mpp,
            tile_size=self._tile_size,
            tile_overlap=self._tile_overlap,
            num_samples=len(self._dataset),
            data_format=DataFormat.IMAGE,
        )

        @torch.no_grad()
        def generator():
            for sample in tqdm(self._dataset):
                image = sample.image.flatten(background=[255, 255, 255]).numpy()
                data = torch.from_numpy(image).permute(2, 0, 1).unsqueeze(0).float().to(self._device)
                coordinates = np.asarray(sample.coordinates)[np.newaxis, ...]
                output = self._jitted_model(data).cpu()
                segmentation_map = output[0].argmax(dim=0).numpy().astype("uint8")[np.newaxis, np.newaxis, ...]

                yield coordinates, segmentation_map

        writer.consume(generator())

    def process_results(self, temp_file: Path):
        temp_file = Path(temp_file)
        reader = ZarrFileImageReader(temp_file, stitching_mode=StitchingMode.CROP)

        grid = Grid.from_tiling(
            (0, 0), reader.size, tile_size=self._tile_size, tile_overlap=(0, 0), mode="overflow", order="C"
        )

        def generate_regions():
            for coordinates in grid:
                region = reader.read_region(coordinates, 0, self._tile_size)
                data = region.numpy().astype(np.uint8)
                yield data

        writer = TifffileImageWriter(
            self._output_file,
            size=reader.size,
            mpp=reader.mpp,
            tile_size=self._tile_size,
            pyramid=True,
            compression=TiffCompression.ZSTD,
        )
        writer.from_tiles_iterator(generate_regions())

        if self._create_thumbnail:
            self.create_overlay()

    def create_overlay(self):
        slide_image = self._dataset.slide_image
        full_size = slide_image.size
        scaling = 4096 / max(full_size)
        scaled_size = (int(full_size[0] * scaling), int(full_size[1] * scaling))
        mask = iio.imread(self._output_file)

        image = slide_image.read_region((0, 0), scaling, scaled_size).numpy()
        overlay_segmentation(image, mask).save(self._output_file.with_suffix(".thumbnail.png"))


def run_inference_command(args):
    """Run inference on an image using a JIT model.

    Args:
        args: Command line arguments
    """
    processor = ProcessImage(
        model_path=args.model_path, image_file=args.image, output_file=args.output, create_thumbnail=args.save_thumbnail
    )
    processor.infer_and_save()


def register_parser(
    parser: argparse._SubParsersAction,  # pylint: disable=unsubscriptable-object
) -> None:  # pylint: disable=E1136
    """Register inference commands to a root parser."""
    inference_parser = parser.add_parser("inference", help="Inference utilities")
    inference_subparsers = inference_parser.add_subparsers(help="Inference subparser")
    inference_subparsers.required = True
    inference_subparsers.dest = "subcommand"

    # Add segmentation command
    segmentation_parser = inference_subparsers.add_parser(
        "segmentation",
        help="Run segmentation on image and produce overlay results",
    )

    segmentation_parser.add_argument(
        "--image",
        type=Path,
        required=True,
        help="Path to the image file",
    )

    segmentation_parser.add_argument(
        "--output",
        type=Path,
        required=True,
        help="Path to the output TIFF file",
    )

    segmentation_parser.add_argument(
        "--save-thumbnail",
        action="store_true",
        help="Save a thumbnail of the overlay",
    )

    segmentation_parser.add_argument(
        "--model-path",
        type=Path,
        required=True,
        help="Path to the model file",
    )

    segmentation_parser.set_defaults(subcommand=run_inference_command)
