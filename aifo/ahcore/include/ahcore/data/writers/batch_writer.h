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
#ifndef AIFO_AHCORE_INCLUDE_AHCORE_DATA_WRITERS_BATCH_WRITER_H_
#define AIFO_AHCORE_INCLUDE_AHCORE_DATA_WRITERS_BATCH_WRITER_H_

#include <spdlog/spdlog.h>
#include <vips/vips8>

#include <algorithm>  // for std::clamp
#include <condition_variable>
#include <filesystem>
#include <functional>
#include <memory>
#include <mutex>
#include <queue>
#include <stdexcept>
#include <string>
#include <thread>
#include <tuple>
#include <unordered_map>
#include <utility>
#include <vector>

#include <xtensor/containers/xarray.hpp>

#include "ahcore/data/writers/disk_tile_writer.h"
#include "ahcore/utilities/logging.h"
#include "aifocore/concepts/numeric.h"
#include "aifocore/tiling/grid.h"

auto global_logger = aifo::logging::GetGlobalLogger();
using aifocore::Size;

namespace aifo {

// Struct for configuration
struct Config {
  double mpp;                                          // Microns per pixel
  Size<int, 2> offset;                                 // (x_offset, y_offset)
  Size<int, 2> size;                                   // (width, height)
  Size<int, 2> tile_size;                              // (width, height)
  Size<int, 2> tile_overlap;                           // (x_overlap, y_overlap)
  std::vector<std::tuple<int, int>> grid_coordinates;  // Grid for tiling

  Config(double mpp = 0.5, Size<int, 2> offset = {0, 0},
         Size<int, 2> size = {0, 0}, Size<int, 2> tile_size = {64, 64},
         Size<int, 2> tile_overlap = {0, 0})
      : mpp(mpp),
        offset(offset),
        size(size),
        tile_size(tile_size),
        tile_overlap(tile_overlap) {}
};

// Template struct for batch data
template <typename T>
struct BatchData {
  xt::xarray<T> coordinates;  // (x, y) coordinate
  xt::xarray<T> data;         // Image data (shape: height, width, channels)
  std::string filename;       // Output filename
  bool last_batch;  // Indicates if this is the last batch for the file
};

// Templated MultiFileBatchProcessor for int/float types
template <typename T>
class MultiFileBatchProcessor {
 public:
  explicit MultiFileBatchProcessor(size_t max_threads = 4)
      : max_threads_(max_threads), active_threads_(0) {
    if (VIPS_INIT("MultiFileBatchProcessor")) {
      throw std::runtime_error("Failed to initialize libvips");
    }
    spdlog::info("MultiFileBatchProcessor created with max_threads = {}",
                 max_threads_);
  }

  ~MultiFileBatchProcessor() {
    spdlog::info("Destructor called. Stopping all threads...");
    StopAllThreads();
    vips_shutdown();  // Clean up libvips resources
    spdlog::info("Destructor finished.");
  }

  void RegisterCallback(std::function<Config(const std::string&)> callback) {
    if (!callback) {
      throw std::invalid_argument("Callback cannot be null");
    }
    spdlog::info("Callback registered.");
    config_callback_ = std::move(callback);
    callback_registered_ = true;
  }

  void AddBatch(const xt::xarray<T>& coordinates, const xt::xarray<T>& data,
                const std::string& filename, bool last_batch) {
    if (!callback_registered_) {
      throw std::runtime_error(
          "No callback registered! Please register a callback using "
          "`register_callback`.");
    }

    {
      std::lock_guard<std::mutex> lock(mutex_);
      if (batch_queues_.find(filename) == batch_queues_.end()) {
        {
          std::unique_lock<std::mutex> thread_lock(mutex_threads_);
          cv_threads_.wait(thread_lock,
                           [this] { return active_threads_ < max_threads_; });
        }

        spdlog::info("New filename detected: {}", filename);
        Config config = config_callback_(filename);
        spdlog::info(
            "Configuration for {}: mpp = {}, tile_size = {}, offset = "
            "{}, tile_overlap = {}",
            filename, config.mpp, config.tile_size, config.offset,
            config.tile_overlap);

        configs_[filename] = config;  // Store configuration
        batch_queues_[filename] = std::make_shared<std::queue<BatchData<T>>>();
        is_running_[filename] = true;
        StartThreadForFile(filename);
      }

      batch_queues_[filename]->push({coordinates, data, filename, last_batch});
      spdlog::info("Added batch to queue for {}. Last batch: {}", filename,
                   last_batch);
    }
    cv_.notify_all();
  }

 private:
  void StartThreadForFile(const std::string& filename) {
    {
      std::lock_guard<std::mutex> lock(mutex_threads_);
      active_threads_++;
      spdlog::info("Active threads incremented to {}", active_threads_);
    }

    consumer_threads_[filename] = std::thread([this, filename]() {
      try {
        spdlog::info("Thread started for {}", filename);
        ProcessFile(filename);
      } catch (const std::exception& e) {
        spdlog::error("Exception in thread for {}: {}", filename, e.what());
      } catch (...) {
        spdlog::error("Unknown exception in thread for {}", filename);
      }

      {
        std::lock_guard<std::mutex> lock(mutex_threads_);
        active_threads_--;
        spdlog::info("Thread for {} finished. Active threads: {}", filename,
                     active_threads_);
      }
      cv_threads_.notify_one();  // Notify that a thread slot is available
    });
  }

