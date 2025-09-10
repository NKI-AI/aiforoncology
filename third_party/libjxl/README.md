# libjxl Vendoring Guide

This directory contains a vendored version of libjxl (JPEG XL reference implementation) following the repository's third-party library pattern.

## Current Version

- **Version**: 0.11.1
- **Release Date**: September 2024
- **Source**: https://github.com/libjxl/libjxl

## Available Targets

### Core Libraries

- `//third_party/libjxl:libjxl` - Main JPEG XL decoder/encoder library
- `//third_party/libjxl:libjxl_threads` - Threading support for JPEG XL
- `//third_party/libjxl:jpegli` - JPEG implementation based on JPEG XL technology
- `//third_party/libjxl:libjxl_base` - Base utilities and core functionality
- `//third_party/libjxl:libjxl_includes` - Public headers only

### Extended Libraries

- `//third_party/libjxl:libjxl_extras` - Additional codec support (APNG, EXR, GIF, JPEG, PNG)
- `//third_party/libjxl:libjxl_test_utils` - Test utilities (testonly)

## Directory Structure

```
third_party/libjxl/
├── BUILD.bazel           # Public API wrapper targets
├── version.bzl           # Version configuration
├── README.md            # This file
└── upstream/
    └── libjxl/          # Complete libjxl source code
        ├── MODULE.bazel  # libjxl's own Bazel module
        └── lib/BUILD     # Modified for vendored usage
```

## Dependencies

libjxl requires these dependencies, which are already included in the main MODULE.bazel:

- `brotli` - Compression library
- `skcms` - Skia Color Management System
- `hwy` - Google Highway (SIMD library)
- `giflib` - GIF format support (for extras)
- `openexr` - OpenEXR format support (for extras)
- `libpng` - PNG format support (for extras)
- `libjpeg_turbo` - JPEG format support (for extras)

## Changes Made for Vendoring

### 1. Include Path Fixes

The main issue was that libjxl's BUILD file expected to be at the repository root, but was nested under `third_party/libjxl/upstream/libjxl/lib/`.

**Fix**: Added `includes = [".."]` to all relevant cc_library targets in `upstream/libjxl/lib/BUILD`:

- `jpegxl`
- `base`
- `jpegli`
- `jpegxl_extras`
- `test_utils`

### 2. Dependency Mapping

- Renamed `@highway` to `@hwy` to match the main MODULE.bazel naming

## How to Update libjxl

### 1. Download New Version

```bash
cd third_party/libjxl
# Check latest release at: https://github.com/libjxl/libjxl/releases
wget https://github.com/libjxl/libjxl/archive/refs/tags/vX.Y.Z.tar.gz
tar -xzf vX.Y.Z.tar.gz
rm -rf upstream/libjxl
mv libjxl-X.Y.Z upstream/libjxl
```

### 2. Update Version Information

Edit `version.bzl`:

```python
LIBJXL_VERSION = "X.Y.Z"
LIBJXL_COMMIT = "commit_hash_from_github"  # Get from GitHub release
LIBJXL_SOURCE = "https://github.com/libjxl/libjxl/archive/refs/tags/vX.Y.Z.tar.gz"
```

### 3. Apply Include Path Fixes

Add `includes = [".."]` to these targets in `upstream/libjxl/lib/BUILD`:

```python
cc_library(
    name = "jpegxl",
    # ... existing config ...
    includes = [".."],  # Add this line
    # ... rest of config ...
)

cc_library(
    name = "base",
    # ... existing config ...
    includes = [".."],  # Add this line
    # ... rest of config ...
)

# Repeat for: jpegli, jpegxl_extras, test_utils
```

### 4. Check Dependencies

Compare `upstream/libjxl/MODULE.bazel` with the main `MODULE.bazel` to ensure all required dependencies are available and versions are compatible.

### 5. Test the Build

```bash
# Test core library
bazelisk build //third_party/libjxl:libjxl

# Test extras (requires OpenEXR, giflib, etc.)
bazelisk build //third_party/libjxl:libjxl_extras

# Test all targets
bazelisk build //third_party/libjxl:all
```

### 6. Update This README

Update the version information and any new dependencies or changes.

## Common Issues

### Missing Dependencies

If you get dependency errors, check that all bazel_dep entries from `upstream/libjxl/MODULE.bazel` are present in the main `MODULE.bazel`.

### Include Path Errors

If you see errors like `fatal error: 'lib/jxl/something.h' file not found`, you need to add `includes = [".."]` to the failing cc_library target.

### Version Compatibility

Ensure the versions of shared dependencies (like `hwy`, `brotli`) in the main MODULE.bazel are compatible with libjxl's requirements. The default libjxl bazel build files use `@highway` rather than `@hwy`, so that needs to be modified.

## Usage Example

### Public API (Recommended)

For external code using libjxl, use the clean public API:

```cpp
#include "jxl/decode.h"     // ✅ Correct - Clean public API
#include "jxl/encode.h"     // ✅ Correct - Clean public API

// In your BUILD.bazel:
cc_library(
    name = "my_library",
    srcs = ["my_code.cc"],
    deps = [
        "//third_party/libjxl:libjxl_includes", # Public headers only
        "//third_party/libjxl:libjxl",          # Core functionality
        "//third_party/libjxl:libjxl_extras",   # Additional codecs
    ],
)
```

### Internal Headers (Advanced)

If you need access to internal libjxl implementation details:

```cpp
#include "lib/jxl/dec_frame.h"        // ⚠️  Internal implementation
#include "lib/jxl/enc_butteraugli.h"  // ⚠️  Internal implementation

// Note: Internal headers may change between versions
// In your BUILD.bazel:
cc_library(
    name = "my_advanced_library",
    srcs = ["my_advanced_code.cc"],
    deps = [
        "//third_party/libjxl/upstream/libjxl/lib:jpegxl_private", # Internal access
    ],
)
```

### Include Path Explanation

libjxl has two types of headers:

1. **Public API**: `include/jxl/*.h` → Include as `#include "jxl/decode.h"`

   - Use `//third_party/libjxl:libjxl_includes` dependency
   - Stable interface, recommended for external code

2. **Internal Headers**: `lib/jxl/*.h` → Include as `#include "lib/jxl/some_header.h"`
   - Use `//third_party/libjxl/upstream/libjxl/lib:*` dependencies
   - Implementation details, may change between versions
