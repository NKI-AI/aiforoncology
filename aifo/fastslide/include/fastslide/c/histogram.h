// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of FastSlide.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the FastSlide project root.

#ifndef AIFO_FASTSLIDE_INCLUDE_FASTSLIDE_C_HISTOGRAM_H_
#define AIFO_FASTSLIDE_INCLUDE_FASTSLIDE_C_HISTOGRAM_H_

#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>

#include "fastslide/c/image.h"

#ifdef __cplusplus
extern "C" {
#endif

/// @brief Opaque histogram handle
typedef struct FastSlideHistogram FastSlideHistogram;

/// @brief Display range pair
typedef struct {
  double min_value;
  double max_value;
} FastSlideDisplayRange;

// Factory functions

/// @brief Create histogram from double array
/// @param values Array of double values
/// @param num_values Number of values
/// @param n_bins Number of histogram bins
/// @return Histogram handle or NULL on failure
FastSlideHistogram* fastslide_histogram_create_from_doubles(
    const double* values, size_t num_values, int n_bins);

/// @brief Create histogram from float array
/// @param values Array of float values
/// @param num_values Number of values
/// @param n_bins Number of histogram bins
/// @return Histogram handle or NULL on failure
FastSlideHistogram* fastslide_histogram_create_from_floats(const float* values,
                                                           size_t num_values,
                                                           int n_bins);

/// @brief Create histogram from uint8 array
/// @param values Array of uint8 values
/// @param num_values Number of values
/// @param n_bins Number of histogram bins
/// @return Histogram handle or NULL on failure
FastSlideHistogram* fastslide_histogram_create_from_uint8(const uint8_t* values,
                                                          size_t num_values,
                                                          int n_bins);

/// @brief Create histogram from uint16 array
/// @param values Array of uint16 values
/// @param num_values Number of values
/// @param n_bins Number of histogram bins
/// @return Histogram handle or NULL on failure
FastSlideHistogram* fastslide_histogram_create_from_uint16(
    const uint16_t* values, size_t num_values, int n_bins);

/// @brief Create histogram from image channel
/// @param image Image handle
/// @param channel Channel index
/// @param n_bins Number of histogram bins
/// @return Histogram handle or NULL on failure
FastSlideHistogram* fastslide_histogram_create_from_image_channel(
    const FastSlideImage* image, uint32_t channel, int n_bins);

/// @brief Create histograms from all image channels
/// @param image Image handle
/// @param n_bins Number of histogram bins
/// @param histograms Output array of histogram handles (allocated by function)
/// @param num_histograms Output number of histograms created
/// @return 1 on success, 0 on failure
int fastslide_histogram_create_from_image_channels(
    const FastSlideImage* image, int n_bins, FastSlideHistogram*** histograms,
    int* num_histograms);

// Properties

/// @brief Get histogram edge minimum
/// @param histogram Histogram handle
/// @return Edge minimum value or NaN on failure
double fastslide_histogram_get_edge_min(const FastSlideHistogram* histogram);

/// @brief Get histogram edge maximum
/// @param histogram Histogram handle
/// @return Edge maximum value or NaN on failure
double fastslide_histogram_get_edge_max(const FastSlideHistogram* histogram);

/// @brief Get histogram edge range
/// @param histogram Histogram handle
/// @return Edge range (max - min) or NaN on failure
double fastslide_histogram_get_edge_range(const FastSlideHistogram* histogram);

/// @brief Get number of bins
/// @param histogram Histogram handle
/// @return Number of bins or -1 on failure
int fastslide_histogram_get_num_bins(const FastSlideHistogram* histogram);

/// @brief Check if histogram represents integer values
/// @param histogram Histogram handle
/// @return 1 if integer, 0 if not or on failure
int fastslide_histogram_is_integer(const FastSlideHistogram* histogram);

// Bin access

/// @brief Get bin left edge
/// @param histogram Histogram handle
/// @param bin_index Bin index
/// @return Left edge value or NaN on failure
double fastslide_histogram_get_bin_left_edge(
    const FastSlideHistogram* histogram, int bin_index);

/// @brief Get bin right edge
/// @param histogram Histogram handle
/// @param bin_index Bin index
/// @return Right edge value or NaN on failure
double fastslide_histogram_get_bin_right_edge(
    const FastSlideHistogram* histogram, int bin_index);

/// @brief Get bin center
/// @param histogram Histogram handle
/// @param bin_index Bin index
/// @return Center value or NaN on failure
double fastslide_histogram_get_bin_center(const FastSlideHistogram* histogram,
                                          int bin_index);

/// @brief Get bin width
/// @param histogram Histogram handle
/// @param bin_index Bin index
/// @return Width value or NaN on failure
double fastslide_histogram_get_bin_width(const FastSlideHistogram* histogram,
                                         int bin_index);

/// @brief Get bin count
/// @param histogram Histogram handle
/// @param bin_index Bin index
/// @return Count value or -1 on failure
int64_t fastslide_histogram_get_bin_count(const FastSlideHistogram* histogram,
                                          int bin_index);

/// @brief Get normalized bin count
/// @param histogram Histogram handle
/// @param bin_index Bin index
/// @return Normalized count value or NaN on failure
double fastslide_histogram_get_bin_normalized_count(
    const FastSlideHistogram* histogram, int bin_index);

/// @brief Get bin index for value
/// @param histogram Histogram handle
/// @param value Value to find
/// @return Bin index or -1 if out of range
int fastslide_histogram_get_bin_index_for_value(
    const FastSlideHistogram* histogram, double value);

// Statistics

/// @brief Get minimum value of input data
/// @param histogram Histogram handle
/// @return Minimum value or NaN on failure
double fastslide_histogram_get_min_value(const FastSlideHistogram* histogram);

/// @brief Get maximum value of input data
/// @param histogram Histogram handle
/// @return Maximum value or NaN on failure
double fastslide_histogram_get_max_value(const FastSlideHistogram* histogram);

/// @brief Get mean value of input data
/// @param histogram Histogram handle
/// @return Mean value or NaN on failure
double fastslide_histogram_get_mean_value(const FastSlideHistogram* histogram);

/// @brief Get variance of input data
/// @param histogram Histogram handle
/// @return Variance or NaN on failure
double fastslide_histogram_get_variance(const FastSlideHistogram* histogram);

/// @brief Get standard deviation of input data
/// @param histogram Histogram handle
/// @return Standard deviation or NaN on failure
double fastslide_histogram_get_std_dev(const FastSlideHistogram* histogram);

/// @brief Get sum of input data
/// @param histogram Histogram handle
/// @return Sum value or NaN on failure
double fastslide_histogram_get_sum(const FastSlideHistogram* histogram);

/// @brief Get number of values in histogram
/// @param histogram Histogram handle
/// @return Number of values or -1 on failure
int64_t fastslide_histogram_get_num_values(const FastSlideHistogram* histogram);

/// @brief Get number of missing/NaN values
/// @param histogram Histogram handle
/// @return Number of missing values or -1 on failure
int64_t fastslide_histogram_get_num_missing_values(
    const FastSlideHistogram* histogram);

/// @brief Get maximum count in any bin
/// @param histogram Histogram handle
/// @return Maximum count or -1 on failure
int64_t fastslide_histogram_get_max_count(const FastSlideHistogram* histogram);

/// @brief Get maximum normalized count
/// @param histogram Histogram handle
/// @return Maximum normalized count or NaN on failure
double fastslide_histogram_get_max_normalized_count(
    const FastSlideHistogram* histogram);

/// @brief Get total count sum
/// @param histogram Histogram handle
/// @return Total count sum or -1 on failure
int64_t fastslide_histogram_get_count_sum(const FastSlideHistogram* histogram);

// Display range computation

/// @brief Compute display range based on saturation
/// @param histogram Histogram handle
/// @param saturation Saturation level (0.0 to 1.0)
/// @param range Output display range
/// @return 1 on success, 0 on failure
int fastslide_histogram_compute_display_range(
    const FastSlideHistogram* histogram, double saturation,
    FastSlideDisplayRange* range);

// Array access

/// @brief Get all bin edges
/// @param histogram Histogram handle
/// @param edges Output array pointer (allocated by function)
/// @param num_edges Output number of edges
/// @return 1 on success, 0 on failure
int fastslide_histogram_get_edges(const FastSlideHistogram* histogram,
                                  double** edges, int* num_edges);

/// @brief Get all bin counts
/// @param histogram Histogram handle
/// @param counts Output array pointer (allocated by function)
/// @param num_counts Output number of counts
/// @return 1 on success, 0 on failure
int fastslide_histogram_get_counts(const FastSlideHistogram* histogram,
                                   int64_t** counts, int* num_counts);

// Export functions

/// @brief Export histogram to CSV file
/// @param histogram Histogram handle
/// @param filename Output filename
/// @return 1 on success, 0 on failure
int fastslide_histogram_export_to_csv(const FastSlideHistogram* histogram,
                                      const char* filename);

/// @brief Export histogram to binary file
/// @param histogram Histogram handle
/// @param filename Output filename
/// @return 1 on success, 0 on failure
int fastslide_histogram_export_to_binary(const FastSlideHistogram* histogram,
                                         const char* filename);

/// @brief Export histogram to binary buffer
/// @param histogram Histogram handle
/// @param buffer Output buffer pointer (allocated by function)
/// @param buffer_size Output buffer size
/// @return 1 on success, 0 on failure
int fastslide_histogram_export_to_binary_buffer(
    const FastSlideHistogram* histogram, uint8_t** buffer, size_t* buffer_size);

/// @brief Get string representation
/// @param histogram Histogram handle
/// @param buffer Output string buffer
/// @param buffer_size Size of output buffer
/// @return Length of string or -1 on failure
int fastslide_histogram_to_string(const FastSlideHistogram* histogram,
                                  char* buffer, size_t buffer_size);

// Memory management

/// @brief Free histogram handle
/// @param histogram Histogram handle
void fastslide_histogram_free(FastSlideHistogram* histogram);

/// @brief Free array of histogram handles
/// @param histograms Array of histogram handles
/// @param num_histograms Number of histograms
void fastslide_histogram_free_array(FastSlideHistogram** histograms,
                                    int num_histograms);

/// @brief Free edges array
/// @param edges Edges array
void fastslide_histogram_free_edges(double* edges);

/// @brief Free counts array
/// @param counts Counts array
void fastslide_histogram_free_counts(int64_t* counts);

/// @brief Free binary buffer
/// @param buffer Binary buffer
void fastslide_histogram_free_binary_buffer(uint8_t* buffer);

#ifdef __cplusplus
}
#endif

#endif  // AIFO_FASTSLIDE_INCLUDE_FASTSLIDE_C_HISTOGRAM_H_
