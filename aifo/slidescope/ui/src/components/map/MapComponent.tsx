// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
import { useEffect, useRef, useState } from "react";
import Map from "ol/Map";
import TileLayer from "ol/layer/WebGLTile";
import { WholeSlideImage } from "./SlideImage";
import { OverviewMapControl } from "./OverviewMapControl";
import { MapControls } from "./MapControls";
import { createMap } from "./MapUtils";
import { processMetadataForDisplay } from "./utils";
import { apiFetch } from "../../utils/fetchUtils";
import {
  SlideMetadata,
  MaskMetadata,
  MaskList,
  Mask,
  DisplayMetadata,
  FetchError,
  ConfigurationError,
} from "./types";
import {
  DEFAULT_TILE_SIZE,
  DEFAULT_MASK_OPACITY,
  DEFAULT_PRELOAD_LEVEL,
  DEFAULT_INTERPOLATE_SLIDE,
  DEFAULT_INTERPOLATE_MASK,
} from "./constants";

// Simple logger utility
const logger = {
  debug: (message: string, ...args: any[]) => {
    console.log(`[DEBUG] ${message}`, ...args);
  },
  info: (message: string, ...args: any[]) => {
    console.log(`[INFO] ${message}`, ...args);
  },
  warn: (message: string, ...args: any[]) => {
    console.warn(`[WARN] ${message}`, ...args);
  },
  error: (message: string, ...args: any[]) => {
    console.error(`[ERROR] ${message}`, ...args);
  },
};

interface MapComponentProps {
  slideId: string;
  metadataLoaded: (metadata: DisplayMetadata) => void;
  onMapCreated: (map: Map) => void;
  onMaskLayerCreated: (maskLayer: TileLayer) => void;
  onLoadingChange: (isLoading: boolean) => void;
}

/**
 * Core map component that initializes the OpenLayers map with slide data
 */
