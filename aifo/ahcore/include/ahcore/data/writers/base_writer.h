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
#ifndef AIFO_AHCORE_INCLUDE_AHCORE_DATA_WRITERS_BASE_WRITER_H_
#define AIFO_AHCORE_INCLUDE_AHCORE_DATA_WRITERS_BASE_WRITER_H_

#include <filesystem>
#include <functional>
#include <iostream>
#include <memory>
#include <string>
#include <vector>

#include "absl/status/status.h"
#include "absl/status/statusor.h"
#include "ahcore/data/metadata.h"
#include "aifocore/status/status_macros.h"
#include "aifocore/tiling/grid.h"
#include "dlup/slide_dataset.h"

namespace tiling = aifocore::tiling;

namespace aifo::data::writers {

/**
 * @class Writer
 * @brief Abstract base class for writing tiles and metadata to storage.
 *
 * This class provides an interface for writing tiles and metadata using a defined grid.
 * Derived classes should implement the `Open`, `Close`, `WriteTile`, and `WriteMetadata` methods.
 */
class Writer {
 public:
  /**
   * @brief Default constructor for Writer.
   *
   * Initializes the metadata object.
   */
  Writer() : metadata_(Metadata::Create()) {}

  /**
   * @brief Virtual destructor for Writer.
   */
  virtual ~Writer() = default;

  /**
   * @brief Opens the writer for writing.
   *
   * This method should be implemented by derived classes to initialize resources.
   * For example, opening files or initializing connections.
   * 
   * @return absl::Status indicating success or failure.
   */
  virtual absl::Status Open() = 0;

  /**
   * @brief Closes the writer and releases resources.
   *
   * This method should be implemented by derived classes to clean up resources.
   * For example, closing files or finalizing connections.
   * 
   * @return absl::Status indicating success or failure.
   */
  virtual absl::Status Close() = 0;

  /**
   * @brief Sets the grid for the writer.
   *
   * The grid defines the tiling layout.
   *
   * @param grid A shared pointer to the grid object.
   * @return absl::Status indicating success or failure.
   */
  absl::Status SetGrid(std::shared_ptr<tiling::Grid<int>> grid) {
    // TODO(jonasteuwen): We need to also check the tile size etc from the Grid
    if (!grid) {
      return MAKE_STATUS(absl::StatusCode::kInvalidArgument,
                         "Grid cannot be null");
    }
    bool metadata_locked = metadata_->IsLocked();
    if (metadata_locked) {
      metadata_->Unlock();
    }
    // Let's check if the metadata of the grid matches the setting
    // of the writer
    if (metadata_->HasKey(MetadataKeys::GridOrder)) {
      if (metadata_->GetGridOrder() != grid->GetGridOrder()) {
        std::cerr << "Warning: Grid order mismatch between metadata and grid. "
                  << "Will overwrite" << std::endl;
      }
      GetMetadata()->SetGridOrder(grid->GetGridOrder());
    }

    if (metadata_locked) {
      metadata_->Lock();
    }
    grid_ = grid;

    tile_indices_.reserve(grid_->Length());
    for (size_t i = 0; i < grid_->Length(); ++i) {
      tile_indices_.push_back(-1);
    }
    return absl::OkStatus();
  }

  /** Sets the metadata for the writer.
   *
   * The metadata defines the properties of the data being written.
   *
   * @param metadata A unique pointer to the metadata object.
   * @return absl::Status indicating success or failure.
   */
  absl::Status SetMetadata(std::shared_ptr<Metadata> metadata) {
    if (!metadata) {
      return MAKE_STATUS(absl::StatusCode::kInvalidArgument,
                         "Metadata cannot be null.");
    }
    metadata_ = metadata;
    return absl::OkStatus();
  }

  /**
   * @brief Optionally writes metadata.
   *
   * This method should be implemented by derived classes to handle
   * writing metadata to storage.
   * 
   * @return absl::Status indicating success or failure.
   */
  virtual absl::Status WriteMetadata() = 0;

  /**
   * @brief Provides access to the metadata object.
   *
   * @return A pointer to the metadata object.
   */
  Metadata* GetMetadata();

  /**
   * @brief Gets the total size in bytes of all stored data.
   *
   * @return Size in bytes of all stored data.
   */
  virtual std::uintmax_t GetTotalSize() const = 0;

