// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { useCallback, useRef } from "react";
import Map from "ol/Map";
import View from "ol/View";
import WebGLTileLayer from "ol/layer/WebGLTile";
import TileImageLayer from "ol/layer/Tile";
import { getCenter } from "ol/extent";
import { WholeSlideImage, TileFormat } from "../components/map/SlideImage";
import { generateSlideStyle } from "../components/map/slideStyleUtils";
import { processMetadataForDisplay } from "../components/map/utils";
import { createMaskStyle } from "../components/map/webglStyleUtils";
import { apiFetch } from "@/utils/fetchUtils";
import {
  SlideMetadata,
  MaskList,
  Mask,
  DisplayMetadata,
  FetchError,
  ConfigurationError,
} from "../components/map/types";
import {
  DEFAULT_TILE_SIZE,
  DEFAULT_MASK_OPACITY,
  DEFAULT_PRELOAD_LEVEL,
  DEFAULT_INTERPOLATE_SLIDE,
} from "../components/map/constants";
import { dbg } from "@/features/viewer/utils/debug";

interface UseMapInitializationProps {
  onMetadataLoaded: (metadata: DisplayMetadata) => void;
  onRawMetadataLoaded?: (metadata: SlideMetadata) => void;
  onSlideLayerCreated?: (slideLayer: WebGLTileLayer | TileImageLayer) => void;
  onMaskLayerCreated: (maskLayer: WebGLTileLayer) => void;
  onVectorLayerCreated?: (map: Map, metadata: SlideMetadata) => void;
  onLoadingChange: (isLoading: boolean) => void;
  onTileProgress?: (progress: {
    inFlight: number;
    loaded: number;
    errors: number;
    started: number;
  }) => void;
  quality?: number; // Add quality parameter for JPG/JXL compression
}

