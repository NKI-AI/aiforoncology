// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { useEffect, useRef, useState, useCallback } from "react";
import Map from "ol/Map";
import TileLayer from "ol/layer/WebGLTile";
import {
  defaults as defaultInteractions,
  DragPan,
  MouseWheelZoom,
} from "ol/interaction";
import { useMapInitialization } from "./useMapInitialization";
import { WholeSlideImage, TileFormat } from "../components/map/SlideImage";
import { SlideMetadata, DisplayMetadata } from "../components/map/types";

// Custom DragPan with adjustable pan sensitivity
class CustomDragPan extends DragPan {
  panSensitivity: number;

  constructor(options: any = {}) {
    super(options);
    this.panSensitivity = options.panSensitivity ?? 1.0;
  }

  handleDragEvent(event: any) {
    event.deltaX *= this.panSensitivity;
    event.deltaY *= this.panSensitivity;
    super.handleDragEvent(event);
  }
}

interface UseSlideMapProps {
  slideUid: string;
  metadataLoaded: (metadata: DisplayMetadata) => void;
  onRawMetadataLoaded?: (metadata: SlideMetadata) => void;
  onSlideLayerCreated?: (slideLayer: any) => void;
  onMaskLayerCreated: (maskLayer: TileLayer) => void;
  onVectorLayerCreated?: (map: Map, metadata: SlideMetadata) => void;
  onLoadingChange: (isLoading: boolean) => void;
  onTileProgress?: (progress: {
    inFlight: number;
    loaded: number;
    errors: number;
    started: number;
  }) => void;
  panSensitivity?: number;
  zoomSensitivity?: number;
  quality?: number;
}

interface UseSlideMapReturn {
  map: Map | null;
  slideSource: WholeSlideImage | null;
  mapContainerRef: React.RefObject<HTMLDivElement>;
  initializeMap: () => (() => void) | undefined;
}

/**
 * Custom hook for managing OpenLayers map with slide data, including sensitivity controls
 */
export function useSlideMap({
  slideUid,
  metadataLoaded,
  onRawMetadataLoaded,
  onSlideLayerCreated,
  onMaskLayerCreated,
  onVectorLayerCreated,
  onLoadingChange,
  onTileProgress,
  panSensitivity,
  zoomSensitivity,
  quality,
}: UseSlideMapProps): UseSlideMapReturn {
  const mapContainerRef = useRef<HTMLDivElement>(null);
  const [map, setMap] = useState<Map | null>(null);
  const [slideSource, setSlideSource] = useState<WholeSlideImage | null>(null);
  const customDragPanRef = useRef<CustomDragPan | null>(null);

  // Use the existing map initialization hook
  const { initializeSlide } = useMapInitialization({
    onMetadataLoaded: metadataLoaded,
    onRawMetadataLoaded,
    onSlideLayerCreated,
    onMaskLayerCreated,
    onVectorLayerCreated,
    onLoadingChange,
    onTileProgress,
    quality,
  });

  // Initialize map
  const initializeMap = useCallback(() => {
    if (!mapContainerRef.current) {
      return;
    }

    // Create custom interactions
    const customDragPan = new CustomDragPan({
      panSensitivity: panSensitivity || 1.0,
    });
    const standardMouseWheelZoom = new MouseWheelZoom({
      maxDelta: zoomSensitivity || 1.0,
      duration: 250,
      timeout: 80,
      useAnchor: true,
    });

    // Store references for later updates
    customDragPanRef.current = customDragPan;

    // Create initial empty map with custom interactions
    const initialMap = new Map({
      target: mapContainerRef.current,
      maxTilesLoading: 100,
      interactions: defaultInteractions({
        dragPan: false,
        mouseWheelZoom: false,
      }).extend([customDragPan, standardMouseWheelZoom]),
      controls: [], // Start with no controls, we'll add them via components
    });

    setMap(initialMap);

    return () => {
      // Clean up the map when component unmounts
      if (initialMap) {
        initialMap.setTarget(undefined);
      }
    };
  }, [panSensitivity, zoomSensitivity]);

  // Update pan sensitivity when it changes
  useEffect(() => {
    if (panSensitivity === undefined || !customDragPanRef.current) return;
    customDragPanRef.current.panSensitivity = panSensitivity;
  }, [panSensitivity]);

  // Update zoom sensitivity when it changes
  useEffect(() => {
    if (zoomSensitivity === undefined || !map) return;

    // Get current interactions and replace the MouseWheelZoom
    const interactions = map.getInteractions();
    const currentMouseWheelZoom = interactions
      .getArray()
      .find((interaction) => interaction instanceof MouseWheelZoom);

    if (currentMouseWheelZoom) {
      interactions.remove(currentMouseWheelZoom);

      // Create new MouseWheelZoom with updated sensitivity
      const newMouseWheelZoom = new MouseWheelZoom({
        maxDelta: zoomSensitivity,
        duration: 250,
        timeout: 80,
        useAnchor: true,
      });

      interactions.push(newMouseWheelZoom);
    }
  }, [zoomSensitivity, map]);

  // Handle slide changes
  useEffect(() => {
    if (!slideUid || !map) {
      onLoadingChange(false);
      setSlideSource(null);
      return;
    }

    // Initialize slide and capture the source
    const initializeAndCaptureSource = async () => {
      try {
        await initializeSlide(slideUid, map);

        // Find the newly added slide layer (should be the first layer)
        const layers = map.getLayers().getArray();

        if (layers.length > 0) {
          const slideLayer = layers[0] as TileLayer;
          const source = slideLayer.getSource();
          if (source && source instanceof WholeSlideImage) {
            setSlideSource(source);
          } else {
            console.warn(
              "⚠️  Slide layer found but source is not WholeSlideImage:",
              source
            );
          }
        } else {
          console.warn("⚠️  No layers found after slide initialization");
        }
      } catch (error) {
        console.error("❌ Error during slide initialization:", error);
        setSlideSource(null);

        // Throw critical errors to stop the feedback loop
        if (error instanceof Error && error.message.includes("abort")) {
          return;
        }

        // For other errors, we might want to stop the cycle
        throw new Error(`Slide initialization failed: ${error}`);
      }
    };

    initializeAndCaptureSource();
  }, [slideUid, map, initializeSlide, onLoadingChange]);

  return {
    map,
    slideSource,
    mapContainerRef,
    initializeMap,
  };
}
