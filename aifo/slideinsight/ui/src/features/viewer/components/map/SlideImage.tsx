// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import ImageTileSource from "ol/source/ImageTile.js";
import DataTileSource from "ol/source/DataTile.js";
import TileGrid from "ol/tilegrid/TileGrid.js";
import Projection from "ol/proj/Projection.js";
import type { Size } from "ol/size.js";
import type { Extent } from "ol/extent.js";
import JXLLoader from "./JXLLoader.js";
import {
  trackTileError,
  shouldStopLoading,
  extractSlideUidFromUrl,
} from "./defaultRegistrations.js";

// Supported tile formats
export type TileFormat = "png" | "jpg" | "jxl";

// Channel metadata for fluorescent imaging
interface ChannelMetadata {
  [channelIndex: string]: {
    name: string;
    biomarker: string;
    color: string; // Hex color like "#0000ff"
  };
}

// Slide metadata containing channel information
interface SlideMetadata {
  channels?: ChannelMetadata;
}

// Options interface for WholeSlideImage source.
interface WholeSlideImageOptions {
  width: number;
  height: number;
  tileSize: number;
  mpp: number; // Physical size of a pixel in map units (e.g. meter per pixel).
  crossOrigin?: string; // Cross-origin setting for image loading.
  interpolate?: boolean; // Whether to interpolate (smooth) images when scaling.
  transition?: number; // Duration of opacity transition for new tiles (ms).
  attributions?: string | string[]; // Source attributions if any.
  url?: (z: number, x: number, y: number) => string; // URL template(s) or function for tile URLs.
  slideUid?: string; // Slide identifier to use in URL templates
  pixelSize?: number; // Alternative name for mpp to maintain backward compatibility
  tileFormat?: TileFormat; // Format of tiles to load (default: 'png')
  bandCount?: number; // Number of bands for fluorescence tiles (default: 32 for fluorescence)
  metadata?: SlideMetadata; // Slide metadata including channel information
  isFluorescent?: boolean; // Whether this is a fluorescent image (determines JXL decoding mode)
}

// Helper class for managing individual tile loading. In OpenLayers ≥10, custom tile
// classes are generally not required because the loader handles asynchronous loading.
// Here, SlideImageTile is a helper that encapsulates the loading logic for a single tile.
class SlideImageTile {
  private z: number;
  private x: number;
  private y: number;
  private source: WholeSlideImage;

  constructor(z: number, x: number, y: number, source: WholeSlideImage) {
    this.z = z;
    this.x = x;
    this.y = y;
    this.source = source;
  }

  // Load the tile image asynchronously. Returns a Promise that resolves to an HTMLImageElement,
  // HTMLCanvasElement, or ImageBitmap. On failure, it falls back to a blank canvas.
  async load(
    signal?: AbortSignal
  ): Promise<HTMLImageElement | HTMLCanvasElement | ImageBitmap> {
    // Compute the tile URL using the source's URL function (if any).
    let src: string | undefined;
    if (typeof this.source.tileUrlFunction === "function") {
      src = this.source.tileUrlFunction(this.z, this.x, this.y);
    }

    // Extract slideUid from URL for error tracking
    const slideUid = src ? extractSlideUidFromUrl(src) : "unknown";

    // Check if we should stop loading due to persistent errors
    if (shouldStopLoading(slideUid)) {
      console.debug(
        `⏸️ Skipping raster tile load for slide ${slideUid} (error recovery mode)`
      );
      return WholeSlideImage.createBlankTile(this.source.tilePixelSize);
    }

    if (!src) {
      // If no URL (tile outside bounds or undefined), return a blank tile.
      return WholeSlideImage.createBlankTile(this.source.tilePixelSize);
    }

    try {
      // Check if already aborted
      if (signal?.aborted) {
        throw new DOMException("Request was aborted", "AbortError");
      }

      // Load the image from the URL.
      const img = await WholeSlideImage.loadImage(
        src,
        this.source.crossOrigin,
        signal
      );
      return img;
    } catch (error) {
      // Re-throw abort errors to properly cancel
      if (error instanceof DOMException && error.name === "AbortError") {
        throw error;
      }

      // Track the error for persistent error handling
      trackTileError(slideUid, error as Error);

      console.warn(
        `🔄 Raster tile loading issue (z=${this.z}, x=${this.x}, y=${this.y}) for slide ${slideUid}:`,
        error
      );

      // If loading fails, return a blank canvas as fallback.
      return WholeSlideImage.createBlankTile(this.source.tilePixelSize);
    }
  }
}

