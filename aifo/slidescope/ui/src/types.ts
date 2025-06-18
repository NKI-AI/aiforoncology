// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.

/**
 * Interface for mask color representation with RGBA values
 */
export interface MaskColor {
  r: number;
  g: number;
  b: number;
  a: number;
}

/**
 * Default color palette for mask labels, optimized for visibility in histology images
 */
export const DEFAULT_MASK_COLORS: MaskColor[] = [
  { r: 220, g: 40, b: 40, a: 1 }, // Red for tumor/primary features
  { r: 40, g: 180, b: 40, a: 1 }, // Green for stroma/secondary features
  { r: 40, g: 120, b: 220, a: 1 }, // Blue for nuclei/tertiary features
  { r: 220, g: 220, b: 40, a: 1 }, // Yellow for additional features
];
