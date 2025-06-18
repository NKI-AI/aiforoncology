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
#ifndef AIFO_AHCORE_INCLUDE_AHCORE_DATA_WRITERS_DISK_TILE_WRITER_H_
#define AIFO_AHCORE_INCLUDE_AHCORE_DATA_WRITERS_DISK_TILE_WRITER_H_

#include <filesystem>
#include <fstream>
#include <iostream>
#include <memory>
#include "ahcore/data/writers/base_writer.h"
#include "aifo/aifocore/include/aifocore/status/status_macros.h"

namespace aifo::data::writers {

namespace fs = std::filesystem;

/**
 * @class DiskTileWriter
 * @brief A writer for storing tiles and metadata to disk.
 *
 * This class implements the `Writer` interface and provides functionality
 * for writing image tiles to a temporary folder on disk. Metadata is serialized
 * and saved in a binary format.
 */
class DiskTileWriter : public ImageWriter {
 public:
  /**
   * @brief Constructs a `DiskTileWriter` with a specified temporary folder.
   *
   * @param temp_folder The folder where tiles and metadata will be stored.
   * @param metadata A shared pointer to metadata (optional).
   */
  inline DiskTileWriter(const fs::path& temp_folder,
                        std::shared_ptr<Metadata> metadata = nullptr)
      : ImageWriter(),
        temp_folder_(temp_folder),
        is_open_(false),
        tile_count_(0) {
    if (metadata) {
      absl::Status status = SetMetadata(metadata);
      if (!status.ok()) {
        // In a constructor we can't return a status, so we'll log the error
        std::cerr << "Error setting metadata: " << status.message()
                  << std::endl;
      }
    } else {
      metadata_ = Metadata::Create();
    }
    metadata_->Lock();  // Lock metadata to prevent further modifications
  }

  /**
   * @brief Destructor for `DiskTileWriter`.
   *
   * Ensures that resources are released and the writer is closed if open.
   */
  ~DiskTileWriter();

  /**
   * @brief Opens the writer for writing tiles and metadata.
   *
   * Creates the temporary folder if it does not exist.
   *
   * @return absl::Status indicating success or failure.
   */
  absl::Status Open() override;

  /**
   * @brief Closes the writer and releases any resources.
   * 
   * @return absl::Status indicating success or failure.
   */
  absl::Status Close() override;

  /**
   * @brief Writes a tile to the specified temporary folder.
   *
   * Each tile is saved as a VIPS native format file.
   *
   * @param index The index of the tile to write.
   * @param tile The tile data as a `vips::VImage`.
   * @return absl::Status indicating success or failure.
   */
  absl::Status WriteTile(int index, const vips::VImage& tile) override;

  /**
   * @brief Writes metadata to a binary file in the temporary folder.
   *
   * Metadata is serialized and stored in `metadata.bin`.
   *
   * @return absl::Status indicating success or failure.
   */
  absl::Status WriteMetadata() override;

  /**
   * @brief Removes the files which are written to the temporary folder.
   */

  void Unlink() {
    fs::remove(temp_folder_ / "metadata.bin");
    for (int i = 0; i < tile_count_; i++) {
      fs::remove(temp_folder_ / ("tile_" + std::to_string(i) + ".v"));
    }
  }

  /**
   * @brief Calculates the total size in bytes of all stored data on disk.
   *
   * This includes all tile files and the metadata file.
   *
   * @return Size in bytes of all stored data.
   */
  std::uintmax_t GetTotalSize() const override {
    std::uintmax_t total_size = 0;

    // Add size of metadata file if it exists
    fs::path metadata_path = temp_folder_ / "metadata.bin";
    if (fs::exists(metadata_path)) {
      total_size += fs::file_size(metadata_path);
    }

    // Get size of one tile and multiply by total number of tiles
    if (tile_count_ > 0) {
      fs::path first_tile_path = temp_folder_ / "tile_0.v";
      if (fs::exists(first_tile_path)) {
        total_size += fs::file_size(first_tile_path) * tile_count_;
      }
    }

    return total_size;
  }

 private:
  fs::path temp_folder_;  ///< The folder where tiles and metadata are stored.
  bool is_open_;          ///< Indicates whether the writer is open.
  int tile_count_;        ///< The number of tiles written.
};

inline DiskTileWriter::~DiskTileWriter() {
  if (is_open_) {
    absl::Status status = Close();
    if (!status.ok()) {
      // Cannot throw in destructor, so just log the error
      std::cerr << "Error closing writer in destructor: " << status.message()
                << std::endl;
    }
  }
}

inline absl::Status DiskTileWriter::Open() {
  if (!fs::exists(temp_folder_)) {
    fs::create_directories(temp_folder_);
  }
  is_open_ = true;
  return absl::OkStatus();
}

inline absl::Status DiskTileWriter::Close() {
  is_open_ = false;
  return absl::OkStatus();
}

inline absl::Status DiskTileWriter::WriteTile(int index,
                                              const vips::VImage& tile) {
  if (!is_open_) {
    return MAKE_STATUS(absl::StatusCode::kFailedPrecondition,
                       "Writer is not open");
  }

  fs::path filename = temp_folder_ / ("tile_" + std::to_string(index) + ".v");

  try {
    // Write the tile directly to file in VIPS native format
    tile.write_to_file(filename.c_str());
  } catch (const vips::VError& e) {
    return MAKE_STATUS(
        absl::StatusCode::kUnknown,
        "Failed to write tile " + std::to_string(index) + ": " + e.what());
  } catch (const std::exception& e) {
    std::cerr << "Standard error in WriteTile: " << e.what() << std::endl;
    return MAKE_STATUS(absl::StatusCode::kUnknown,
                       "Standard error in WriteTile");
  } catch (...) {
    std::cerr << "Unknown error in WriteTile" << std::endl;
    return MAKE_STATUS(absl::StatusCode::kUnknown,
                       "Unknown error in WriteTile");
  }
  tile_count_++;
  return absl::OkStatus();
}

inline absl::Status DiskTileWriter::WriteMetadata() {
  if (!metadata_) {
    std::cerr << "No metadata available to write." << std::endl;
    return absl::OkStatus();
  }

  fs::path metadata_filename = temp_folder_ / "metadata.bin";

  try {
    metadata_->Serialize(metadata_filename);
  } catch (const std::exception& e) {
    std::cerr << "Error writing metadata: " << e.what() << std::endl;
    return MAKE_STATUS(absl::StatusCode::kUnknown, "Error writing metadata");
  }
  return absl::OkStatus();
}

}  // namespace aifo::data::writers

#endif  // AIFO_AHCORE_INCLUDE_AHCORE_DATA_WRITERS_DISK_TILE_WRITER_H_
