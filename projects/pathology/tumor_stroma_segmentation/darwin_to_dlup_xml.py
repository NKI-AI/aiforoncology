import argparse
from pathlib import Path

from dlup import SlideAnnotations
from tqdm import tqdm
from lib.fix_rois import check_rois


ROI_NAME = "ROI (segmentation)"
"""
Next, we have the INDEX_MAP, which is a dictionary that maps the annotation names to the index that will be used in the DLUP XML file.

1 is stroma compartment, 
2 is tumor compartment,
3 are the compartments to ignore. 

The rest of the annotations are mapped to 0.
"""
INDEX_MAP = {
    "stroma (area)": 1,
    "tumor (area)": 2,
    "inflamed (area)": 1,
    "dcis (area)": 2,
    "lymphoid aggregate (area)": 1,
    "dcis immune cell (area)": 1,
    "necrotic (area)": 3,
    "normal gland (area)": 1,
    "blood vessel (area)": 1,
    "fat cell (area)": 3,
    "red blood cell (area)": 1,
    "background (area)": 3,
    "fibrosis (area)": 1,
    "artefacts": 3,
    "skin tissue": 3,
    "liver tissue": 3,
    "unlabeled (area)": 3,
}


def main():
    parser = argparse.ArgumentParser(description="Convert Darwin v7 annotations to GeoJSON.")
    parser.add_argument("annotation_folder", type=Path, help="Root path to the Darwin v7 json files.")
    parser.add_argument(
        "output_folder",
        type=Path,
        help="Path to the output folder. Will be created (including parents) if it does not exist.",
    )
    args = parser.parse_args()
    annotations_folder = args.annotation_folder
    output_folder = args.output_folder

    for annotation_file in tqdm(list(annotations_folder.glob("*.json"))):
        output_annotation_path = Path(output_folder / annotation_file.stem)
        wsi_annotations = SlideAnnotations.from_darwin_json(
            annotation_file, scaling=1, sorting="AREA", roi_names=[ROI_NAME]
        )

        wsi_annotations.filter_polygons(ROI_NAME)

        point_annotations = wsi_annotations.points
        point_annotation_names = [point.label for point in point_annotations]

        for point_name in point_annotation_names:
            wsi_annotations.filter_points(point_name)

        box_annotations = wsi_annotations.boxes
        box_annotation_names = [box.label for box in box_annotations]

        for box_name in box_annotation_names:
            wsi_annotations.filter_boxes(box_name)

        wsi_annotations.reindex_polygons(INDEX_MAP)

        for polygon in wsi_annotations.polygons:
            assert polygon.label in INDEX_MAP, f"Unexpected label: {polygon.label}"
            assert INDEX_MAP[polygon.label] == polygon.index, f"Index mismatch: {polygon.label} -> {polygon.index}"

        if len(wsi_annotations.rois) == 0:
            print(f"No ROIs found in {annotation_file.stem}.json")
            continue
        else:
            assert wsi_annotations.rois is not None
            assert [roi.label == ROI_NAME for roi in wsi_annotations.rois]
            wsi_annotations.rebuild_rtree()
            wsi_annotations = check_rois(wsi_annotations)
            output_annotation_path.mkdir(parents=True, exist_ok=True)

            with open(output_annotation_path / "annotations.xml", "w") as f:
                f.write(wsi_annotations.as_dlup_xml())


if __name__ == "__main__":
    main()
