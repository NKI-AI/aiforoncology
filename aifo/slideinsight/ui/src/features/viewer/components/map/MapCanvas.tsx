// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React, { useEffect } from "react";
import Map from "ol/Map";
import TileLayer from "ol/layer/WebGLTile";
import { useSlideMap } from "@/features/viewer/hooks";
import { OverviewMapControl } from "@/features/viewer/components/map/OverviewMapControl";
import { MapControls } from "@/features/viewer/components/map/MapControls";
import type {
  DisplayMetadata,
  SlideMetadata,
} from "@/features/viewer/components/map/types";
import type { TileFormat } from "@/features/viewer/components/map/SlideImage";

interface MapCanvasProps {
  slideUid: string;
  metadataLoaded: (metadata: DisplayMetadata) => void;
  onRawMetadataLoaded?: (metadata: SlideMetadata) => void;
  onMapCreated: (map: Map) => void;
  onSlideLayerCreated?: (slideLayer: TileLayer) => void;
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
  maxTilesLoading?: number;
  showMeasurementBar?: boolean;
}

export default function MapCanvas({
  slideUid,
  metadataLoaded,
  onRawMetadataLoaded,
  onMapCreated,
  onSlideLayerCreated,
  onMaskLayerCreated,
  onVectorLayerCreated,
  onLoadingChange,
  onTileProgress,
  panSensitivity,
  zoomSensitivity,
  quality,
  showMeasurementBar = true,
}: MapCanvasProps) {
  const { map, slideSource, mapContainerRef, initializeMap } = useSlideMap({
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
  });

  useEffect(() => {
    if (mapContainerRef.current) {
      const cleanup = initializeMap();
      return cleanup;
    }
  }, [initializeMap, mapContainerRef]);

  useEffect(() => {
    if (map) {
      onMapCreated(map);
    }
  }, [map, onMapCreated]);

  return (
    <div ref={mapContainerRef} className="w-full h-full">
      {map && (
        <>
          <OverviewMapControl map={map} source={slideSource} collapsed={true} />
          <MapControls map={map} showScaleLine={showMeasurementBar} />
        </>
      )}
    </div>
  );
}
