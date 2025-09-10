#!/usr/bin/env python3
"""
FastSlide Python Usage Examples

This example demonstrates how to use the FastSlide Python extension with the new
Google C++ style API. FastSlide provides high-performance, thread-safe reading
of digital pathology slides in SVS, QPTIFF, and other formats.

KEY DIFFERENCES FROM OPENSLIDE:
===============================

- OpenSlide: Always uses level-0 coordinates, automatically converts
- FastSlide: Uses level-native coordinates (more efficient, no conversion overhead)

See detailed comparisons in each example section below.
"""

from pathlib import Path
import argparse
import time
import threading
from typing import Optional, Tuple

# Import the FastSlide Python extension
import fastslide


def basic_usage_example(slide_path: str):
    """
    Basic usage example showing core functionality

    OPENSLIDE COMPARISON:
    OpenSlide: slide = openslide.OpenSlide(slide_path)
    FastSlide: slide = fastslide.FastSlide.from_file_path(slide_path)
    """
    print(f"=== Basic Usage: {Path(slide_path).name} ===")

    # Create slide reader using factory method (Google C++ style)
    # This replaces OpenSlide's direct constructor
    slide = fastslide.FastSlide.from_file_path(slide_path)

    # Get basic slide information using properties (not methods)
    # OpenSlide: slide.detect_format(slide_path)
    # FastSlide: slide.format (property)
    print(f"Format: {slide.format}")
    print(f"Levels: {slide.level_count}")

    # Get slide properties - combined properties + metadata
    # OpenSlide: slide.properties (dict of strings)
    # FastSlide: slide.properties (dict with typed values + metadata)
    props = slide.properties
    print(f"Resolution: {props['mpp_x']:.3f} x {props['mpp_y']:.3f} μm/pixel")
    print(f"Magnification: {props['objective_magnification']}x")
    print(f"Scanner: {props['scanner_model']}")
    print(f"Source file: {props['source_path']}")

    # Print pyramid information
    # OpenSlide: slide.level_dimensions, slide.level_downsamples
    # FastSlide: slide.level_dimensions, slide.level_downsamples (same)
    print("\nPyramid levels:")
    dimensions = slide.level_dimensions
    downsamples = slide.level_downsamples

    for level in range(slide.level_count):
        width, height = dimensions[level]
        downsample = downsamples[level]
        print(f"  Level {level}: {width}x{height} (downsample: {downsample:.1f}x)")

    # List associated images - lazy loading
    # OpenSlide: slide.associated_images.keys() (all loaded immediately)
    # FastSlide: slide.associated_images.keys() (not loaded until accessed)
    assoc_names = slide.associated_images.keys()
    print(f"\nAssociated images: {list(assoc_names)}")
    print(f"Images currently in memory: {slide.associated_images.get_cache_size()}")

    slide.close()


