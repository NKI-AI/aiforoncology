// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
import ImageTile from "ol/ImageTile";
import { FetchError, SlideMetadata, DisplayMetadata } from "./types";

import { apiFetch } from "../../utils/fetchUtils";

/**
 * Handle errors in tile loading
 */
export function handleTileLoadError(
  tile: ImageTile,
  z: number,
  x: number,
  y: number,
  error: any
): void {
  console.error(`[ERROR] Error loading tile z=${z} x=${x} y=${y}:`, error);
  // Log the tile URL if possible
  const img = tile.getImage();
  if (img instanceof HTMLImageElement && img.src) {
    console.error(`[ERROR] Failed image URL: ${img.src}`);
  }
  // Future enhancement: Could set a default error image or implement retry logic
}

/**
 * Process slide metadata for display
 */
export function processMetadataForDisplay(
  slideMeta: SlideMetadata
): DisplayMetadata {
  // Get slide dimensions
  const width = slideMeta.slideWidth.toString();
  const height = slideMeta.slideHeight.toString();
  const slideMpp = slideMeta.slideMpp.toString();
  const objective = slideMeta.magnification || "Unknown";
  const vendor = slideMeta.vendor || "Unknown";

  // Calculate size in megapixels
  const widthNum = slideMeta.slideWidth;
  const heightNum = slideMeta.slideHeight;
  const megapixels = ((widthNum * heightNum) / 1000000).toFixed(1);

  // Calculate physical dimensions if mpp is available
  let physicalDimensions = "Unknown";

  if (slideMpp) {
    const widthMm = ((widthNum * slideMeta.slideMpp) / 1000).toFixed(1);
    const heightMm = ((heightNum * slideMeta.slideMpp) / 1000).toFixed(1);
    physicalDimensions = `${widthMm} mm × ${heightMm} mm`;
  }

  return {
    resolution: `${width} × ${height}`,
    objective: `${objective}×`,
    vendor: vendor,
    digitalSize: `${megapixels} MP`,
    physicalDimensions,
    slideId: slideMeta.slideId || "Unknown",
    slideMpp: slideMeta.slideMpp ? `${slideMeta.slideMpp} µm/px` : "Unknown",
  };
}
