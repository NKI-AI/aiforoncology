"""
FastSlide: High-performance, thread-safe digital pathology slide reader

This module provides Python bindings for the FastSlide C++ library,
enabling fast reading of digital pathology slides in various formats.

Key Features:
- Thread-safe slide reading for SVS, QPTIFF, and other formats
- Optional tile caching for improved performance with repeated/overlapping reads
- OpenSlide-compatible API for easy migration
- Memory-efficient handling of large slide files
- High-performance C++ backend with Python convenience

Usage Example:
    import fastslide
    from pathlib import Path

    # Create reader (supports both str and pathlib.Path)
    slide = fastslide.FastSlide.from_file_path("slide.svs")
    # or: slide = fastslide.FastSlide.from_file_path(Path("slide.svs"))

    # Create and set cache for improved performance
    cache = fastslide.CacheManager.create(capacity=100)
    slide.set_cache_manager(cache)

    # Read regions - OpenSlide compatible API
    tile = slide.read_region((1000, 1000), 0, (512, 512))

    # Check cache performance
    stats = cache.get_detailed_stats()
    print(f"Hit ratio: {stats.hit_ratio:.3f}")
"""

from fastslide._fastslide import *  # Import all symbols from the C++ extension

__version__ = "1.0.0"
