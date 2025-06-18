// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
import ImageTileSource from "ol/source/ImageTile.js";
import TileGrid from "ol/tilegrid/TileGrid.js";
import Projection from "ol/proj/Projection.js";
import type { Size } from "ol/size.js";
import type { Extent } from "ol/extent.js";

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
  slideId?: string; // Slide identifier to use in URL templates
  pixelSize?: number; // Alternative name for mpp to maintain backward compatibility
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
  async load(): Promise<HTMLImageElement | HTMLCanvasElement | ImageBitmap> {
    // Compute the tile URL using the source's URL function (if any).
    let src: string | undefined;
    if (typeof this.source.tileUrlFunction === "function") {
      src = this.source.tileUrlFunction(this.z, this.x, this.y);
    }
    if (!src) {
      // If no URL (tile outside bounds or undefined), return a blank tile.
      return WholeSlideImage.createBlankTile(this.source.tilePixelSize);
    }
    try {
      // Load the image from the URL.
      const img = await WholeSlideImage.loadImage(src, this.source.crossOrigin);
      return img;
    } catch (error) {
      // If loading fails, return a blank canvas as fallback.
      return WholeSlideImage.createBlankTile(this.source.tilePixelSize);
    }
  }
}

// The WholeSlideImage source class extends the new OpenLayers ImageTile source.
export class WholeSlideImage extends ImageTileSource {
  // Store properties needed for loader logic and projection.
  public readonly tilePixelSize: Size;
  public readonly crossOrigin: string;
  public tileUrlFunction?: (z: number, x: number, y: number) => string;
  protected readonly tileGrid: TileGrid;

  constructor(options: WholeSlideImageOptions) {
    // Prepare projection and tile grid based on image dimensions and resolution.
    // Determine the coordinate extent of the image. Default is the "fourth quadrant"
    // where the origin (0,0) is the top-left of the image and y increases upward.

    // Default 1 micron per pixel
    const MicronsPerMeter = options.mpp ? 1e-6 * options.mpp : 1e-6;

    const heightMeters = options.height * MicronsPerMeter;
    const widthMeters = options.width * MicronsPerMeter;

    const imageExtent: Extent = [0, -heightMeters, widthMeters, 0];
    const projection = new Projection({
      code: "none",
      units: "m",
      extent: imageExtent,
    });

    let tileGrid: TileGrid;
    const tilePixelSize = [options.tileSize, options.tileSize];
    const maxZoom = Math.ceil(
      Math.log2(Math.max(options.width, options.height) / options.tileSize)
    );
    const resolutions = Array.from(
      { length: maxZoom + 1 },
      (_, z) => 1 << (maxZoom - z)
    ).map((x) => x * MicronsPerMeter);

    tileGrid = new TileGrid({
      minZoom: 0,
      extent: imageExtent,
      origin: [imageExtent[0], imageExtent[3]], // top-left corner of the image (e.g., [0,0] in this coordinate system).
      resolutions: resolutions,
      tileSize: [options.tileSize, options.tileSize],
    });

    // Store properties for use in the loader.
    const crossOrigin =
      options.crossOrigin !== undefined ? options.crossOrigin : "anonymous";
    const interpolate =
      options.interpolate !== undefined ? options.interpolate : true;
    // Keep references for loader usage.
    const sourceTileSize: Size = tilePixelSize;
    const sourceCrossOrigin: string = crossOrigin;

    let tileUrlFunction:
      | ((z: number, x: number, y: number) => string)
      | undefined = undefined;
    if (options.url) {
      if (typeof options.url === "function") {
        tileUrlFunction = options.url;
      } else {
        throw Error("URL should be a function for WholeSlideImage");
      }
    }

    // Save the URL function for use in SlideImageTile (if needed).
    // (These are not part of the official API, but we keep them as internal references.)
    super({
      projection: projection,
      tileGrid: tileGrid,
      crossOrigin: crossOrigin as any, // Type coercion to match the expected CrossOriginAttribute
      interpolate: interpolate,
      attributions: options.attributions,
      transition: options.transition,
      // Define the loader function for this source. The loader handles asynchronous tile fetching.
      loader: async (
        z: number,
        x: number,
        y: number,
        loaderOptions: { signal?: AbortSignal }
      ) => {
        // Use SlideImageTile to handle the loading and fallback.
        const tile = new SlideImageTile(z, x, y, this);
        // Delegate to the SlideImageTile's load method (returns a Promise).
        // If the request is aborted (signal), we do not explicitly cancel the Image load,
        // but OpenLayers will ignore the result if the tile is no longer needed.
        if (loaderOptions.signal) {
          // If an abort signal is provided, listen for abort to possibly handle cancellation.
          // (Note: HTMLImageElement loading cannot be easily aborted, so we rely on GC if aborted.)
          loaderOptions.signal.addEventListener("abort", () => {
            // On abort, we could optionally stop loading logic if possible.
            // Here, we simply let it abort usage; the tile promise may resolve later but will be ignored by OL.
          });
        }
        return tile.load();
      },
    });

    // After calling super(), `this` is fully initialized as an ImageTile source.
    // Store the configuration for use in SlideImageTile:
    this.tilePixelSize = sourceTileSize;
    this.crossOrigin = sourceCrossOrigin;
    this.tileGrid = tileGrid; // Store the tileGrid for direct access

    if (tileUrlFunction) {
      // Normalize the URL function to (z,x,y) signature.
      this.tileUrlFunction = tileUrlFunction;
    }
  }

  /**
   * Gets the tile grid used by this source.
   * @returns The tile grid.
   */
  getTileGrid(): TileGrid {
    return this.tileGrid;
  }

  /**
   * Static helper: Load an image from a URL, returning a Promise that resolves when the image is fully loaded.
   * If the image fails to load (error), the promise will be rejected.
   * The `crossOrigin` parameter is applied to the image element if provided.
   */
  static loadImage(
    src: string,
    crossOrigin?: string
  ): Promise<HTMLImageElement> {
    return new Promise((resolve, reject) => {
      const img = new Image();
      if (crossOrigin !== undefined) {
        img.crossOrigin = crossOrigin;
      }
      img.onload = () => resolve(img);
      img.onerror = () =>
        reject(new Error(`Failed to load tile image: ${src}`));
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
