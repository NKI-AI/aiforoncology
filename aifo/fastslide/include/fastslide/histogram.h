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

#ifndef AIFO_FASTSLIDE_INCLUDE_FASTSLIDE_HISTOGRAM_H_
#define AIFO_FASTSLIDE_INCLUDE_FASTSLIDE_HISTOGRAM_H_

#include <cmath>
#include <cstdint>
#include <limits>
#include <memory>
#include <span>
#include <string>
#include <string_view>
#include <utility>
#include <vector>

#include "absl/status/status.h"
#include "absl/status/statusor.h"
#include "fastslide/image.h"

namespace fastslide {

/**
 * @brief Running statistics calculator for computing mean, variance, etc. on the fly
 * 
 * This class implements Welford's online algorithm for computing running statistics
 * efficiently without storing all values in memory.
 */
class RunningStatistics {
 public:
  /// @brief Default constructor
  RunningStatistics() = default;

  /// @brief Add a value to the running statistics
  /// @param value The value to add
  void AddValue(double value);

  /// @brief Get the number of values processed
  /// @return Number of values
  [[nodiscard]] std::size_t GetSize() const noexcept { return count_; }

  /// @brief Get the number of NaN values encountered
  /// @return Number of NaN values
  [[nodiscard]] std::size_t GetNumNaNs() const noexcept { return nan_count_; }

  /// @brief Get the minimum value
  /// @return Minimum value or NaN if no values
  [[nodiscard]] double GetMin() const noexcept { return min_value_; }

  /// @brief Get the maximum value
  /// @return Maximum value or NaN if no values
  [[nodiscard]] double GetMax() const noexcept { return max_value_; }

  /// @brief Get the mean value
  /// @return Mean value or NaN if no values
  [[nodiscard]] double GetMean() const noexcept { return mean_; }

  /// @brief Get the variance
  /// @return Variance or NaN if fewer than 2 values
  [[nodiscard]] double GetVariance() const noexcept;

  /// @brief Get the standard deviation
  /// @return Standard deviation or NaN if fewer than 2 values
  [[nodiscard]] double GetStdDev() const noexcept;

  /// @brief Get the sum of all values
  /// @return Sum of all values
  [[nodiscard]] double GetSum() const noexcept { return sum_; }

  /// @brief Reset all statistics
  void Reset();

 private:
  std::size_t count_ = 0;      ///< Number of values processed
  std::size_t nan_count_ = 0;  ///< Number of NaN values encountered
  double min_value_ = std::numeric_limits<double>::max();
  double max_value_ = std::numeric_limits<double>::lowest();
  double mean_ = 0.0;  ///< Running mean
  double m2_ = 0.0;    ///< Running sum of squared differences from mean
  double sum_ = 0.0;   ///< Running sum
};

/**
 * @brief Type-erased wrapper for different array types
 * 
 * This allows the histogram to work with different data types (uint8, uint16, float, etc.)
 * in a uniform way, similar to QuPath's ArrayWrapper.
 */
class ArrayWrapper {
 public:
  /// @brief Virtual destructor
  virtual ~ArrayWrapper() = default;

  /// @brief Get the size of the array
  /// @return Number of elements
  [[nodiscard]] virtual std::size_t GetSize() const = 0;

  /// @brief Get value at index as double
  /// @param index Element index
  /// @return Value as double
  [[nodiscard]] virtual double GetDouble(std::size_t index) const = 0;

  /// @brief Check if this wrapper contains integer values
  /// @return True if all values are integers
  [[nodiscard]] virtual bool IsIntegerWrapper() const = 0;
};

/**
 * @brief Concrete implementation of ArrayWrapper for different types
 * @tparam T The underlying data type
 */
template <typename T>
class TypedArrayWrapper final : public ArrayWrapper {
 public:
  /// @brief Constructor from span
  /// @param data Span of data elements
  explicit TypedArrayWrapper(std::span<const T> data) : data_(data) {}

  /// @brief Get the size of the array
  /// @return Number of elements
  [[nodiscard]] std::size_t GetSize() const override { return data_.size(); }

  /// @brief Get value at index as double
  /// @param index Element index
  /// @return Value as double
  [[nodiscard]] double GetDouble(std::size_t index) const override {
    return static_cast<double>(data_[index]);
  }

  /// @brief Check if this wrapper contains integer values
  /// @return True if T is an integer type
  [[nodiscard]] bool IsIntegerWrapper() const override {
    return std::is_integral_v<T>;
  }

