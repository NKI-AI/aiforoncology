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

#ifndef AIFO_FASTSLIDE_INCLUDE_FASTSLIDE_READERS_SVS_H_
#define AIFO_FASTSLIDE_INCLUDE_FASTSLIDE_READERS_SVS_H_

#include <cstdint>
#include <filesystem>
#include <map>
#include <memory>
#include <string>
#include <utility>
#include <vector>

#include "fastslide/image.h"
#include "fastslide/readers/tiff_based_reader.h"

namespace fs = std::filesystem;

namespace fastslide {

/// @brief Pyramid level metadata for SVS
struct SvsLevelInfo {
  uint16_t page = 0;               ///< TIFF page number
  ImageDimensions size = {0, 0};   ///< Level dimensions (width, height)
  double downsample_factor = 0.0;  ///< Downsample factor relative to level 0
};

/// @brief Associated image metadata for SVS
struct SvsAssociatedInfo {
  uint16_t page;                  ///< TIFF page number
  ImageDimensions size = {0, 0};  ///< Image dimensions (width, height)
  std::string name;               ///< Image name (e.g., "thumbnail", "macro")
};

/// @brief Simple Aperio metadata parser for SVS files
struct AperioMetadata {
  std::array<double, 2> mpp;  ///< (x, y) microns per pixel
  double app_mag;             ///< Apparent magnification
  std::string scanner_id;     ///< Scanner identifier

  /// @brief Default constructor
  AperioMetadata() : mpp{0.0, 0.0}, app_mag(0.0) {}

  /// @brief Parse Aperio metadata string
  /// @param metadata_str Aperio metadata string from ImageDescription
  /// @return True if parsing succeeded
  bool ParseFromString(const std::string& metadata_str);
};

/// @brief SVS reader class implementing the SlideReader interface
class SvsReader : public TiffBasedReader {
 public:
  /// @brief Factory method to create an SvsReader instance
  /// @param filename Path to the SVS file
  /// @return StatusOr containing the reader instance or an error
  static absl::StatusOr<std::unique_ptr<SvsReader>> Create(fs::path filename);

  /// @brief Destructor
  ~SvsReader() override = default;

  // SlideReader interface implementation
  [[nodiscard]] int GetLevelCount() const override;
  [[nodiscard]] absl::StatusOr<LevelInfo> GetLevelInfo(
      int level) const override;
  [[nodiscard]] const SlideProperties& GetProperties() const override;
  [[nodiscard]] std::vector<ChannelMetadata> GetChannelMetadata()
      const override;
  [[nodiscard]] std::vector<std::string> GetAssociatedImageNames()
      const override;
  [[nodiscard]] absl::StatusOr<ImageDimensions> GetAssociatedImageDimensions(
      std::string_view name) const override;
  [[nodiscard]] absl::StatusOr<RGBImage> ReadRegion(
      const RegionSpec& region) const override;
  [[nodiscard]] absl::StatusOr<RGBImage> ReadAssociatedImage(
      std::string_view name) const override;

  [[nodiscard]] Metadata GetMetadata() const override;

  [[nodiscard]] std::string GetFormatName() const override { return "SVS"; }

  [[nodiscard]] ImageDimensions GetTileSize() const override;

  /// @brief Get Aperio metadata (format-specific)
  /// @return Reference to Aperio metadata
  [[nodiscard]] const AperioMetadata& GetAperioMetadata() const {
    return aperio_metadata_;
  }

  /// @brief Get pyramid levels (format-specific)
  /// @return Reference to pyramid levels
  [[nodiscard]] const std::vector<SvsLevelInfo>& GetPyramidLevels() const {
    return pyramid_levels_;
  }

  /// @brief Get associated images (format-specific)
  /// @return Reference to associated images
  [[nodiscard]] const std::vector<SvsAssociatedInfo>& GetAssociatedImages()
      const {
    return associated_images_;
  }

  /// @brief Parse associated image name from ImageDescription
  /// @param description ImageDescription string
  /// @return Associated image name
  static std::string ParseAssociatedImageName(const std::string& description);

 private:
  /// @brief Private constructor - use Create() factory method instead
  /// @param filename Path to the SVS file
  explicit SvsReader(fs::path filename);

  AperioMetadata aperio_metadata_;            ///< Aperio-specific metadata
  std::vector<SvsLevelInfo> pyramid_levels_;  ///< Pyramid levels
  std::vector<SvsAssociatedInfo> associated_images_;  ///< Associated images

  /// @brief Process SVS metadata and build pyramid structure
  /// @return Status indicating success or failure
  absl::Status ProcessMetadata();

  /// @brief Load level/associated image information from TIFF directories
  /// @return Status indicating success or failure
  absl::Status LoadDirectories();

  /// @brief Populate slide properties
  void PopulateSlideProperties();
};

}  // namespace fastslide

#endif  // AIFO_FASTSLIDE_INCLUDE_FASTSLIDE_READERS_SVS_H_