export default function MapComponent({
  slideId,
  metadataLoaded,
  onMapCreated,
  onMaskLayerCreated,
  onLoadingChange,
}: MapComponentProps) {
  const mapContainerRef = useRef<HTMLDivElement>(null);
  const [map, setMap] = useState<Map | null>(null);
  const [slideSource, setSlideSource] = useState<WholeSlideImage | null>(null);

  useEffect(() => {
    if (!slideId || !mapContainerRef.current) {
      logger.warn(
        "Missing slideId or map container not ready. Using default configuration."
      );
      onLoadingChange(false);
      return;
    }

    // Log the actual slideId being used for this effect run
    logger.info(`Initializing map for slide: ${slideId}`);

    const initializeMap = async () => {
      try {
        onLoadingChange(true);

        // Fetch slide metadata with improved error handling
        logger.debug(`Fetching slide metadata for: ${slideId}`);
        const slideMeta = await apiFetch<SlideMetadata>(
          `/api/v1/slides/${slideId}/metadata`
        );

        // Process metadata for display using utility function
        const displayMetadata = processMetadataForDisplay(slideMeta);
        metadataLoaded(displayMetadata);

        // Create slide layer and setup map
        const slideSource = setupSlideLayer(slideMeta);
        const map = createMap(slideSource, mapContainerRef.current!);

        // Store references
        setMap(map);
        setSlideSource(slideSource);

        // Notify parent that map was created
        onMapCreated(map);

        // Create mask layer if available - use the current slideId captured in the closure
        const maskLayer = await setupMaskLayer(
          slideId,
          slideMeta,
          slideSource,
          slideMeta.tileSize || DEFAULT_TILE_SIZE,
          map
        );

        // Notify parent about mask layer if it was created
        if (maskLayer) {
          onMaskLayerCreated(maskLayer);
        } else {
          logger.debug(`No mask layer created for: ${slideId}`);
        }

        onLoadingChange(false);
      } catch (e) {
        handleMapError(e);
        onLoadingChange(false);
      }
    };

    initializeMap();

    return () => {
      // Clean up the map when component unmounts
      if (map) {
        map.setTarget(undefined);
        setMap(null);
      }
    };
  }, [slideId]);

  /**
   * Handle errors during map initialization
   */
  const handleMapError = (error: unknown) => {
    logger.error("Error loading slide/mask:", error);

    // Create user-friendly error message
    let errorMessage = `Failed to load slide: ${slideId}. Please check if the slide exists and try again.`;

    if (error instanceof FetchError) {
      errorMessage = `Failed to fetch slide data: ${error.status} ${error.statusText}`;
    } else if (error instanceof ConfigurationError) {
      errorMessage = `Slide configuration error: ${error.message}`;
    } else if (error instanceof Error) {
      errorMessage = `Error: ${error.message}`;
    }

    // Display error message to user
    const errorElement = document.createElement("div");
    errorElement.className =
      "text-red-500 bg-black/80 p-4 rounded absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2";
    errorElement.textContent = errorMessage;
    mapContainerRef.current?.appendChild(errorElement);
  };

  /**
   * Setup the slide layer
   */
  const setupSlideLayer = (slideMeta: SlideMetadata): WholeSlideImage => {
    const {
      slideWidth,
      slideHeight,
      slideMpp,
      tileSize = DEFAULT_TILE_SIZE,
      format = "png",
    } = slideMeta;

    if (!slideWidth || !slideHeight) {
      logger.error("Missing slide dimensions in metadata");
      throw new ConfigurationError("Missing slide dimensions in metadata");
    }

    // Create slide image source
    logger.debug(
      "Creating slide source with dimensions:",
      slideWidth,
      slideHeight
    );

    const wsiSource = new WholeSlideImage({
      width: slideWidth,
      height: slideHeight,
      tileSize: tileSize,
      crossOrigin: "anonymous",
      interpolate: DEFAULT_INTERPOLATE_SLIDE,
      transition: 0, // no fade transition between tile updates
      url: (z: number, x: number, y: number) =>
        `/api/v1/slides/${slideId}/tiles/${z}/${x}/${y}.${format}`, // template for tile URLs
      mpp: slideMpp,
    });
    return wsiSource;
  };

  /**
   * Setup the mask layer if available
   */
  const setupMaskLayer = async (
    slideId: string,
    slideMeta: SlideMetadata,
    slideSource: WholeSlideImage,
    tileSize: number,
    mapInstance: Map | null = null
  ): Promise<TileLayer | null> => {
    // Store the slideId locally to ensure it doesn't change during async operations
    const currentSlideId = slideId;

    logger.debug(`Starting mask layer setup for slideId: ${currentSlideId}`);
    try {
      // Fetch mask metadata using the authenticated fetch utility
      logger.debug(
        `Fetching masks from: /api/v1/slides/${currentSlideId}/annotations/raster`
      );

      try {
        const maskList = await apiFetch<MaskList>(
          `/api/v1/slides/${currentSlideId}/annotations/raster`
        );

        logger.debug("Masks response received:", maskList);

        // Check if there are any masks available
        if (!maskList.masks || maskList.masks.length === 0) {
          logger.info("No masks available for this slide");
          return null;
        }

        // Log if there are multiple masks
        if (maskList.masks.length > 1) {
          logger.info(
            `Multiple masks found (${maskList.masks.length}). Using the first mask: ${maskList.masks[0].maskName}`
          );
        }

        // Get the first mask
        const maskMeta: Mask = maskList.masks[0];

        // Check if mask dimensions are valid
        if (!maskMeta.maskWidth || !maskMeta.maskHeight) {
          logger.error(
            "Invalid mask dimensions:",
            maskMeta.maskWidth,
            maskMeta.maskHeight
          );
          return null;
        }

        // TODO: This can go away
        const maskResolutionPixelsPerMeter = maskMeta.maskMpp
          ? 1e6 / maskMeta.maskMpp
          : 1e6;

        const maskSource = new WholeSlideImage({
          width: maskMeta.maskWidth,
          height: maskMeta.maskHeight,
          mpp: maskMeta.maskMpp,
          tileSize: tileSize,
          crossOrigin: "anonymous",
          interpolate: false, // preserve pixelated rendering (no interpolation)
          transition: 0, // no fade transition between tile updates
          // Provide a custom URL template or function for tile loading:
          url: (z: number, x: number, y: number) => {
            // Check if the mask has a tilesUrl property and it's not empty
            if (maskMeta.tilesUrl && maskMeta.tilesUrl.trim() !== "") {
              // Use the tilesUrl from the mask metadata with format replacement
              return maskMeta.tilesUrl
                .replace("{z}", z.toString())
                .replace("{x}", x.toString())
                .replace("{y}", y.toString());
            } else {
              // Fall back to the default URL pattern if tilesUrl is not provided
              return `/api/v1/slides/${slideId}/annotations/raster/${maskMeta.maskId}/tiles/${z}/${x}/${y}.png`;
            }
          },
          pixelSize: maskResolutionPixelsPerMeter,
        });

        // Create the mask layer
        const maskLayer = new TileLayer({
          preload: DEFAULT_PRELOAD_LEVEL,
          extent: slideSource.getTileGrid().getExtent(),
          source: maskSource as any,
          visible: true,
          opacity: DEFAULT_MASK_OPACITY,
        });

        // Use the mapInstance parameter first, fall back to state if not provided
        const targetMap = mapInstance || map;

        // Add the mask layer to the map
        if (targetMap) {
          targetMap.addLayer(maskLayer);
          // Force map re-render
          targetMap.renderSync();
        } else {
          logger.error("Cannot add mask layer: map is null");
          return null;
        }
        return maskLayer;
      } catch (error) {
        logger.error("Error loading mask metadata:", error);
        return null;
      }
    } catch (error) {
      logger.error("Error loading mask layer:", error);
      return null;
    }
  };

  return (
    <div ref={mapContainerRef} className="w-full h-full">
      {map && slideSource && (
        <>
          <OverviewMapControl map={map} source={slideSource} collapsed={true} />
          <MapControls map={map} showScaleLine={true} />
        </>
      )}
    </div>
  );
}