 private:
  std::span<const T> data_;  ///< View into the data
};

/**
 * @brief Factory functions for creating ArrayWrapper instances
 */
namespace array_wrappers {

/// @brief Create wrapper for uint8 data
/// @param data Span of uint8 data
/// @return Unique pointer to ArrayWrapper
[[nodiscard]] std::unique_ptr<ArrayWrapper> MakeUInt8ArrayWrapper(
    std::span<const uint8_t> data);

/// @brief Create wrapper for uint16 data
/// @param data Span of uint16 data
/// @return Unique pointer to ArrayWrapper
[[nodiscard]] std::unique_ptr<ArrayWrapper> MakeUInt16ArrayWrapper(
    std::span<const uint16_t> data);

/// @brief Create wrapper for float data
/// @param data Span of float data
/// @return Unique pointer to ArrayWrapper
[[nodiscard]] std::unique_ptr<ArrayWrapper> MakeFloatArrayWrapper(
    std::span<const float> data);

/// @brief Create wrapper for double data
/// @param data Span of double data
/// @return Unique pointer to ArrayWrapper
[[nodiscard]] std::unique_ptr<ArrayWrapper> MakeDoubleArrayWrapper(
    std::span<const double> data);

}  // namespace array_wrappers

/**
 * @brief Histogram class that matches QuPath's histogram algorithm
 * 
 * This class computes histograms from arrays of values with the same algorithm
 * used by QuPath, including special handling for integer values and automatic
 * bin size adjustment.
 */
class Histogram {
 public:
  /// @brief Default constructor for empty histogram
  Histogram() = default;

  /// @brief Constructor from double array
  /// @param values Array of values
  /// @param n_bins Number of bins (default: 1024)
  explicit Histogram(std::span<const double> values, int n_bins = 1024);

  /// @brief Constructor from float array
  /// @param values Array of values
  /// @param n_bins Number of bins (default: 1024)
  explicit Histogram(std::span<const float> values, int n_bins = 1024);

  /// @brief Constructor from uint8 array
  /// @param values Array of values
  /// @param n_bins Number of bins (default: 1024)
  explicit Histogram(std::span<const uint8_t> values, int n_bins = 1024);

  /// @brief Constructor from uint16 array
  /// @param values Array of values
  /// @param n_bins Number of bins (default: 1024)
  explicit Histogram(std::span<const uint16_t> values, int n_bins = 1024);

  /// @brief Constructor from ArrayWrapper (generic)
  /// @param values Array wrapper
  /// @param n_bins Number of bins (default: 1024)
  explicit Histogram(const ArrayWrapper& values, int n_bins = 1024);

  /// @brief Copy constructor
  Histogram(const Histogram& other);

  /// @brief Move constructor
  Histogram(Histogram&& other) noexcept = default;

  /// @brief Copy assignment
  Histogram& operator=(const Histogram& other);

  /// @brief Move assignment
  Histogram& operator=(Histogram&& other) noexcept = default;

  /// @brief Destructor
  ~Histogram() = default;

  // Accessors

  /// @brief Get the minimum edge of the histogram
  /// @return Minimum edge
  [[nodiscard]] double GetEdgeMin() const noexcept { return edge_min_; }

  /// @brief Get the maximum edge of the histogram
  /// @return Maximum edge
  [[nodiscard]] double GetEdgeMax() const noexcept { return edge_max_; }

  /// @brief Get the histogram edge range
  /// @return Edge range (max - min)
  [[nodiscard]] double GetEdgeRange() const noexcept {
    return GetEdgeMax() - GetEdgeMin();
  }

  /// @brief Get the lower edge for a specified bin
  /// @param bin_index Index of the bin
  /// @return Lower edge value
  [[nodiscard]] double GetBinLeftEdge(int bin_index) const;

  /// @brief Get the upper edge for a specified bin
  /// @param bin_index Index of the bin
  /// @return Upper edge value
  [[nodiscard]] double GetBinRightEdge(int bin_index) const;

  /// @brief Get the center of a bin
  /// @param bin_index Index of the bin
  /// @return Center value
  [[nodiscard]] double GetBinCenter(int bin_index) const;

  /// @brief Get the width of a bin
  /// @param bin_index Index of the bin
  /// @return Bin width
  [[nodiscard]] double GetBinWidth(int bin_index) const;

  /// @brief Get the count for a specific bin
  /// @param bin_index Index of the bin
  /// @return Count for the bin
  [[nodiscard]] std::int64_t GetCountsForBin(int bin_index) const;

  /// @brief Get the normalized count for a specific bin
  /// @param bin_index Index of the bin
  /// @return Normalized count (count / total_count)
  [[nodiscard]] double GetNormalizedCountsForBin(int bin_index) const;

  /// @brief Check if histogram was created from integer values only
  /// @return True if all values were integers
  [[nodiscard]] bool IsInteger() const noexcept { return is_integer_; }

  /// @brief Get the minimum of all input values
  /// @return Minimum value
  [[nodiscard]] double GetMinValue() const noexcept;

  /// @brief Get the maximum of all input values
  /// @return Maximum value
  [[nodiscard]] double GetMaxValue() const noexcept;

  /// @brief Get the mean of all input values
  /// @return Mean value
  [[nodiscard]] double GetMeanValue() const noexcept;

  /// @brief Get the variance of all input values
  /// @return Variance
  [[nodiscard]] double GetVariance() const noexcept;

  /// @brief Get the standard deviation of all input values
  /// @return Standard deviation
  [[nodiscard]] double GetStdDev() const noexcept;

  /// @brief Get the sum of all input values
  /// @return Sum of values
  [[nodiscard]] double GetSum() const noexcept;

