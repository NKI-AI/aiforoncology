# FIM (FastImage) Library

A minimal C++20 library for tile-based image processing using CRTP (Curiously Recurring Template Pattern) with Bazel build system.

## Features

- **Tile-based processing**: Efficient processing of large images through tiling
- **CRTP pipeline**: Zero-overhead pipeline composition using compile-time polymorphism
- **Chained API**: Fluent interface for easy pipeline construction
- **Multiple sources**: Support for TIFF (via libtiff) and PNG (via lodepng)
- **Operators**: Crop and downsample operations with extensible design
- **Flexible sinks**: Write to tiled TIFF files

## Dependencies

- C++20 compatible compiler
- Bazelisk (Bazel wrapper)
- libtiff-4 (for TIFF support)
- lodepng (included in third_party/)

## Building

```bash
# Build the library
bazelisk build //aifo/fimage:all

# Run tests
bazelisk test //aifo/fimage/...

# Build with optimizations
bazelisk build -c opt //aifo/fimage:all
```

## Usage

### Basic Pipeline

```cpp
#include <fim/image.h>
#include <fim/sources/tiff_source.h>
#include <fim/operators/crop.h>
#include <fim/operators/downsample.h>
#include <fim/sinks/tiff_sink.h>

int main() {
    // Create a chained pipeline
    fim::Image<fim::TiffSource>("input.tiff")
        .Crop(10, 10, 512, 512)
        .Downsample(2)
        .Render(fim::TiffSink("output.tiff"));

    return 0;
}
```

### Step-by-step Pipeline

```cpp
// Create image from TIFF source
auto image = fim::Image<fim::TiffSource>("input.tiff");

// Apply crop operation
auto cropped = image.Crop(10, 10, 512, 512);

// Apply downsample operation
auto downsampled = cropped.Downsample(2);

// Render to output
downsampled.Render(fim::TiffSink("output.tiff"));
```

### PNG Source Example

```cpp
fim::Image<fim::PngSource>("input.png")
    .Crop(0, 0, 256, 256)
    .Downsample(4)
    .Render(fim::TiffSink("output.tiff"));
```

### Getting Image Information

```cpp
auto image = fim::Image<fim::TiffSource>("input.tiff");
auto dims = image.GetSource().GetDimensions();
auto tile_size = image.GetSource().GetIdealTileSize();

std::cout << "Image: " << dims.width << "x" << dims.height
          << " channels: " << dims.channels << std::endl;
std::cout << "Tile size: " << tile_size.width << "x" << tile_size.height << std::endl;
```

## Architecture

### Sources

- **TiffSource**: Reads from TIFF files (tiled and strip-based)
- **PngSource**: Reads from PNG files using lodepng

### Operators

- **Crop**: Extracts a rectangular region from the image
- **Downsample**: Performs average-pooling downsampling by integer factor

### Sinks

- **TiffSink**: Writes to TIFF files (automatically chooses tiled vs strip format)

### CRTP Pipeline

All pipeline stages use CRTP for zero-overhead composition:

- `SourceBase<Derived>`: Base class for all sources
- `OperatorBase<Derived, InputType>`: Base class for all operators
- `SinkBase<Derived>`: Base class for all sinks

## Design Principles

1. **Zero-overhead abstraction**: CRTP ensures no virtual function calls
2. **Tile-based processing**: Efficient memory usage for large images
3. **Composable pipeline**: Operators can be chained in any order
4. **Type safety**: Compile-time type checking for pipeline stages
5. **Extensible**: Easy to add new sources, operators, and sinks

## Example Pipeline Flow

```
TiffSource -> Crop -> Downsample -> TiffSink
     ^         ^          ^           ^
     |         |          |           |
  GetTile() GetTile() GetTile()   Render()
```

Each stage requests tiles from the previous stage, transforming the data as needed.

## License

Apache 2.0 License - see LICENSE file for details.
