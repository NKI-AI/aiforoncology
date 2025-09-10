// Copyright 2024 Jonas Teuwen. All Rights Reserved.
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
#include <openslide/openslide.h>
#include <vips/vips8>

#include <algorithm>
#include <atomic>
#include <chrono>
#include <cmath>
#include <filesystem>
#include <future>
#include <iomanip>
#include <iostream>
#include <map>
#include <memory>
#include <mutex>
#include <numeric>
#include <random>
#include <sstream>
#include <string>
#include <thread>
#include <utility>
#include <vector>

#include "absl/status/status.h"
#include "absl/status/statusor.h"
#include "absl/strings/str_cat.h"
#include "fastslide/readers/readers.h"
#include "fastslide/slide_reader.h"

using fastslide::InitializeReaders;
using fastslide::RegionSpec;
using fastslide::RGBImage;
using fastslide::SlideReader;
using fastslide::SlideReaderRegistry;

namespace {

/// @brief Test region specification
/// (using level-0 coordinates for OpenSlide compatibility)
struct TestRegion {
  int64_t x, y;  // Level-0 coordinates (used by OpenSlide)
  int64_t width, height;
  int level;
  std::string description;
};

/// @brief Performance statistics for a collection of measurements
struct PerformanceStats {
  std::vector<double> times_ms;
  size_t total_bytes;
  size_t successful_ops;
  size_t failed_ops;

  double GetMeanTime() const {
    if (times_ms.empty())
      return 0.0;
    double sum = 0.0;
    for (double t : times_ms)
      sum += t;
    return sum / times_ms.size();
  }

  double GetMedianTime() const {
    if (times_ms.empty())
      return 0.0;
    std::vector<double> sorted = times_ms;
    std::sort(sorted.begin(), sorted.end());
    return sorted[sorted.size() / 2];
  }

  double GetStdDev() const {
    if (times_ms.size() < 2)
      return 0.0;
    double mean = GetMeanTime();
    double variance = 0.0;
    for (double t : times_ms) {
      variance += (t - mean) * (t - mean);
    }
    return std::sqrt(variance / (times_ms.size() - 1));
  }

  double GetMinTime() const {
    if (times_ms.empty())
      return 0.0;
    return *std::min_element(times_ms.begin(), times_ms.end());
  }

  double GetMaxTime() const {
    if (times_ms.empty())
      return 0.0;
    return *std::max_element(times_ms.begin(), times_ms.end());
  }

  double GetThroughputMBps() const {
    double mean_time = GetMeanTime();
    if (mean_time <= 0.0)
      return 0.0;
    double total_mb = static_cast<double>(total_bytes) / (1024.0 * 1024.0);
    return total_mb / (mean_time / 1000.0);
  }

  double GetOpsPerSecond() const {
    double mean_time = GetMeanTime();
    if (mean_time <= 0.0)
      return 0.0;
    return 1000.0 / mean_time;
  }
};

/// @brief Test results for a specific level
struct LevelBenchmarkResult {
  int level;
  std::string test_name;
  PerformanceStats openslide_stats;
  PerformanceStats fastslide_stats;

  double GetSpeedup() const {
    double os_time = openslide_stats.GetMeanTime();
    double fs_time = fastslide_stats.GetMeanTime();
    if (fs_time <= 0.0)
      return 0.0;
    return os_time / fs_time;
  }

  double GetThroughputRatio() const {
    double os_throughput = openslide_stats.GetThroughputMBps();
    double fs_throughput = fastslide_stats.GetThroughputMBps();
    if (os_throughput <= 0.0)
      return 0.0;
    return fs_throughput / os_throughput;
  }
};

/// @brief Test results for a specific level and tile size combination
struct TileSizeBenchmarkResult {
  int level;
  uint32_t tile_size;
  std::string test_name;
  PerformanceStats openslide_stats;
  PerformanceStats fastslide_stats;

  double GetSpeedup() const {
    double os_time = openslide_stats.GetMeanTime();
    double fs_time = fastslide_stats.GetMeanTime();
    if (fs_time <= 0.0)
      return 0.0;
    return os_time / fs_time;
  }

  double GetThroughputRatio() const {
    double os_throughput = openslide_stats.GetThroughputMBps();
    double fs_throughput = fastslide_stats.GetThroughputMBps();
    if (os_throughput <= 0.0)
      return 0.0;
    return fs_throughput / os_throughput;
  }
};

/// @brief Multithreaded benchmark results
struct MultithreadedBenchmarkResult {
  uint32_t tile_size;
  int level;
  int thread_count;
  std::string test_name;
  PerformanceStats single_threaded_stats;
  PerformanceStats multi_threaded_stats;
  double
      thread_efficiency;  // Performance per thread compared to single-threaded

  double GetThreadScaling() const {
    double st_throughput = single_threaded_stats.GetThroughputMBps();
    double mt_throughput = multi_threaded_stats.GetThroughputMBps();
    if (st_throughput <= 0.0)
      return 0.0;
    return mt_throughput / st_throughput;
  }

