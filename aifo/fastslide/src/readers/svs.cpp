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

#include "fastslide/readers/svs.h"

#include <tiffio.h>
#include <algorithm>
#include <cstdint>
#include <map>
#include <memory>
#include <string>
#include <utility>
#include <vector>

#include "absl/status/status.h"
#include "absl/strings/numbers.h"
#include "absl/strings/str_split.h"
#include "aifocore/status/status_macros.h"
#include "aifocore/utilities/fmt.h"

#include "fastslide/utilities/tiff/directory_processor.h"
#include "fastslide/utilities/tiff/tiff_file.h"

namespace fs = std::filesystem;

namespace fastslide {

namespace {

/// @brief Extracts Aperio metadata from image description
/// @param description Image description string
/// @param metadata Output metadata structure
/// @return Status indicating success or failure
absl::Status ExtractAperioMetadata(const std::string& description,
                                   AperioMetadata& metadata) {
  if (description.find("Aperio") == std::string::npos) {
    return MAKE_STATUS(absl::StatusCode::kInvalidArgument,
                       "Not an Aperio SVS file: missing Aperio signature");
  }

  // Split by "|" to get key-value pairs
  std::vector<std::string> parts;
  for (const auto& part : absl::StrSplit(description, '|')) {
    parts.push_back(std::string(part));
  }

  bool found_any_metadata = false;

  // Parse key=value pairs
  for (const auto& part : parts) {
    std::vector<std::string> kv;
    for (const auto& piece : absl::StrSplit(part, '=')) {
      kv.push_back(std::string(piece));
    }
    if (kv.size() != 2)
      continue;

    std::string key = kv[0];
    std::string value = kv[1];

    // Trim whitespace
    key.erase(key.find_last_not_of(" \t\r\n") + 1);
    key.erase(0, key.find_first_not_of(" \t\r\n"));
    value.erase(value.find_last_not_of(" \t\r\n") + 1);
    value.erase(0, value.find_first_not_of(" \t\r\n"));

    if (key == "MPP") {
      double mpp_val;
      if (absl::SimpleAtod(value, &mpp_val)) {
        metadata.mpp = {mpp_val, mpp_val};
        found_any_metadata = true;
      }
    } else if (key == "AppMag") {
      if (absl::SimpleAtod(value, &metadata.app_mag)) {
        found_any_metadata = true;
      }
    } else if (key == "ScanScope ID") {
      metadata.scanner_id = value;
      found_any_metadata = true;
    }
  }

  if (!found_any_metadata) {
    return MAKE_STATUS(absl::StatusCode::kNotFound,
                       "No Aperio metadata found in image description");
  }

  return absl::OkStatus();
}

/// @brief Callback class for SVS-specific directory processing
class SvsDirectoryCallback : public TiffDirectoryCallback {
 public:
  SvsDirectoryCallback(std::vector<SvsLevelInfo>& pyramid_levels,
                       std::vector<SvsAssociatedInfo>& associated_images,
                       AperioMetadata& metadata)
      : pyramid_levels_(pyramid_levels),
        associated_images_(associated_images),
        metadata_(metadata),
        metadata_extracted_(false) {}

  absl::Status ProcessDirectory(
      const TiffDirectoryInfoWithPage& dir_info) override {
    const auto& image_dims = dir_info.info.image_dims;
    if (image_dims[0] == 0 || image_dims[1] == 0) {
      return MAKE_STATUS(
          absl::StatusCode::kInvalidArgument,
          aifocore::fmt::format("Invalid image dimensions: {}x{}",
                                image_dims[0], image_dims[1]));
    }

    // Extract metadata from first directory
    if (!metadata_extracted_ && dir_info.page == 0) {
      if (dir_info.info.image_description.has_value() &&
          !dir_info.info.image_description->empty()) {
        auto status =
            ExtractAperioMetadata(*dir_info.info.image_description, metadata_);
        if (status.ok()) {
          metadata_extracted_ = true;
        }
        // Continue even if metadata extraction fails for this directory
      }
    }

    if (dir_info.info.is_tiled) {
      // Store tiled directory info for later processing
      TiledDirectoryInfo tiled_info;
      tiled_info.page = dir_info.page;
      tiled_info.size = {image_dims[0], image_dims[1]};
      tiled_info.area = static_cast<uint64_t>(image_dims[0]) * image_dims[1];

      tiled_directories_.push_back(tiled_info);
    } else {
      // Non-tiled → associated image
      std::string name;
      if (dir_info.page == 1) {
        name = "thumbnail";
      } else {
        // Parse name from ImageDescription
        if (dir_info.info.image_description.has_value() &&
            !dir_info.info.image_description->empty()) {
          name = SvsReader::ParseAssociatedImageName(
              *dir_info.info.image_description);
        }
        if (name.empty()) {
          name = "unknown";
        }
      }
      associated_images_.push_back(
          SvsAssociatedInfo{.page = dir_info.page,
                            .size = {image_dims[0], image_dims[1]},
                            .name = name});
    }

    return absl::OkStatus();
  }

