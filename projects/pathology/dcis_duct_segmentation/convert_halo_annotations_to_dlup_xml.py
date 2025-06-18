# Copyright (c) dlup contributors
"""This script converts the Halo XML annotations to XML and GeoJSON annotations (including for QuPath).
It also generates thumbnails over the overlays.
"""

import json
from pathlib import Path
import argparse

import numpy as np
import PIL.Image

from dlup import SlideImage, SlideAnnotations


FILTER_LABELS = [
    "Jelle_check",
    "Not_DCIS_Joyce",
    "DCIS_90PM",
    "Jelle to check",
    "Check with Jelle",
    "Glass",
    "al gehad",
    "check with Jelle",
    "Stroma",
    "tissue (for stroma AI classifier)",
    "tissue (for AI classifier)",
    "Stroma",
    "tissue",
    "Tissue",
]
RELABEL_MAP = {"DCIS_SV": "DCIS", "DCIS_Joyce": "DCIS", "DCIS_PK": "DCIS"}


def parse_arguments():
    parser = argparse.ArgumentParser(description="Process HALO annotations for DCIS duct segmentation")
    parser.add_argument("--images-root", type=Path, required=True, help="Root path for the image files")
    parser.add_argument("--annotations-root", type=Path, required=True, help="Path to the annotation files")
    parser.add_argument("--output-path", type=Path, required=True, help="Output path for processed files")
    parser.add_argument(
        "--data-list",
        type=Path,
        required=True,
        help="Path to the DCIS list file. It should have two columns: relative path to the image and the annotation file. First line is the header",
    )
    return parser.parse_args()


def main():
    args = parse_arguments()

    # Create output directory
    args.output_path.mkdir(exist_ok=True)

    with open(args.data_list, "r") as f:
        dcis_list = f.readlines()[1:]
        dcis_list = [dcis.strip() for dcis in dcis_list if dcis.strip()]
        data_pairs = [_.split(",") for _ in dcis_list]
        data_pairs = [(args.images_root / _[0], args.annotations_root / _[1]) for _ in data_pairs]

    filtered_pairs = []
    # Let's filter those that don't exists
    for image_fn, annotation_fn in data_pairs:
        if not image_fn.exists():
            print(f"Image file {image_fn} does not exist")
            continue
        if not annotation_fn.exists():
            print(f"Annotation file {annotation_fn} does not exist")
            continue
        filtered_pairs.append((image_fn, annotation_fn))

    print(f"Working on {len(filtered_pairs)} filtered images")

    for image_fn, annotation_fn in filtered_pairs:
        if annotation_fn.name not in ["T14-68021_A3.annotations"]:
            continue

        output_path_images = args.output_path / "thumbnails" / (annotation_fn.stem + ".png")
        if output_path_images.exists():
            print(f"Skipping {annotation_fn} as it is done")
            continue
        print(f"Working on {annotation_fn}")
        with SlideImage.from_file_path(image_fn, backend="PYVIPS") as slide:
            # Let's find a scaling factor that makes sure the longest side is 2048 pixels
            _, full_bounded_size = slide.slide_bounds

            scaling = 4096 / max(full_bounded_size)
            start_coords, scaled_size = slide.get_scaled_slide_bounds(scaling)

            image_region_vips = slide.read_region(start_coords, scaling, scaled_size)
            offset, _ = slide.slide_bounds

        annotations = SlideAnnotations.from_halo_xml(annotation_fn)
        actual_offset = offset - np.asarray(offset) % 256
        assert np.all(actual_offset % 256 == 0)
        annotations.set_offset(tuple(actual_offset))

        annotations_qupath = SlideAnnotations.from_halo_xml(annotation_fn)
        qupath_offset = -(np.asarray(offset) % 256)
        annotations_qupath.set_offset((qupath_offset[0], qupath_offset[1]))

        for label in FILTER_LABELS:
            annotations.filter_polygons(label)
            annotations_qupath.filter_polygons(label)
            annotations.filter_boxes(label)
            annotations_qupath.filter_boxes(label)

        annotations.rebuild_rtree()
        annotations_qupath.rebuild_rtree()

        annotations.relabel_polygons(RELABEL_MAP)
        annotations.reindex_polygons({"DCIS": 1})

        annotations_qupath.relabel_polygons(RELABEL_MAP)
        annotations_qupath.reindex_polygons({"DCIS": 1})

        # Let's save it to GeoJSON and XML
        (args.output_path / "geojsons").mkdir(exist_ok=True)
        (args.output_path / "qupath_geojsons").mkdir(exist_ok=True)
        (args.output_path / "xmls").mkdir(exist_ok=True)
        output_path_json = args.output_path / "geojsons" / (annotation_fn.stem + ".json")
        output_path_json_qupath = args.output_path / "qupath_geojsons" / (annotation_fn.stem + ".qupath.json")
        output_path_xml = args.output_path / "xmls" / (annotation_fn.stem + ".xml")

        with open(output_path_json_qupath, "w") as f:
            json.dump(annotations_qupath.as_geojson(), f, indent=2)

        with open(output_path_json, "w") as f:
            json.dump(annotations.as_geojson(), f, indent=2)

        metadata = {"image_id": image_fn.name, "description": "DCIS duct annotations", "version": "v1.0"}
        annotations.metadata = metadata

        with open(output_path_xml, "w") as f:
            f.write(annotations.as_dlup_xml())
        image_region = PIL.Image.fromarray(image_region_vips.numpy()).convert("RGBA")

        assert annotations.available_classes == {"DCIS"}

        region = annotations.read_region(start_coords, scaling, scaled_size)
        image_mask = PIL.Image.fromarray(annotations.color_lut[region.polygons.to_mask().numpy()]).convert("RGBA")

        mask_array = np.array(image_mask)
        non_black_mask = np.any(mask_array[:, :, :3] != 0, axis=2)
        new_alpha = np.where(non_black_mask, 90, 0).astype(np.uint8)  # 90 is alpha value
        transparent_mask = mask_array.copy()
        transparent_mask[:, :, 3] = new_alpha
        transparent_mask_image = PIL.Image.fromarray(transparent_mask, "RGBA")
        image_region_image = PIL.Image.fromarray(np.array(image_region), "RGBA")
        combined_image = PIL.Image.alpha_composite(image_region_image, transparent_mask_image)

        (args.output_path / "thumbnails").mkdir(exist_ok=True)
        combined_image.save(output_path_images)


if __name__ == "__main__":
    main()
