// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

/**
 * WebGL style utilities for masks and annotations
 */

import { MaskColor } from "@/types";
import { generateColorExpression } from "./utils";

export interface MaskLayer {
  id: string;
  name: string;
  color: string;
  visible: boolean;
  index: number;
  maskName?: string;
}

/**
 * Generate WebGL color expression from API-based mask layers
 */
export function generateColorExpressionFromLayers(layers: MaskLayer[]): any[] {
  // Start building our case expression
  const colorExpression: any[] = ["case"];

  // First case is always transparent for value 0
  colorExpression.push(["==", ["*", ["band", 1], 255], 0], [0, 0, 0, 0.0]);

  // Add each layer based on its index and visibility
  layers.forEach((layer) => {
    // Convert hex color to RGB
    const hex = layer.color.replace("#", "");
    const r = parseInt(hex.substr(0, 2), 16);
    const g = parseInt(hex.substr(2, 2), 16);
    const b = parseInt(hex.substr(4, 2), 16);

    // Use alpha 1.0 if visible, 0.0 if hidden
    const alpha = layer.visible ? 1.0 : 0.0;

    colorExpression.push(
      ["==", ["*", ["band", 1], 255], layer.index],
      [r, g, b, alpha]
    );
  });

  // Default case - transparent
  colorExpression.push([0, 0, 0, 0.0]);

  return colorExpression;
}

/**
 * Create WebGL mask style from layers or fallback colors
 */
export function createMaskStyle(
  maskLayers: MaskLayer[],
  fallbackColors: MaskColor[]
): { color: any[]; gamma: number } {
  let colorExpression: any[];

  if (maskLayers.length > 0) {
    colorExpression = generateColorExpressionFromLayers(maskLayers);
  } else {
    colorExpression = generateColorExpression(fallbackColors);
  }

  return {
    color: colorExpression,
    gamma: 1.0,
  };
}
