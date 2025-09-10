// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React, {
  useState,
  useRef,
  useEffect,
  useCallback,
  useMemo,
} from "react";
import type Map from "ol/Map";
import TileLayer from "ol/layer/WebGLTile";
import { useMaskContext } from "@/features/viewer/contexts/MaskContext";
import { createMaskStyle } from "@/features/viewer/components/map/webglStyleUtils";
import { useMaskLayers, useVectorLayers } from "@/features/viewer/hooks";

import {
  DisplayMetadata,
  SlideMetadata,
} from "@/features/viewer/components/map/types";
import {
  WholeSlideImage,
  type TileFormat,
} from "@/features/viewer/components/map/SlideImage";
import { SlideStyleOptions } from "@/features/viewer/components/map/slideStyleUtils";
import { LoadingOverlay } from "@/features/viewer/components/LoadingOverlay";
import MapCanvas from "@/features/viewer/components/map/MapCanvas";
import {
  useFluorescentStyle,
  useMapLoadSpinner,
} from "@/features/viewer/hooks";

import { onJXLLoadingStateChange } from "@/features/viewer/components/map/JXLLoader";

// Map component moved to MapCanvas.tsx

interface SlideViewerProps {
  /** Slide unique identifier */
  slideUid?: string;
  /** Study unique identifier for fetching metadata */
  studyUid?: string;
  /** Callback fired when slide metadata is loaded */
  onMetadataLoaded?: (metadata: DisplayMetadata) => void;
  /** Notify parent when the OpenLayers map is created */
  onMapReady?: (map: Map) => void;
  /** Raw metadata callback (e.g. for channels) */
  onRawMetadataLoaded?: (metadata: SlideMetadata) => void;
  /** Slide layer created callback */
  onSlideLayerCreated?: (
    layer: TileLayer,
    slideImage?: WholeSlideImage
  ) => void;
  /** Mask layer created callback */
  onMaskLayerCreated?: (layer: TileLayer) => void;
  /** Vector layer created callback */
  onVectorLayerCreated?: (map: Map, metadata: SlideMetadata) => void;
  /** Loading state change passthrough */
  onLoadingChange?: (isLoading: boolean) => void;
  /** Tile progress reporting */
  onTileProgress?: (progress: {
    inFlight: number;
    loaded: number;
    errors: number;
    started: number;
  }) => void;
  /** Viewer sensitivity/settings */
  panSensitivity?: number;
  zoomSensitivity?: number;
  quality?: number;
  /** Whether to show the measurement bar */
  showMeasurementBar?: boolean;
  /** Callback for non-default brightness settings state */
  onHasNonDefaultBrightnessSettings?: (hasNonDefault: boolean) => void;
}

/**
 * SlideViewer renders the WholeSlideImage and in-viewer controls.
 * External layout (e.g., CaseSlideBar, surrounding panels) is handled by SlideWorkspace.
 */
