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
#ifndef AIFO_AHCORE_INCLUDE_AHCORE_DATA_READERS_DISK_TILE_READER_H_
#define AIFO_AHCORE_INCLUDE_AHCORE_DATA_READERS_DISK_TILE_READER_H_

#include <vips/vips8>

#include <filesystem>
#include <memory>
#include <string>

#include "absl/status/status.h"
#include "ahcore/data/metadata.h"
#include "ahcore/data/readers/base_reader.h"
#include "aifocore/status/status_macros.h"
#include "aifocore/utilities/fmt.h"

namespace fs = std::filesystem;

namespace aifo::data::readers {

inline vips::VImage ReadTile(const fs::path& filename) {
  if (!fs::exists(filename)) {
    throw std::runtime_error(
        aifocore::fmt::format("File does not exist: {}", filename.string()));
  }
  try {
    vips::VImage image = vips::VImage::new_from_file(filename.c_str());
    return image;
  } catch (const vips::VError& e) {
    throw std::runtime_error(
        aifocore::fmt::format("VIPS error reading tile from {}: {}",
                              filename.string(), std::string(e.what())));
  } catch (const std::exception& e) {
    throw std::runtime_error(
        aifocore::fmt::format("Error reading tile from {}: {}",
                              filename.string(), std::string(e.what())));
  }
}

/**
 * @class DiskTileReader
 * @brief A reader for loading tiles and metadata from disk.
 *
 * This class reads image tiles stored in a directory, with metadata to handle
 * image dimensions, tile sizes, and other properties.
 */
class DiskTileReader : public ImageReader {
 public:
  DiskTileReader(const fs::path& directory, StitchingMode stitching_mode);

  // Add constructor that forwards post_transform to ImageReader
  DiskTileReader(const fs::path& directory, StitchingMode stitching_mode,
                 std::shared_ptr<PostStitchTransform> post_transform)
      : ImageReader(directory, stitching_mode, post_transform),
        directory_(directory),
        is_open_(false) {}

  ~DiskTileReader() {
    if (is_open_) {
      Close();
    }
  }

  absl::Status Open() override;
  void Close() override;
  absl::StatusOr<vips::VImage> ReadTile(int index) const override;

 private:
  fs::path directory_;
  bool is_open_;
};

DiskTileReader::DiskTileReader(const fs::path& directory,
                               StitchingMode stitching_mode)
    : ImageReader(directory, stitching_mode),
      directory_(directory),
      is_open_(false) {}

absl::Status DiskTileReader::Open() {
  if (!fs::exists(directory_)) {
    return MAKE_STATUS(absl::StatusCode::kNotFound,
                       "Directory does not exist: " + directory_.string());
  }

  // Load metadata from metadata.bin
  const fs::path metadata_file = directory_ / "metadata.bin";
  if (!fs::exists(metadata_file)) {
    return MAKE_STATUS(absl::StatusCode::kNotFound,
                       "Metadata file not found: " + metadata_file.string());
  }

  try {
    this->metadata_ = aifo::data::Metadata::LoadFromFile(
        metadata_file, true);  // Load from file and lock
    this->metadata_loaded_ = true;
  } catch (const std::ios::failure& e) {
    return MAKE_STATUS(absl::StatusCode::kDataLoss,
                       "Failed to read metadata file " +
                           metadata_file.string() + ": " + e.what());
  } catch (const std::exception& e) {
    return MAKE_STATUS(absl::StatusCode::kUnknown,
                       "Failed to load metadata from " +
                           metadata_file.string() + ": " + e.what());
  }

  try {
    ImageReader::ComputeParameters();
  } catch (const std::exception& e) {
    return MAKE_STATUS(
        absl::StatusCode::kInternal,
        "Failed to compute parameters from metadata: " + std::string(e.what()));
  }
  is_open_ = true;
  return absl::OkStatus();
}

void DiskTileReader::Close() {
  is_open_ = false;
  this->metadata_loaded_ = false;  // Mark metadata as no longer loaded
  // Reset metadata to default-constructed state
  // TODO(jonasteuwen): Should this be done differently?
  // this->metadata_ = aifo::data::Metadata();
}

absl::StatusOr<vips::VImage> DiskTileReader::ReadTile(int index) const {
  if (!is_open_) {
    return MAKE_STATUS(absl::StatusCode::kFailedPrecondition,
                       "Reader is not open.");
  }

  fs::path filename = directory_ / ("tile_" + std::to_string(index) + ".v");
  if (!fs::exists(filename)) {
    return MAKE_STATUS(absl::StatusCode::kNotFound,
                       aifocore::fmt::format("Tile file does not exist: {}",
                                             filename.string()));
  }

  try {
    return vips::VImage::new_from_file(filename.string().c_str());
  } catch (const vips::VError& e) {
    return MAKE_STATUS(
        absl::StatusCode::kUnknown,
        aifocore::fmt::format("VIPS error reading tile from {}: {}",
                              filename.string(), std::string(e.what())));
  }
}

}  // namespace aifo::data::readers

#endif  // AIFO_AHCORE_INCLUDE_AHCORE_DATA_READERS_DISK_TILE_READER_H_
