// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of FastSlide.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the FastSlide project root.

#include "fastslide/c/histogram.h"

#include <cstring>
#include <limits>
#include <span>
#include <string>
#include <utility>

#include "fastslide/c/image.h"
#include "fastslide/histogram.h"

// Forward declarations for external functions
extern "C" void fastslide_set_last_error(const char* message);

// Forward declaration for function in image.cpp
const fastslide::Image& fastslide_image_get_cpp_image(
    const FastSlideImage* image);

// Wrapper struct to hold the C++ Histogram
struct FastSlideHistogram {
  fastslide::Histogram histogram;

  explicit FastSlideHistogram(fastslide::Histogram h)
      : histogram(std::move(h)) {}
};

namespace {

void SetLastError(const char* message) {
  fastslide_set_last_error(message);
}

// Helper to get FastSlideImage internal image
const fastslide::Image& GetCppImage(const FastSlideImage* image) {
  return fastslide_image_get_cpp_image(image);
}

}  // namespace

// Factory functions

FastSlideHistogram* fastslide_histogram_create_from_doubles(
    const double* values, size_t num_values, int n_bins) {
  if (!values || num_values == 0 || n_bins <= 0) {
    SetLastError("invalid parameters for histogram creation");
    return nullptr;
  }

  try {
    std::span<const double> data_span(values, num_values);
    auto histogram = fastslide::Histogram(data_span, n_bins);
    return new FastSlideHistogram(std::move(histogram));
  } catch (const std::exception& e) {
    SetLastError(e.what());
    return nullptr;
  }
}

FastSlideHistogram* fastslide_histogram_create_from_floats(const float* values,
                                                           size_t num_values,
                                                           int n_bins) {
  if (!values || num_values == 0 || n_bins <= 0) {
    SetLastError("invalid parameters for histogram creation");
    return nullptr;
  }

  try {
    std::span<const float> data_span(values, num_values);
    auto histogram = fastslide::Histogram(data_span, n_bins);
    return new FastSlideHistogram(std::move(histogram));
  } catch (const std::exception& e) {
    SetLastError(e.what());
    return nullptr;
  }
}

FastSlideHistogram* fastslide_histogram_create_from_uint8(const uint8_t* values,
                                                          size_t num_values,
                                                          int n_bins) {
  if (!values || num_values == 0 || n_bins <= 0) {
    SetLastError("invalid parameters for histogram creation");
    return nullptr;
  }

  try {
    std::span<const uint8_t> data_span(values, num_values);
    auto histogram = fastslide::Histogram(data_span, n_bins);
    return new FastSlideHistogram(std::move(histogram));
  } catch (const std::exception& e) {
    SetLastError(e.what());
    return nullptr;
  }
}

FastSlideHistogram* fastslide_histogram_create_from_uint16(
    const uint16_t* values, size_t num_values, int n_bins) {
  if (!values || num_values == 0 || n_bins <= 0) {
    SetLastError("invalid parameters for histogram creation");
    return nullptr;
  }

  try {
    std::span<const uint16_t> data_span(values, num_values);
    auto histogram = fastslide::Histogram(data_span, n_bins);
    return new FastSlideHistogram(std::move(histogram));
  } catch (const std::exception& e) {
    SetLastError(e.what());
    return nullptr;
  }
}

FastSlideHistogram* fastslide_histogram_create_from_image_channel(
    const FastSlideImage* image, uint32_t channel, int n_bins) {
  if (!image || n_bins <= 0) {
    SetLastError("invalid parameters for histogram creation");
    return nullptr;
  }

  try {
    const auto& cpp_image = GetCppImage(image);
    auto histogram_or =
        fastslide::CreateHistogramFromImageChannel(cpp_image, channel, n_bins);
    if (!histogram_or.ok()) {
      SetLastError(std::string(histogram_or.status().message()).c_str());
      return nullptr;
    }

    return new FastSlideHistogram(std::move(histogram_or.value()));
  } catch (const std::exception& e) {
    SetLastError(e.what());
    return nullptr;
  }
}