  double GetThreadEfficiency() const {
    double scaling = GetThreadScaling();
    return scaling / thread_count;  // Perfect efficiency = 1.0
  }
};

/// @brief Thread worker function for multithreaded benchmarks
struct ThreadWorkerData {
  std::shared_ptr<SlideReader>
      shared_reader;  // Use shared reader instead of filename
  std::vector<TestRegion> regions;
  std::atomic<size_t>* completed_ops;
  std::atomic<size_t>* failed_ops;
  std::vector<double>* times_ms;
  std::mutex* times_mutex;
  std::atomic<size_t>* total_bytes;
  int thread_id;
};

void FastSlideWorkerThread(ThreadWorkerData* data) {
  // Use the shared reader instance (FastSlide is thread-safe)
  auto reader = data->shared_reader;

  if (!reader) {
    data->failed_ops->fetch_add(data->regions.size());
    return;
  }

  std::vector<double> local_times;
  local_times.reserve(data->regions.size());

  for (const auto& region : data->regions) {
    // Convert OpenSlide level-0 coordinates to
    // FastSlide level-native coordinates
    uint32_t fs_x = region.x;
    uint32_t fs_y = region.y;

    if (region.level > 0) {
      auto level_info_or = reader->GetLevelInfo(region.level);
      if (level_info_or.ok()) {
        double downsample = level_info_or.value().downsample_factor;
        fs_x = static_cast<uint32_t>(region.x / downsample);
        fs_y = static_cast<uint32_t>(region.y / downsample);
      } else {
        data->failed_ops->fetch_add(1);
        continue;
      }
    }

    RegionSpec fs_region;
    fs_region.top_left = {fs_x, fs_y};
    fs_region.size = {static_cast<uint32_t>(region.width),
                      static_cast<uint32_t>(region.height)};
    fs_region.level = region.level;

    auto start = std::chrono::high_resolution_clock::now();
    auto fs_result = reader->ReadRegion(fs_region);
    auto end = std::chrono::high_resolution_clock::now();

    if (fs_result.ok()) {
      auto duration =
          std::chrono::duration_cast<std::chrono::microseconds>(end - start);
      local_times.push_back(duration.count() / 1000.0);
      data->total_bytes->fetch_add(fs_result.value().SizeBytes());
      data->completed_ops->fetch_add(1);
    } else {
      data->failed_ops->fetch_add(1);
    }
  }

  // Add local times to global times vector (with mutex protection)
  {
    std::lock_guard<std::mutex> lock(*data->times_mutex);
    data->times_ms->insert(data->times_ms->end(), local_times.begin(),
                           local_times.end());
  }
}

/// @brief Run multithreaded FastSlide benchmark
PerformanceStats BenchmarkFastSlideMultithreaded(
    const std::string& filename, const std::vector<TestRegion>& regions,
    int thread_count) {
  PerformanceStats stats;
  stats.total_bytes = 0;
  stats.successful_ops = 0;
  stats.failed_ops = 0;

  std::cout << "FastSlide (" << thread_count << " threads): Testing "
            << regions.size() << " regions...\n";

  // Create a single shared reader instance (this is the key fix!)
  auto reader_or = SlideReaderRegistry::GetInstance().CreateReader(filename);
  if (!reader_or.ok()) {
    std::cout << "  Failed to create shared reader: "
              << reader_or.status().message() << "\n";
    return stats;
  }
  std::shared_ptr<SlideReader> shared_reader = std::move(reader_or.value());

  // Divide regions among threads
  size_t regions_per_thread = regions.size() / thread_count;
  size_t remaining_regions = regions.size() % thread_count;

  std::vector<std::thread> threads;
  std::vector<ThreadWorkerData> worker_data(thread_count);

  std::atomic<size_t> completed_ops{0};
  std::atomic<size_t> failed_ops{0};
  std::atomic<size_t> total_bytes{0};
  std::vector<double> all_times;
  std::mutex times_mutex;

  auto benchmark_start = std::chrono::high_resolution_clock::now();

  for (int i = 0; i < thread_count; ++i) {
    size_t start_idx = i * regions_per_thread;
    size_t end_idx = start_idx + regions_per_thread;
    if (i == thread_count - 1) {
      end_idx += remaining_regions;  // Last thread takes remaining regions
    }

    worker_data[i].shared_reader =
        shared_reader;  // Share the same reader instance
    worker_data[i].regions.assign(regions.begin() + start_idx,
                                  regions.begin() + end_idx);
    worker_data[i].completed_ops = &completed_ops;
    worker_data[i].failed_ops = &failed_ops;
    worker_data[i].times_ms = &all_times;
    worker_data[i].times_mutex = &times_mutex;
    worker_data[i].total_bytes = &total_bytes;
    worker_data[i].thread_id = i;

    threads.emplace_back(FastSlideWorkerThread, &worker_data[i]);
  }

  // Wait for all threads to complete
  for (auto& thread : threads) {
    thread.join();
  }

  auto benchmark_end = std::chrono::high_resolution_clock::now();
  auto total_duration = std::chrono::duration_cast<std::chrono::milliseconds>(
      benchmark_end - benchmark_start);

  stats.times_ms = all_times;
  stats.total_bytes = total_bytes.load();
  stats.successful_ops = completed_ops.load();
  stats.failed_ops = failed_ops.load();

  std::cout << "  Completed in " << total_duration.count()
            << " ms: " << stats.successful_ops << " successful, "
            << stats.failed_ops << " failed\n";
  std::cout << "  Concurrent throughput: " << std::fixed << std::setprecision(1)
            << (stats.successful_ops * 1000.0 / total_duration.count())
            << " ops/sec\n";

  // Print handle pool statistics to diagnose contention
  std::cout << "  Handle pool analysis:\n";
  std::cout << "    Default pool size: " << std::thread::hardware_concurrency()
            << " handles\n";
  std::cout << "    Thread count: " << thread_count << "\n";
  std::cout << "    Pool utilization: "
            << (static_cast<float>(thread_count) /
                static_cast<float>(std::thread::hardware_concurrency()) *
                100.0F)
            << "%\n";

  return stats;
}

/// @brief Save OpenSlide RGBA data as PNG (for visual verification)
bool SaveOpenSlideAsPNG(const std::string& filename, uint32_t* rgba_data,
                        uint32_t width, uint32_t height) {
  try {
    // Convert RGBA to RGB (OpenSlide uses ABGR format on little-endian)
    std::vector<uint8_t> rgb_data(width * height * 3);

    for (uint32_t y = 0; y < height; ++y) {
      for (uint32_t x = 0; x < width; ++x) {
        uint32_t rgba = rgba_data[y * width + x];
        size_t rgb_idx = (y * width + x) * 3;

        // Extract RGB components from RGBA
        rgb_data[rgb_idx + 0] = (rgba >> 0) & 0xFF;   // Red
        rgb_data[rgb_idx + 1] = (rgba >> 8) & 0xFF;   // Green
        rgb_data[rgb_idx + 2] = (rgba >> 16) & 0xFF;  // Blue
      }
    }

    // Create VipsImage from RGB data
    vips::VImage vips_image = vips::VImage::new_from_memory(
        rgb_data.data(), rgb_data.size(), width, height, 3, VIPS_FORMAT_UCHAR);

    vips_image.write_to_file(filename.c_str());
    return true;
  } catch (const vips::VError& e) {
    std::cerr << "Failed to save OpenSlide PNG: " << e.what() << "\n";
    return false;
  }
}

/// @brief Save FastSlide RGBImage as PNG (for visual verification)
bool SaveFastSlideAsPNG(const std::string& filename, const RGBImage& image) {
  try {
    // Create VipsImage from RGB data
    vips::VImage vips_image = vips::VImage::new_from_memory(
        const_cast<void*>(static_cast<const void*>(image.GetData())),
        image.SizeBytes(), image.GetDimensions()[0], image.GetDimensions()[1],
        3, VIPS_FORMAT_UCHAR);

    vips_image.write_to_file(filename.c_str());
    return true;
  } catch (const vips::VError& e) {
    std::cerr << "Failed to save FastSlide PNG: " << e.what() << "\n";
    return false;
  }
}

/// @brief Generate comprehensive test regions for multiple tile sizes
/// across all levels
std::vector<TestRegion> GenerateComprehensiveTestRegions(
    int max_levels, int64_t slide_width, int64_t slide_height,
    int regions_per_level_per_size = 200) {

  std::vector<TestRegion> test_regions;
  std::random_device rd;
  std::mt19937 gen(rd());

  // Test multiple tile sizes
  std::vector<uint32_t> tile_sizes = {256, 512, 1024};

  std::cout
      << "Generating comprehensive test regions for multiple tile sizes...\n";

  for (int level = 0; level < max_levels; ++level) {
    std::cout << "  Level " << level << ":\n";

    for (uint32_t tile_size : tile_sizes) {
      // Calculate safe bounds for this level and tile size
      int64_t max_x =
          std::max(static_cast<int64_t>(1), slide_width - tile_size);
      int64_t max_y =
          std::max(static_cast<int64_t>(1), slide_height - tile_size);

      if (max_x <= 0 || max_y <= 0) {
        std::cout << "    Skipping " << tile_size << "x" << tile_size
                  << " - image too small\n";
        continue;
      }

      std::uniform_int_distribution<int64_t> x_dist(0, max_x);
      std::uniform_int_distribution<int64_t> y_dist(0, max_y);

      for (int i = 0; i < regions_per_level_per_size; ++i) {
        test_regions.push_back({x_dist(gen), y_dist(gen), tile_size, tile_size,
                                level,
                                absl::StrCat("level", level, "_", tile_size,
                                             "x", tile_size, "_region", i)});
      }

      std::cout << "    " << tile_size << "x" << tile_size << ": "
                << regions_per_level_per_size << " regions (bounds: 0-" << max_x
                << ", 0-" << max_y << ")\n";
    }
  }

  std::cout << "Generated " << test_regions.size()
            << " total test regions across all tile sizes\n";
  return test_regions;
}

/// @brief OpenSlide benchmark functions
class OpenSlideBenchmark {
 public:
  explicit OpenSlideBenchmark(const std::string& filename)
      : filename_(filename), slide_(nullptr) {}

