// Copyright 2025 Jonas Teuwen. All Rights Reserved.
//
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

#include "fastslide/histogram.h"

#include <algorithm>
#include <cmath>
#include <cstring>
#include <fstream>
#include <iomanip>
#include <limits>
#include <memory>
#include <stdexcept>
#include <string>
#include <utility>
#include <vector>

#include "aifocore/status/status_macros.h"
#include "aifocore/utilities/fmt.h"

namespace fastslide {

using aifocore::fmt::format;

// RunningStatistics implementation

void RunningStatistics::AddValue(double value) {
  if (std::isnan(value)) {
    ++nan_count_;
    return;
  }

  ++count_;
  sum_ += value;

  // Update min/max
  min_value_ = std::min(min_value_, value);
  max_value_ = std::max(max_value_, value);

  // Welford's online algorithm for mean and variance
  double delta = value - mean_;
  mean_ += delta / static_cast<double>(count_);
  double delta2 = value - mean_;
  m2_ += delta * delta2;
}

double RunningStatistics::GetVariance() const noexcept {
  if (count_ < 2) {
    return std::numeric_limits<double>::quiet_NaN();
  }
  return m2_ / static_cast<double>(count_ - 1);
}

double RunningStatistics::GetStdDev() const noexcept {
  return std::sqrt(GetVariance());
}

void RunningStatistics::Reset() {
  count_ = 0;
  nan_count_ = 0;
  min_value_ = std::numeric_limits<double>::max();
  max_value_ = std::numeric_limits<double>::lowest();
  mean_ = 0.0;
  m2_ = 0.0;
  sum_ = 0.0;
}

// ArrayWrapper factory functions

namespace array_wrappers {

std::unique_ptr<ArrayWrapper> MakeUInt8ArrayWrapper(
    std::span<const uint8_t> data) {
  return std::make_unique<TypedArrayWrapper<uint8_t>>(data);
}

std::unique_ptr<ArrayWrapper> MakeUInt16ArrayWrapper(
    std::span<const uint16_t> data) {
  return std::make_unique<TypedArrayWrapper<uint16_t>>(data);
}

std::unique_ptr<ArrayWrapper> MakeFloatArrayWrapper(
    std::span<const float> data) {
  return std::make_unique<TypedArrayWrapper<float>>(data);
}

std::unique_ptr<ArrayWrapper> MakeDoubleArrayWrapper(
    std::span<const double> data) {
  return std::make_unique<TypedArrayWrapper<double>>(data);
}

}  // namespace array_wrappers

// Histogram implementation

Histogram::Histogram(std::span<const double> values, int n_bins) {
  auto wrapper = array_wrappers::MakeDoubleArrayWrapper(values);
  BuildHistogram(*wrapper, n_bins);
}

Histogram::Histogram(std::span<const float> values, int n_bins) {
  auto wrapper = array_wrappers::MakeFloatArrayWrapper(values);
  BuildHistogram(*wrapper, n_bins);
}

Histogram::Histogram(std::span<const uint8_t> values, int n_bins) {
  auto wrapper = array_wrappers::MakeUInt8ArrayWrapper(values);
  BuildHistogram(*wrapper, n_bins);
}

Histogram::Histogram(std::span<const uint16_t> values, int n_bins) {
  auto wrapper = array_wrappers::MakeUInt16ArrayWrapper(values);
  BuildHistogram(*wrapper, n_bins);
}

Histogram::Histogram(const ArrayWrapper& values, int n_bins) {
  BuildHistogram(values, n_bins);
}

Histogram::Histogram(const Histogram& other)
    : edges_(other.edges_),
      counts_(other.counts_),
      max_count_(other.max_count_),
      edge_min_(other.edge_min_),
      edge_max_(other.edge_max_),
      count_sum_(other.count_sum_),
      is_integer_(other.is_integer_) {
  // Deep copy the statistics
  if (other.stats_) {
    stats_ = std::make_unique<RunningStatistics>(*other.stats_);
  }
}

Histogram& Histogram::operator=(const Histogram& other) {
  if (this != &other) {
    edges_ = other.edges_;
    counts_ = other.counts_;
    max_count_ = other.max_count_;
    edge_min_ = other.edge_min_;
    edge_max_ = other.edge_max_;
    count_sum_ = other.count_sum_;
    is_integer_ = other.is_integer_;

    // Deep copy the statistics
    if (other.stats_) {
      stats_ = std::make_unique<RunningStatistics>(*other.stats_);
    } else {
      stats_.reset();
    }
  }
  return *this;
}

double Histogram::GetBinLeftEdge(int bin_index) const {
  if (bin_index < 0 || bin_index >= static_cast<int>(edges_.size() - 1)) {
    throw std::out_of_range("Bin index out of range");
  }
  return edges_[bin_index];
}

double Histogram::GetBinRightEdge(int bin_index) const {
  if (bin_index < 0 || bin_index >= static_cast<int>(edges_.size() - 1)) {
    throw std::out_of_range("Bin index out of range");
  }
  return edges_[bin_index + 1];
}

double Histogram::GetBinCenter(int bin_index) const {
  return (GetBinLeftEdge(bin_index) + GetBinRightEdge(bin_index)) / 2.0;
}

double Histogram::GetBinWidth(int bin_index) const {
  return GetBinRightEdge(bin_index) - GetBinLeftEdge(bin_index);
}

std::int64_t Histogram::GetCountsForBin(int bin_index) const {
  if (bin_index < 0 || bin_index >= static_cast<int>(counts_.size())) {
    throw std::out_of_range("Bin index out of range");
  }
  return counts_[bin_index];
}

double Histogram::GetNormalizedCountsForBin(int bin_index) const {
  if (count_sum_ == 0) {
    return 0.0;
  }
  return static_cast<double>(GetCountsForBin(bin_index)) /
         static_cast<double>(count_sum_);
}

double Histogram::GetMinValue() const noexcept {
  return stats_ ? stats_->GetMin() : std::numeric_limits<double>::quiet_NaN();
}

double Histogram::GetMaxValue() const noexcept {
  return stats_ ? stats_->GetMax() : std::numeric_limits<double>::quiet_NaN();
}

double Histogram::GetMeanValue() const noexcept {
  return stats_ ? stats_->GetMean() : std::numeric_limits<double>::quiet_NaN();
}

double Histogram::GetVariance() const noexcept {
  return stats_ ? stats_->GetVariance()
                : std::numeric_limits<double>::quiet_NaN();
}

double Histogram::GetStdDev() const noexcept {
  return stats_ ? stats_->GetStdDev()
                : std::numeric_limits<double>::quiet_NaN();
}

double Histogram::GetSum() const noexcept {
  return stats_ ? stats_->GetSum() : std::numeric_limits<double>::quiet_NaN();
}

std::int64_t Histogram::GetNValues() const noexcept {
  return stats_ ? static_cast<std::int64_t>(stats_->GetSize()) : -1;
}

std::int64_t Histogram::GetNMissingValues() const noexcept {
  return stats_ ? static_cast<std::int64_t>(stats_->GetNumNaNs()) : 0;
}

int Histogram::GetBinIndexForValue(double value) const {
  // Return -1 if out of range
  if (value > edge_max_ || value < edge_min_) {
    return -1;
  }

  int i = static_cast<int>(edges_.size()) - 2;
  while (i >= 0) {
    if (edges_[i] <= value) {
      return i;
    }
    --i;
  }
  return i;
}

int Histogram::GetBinIndexForValue(double value, double bin_width) const {
  int bin = static_cast<int>((value - edge_min_) / bin_width);
  if (bin >= static_cast<int>(counts_.size())) {
    bin = static_cast<int>(counts_.size()) - 1;
  }
  return bin;
}

double Histogram::GetMaxNormalizedCount() const noexcept {
  if (count_sum_ == 0) {
    return 0.0;
  }
  return static_cast<double>(GetMaxCount()) / static_cast<double>(count_sum_);
}

absl::Status Histogram::ExportToCSV(std::string_view filename) const {
  std::ofstream file{std::string(filename)};
  if (!file.is_open()) {
    return MAKE_STATUS(absl::StatusCode::kInvalidArgument,
                       format("Failed to open file: {}", filename));
  }

  return ExportToCSV(file);
}

absl::Status Histogram::ExportToCSV(std::ostream& stream) const {
  // Write header
  stream
      << "bin_index,left_edge,right_edge,center,width,count,normalized_count\n";

  // Write data
  for (int i = 0; i < GetNBins(); ++i) {
    stream << i << "," << std::fixed << std::setprecision(6)
           << GetBinLeftEdge(i) << "," << GetBinRightEdge(i) << ","
           << GetBinCenter(i) << "," << GetBinWidth(i) << ","
           << GetCountsForBin(i) << "," << GetNormalizedCountsForBin(i) << "\n";
  }

  if (stream.fail()) {
    return MAKE_STATUS(absl::StatusCode::kInternal,
                       "Failed to write histogram data to stream");
  }

  return absl::OkStatus();
}

absl::Status Histogram::ExportToBinary(std::string_view filename) const {
  std::ofstream file{std::string(filename), std::ios::binary};
  if (!file.is_open()) {
    return MAKE_STATUS(absl::StatusCode::kInvalidArgument,
                       format("Failed to open file: {}", filename));
  }

  return ExportToBinary(file);
}

absl::Status Histogram::ExportToBinary(std::ostream& stream) const {
  auto binary_data = ExportToBinaryVector();

  stream.write(reinterpret_cast<const char*>(binary_data.data()),
               binary_data.size());

  if (stream.fail()) {
    return MAKE_STATUS(absl::StatusCode::kInternal,
                       "Failed to write binary histogram data to stream");
  }

  return absl::OkStatus();
}

std::vector<uint8_t> Histogram::ExportToBinaryVector() const {
  // Binary format:
  // - Magic header: "HIST" (4 bytes)
  // - Version: uint32 (4 bytes)
  // - Number of bins: uint32 (4 bytes)
  // - Bin edges: array of double (8 * (n_bins + 1) bytes)
  // - Bin counts: array of int64 (8 * n_bins bytes)
  // - Statistics: min, max, mean, stddev, sum, n_values, n_missing (7 * 8 = 56
  // bytes)
  // All data is little-endian for cross-platform compatibility

  const uint32_t version = 1;
  const uint32_t n_bins = static_cast<uint32_t>(GetNBins());

  // Calculate total size
  const size_t header_size = 4 + 4 + 4;  // magic + version + n_bins
  const size_t edges_size = (n_bins + 1) * sizeof(double);
  const size_t counts_size = n_bins * sizeof(std::int64_t);
  const size_t stats_size = 7 * sizeof(double);
  const size_t total_size = header_size + edges_size + counts_size + stats_size;

  std::vector<uint8_t> data;
  data.reserve(total_size);

  auto write_bytes = [&data](const void* src, size_t size) {
    const uint8_t* bytes = static_cast<const uint8_t*>(src);
    data.insert(data.end(), bytes, bytes + size);
  };

  auto write_le_uint32 = [&write_bytes](uint32_t value) {
    uint32_t le_value = value;  // Assume little-endian host for now
    write_bytes(&le_value, sizeof(le_value));
  };

  auto write_le_double = [&write_bytes](double value) {
    write_bytes(&value, sizeof(value));
  };

  auto write_le_int64 = [&write_bytes](std::int64_t value) {
    write_bytes(&value, sizeof(value));
  };

  // Write magic header
  const char magic[4] = {'H', 'I', 'S', 'T'};
  write_bytes(magic, 4);

  // Write version
  write_le_uint32(version);

  // Write number of bins
  write_le_uint32(n_bins);

  // Write bin edges
  for (const auto& edge : edges_) {
    write_le_double(edge);
  }

  // Write bin counts
  for (const auto& count : counts_) {
    write_le_int64(count);
  }

  // Write statistics
  write_le_double(GetMinValue());
  write_le_double(GetMaxValue());
  write_le_double(GetMeanValue());
  write_le_double(GetStdDev());
  write_le_double(GetSum());
  write_le_double(static_cast<double>(GetNValues()));
  write_le_double(static_cast<double>(GetNMissingValues()));

  return data;
}

std::string Histogram::ToString() const {
  return format("Histogram: Min {:.2f}, Max {:.2f}, Total count: {}, N Bins {}",
                GetEdgeMin(), GetEdgeMax(), GetCountSum(), GetNBins());
}

std::pair<double, double> Histogram::ComputeDisplayRange(
    double saturation) const {
  // For unsupported saturation values, just return min/max
  if (saturation <= 0.0 || saturation >= 1.0) {
    return {GetEdgeMin(), GetEdgeMax()};
  }

  // Get initial values
  std::int64_t count_sum = GetCountSum();
  int n_bins = GetNBins();
  int ind = 0;

  // Possibly skip the first and/or last bins; these can often represent
  // unscanned/clipped regions
  if (n_bins > 2) {
    std::int64_t first_count = GetCountsForBin(0);
    if (first_count > GetCountsForBin(1)) {
      count_sum -= first_count;
      ind = 1;
    }

    std::int64_t last_count = GetCountsForBin(n_bins - 1);
    if (last_count > GetCountsForBin(n_bins - 2)) {
      count_sum -= last_count;
      n_bins -= 1;
    }
  }

  double count_max = static_cast<double>(count_sum) * saturation;

  // Compute minDisplay (scan from left)
  double count = count_max;
  double min_display = GetEdgeMin();  // histogram edge min

  while (ind < GetNBins()) {
    double next_count = static_cast<double>(GetCountsForBin(ind));
    if (count < next_count) {
      // Interpolate within this bin
      double bin_left_edge = GetBinLeftEdge(ind);
      double bin_width = GetBinWidth(ind);
      min_display = bin_left_edge + (count / next_count) * bin_width;
      break;
    }
    count -= next_count;
    ind++;
  }

  // Compute maxDisplay (scan from right)
  count = count_max;
  double max_display = GetEdgeMax();  // histogram edge max
  ind = GetNBins() - 1;

  while (ind >= 0) {
    double next_count = static_cast<double>(GetCountsForBin(ind));
    if (count < next_count) {
      // Interpolate within this bin
      double bin_right_edge = GetBinRightEdge(ind);
      double bin_width = GetBinWidth(ind);
      max_display = bin_right_edge - (count / next_count) * bin_width;
      break;
    }
    count -= next_count;
    ind--;
  }

  return {min_display, max_display};
}

void Histogram::BuildHistogram(const ArrayWrapper& values, int n_bins) {
  // Initialize statistics
  stats_ = std::make_unique<RunningStatistics>();

  // Compute running statistics and check for integer values
  is_integer_ = values.IsIntegerWrapper();
  bool maybe_integer = !is_integer_;

  const std::size_t n = values.GetSize();
  for (std::size_t i = 0; i < n; ++i) {
    double v = values.GetDouble(i);
    stats_->AddValue(v);

    // Check if we have integers only
    if (maybe_integer && v != static_cast<int>(v)) {
      maybe_integer = false;
    }
  }

  if (!is_integer_) {
    is_integer_ = maybe_integer;
  }

  // Always use the full data range from statistics
  double stats_min = stats_->GetMin();
  double stats_max = stats_->GetMax();

  if (stats_->GetSize() > 0 && std::isfinite(stats_min) &&
      stats_min != std::numeric_limits<double>::max()) {
    edge_min_ = stats_min;
  } else {
    edge_min_ = 0.0;  // Fallback for edge case
  }

  if (stats_->GetSize() > 0 && std::isfinite(stats_max) &&
      stats_max != std::numeric_limits<double>::lowest()) {
    edge_max_ = stats_max;
  } else {
    edge_max_ = 255.0;  // Fallback for common image data range
  }

  // If min and max are the same, expand the range slightly
  if (std::abs(edge_max_ - edge_min_) <
      std::numeric_limits<double>::epsilon()) {
    if (edge_min_ == 0.0) {
      edge_max_ = 1.0;
    } else {
      double mid = (edge_min_ + edge_max_) / 2.0;
      double delta = std::abs(mid) * 0.1;
      if (delta == 0.0)
        delta = 1.0;
      edge_min_ = mid - delta;
      edge_max_ = mid + delta;
    }
  }

  // Compute bin width
  double bin_width = (edge_max_ - edge_min_) / static_cast<double>(n_bins);

  // If we have integer values, don't set the bin width to be < 1
  if (!std::isfinite(bin_width)) {
    n_bins = 0;
  } else if (bin_width < 1.0 && is_integer_) {
    bin_width = 1.0;
    n_bins = static_cast<int>(edge_max_ - edge_min_ + 1);
  }

  // Create arrays
  edges_.resize(n_bins + 1);
  counts_.resize(n_bins, 0);

  if (n_bins == 0) {
    return;
  }

  // Fill in edges
  for (int i = 0; i <= n_bins; ++i) {
    edges_[i] = edge_min_ + static_cast<double>(i) * bin_width;
  }

  // Compute counts
  max_count_ = 0;
  count_sum_ = 0;

  for (std::size_t i = 0; i < n; ++i) {
    double v = values.GetDouble(i);

    // Skip NaNs or out of range values
    if (std::isnan(v) || v < edge_min_ || v > edge_max_) {
      continue;
    }

    int bin = GetBinIndexForValue(v, bin_width);
    if (bin >= 0 && bin < static_cast<int>(counts_.size())) {
      std::int64_t count = counts_[bin] + 1;
      counts_[bin] = count;
      if (count > max_count_) {
        max_count_ = count;
      }
      ++count_sum_;
    }
  }
}

// Image integration functions

absl::StatusOr<Histogram> CreateHistogramFromImageChannel(const Image& image,
                                                          uint32_t channel,
                                                          int n_bins) {
  if (image.Empty()) {
    return MAKE_STATUS(absl::StatusCode::kInvalidArgument,
                       "Cannot create histogram from empty image");
  }

  if (channel >= image.GetChannels()) {
    return MAKE_STATUS(
        absl::StatusCode::kInvalidArgument,
        format("Channel {} is out of range (image has {} channels)", channel,
               image.GetChannels()));
  }

  // Extract channel data based on image data type
  const std::size_t pixel_count = image.GetPixelCount();
  std::vector<double> channel_data;
  channel_data.reserve(pixel_count);

  // Extract channel data using Image::At method for consistency
  for (uint32_t y = 0; y < image.GetHeight(); ++y) {
    for (uint32_t x = 0; x < image.GetWidth(); ++x) {
      double value;
      switch (image.GetDataType()) {
        case DataType::kUInt8:
          value = static_cast<double>(image.At<uint8_t>(y, x, channel));
          break;
        case DataType::kUInt16:
          value = static_cast<double>(image.At<uint16_t>(y, x, channel));
          break;
        case DataType::kFloat32:
          value = static_cast<double>(image.At<float>(y, x, channel));
          break;
        case DataType::kFloat64:
          value = image.At<double>(y, x, channel);
          break;
        default:
          return MAKE_STATUS(absl::StatusCode::kUnimplemented,
                             format("Unsupported data type for histogram: {}",
                                    GetDataTypeName(image.GetDataType())));
      }
      channel_data.push_back(value);
    }
  }

  return Histogram(std::span<const double>(channel_data), n_bins);
}

absl::StatusOr<Histogram> CreateHistogramFromImage(const Image& image,
                                                   int n_bins) {
  if (image.Empty()) {
    return MAKE_STATUS(absl::StatusCode::kInvalidArgument,
                       "Cannot create histogram from empty image");
  }

  // Extract all pixel data
  const std::size_t total_elements =
      image.GetPixelCount() * image.GetChannels();
  std::vector<double> all_data;
  all_data.reserve(total_elements);

  // Extract all data efficiently based on data type
  switch (image.GetDataType()) {
    case DataType::kUInt8: {
      const uint8_t* data = image.GetDataAs<uint8_t>();
      for (std::size_t i = 0; i < total_elements; ++i) {
        all_data.push_back(static_cast<double>(data[i]));
      }
      break;
    }
    case DataType::kUInt16: {
      const uint16_t* data = image.GetDataAs<uint16_t>();
      for (std::size_t i = 0; i < total_elements; ++i) {
        all_data.push_back(static_cast<double>(data[i]));
      }
      break;
    }
    case DataType::kFloat32: {
      const float* data = image.GetDataAs<float>();
      for (std::size_t i = 0; i < total_elements; ++i) {
        all_data.push_back(static_cast<double>(data[i]));
      }
      break;
    }
    case DataType::kFloat64: {
      const double* data = image.GetDataAs<double>();
      for (std::size_t i = 0; i < total_elements; ++i) {
        all_data.push_back(data[i]);
      }
      break;
    }
    default:
      return MAKE_STATUS(absl::StatusCode::kUnimplemented,
                         format("Unsupported data type for histogram: {}",
                                GetDataTypeName(image.GetDataType())));
  }

  return Histogram(std::span<const double>(all_data), n_bins);
}

absl::StatusOr<std::vector<Histogram>> CreateHistogramsFromImageChannels(
    const Image& image, int n_bins) {
  if (image.Empty()) {
    return MAKE_STATUS(absl::StatusCode::kInvalidArgument,
                       "Cannot create histograms from empty image");
  }

  std::vector<Histogram> histograms;
  histograms.reserve(image.GetChannels());

  for (uint32_t channel = 0; channel < image.GetChannels(); ++channel) {
    Histogram histogram;
    ASSIGN_OR_RETURN(histogram,
                     CreateHistogramFromImageChannel(image, channel, n_bins),
                     "Failed to create histogram for channel");
    histograms.push_back(std::move(histogram));
  }

  return histograms;
}

}  // namespace fastslide
