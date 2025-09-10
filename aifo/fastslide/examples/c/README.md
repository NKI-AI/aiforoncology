# FastSlide C Examples

This directory contains C examples demonstrating how to use the FastSlide C wrapper.

## Building

Build the examples using Bazel:

```bash
# Build the spectral example
bazel build //aifo/fastslide/examples/c:spectral_example_c
```

## Running the Spectral Example

The spectral example demonstrates advanced multiplex/spectral slide processing using the C API:

```bash
# Run with a spectral slide file
bazel run //aifo/fastslide/examples/c:spectral_example_c -- \
  --file /path/to/your/spectral_slide.qptiff \
  --output /tmp/fastslide_c_output
```

### Features Demonstrated

The C spectral example replicates the functionality of the C++ version and demonstrates:

1. **Channel Metadata Access**: Retrieves channel names, biomarkers, and colors
2. **Histogram Computation**: Creates histograms for each channel using the highest resolution level
3. **Display Range Extraction**: Computes QuPath-compatible display ranges using saturation-based algorithm
4. **Spectral Blending**: Converts spectral data to RGB using channel colors and display ranges
5. **Associated Images**: Saves thumbnail and other associated images

### Output

The example produces the same output as the C++ version:

- Channel information with biomarkers and display ranges printed to console
- Main slide level saved as PNG image with spectral blending applied
- Associated images (thumbnails, etc.) saved as separate PNG files

### Image Format

Images are saved in PNG format using libvips, providing high-quality output with good compression. This matches the format used by the C++ spectral example.

## API Comparison

The C example closely mirrors the C++ spectral example functionality:

| C++ Feature                                  | C Equivalent                                  |
| -------------------------------------------- | --------------------------------------------- |
| `SlideReader::GetChannelMetadata()`          | `fastslide_get_channel_metadata()`            |
| `CreateHistogramsFromImageChannels()`        | `fastslide_create_histograms_from_channels()` |
| `Histogram::ComputeDisplayRange()`           | `fastslide_histogram_compute_display_range()` |
| `CombineSpectralChannelsWithDisplayRanges()` | `fastslide_read_region_spectral_blend()`      |
| `SaveAsPNG()` (libvips)                      | `save_rgb_as_png()` (libvips-c)               |

## Dependencies

The C example uses:

- **FastSlide C wrapper**: For slide reading and processing
- **libvips-c**: For high-quality PNG image saving

## Error Handling

The C wrapper provides comprehensive error handling:

```c
// Check for errors after any operation
if (!fastslide_read_region_spectral_blend(...)) {
    printf("Error: %s\n", fastslide_get_last_error());
}
```

## Memory Management

The C wrapper handles memory management for complex objects, but you must free allocated arrays:

```c
// Free channel metadata
fastslide_free_channel_metadata(metadata, count);

// Free histograms
fastslide_free_histograms(histograms, count);

// Free associated image names
fastslide_free_associated_image_names(names, count);
```

## libvips Integration

The example properly initializes and cleans up libvips:

```c
// Initialize libvips at startup
if (vips_init(argv[0]) != 0) {
    printf("Failed to initialize libvips\n");
    return 1;
}

// ... process images ...

// Cleanup libvips before exit
vips_shutdown();
```

## Spectral Resampling Example

The `spectral_resample_c` example demonstrates high-quality image resampling using Lanczos algorithms on spectral images:

### Features

- **Multiple Lanczos Algorithms**: Tests Lanczos2, Lanczos3, and Cosine-windowed sinc resampling
- **Performance Timing**: Benchmarks resampling operations with configurable timing runs
- **Spectral Channel Processing**: Properly handles spectral images with display range computation
- **Planar Configuration Handling**: Automatically converts images to the required separate planar format

### Usage

```bash
# Run with a spectral slide file
bazelisk run //aifo/fastslide/examples/c:spectral_resample_c -- --file slide.qptiff

# Customize output directory and timing
bazelisk run //aifo/fastslide/examples/c:spectral_resample_c -- \
    --file slide.qptiff \
    --output output_resample \
    --runs 50 \
    --downsample 2.5
```

### Arguments

- `--file <path>`: Input slide file (required)
- `--output <dir>`: Output directory for resampled images (default: "output_resample")
- `--runs <num>`: Number of timing iterations for performance measurement (default: 20)
- `--downsample <factor>`: Downsample factor for resampling (default: 2.0)

### Output Files

The example generates several output files:

- `original_spectral.png`: Original image with proper spectral blending
- `resampled_spectral_WxH.png`: Resampled image with spectral blending
- `test_lanczos3_WxH.png`: Lanczos3 algorithm test result
- `test_lanczos2_WxH.png`: Lanczos2 algorithm test result
- `test_cosine_WxH.png`: Cosine-windowed sinc test result

### Example Output

```
FastSlide Spectral Resampling Example - C API
Slide file: example.qptiff
Output directory: output_resample
Timing runs: 20
Downsample factor: 2.0x

=== Slide Information ===
Format: QuPath TIFF
Level count: 5
Base dimensions: 46000x32914
MPP: 0.220 x 0.220
Objective magnification: 40.0x

Computing spectral channel display ranges...
  Found 4 spectral channels
  Computing histograms from level 4 (719x516)
  Computed display ranges for 4 channels

Processing spectral image for resampling demonstration...
  Using level 1 (2875x2057)
  Processing spectral image with 4 channels
  Converting original using spectral blending... ✓
  Saved original: original_spectral.png
  Resampling spectral data (2875x2057 → 1437x1028, downsample 2.0x)
    Converting to separate planar configuration... ✓
  Running 20 timing iterations..... ✓ (avg: 234 ms over 20 runs)
  Converting resampled spectral to RGB... ✓
  Saved resampled: resampled_spectral_1437x1028.png

Testing different resampling algorithms...
  Using test region: 719x516 from level 4
    Converting to separate planar configuration... ✓
  Testing Lanczos3 resampling... ✓ (15 ms)
  Testing Lanczos2 resampling... ✓ (12 ms)
  Testing Cosine-windowed sinc resampling... ✓ (18 ms)
  Algorithm comparison complete

✓ All resampled images saved to output_resample
✓ Spectral resampling demonstration complete
```

The resampling functions require images in separate planar configuration (`FASTSLIDE_PLANAR_CONFIG_SEPARATE`) and the example automatically handles this conversion when needed.
