// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, {
  createContext,
  useState,
  useContext,
  ReactNode,
  useEffect,
} from "react";
import { MaskColor } from "@/types";
import { useMaskKeyboardShortcuts } from "@/features/viewer/hooks/useMaskKeyboardShortcuts";
import { useMaskColors } from "@/features/viewer/hooks/useMaskColors";
import { type MaskLayer } from "@/features/viewer/components/map/webglStyleUtils";

interface MaskContextType {
  maskOpacity: number;
  showMask: boolean;
  maskColors: MaskColor[];
  maskLayers: MaskLayer[]; // Individual mask layers with their info
  setMaskOpacity: (opacity: number) => void;
  setShowMask: (show: boolean) => void;
  setMaskLayers: (layers: MaskLayer[]) => void;
  toggleMaskLayerVisibility: (id: string) => void;
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
  const [maskLayers, setMaskLayers] = useState<MaskLayer[]>([]);

  // Use the extracted mask colors hook
  const { maskColors, setMaskColors, handleMaskColorChange } = useMaskColors();

  // Use the extracted keyboard shortcuts hook
  useMaskKeyboardShortcuts({
    onToggleMask: () => setShowMask((prevShow) => !prevShow),
    onIncreaseMaskOpacity: () =>
      setMaskOpacity((prev) => Math.min(1, prev + 0.1)),
    onDecreaseMaskOpacity: () =>
      setMaskOpacity((prev) => Math.max(0, prev - 0.1)),
  });

  const toggleMaskLayerVisibility = (id: string) => {
    setMaskLayers((prev) =>
      prev.map((layer) =>
        layer.id === id ? { ...layer, visible: !layer.visible } : layer
      )
    );
  };

  const value = {
    maskOpacity,
    showMask,
    maskColors,
    maskLayers,
    setMaskOpacity,
    setShowMask,
    setMaskLayers,
    toggleMaskLayerVisibility,
    handleMaskColorChange,
  };

  return <MaskContext.Provider value={value}>{children}</MaskContext.Provider>;
}