int fastslide_histogram_create_from_image_channels(
    const FastSlideImage* image, int n_bins, FastSlideHistogram*** histograms,
    int* num_histograms) {
  if (!image || n_bins <= 0 || !histograms || !num_histograms) {
    SetLastError("invalid parameters for histogram creation");
    return 0;
  }

  try {
    const auto& cpp_image = GetCppImage(image);
    auto histograms_or =
        fastslide::CreateHistogramsFromImageChannels(cpp_image, n_bins);
    if (!histograms_or.ok()) {
      SetLastError(std::string(histograms_or.status().message()).c_str());
      return 0;
    }

    const auto& cpp_histograms = histograms_or.value();
    *num_histograms = static_cast<int>(cpp_histograms.size());

    if (*num_histograms == 0) {
      *histograms = nullptr;
      return 1;
    }

    *histograms = static_cast<FastSlideHistogram**>(
        malloc(*num_histograms * sizeof(FastSlideHistogram*)));
    if (!*histograms) {
      SetLastError("failed to allocate memory for histograms array");
      return 0;
    }

    for (int i = 0; i < *num_histograms; ++i) {
      (*histograms)[i] =
          new FastSlideHistogram(fastslide::Histogram(cpp_histograms[i]));
    }

    return 1;
  } catch (const std::exception& e) {
    SetLastError(e.what());
    return 0;
  }
}

// Properties

double fastslide_histogram_get_edge_min(const FastSlideHistogram* histogram) {
  if (!histogram) {
    SetLastError("histogram is null");
    return std::numeric_limits<double>::quiet_NaN();
  }
  return histogram->histogram.GetEdgeMin();
}

double fastslide_histogram_get_edge_max(const FastSlideHistogram* histogram) {
  if (!histogram) {
    SetLastError("histogram is null");
    return std::numeric_limits<double>::quiet_NaN();
  }
  return histogram->histogram.GetEdgeMax();
}

double fastslide_histogram_get_edge_range(const FastSlideHistogram* histogram) {
  if (!histogram) {
    SetLastError("histogram is null");
    return std::numeric_limits<double>::quiet_NaN();
  }
  return histogram->histogram.GetEdgeRange();
}

int fastslide_histogram_get_num_bins(const FastSlideHistogram* histogram) {
  if (!histogram) {
    SetLastError("histogram is null");
    return -1;
  }
  return histogram->histogram.GetNBins();
}

int fastslide_histogram_is_integer(const FastSlideHistogram* histogram) {
  if (!histogram) {
    SetLastError("histogram is null");
    return 0;
  }
  return histogram->histogram.IsInteger() ? 1 : 0;
}

// Bin access

double fastslide_histogram_get_bin_left_edge(
    const FastSlideHistogram* histogram, int bin_index) {
  if (!histogram) {
    SetLastError("histogram is null");
    return std::numeric_limits<double>::quiet_NaN();
  }

  try {
    return histogram->histogram.GetBinLeftEdge(bin_index);
  } catch (const std::exception& e) {
    SetLastError(e.what());
    return std::numeric_limits<double>::quiet_NaN();
  }
}

double fastslide_histogram_get_bin_right_edge(
    const FastSlideHistogram* histogram, int bin_index) {
  if (!histogram) {
    SetLastError("histogram is null");
    return std::numeric_limits<double>::quiet_NaN();
  }

  try {
    return histogram->histogram.GetBinRightEdge(bin_index);
  } catch (const std::exception& e) {
    SetLastError(e.what());
    return std::numeric_limits<double>::quiet_NaN();
  }
}

double fastslide_histogram_get_bin_center(const FastSlideHistogram* histogram,
                                          int bin_index) {
  if (!histogram) {
    SetLastError("histogram is null");
    return std::numeric_limits<double>::quiet_NaN();
  }

  try {
    return histogram->histogram.GetBinCenter(bin_index);
  } catch (const std::exception& e) {
    SetLastError(e.what());
    return std::numeric_limits<double>::quiet_NaN();
  }
}

double fastslide_histogram_get_bin_width(const FastSlideHistogram* histogram,
                                         int bin_index) {
  if (!histogram) {
    SetLastError("histogram is null");
    return std::numeric_limits<double>::quiet_NaN();
  }

  try {
    return histogram->histogram.GetBinWidth(bin_index);
  } catch (const std::exception& e) {
    SetLastError(e.what());
    return std::numeric_limits<double>::quiet_NaN();
  }
}

