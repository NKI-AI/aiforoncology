// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
import { useEffect } from "react";
import Map from "ol/Map";

interface CoordinateTrackerProps {
  map: Map | null;
  mapContainer: HTMLDivElement | null;
  onPositionChange: (position: { x: number; y: number }) => void;
  onCoordinatesChange: (coordinates: { x: string; y: string }) => void;
}

/**
 * Component that tracks mouse coordinates on the map
 */
export default function CoordinateTracker({
  map,
  mapContainer,
  onPositionChange,
  onCoordinatesChange,
}: CoordinateTrackerProps) {
  useEffect(() => {
    if (!map || !mapContainer) return;

    // Function to handle pointer move
    let pending = false;

    const handlePointerMove = (e: MouseEvent) => {
      if (pending) return;
      pending = true;

      requestAnimationFrame(() => {
        // Get mouse position relative to the map container
        const rect = mapContainer.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const y = e.clientY - rect.top;

        // Update cursor position for crosshair
        onPositionChange({ x, y });

        // Get OpenLayers coordinates
        const coordinate = map.getCoordinateFromPixel([x, y]);

        if (coordinate) {
          // Format coordinates in millimeters (x1000) with 2 decimal places
          const xCoord = coordinate[0] * 1000; // Convert meters to mm
          const yCoord = Math.abs(coordinate[1]) * 1000; // Convert meters to mm

          if (!isNaN(xCoord) && !isNaN(yCoord)) {
            onCoordinatesChange({
              x: xCoord.toFixed(2),
              y: yCoord.toFixed(2),
            });
          }
        }

        pending = false;
      });
    };

    // Add pointer move listener to the map container
    mapContainer.addEventListener("mousemove", handlePointerMove);

    return () => {
      mapContainer.removeEventListener("mousemove", handlePointerMove);
    };
  }, [map, mapContainer]);

  // This component doesn't render anything
  return null;
}