  void ProcessFile(const std::string& filename) {
    std::shared_ptr<std::queue<BatchData<T>>> queue;
    Config config;

    {
      std::lock_guard<std::mutex> lock(mutex_);
      queue = batch_queues_[filename];
      config = configs_[filename];
    }

    fs::path temp_folder = fs::temp_directory_path() / ("batch_" + filename);
    spdlog::info("Created temporary folder for writer: {}",
                 temp_folder.string());

    auto writer = std::make_shared<aifo::data::writers::DiskTileWriter>(
        temp_folder,
        aifo::data::Metadata::Create()
            ->SetMpp(config.mpp)
            ->SetGeometry(dlup::SlideGeometry{config.size, {0, 0}, config.size})
            ->SetTileSize(config.tile_size)
            ->SetTileOverlap(config.tile_overlap)
            ->SetGridOrder(aifocore::tiling::GridOrder::kF)
            ->Lock());

    auto grid_ = std::make_shared<aifocore::tiling::Grid<int>>(
        aifocore::tiling::Grid<int>::FromTiling(
            config.offset, config.size, config.tile_size, config.tile_overlap,
            aifocore::tiling::TilingMode::kOverflow,
            aifocore::tiling::GridOrder::kF));

    absl::Status status = writer->SetGrid(grid_);
    if (!status.ok()) {
      spdlog::error("Failed to set grid for {}: {}", filename,
                    status.ToString());
      return;
    }

    status = writer->Open();
    if (!status.ok()) {
      spdlog::error("Failed to open writer for {}: {}", filename,
                    status.ToString());
      return;
    }

    auto batch_generator =
        [this, queue,
         filename](int batch_index) -> std::vector<dlup::DatasetSample> {
      std::vector<dlup::DatasetSample> batch_results;

      {
        std::unique_lock<std::mutex> lock(mutex_);
        cv_.wait(lock, [&queue, this, filename] {
          return !queue->empty() || !is_running_[filename];
        });

        if (!is_running_[filename] && queue->empty()) {
          spdlog::info("No more data for {}. Exiting batch generation.",
                       filename);
          return {};  // Return an empty vector to signal the end
        }

        // Process the queue for a batch
        while (!queue->empty() &&
               batch_results.size() < static_cast<size_t>(batch_index + 1)) {
          auto batch = queue->front();
          queue->pop();

          int x = batch.coordinates(0, 0);
          int y = batch.coordinates(0, 1);

          // Convert xtensor data to vips::VImage
          std::vector<uint8_t> image_data(batch.data.size());
          std::transform(batch.data.begin(), batch.data.end(),
                         image_data.begin(), [](T val) {
                           return static_cast<uint8_t>(
                               std::clamp(static_cast<int>(val), 0, 255));
                         });

          int height = batch.data.shape(0);
          int width = batch.data.shape(1);
          int channels = batch.data.shape(2);

          auto vimage = vips::VImage::new_from_memory(
              image_data.data(), width * height * channels, width, height,
              channels, VIPS_FORMAT_UCHAR);

          batch_results.emplace_back(dlup::DatasetSample{
              std::make_shared<vips::VImage>(vimage), {x, y}, batch.filename});

          if (batch.last_batch) {
            spdlog::info("Last batch detected for {}", filename);
            break;
          }
        }
      }
      return batch_results;
    };

    // Pass the batch generator to the writer
    status = writer->Consume(batch_generator);
    if (!status.ok()) {
      spdlog::error("Failed to consume batches for {}: {}", filename,
                    status.ToString());
      return;
    }

    status = writer->Close();
    if (!status.ok()) {
      spdlog::error("Failed to close writer for {}: {}", filename,
                    status.ToString());
      return;
    }

    spdlog::info("Finished processing file: {}", filename);

    {
      std::lock_guard<std::mutex> lock(mutex_);
      is_running_[filename] = false;
      batch_queues_.erase(filename);
      configs_.erase(filename);
      spdlog::info("Deleted queue and config for: {}", filename);
    }
  }

  void StopAllThreads() {
    {
      std::lock_guard<std::mutex> lock(mutex_);
      for (auto& [filename, is_running] : is_running_) {
        is_running = false;
      }
    }
    cv_.notify_all();

    for (auto& [filename, thread] : consumer_threads_) {
      if (thread.joinable()) {
        spdlog::info("Joining thread for {}", filename);
        thread.join();
        spdlog::info("Thread for {} joined.", filename);
      }
    }
  }

  size_t max_threads_;
  size_t active_threads_;
  std::unordered_map<std::string, std::shared_ptr<std::queue<BatchData<T>>>>
      batch_queues_;
  std::unordered_map<std::string, Config> configs_;
  std::unordered_map<std::string, std::thread> consumer_threads_;
  std::unordered_map<std::string, bool> is_running_;
  std::mutex mutex_;
  std::mutex mutex_threads_;
  std::condition_variable cv_;
  std::condition_variable cv_threads_;
  std::function<Config(const std::string&)> config_callback_;
  bool callback_registered_ = false;
};

}  // namespace aifo

#endif  // AIFO_AHCORE_INCLUDE_AHCORE_DATA_WRITERS_BATCH_WRITER_H_