int64_t fastslide_histogram_get_bin_count(const FastSlideHistogram* histogram,
                                          int bin_index) {
  if (!histogram) {
    SetLastError("histogram is null");
    return -1;
  }

  try {
    return histogram->histogram.GetCountsForBin(bin_index);
  } catch (const std::exception& e) {
    SetLastError(e.what());
    return -1;
  }
}

double fastslide_histogram_get_bin_normalized_count(
    const FastSlideHistogram* histogram, int bin_index) {
  if (!histogram) {
    SetLastError("histogram is null");
    return std::numeric_limits<double>::quiet_NaN();
  }

  try {
    return histogram->histogram.GetNormalizedCountsForBin(bin_index);
  } catch (const std::exception& e) {
    SetLastError(e.what());
    return std::numeric_limits<double>::quiet_NaN();
  }
}

int fastslide_histogram_get_bin_index_for_value(
    const FastSlideHistogram* histogram, double value) {
  if (!histogram) {
    SetLastError("histogram is null");
    return -1;
  }

  return histogram->histogram.GetBinIndexForValue(value);
}

// Statistics

double fastslide_histogram_get_min_value(const FastSlideHistogram* histogram) {
  if (!histogram) {
    SetLastError("histogram is null");
    return std::numeric_limits<double>::quiet_NaN();
  }
  return histogram->histogram.GetMinValue();
}

double fastslide_histogram_get_max_value(const FastSlideHistogram* histogram) {
  if (!histogram) {
    SetLastError("histogram is null");
    return std::numeric_limits<double>::quiet_NaN();
  }
  return histogram->histogram.GetMaxValue();
}

double fastslide_histogram_get_mean_value(const FastSlideHistogram* histogram) {
  if (!histogram) {
    SetLastError("histogram is null");
    return std::numeric_limits<double>::quiet_NaN();
  }
  return histogram->histogram.GetMeanValue();
}

double fastslide_histogram_get_variance(const FastSlideHistogram* histogram) {
  if (!histogram) {
    SetLastError("histogram is null");
    return std::numeric_limits<double>::quiet_NaN();
  }
  return histogram->histogram.GetVariance();
}

double fastslide_histogram_get_std_dev(const FastSlideHistogram* histogram) {
  if (!histogram) {
    SetLastError("histogram is null");
    return std::numeric_limits<double>::quiet_NaN();
  }
  return histogram->histogram.GetStdDev();
}

double fastslide_histogram_get_sum(const FastSlideHistogram* histogram) {
  if (!histogram) {
    SetLastError("histogram is null");
    return std::numeric_limits<double>::quiet_NaN();
  }
  return histogram->histogram.GetSum();
}

int64_t fastslide_histogram_get_num_values(
    const FastSlideHistogram* histogram) {
  if (!histogram) {
    SetLastError("histogram is null");
    return -1;
  }
  return histogram->histogram.GetNValues();
}

int64_t fastslide_histogram_get_num_missing_values(
    const FastSlideHistogram* histogram) {
  if (!histogram) {
    SetLastError("histogram is null");
    return -1;
  }
  return histogram->histogram.GetNMissingValues();
}

int64_t fastslide_histogram_get_max_count(const FastSlideHistogram* histogram) {
  if (!histogram) {
    SetLastError("histogram is null");
    return -1;
  }
  return histogram->histogram.GetMaxCount();
}

double fastslide_histogram_get_max_normalized_count(
    const FastSlideHistogram* histogram) {
  if (!histogram) {
    SetLastError("histogram is null");
    return std::numeric_limits<double>::quiet_NaN();
  }
  return histogram->histogram.GetMaxNormalizedCount();
}

int64_t fastslide_histogram_get_count_sum(const FastSlideHistogram* histogram) {
  if (!histogram) {
    SetLastError("histogram is null");
    return -1;
  }
  return histogram->histogram.GetCountSum();
}

// Display range computation

int fastslide_histogram_compute_display_range(
    const FastSlideHistogram* histogram, double saturation,
    FastSlideDisplayRange* range) {
  if (!histogram || !range) {
    SetLastError("histogram and range cannot be null");
    return 0;
  }

  auto display_range = histogram->histogram.ComputeDisplayRange(saturation);
  range->min_value = display_range.first;
  range->max_value = display_range.second;

  return 1;
}

// Array access