  /**
   * @brief Removes all written data
   */
  void Unlink();

 protected:
  std::shared_ptr<Metadata> metadata_;  ///< Pointer to the metadata object.
  std::shared_ptr<tiling::Grid<int>>
      grid_;  ///< Pointer to the grid defining the tiling layout.
  int num_tiles_written_{0};  ///< The number of tiles written.
  ///< The indices of the tiles actually written compared to the internal grid
  std::vector<int> tile_indices_;
};

/**
 * @brief Provides access to the metadata object.
 *
 * This method returns a raw pointer to the metadata object. It is intended
 * for use in configuring metadata.
 *
 * @return A pointer to the metadata object.
 */
Metadata* Writer::GetMetadata() {
  return metadata_.get();
}

/**
 * @class ImageWriter
 * @brief Abstract base class for writing image tiles.
 */
class ImageWriter : public Writer {
 public:
  using Writer::Writer;

  /**
   * @brief Writes a single tile to the storage medium.
   *
   * @param index The index of the tile to write.
   * @param tile The tile data as a `vips::VImage`.
   * @return absl::Status indicating success or failure.
   *
   * Derived classes should implement this method to handle the actual
   * writing of tile data to the desired storage medium.
   */
  virtual absl::Status WriteTile(int index, const vips::VImage& tile) = 0;

  /**
   * @brief Consumes dataset samples from a generator function and writes them to storage.
   *
   * @param sample_generator A function that generates `DatasetSample` for a given index.
   * @return absl::Status indicating success or failure.
   */
  absl::Status Consume(
      std::function<std::optional<dlup::DatasetSample>(int)> sample_generator) {
    if (!grid_) {
      return MAKE_STATUS(absl::StatusCode::kFailedPrecondition,
                         "Grid not set before Consume operation");
    }

    size_t grid_counter = 0;     // Counter to track progress through the grid
    size_t processed_tiles = 0;  // Counter for successfully processed tiles

    try {
      while (grid_counter < grid_->Length()) {
        // Generate dataset sample
        auto maybe_sample = sample_generator(processed_tiles);
        if (!maybe_sample) {
          // Generator signals it's done
          break;
        }

        const auto& sample = *maybe_sample;
        const auto& sample_coordinates = sample.coordinates;

        // Find matching coordinates in the grid
        while (grid_counter < grid_->Length()) {
          auto grid_coordinates = (*grid_)[grid_counter];

          // Check if the current grid coordinates match the sample coordinates
          if (grid_coordinates[0] == sample_coordinates[0] &&
              grid_coordinates[1] == sample_coordinates[1]) {
            // Assign the current index to the tile_indices_ vector
            tile_indices_[grid_counter] = processed_tiles;

            // Increment grid_counter to move
            // to the next tile for future samples
            ++grid_counter;
            break;
          }

          // Increment grid_counter if no match is found
          ++grid_counter;
        }

        // Validate and process the sample
        if (!sample.tile || sample.tile->is_null()) {
          return MAKE_STATUS(absl::StatusCode::kInvalidArgument,
                             "Null or invalid tile in DatasetSample at index " +
                                 std::to_string(processed_tiles));
        }

        // Process and write the tile from the sample
        absl::Status status =
            ProcessAndWriteTile(processed_tiles, *sample.tile);
        if (!status.ok()) {
          return MAKE_STATUS(absl::StatusCode::kInternal,
                             "Failed to process and write tile at index " +
                                 std::to_string(processed_tiles) + ": " +
                                 std::string(status.message()));
        }
        ++processed_tiles;
      }
    } catch (const vips::VError& e) {
      std::cerr << "VIPS error processing tile " << processed_tiles << ": "
                << e.what() << std::endl;
      return MAKE_STATUS(absl::StatusCode::kInternal,
                         "VIPS error processing tile " +
                             std::to_string(processed_tiles) + ": " + e.what());
    } catch (const std::exception& e) {
      std::cerr << "Error processing tile " << processed_tiles << ": "
                << e.what() << std::endl;
      return MAKE_STATUS(absl::StatusCode::kInternal,
                         "Error processing tile " +
                             std::to_string(processed_tiles) + ": " + e.what());
    } catch (...) {
      std::cerr << "Unknown error occurred during processing" << std::endl;
      return MAKE_STATUS(absl::StatusCode::kInternal,
                         "Unknown error occurred during processing");
    }

    return FinalizeConsume(processed_tiles);
  }

