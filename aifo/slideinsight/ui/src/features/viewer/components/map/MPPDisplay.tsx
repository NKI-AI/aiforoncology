// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React, { useEffect, useState } from "react";
import Map from "ol/Map";

interface MPPDisplayProps {
  map: Map;
  baseMpp?: number; // The MPP at the highest resolution (zoom level with actual tiles)
  showMppDisplay?: boolean;
}

/**
 * Component that displays the current MPP (microns per pixel) at the current zoom level
 * Positioned similar to the scale line control
 */
export function MPPDisplay({
  map,
  baseMpp,
  showMppDisplay = true,
}: MPPDisplayProps) {
  const [currentMpp, setCurrentMpp] = useState<number | null>(null);

  useEffect(() => {
    if (!map || !baseMpp || !showMppDisplay) {
      setCurrentMpp(null);
      return;
    }

    const updateMpp = () => {
      const view = map.getView();
      const resolution = view.getResolution();

      if (resolution === undefined) {
        setCurrentMpp(null);
        return;
      }

      // Get the tile grid resolutions to understand the zoom structure
      const layers = map.getLayers().getArray();
      const slideLayer = layers.find((layer) => {
        const source = (layer as any).getSource?.();
        return source && typeof source.getTileGrid === "function";
      });

      if (!slideLayer) {
        setCurrentMpp(null);
        return;
      }

      const source = (slideLayer as any).getSource();
      if (!source || typeof source.getTileGrid !== "function") {
        setCurrentMpp(null);
        return;
      }

      // The current resolution is in meters per pixel
      // We need to convert it to microns per pixel
      // 1 meter = 1,000,000 microns
      const currentMppFromResolution = resolution * 1000000;

      setCurrentMpp(currentMppFromResolution);
    };

    // Update MPP when view changes
    const view = map.getView();
    view.on("change:resolution", updateMpp);
    view.on("change:center", updateMpp);

    // Initial calculation
    updateMpp();

    return () => {
      view.un("change:resolution", updateMpp);
      view.un("change:center", updateMpp);
    };
  }, [map, baseMpp, showMppDisplay]);

  if (!showMppDisplay || currentMpp === null) {
    return null;
  }

  // Format MPP for display
  const formatMpp = (mpp: number): string => {
    if (mpp >= 1) {
      return mpp.toFixed(2);
    } else if (mpp >= 0.1) {
      return mpp.toFixed(3);
    } else {
      return mpp.toFixed(4);
    }
  };

  return (
    <div
      className="absolute bottom-2 left-2 z-10"
      style={{
        // Position it just above the scale line if both are visible
        marginBottom: "28px",
      }}
    >
      <div className="bg-white/90 backdrop-blur-sm text-black text-xs px-2 py-1 rounded-sm border border-gray-300 shadow-sm font-mono">
        <span className="text-gray-600">MPP:</span>{" "}
        <span className="font-semibold">{formatMpp(currentMpp)} µm/px</span>
      </div>
    </div>
  );
}