  absl::Status Finalize() override {

    // Sort tiled directories by area (largest first)
    std::sort(tiled_directories_.begin(), tiled_directories_.end(),
              [](const TiledDirectoryInfo& a, const TiledDirectoryInfo& b) {
                return a.area > b.area;
              });

    // Convert to pyramid levels with proper downsample factors
    pyramid_levels_.clear();
    pyramid_levels_.reserve(tiled_directories_.size());

    if (tiled_directories_.empty()) {
      return absl::OkStatus();
    }

    // First (largest) becomes level 0
    const auto& level0 = tiled_directories_[0];
    pyramid_levels_.push_back(
        SvsLevelInfo{.page = level0.page,
                     .size = {level0.size[0], level0.size[1]},
                     .downsample_factor = 1.0});

    // Calculate downsample factors for remaining levels
    for (size_t i = 1; i < tiled_directories_.size(); ++i) {
      const auto& level = tiled_directories_[i];

      // Calculate downsample as average of width and height ratios
      double downsample = (static_cast<double>(level0.size[0]) /
                               static_cast<double>(level.size[0]) +
                           static_cast<double>(level0.size[1]) /
                               static_cast<double>(level.size[1])) /
                          2.0;

      pyramid_levels_.push_back(
          SvsLevelInfo{.page = level.page,
                       .size = {level.size[0], level.size[1]},
                       .downsample_factor = downsample});
    }

    return absl::OkStatus();
  }

 private:
  struct TiledDirectoryInfo {
    uint16_t page;
    std::array<uint32_t, 2> size;
    uint64_t area;
  };

