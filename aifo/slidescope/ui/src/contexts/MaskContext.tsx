// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
import React, {
  createContext,
  useState,
  useContext,
  ReactNode,
  useEffect,
} from "react";
import { MaskColor, DEFAULT_MASK_COLORS } from "../types";
import { useKeyboardShortcuts } from "../hooks/useKeyboardShortcuts";

interface MaskContextType {
  maskOpacity: number;
  showMask: boolean;
  maskColors: MaskColor[];
  setMaskOpacity: (opacity: number) => void;
  setShowMask: (show: boolean) => void;
  handleMaskColorChange: (labelIndex: number, color: MaskColor) => void;
}

const MaskContext = createContext<MaskContextType | null>(null);

export function useMaskContext() {
  const context = useContext(MaskContext);
  if (!context) {
    throw new Error("useMaskContext must be used within a MaskProvider");
  }
  return context;
}

interface MaskProviderProps {
  children: ReactNode;
}

export function MaskProvider({ children }: MaskProviderProps) {
  const [maskOpacity, setMaskOpacity] = useState(0.5);
  const [showMask, setShowMask] = useState(true);
  const [maskColors, setMaskColors] =
    useState<MaskColor[]>(DEFAULT_MASK_COLORS);

  // Handle keyboard shortcuts for mask controls
  useKeyboardShortcuts({
    onKeyT: () => setShowMask((prevShow) => !prevShow),
    onKeyPlus: () => setMaskOpacity((prev) => Math.min(1, prev + 0.1)),
    onKeyMinus: () => setMaskOpacity((prev) => Math.max(0, prev - 0.1)),
  });

  // Handler for mask color changes
  const handleMaskColorChange = (labelIndex: number, color: MaskColor) => {
    console.log(`[DEBUG] Changed color for label ${labelIndex} to:`, color);

    // Defensive copy to ensure React detects the state change
    const newColors = [...maskColors];
    if (labelIndex > 0 && labelIndex <= newColors.length) {
      newColors[labelIndex - 1] = { ...color }; // Create a new object to ensure state update
      console.log("[DEBUG] New colors array:", newColors);
      setMaskColors(newColors);
    } else {
      console.warn(`[DEBUG] Invalid label index: ${labelIndex}`);
    }
  };

  const value = {
    maskOpacity,
    showMask,
    maskColors,
    setMaskOpacity,
    setShowMask,
    handleMaskColorChange,
  };

  return <MaskContext.Provider value={value}>{children}</MaskContext.Provider>;
}
