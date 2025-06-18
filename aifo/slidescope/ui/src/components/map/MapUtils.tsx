// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
import Map from "ol/Map";
import View from "ol/View";
import TileLayer from "ol/layer/WebGLTile";
import { Extent, getCenter } from "ol/extent";
import { Projection } from "ol/proj";
import { defaults as defaultControls } from "ol/control";
import { WholeSlideImage } from "./SlideImage";

/**
 * Constants for map configuration
 */
const MAP_PRELOAD = 8;
const DEFAULT_MAX_TILES_LOADING = 16;

/**
 * Create a map with a slide source
 *
 * @param source The slide image source
 * @param target The HTML element to render the map into
 * @returns A new OpenLayers Map instance
 */
export function createMap(source: WholeSlideImage, target: HTMLElement): Map {
  // Create view
  const view = createMapView(source);

  // Create map
  const map = createMapInstance(target, source, view);

  return map;
}

/**
 * Create a view for the map
 */
function createMapView(source: WholeSlideImage): View {
  const extent = source.getTileGrid().getExtent();
  const projection = source.getProjection();
  const tileGrid = source.getTileGrid();

  // Handle case where projection might be null
  if (!projection) {
    throw new Error("Source projection is null - cannot create map view");
  }

  return new View({
    projection,
    center: getCenter(extent),
    resolutions: tileGrid.getResolutions(),
    zoom: tileGrid.getMinZoom(),
    extent,
  });
}

/**
 * Create a map instance
 */
function createMapInstance(
  target: HTMLElement,
  source: WholeSlideImage,
  view: View
): Map {
  const extent = source.getTileGrid().getExtent();

  return new Map({
    target,
    layers: [
      new TileLayer({
        preload: MAP_PRELOAD,
        extent,
        source: source as any, // Type assertion needed for WebGLTile
        cacheSize: 0, // Disable tile cache at layer level
        useInterimTilesOnError: true,
        properties: {
          uniqueId: `layer-${Date.now()}`,
        },
      }),
    ],
    view,
    controls: defaultControls(),
    maxTilesLoading: DEFAULT_MAX_TILES_LOADING,
  });
}

/**
 * Calculate the appropriate resolution to fit an image in a viewport
 *
 * @param imageWidth Image width in pixels
 * @param imageHeight Image height in pixels
 * @param viewportWidth Viewport width in pixels
 * @param viewportHeight Viewport height in pixels
 * @param pixelRatio Device pixel ratio
 * @returns The resolution needed to fit the image
 */
export function calculateFitResolution(
  imageWidth: number,
  imageHeight: number,
  viewportWidth: number,
  viewportHeight: number,
  pixelRatio: number = 1
): number {
  const imageAspect = imageWidth / imageHeight;
  const viewportAspect = viewportWidth / viewportHeight;

  if (imageAspect > viewportAspect) {
    // Image is wider than viewport (relative to height)
    return imageWidth / (viewportWidth * pixelRatio);
  } else {
    // Image is taller than viewport (relative to width)
    return imageHeight / (viewportHeight * pixelRatio);
  }
}
