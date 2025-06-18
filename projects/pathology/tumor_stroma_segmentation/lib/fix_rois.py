from dlup.geometry import Polygon as DlupPolygon
from shapely.ops import unary_union


def rois_match(roi1, roi2, tolerance=1e-7):
    """Check if two geometries are essentially the same, within a small area tolerance."""
    return roi1.symmetric_difference(roi2).area < tolerance


def check_rois(annotations, tolerance=1e-7):
    """
    For each ROI in 'annotations.rois':
      1) Collect polygons that intersect it.
      2) If none, remove the ROI.
      3) Otherwise, compute the union and then minimum rotated rectangle (MRR) of those polygons.
      4) If MRR differs from the old ROI, replace the old ROI with the MRR.

    Parameters
    ----------
    annotations : SlideAnnotations
        The annotations for which ROIs should be checked.
    tolerance: float
        Numeric threshold for deciding if two geometries are the "same."

    Return
    ------
        The updated annotations object.
    """
    # Convert all annotation polygons to Shapely once
    poly_geoms = [ann.to_shapely() for ann in annotations.polygons]

    # Iterate over a copy of current ROIs to avoid modifying while iterating
    original_rois = annotations.rois[:]
    for old_roi in original_rois:
        roi_geom = old_roi.to_shapely()

        # 1) Find polygons that intersect this ROI
        relevant_polys = [pg.buffer(0) for pg in poly_geoms if pg.intersects(roi_geom)]
        if not relevant_polys:
            # 2) If none intersect, remove this ROI
            annotations.remove_roi(old_roi)
            annotations.rebuild_rtree()
            continue

        # 3) Compute union and minimum rotated rectangle
        combined = unary_union(relevant_polys)
        mrr_geom = combined.minimum_rotated_rectangle

        # 4) Replace ROI if the MRR differs from the old geometry
        if not rois_match(roi_geom, mrr_geom, tolerance):
            annotations.remove_roi(old_roi)
            annotations.rebuild_rtree()

            new_roi = DlupPolygon.from_shapely(mrr_geom)
            new_roi.label = getattr(old_roi, "label", None)
            annotations.add_roi(new_roi)
            annotations.rebuild_rtree()

    return annotations