// Base class for both image and data tile sources
abstract class BaseWholeSlideImage {
  public readonly tilePixelSize: Size;
  public readonly crossOrigin: string;
  public tileUrlFunction?: (z: number, x: number, y: number) => string;
  protected readonly tileGrid: TileGrid;
  public readonly tileFormat: TileFormat;
  public readonly metadata?: SlideMetadata;
  public readonly isFluorescent: boolean;

  constructor(options: WholeSlideImageOptions) {
    this.tileFormat = options.tileFormat || "png";
    this.metadata = options.metadata;
    this.isFluorescent = options.isFluorescent ?? false;

    // Prepare projection and tile grid based on image dimensions and resolution.
    const MicronsPerMeter = options.mpp ? 1e-6 * options.mpp : 1e-6;
    const heightMeters = options.height * MicronsPerMeter;
    const widthMeters = options.width * MicronsPerMeter;

    const imageExtent: Extent = [0, -heightMeters, widthMeters, 0];
    const projection = new Projection({
      code: "none",
      units: "m",
      extent: imageExtent,
    });

    const tilePixelSize = [options.tileSize, options.tileSize];

    // Calculate the zoom level for native resolution (1:1 pixel ratio)
    // This represents the maximum zoom level where actual tiles exist on the server
    const maxZoom = Math.ceil(
      Math.log2(Math.max(options.width, options.height) / options.tileSize)
    );

    // Generate resolutions array only for zoom levels that have actual tiles
    const resolutions = Array.from(
      { length: maxZoom + 1 },
      (_, z) => (1 << (maxZoom - z)) * MicronsPerMeter
    );

    this.tileGrid = new TileGrid({
      minZoom: 0,
      extent: imageExtent,
      origin: [imageExtent[0], imageExtent[3]],
      resolutions: resolutions,
      tileSize: [options.tileSize, options.tileSize],
    });

    this.crossOrigin =
      options.crossOrigin !== undefined ? options.crossOrigin : "anonymous";
    this.tilePixelSize = tilePixelSize;

    if (options.url) {
      if (typeof options.url === "function") {
        this.tileUrlFunction = options.url;
      } else {
        throw Error("URL should be a function for WholeSlideImage");
      }
    }
  }

  getTileGrid(): TileGrid {
    return this.tileGrid;
  }

  // Helper function to normalize typed arrays to Float32Array [0, 1] range
  static normalizeData(data: any, targetSize: number): Float32Array {
    const result = new Float32Array(targetSize);

    // Determine normalization factor based on data type
    let divisor = 1.0;
    if (data instanceof Uint8Array) {
      divisor = 255.0;
    } else if (data instanceof Uint16Array) {
      divisor = 65535.0;
    } else if (data instanceof Int8Array) {
      divisor = 127.0;
    } else if (data instanceof Int16Array) {
      divisor = 32767.0;
    } else if (data instanceof Float32Array || data instanceof Float64Array) {
      divisor = 1.0; // Already normalized
    } else {
      divisor = 255.0; // Default
    }

    // Normalize the data
    for (let i = 0; i < data.length; i++) {
      result[i] = data[i] / divisor;
    }

    return result;
  }
}

// Image-based tile source (for PNG/JPG tiles)
class WholeSlideImageSource extends ImageTileSource {
  private baseImage: BaseWholeSlideImage;

  constructor(options: WholeSlideImageOptions, baseImage: BaseWholeSlideImage) {
    const imageExtent = baseImage.getTileGrid().getExtent();
    const projection = new Projection({
      code: "none",
      units: "m",
      extent: imageExtent,
    });

    super({
      projection: projection,
      tileGrid: baseImage.getTileGrid(),
      crossOrigin: baseImage.crossOrigin as any,
      interpolate:
        options.interpolate !== undefined ? options.interpolate : true,
      attributions: options.attributions,
      transition: options.transition,
      loader: async (
        z: number,
        x: number,
        y: number,
        loaderOptions: { signal?: AbortSignal }
      ) => {
        const tile = new SlideImageTile(z, x, y, baseImage as any);
        return tile.load(loaderOptions.signal);
      },
    });

    this.baseImage = baseImage;
  }