export function useMapInitialization({
  onMetadataLoaded,
  onRawMetadataLoaded,
  onSlideLayerCreated,
  onMaskLayerCreated,
  onVectorLayerCreated,
  onLoadingChange,
  onTileProgress,
  quality,
}: UseMapInitializationProps) {
  // Track last initialized slide to prevent duplicates
  const lastInitializedSlideRef = useRef<string | null>(null);
  /**
   * Setup the slide layer
   */
  const setupSlideLayer = useCallback(
    (
      slideUid: string,
      slideMeta: SlideMetadata,
      format: TileFormat = "png",
      bandCount?: number
    ): WholeSlideImage => {
      const {
        slideWidth,
        slideHeight,
        slideMpp,
        tileSize = DEFAULT_TILE_SIZE,
      } = slideMeta;

      if (!slideWidth || !slideHeight) {
        throw new ConfigurationError("Missing slide dimensions in metadata");
      }

      let actualFormat = format;

      // Determine URL format based on actual tile format
      const getUrl = (z: number, x: number, y: number) => {
        const fileFormat = actualFormat === "jpg" ? "jpg" : actualFormat;
        let url = `/api/v1/slides/${slideUid}/tiles/${z}/${x}/${y}.${fileFormat}`;

        // Add quality parameter for JPG/JXL formats when specified
        if (
          quality !== undefined &&
          (actualFormat === "jpg" || actualFormat === "jxl")
        ) {
          url += `?q=${quality}`;
        }

        // Debug: Log tile URL generation for first few tiles
        // if (z <= 1 && x <= 1 && y <= 1) {
        //   console.log(`🔍 DEBUG: Generated tile URL for ${z}/${x}/${y}: ${url} (format=${actualFormat}, quality=${quality})`);
        // }

        return url;
      };

      // console.log(`🛠️ URL generator configured for format: ${actualFormat} (requested: ${format})`);

      const wsiSource = new WholeSlideImage({
        width: slideWidth,
        height: slideHeight,
        tileSize: tileSize,
        crossOrigin: "anonymous",
        interpolate: DEFAULT_INTERPOLATE_SLIDE,
        transition: 0,
        url: getUrl,
        mpp: slideMpp,
        tileFormat: format, // Pass original format, let WholeSlideImage handle fallback
        bandCount: bandCount,
        metadata: slideMeta.metadata, // Pass the metadata including channel information
        // Determine fluorescent mode from authoritative metadata to avoid timing issues
        isFluorescent: slideMeta.imageTypeId === "img_type_fluor",
      });
      dbg("WSI source created", {
        format,
        bandCount,
        slideWidth,
        slideHeight,
        tileSize,
        isFluorescent: slideMeta.imageTypeId === "img_type_fluor",
      });
      return wsiSource;
    },
    [quality]
  );

  /**
   * Setup the mask layer if available
   */
  const setupMaskLayer = useCallback(
    async (
      slideUid: string,
      slideMeta: SlideMetadata,
      slideSource: WholeSlideImage,
      tileSize: number,
      map: Map
    ): Promise<WebGLTileLayer | null> => {
      // Check if mask layer already exists
      const existingMaskLayers = map
        .getLayers()
        .getArray()
        .filter((layer) => layer.get("layerType") === "mask-main");

      if (existingMaskLayers.length > 0) {
        console.log(
          `⚠️ setupMaskLayer: Found ${existingMaskLayers.length} existing mask-main layers, skipping creation`
        );
        return existingMaskLayers[0] as WebGLTileLayer;
      }

      console.log(
        `🎭 setupMaskLayer: Creating mask layer for slide ${slideUid}`
      );

      try {
        const maskList = await apiFetch<MaskList>(
          `/api/v1/slides/${slideUid}/annotations/raster`
        );

        if (!maskList.masks || maskList.masks.length === 0) {
          return null;
        }

        const maskMeta: Mask = maskList.masks[0];

        if (!maskMeta.maskWidth || !maskMeta.maskHeight) {
          console.error(
            "Invalid mask dimensions:",
            maskMeta.maskWidth,
            maskMeta.maskHeight
          );
          return null;
        }

        const maskResolutionPixelsPerMeter = maskMeta.maskMpp
          ? 1e6 / maskMeta.maskMpp
          : 1e6;

        const maskSource = new WholeSlideImage({
          width: maskMeta.maskWidth,
          height: maskMeta.maskHeight,
          mpp: maskMeta.maskMpp,
          tileSize: tileSize,
          crossOrigin: "anonymous",
          interpolate: false,
          transition: 0,
          url: (z: number, x: number, y: number) => {
            if (maskMeta.tilesUrl && maskMeta.tilesUrl.trim() !== "") {
              return maskMeta.tilesUrl
                .replace("{z}", z.toString())
                .replace("{x}", x.toString())
                .replace("{y}", y.toString());
            } else {
              return `/api/v1/slides/${slideUid}/annotations/raster/${maskMeta.maskUid}/tiles/${z}/${x}/${y}.png`;
            }
          },
          pixelSize: maskResolutionPixelsPerMeter,
          tileFormat: "png", // Masks are always PNG for now
        });

        const maskLayer = new WebGLTileLayer({
          preload: DEFAULT_PRELOAD_LEVEL,
          extent: slideSource.getTileGrid().getExtent(),
          source: maskSource.getSource() as any,
          visible: true,
          opacity: DEFAULT_MASK_OPACITY,
          style: createMaskStyle([], []),
        } as any);

        // DEBUG: Mark this as the main mask layer
        maskLayer.set("layerType", "mask-main");
        maskLayer.set(
          "debugInfo",
          "Created by useMapInitialization.setupMaskLayer"
        );

        map.addLayer(maskLayer);
        map.renderSync();

        // DEBUG: Log mask layer creation
        console.log(
          `🎭 DEBUG: Created main mask layer for slide ${slideUid}`,
          maskLayer
        );
        console.log(
          `🎭 DEBUG: Map now has ${map
            .getLayers()
            .getLength()} layers after adding mask layer`
        );

        return maskLayer;
      } catch (error) {
        console.error("Error loading mask layer:", error);
        return null;
      }
    },
    []
  );

  /**
   * Initialize slide and mask layers
   */
  const initializeSlide = useCallback(
    async (slideUid: string, map: Map) => {
      console.log(
        `🔄 initializeSlide called for slideUid: ${slideUid}, lastInitialized: ${lastInitializedSlideRef.current}`
      );

      // Prevent double initialization of the same slide
      if (lastInitializedSlideRef.current === slideUid) {
        console.log(
          `⚠️ Skipping initialization - slide ${slideUid} already initialized`
        );
        return;
      }

      // Set the slideUid immediately to prevent concurrent calls
      const previousSlideUid = lastInitializedSlideRef.current;
      lastInitializedSlideRef.current = slideUid;

      console.log(
        `🔄 Initializing slide ${slideUid} (previous: ${previousSlideUid})`
      );

      // If switching to a different slide, we need to clear layers
      if (previousSlideUid && previousSlideUid !== slideUid) {
        console.log(
          `🧹 Switching from slide ${previousSlideUid} to ${slideUid}, clearing layers`
        );
      }

      // Check if we already have layers to prevent duplicate initialization
      const currentLayers = map.getLayers().getArray();
      if (currentLayers.length > 0) {
        // Preserve annotation layers (Z-index 1500+) during re-initialization
        const annotationLayers: any[] = [];
        currentLayers.forEach((layer) => {
          const zIndex = layer.getZIndex();
          if (zIndex !== undefined && zIndex >= 1500) {
            annotationLayers.push(layer);
          }
        });

        // Clear all layers
        map.getLayers().clear();

        // Re-add preserved annotation layers
        annotationLayers.forEach((layer) => {
          map.addLayer(layer);
        });
      }
      try {
        onLoadingChange(true);

        // Fetch slide metadata (including channels)

        const slideMeta = await apiFetch<SlideMetadata>(
          `/api/v1/slides/${slideUid}`
        );

        // img_type_fluor = JXL, normal images = JPG, masks = PNG
        const finalTileFormat: TileFormat =
          slideMeta.imageTypeId === "img_type_fluor" ? "jxl" : "jpg";
        const finalBandCount =
          slideMeta.imageTypeId === "img_type_fluor" ? undefined : 4; // Let JXL determine bands, use 4 for JPG

        // Process metadata for display
        const displayMetadata = processMetadataForDisplay(slideMeta);
        onMetadataLoaded(displayMetadata);

        // Pass raw metadata for channel information
        if (onRawMetadataLoaded) {
          onRawMetadataLoaded(slideMeta);
        }

        // Note: Layer clearing is already handled at the start of this function
        // We preserved annotation layers (Z-index >= 1500) and only need to add new slide/mask layers

        // Create new slide layer with auto-detected tile format
        const newSlideSource = setupSlideLayer(
          slideUid,
          slideMeta,
          finalTileFormat,
          finalBandCount
        );

        // Create slide layer
        const baseLayerOptions = {
          preload: 8,
          extent: newSlideSource.getTileGrid().getExtent(),
          source: newSlideSource.getSource() as any,
          properties: {
            uniqueId: `layer-${Date.now()}`,
            tileFormat: finalTileFormat,
            wholeSlideImageSource: newSlideSource,
          },
        } as const;

        const slideLayer =
          finalTileFormat === "jxl"
            ? new WebGLTileLayer({
                ...baseLayerOptions,
              } as any)
            : new TileImageLayer({
                ...baseLayerOptions,
              } as any);

        // Mark slide layer with proper type for debugging
        slideLayer.set("layerType", "slide-main");
        slideLayer.set(
          "debugInfo",
          `Slide layer for ${slideUid} (${finalTileFormat})`
        );

        dbg("Slide layer created", { finalTileFormat });

        // For JXL/DataTile sources, ensure a safe WebGL style is set
        if (finalTileFormat === "jxl") {
          try {
            const channelMeta = slideMeta.metadata?.channels;
            if (channelMeta && Object.keys(channelMeta).length > 0) {
              const defaultChannelParams: Record<
                string,
                { visible: boolean; min: number; max: number }
              > = {};
              Object.keys(channelMeta).forEach((id) => {
                defaultChannelParams[id] = {
                  visible: true,
                  min: 0,
                  max: 65535,
                };
              });
              const initialStyle = generateSlideStyle({
                showGrayscale: false,
                invertBackground: false,
                gamma: 1.0,
                channels: defaultChannelParams,
                channelMetadata: channelMeta,
              });
              (slideLayer as WebGLTileLayer).setStyle(initialStyle as any);
              dbg(
                "Applied initial JXL style with channels",
                Object.keys(channelMeta)
              );
            } else {
              // Safe grayscale fallback using band 1 for RGB to avoid out-of-range band access
              const fallbackStyle = {
                color: ["array", ["band", 1], ["band", 1], ["band", 1], 1],
                gamma: 1.0,
              } as any;
              (slideLayer as WebGLTileLayer).setStyle(fallbackStyle);
              dbg("Applied fallback grayscale style");
            }
          } catch (e) {
            console.warn(
              "Failed to set initial JXL style; viewer may apply later",
              e
            );
          }
        }

        // Add slide layer to map and schedule a render to initialize WebGL resources
        map.addLayer(slideLayer);
        // Use requestAnimationFrame to avoid immediate GL flush during mutation
        try {
          requestAnimationFrame(() => {
            try {
              (map as any).renderSync?.();
            } catch {}
          });
        } catch {}
        dbg("Slide layer added to map");

        // Notify about slide layer creation
        if (onSlideLayerCreated) {
          onSlideLayerCreated(slideLayer as any);
        }

        // Hook into tile loading progress if callback provided
        if (onTileProgress) {
          // OpenLayers sources emit tileloadstart/tileloadend/tileloaderror events
          let inFlight = 0;
          let loaded = 0;
          let errors = 0;
          let started = 0;

          const source: any = (slideLayer as any).getSource?.();
          const handleStart = () => {
            inFlight += 1;
            started += 1;
            onTileProgress({ inFlight, loaded, errors, started });
          };
          const handleEnd = () => {
            inFlight = Math.max(0, inFlight - 1);
            loaded += 1;
            onTileProgress({ inFlight, loaded, errors, started });
          };
          const handleError = () => {
            inFlight = Math.max(0, inFlight - 1);
            errors += 1;
            onTileProgress({ inFlight, loaded, errors, started });
          };

          source?.on?.("tileloadstart", handleStart);
          source?.on?.("tileloadend", handleEnd);
          source?.on?.("tileloaderror", handleError);

          // Clean up listeners when layers are rebuilt
          map.once("change:layergroup", () => {
            source?.un?.("tileloadstart", handleStart);
            source?.un?.("tileloadend", handleEnd);
            source?.un?.("tileloaderror", handleError);
          });
        }

        // Update map view with extended resolutions for digital zoom
        const tileGridResolutions = newSlideSource
          .getTileGrid()
          .getResolutions();
        const maxTileResolution =
          tileGridResolutions[tileGridResolutions.length - 1];

        // Add extra zoom levels beyond tile grid for digital zoom (2x, 4x, 8x)
        const extraDigitalZoomLevels = 3;
        const digitalZoomResolutions = Array.from(
          { length: extraDigitalZoomLevels },
          (_, i) => maxTileResolution / 2 ** (i + 1)
        );

        // Combine tile grid resolutions with digital zoom resolutions
        const allResolutions = [
          ...tileGridResolutions,
          ...digitalZoomResolutions,
        ];

        // Always start at the most zoomed-out level
        const extent = newSlideSource.getTileGrid().getExtent();
        const view = new View({
          projection: newSlideSource.getProjection()!,
          center: getCenter(extent),
          resolutions: allResolutions,
          // Zoom index 0 corresponds to the lowest resolution (most zoomed out)
          zoom: 0,
          showFullExtent: true,
          extent: extent,
        });
        map.setView(view);

        // Create mask layer if available
        const maskLayer = await setupMaskLayer(
          slideUid,
          slideMeta,
          newSlideSource,
          slideMeta.tileSize || DEFAULT_TILE_SIZE,
          map
        );

        if (maskLayer) {
          onMaskLayerCreated(maskLayer);
        }

        // Setup vector layers if callback is provided
        if (onVectorLayerCreated) {
          onVectorLayerCreated(map, slideMeta);
        }

        // Mark slide as successfully initialized
        lastInitializedSlideRef.current = slideUid;

        onLoadingChange(false);
      } catch (error) {
        console.error("❌ Error during slide initialization:", error);
        handleMapError(slideUid, error);
        onLoadingChange(false);
      }
    },
    [
      setupSlideLayer,
      setupMaskLayer,
      onMetadataLoaded,
      onMaskLayerCreated,
      onVectorLayerCreated,
      onLoadingChange,
      quality, // Include quality parameter
    ]
  );

  /**
   * Handle errors during map initialization
   */
  const handleMapError = useCallback((slideUid: string, error: unknown) => {
    console.error("Error loading slide/mask:", error);

    let errorMessage = `Failed to load slide: ${slideUid}. Please check if the slide exists and try again.`;

    if (error instanceof FetchError) {
      errorMessage = `Failed to fetch slide data: ${error.status} ${error.statusText}`;
    } else if (error instanceof ConfigurationError) {
      errorMessage = `Slide configuration error: ${error.message}`;
    } else if (error instanceof Error) {
      errorMessage = `Error: ${error.message}`;
    }

    // Note: Error display logic could be extracted to a separate utility
    // For now, keeping the DOM manipulation here to maintain existing behavior
    const errorElement = document.createElement("div");
    errorElement.className =
      "text-red-500 bg-black/80 p-4 rounded absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2";
    errorElement.textContent = errorMessage;

    // Find the map container (this is a simplification - in practice you'd pass the container ref)
    const mapContainer =
      document.querySelector(".ol-viewport")?.parentElement ||
      document.querySelector('[class*="map"]') ||
      document.body;
    mapContainer.appendChild(errorElement);
  }, []);

  return {
    initializeSlide,
    handleMapError,
  };
}
