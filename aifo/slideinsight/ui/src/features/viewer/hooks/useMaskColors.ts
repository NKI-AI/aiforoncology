// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { useState } from "react";
import { MaskColor, DEFAULT_MASK_COLORS } from "@/types";

/**
 * Hook for managing mask colors state and operations
 */
export function useMaskColors() {
  const [maskColors, setMaskColors] =
    useState<MaskColor[]>(DEFAULT_MASK_COLORS);

  const handleMaskColorChange = (labelIndex: number, color: MaskColor) => {
    // Defensive copy to ensure React detects the state change
    const newColors = [...maskColors];
    if (labelIndex > 0 && labelIndex <= newColors.length) {
      newColors[labelIndex - 1] = { ...color }; // Create a new object to ensure state update
      setMaskColors(newColors);
    } else {
      console.warn(`Invalid label index: ${labelIndex}`);
    }
  };

  return {
    maskColors,
    setMaskColors,
    handleMaskColorChange,
  };
}
