// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
import { useEffect } from "react";
import Map from "ol/Map";
import { WholeSlideImage } from "./SlideImage";
import {
  createOverviewMapControl,
  forceOverviewPosition,
} from "./OverviewMapFactory";
import "./OverviewMap.css";

interface OverviewMapControlProps {
  map: Map;
  source: WholeSlideImage;
  collapsed?: boolean;
}

/**
 * Component that creates and manages an overview map control
 */
export function OverviewMapControl({
  map,
  source,
  collapsed = true,
}: OverviewMapControlProps) {
  useEffect(() => {
    // Create the overview map control
    const overviewMapControl = createOverviewMapControl(map, source, collapsed);

    // Add the control to the map
    map.addControl(overviewMapControl);

    // Force the overview map position at different times to ensure it's applied
    setTimeout(forceOverviewPosition, 0);
    setTimeout(forceOverviewPosition, 100);
    setTimeout(forceOverviewPosition, 500);

    // Clean up on unmount
    return () => {
      map.removeControl(overviewMapControl);
    };
  }, [map, source, collapsed]);

  // This is a functional component that doesn't render anything directly
  return null;
}
