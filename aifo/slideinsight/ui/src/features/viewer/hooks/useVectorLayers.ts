// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { useCallback } from "react";
import Map from "ol/Map";
import { useVectorContext } from "@/features/viewer/contexts/VectorContext";
import { useVectorLayer } from "@/features/viewer/hooks/useVectorLayer";
import { SlideMetadata } from "@/features/viewer/components/map/types";

interface VectorLayer {
  id: string;
  name: string;
  color: string;
  visible: boolean;
  defaultColor?: string;
  vectorName?: string;
}

export interface UseVectorLayersResult {
  vectorLayers: VectorLayer[];
  handleVectorLayerCreated: (
    map: Map,
    metadata: SlideMetadata
  ) => Promise<void>;
  updateVectorLayer: (id: string, updates: Partial<VectorLayer>) => void;
  toggleVectorLayerVisibility: (id: string) => void;
}

/**
 * Custom hook for managing vector layers state and API interactions.
 * Delegates to the existing useVectorLayer hook for vector layer creation.
 *
 * @param slideUid - The slide identifier for vector layer management
 * @returns Object containing vector layers array and update functions
 */
export function useVectorLayers(slideUid?: string): UseVectorLayersResult {
  const { vectorLayers, updateVectorLayer, toggleVectorLayerVisibility } =
    useVectorContext();

  // Use the existing vector layer hook for proper label handling
  const { handleVectorLayerCreated: vectorLayerCreatedHandler } =
    useVectorLayer({
      slideUid,
    });

  // Memoize vector layer creation handler to prevent infinite loops
  const handleVectorLayerCreated = useCallback(
    async (map: Map, metadata: SlideMetadata) => {
      // Call the vector layer hook handler directly
      if (slideUid && vectorLayerCreatedHandler) {
        await vectorLayerCreatedHandler(map, metadata);
      }
    },
    [slideUid, vectorLayerCreatedHandler]
  );

  return {
    vectorLayers,
    handleVectorLayerCreated,
    updateVectorLayer,
    toggleVectorLayerVisibility,
  };
}