  getTileGrid(): TileGrid {
    return this.baseImage.getTileGrid();
  }
}

// Data-based tile source (for JXL tiles)
class WholeSlideDataSource extends DataTileSource {
  private baseImage: BaseWholeSlideImage;
  private readonly configuredBandCount: number;

  constructor(options: WholeSlideImageOptions, baseImage: BaseWholeSlideImage) {
    const imageExtent = baseImage.getTileGrid().getExtent();
    const projection = new Projection({
      code: "none",
      units: "m",
      extent: imageExtent,
    });

    // For JXL files, use a higher default band count since they can have many components
    const bandCount =
      baseImage.tileFormat === "jxl"
        ? options.bandCount || 32 // Default to 32 for JXL files
        : options.bandCount || 3; // Default RGB for other formats

    super({
      projection: projection,
      tileGrid: baseImage.getTileGrid(),
      // Ensure WebGL resampling uses linear filtering for smoother visuals
      interpolate:
        options.interpolate !== undefined ? options.interpolate : true,
      bandCount: bandCount,
      loader: async (
        z: number,
        x: number,
        y: number,
        loaderOptions: { signal?: AbortSignal }
      ) => {
        return this.loadDataTile(z, x, y, loaderOptions.signal);
      },
      attributions: options.attributions,
      transition: options.transition,
    });

    this.baseImage = baseImage;
    this.configuredBandCount = bandCount;
  }

  private async loadDataTile(
    z: number,
    x: number,
    y: number,
    signal?: AbortSignal
  ): Promise<Float32Array> {
    const format = this.baseImage.tileFormat;
    if (format !== "jxl") {
      throw new Error("Only JXL tiles are supported");
    }

    let src: string | undefined;
    if (typeof this.baseImage.tileUrlFunction === "function") {
      src = this.baseImage.tileUrlFunction(z, x, y);
    }

    // Extract slideUid from URL for error tracking
    const slideUid = src ? extractSlideUidFromUrl(src) : "unknown";

    // Check if we should stop loading due to persistent errors
    if (shouldStopLoading(slideUid)) {
      console.debug(
        `⏸️ Skipping JXL data tile load for slide ${slideUid} (error recovery mode)`
      );
      const tileSize = this.baseImage.tilePixelSize[0];
      return new Float32Array(tileSize * tileSize * 32);
    }

    if (!src) {
      // Return blank tile data
      const tileSize = this.baseImage.tilePixelSize[0]; // Assuming square tiles
      const blankTile = new Float32Array(tileSize * tileSize * 32); // Use 32 as default
      return blankTile;
    }

    try {
      // Check if already aborted before starting fetch
      if (signal?.aborted) {
        throw new DOMException("Request was aborted", "AbortError");
      }

      const response = await fetch(src, { signal });
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
      }

      const buffer = await response.arrayBuffer();

      // Check if aborted after download but before expensive JXL decoding
      if (signal?.aborted) {
        throw new DOMException("Request was aborted", "AbortError");
      }

      // Use the authoritative isFluorescent flag to determine decoding mode
      const isMultiplex = this.baseImage.isFluorescent;

      let tileData: {
        data: Uint16Array | Uint8Array;
        dataType: string;
        isMultiplex: boolean;
        channels: number;
        width: number;
        height: number;
      };

      // TODO: Future enhancement - pass signal to JXLLoader for WASM-level cancellation
      // Currently WASM decoding cannot be easily cancelled, but we've prevented starting
      // unnecessary decoding operations when already aborted
      tileData = await JXLLoader.fromArrayBuffer(buffer, isMultiplex);

      const { data, dataType } = tileData;

      // Note: Multiplex data is now already in INTERLEAVED format from WASM decoder
      // No conversion needed - WebGL can use the data directly

      const dataTile = BaseWholeSlideImage.normalizeData(data, data.length);
      return dataTile;
    } catch (error) {
      // Don't log abort errors as failures - they're expected when navigating away
      if (error instanceof DOMException && error.name === "AbortError") {
        throw error; // Re-throw abort errors to properly cancel the operation
      }

      // Track the error for persistent error handling
      trackTileError(slideUid, error as Error);

      console.warn(
        `🔄 ${format.toUpperCase()} tile loading issue (z=${z}, x=${x}, y=${y}) for slide ${slideUid}:`,
        error
      );
      // Return blank tile data on error
      const tileSize = this.baseImage.tilePixelSize[0]; // Assuming square tiles
      const blankTile = new Float32Array(tileSize * tileSize * 32); // Use 32 as default
      return blankTile;
    }
  }

  getTileGrid(): TileGrid {
    return this.baseImage.getTileGrid();
  }
}

