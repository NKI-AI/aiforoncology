import type { Size } from "ol/size.js";

export function loadImage(
  src: string,
  crossOrigin?: string,
  signal?: AbortSignal
): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    if (signal?.aborted) {
      reject(new DOMException("Request was aborted", "AbortError"));
      return;
    }
    const img = new Image();
    if (crossOrigin !== undefined) {
      img.crossOrigin = crossOrigin;
    }
    let abortHandler: (() => void) | undefined;
    if (signal) {
      abortHandler = () => {
        img.src = "";
        reject(new DOMException("Request was aborted", "AbortError"));
      };
      signal.addEventListener("abort", abortHandler);
    }
    img.onload = () => {
      if (signal && abortHandler) {
        signal.removeEventListener("abort", abortHandler);
      }
      resolve(img);
    };
    img.onerror = () => {
      if (signal && abortHandler) {
        signal.removeEventListener("abort", abortHandler);
      }
      reject(new Error(`Failed to load tile image: ${src}`));
    };
    img.src = src;
  });
}

export function createBlankTile(tileSize: Size): HTMLCanvasElement {
  const canvas = document.createElement("canvas");
  canvas.width = Array.isArray(tileSize) ? tileSize[0] : (tileSize as number);
  canvas.height = Array.isArray(tileSize) ? tileSize[1] : (tileSize as number);
  const ctx = canvas.getContext("2d");
  if (ctx) {
    ctx.fillStyle = "rgba(0,0,0,0)";
    ctx.fillRect(0, 0, canvas.width, canvas.height);
  }
  return canvas;
}

export function normalizeData(data: any, targetSize: number): Float32Array {
  const result = new Float32Array(targetSize);
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
    divisor = 1.0;
  } else {
    divisor = 255.0;
  }
  for (let i = 0; i < data.length; i++) {
    result[i] = data[i] / divisor;
  }
  return result;
}
