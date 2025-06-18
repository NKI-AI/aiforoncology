// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
import Map from "ol/Map";
import View from "ol/View";
import TileLayer from "ol/layer/WebGLTile";
import { OverviewMap } from "ol/control";
import { getCenter } from "ol/extent";
import { WholeSlideImage } from "./SlideImage";

/**
 * Creates an OpenLayers OverviewMap control for a WholeSlideImage
 *
 * @param map - The OpenLayers map to add the control to
 * @param source - The WholeSlideImage source
 * @param collapsed - Whether the overview map should start collapsed
 * @returns The created OverviewMap control
 */
export function createOverviewMapControl(
  map: Map,
  source: WholeSlideImage,
  collapsed = true
): OverviewMap {
  // Get the tile grid from the source
  const grid = source.getTileGrid();
  const allRes = grid.getResolutions();
  const overviewRes0 = allRes[grid.getMinZoom()]; // The "zoom 0" resolution
  const extent = grid.getExtent();
  const projection = source.getProjection();

  // Handle case where projection might be null
  if (!projection) {
    throw new Error("Source projection is null - cannot create overview map");
  }

  // Create the overview map control
  const overviewMapControl = new OverviewMap({
    className: "ol-overviewmap ol-overviewmap-topright",
    layers: [
      new TileLayer({
        source: source as any,
      }),
    ],
    collapsed: collapsed, // Start collapsed if specified
    collapsible: true, // Can be collapsed/expanded
    tipLabel: "Toggle overview map",
    // Use simple < and > symbols
    collapseLabel: "<", // Symbol to show when expanded (to collapse)
    label: ">", // Symbol to show when collapsed (to expand)
    view: new View({
      projection: projection,
      resolutions: [overviewRes0 * 3], // Only zoom level 0 available
      resolution: overviewRes0 * 3, // Start at that resolution
      center: getCenter(extent), // Center over the whole slide
    }),
  });

  return overviewMapControl;
}

/**
 * Forces the overview map to be positioned correctly in the top-right
 */
export function forceOverviewPosition(): void {
  const overviewElement = document.querySelector(".ol-overviewmap-topright");
  if (overviewElement) {
    // Force inline styles
    (overviewElement as HTMLElement).style.right = "10px";
    (overviewElement as HTMLElement).style.left = "auto";
    (overviewElement as HTMLElement).style.top = "10px";
  }
}