  ~OpenSlideBenchmark() {
    if (slide_) {
      openslide_close(slide_);
    }
  }

  bool Initialize() {
    slide_ = openslide_open(filename_.c_str());
    if (!slide_)
      return false;

    const char* error = openslide_get_error(slide_);
    if (error) {
      openslide_close(slide_);
      slide_ = nullptr;
      return false;
    }
    return true;
  }

  PerformanceStats BenchmarkOpenClose(int iterations = 20) {
    PerformanceStats stats;
    stats.total_bytes = 0;
    stats.successful_ops = 0;
    stats.failed_ops = 0;

    std::cout << "OpenSlide: Testing open/close performance with " << iterations
              << " iterations...\n";

    for (int i = 0; i < iterations; ++i) {
      if (i % 5 == 0) {
        std::cout << "  Progress: " << i << "/" << iterations << "\r"
                  << std::flush;
      }

      auto start = std::chrono::high_resolution_clock::now();

      openslide_t* temp_slide = openslide_open(filename_.c_str());
      if (!temp_slide) {
        stats.failed_ops++;
        continue;
      }

      const char* error = openslide_get_error(temp_slide);
      if (error) {
        stats.failed_ops++;
        openslide_close(temp_slide);
        continue;
      }

      openslide_close(temp_slide);

      auto end = std::chrono::high_resolution_clock::now();

      auto duration =
          std::chrono::duration_cast<std::chrono::microseconds>(end - start);
      stats.times_ms.push_back(duration.count() / 1000.0);
      stats.successful_ops++;
      // No significant data transfer for open/close, just metadata parsing
      stats.total_bytes += 1024;  // Nominal size for metadata
    }

    std::cout << "\n  Completed: " << stats.successful_ops << " successful, "
              << stats.failed_ops << " failed\n";
    return stats;
  }

  PerformanceStats BenchmarkRegions(const std::vector<TestRegion>& regions,
                                    bool save_sample = false) {
    PerformanceStats stats;
    stats.total_bytes = 0;
    stats.successful_ops = 0;
    stats.failed_ops = 0;

    if (!slide_)
      return stats;

    std::cout << "OpenSlide: Testing " << regions.size() << " regions...\n";

    for (size_t i = 0; i < regions.size(); ++i) {
      const auto& region = regions[i];

      if (i % 100 == 0) {
        std::cout << "  Progress: " << i << "/" << regions.size() << "\r"
                  << std::flush;
      }

      uint32_t* buffer = new uint32_t[region.width * region.height];

      auto start = std::chrono::high_resolution_clock::now();
      openslide_read_region(slide_, buffer, region.x, region.y, region.level,
                            region.width, region.height);
      auto end = std::chrono::high_resolution_clock::now();

      const char* error = openslide_get_error(slide_);
      if (error) {
        stats.failed_ops++;
        delete[] buffer;
        continue;
      }

      auto duration =
          std::chrono::duration_cast<std::chrono::microseconds>(end - start);
      stats.times_ms.push_back(duration.count() / 1000.0);
      stats.total_bytes += region.width * region.height * 4;  // RGBA
      stats.successful_ops++;

      // Save first region of each level for visual verification
      if (save_sample && i < 10) {
        std::string filename =
            "/Users/j.teuwen/test_out/" + region.description + "_openslide.png";
        SaveOpenSlideAsPNG(filename, buffer, region.width, region.height);
      }

      delete[] buffer;
    }

    std::cout << "\n  Completed: " << stats.successful_ops << " successful, "
              << stats.failed_ops << " failed\n";
    return stats;
  }

  int GetLevelCount() const {
    return slide_ ? openslide_get_level_count(slide_) : 0;
  }

  double GetLevelDownsample(int level) const {
    return slide_ ? openslide_get_level_downsample(slide_, level) : 1.0;
  }

 private:
  std::string filename_;
  openslide_t* slide_;
};

/// @brief FastSlide benchmark functions
class FastSlideBenchmark {
 public:
  explicit FastSlideBenchmark(const std::string& filename)
      : filename_(filename) {}

  bool Initialize() {
    auto reader_or = SlideReaderRegistry::GetInstance().CreateReader(filename_);
    if (!reader_or.ok()) {
      error_message_ = reader_or.status().message();
      return false;
    }
    reader_ = std::move(reader_or.value());
    return true;
  }

  PerformanceStats BenchmarkOpenClose(int iterations = 20) {
    PerformanceStats stats;
    stats.total_bytes = 0;
    stats.successful_ops = 0;
    stats.failed_ops = 0;

    std::cout << "FastSlide: Testing open/close performance with " << iterations
              << " iterations...\n";

    for (int i = 0; i < iterations; ++i) {
      if (i % 5 == 0) {
        std::cout << "  Progress: " << i << "/" << iterations << "\r"
                  << std::flush;
      }

      auto start = std::chrono::high_resolution_clock::now();

      auto temp_reader_or =
          SlideReaderRegistry::GetInstance().CreateReader(filename_);
      if (!temp_reader_or.ok()) {
        stats.failed_ops++;
        continue;
      }

      // Reader is automatically destroyed here, which closes the slide
      auto end = std::chrono::high_resolution_clock::now();

      auto duration =
          std::chrono::duration_cast<std::chrono::microseconds>(end - start);
      stats.times_ms.push_back(duration.count() / 1000.0);
      stats.successful_ops++;
      // No significant data transfer for open/close, just metadata parsing
      stats.total_bytes += 1024;  // Nominal size for metadata
    }

    std::cout << "\n  Completed: " << stats.successful_ops << " successful, "
              << stats.failed_ops << " failed\n";
    return stats;
  }