  /**
   * @brief Consumes batches of dataset samples from a generator function.
   *
   * This method converts a batch generator into a single sample generator internally,
   * allowing efficient processing of batched data while maintaining compatibility
   * with the single-sample Consume interface. The batch generator is called
   * incrementally as samples are needed.
   *
   * @param batch_generator Function that generates vectors of DatasetSample objects.
   *        Takes a batch index and returns a vector of samples for that batch.
   * @return absl::Status indicating success or failure.
   */
  absl::Status Consume(
      std::function<std::vector<dlup::DatasetSample>(int)> batch_generator) {
    struct BatchState {
      std::vector<dlup::DatasetSample> current_batch;
      size_t batch_index = 0;
      size_t sample_index = 0;
    };

    auto state = std::make_shared<BatchState>();

    std::function<std::optional<dlup::DatasetSample>(int)> sample_generator =
        [state,
         batch_generator](int index) -> std::optional<dlup::DatasetSample> {
      // Load first batch or next batch if needed
      if (state->current_batch.empty() ||
          state->sample_index >= state->current_batch.size()) {
        state->current_batch = batch_generator(state->batch_index++);
        state->sample_index = 0;

        if (state->current_batch.empty()) {
          return std::nullopt;  // Signal end of processing
        }
      }

      // Get current sample and advance index
      auto sample = state->current_batch[state->sample_index++];
      return sample;
    };

    return Consume(sample_generator);
  }

 private:
  /**
   * @brief Helper function to process and write a tile.
   *
   * This function processes the metadata based on the first tile and writes
   * the tile to storage.
   *
   * @param index Index of the tile.
   * @param tile The tile data as a `vips::VImage`.
   * @return absl::Status indicating success or failure.
   */
  absl::Status ProcessAndWriteTile(int index, const vips::VImage& tile) {
    if (num_tiles_written_ == 0) {
      bool is_metadata_locked = metadata_->IsLocked();
      if (is_metadata_locked) {
        metadata_->Unlock();
      }
      metadata_->Set(MetadataKeys::PixelFormat,
                     static_cast<int>(tile.format()));
      metadata_->Set(MetadataKeys::Interpretation,
                     static_cast<int>(tile.interpretation()));
      metadata_->Set(MetadataKeys::NumChannels, tile.bands());
      // metadata_->SetGridOffset(grid_[0][0], grid_[0][1]);
      if (is_metadata_locked) {
        metadata_->Lock();
      }
    }

    if (tile.is_null()) {
      return MAKE_STATUS(absl::StatusCode::kInvalidArgument,
                         "Invalid VImage at index " + std::to_string(index));
    }
    return WriteTile(index, tile);
  }

  /**
   * @brief Helper function to finalize the consume process.
   *
   * This function writes metadata and closes the writer.
   *
   * @param num_tiles The total number of tiles processed.
   * @return absl::Status indicating success or failure.
   */
  absl::Status FinalizeConsume(size_t num_tiles) {
    num_tiles_written_ = num_tiles;
    // Let's set internal indices to the metadata
    bool is_metadata_locked = metadata_->IsLocked();
    if (is_metadata_locked) {
      metadata_->Unlock();
    }
    metadata_->Set(MetadataKeys::TileIndices, tile_indices_);
    if (is_metadata_locked) {
      metadata_->Lock();
    }
    absl::Status status = WriteMetadata();
    if (!status.ok()) {
      return status;
    }
    return Close();
  }
};

/**
 * @class FeatureWriter
 * @brief Abstract base class for writing numerical features.
 */
class FeatureWriter : public Writer {
 public:
  using Writer::Writer;

  /**
   * @brief Writes a single feature array.
   *
   * @param index Index of the feature.
   * @param feature The feature data as an `xt::xarray`.
   */
  // virtual void WriteFeature(int index, const xt::xarray<float>& feature) = 0;
};

}  // namespace aifo::data::writers

#endif  // AIFO_AHCORE_INCLUDE_AHCORE_DATA_WRITERS_BASE_WRITER_H_