  std::vector<SvsLevelInfo>& pyramid_levels_;
  std::vector<SvsAssociatedInfo>& associated_images_;
  AperioMetadata& metadata_;
  bool metadata_extracted_;
  std::vector<TiledDirectoryInfo> tiled_directories_;
};

}  // namespace

// Simple Aperio metadata parser - just extract key fields
bool AperioMetadata::ParseFromString(const std::string& metadata_str) {
  AperioMetadata temp_metadata;
  auto status = ExtractAperioMetadata(metadata_str, temp_metadata);
  if (status.ok()) {
    *this = temp_metadata;
    return true;
  }
  return false;
}

// SvsReader implementation
absl::StatusOr<std::unique_ptr<SvsReader>> SvsReader::Create(
    fs::path filename) {
  auto reader = std::unique_ptr<SvsReader>(new SvsReader(filename));

  RETURN_IF_ERROR(ValidateTiffFile(filename), "Failed to validate TIFF file");

  RETURN_IF_ERROR(reader->InitializeHandlePool(),
                  "Failed to initialize handle pool");

  RETURN_IF_ERROR(reader->ProcessMetadata(), "Failed to process metadata");

  reader->PopulateSlideProperties();

  return reader;
}

SvsReader::SvsReader(fs::path filename) : TiffBasedReader(filename) {
  // Constructor is now private - initialization is done in Create()
}

int SvsReader::GetLevelCount() const {
  return static_cast<int>(pyramid_levels_.size());
}

absl::StatusOr<LevelInfo> SvsReader::GetLevelInfo(int level) const {
  if (level < 0 || level >= static_cast<int>(pyramid_levels_.size())) {
    return MAKE_STATUSOR(LevelInfo, absl::StatusCode::kNotFound,
                         aifocore::fmt::format("Level {} not found", level));
  }

  const auto& svs_level = pyramid_levels_[level];
  LevelInfo level_info;
  level_info.dimensions = {svs_level.size[0], svs_level.size[1]};
  level_info.downsample_factor = svs_level.downsample_factor;

  return level_info;
}

const SlideProperties& SvsReader::GetProperties() const {
  return properties_;
}

std::vector<ChannelMetadata> SvsReader::GetChannelMetadata() const {
  // SVS files typically have RGB channels
  std::vector<ChannelMetadata> metadata;
  metadata.emplace_back("RGB", "Histological stain", ColorRGB{255, 255, 255});
  return metadata;
}

std::vector<std::string> SvsReader::GetAssociatedImageNames() const {
  std::vector<std::string> names;
  names.reserve(associated_images_.size());
  for (const auto& img : associated_images_) {
    names.push_back(img.name);
  }
  return names;
}

absl::StatusOr<ImageDimensions> SvsReader::GetAssociatedImageDimensions(
    std::string_view name) const {
  for (const auto& img : associated_images_) {
    if (img.name == name) {
      return ImageDimensions{img.size[0], img.size[1]};
    }
  }
  return MAKE_STATUSOR(
      ImageDimensions, absl::StatusCode::kNotFound,
      aifocore::fmt::format("Associated image '{}' not found", name));
}

absl::StatusOr<RGBImage> SvsReader::ReadRegion(const RegionSpec& region) const {
  if (!region.IsValid()) {
    return MAKE_STATUSOR(RGBImage, absl::StatusCode::kInvalidArgument,
                         "Invalid region specification");
  }

  if (region.level < 0 ||
      region.level >= static_cast<int>(pyramid_levels_.size())) {
    return MAKE_STATUSOR(
        RGBImage, absl::StatusCode::kInvalidArgument,
        aifocore::fmt::format("Invalid level: {}", region.level));
  }

  const auto& level_info = pyramid_levels_[region.level];

  // Validate that the region is within level bounds
  if (region.top_left[0] >= level_info.size[0] ||
      region.top_left[1] >= level_info.size[1]) {
    return MAKE_STATUSOR(
        RGBImage, absl::StatusCode::kOutOfRange,
        aifocore::fmt::format(
            "Region top-left ({}, {}) is outside level {} bounds ({}x{})",
            region.top_left[0], region.top_left[1], region.level,
            level_info.size[0], level_info.size[1]));
  }

  // Validate that the region size fits within the level
  if (region.top_left[0] + region.size[0] > level_info.size[0] ||
      region.top_left[1] + region.size[1] > level_info.size[1]) {
    return MAKE_STATUSOR(
        RGBImage, absl::StatusCode::kOutOfRange,
        aifocore::fmt::format(
            "Region ({}x{} at {},{}) extends beyond level {} bounds ({}x{})",
            region.size[0], region.size[1], region.top_left[0],
            region.top_left[1], region.level, level_info.size[0],
            level_info.size[1]));
  }

  uint32_t actual_width;
  uint32_t actual_height;
  auto data =
      ReadTiffRegion(level_info.page, region, actual_width, actual_height, 3);

  if (data.empty()) {
    return MAKE_STATUSOR(RGBImage, absl::StatusCode::kInternal,
                         "Failed to read region from TIFF");
  }

  RGBImage rgb_image(ImageDimensions{actual_width, actual_height},
                     ImageFormat::kRGB, DataType::kUInt8);
  std::ranges::copy(data, rgb_image.GetDataVector().begin());

  return rgb_image;
}

absl::StatusOr<RGBImage> SvsReader::ReadAssociatedImage(
    std::string_view name) const {
  const SvsAssociatedInfo* info = nullptr;
  for (const auto& img : associated_images_) {
    if (img.name == name) {
      info = &img;
      break;
    }
  }

  if (!info) {
    return MAKE_STATUSOR(
        RGBImage, absl::StatusCode::kNotFound,
        aifocore::fmt::format("Associated image '{}' not found", name));
  }

  return ReadAssociatedImageFromPage(info->page, info->size[0], info->size[1],
                                     std::string(name));
}

// GetBestLevelForDownsample uses the base class implementation

ImageDimensions SvsReader::GetTileSize() const {
  // Try to get tile size from level 0
  if (pyramid_levels_.empty()) {
    return ImageDimensions{256, 256};  // Default for SVS
  }

  // Get tile dimensions from the first level
  auto tiff_file_result = TiffFile::Create(handle_pool_.get());
  if (!tiff_file_result.ok()) {
    return ImageDimensions{256, 256};  // Default fallback
  }
  auto tiff_file = std::move(tiff_file_result.value());

  auto status = tiff_file.SetDirectory(pyramid_levels_[0].page);
  if (!status.ok()) {
    return ImageDimensions{256, 256};  // Default fallback
  }

  if (tiff_file.IsTiled()) {
    auto tile_dims_result = tiff_file.GetTileDimensions();
    if (tile_dims_result.ok()) {
      const auto& dims = tile_dims_result.value();
      return ImageDimensions{dims[0], dims[1]};
    }
  }

  return ImageDimensions{256, 256};  // Default for SVS
}

Metadata SvsReader::GetMetadata() const {
  std::map<std::string, std::variant<std::string, size_t, double>> metadata;
  metadata["format"] = "SVS";
  metadata["mpp_x"] = aperio_metadata_.mpp[0];
  metadata["mpp_y"] = aperio_metadata_.mpp[1];
  metadata["app_mag"] = aperio_metadata_.app_mag;
  metadata["scanner_id"] = aperio_metadata_.scanner_id;
  metadata["levels"] = pyramid_levels_.size();
  metadata["associated_images"] = associated_images_.size();
  return Metadata(metadata);
}

absl::Status SvsReader::ProcessMetadata() {
  // Load directories and extract basic information
  RETURN_IF_ERROR(LoadDirectories(), "Failed to load TIFF directories");

  return absl::OkStatus();
}

absl::Status SvsReader::LoadDirectories() {
  // Create TiffFile wrapper using the handle pool
  auto tiff_file_result = TiffFile::Create(handle_pool_.get());
  if (!tiff_file_result.ok()) {
    return MAKE_STATUS(
        absl::StatusCode::kInternal,
        aifocore::fmt::format("Failed to create TiffFile wrapper: {}",
                              tiff_file_result.status().message()));
  }
  auto tiff_file = std::move(tiff_file_result.value());

  pyramid_levels_.clear();
  associated_images_.clear();

  // Create directory processor and callback
  TiffDirectoryProcessor processor(std::move(tiff_file));
  SvsDirectoryCallback callback(pyramid_levels_, associated_images_,
                                aperio_metadata_);

  // Process all directories using the common processor
  RETURN_IF_ERROR(processor.ProcessAllDirectories(callback),
                  "Failed to process TIFF directories");

  return absl::OkStatus();
}

void SvsReader::PopulateSlideProperties() {
  properties_.mpp = aifocore::Size<double, 2>{aperio_metadata_.mpp[0],
                                              aperio_metadata_.mpp[1]};
  properties_.objective_magnification = aperio_metadata_.app_mag;
  properties_.objective_name =
      aifocore::fmt::format("{}x", aperio_metadata_.app_mag);
  properties_.scanner_model =
      aifocore::fmt::format("Aperio/{}", aperio_metadata_.scanner_id);
}

std::string SvsReader::ParseAssociatedImageName(
    const std::string& description) {
  // Parse name from ImageDescription following OpenSlide approach:
  // 1. Split into lines at newlines
  // 2. Take second line (index 1)
  // 3. Split on spaces
  // 4. Take first word

  std::vector<std::string> lines;
  for (const auto& line :
       absl::StrSplit(description, absl::ByAnyChar("\r\n"))) {
    lines.push_back(std::string(line));
  }

  if (lines.size() < 2) {
    return "macro";  // Default fallback
  }

  std::string second_line = lines[1];
  std::vector<std::string> words;
  for (const auto& word : absl::StrSplit(second_line, ' ')) {
    words.push_back(std::string(word));
  }

  if (words.empty() || words[0].empty()) {
    return "unknown";  // Default fallback
  }

  return words[0];
}

}  // namespace fastslide