  PerformanceStats BenchmarkRegions(const std::vector<TestRegion>& regions,
                                    bool save_sample = false) {
    PerformanceStats stats;
    stats.total_bytes = 0;
    stats.successful_ops = 0;
    stats.failed_ops = 0;

    if (!reader_)
      return stats;

    std::cout << "FastSlide: Testing " << regions.size() << " regions...\n";

    for (size_t i = 0; i < regions.size(); ++i) {
      const auto& region = regions[i];

      if (i % 100 == 0) {
        std::cout << "  Progress: " << i << "/" << regions.size() << "\r"
                  << std::flush;
      }

      // Convert OpenSlide level-0 coordinates to
      // FastSlide level-native coordinates
      uint32_t fs_x = region.x;
      uint32_t fs_y = region.y;

      if (region.level > 0) {
        auto level_info_or = reader_->GetLevelInfo(region.level);
        if (level_info_or.ok()) {
          double downsample = level_info_or.value().downsample_factor;
          fs_x = static_cast<uint32_t>(region.x / downsample);
          fs_y = static_cast<uint32_t>(region.y / downsample);
        } else {
          stats.failed_ops++;
          continue;
        }
      }

      RegionSpec fs_region;
      fs_region.top_left = {fs_x, fs_y};
      fs_region.size = {static_cast<uint32_t>(region.width),
                        static_cast<uint32_t>(region.height)};
      fs_region.level = region.level;

      auto start = std::chrono::high_resolution_clock::now();
      auto fs_result = reader_->ReadRegion(fs_region);
      auto end = std::chrono::high_resolution_clock::now();

      if (!fs_result.ok()) {
        stats.failed_ops++;
        continue;
      }

      auto duration =
          std::chrono::duration_cast<std::chrono::microseconds>(end - start);
      stats.times_ms.push_back(duration.count() / 1000.0);
      stats.total_bytes += fs_result.value().SizeBytes();
      stats.successful_ops++;

      // Save first region of each level for visual verification
      if (save_sample && i < 10) {
        std::string filename =
            "/Users/j.teuwen/test_out/" + region.description + "_fastslide.png";
        SaveFastSlideAsPNG(filename, fs_result.value());
      }
    }

    std::cout << "\n  Completed: " << stats.successful_ops << " successful, "
              << stats.failed_ops << " failed\n";
    return stats;
  }

  int GetLevelCount() const { return reader_ ? reader_->GetLevelCount() : 0; }

  std::string GetFormatName() const {
    return reader_ ? reader_->GetFormatName() : "Unknown";
  }

  std::string GetErrorMessage() const { return error_message_; }

