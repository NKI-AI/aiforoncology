// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { useEffect } from "react";
import Map from "ol/Map";
import {
  ScaleLine,
  MousePosition,
  defaults as defaultControls,
  Control,
} from "ol/control";
import { format, Coordinate } from "ol/coordinate";

interface MapControlsProps {
  map: Map;
  showScaleLine?: boolean;
}

/**
 * Component that adds standard controls to the map
 */
export function MapControls({ map, showScaleLine = true }: MapControlsProps) {
  useEffect(() => {
    const controls: Control[] = [];

    // Add scale line if requested
    if (showScaleLine) {
      const scaleControl = new ScaleLine({
        units: "metric",
        bar: false,
        steps: 1,
        minWidth: 140,
      });
      controls.push(scaleControl);
    }

    // Add all controls to the map
    controls.forEach((control) => map.addControl(control));

    // Clean up on unmount
    return () => {
      controls.forEach((control) => map.removeControl(control));
    };
  }, [map, showScaleLine]);

  // This is a functional component that doesn't render anything directly
  return null;
}
