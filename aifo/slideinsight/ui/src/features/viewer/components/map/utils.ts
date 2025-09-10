// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { SlideMetadata, DisplayMetadata } from "./types";
import { MaskColor } from "@/types";

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
    slideUid: slideMeta.slideUid || "Unknown",
    slideMpp: slideMeta.slideMpp ? `${slideMeta.slideMpp} µm/px` : "Unknown",
  };
}

/**
 * Generate WebGL color expression from legacy mask colors (fallback)
 */
export function generateColorExpression(colors: MaskColor[]): any[] {
  // Start building our case expression
  const colorExpression: any[] = ["case"];

  // First case is always transparent for value 0
  colorExpression.push(["==", ["*", ["band", 1], 255], 0], [0, 0, 0, 0.0]);

  // Add each color from the maskColors array
  colors.forEach((color, index) => {
    const labelValue = index + 1;
    // Use the alpha value from the color object to control visibility
    colorExpression.push(
      ["==", ["*", ["band", 1], 255], labelValue],
      [color.r, color.g, color.b, color.a]
    );
  });
  colorExpression.push([0, 0, 0, 0.0]);

  return colorExpression;
}