// Main WholeSlideImage class that delegates to either image or data source
export class WholeSlideImage extends BaseWholeSlideImage {
  private source: WholeSlideImageSource | WholeSlideDataSource;
  private actualTileFormat: TileFormat; // Track the actual format being used

  constructor(options: WholeSlideImageOptions) {
    super(options);

    // Create appropriate source based on tile format
    if (this.tileFormat === "jxl") {
      this.actualTileFormat = this.tileFormat;
      this.source = new WholeSlideDataSource(options, this);
    } else {
      this.actualTileFormat = this.tileFormat;
      this.source = new WholeSlideImageSource(options, this);
    }
  }

  // Get the actual tile format being used (may differ from requested if fallback occurred)
  getActualTileFormat(): TileFormat {
    return this.actualTileFormat;
  }

  // Delegate methods to the underlying source
  getProjection() {
    return this.source.getProjection();
  }

  on(type: any, listener: any) {
    return this.source.on(type, listener);
  }

  once(type: any, listener: any) {
    return this.source.once(type, listener);
  }

  un(type: any, listener: any) {
    return this.source.un(type, listener);
  }

  dispatchEvent(event: any) {
    return this.source.dispatchEvent(event);
  }

  getRevision() {
    return this.source.getRevision();
  }

  // Expose the underlying source for use with layers
  getSource() {
    return this.source;
  }

  /**
   * Static helper: Load an image from a URL, returning a Promise that resolves when the image is fully loaded.
   * If the image fails to load (error), the promise will be rejected.
   * The `crossOrigin` parameter is applied to the image element if provided.
   * The `signal` parameter allows cancellation of the image loading.
   */
  static loadImage(
    src: string,
    crossOrigin?: string,
    signal?: AbortSignal
  ): Promise<HTMLImageElement> {
    return new Promise((resolve, reject) => {
      // Check if already aborted
      if (signal?.aborted) {
        reject(new DOMException("Request was aborted", "AbortError"));
        return;
      }

      const img = new Image();
      if (crossOrigin !== undefined) {
        img.crossOrigin = crossOrigin;
      }

      // Set up abort handling
      let abortHandler: (() => void) | undefined;
      if (signal) {
        abortHandler = () => {
          // Cancel the image loading by setting src to empty
          img.src = "";
          reject(new DOMException("Request was aborted", "AbortError"));
        };
        signal.addEventListener("abort", abortHandler);
      }

      img.onload = () => {
        // Clean up abort listener
        if (signal && abortHandler) {
          signal.removeEventListener("abort", abortHandler);
        }
        resolve(img);
      };

      img.onerror = () => {
        // Clean up abort listener
        if (signal && abortHandler) {
          signal.removeEventListener("abort", abortHandler);
        }
        reject(new Error(`Failed to load tile image: ${src}`));
      };

      img.src = src;
    });
  }

  /**
   * Static helper: Create a blank canvas of the given tile size. This is used as a fallback for tiles
   * that fail to load. The returned canvas is transparent.
   */
  static createBlankTile(tileSize: Size): HTMLCanvasElement {
    const canvas = document.createElement("canvas");
    canvas.width = Array.isArray(tileSize) ? tileSize[0] : tileSize;
    canvas.height = Array.isArray(tileSize) ? tileSize[1] : tileSize;
    // Fill the canvas with transparent pixels (optional, as canvas is transparent by default).
    const ctx = canvas.getContext("2d");
    if (ctx) {
      ctx.fillStyle = "rgba(0,0,0,0)";
      ctx.fillRect(0, 0, canvas.width, canvas.height);
    }
    return canvas;
  }
}
