// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { useCallback, useEffect, useRef } from "react";
import TileLayer from "ol/layer/WebGLTile";
import { useMaskContext } from "@/features/viewer/contexts/MaskContext";
import { MaskLayer } from "@/features/viewer/components/map/webglStyleUtils";
import { apiFetch } from "@/utils/fetchUtils";
import { useStudyMetadataField } from "@/api/hooks";

export interface UseMaskLayersResult {
  maskLayers: MaskLayer[];
  handleMaskLayerCreated: (maskLayer: TileLayer) => Promise<void>;
  toggleMaskLayerVisibility: (id: string) => void;
}

/**
 * Custom hook for managing mask layers state and API interactions.
 * Handles fetching mask metadata and creating MaskLayer objects with API-provided colors.
 *
 * @param slideUid - The slide identifier for fetching mask data
 * @param studyUid - The study identifier for fetching color metadata
 * @returns Object containing mask layers array and update functions
 */
export function useMaskLayers(
  slideUid?: string,
  studyUid?: string
): UseMaskLayersResult {
  const { maskLayers, setMaskLayers, toggleMaskLayerVisibility } =
    useMaskContext();

  // Track slideUid changes to clear old mask layers
  const prevSlideUidRef = useRef<string | undefined>(undefined);

  // Clear mask layers when slideUid changes
  useEffect(() => {
    if (
      prevSlideUidRef.current !== undefined &&
      prevSlideUidRef.current !== slideUid
    ) {
      console.log(
        `useMaskLayers: Slide changed from ${prevSlideUidRef.current} to ${slideUid}, clearing mask layers`
      );
      setMaskLayers([]);
    }
    prevSlideUidRef.current = slideUid;
  }, [slideUid, setMaskLayers]);

  // Fetch study metadata to get color and index mappings
  const { data: studyMetadata } = useStudyMetadataField(studyUid || "");

  // Memoize mask layer creation handler to prevent infinite loops
  const handleMaskLayerCreated = useCallback(
    async (maskLayer: TileLayer) => {
      console.log("Mask layer created:", maskLayer);

      // Clear any existing mask layers from context when a new slide loads
      // This prevents accumulation of layers from previous slides
      setMaskLayers([]);

      // Integrate with mask context by fetching mask metadata and creating MaskLayer objects
      if (slideUid) {
        try {
          const maskList = (await apiFetch(
            `/api/v1/slides/${slideUid}/annotations/raster`
          )) as any;

          if (maskList?.masks?.length > 0) {
            // Extract color and index mappings from study metadata
            const colorMap = studyMetadata?.metadata?.color_map || {};
            const indexMap = studyMetadata?.metadata?.index_map || {};

            // Helper function to get color for a mask index
            const getColorForIndex = (index: number): string => {
              const labelName = indexMap[index.toString()];
              if (labelName && colorMap[labelName]) {
                return colorMap[labelName];
              }
              // Fallback colors if not found in mapping
              const fallbackColors = [
                "#ff0000",
                "#00ff00",
                "#0000ff",
                "#ffff00",
                "#ff00ff",
                "#00ffff",
              ];
              return (
                fallbackColors[(index - 1) % fallbackColors.length] || "#ff0000"
              );
            };

            // Helper function to get name for a mask index
            const getNameForIndex = (
              index: number,
              originalName?: string
            ): string => {
              const labelName = indexMap[index.toString()];
              if (labelName) {
                return `${labelName} (index: ${index})`;
              }
              // Fallback to original name or generic name
              return originalName || `Mask ${index}`;
            };

            const maskLayerObjects: MaskLayer[] = maskList.masks.map(
              (mask: any, index: number) => ({
                id: mask.maskUid || `mask-${index}`,
                name: getNameForIndex(index + 1, mask.maskName),
                color: getColorForIndex(index + 1),
                visible: true,
                index: index + 1,
                maskName: mask.maskName,
              })
            );

            setMaskLayers(maskLayerObjects);
            console.log(
              `Added ${maskLayerObjects.length} mask layers to context:`,
              maskLayerObjects
            );
          }
        } catch (error) {
          console.error("Failed to fetch mask metadata for context:", error);
        }
      }
    },
    [slideUid, studyUid, studyMetadata, setMaskLayers]
  );

  return {
    maskLayers,
    handleMaskLayerCreated,
    toggleMaskLayerVisibility,
  };
}
