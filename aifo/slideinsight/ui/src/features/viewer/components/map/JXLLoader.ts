// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import jsUrl from "./libjxl/libjxl_multiplex.js?url";
import wasmUrl from "./libjxl/libjxl_multiplex.wasm?url";

interface JXLData {
  data: Uint16Array | Uint8Array; // Raw data - no normalization
  dataType: string; // Data type string for caching
  isMultiplex: boolean; // Whether this is multiplex fluorescent data
  channels: number; // Number of channels
  width: number; // Image width
  height: number; // Image height
}

interface JXLModule {
  _malloc: (size: number) => number;
  _free: (ptr: number) => void;
  _encode_jxl_multiplex: (
    inputPtr: number,
    xsize: number,
    ysize: number,
    channels: number,
    quality: number
  ) => number;
  _decode_jxl_multiplex: (
    compressedPtr: number,
    compressedSize: number
  ) => number;
  _decode_jxl_rgba: (compressedPtr: number, compressedSize: number) => number;
  _get_compressed_size: () => number;
  _get_compressed_data: () => number;
  _get_decoded_size: () => number;
  _get_decoded_data: () => number;
  _get_decoded_rgba_data: () => number;
  _get_decoded_rgba_size: () => number;
  _get_rgba_bits_per_sample: () => number;
  HEAPU8: Uint8Array;
}

let jxlModule: JXLModule | null = null;
let moduleLoadPromise: Promise<JXLModule> | null = null;

// Loading state tracking
let isLoading = false;
const loadingCallbacks = new Set<(loading: boolean) => void>();

// Export functions for external components to track loading state
export const onJXLLoadingStateChange = (
  callback: (loading: boolean) => void
) => {
  loadingCallbacks.add(callback);
  // Immediately call with current state
  callback(isLoading);

  // Return cleanup function
  return () => {
    loadingCallbacks.delete(callback);
  };
};

const setLoadingState = (loading: boolean) => {
  if (isLoading !== loading) {
    isLoading = loading;
    loadingCallbacks.forEach((callback) => callback(loading));
  }
};

class JXLLoader {
  /**
   * Load and initialize the JPEG XL WASM module using dynamic import
   */
  private static async loadModule(): Promise<JXLModule> {
    if (jxlModule) {
      return jxlModule;
    }

    if (moduleLoadPromise) {
      return moduleLoadPromise;
    }

    moduleLoadPromise = (async () => {
      try {
        setLoadingState(true);
        console.log("🔧 Loading JPEG XL WASM module...");

        // Use dynamic import to load the ES6 module with Vite URL import
        const moduleFactory = await import(/* @vite-ignore */ jsUrl);

        // Initialize the module with WASM file location
        const module = await moduleFactory.default({
          locateFile: (path: string) => {
            if (path.endsWith(".wasm")) {
              return wasmUrl;
            }
            return path;
          },
        });

        jxlModule = module;
        console.log("✅ JPEG XL WASM module loaded successfully");
        setLoadingState(false);
        return module;
      } catch (error) {
        setLoadingState(false);
        console.error("❌ Failed to load JPEG XL WASM module:", error);
        throw new Error(
          `Failed to load JPEG XL WASM module: ${
            error instanceof Error ? error.message : error
          }`
        );
      }
    })();

    return moduleLoadPromise;
  }

  /**
   * Decode a JXL file from ArrayBuffer
   * Returns raw data without normalization
   * @param buffer - The JXL file as ArrayBuffer
   * @param isMultiplex - Whether this is multiplex fluorescent data (default: true for backward compatibility)
   */
  static async fromArrayBuffer(
    buffer: ArrayBuffer,
    isMultiplex: boolean = true
  ): Promise<JXLData> {
    // const startTime = performance.now();
    const Module = await this.loadModule();

    try {
      // Convert ArrayBuffer to Uint8Array and allocate WASM memory
      const array = new Uint8Array(buffer);

      const compressedPtr = Module._malloc(array.length);

      try {
        // Copy compressed data to WASM memory
        const compressedHeap = new Uint8Array(
          Module.HEAPU8.buffer,
          compressedPtr,
          array.length
        );
        compressedHeap.set(array);

        let resultPtr: number;
        let width: number;
        let height: number;
        let numComponents: number;
        let data: Uint16Array | Uint8Array;
        let dataType: string;

        if (isMultiplex) {
          // Decode multiplex fluorescent data
          resultPtr = Module._decode_jxl_multiplex(compressedPtr, array.length);

          if (!resultPtr) {
            throw new Error("JXL multiplex decoding failed");
          }

          // Parse the result structure
          // The decode function returns a pointer to a struct with:
          // [decodedDataPtr, xsize, ysize, channels, success]
          const resultView = new Uint32Array(
            Module.HEAPU8.buffer,
            resultPtr,
            5
          );
          const decodedDataPtr = resultView[0];
          width = resultView[1];
          height = resultView[2];
          numComponents = resultView[3];
          const success = resultView[4];

          if (!success) {
            throw new Error("JXL multiplex decode operation failed");
          }

          // Get the decoded data size and copy it
          const decodedSize = Module._get_decoded_size();
          const decodedDataHeap = new Uint16Array(
            Module.HEAPU8.buffer,
            Module._get_decoded_data(),
            decodedSize
          );

          // Copy the data to a new array to avoid memory issues
          data = new Uint16Array(decodedDataHeap);
          dataType = "uint16";
        } else {
          // Decode regular RGBA data
          resultPtr = Module._decode_jxl_rgba(compressedPtr, array.length);

          if (!resultPtr) {
            throw new Error("JXL RGBA decoding failed");
          }

          // Parse the result structure
          const resultView = new Uint32Array(
            Module.HEAPU8.buffer,
            resultPtr,
            5
          );
          const decodedDataPtr = resultView[0];
          width = resultView[1];
          height = resultView[2];
          numComponents = resultView[3];
          const success = resultView[4];

          if (!success) {
            throw new Error("JXL RGBA decode operation failed");
          }

          // Get bits per sample to determine data type
          const bitsPerSample = Module._get_rgba_bits_per_sample();
          const decodedSize = Module._get_decoded_rgba_size();

          if (bitsPerSample <= 8) {
            const decodedDataHeap = new Uint8Array(
              Module.HEAPU8.buffer,
              Module._get_decoded_rgba_data(),
              decodedSize
            );
            data = new Uint8Array(decodedDataHeap);
            dataType = "uint8";
          } else {
            // For 16-bit data, we need to reinterpret the bytes as uint16
            const decodedDataHeap = new Uint16Array(
              Module.HEAPU8.buffer,
              Module._get_decoded_rgba_data(),
              decodedSize / 2
            );
            data = new Uint16Array(decodedDataHeap);
            dataType = "uint16";
          }
        }

        // const loadTime = performance.now() - startTime;
        // const modeText = isMultiplex ? 'multiplex' : 'RGBA';
        // console.log(`✅ JXL ${modeText} decoded in ${loadTime.toFixed(1)}ms: ${width}x${height}x${numComponents} (${dataType} data)`);

        return {
          data,
          dataType,
          isMultiplex,
          channels: numComponents,
          width,
          height,
        };
      } finally {
        Module._free(compressedPtr);
      }
    } catch (error) {
      console.error("❌ Failed to decode JXL data:", error);
      throw new Error(
        `JXL decoding failed: ${error instanceof Error ? error.message : error}`
      );
    }
  }
}

// Export as default
export default JXLLoader;

// Make available globally for debugging
if (typeof window !== "undefined") {
  (window as any).JXLLoader = JXLLoader;
}