def coordinate_system_example(slide_path: str):
    """
    Example showing FastSlide's level-native coordinate system

    MAJOR DIFFERENCE FROM OPENSLIDE:
    ================================

    OpenSlide coordinates:
    - Always specify coordinates in level-0 space
    - read_region(location, level, size) where location is always level-0
    - Library automatically converts coordinates to level-native space

    FastSlide coordinates:
    - Specify coordinates in the native space of the level you're reading
    - read_region(x, y, width, height, level) where x,y are level-native
    - More efficient (no coordinate conversion overhead)
    - Conversion utilities provided when needed
    """
    print(f"\n=== Coordinate System: {Path(slide_path).name} ===")

    with fastslide.FastSlide.from_file_path(slide_path) as slide:
        print("FastSlide uses LEVEL-NATIVE coordinates (more efficient)")
        print("OpenSlide uses LEVEL-0 coordinates (requires conversion)\n")

        # Get slide dimensions for reference
        level0_width, level0_height = slide.dimensions
        print(f"Level 0 dimensions: {level0_width} x {level0_height}")

        if slide.level_count > 1:
            level1_width, level1_height = slide.level_dimensions[1]
            downsample = slide.level_downsamples[1]
            print(f"Level 1 dimensions: {level1_width} x {level1_height}")
            print(f"Level 1 downsample: {downsample:.1f}x\n")

            # Example: Reading a region from the center of the slide
            center_level0_x = level0_width // 2
            center_level0_y = level0_height // 2
            region_size = 512

            print(f"Target region: center of slide, {region_size}x{region_size} pixels")
            print(f"Level-0 center coordinates: ({center_level0_x}, {center_level0_y})")

            # Method 1: FastSlide level-native approach
            print("\n--- FastSlide Level-Native Approach ---")

            # Read from level 0 (coordinates are already level-native)
            region_l0 = slide.read_region((center_level0_x, center_level0_y), 0, (region_size, region_size))
            print(f"Level 0: read_region({center_level0_x}, {center_level0_y}, {region_size}, {region_size}, level=0)")
            print(f"         Result: {region_l0.shape}")

            # Read from level 1 (need level-1 native coordinates)
            center_level1_x, center_level1_y = slide.convert_level0_to_level_native(
                center_level0_x, center_level0_y, level=1
            )
            region_l1 = slide.read_region((center_level1_x, center_level1_y), 1, (region_size, region_size))
            print(
                f"Level 1: convert_coords({center_level0_x}, {center_level0_y}) -> ({center_level1_x}, {center_level1_y})"
            )
            print(f"         read_region({center_level1_x}, {center_level1_y}, {region_size}, {region_size}, level=1)")
            print(f"         Result: {region_l1.shape}")

            # Method 2: OpenSlide-compatible helper function
            print("\n--- OpenSlide-Compatible Helper ---")

            def read_region_openslide_style(slide, location, level, size):
                """Helper function that mimics OpenSlide's coordinate system"""
                x, y = location
                width, height = size

                if level == 0:
                    return slide.read_region((x, y), 0, (width, height))
                else:
                    # Convert level-0 coordinates to level-native
                    native_x, native_y = slide.convert_level0_to_level_native(x, y, level)
                    return slide.read_region((native_x, native_y), level, (width, height))

            # This mimics OpenSlide's read_region signature
            region_openslide_style = read_region_openslide_style(
                slide, location=(center_level0_x, center_level0_y), level=1, size=(region_size, region_size)
            )
            print(
                f"OpenSlide style: read_region(location=({center_level0_x}, {center_level0_y}), level=1, size=({region_size}, {region_size}))"
            )
            print(f"                 Result: {region_openslide_style.shape}")

            # Verify both methods give same result
            import numpy as np

            if np.array_equal(region_l1, region_openslide_style):
                print("✓ Both methods produce identical results")
            else:
                print("✗ Methods produced different results (unexpected)")


