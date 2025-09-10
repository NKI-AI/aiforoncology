// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { useEffect, useRef } from "react";
import Map from "ol/Map";
import { OverviewMap } from "ol/control";
import { WholeSlideImage } from "./SlideImage";
import { createOverviewMapControl } from "./OverviewMapFactory";
import {
  forceOverviewPositionWithRetry,
  ensureOverviewButtonVisibility,
} from "./overviewUtils";
import "./OverviewMap.css";

interface OverviewMapControlProps {
  map: Map;
  source: WholeSlideImage | null;
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
  const overviewMapControlRef = useRef<OverviewMap | null>(null);

  // Create the overview map control once when map is available
  useEffect(() => {
    if (!map) return;

    const cleanup = () => {
      if (overviewMapControlRef.current) {
        map.removeControl(overviewMapControlRef.current);
        overviewMapControlRef.current = null;
      }
    };

    return cleanup;
  }, [map]);

  // Update the overview map when source changes
  useEffect(() => {
    if (!map || !source) {
      // Remove overview map if no source
      if (overviewMapControlRef.current) {
        map.removeControl(overviewMapControlRef.current);
        overviewMapControlRef.current = null;
      }
      return;
    }

    // Remove existing overview map if it exists
    if (overviewMapControlRef.current) {
      map.removeControl(overviewMapControlRef.current);
    }

    try {
      // Create new overview map control with the new source
      const overviewMapControl = createOverviewMapControl(
        map,
        source,
        collapsed
      );
      overviewMapControlRef.current = overviewMapControl;

      // Add the control to the map
      map.addControl(overviewMapControl);

      // Simplified setup without complex retry logic
      const setupOverviewMap = async () => {
        await forceOverviewPositionWithRetry();
        await ensureOverviewButtonVisibility();
      };

      setupOverviewMap().catch(() => {
        // Silently handle errors to avoid console noise
      });
    } catch (error) {
      // Silently handle errors to avoid console noise
    }
  }, [map, source, collapsed]);

  return null;
}