  /// @brief Get the number of values represented in the histogram
  /// @return Number of values
  [[nodiscard]] std::int64_t GetNValues() const noexcept;

  /// @brief Get the number of NaN values in the input
  /// @return Number of NaN values
  [[nodiscard]] std::int64_t GetNMissingValues() const noexcept;

  /// @brief Get the bin index for a value (general method)
  /// @param value The value to find
  /// @return Bin index or -1 if out of range
  [[nodiscard]] int GetBinIndexForValue(double value) const;

  /// @brief Get the bin index for a value (optimized for uniform bins)
  /// @param value The value to find
  /// @param bin_width Width of each bin (assumed uniform)
  /// @return Bin index
  [[nodiscard]] int GetBinIndexForValue(double value, double bin_width) const;

  /// @brief Get the maximum count in any bin
  /// @return Maximum count
  [[nodiscard]] std::int64_t GetMaxCount() const noexcept { return max_count_; }

  /// @brief Get the maximum normalized count
  /// @return Maximum normalized count
  [[nodiscard]] double GetMaxNormalizedCount() const noexcept;

  /// @brief Get the number of bins
  /// @return Number of bins
  [[nodiscard]] int GetNBins() const noexcept {
    return static_cast<int>(counts_.size());
  }

  /// @brief Get the sum of all counts
  /// @return Total count
  [[nodiscard]] std::int64_t GetCountSum() const noexcept { return count_sum_; }

  /// @brief Get all bin edges
  /// @return Vector of bin edges (size = n_bins + 1)
  [[nodiscard]] const std::vector<double>& GetEdges() const noexcept {
    return edges_;
  }

  /// @brief Get all bin counts
  /// @return Vector of bin counts
  [[nodiscard]] const std::vector<std::int64_t>& GetCounts() const noexcept {
    return counts_;
  }

  // QuPath compatibility methods

  /// @brief Compute QuPath-style display range using saturation-based algorithm
  /// @param saturation Saturation percentage (default: 0.001 = 0.1%)
  /// @return Pair of (min_display, max_display) values
  [[nodiscard]] std::pair<double, double> ComputeDisplayRange(
      double saturation = 0.001) const;

  // Export functionality

  /// @brief Export histogram data to CSV format
  /// @param filename Output filename
  /// @return Status of the operation
  [[nodiscard]] absl::Status ExportToCSV(std::string_view filename) const;

  /// @brief Export histogram data to CSV stream
  /// @param stream Output stream
  /// @return Status of the operation
  [[nodiscard]] absl::Status ExportToCSV(std::ostream& stream) const;

  /// @brief Export histogram data to binary format
  /// @param filename Output filename
  /// @return Status of the operation
  [[nodiscard]] absl::Status ExportToBinary(std::string_view filename) const;

  /// @brief Export histogram data to binary stream
  /// @param stream Output stream
  /// @return Status of the operation
  [[nodiscard]] absl::Status ExportToBinary(std::ostream& stream) const;

  /// @brief Export histogram data to binary vector
  /// @return Vector containing binary data
  [[nodiscard]] std::vector<uint8_t> ExportToBinaryVector() const;

  /// @brief Get string representation
  /// @return String description
  [[nodiscard]] std::string ToString() const;

 private:
  std::vector<double> edges_;         ///< Bin edges (size = n_bins + 1)
  std::vector<std::int64_t> counts_;  ///< Bin counts
  std::int64_t max_count_ = 0;        ///< Maximum count in any bin
  double edge_min_ = 0.0;             ///< Minimum edge
  double edge_max_ = 0.0;             ///< Maximum edge
  std::int64_t count_sum_ = 0;        ///< Sum of all counts
  bool is_integer_ = true;            ///< Whether input was all integers
  std::unique_ptr<RunningStatistics> stats_;  ///< Statistics calculator

  /// @brief Build histogram from array wrapper
  /// @param values Array wrapper
  /// @param n_bins Number of bins
  void BuildHistogram(const ArrayWrapper& values, int n_bins);
};

// Image integration functions

/// @brief Create histogram from a single channel of an Image
/// @param image Input image
/// @param channel Channel index
/// @param n_bins Number of bins (default: 1024)
/// @return Histogram or error status
[[nodiscard]] absl::StatusOr<Histogram> CreateHistogramFromImageChannel(
    const Image& image, uint32_t channel, int n_bins = 1024);

/// @brief Create histogram from entire Image (all channels combined)
/// @param image Input image
/// @param n_bins Number of bins (default: 1024)
/// @return Histogram or error status
[[nodiscard]] absl::StatusOr<Histogram> CreateHistogramFromImage(
    const Image& image, int n_bins = 1024);

/// @brief Create histograms for each channel of an Image
/// @param image Input image
/// @param n_bins Number of bins (default: 1024)
/// @return Vector of histograms (one per channel) or error status
[[nodiscard]] absl::StatusOr<std::vector<Histogram>>
CreateHistogramsFromImageChannels(const Image& image, int n_bins = 1024);

}  // namespace fastslide

#endif  // AIFO_FASTSLIDE_INCLUDE_FASTSLIDE_HISTOGRAM_H_
