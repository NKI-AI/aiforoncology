import ImageTileSource from "ol/source/ImageTile.js";
import DataTileSource from "ol/source/DataTile.js";
import {
  registerTileSource,
  type TileSourceFactory,
} from "./TileSourceRegistry";
import { registerJxlParser, getJxlParser } from "./JxlParsers";
import JXLLoader from "./JXLLoader";
import { createBlankTile, loadImage, normalizeData } from "./TileHelpers";

// Global tile error tracking
interface TileErrorInfo {
  errorCount: number;
  lastError: Date;
  firstError: Date;
  stopped: boolean;
}

const tileErrorTracker = new Map<string, TileErrorInfo>();
const MAX_TILE_ERRORS = 10; // Stop loading after this many errors
const ERROR_RESET_TIME = 30000; // Reset error count after 30 seconds

// Global error callback for persistent tile failures
let onPersistentTileError:
  | ((slideUid: string, errorCount: number) => void)
  | null = null;

export function setTileErrorCallback(
  callback: (slideUid: string, errorCount: number) => void
) {
  onPersistentTileError = callback;
}

export function trackTileError(slideUid: string, error: Error): boolean {
  const now = new Date();
  const key = slideUid;

  let errorInfo = tileErrorTracker.get(key);
  if (!errorInfo) {
    errorInfo = {
      errorCount: 0,
      lastError: now,
      firstError: now,
      stopped: false,
    };
    tileErrorTracker.set(key, errorInfo);
  }

  // Reset error count if enough time has passed since last error
  if (now.getTime() - errorInfo.lastError.getTime() > ERROR_RESET_TIME) {
    errorInfo.errorCount = 0;
    errorInfo.firstError = now;
    errorInfo.stopped = false;
  }

  errorInfo.errorCount++;
  errorInfo.lastError = now;

  // Check if we should stop loading tiles
  if (errorInfo.errorCount >= MAX_TILE_ERRORS && !errorInfo.stopped) {
    errorInfo.stopped = true;
    console.info(
      `⏸️ Pausing tile loading for slide ${slideUid} after ${errorInfo.errorCount} attempts (will show user dialog)`
    );

    // Notify about persistent error
    if (onPersistentTileError) {
      onPersistentTileError(slideUid, errorInfo.errorCount);
    }

    return true; // Should stop loading
  }

  return errorInfo.stopped;
}

export function shouldStopLoading(slideUid: string): boolean {
  const errorInfo = tileErrorTracker.get(slideUid);
  return errorInfo?.stopped ?? false;
}

export function resetTileErrors(slideUid: string) {
  tileErrorTracker.delete(slideUid);
}

// Default JXL parser wired to the bundled WASM loader
registerJxlParser("wasm-jxl", async (buffer, isMultiplex) => {
  return JXLLoader.fromArrayBuffer(buffer, isMultiplex);
});

const rasterFactory: TileSourceFactory = (options, base) =>
  new ImageTileSource({
    projection: base.getProjection(),
    tileGrid: base.getTileGrid(),
    crossOrigin: base.crossOrigin as
      | "anonymous"
      | "use-credentials"
      | undefined,
    interpolate: options.interpolate !== undefined ? options.interpolate : true,
    attributions: options.attributions,
    transition: options.transition,
    loader: async (z, x, y, loaderOptions: { signal?: AbortSignal }) => {
      let src: string | undefined;
      if (typeof base.tileUrlFunction === "function") {
        src = base.tileUrlFunction(z, x, y);
      }

      // Extract slideUid from URL for error tracking
      const slideUid = src ? extractSlideUidFromUrl(src) : "unknown";

      // Check if we should stop loading due to persistent errors
      if (shouldStopLoading(slideUid)) {
        console.debug(
          `⏸️ Skipping raster tile load for slide ${slideUid} (error recovery mode)`
        );
        return createBlankTile(base.tilePixelSize);
      }

      if (!src) {
        return createBlankTile(base.tilePixelSize);
      }
      try {
        if (loaderOptions?.signal?.aborted) {
          throw new DOMException("Request was aborted", "AbortError");
        }
        const img = await loadImage(
          src,
          base.crossOrigin,
          loaderOptions?.signal
        );
        return img;
      } catch (error) {
        if (error instanceof DOMException && error.name === "AbortError") {
          throw error;
        }

        // Track the error for raster tiles too
        trackTileError(slideUid, error as Error);

        console.warn(
          `🔄 Raster tile loading issue (z=${z}, x=${x}, y=${y}) for slide ${slideUid}:`,
          error
        );

        return createBlankTile(base.tilePixelSize);
      }
    },
  });

const jxlFactory: TileSourceFactory = (options, base) => {
  const bandCount = options.bandCount ?? 32;
  const parse = getJxlParser("wasm-jxl");
  return new DataTileSource({
    projection: base.getProjection(),
    tileGrid: base.getTileGrid(),
    bandCount,
    attributions: options.attributions,
    transition: options.transition,
    loader: async (z, x, y, loaderOptions: { signal?: AbortSignal }) => {
      let src: string | undefined;
      if (typeof base.tileUrlFunction === "function") {
        src = base.tileUrlFunction(z, x, y);
      }
      const tileSize = Array.isArray(base.tilePixelSize)
        ? base.tilePixelSize[0]
        : (base.tilePixelSize as number);

      // Extract slideUid from URL for error tracking
      const slideUid = src ? extractSlideUidFromUrl(src) : "unknown";

      // Check if we should stop loading due to persistent errors
      if (shouldStopLoading(slideUid)) {
        console.debug(
          `⏸️ Skipping JXL tile load for slide ${slideUid} (error recovery mode)`
        );
        return new Float32Array(tileSize * tileSize * 32);
      }

      if (!src) {
        return new Float32Array(tileSize * tileSize * 32);
      }
      try {
        if (loaderOptions?.signal?.aborted) {
          throw new DOMException("Request was aborted", "AbortError");
        }
        const response = await fetch(src, { signal: loaderOptions?.signal });
        if (!response.ok) {
          throw new Error(`HTTP ${response.status}`);
        }
        const buffer = await response.arrayBuffer();
        if (loaderOptions?.signal?.aborted) {
          throw new DOMException("Request was aborted", "AbortError");
        }
        const decoded = await parse(buffer, base.isFluorescent);
        return normalizeData(decoded.data, decoded.data.length);
      } catch (error) {
        if (error instanceof DOMException && error.name === "AbortError") {
          throw error;
        }

        // Track the error and check if we should stop loading
        const shouldStop = trackTileError(slideUid, error as Error);

        console.warn(
          `🔄 JXL tile loading issue (z=${z}, x=${x}, y=${y}) for slide ${slideUid}:`,
          error
        );

        return new Float32Array(tileSize * tileSize * 32);
      }
    },
  });
};

// Helper function to extract slideUid from tile URL
export function extractSlideUidFromUrl(url: string): string {
  try {
    // URL format: /api/v1/slides/{slideUid}/tiles/{z}/{x}/{y}.{format}
    const match = url.match(/\/api\/v1\/slides\/([^\/]+)\/tiles/);
    return match ? match[1] : "unknown";
  } catch {
    return "unknown";
  }
}

// Register defaults
registerTileSource("png", rasterFactory);
registerTileSource("jpg", rasterFactory);
registerTileSource("jxl", jxlFactory);