int fastslide_histogram_get_edges(const FastSlideHistogram* histogram,
                                  double** edges, int* num_edges) {
  if (!histogram || !edges || !num_edges) {
    SetLastError("histogram, edges, and num_edges cannot be null");
    return 0;
  }

  const auto& edges_vec = histogram->histogram.GetEdges();
  *num_edges = static_cast<int>(edges_vec.size());

  if (*num_edges == 0) {
    *edges = nullptr;
    return 1;
  }

  *edges = static_cast<double*>(malloc(*num_edges * sizeof(double)));
  if (!*edges) {
    SetLastError("failed to allocate memory for edges array");
    return 0;
  }

  for (int i = 0; i < *num_edges; ++i) {
    (*edges)[i] = edges_vec[i];
  }

  return 1;
}

int fastslide_histogram_get_counts(const FastSlideHistogram* histogram,
                                   int64_t** counts, int* num_counts) {
  if (!histogram || !counts || !num_counts) {
    SetLastError("histogram, counts, and num_counts cannot be null");
    return 0;
  }

  const auto& counts_vec = histogram->histogram.GetCounts();
  *num_counts = static_cast<int>(counts_vec.size());

  if (*num_counts == 0) {
    *counts = nullptr;
    return 1;
  }

  *counts = static_cast<int64_t*>(malloc(*num_counts * sizeof(int64_t)));
  if (!*counts) {
    SetLastError("failed to allocate memory for counts array");
    return 0;
  }

  for (int i = 0; i < *num_counts; ++i) {
    (*counts)[i] = counts_vec[i];
  }

  return 1;
}

// Export functions

int fastslide_histogram_export_to_csv(const FastSlideHistogram* histogram,
                                      const char* filename) {
  if (!histogram || !filename) {
    SetLastError("histogram and filename cannot be null");
    return 0;
  }

  auto status = histogram->histogram.ExportToCSV(filename);
  if (!status.ok()) {
    SetLastError(std::string(status.message()).c_str());
    return 0;
  }

  return 1;
}

int fastslide_histogram_export_to_binary(const FastSlideHistogram* histogram,
                                         const char* filename) {
  if (!histogram || !filename) {
    SetLastError("histogram and filename cannot be null");
    return 0;
  }

  auto status = histogram->histogram.ExportToBinary(filename);
  if (!status.ok()) {
    SetLastError(std::string(status.message()).c_str());
    return 0;
  }

  return 1;
}

int fastslide_histogram_export_to_binary_buffer(
    const FastSlideHistogram* histogram, uint8_t** buffer,
    size_t* buffer_size) {
  if (!histogram || !buffer || !buffer_size) {
    SetLastError("histogram, buffer, and buffer_size cannot be null");
    return 0;
  }

  try {
    auto binary_data = histogram->histogram.ExportToBinaryVector();

    *buffer_size = binary_data.size();
    *buffer = static_cast<uint8_t*>(malloc(*buffer_size));

    if (!*buffer) {
      SetLastError("failed to allocate memory for binary buffer");
      return 0;
    }

    std::memcpy(*buffer, binary_data.data(), *buffer_size);
    return 1;

  } catch (const std::exception& e) {
    SetLastError(e.what());
    return 0;
  }
}

int fastslide_histogram_to_string(const FastSlideHistogram* histogram,
                                  char* buffer, size_t buffer_size) {
  if (!histogram || !buffer) {
    SetLastError("histogram and buffer cannot be null");
    return -1;
  }

  std::string str = histogram->histogram.ToString();
  size_t str_len = str.length();

  if (buffer_size <= str_len) {
    SetLastError("buffer size is too small");
    return -1;
  }

  std::strncpy(buffer, str.c_str(), buffer_size - 1);
  buffer[buffer_size - 1] = '\0';

  return static_cast<int>(str_len);
}

// Memory management

void fastslide_histogram_free(FastSlideHistogram* histogram) {
  delete histogram;
}

void fastslide_histogram_free_array(FastSlideHistogram** histograms,
                                    int num_histograms) {
  if (!histograms || num_histograms <= 0) {
    return;
  }

  for (int i = 0; i < num_histograms; ++i) {
    delete histograms[i];
  }
  free(histograms);
}

void fastslide_histogram_free_edges(double* edges) {
  free(edges);
}

void fastslide_histogram_free_counts(int64_t* counts) {
  free(counts);
}

void fastslide_histogram_free_binary_buffer(uint8_t* buffer) {
  free(buffer);
}
