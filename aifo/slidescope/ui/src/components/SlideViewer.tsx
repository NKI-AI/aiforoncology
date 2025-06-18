// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
import { useEffect, useRef, useState } from "react";
import Map from "ol/Map";
import WebGLTileLayer from "ol/layer/WebGLTile";
import MapComponent from "./map/MapComponent";
import Crosshair from "./map/Crosshair";
import CoordinateTracker from "./map/CoordinateTracker";
import { DisplayMetadata } from "./map/types";
import { DEFAULT_MASK_OPACITY } from "./map/constants";
import { MaskColor } from "../types";
import { useMaskContext } from "../contexts/MaskContext";

interface Coordinates {
  x: string;
  y: string;
}

interface CursorPosition {
  x: number;
  y: number;
}

interface SlideViewerProps {
  slideId?: string;
  showCrosshair: boolean;
  onMetadataLoaded?: (metadata: DisplayMetadata) => void;
}

export default function SlideViewer({
  slideId,
  showCrosshair = false,
  onMetadataLoaded,
}: SlideViewerProps) {
  const { maskOpacity, showMask, maskColors } = useMaskContext();

  const mapRef = useRef<Map | null>(null);
  const mapContainerRef = useRef<HTMLDivElement>(null);
  const maskLayerRef = useRef<WebGLTileLayer | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [coordinates, setCoordinates] = useState<Coordinates>({
    x: "0.0",
    y: "0.0",
  });
  const [cursorPosition, setCursorPosition] = useState<CursorPosition>({
    x: 0,
    y: 0,
  });
  const [error, setError] = useState<string | null>(null);

  // Check if slideId is valid
  useEffect(() => {
    if (!slideId) {
      setError("No slide ID provided. Please select a slide to view.");
      setIsLoading(false);
    } else {
      setError(null);

      // Reset mask layer reference when slideId changes
      console.log("[DEBUG] Slide ID changed, resetting mask layer reference");
      maskLayerRef.current = null;
    }
  }, [slideId]);

  // Update mask opacity when it changes
  useEffect(() => {
    if (maskLayerRef.current) {
      console.log("[DEBUG] Updating mask opacity to:", maskOpacity);
      maskLayerRef.current.setOpacity(maskOpacity);
    } else {
      console.log("[DEBUG] Cannot update opacity: maskLayer is null");
    }
  }, [maskOpacity]);

  // Update mask visibility when it changes
  useEffect(() => {
    if (maskLayerRef.current) {
      console.log("[DEBUG] Updating mask visibility to:", showMask);
      maskLayerRef.current.setVisible(showMask);
    } else {
      console.log("[DEBUG] Cannot update visibility: maskLayer is null");
    }
  }, [showMask]);

  // Update mask colors when they change
  useEffect(() => {
    if (maskLayerRef.current) {
      const maskStyle = {
        color: generateColorExpression(maskColors),
        gamma: 1.0,
      };
      maskLayerRef.current.setStyle(maskStyle);
    }
  }, [maskColors]);

  // Generate the color expression using current mask colors
  const generateColorExpression = (colors: MaskColor[]) => {
    // Start building our case expression
    const colorExpression: any[] = ["case"];

    // First case is always transparent for value 0
    colorExpression.push(["==", ["*", ["band", 1], 255], 0], [0, 0, 0, 0.0]);

    // Add each color from the maskColors array
    colors.forEach((color, index) => {
      const labelValue = index + 1;
      // Use the alpha value from the color object to control visibility
      colorExpression.push(
        ["==", ["*", ["band", 1], 255], labelValue],
        [color.r, color.g, color.b, color.a]
      );
    });
    colorExpression.push([0, 0, 0, 0.0]);

    return colorExpression;
  };

  // Handle map creation
  const handleMapCreated = (map: Map) => {
    console.log("[DEBUG] Map created and received in SlideViewer");
    mapRef.current = map;
  };

  // Handle mask layer creation
  const handleMaskLayerCreated = (maskLayer: WebGLTileLayer) => {
    console.log("[DEBUG] Mask layer received in SlideViewer");
    maskLayerRef.current = maskLayer;

    // Apply current settings (in case they've changed during loading)
    console.log(
      "[DEBUG] Initial mask settings - opacity:",
      maskOpacity,
      "visible:",
      showMask
    );

    // Initialize with current colors
    const maskStyle = {
      color: generateColorExpression(maskColors),
      gamma: 1.0,
    };

    // Apply the style, opacity and visibility:
    maskLayer.setStyle(maskStyle);
    maskLayer.setOpacity(maskOpacity);
    maskLayer.setVisible(showMask);
  };

  // Handle metadata loaded
  const handleMetadataLoaded = (metadata: DisplayMetadata) => {
    if (onMetadataLoaded) {
      onMetadataLoaded(metadata);
    }
  };

  return (
    <div id="view" className="absolute inset-0 pt-12 bg-black">
      <div ref={mapContainerRef} className="w-full h-full relative">
        {isLoading && <div className="loading-spinner"></div>}

        {error && (
          <div className="absolute inset-0 flex items-center justify-center">
            <div className="bg-red-900/80 text-white p-4 rounded shadow-lg max-w-md text-center">
              <p>{error}</p>
            </div>
          </div>
        )}

        {/* Map Component - only render if we have a slideId */}
        {slideId && (
          <MapComponent
            slideId={slideId}
            metadataLoaded={handleMetadataLoaded}
            onMapCreated={handleMapCreated}
            onMaskLayerCreated={handleMaskLayerCreated}
            onLoadingChange={setIsLoading}
          />
        )}

        {/* Coordinate Tracker */}
        {mapRef.current && (
          <CoordinateTracker
            map={mapRef.current}
            mapContainer={mapContainerRef.current}
            onPositionChange={setCursorPosition}
            onCoordinatesChange={setCoordinates}
          />
        )}

        {/* Crosshair */}
        <Crosshair
          show={showCrosshair}
          position={cursorPosition}
          coordinates={coordinates}
        />
      </div>
    </div>
  );
}