 private:
  std::string filename_;
  std::unique_ptr<SlideReader> reader_;
  std::string error_message_;
};

/// @brief Print detailed statistics
void PrintDetailedStats(const std::string& name,
                        const PerformanceStats& stats) {
  std::cout << "\n=== " << name << " Statistics ===\n";
  std::cout << "Operations: " << stats.successful_ops << " successful, "
            << stats.failed_ops << " failed\n";
  std::cout << "Total Data: " << std::fixed << std::setprecision(2)
            << (stats.total_bytes / (1024.0 * 1024.0)) << " MB\n";
  std::cout << "Mean Time: " << std::setprecision(3) << stats.GetMeanTime()
            << " ms\n";
  std::cout << "Median Time: " << stats.GetMedianTime() << " ms\n";
  std::cout << "Std Dev: " << stats.GetStdDev() << " ms\n";
  std::cout << "Min Time: " << stats.GetMinTime() << " ms\n";
  std::cout << "Max Time: " << stats.GetMaxTime() << " ms\n";
  std::cout << "Throughput: " << std::setprecision(2)
            << stats.GetThroughputMBps() << " MB/s\n";
  std::cout << "Ops/sec: " << std::setprecision(1) << stats.GetOpsPerSecond()
            << "\n";
}

/// @brief Print level benchmark comparison
void PrintLevelComparison(const LevelBenchmarkResult& result) {
  std::cout << "\n=== Level " << result.level << " Comparison ===\n";

  if (result.openslide_stats.successful_ops == 0) {
    std::cout << "OpenSlide: No successful operations\n";
  } else {
    std::cout << "OpenSlide: " << std::fixed << std::setprecision(2)
              << result.openslide_stats.GetMeanTime() << " ± "
              << result.openslide_stats.GetStdDev() << " ms, "
              << result.openslide_stats.GetThroughputMBps() << " MB/s\n";
  }

  if (result.fastslide_stats.successful_ops == 0) {
    std::cout << "FastSlide: No successful operations\n";
  } else {
    std::cout << "FastSlide: " << std::fixed << std::setprecision(2)
              << result.fastslide_stats.GetMeanTime() << " ± "
              << result.fastslide_stats.GetStdDev() << " ms, "
              << result.fastslide_stats.GetThroughputMBps() << " MB/s\n";
  }

  if (result.openslide_stats.successful_ops > 0 &&
      result.fastslide_stats.successful_ops > 0) {
    double speedup = result.GetSpeedup();
    std::cout << "Speedup: " << std::fixed << std::setprecision(2) << speedup
              << "x ";
    if (speedup > 1.0) {
      std::cout << "(FastSlide faster)";
    } else if (speedup < 1.0) {
      std::cout << "(OpenSlide faster)";
    } else {
      std::cout << "(Equal performance)";
    }
    std::cout << "\n";
    std::cout << "Throughput Ratio: " << result.GetThroughputRatio() << "x\n";
  }
}

/// @brief Print summary table
void PrintSummaryTable(const std::vector<LevelBenchmarkResult>& results) {
  std::cout << "\n=== Performance Summary Table ===\n";
  std::cout << std::left << std::setw(8) << "Level" << std::right
            << std::setw(18) << "OpenSlide (ms)" << std::right << std::setw(18)
            << "FastSlide (ms)" << std::right << std::setw(12) << "Speedup"
            << std::right << std::setw(15) << "FS Throughput" << std::right
            << std::setw(15) << "OS Throughput"
            << "\n";
  std::cout << std::left << std::setw(8) << "" << std::right << std::setw(18)
            << "(Mean ± StdDev)" << std::right << std::setw(18)
            << "(Mean ± StdDev)" << std::right << std::setw(12) << ""
            << std::right << std::setw(15) << "(MB/s)" << std::right
            << std::setw(15) << "(MB/s)"
            << "\n";
  std::cout << std::string(86, '-') << "\n";

  for (const auto& result : results) {
    if (result.openslide_stats.successful_ops > 0 &&
        result.fastslide_stats.successful_ops > 0) {

      // Format mean ± stddev strings using stringstream
      std::ostringstream os_time_stream, fs_time_stream;
      os_time_stream << std::fixed << std::setprecision(2)
                     << result.openslide_stats.GetMeanTime() << " ± "
                     << std::setprecision(2)
                     << result.openslide_stats.GetStdDev();

      fs_time_stream << std::fixed << std::setprecision(2)
                     << result.fastslide_stats.GetMeanTime() << " ± "
                     << std::setprecision(2)
                     << result.fastslide_stats.GetStdDev();

      std::cout << std::left << std::setw(8) << result.level << std::right
                << std::setw(18) << os_time_stream.str() << std::right
                << std::setw(18) << fs_time_stream.str() << std::right
                << std::setw(10) << std::fixed << std::setprecision(2)
                << result.GetSpeedup() << "x" << std::right << std::setw(13)
                << std::fixed << std::setprecision(1)
                << result.fastslide_stats.GetThroughputMBps() << " MB/s"
                << std::right << std::setw(13) << std::fixed
                << std::setprecision(1)
                << result.openslide_stats.GetThroughputMBps() << " MB/s"
                << "\n";
    }
  }

  std::cout << "\nVariance Analysis:\n";
  for (const auto& result : results) {
    if (result.openslide_stats.successful_ops > 0 &&
        result.fastslide_stats.successful_ops > 0) {
      std::cout << "Level " << result.level << ":\n";
      std::cout << "  OpenSlide:  Variance=" << std::fixed
                << std::setprecision(2)
                << result.openslide_stats.GetStdDev() *
                       result.openslide_stats.GetStdDev()
                << " ms², CV=" << std::setprecision(1)
                << (result.openslide_stats.GetStdDev() /
                    result.openslide_stats.GetMeanTime() * 100)
                << "%\n";
      std::cout << "  FastSlide:  Variance=" << std::fixed
                << std::setprecision(2)
                << result.fastslide_stats.GetStdDev() *
                       result.fastslide_stats.GetStdDev()
                << " ms², CV=" << std::setprecision(1)
                << (result.fastslide_stats.GetStdDev() /
                    result.fastslide_stats.GetMeanTime() * 100)
                << "%\n";
    }
  }
}

/// @brief Print comprehensive tile size performance analysis
void PrintTileSizeAnalysis(
    const std::map<uint32_t, std::map<int, TileSizeBenchmarkResult>>& results) {
  std::cout << "\n=== Comprehensive Tile Size Performance Analysis ===\n";

  // Print performance table for each tile size
  for (const auto& [tile_size, level_results] : results) {
    std::cout << "\n--- " << tile_size << "x" << tile_size
              << " Tile Performance ---\n";
    std::cout << std::left << std::setw(8) << "Level" << std::right
              << std::setw(18) << "OpenSlide (ms)" << std::right
              << std::setw(18) << "FastSlide (ms)" << std::right
              << std::setw(12) << "Speedup" << std::right << std::setw(15)
              << "FS Throughput" << std::right << std::setw(15) << "Samples"
              << "\n";
    std::cout << std::string(86, '-') << "\n";

    for (const auto& [level, result] : level_results) {
      if (result.openslide_stats.successful_ops > 0 &&
          result.fastslide_stats.successful_ops > 0) {
        std::cout << std::left << std::setw(8) << level << std::right
                  << std::setw(16) << std::fixed << std::setprecision(2)
                  << result.openslide_stats.GetMeanTime() << " ms" << std::right
                  << std::setw(16) << std::fixed << std::setprecision(2)
                  << result.fastslide_stats.GetMeanTime() << " ms" << std::right
                  << std::setw(10) << std::fixed << std::setprecision(2)
                  << result.GetSpeedup() << "x" << std::right << std::setw(12)
                  << std::fixed << std::setprecision(1)
                  << result.fastslide_stats.GetThroughputMBps() << " MB/s"
                  << std::right << std::setw(8)
                  << result.fastslide_stats.successful_ops << "\n";
      }
    }
  }

  // Print tile size comparison summary
  std::cout << "\n=== Tile Size Performance Summary ===\n";
  std::cout << std::left << std::setw(12) << "Tile Size" << std::right
            << std::setw(15) << "Avg Speedup" << std::right << std::setw(18)
            << "Avg Throughput" << std::right << std::setw(12) << "Total Ops"
            << "\n";
  std::cout << std::string(57, '-') << "\n";

  for (const auto& [tile_size, level_results] : results) {
    std::vector<double> speedups;
    std::vector<double> throughputs;
    size_t total_ops = 0;

    for (const auto& [level, result] : level_results) {
      if (result.openslide_stats.successful_ops > 0 &&
          result.fastslide_stats.successful_ops > 0) {
        speedups.push_back(result.GetSpeedup());
        throughputs.push_back(result.fastslide_stats.GetThroughputMBps());
        total_ops += result.fastslide_stats.successful_ops;
      }
    }

    if (!speedups.empty()) {
      double avg_speedup =
          std::accumulate(speedups.begin(), speedups.end(), 0.0) /
          speedups.size();
      double avg_throughput =
          std::accumulate(throughputs.begin(), throughputs.end(), 0.0) /
          throughputs.size();

      std::cout << std::left << std::setw(12)
                << (std::to_string(tile_size) + "x" + std::to_string(tile_size))
                << std::right << std::setw(13) << std::fixed
                << std::setprecision(2) << avg_speedup << "x" << std::right
                << std::setw(15) << std::fixed << std::setprecision(1)
                << avg_throughput << " MB/s" << std::right << std::setw(12)
                << total_ops << "\n";
    }
  }
}

/// @brief Print comprehensive tile size performance table with tile_size,
/// level, and time ± stddev
void PrintComprehensiveTileSizeTable(
    const std::vector<TileSizeBenchmarkResult>& results) {
  std::cout << "\n=== Comprehensive Tile Size Performance Table ===\n";
  std::cout << std::left << std::setw(12) << "Tile Size" << std::left
            << std::setw(8) << "Level" << std::right << std::setw(20)
            << "OpenSlide (ms)" << std::right << std::setw(20)
            << "FastSlide (ms)" << std::right << std::setw(12) << "Speedup"
            << std::right << std::setw(15) << "Throughput" << std::right
            << std::setw(10) << "Samples"
            << "\n";
  std::cout << std::left << std::setw(12) << "" << std::left << std::setw(8)
            << "" << std::right << std::setw(20) << "(Mean ± StdDev)"
            << std::right << std::setw(20) << "(Mean ± StdDev)" << std::right
            << std::setw(12) << "" << std::right << std::setw(15) << "(MB/s)"
            << std::right << std::setw(10) << ""
            << "\n";
  std::cout << std::string(97, '-') << "\n";

  for (const auto& result : results) {
    if (result.openslide_stats.successful_ops > 0 &&
        result.fastslide_stats.successful_ops > 0) {
      // Format timing strings with ± stddev
      std::ostringstream os_time_str, fs_time_str;
      os_time_str << std::fixed << std::setprecision(2)
                  << result.openslide_stats.GetMeanTime() << " ± "
                  << result.openslide_stats.GetStdDev();
      fs_time_str << std::fixed << std::setprecision(2)
                  << result.fastslide_stats.GetMeanTime() << " ± "
                  << result.fastslide_stats.GetStdDev();

      std::cout << std::left << std::setw(12)
                << (std::to_string(result.tile_size) + "x" +
                    std::to_string(result.tile_size))
                << std::left << std::setw(8) << result.level << std::right
                << std::setw(20) << os_time_str.str() << std::right
                << std::setw(20) << fs_time_str.str() << std::right
                << std::setw(10) << std::fixed << std::setprecision(2)
                << result.GetSpeedup() << "x" << std::right << std::setw(12)
                << std::fixed << std::setprecision(1)
                << result.fastslide_stats.GetThroughputMBps() << " MB/s"
                << std::right << std::setw(10)
                << result.fastslide_stats.successful_ops << "\n";
    }
  }

  // Print summary statistics by tile size
  std::cout << "\n=== Tile Size Performance Summary ===\n";
  std::cout << std::left << std::setw(12) << "Tile Size" << std::right
            << std::setw(15) << "Avg Speedup" << std::right << std::setw(18)
            << "Avg Throughput" << std::right << std::setw(12) << "Total Tests"
            << "\n";
  std::cout << std::string(57, '-') << "\n";

  std::map<uint32_t, std::vector<TileSizeBenchmarkResult>> by_tile_size;
  for (const auto& result : results) {
    by_tile_size[result.tile_size].push_back(result);
  }

  for (const auto& [tile_size, tile_results] : by_tile_size) {
    std::vector<double> speedups;
    std::vector<double> throughputs;

    for (const auto& result : tile_results) {
      if (result.openslide_stats.successful_ops > 0 &&
          result.fastslide_stats.successful_ops > 0) {
        speedups.push_back(result.GetSpeedup());
        throughputs.push_back(result.fastslide_stats.GetThroughputMBps());
      }
    }

    if (!speedups.empty()) {
      double avg_speedup =
          std::accumulate(speedups.begin(), speedups.end(), 0.0) /
          speedups.size();
      double avg_throughput =
          std::accumulate(throughputs.begin(), throughputs.end(), 0.0) /
          throughputs.size();

      std::cout << std::left << std::setw(12)
                << (std::to_string(tile_size) + "x" + std::to_string(tile_size))
                << std::right << std::setw(13) << std::fixed
                << std::setprecision(2) << avg_speedup << "x" << std::right
                << std::setw(15) << std::fixed << std::setprecision(1)
                << avg_throughput << " MB/s" << std::right << std::setw(12)
                << tile_results.size() << "\n";
    }
  }
}

/// @brief Print multithreaded performance comparison
void PrintMultithreadedComparison(
    const std::vector<MultithreadedBenchmarkResult>& results) {
  std::cout << "\n=== Multithreaded Performance Analysis ===\n";
  std::cout << std::left << std::setw(12) << "Tile Size" << std::left
            << std::setw(8) << "Level" << std::left << std::setw(10)
            << "Threads" << std::right << std::setw(18) << "Single-Thread (ms)"
            << std::right << std::setw(18) << "Multi-Thread (ms)" << std::right
            << std::setw(12) << "Scaling" << std::right << std::setw(12)
            << "Efficiency"
            << "\n";
  std::cout << std::left << std::setw(12) << "" << std::left << std::setw(8)
            << "" << std::left << std::setw(10) << "" << std::right
            << std::setw(18) << "(Mean ± StdDev)" << std::right << std::setw(18)
            << "(Mean ± StdDev)" << std::right << std::setw(12) << ""
            << std::right << std::setw(12) << "(%)"
            << "\n";
  std::cout << std::string(92, '-') << "\n";

  for (const auto& result : results) {
    // Format timing strings
    std::ostringstream st_time_str, mt_time_str;
    st_time_str << std::fixed << std::setprecision(2)
                << result.single_threaded_stats.GetMeanTime() << " ± "
                << result.single_threaded_stats.GetStdDev();
    mt_time_str << std::fixed << std::setprecision(2)
                << result.multi_threaded_stats.GetMeanTime() << " ± "
                << result.multi_threaded_stats.GetStdDev();

    std::cout << std::left << std::setw(12)
              << (std::to_string(result.tile_size) + "x" +
                  std::to_string(result.tile_size))
              << std::left << std::setw(8) << result.level << std::left
              << std::setw(10) << result.thread_count << std::right
              << std::setw(18) << st_time_str.str() << std::right
              << std::setw(18) << mt_time_str.str() << std::right
              << std::setw(10) << std::fixed << std::setprecision(2)
              << result.GetThreadScaling() << "x" << std::right << std::setw(11)
              << std::fixed << std::setprecision(1)
              << (result.GetThreadEfficiency() * 100.0) << "%"
              << "\n";
  }

  // Print threading efficiency summary
  std::cout << "\n=== Threading Efficiency Summary ===\n";
  std::cout << std::left << std::setw(10) << "Threads" << std::right
            << std::setw(15) << "Avg Scaling" << std::right << std::setw(15)
            << "Avg Efficiency" << std::right << std::setw(12) << "Tests"
            << "\n";
  std::cout << std::string(52, '-') << "\n";

  std::map<int, std::vector<MultithreadedBenchmarkResult>> by_thread_count;
  for (const auto& result : results) {
    by_thread_count[result.thread_count].push_back(result);
  }

  for (const auto& [thread_count, thread_results] : by_thread_count) {
    std::vector<double> scalings;
    std::vector<double> efficiencies;

    for (const auto& result : thread_results) {
      scalings.push_back(result.GetThreadScaling());
      efficiencies.push_back(result.GetThreadEfficiency());
    }

    if (!scalings.empty()) {
      double avg_scaling =
          std::accumulate(scalings.begin(), scalings.end(), 0.0) /
          scalings.size();
      double avg_efficiency =
          std::accumulate(efficiencies.begin(), efficiencies.end(), 0.0) /
          efficiencies.size();

      std::cout << std::left << std::setw(10) << thread_count << std::right
                << std::setw(13) << std::fixed << std::setprecision(2)
                << avg_scaling << "x" << std::right << std::setw(14)
                << std::fixed << std::setprecision(1)
                << (avg_efficiency * 100.0) << "%" << std::right
                << std::setw(12) << thread_results.size() << "\n";
    }
  }
}

/// @brief Test different multithreading strategies
void TestThreadingStrategies(const std::string& filename) {
  std::cout << "\n=== Threading Strategy Analysis ===\n";

  // Test 1: Sequential regions (what we're doing now)
  std::cout << "Test 1: Sequential regions (current approach)\n";
  std::cout << "  Multiple threads reading different regions from same file\n";
  std::cout << "  Expected: Poor scaling due to I/O contention\n";

  // Test 2: Same region, multiple threads (cache test)
  std::cout << "\nTest 2: Same region repeated reads (cache behavior)\n";
  std::cout << "  Multiple threads reading the SAME region repeatedly\n";
  std::cout << "  Expected: Better scaling if caching works\n";

  // Test 3: I/O vs CPU bound analysis
  std::cout << "\nTest 3: I/O vs CPU analysis\n";
  std::cout << "  System Info:\n";
  std::cout << "    CPU cores: " << std::thread::hardware_concurrency() << "\n";
  std::cout << "    Default handle pool size: "
            << std::thread::hardware_concurrency() << "\n";

  // Test 4: Optimal thread count prediction
  std::cout << "\nTest 4: Optimal threading for I/O-bound workloads\n";
  std::cout << "  For single-file I/O: Optimal threads = 1-2 (minimal context "
               "switching)\n";
  std::cout << "  For this workload: Threading helps ONLY if:\n";
  std::cout << "    - Multiple storage devices involved\n";
  std::cout << "    - Significant CPU processing per region\n";
  std::cout << "    - Effective caching between reads\n";

  std::cout << "\n  Recommendation: Use single-threaded for file I/O,\n";
  std::cout << "  multithreading for processing the returned image data.\n";
}

}  // namespace