def region_reading_example(slide_path: str):
    """Example showing different ways to read regions"""
    print(f"\n=== Region Reading: {Path(slide_path).name} ===")

    with fastslide.FastSlide.from_file_path(slide_path) as slide:
        print("Reading regions from multiple levels using level-native coordinates\n")

        # Read tiles from different levels
        for level in range(min(3, slide.level_count)):
            width, height = slide.level_dimensions[level]
            downsample = slide.level_downsamples[level]

            # Calculate center coordinates in this level's native space
            center_x = width // 2
            center_y = height // 2
            tile_size = min(256, width // 4, height // 4)  # Ensure tile fits

            print(f"Level {level} (downsample {downsample:.1f}x):")
            print(f"  Dimensions: {width} x {height}")
            print(f"  Reading {tile_size}x{tile_size} tile from center ({center_x}, {center_y})")

            tile = slide.read_region((center_x, center_y), level, (tile_size, tile_size))
            print(f"  Result: {tile.shape}, dtype: {tile.dtype}")

            # Convert these coordinates to level-0 space for reference
            level0_x, level0_y = slide.convert_level_native_to_level0(center_x, center_y, level)
            print(f"  Equivalent level-0 coordinates: ({level0_x}, {level0_y})")
            print()

        # Performance comparison: level-native vs converted coordinates
        print("Performance comparison:")
        level = min(1, slide.level_count - 1)  # Use level 1 if available

        # Method 1: Direct level-native (efficient)
        start_time = time.time()
        for i in range(10):
            tile = slide.read_region((i * 100, i * 100), level, (256, 256))
        native_time = time.time() - start_time

        # Method 2: With coordinate conversion (less efficient)
        start_time = time.time()
        for i in range(10):
            level0_x, level0_y = 400 + i * 100, 400 + i * 100  # Some level-0 coords
            native_x, native_y = slide.convert_level0_to_level_native(level0_x, level0_y, level)
            tile = slide.read_region((native_x, native_y), level, (256, 256))
        converted_time = time.time() - start_time

        print(f"  Level-native coordinates: {native_time * 1000:.1f} ms")
        print(f"  With coordinate conversion: {converted_time * 1000:.1f} ms")
        print(f"  Performance benefit: {converted_time / native_time:.1f}x faster")


def associated_images_example(slide_path: str):
    """
    Example showing lazy loading of associated images

    OPENSLIDE COMPARISON:
    OpenSlide: slide.associated_images['thumbnail'] (loads immediately)
    FastSlide: slide.associated_images['thumbnail'] (loads on first access)
    """
    print(f"\n=== Associated Images (Lazy Loading): {Path(slide_path).name} ===")

    with fastslide.FastSlide.from_file_path(slide_path) as slide:
        print("FastSlide uses LAZY LOADING for memory efficiency")
        print("OpenSlide loads associated images immediately\n")

        # Check what's available without loading anything
        available_images = slide.associated_images.keys()
        print(f"Available associated images: {list(available_images)}")
        print(f"Images currently in memory: {slide.associated_images.get_cache_size()}")

        for name in available_images:
            print(f"\n--- Processing '{name}' ---")

            # Get dimensions without loading the image (FastSlide feature)
            # OpenSlide doesn't have this - you have to load to get dimensions
            width, height = slide.associated_images.get_dimensions(name)
            print(f"Dimensions: {width} x {height} (obtained without loading)")

            # Check if image exists without loading (FastSlide feature)
            exists = name in slide.associated_images
            print(f"Exists check: {exists} (no loading performed)")

            # Now actually load the image (lazy loading)
            print("Loading image...")
            try:
                start_time = time.time()
                image = slide.associated_images[name]
                load_time = time.time() - start_time

                print(f"Loaded: {image.shape}, dtype: {image.dtype}")
                print(f"Load time: {load_time * 1000:.1f} ms")
                print(f"Images now in memory: {slide.associated_images.get_cache_size()}")

                # Second access should be instant (cached)
                start_time = time.time()
                image_cached = slide.associated_images[name]
                cached_time = time.time() - start_time
                print(f"Cached access time: {cached_time * 1000:.3f} ms")

            except Exception as e:
                print(f"Failed to load: {e}")

        print(f"\nFinal cache status: {slide.associated_images.get_cache_size()} images in memory")

        # Clear cache to free memory
        slide.associated_images.clear_cache()
        print(f"After cache clear: {slide.associated_images.get_cache_size()} images in memory")


def pytorch_integration_example(slide_path: str):
    """Example showing PyTorch integration for deep learning"""
    print(f"\n=== PyTorch Integration: {Path(slide_path).name} ===")

    try:
        import torch
        import torch.nn.functional as F
    except ImportError:
        print("PyTorch not available, skipping example")
        return

    with fastslide.FastSlide.from_file_path(slide_path) as slide:
        # Read a tile for deep learning (common size: 224x224 for pretrained models)
        tile_size = 224
        level = slide.get_best_level_for_downsample(4.0)  # Use ~4x downsampled level

        # Get center coordinates in the selected level
        width, height = slide.level_dimensions[level]
        center_x = width // 2
        center_y = height // 2

        print(f"Reading {tile_size}x{tile_size} tile from level {level}")
        print(f"Level {level} downsample: {slide.level_downsamples[level]:.1f}x")

        tile_np = slide.read_region(center_x, center_y, tile_size, tile_size, level)
        print(f"NumPy array: {tile_np.shape}, dtype: {tile_np.dtype}")

        # Convert to PyTorch tensor (H,W,C) -> (C,H,W)
        tile_tensor = torch.from_numpy(tile_np).permute(2, 0, 1).float() / 255.0
        print(f"PyTorch tensor: {tile_tensor.shape}, dtype: {tile_tensor.dtype}")

        # Add batch dimension for CNN input
        batch_tensor = tile_tensor.unsqueeze(0)  # (1,C,H,W)
        print(f"Batch tensor: {batch_tensor.shape}")

        # Example: Resize if needed (common for pretrained models)
        if tile_tensor.shape[1] != 224 or tile_tensor.shape[2] != 224:
            resized = F.interpolate(batch_tensor, size=(224, 224), mode="bilinear", align_corners=False)
            print(f"Resized tensor: {resized.shape}")
        else:
            resized = batch_tensor

        # Example: Normalize for pretrained models (ImageNet normalization)
        mean = torch.tensor([0.485, 0.456, 0.406]).view(1, 3, 1, 1)
        std = torch.tensor([0.229, 0.224, 0.225]).view(1, 3, 1, 1)
        normalized = (resized - mean) / std
        print(f"Normalized tensor: {normalized.shape}, mean: {normalized.mean():.3f}, std: {normalized.std():.3f}")

        print("✓ Ready for CNN inference!")


def multithreaded_example(slide_path: str):
    """
    Example showing thread-safe usage with concurrent access

    OPENSLIDE COMPARISON:
    OpenSlide: NOT thread-safe, requires one reader per thread
    FastSlide: Thread-safe, single reader can be used by multiple threads
    """
    print(f"\n=== Thread Safety: {Path(slide_path).name} ===")

    with fastslide.FastSlide.from_file_path(slide_path) as slide:
        print("FastSlide readers are THREAD-SAFE")
        print("OpenSlide readers are NOT thread-safe\n")

        results = []
        lock = threading.Lock()

        def read_worker(thread_id: int):
            """Worker function that reads different regions"""
            try:
                # Each thread reads from a different area (level-native coordinates)
                level = 1 if slide.level_count > 1 else 0
                level_width, level_height = slide.level_dimensions[level]

                # Calculate thread-specific coordinates
                x = (thread_id * 200) % (level_width - 512)
                y = (thread_id * 150) % (level_height - 512)

                start_time = time.time()

                # Multiple reads per thread to stress test
                for i in range(3):
                    tile = slide.read_region(x + i * 50, y + i * 50, 256, 256, level)

                end_time = time.time()

                with lock:
                    results.append(
                        {
                            "thread_id": thread_id,
                            "shape": tile.shape,
                            "level": level,
                            "coords": (x, y),
                            "time_ms": (end_time - start_time) * 1000,
                        }
                    )

            except Exception as e:
                with lock:
                    print(f"Thread {thread_id} failed: {e}")

        # Create and start multiple threads
        num_threads = 8
        threads = []

        print(f"Starting {num_threads} concurrent readers...")
        start_time = time.time()

        for i in range(num_threads):
            thread = threading.Thread(target=read_worker, args=(i,))
            threads.append(thread)
            thread.start()

        # Wait for all threads to complete
        for thread in threads:
            thread.join()

        total_time = time.time() - start_time

        print(f"All threads completed in {total_time * 1000:.1f} ms")
        for result in sorted(results, key=lambda x: x["thread_id"]):
            print(
                f"  Thread {result['thread_id']}: {result['shape']} from level {result['level']} "
                f"coords {result['coords']} in {result['time_ms']:.1f} ms"
            )

        print(f"\nTotal reads: {len(results) * 3}")
        print(f"Average time per read: {total_time * 1000 / (len(results) * 3):.1f} ms")


def caching_example(slide_path: str):
    """
    Example showing advanced cache management

    OPENSLIDE COMPARISON:
    OpenSlide: Basic cache, limited introspection
    FastSlide: Advanced cache with detailed statistics and management
    """
    print(f"\n=== Advanced Caching: {Path(slide_path).name} ===")

    with fastslide.FastSlide.from_file_path(slide_path) as slide:
        print("FastSlide provides ADVANCED CACHE MANAGEMENT")
        print("OpenSlide has basic caching with limited control\n")

        # Create a custom cache manager
        cache_manager = fastslide.CacheManager.create(capacity=50)
        slide.set_cache_manager(cache_manager)

        print(f"Created custom cache with capacity: 50 tiles")
        print(f"Cache enabled: {slide.cache_enabled}")

        # Read regions to populate cache
        level = 1 if slide.level_count > 1 else 0
        region_specs = [
            (100, 100, 256, 256),
            (200, 200, 256, 256),
            (300, 300, 256, 256),
            (100, 100, 256, 256),  # Repeat to test cache hit
            (400, 400, 256, 256),
        ]

        print(f"\nReading {len(region_specs)} regions (including one repeat):")

        for i, (x, y, w, h) in enumerate(region_specs):
            start_time = time.time()
            tile = slide.read_region(x, y, w, h, level)
            read_time = (time.time() - start_time) * 1000

            # Get detailed cache statistics
            stats = cache_manager.get_detailed_stats()

            print(f"  Read {i + 1}: ({x}, {y}) -> {tile.shape} in {read_time:.1f} ms")
            print(
                f"    Cache: {stats.hits} hits, {stats.misses} misses, "
                f"hit ratio: {stats.hit_ratio:.3f}, size: {stats.size}/{stats.capacity}"
            )

        # Test cache resizing
        print(f"\nResizing cache from 50 to 100...")
        cache_manager.resize(100)
        stats = cache_manager.get_detailed_stats()
        print(f"New capacity: {stats.capacity}")

        # Test global cache
        print(f"\nSwitching to global cache...")
        global_cache = fastslide.GlobalCacheManager.instance()
        global_cache.set_capacity(200)
        slide.use_global_cache()

        # Read a region with global cache
        tile = slide.read_region(500, 500, 512, 512, level)
        global_stats = global_cache.get_stats()
        print(f"Global cache stats: {global_stats.hits} hits, {global_stats.misses} misses")
        print(f"Global cache capacity: {global_stats.capacity}")

        # Clear global cache
        global_cache.clear()
        print("Global cache cleared")


def metadata_exploration_example(slide_path: str):
    """Example showing comprehensive metadata exploration"""
    print(f"\n=== Metadata Exploration: {Path(slide_path).name} ===")

    with fastslide.FastSlide.from_file_path(slide_path) as slide:
        # FastSlide combines properties + metadata in one dict
        # OpenSlide separates slide.properties and format-specific metadata
        props = slide.properties

        print("All properties and metadata (combined):")
        for key, value in sorted(props.items()):
            print(f"  {key}: {value}")

        print(f"\nSlide format: {slide.format}")
        print(f"Source file: {slide.source_path}")

        # Pyramid analysis
        print(f"\nPyramid structure:")
        print(f"  Levels: {slide.level_count}")

        for level in range(slide.level_count):
            width, height = slide.level_dimensions[level]
            downsample = slide.level_downsamples[level]

            # Calculate physical dimensions
            mpp_x = props.get("mpp_x", 1.0)
            mpp_y = props.get("mpp_y", 1.0)
            width_um = width * mpp_x * downsample
            height_um = height * mpp_y * downsample

            print(f"  Level {level}: {width:,} x {height:,} pixels")
            print(f"            Downsample: {downsample:.1f}x")
            print(f"            Physical: {width_um:.0f} x {height_um:.0f} μm")
            print(f"            Physical: {width_um / 1000:.1f} x {height_um / 1000:.1f} mm")

        # Best level analysis
        print(f"\nBest levels for common downsamples:")
        target_downsamples = [1.0, 2.0, 4.0, 8.0, 16.0, 32.0]
        for target in target_downsamples:
            best_level = slide.get_best_level_for_downsample(target)
            if best_level < slide.level_count:
                actual = slide.level_downsamples[best_level]
                print(f"  {target:4.1f}x -> Level {best_level} (actual: {actual:.1f}x)")


def format_support_example():
    """Example showing format support and detection"""
    print(f"\n=== Format Support ===")

    print(f"FastSlide version: {fastslide.__version__}")

    # Get supported extensions
    extensions = fastslide.get_supported_extensions()
    print(f"Supported extensions: {extensions}")

    # Test format detection
    test_files = [
        "test.svs",
        "test.qptiff",
        "test.tiff",
        "test.jpg",  # Not supported
        "test.unknown",  # Not supported
    ]

    print(f"\nFormat support check:")
    for filename in test_files:
        supported = fastslide.is_supported(filename)
        status = "✓ Supported" if supported else "✗ Not supported"
        print(f"  {filename}: {status}")


def openslide_migration_guide():
    """Guide for migrating from OpenSlide to FastSlide"""
    print(f"\n=== OpenSlide Migration Guide ===")

    migration_examples = [
        {
            "description": "Opening a slide",
            "openslide": "slide = openslide.OpenSlide('slide.svs')",
            "fastslide": "slide = fastslide.FastSlide.from_file_path('slide.svs')",
        },
        {
            "description": "Getting dimensions",
            "openslide": "width, height = slide.dimensions",
            "fastslide": "width, height = slide.dimensions  # Same!",
        },
        {
            "description": "Reading a region (key difference!)",
            "openslide": "region = slide.read_region((1000, 1000), 1, (512, 512))",
            "fastslide": "region = slide.read_region((1000, 1000), 1, (512, 512))",
        },
        {
            "description": "Getting level count",
            "openslide": "count = len(slide.level_dimensions)",
            "fastslide": "count = slide.level_count",
        },
        {
            "description": "Getting level dimensions",
            "openslide": "dims = slide.level_dimensions[level]",
            "fastslide": "dims = slide.level_dimensions[level]  # Same!",
        },
        {
            "description": "Associated images",
            "openslide": "thumb = slide.associated_images['thumbnail']  # Loads immediately",
            "fastslide": "thumb = slide.associated_images['thumbnail']  # Lazy loading",
        },
        {
            "description": "Properties",
            "openslide": "props = slide.properties  # String values only",
            "fastslide": "props = slide.properties  # Typed values + metadata",
        },
        {
            "description": "Best level for downsample",
            "openslide": "level = slide.get_best_level_for_downsample(4.0)",
            "fastslide": "level = slide.get_best_level_for_downsample(4.0)  # Same!",
        },
        {
            "description": "Context manager",
            "openslide": "with openslide.OpenSlide('slide.svs') as slide:",
            "fastslide": "with fastslide.FastSlide.from_file_path('slide.svs') as slide:",
        },
    ]

    for example in migration_examples:
        print(f"\n{example['description']}:")
        print(f"  OpenSlide:  {example['openslide']}")
        print(f"  FastSlide:  {example['fastslide']}")


def main(slide_path: str):
    """Main example function demonstrating all FastSlide features"""

    # Check if slide exists and is supported
    if not Path(slide_path).exists():
        print(f"❌ Slide not found: {slide_path}")
        return

    if not fastslide.is_supported(slide_path):
        print(f"❌ Unsupported format: {slide_path}")
        format_support_example()
        return

    print("🔬 FastSlide Python API Examples")
    print("=" * 50)

    try:
        # Run all examples
        basic_usage_example(slide_path)
        coordinate_system_example(slide_path)
        region_reading_example(slide_path)
        associated_images_example(slide_path)
        pytorch_integration_example(slide_path)
        multithreaded_example(slide_path)
        caching_example(slide_path)
        metadata_exploration_example(slide_path)
        format_support_example()
        openslide_migration_guide()

        print(f"\n✅ All examples completed successfully!")

    except Exception as e:
        print(f"❌ Error processing {slide_path}: {e}")
        import traceback

        traceback.print_exc()


if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="FastSlide Python Usage Examples",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  python python_usage.py slide.svs
  python python_usage.py /path/to/slide.qptiff
  
        """,
    )
    parser.add_argument("slide_path", type=str, help="Path to the slide file")
    args = parser.parse_args()

    main(args.slide_path)