export default function SlideViewer({
  slideUid,
  studyUid,
  onMetadataLoaded,
  onMapReady,
  onRawMetadataLoaded,
  onSlideLayerCreated,
  onMaskLayerCreated,
  onVectorLayerCreated,
  onLoadingChange,
  onTileProgress,
  panSensitivity = 1.0,
  zoomSensitivity = 1.0,
  quality,
  showMeasurementBar = true,
  onHasNonDefaultBrightnessSettings,
}: SlideViewerProps) {
  // Mask context for opacity and visibility
  const { maskOpacity, showMask, maskColors, maskLayers } = useMaskContext();

  // Custom hooks for layer management
  const { handleMaskLayerCreated } = useMaskLayers(slideUid, studyUid);
  const { handleVectorLayerCreated } = useVectorLayers(slideUid);

  // Local state
  const [isLoading, setIsLoading] = useState(false);
  const [isJXLLoading, setIsJXLLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [maskLayer, setMaskLayer] = useState<TileLayer | null>(null);
  const [slideMetadata, setSlideMetadata] = useState<DisplayMetadata | null>(
    null
  );
  const [rawSlideMetadata, setRawSlideMetadata] =
    useState<SlideMetadata | null>(null);
  const [slideLayer, setSlideLayer] = useState<TileLayer | null>(null);
  // Spinner state from map load events

  // Refs
  const mapRef = useRef<Map | null>(null);
  const [mapForSpinner, setMapForSpinner] = useState<Map | null>(null);

  // Determine image characteristics from metadata
  const isFluorescent = useMemo(
    () => rawSlideMetadata?.imageTypeId === "img_type_fluor",
    [rawSlideMetadata]
  );

  const isRefreshing = useMapLoadSpinner(mapForSpinner);

  // Track JXL loading state
  useEffect(() => {
    const cleanup = onJXLLoadingStateChange(setIsJXLLoading);
    return cleanup;
  }, []);

  // Handle map creation from MapComponent
  const handleMapCreated = useCallback(
    (map: Map) => {
      mapRef.current = map;
      setMapForSpinner(map);

      if (onMapReady) {
        onMapReady(map);
      }
    },
    [onMapReady]
  );

  // Handle metadata loading
  const handleMetadataLoaded = useCallback(
    (metadata: DisplayMetadata) => {
      setSlideMetadata(metadata);
      onMetadataLoaded?.(metadata);
    },
    [onMetadataLoaded]
  );

  // Handle raw metadata loading (for channels and auto-detection)
  const handleRawMetadataLoaded = useCallback(
    (metadata: SlideMetadata) => {
      setRawSlideMetadata(metadata);
      onRawMetadataLoaded?.(metadata);
    },
    [onRawMetadataLoaded]
  );

  // Handle slide layer creation
  const handleSlideLayerCreated = useCallback(
    (layer: TileLayer, slideImage?: WholeSlideImage) => {
      setSlideLayer(layer);
      onSlideLayerCreated?.(layer, slideImage);
    },
    [onSlideLayerCreated]
  );

  // Apply initial fluorescent style once when ready
  const { applyStyle: applyFluorStyle } = useFluorescentStyle(
    slideLayer,
    rawSlideMetadata,
    slideUid,
    onHasNonDefaultBrightnessSettings
  );

  // Handle brightness/contrast style changes
  const handleStyleChange = useCallback(
    (options: SlideStyleOptions) => {
      applyFluorStyle(options);
    },
    [applyFluorStyle]
  );

  // Enhanced mask layer creation handler that also tracks local state
  const handleMaskLayerCreatedEnhanced = useCallback(
    async (layer: TileLayer) => {
      setMaskLayer(layer);
      await handleMaskLayerCreated(layer);
      onMaskLayerCreated?.(layer);
    },
    [handleMaskLayerCreated, onMaskLayerCreated]
  );

  const handleVectorLayerCreatedEnhanced = useCallback(
    async (map: Map, metadata: SlideMetadata) => {
      await handleVectorLayerCreated(map, metadata);
      onVectorLayerCreated?.(map, metadata);
    },
    [handleVectorLayerCreated, onVectorLayerCreated]
  );

  const handleLoadingChange = useCallback(
    (loading: boolean) => {
      setIsLoading(loading);
      onLoadingChange?.(loading);
    },
    [onLoadingChange]
  );

  // Update mask layer opacity and visibility (fast updates)
  useEffect(() => {
    if (maskLayer) {
      maskLayer.setOpacity(maskOpacity);
      maskLayer.setVisible(showMask);
    }
  }, [maskLayer, maskOpacity, showMask]);

  // Update WebGL style when mask layers or colors change (slower updates)
  useEffect(() => {
    if (maskLayer) {
      console.log(
        `🐌 SLOW UPDATE: Recompiling WebGL style for ${maskLayers.length} layers`
      );
      const startTime = performance.now();
      const style = createMaskStyle(maskLayers, maskColors);
      maskLayer.setStyle(style);
      const endTime = performance.now();
      console.log(
        `🐌 SLOW UPDATE: WebGL recompilation completed in ${
          endTime - startTime
        }ms`
      );
    }
  }, [maskLayer, maskLayers, maskColors]);

  if (!slideUid) {
    return (
      <div className="absolute inset-0 bg-gray-900 flex items-center justify-center">
        <div className="text-white text-center">
          <h2>No Slide Selected</h2>
          <p>Please select a slide to view.</p>
        </div>
      </div>
    );
  }

  return (
    <div className="absolute inset-0 bg-black flex flex-col min-h-0">
      {/* Main content area - Map (controls are inside the viewer) */}
      <div className="flex-1 flex min-h-0">
        {/* Map container - flex layout with panel and map content */}
        <div className="flex-1 flex min-h-0">
          {/* Map content area */}
          <div className="flex-1 relative min-h-0">
            {(isLoading || isJXLLoading) && (
              <LoadingOverlay
                text={
                  isJXLLoading
                    ? "Loading multiplex decoder..."
                    : "Loading slide..."
                }
              />
            )}

            {error && (
              <div className="absolute inset-0 flex items-center justify-center">
                <div className="bg-red-900/80 text-white p-4 rounded shadow-lg max-w-md text-center">
                  <p>{String(error)}</p>
                </div>
              </div>
            )}

            {/* Map Component */}
            <MapCanvas
              slideUid={slideUid}
              metadataLoaded={handleMetadataLoaded}
              onRawMetadataLoaded={handleRawMetadataLoaded}
              onMapCreated={handleMapCreated}
              onSlideLayerCreated={handleSlideLayerCreated}
              onMaskLayerCreated={handleMaskLayerCreatedEnhanced}
              onVectorLayerCreated={handleVectorLayerCreatedEnhanced}
              onLoadingChange={handleLoadingChange}
              onTileProgress={onTileProgress}
              panSensitivity={panSensitivity}
              zoomSensitivity={zoomSensitivity}
              quality={quality}
              showMeasurementBar={showMeasurementBar}
            />
          </div>
        </div>
      </div>
    </div>
  );
}