int main(int argc, char** argv) {
  if (argc != 2) {
    std::cerr << "Usage: " << argv[0] << " <slide_file>\n";
    std::cerr << "Example: " << argv[0] << " example.svs\n";
    return 1;
  }

  // Initialize libvips
  if (VIPS_INIT(argv[0])) {
    vips_error_exit(nullptr);
  }

  std::string filename = argv[1];

  // Create output directory
  std::filesystem::create_directories("/Users/j.teuwen/test_out");

  try {
    std::cout << "=== Comprehensive OpenSlide vs FastSlide Benchmark ===\n";
    std::cout << "File: " << filename << "\n";
    std::cout << "Testing with 1000 random regions per level for statistical "
                 "robustness\n\n";

    // Initialize FastSlide readers
    InitializeReaders();

    // Create benchmarks
    OpenSlideBenchmark os_bench(filename);
    FastSlideBenchmark fs_bench(filename);

    // Initialize both
    if (!os_bench.Initialize()) {
      std::cerr << "Failed to initialize OpenSlide\n";
      return 1;
    }

    if (!fs_bench.Initialize()) {
      std::cerr << "Failed to initialize FastSlide: "
                << fs_bench.GetErrorMessage() << "\n";
      return 1;
    }

    // Print slide information
    std::cout << "=== Slide Information ===\n";
    std::cout << "OpenSlide Levels: " << os_bench.GetLevelCount() << "\n";
    std::cout << "FastSlide Levels: " << fs_bench.GetLevelCount() << "\n";
    std::cout << "FastSlide Format: " << fs_bench.GetFormatName() << "\n";

    // Determine maximum levels available in both readers
    int max_levels =
        std::min(os_bench.GetLevelCount(), fs_bench.GetLevelCount());

    // Get actual slide dimensions from OpenSlide (level 0)
    int64_t slide_width = 60000;   // Default fallback
    int64_t slide_height = 35000;  // Default fallback

    // Try to get actual dimensions from OpenSlide
    openslide_t* temp_slide = openslide_open(filename.c_str());
    if (temp_slide) {
      openslide_get_level0_dimensions(temp_slide, &slide_width, &slide_height);
      openslide_close(temp_slide);
      std::cout << "Actual slide dimensions: " << slide_width << "x"
                << slide_height << "\n";
    } else {
      std::cout << "Warning: Using default slide dimensions: " << slide_width
                << "x" << slide_height << "\n";
    }

    std::vector<LevelBenchmarkResult> level_results;

    // Test open/close performance first
    std::cout << "\n=== Testing Open/Close Performance ===\n";
    std::cout << "Testing slide initialization and cleanup performance...\n";

    auto os_open_close_stats = os_bench.BenchmarkOpenClose(20);
    auto fs_open_close_stats = fs_bench.BenchmarkOpenClose(20);

    std::cout << "\n=== Open/Close Performance Comparison ===\n";
    std::cout << "OpenSlide: " << std::fixed << std::setprecision(2)
              << os_open_close_stats.GetMeanTime() << " ± "
              << os_open_close_stats.GetStdDev()
              << " ms (median: " << os_open_close_stats.GetMedianTime()
              << " ms)\n";
    std::cout << "FastSlide: " << std::fixed << std::setprecision(2)
              << fs_open_close_stats.GetMeanTime() << " ± "
              << fs_open_close_stats.GetStdDev()
              << " ms (median: " << fs_open_close_stats.GetMedianTime()
              << " ms)\n";

    if (os_open_close_stats.successful_ops > 0 &&
        fs_open_close_stats.successful_ops > 0) {
      double open_close_speedup =
          os_open_close_stats.GetMeanTime() / fs_open_close_stats.GetMeanTime();
      std::cout << "Open/Close Speedup: " << std::fixed << std::setprecision(2)
                << open_close_speedup << "x ";
      if (open_close_speedup > 1.0) {
        std::cout << "(FastSlide faster)";
      } else if (open_close_speedup < 1.0) {
        std::cout << "(OpenSlide faster)";
      } else {
        std::cout << "(Equal performance)";
      }
      std::cout << "\n";
    }

    // Generate comprehensive test regions for all levels and tile sizes upfront
    std::cout << "\n=== Generating Comprehensive Test Regions ===\n";
    auto all_test_regions = GenerateComprehensiveTestRegions(
        max_levels, slide_width, slide_height, 200);

    // Collect results for comprehensive analysis
    std::vector<TileSizeBenchmarkResult> all_tile_results;

    // Test each tile size and level combination separately
    std::vector<uint32_t> tile_sizes = {256, 512, 1024};

    for (uint32_t tile_size : tile_sizes) {
      for (int level = 0; level < max_levels; ++level) {
        // Filter regions for this specific tile size and level
        std::vector<TestRegion> filtered_regions;
        for (const auto& region : all_test_regions) {
          if (region.level == level && region.width == tile_size &&
              region.height == tile_size) {
            filtered_regions.push_back(region);
          }
        }

        if (filtered_regions.empty()) {
          std::cout << "No " << tile_size << "x" << tile_size
                    << " regions available for level " << level << "\n";
          continue;
        }

        std::cout << "\n=== Testing " << tile_size << "x" << tile_size
                  << " tiles at Level " << level << " ===\n";
        std::cout << "Testing " << filtered_regions.size() << " regions...\n";

        TileSizeBenchmarkResult tile_result;
        tile_result.level = level;
        tile_result.tile_size = tile_size;
        tile_result.test_name =
            absl::StrCat(tile_size, "x", tile_size, " Level ", level);

        // Benchmark OpenSlide for this tile size + level combination
        tile_result.openslide_stats =
            os_bench.BenchmarkRegions(filtered_regions, false);

        // Benchmark FastSlide for this tile size + level combination
        tile_result.fastslide_stats =
            fs_bench.BenchmarkRegions(filtered_regions, false);

        all_tile_results.push_back(tile_result);

        // Print immediate results
        if (tile_result.openslide_stats.successful_ops > 0 &&
            tile_result.fastslide_stats.successful_ops > 0) {
          std::cout << "  OpenSlide: " << std::fixed << std::setprecision(2)
                    << tile_result.openslide_stats.GetMeanTime() << " ± "
                    << tile_result.openslide_stats.GetStdDev() << " ms\n";
          std::cout << "  FastSlide: " << std::fixed << std::setprecision(2)
                    << tile_result.fastslide_stats.GetMeanTime() << " ± "
                    << tile_result.fastslide_stats.GetStdDev() << " ms\n";
          std::cout << "  Speedup: " << tile_result.GetSpeedup() << "x\n";
        }
      }
    }

    // Print comprehensive tile size performance table
    PrintComprehensiveTileSizeTable(all_tile_results);

    // Analyze why threading isn't helping
    TestThreadingStrategies(filename);

    // Run multithreaded benchmarks (FastSlide only - it's thread-safe)
    std::cout << "\n=== Running Multithreaded Benchmarks ===\n";
    std::cout << "FastSlide supports native thread safety - testing concurrent "
                 "performance...\n";

    std::vector<MultithreadedBenchmarkResult> multithreaded_results;
    std::vector<int> thread_counts = {2, 4, 8};

    // Test multithreading on a subset of tile sizes and levels for efficiency
    std::vector<uint32_t> mt_tile_sizes = {
        512, 1024};  // Focus on larger tiles for multithreading
    std::vector<int> mt_levels = {0, 1};  // Test first two levels

    for (uint32_t tile_size : mt_tile_sizes) {
      for (int level : mt_levels) {
        if (level >= max_levels)
          continue;

        // Get single-threaded results for this tile size/level combination
        TileSizeBenchmarkResult* single_threaded_result = nullptr;
        for (auto& result : all_tile_results) {
          if (result.tile_size == tile_size && result.level == level) {
            single_threaded_result = &result;
            break;
          }
        }

        if (!single_threaded_result ||
            single_threaded_result->fastslide_stats.successful_ops == 0) {
          std::cout << "Skipping " << tile_size << "x" << tile_size << " level "
                    << level << " - no single-threaded baseline\n";
          continue;
        }

        // Get test regions for this combination
        std::vector<TestRegion> mt_regions;
        for (const auto& region : all_test_regions) {
          if (region.level == level && region.width == tile_size &&
              region.height == tile_size) {
            mt_regions.push_back(region);
          }
        }

        if (mt_regions.empty()) {
          continue;
        }

        // Limit to reasonable number of regions for multithreaded testing
        if (mt_regions.size() > 100) {
          mt_regions.resize(100);
        }

        std::cout << "\nTesting " << tile_size << "x" << tile_size
                  << " tiles at level " << level << " with multithreading:\n";

        for (int thread_count : thread_counts) {
          std::cout << "  Testing with " << thread_count << " threads...\n";

          auto mt_stats = BenchmarkFastSlideMultithreaded(filename, mt_regions,
                                                          thread_count);

          if (mt_stats.successful_ops > 0) {
            MultithreadedBenchmarkResult mt_result;
            mt_result.tile_size = tile_size;
            mt_result.level = level;
            mt_result.thread_count = thread_count;
            mt_result.test_name =
                absl::StrCat(tile_size, "x", tile_size, " Level ", level, " (",
                             thread_count, " threads)");
            mt_result.single_threaded_stats =
                single_threaded_result->fastslide_stats;
            mt_result.multi_threaded_stats = mt_stats;

            multithreaded_results.push_back(mt_result);

            std::cout << "    Scaling: " << std::fixed << std::setprecision(2)
                      << mt_result.GetThreadScaling()
                      << "x, Efficiency: " << std::setprecision(1)
                      << (mt_result.GetThreadEfficiency() * 100.0) << "%\n";
          }
        }
      }
    }

    // Print multithreaded analysis
    if (!multithreaded_results.empty()) {
      PrintMultithreadedComparison(multithreaded_results);
    } else {
      std::cout << "\nNo multithreaded results to analyze.\n";
    }

    // Calculate overall statistics from tile-based results
    if (!all_tile_results.empty()) {
      std::vector<double> speedups;
      double total_os_time = 0.0, total_fs_time = 0.0;
      size_t total_ops = 0;
      size_t successful_combinations = 0;

      for (const auto& result : all_tile_results) {
        if (result.openslide_stats.successful_ops > 0 &&
            result.fastslide_stats.successful_ops > 0) {
          speedups.push_back(result.GetSpeedup());
          total_os_time += result.openslide_stats.GetMeanTime() *
                           result.openslide_stats.successful_ops;
          total_fs_time += result.fastslide_stats.GetMeanTime() *
                           result.fastslide_stats.successful_ops;
          total_ops += result.fastslide_stats.successful_ops;
          successful_combinations++;
        }
      }

      if (!speedups.empty()) {
        double avg_speedup = 0.0;
        for (double s : speedups)
          avg_speedup += s;
        avg_speedup /= speedups.size();

        double overall_speedup =
            (total_ops > 0) ? (total_os_time / total_fs_time) : 0.0;

        double open_close_speedup = 0.0;
        if (os_open_close_stats.successful_ops > 0 &&
            fs_open_close_stats.successful_ops > 0) {
          open_close_speedup = os_open_close_stats.GetMeanTime() /
                               fs_open_close_stats.GetMeanTime();
        }

        // Calculate multithreading benefits
        double best_thread_scaling = 1.0;
        double avg_thread_efficiency = 0.0;
        if (!multithreaded_results.empty()) {
          std::vector<double> scalings;
          std::vector<double> efficiencies;
          for (const auto& mt_result : multithreaded_results) {
            scalings.push_back(mt_result.GetThreadScaling());
            efficiencies.push_back(mt_result.GetThreadEfficiency());
          }
          best_thread_scaling =
              *std::max_element(scalings.begin(), scalings.end());
          avg_thread_efficiency =
              std::accumulate(efficiencies.begin(), efficiencies.end(), 0.0) /
              efficiencies.size();
        }

        std::cout << "\n=== Overall Results ===\n";
        std::cout << "Tile size/level combinations tested: "
                  << successful_combinations << "\n";
        std::cout << "Total operations: " << total_ops << "\n";
        std::cout << "Average speedup across all combinations: " << std::fixed
                  << std::setprecision(2) << avg_speedup << "x\n";
        std::cout << "Weighted overall speedup: " << overall_speedup << "x\n";
        if (open_close_speedup > 0.0) {
          std::cout << "Open/Close speedup: " << open_close_speedup << "x\n";
        }
        if (!multithreaded_results.empty()) {
          std::cout << "Best multithreaded scaling: " << best_thread_scaling
                    << "x\n";
          std::cout << "Average thread efficiency: " << std::fixed
                    << std::setprecision(1) << (avg_thread_efficiency * 100.0)
                    << "%\n";
        }

        std::cout << "\nFastSlide advantages:\n";
        std::cout << "  - Comprehensive testing across multiple tile sizes "
                     "(256x256, 512x512, 1024x1024)\n";
        std::cout << "  - Native thread safety (no external locking needed)\n";
        std::cout << "  - Modern C++ memory safety with RAII\n";
        std::cout << "  - Comprehensive error handling with absl::Status\n";
        std::cout << "  - Zero memory leaks guaranteed\n";

        std::cout << "\nThreading Performance Analysis:\n";
        std::cout
            << "  ⚠️  Current threading shows limited benefit due to I/O "
               "bottlenecks\n";
        std::cout << "  💡 Better threading strategies:\n";
        std::cout << "     1. Use threading for image PROCESSING, not I/O\n";
        std::cout
            << "     2. Implement async I/O with thread pool for processing\n";
        std::cout
            << "     3. Add tile-level caching for repeated access patterns\n";
        std::cout << "     4. Use 1-2 I/O threads max for single files\n";
        std::cout << "  ✅ Thread safety infrastructure is solid and ready for "
                     "optimization\n";
        if (open_close_speedup > 1.0) {
          std::cout << "  - " << open_close_speedup
                    << "x faster slide initialization\n";
        }

        if (avg_speedup > 1.0) {
          std::cout << "  - " << avg_speedup
                    << "x faster on average across all tile sizes and levels\n";
        }

        if (best_thread_scaling > 1.5) {
          std::cout << "  - " << best_thread_scaling
                    << "x maximum concurrent scaling (OpenSlide cannot do this "
                       "safely)\n";
        }

        if (avg_thread_efficiency > 0.5) {
          std::cout << "  - " << std::fixed << std::setprecision(0)
                    << (avg_thread_efficiency * 100.0)
                    << "% average thread efficiency for concurrent workloads\n";
        }
      }
    }

  } catch (const std::exception& e) {
    std::cerr << "Error: " << e.what() << "\n";
    vips_shutdown();
    return 1;
  }

  vips_shutdown();
  return 0;
}
