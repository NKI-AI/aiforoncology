// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of FastSlide.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the FastSlide project root.

#ifndef AIFO_FASTSLIDE_INCLUDE_FASTSLIDE_C_UTILITIES_H_
#define AIFO_FASTSLIDE_INCLUDE_FASTSLIDE_C_UTILITIES_H_

#include <stddef.h>

#include "fastslide/c/histogram.h"
#include "fastslide/c/image.h"
#include "fastslide/c/slide_reader.h"

#ifdef __cplusplus
extern "C" {
#endif

/// @brief Combine spectral image channels using pre-computed display ranges
///
/// This function combines a spectral image using channel metadata
/// and display ranges to create an RGB image with proper spectral blending.
/// This is equivalent to the C++ CombineSpectralChannelsWithDisplayRanges
/// function.
///
/// @param image Input spectral image
/// @param channel_metadata Array of channel metadata containing colors
/// @param num_channels Number of channels in metadata array
/// @param display_ranges Array of display ranges for each channel (can be NULL)
/// @param num_ranges Number of display ranges (ignored if display_ranges is
/// NULL)
/// @return Combined RGB image handle or NULL on failure
FastSlideImage* fastslide_combine_spectral_channels_with_display_ranges(
    const FastSlideImage* image,
    const FastSlideChannelMetadata* channel_metadata, int num_channels,
    const FastSlideDisplayRange* display_ranges, int num_ranges);

/// @brief Combine spectral image channels using computed display ranges
/// from histograms
///
/// This is a convenience function that combines spectral channels
/// using display ranges computed from the provided histograms
/// with the specified saturation level.
///
/// @param image Input spectral image
/// @param channel_metadata Array of channel metadata containing colors
/// @param num_channels Number of channels in metadata array
/// @param histograms Array of histograms for display range computation
/// @param num_histograms Number of histograms
/// @param saturation Saturation level for display range computation (e.g.,
/// 0.001)
/// @return Combined RGB image handle or NULL on failure
FastSlideImage* fastslide_combine_spectral_channels_with_histograms(
    const FastSlideImage* image,
    const FastSlideChannelMetadata* channel_metadata, int num_channels,
    FastSlideHistogram** histograms, int num_histograms, double saturation);

/// @brief Resample an image using Lanczos3 algorithm
///
/// This function performs high-quality image resampling
/// using the Lanczos3 kernel.
/// The input image must have separate planar configuration.
///
/// @param image Input image to resample
/// @param output_width Target output width
/// @param output_height Target output height
/// @return Resampled image handle or NULL on failure
FastSlideImage* fastslide_lanczos_resample(const FastSlideImage* image,
                                           uint32_t output_width,
                                           uint32_t output_height);

/// @brief Resample an image using Lanczos2 algorithm
///
/// This function performs high-quality image resampling
/// using the Lanczos2 kernel.
/// The input image must have separate planar configuration.
///
/// @param image Input image to resample
/// @param output_width Target output width
/// @param output_height Target output height
/// @return Resampled image handle or NULL on failure
FastSlideImage* fastslide_lanczos2_resample(const FastSlideImage* image,
                                            uint32_t output_width,
                                            uint32_t output_height);

/// @brief Resample an image using Cosine-windowed sinc algorithm
///
/// This function performs high-quality image resampling
/// using the Cosine3 kernel.
/// The input image must have separate planar configuration.
///
/// @param image Input image to resample
/// @param output_width Target output width
/// @param output_height Target output height
/// @return Resampled image handle or NULL on failure
FastSlideImage* fastslide_cosine_resample(const FastSlideImage* image,
                                          uint32_t output_width,
                                          uint32_t output_height);

/// @brief Resample an image using average downsampling
///
/// This function performs downsampling by averaging pixels in blocks.
/// The input image must have separate planar configuration.
/// The factor must be a power of two and greater than 0.
///
/// @param image Input image to resample
/// @param factor Downsampling factor (must be power of 2)
/// @return Resampled image handle or NULL on failure
FastSlideImage* fastslide_average_resample(const FastSlideImage* image,
                                           uint32_t factor);

/// @brief Resample an image using 2x2 average downsampling
///
/// Convenience function for 2x2 average downsampling.
/// The input image must have separate planar configuration.
///
/// @param image Input image to resample
/// @return Resampled image handle or NULL on failure
FastSlideImage* fastslide_average_2x2_resample(const FastSlideImage* image);

/// @brief Resample an image using 4x4 average downsampling
///
/// Convenience function for 4x4 average downsampling.
/// The input image must have separate planar configuration.
///
/// @param image Input image to resample
/// @return Resampled image handle or NULL on failure
FastSlideImage* fastslide_average_4x4_resample(const FastSlideImage* image);

/// @brief Resample an image using 8x8 average downsampling
///
/// Convenience function for 8x8 average downsampling.
/// The input image must have separate planar configuration.
///
/// @param image Input image to resample
/// @return Resampled image handle or NULL on failure
FastSlideImage* fastslide_average_8x8_resample(const FastSlideImage* image);

// Example utilities for PNG I/O and vips initialization

/// @brief Initialize vips library for PNG I/O
/// @details Must be called before using PNG save/load functions
void fastslide_examples_initialize_vips(void);

/// @brief Cleanup vips library
/// @details Should be called when done with PNG operations
void fastslide_examples_cleanup_vips(void);

/// @brief Save an RGB image as PNG using libvips
/// @param image RGB image to save (must be uint8 format)
/// @param filename Output PNG filename
/// @return 1 on success, 0 on failure
int fastslide_examples_save_as_png(const FastSlideImage* image,
                                   const char* filename);

/// @brief Load an image from PNG using libvips
/// @param filename Input PNG filename
/// @return Loaded image as RGB uint8, or NULL if failed
FastSlideImage* fastslide_examples_load_from_png(const char* filename);

#ifdef __cplusplus
}
#endif

#endif  // AIFO_FASTSLIDE_INCLUDE_FASTSLIDE_C_UTILITIES_H_
