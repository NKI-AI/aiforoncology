# JPEG XL Multiplex WASM Example

This directory contains a WebAssembly (WASM) implementation of the JPEG XL multiplex encoding/decoding example, providing the same decoding functionality as the C++ and Go versions but running in web browsers.

## Files

- `libjxl_wrapper.cpp` - C++ WASM wrapper that exposes JPEG XL functions to JavaScript
- `BUILD.bazel` - Bazel build configuration for compiling to WASM

## Building the WASM Module

To build the WASM module using Bazel:

```bash
# From the project root directory
bazelisk build //aifo/libjxl_js:wasm --platforms=//platforms:wasm32
```

This will generate:

- `bazel-bin/aifo/libjxl_js/libjxl_multiplex.js` - JavaScript loader
- `bazel-bin/aifo/libjxl_js/libjxl_multiplex.wasm` - WebAssembly binary
