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

#include "fastslide/readers/qptiff.h"

#include <tiffio.h>

#include <algorithm>
#include <cmath>
#include <cstdint>
#include <map>
#include <memory>
#include <numeric>
#include <ranges>
#include <sstream>
#include <string>
#include <utility>
#include <vector>

#include <pugixml.hpp>

#include "absl/log/log.h"
#include "absl/status/status.h"
#include "aifocore/concepts/numeric.h"
#include "aifocore/status/status_macros.h"
#include "aifocore/utilities/fmt.h"
#include "fastslide/slide_reader.h"
#include "fastslide/utilities/colors.h"
#include "fastslide/utilities/scoped_timer.h"

namespace fastslide {

using aifocore::Size;

// TODO(jonasteuwen): Move to color utilities
absl::StatusOr<std::array<uint8_t, 3>> ParseRgb(const std::string& s) {
  std::array<uint8_t, 3> c{};
  std::istringstream ss(s);
  for (auto& x : c) {
    int v = 0;
    ss >> v;
    if (ss.peek() == ',')
      ss.ignore();
    if (v < 0 || v > 255) {
      return MAKE_STATUS(absl::StatusCode::kInvalidArgument,
                         "Invalid RGB value: " + std::to_string(v));
    }
    x = static_cast<uint8_t>(v);
  }
  return c;
}

namespace {

/// @brief Extracts resolution information from TIFF tags (PAGE 0 ONLY!)
/// @param tiff_file TiffFile instance positioned at page 0
/// @param metadata Output metadata structure
/// @param root XML root node (optional, for validation and additional metadata)
/// @return Status indicating success or failure
absl::Status ExtractResolutionMetadata(TiffFile& tiff_file,
                                       SlideMetadata& metadata,
                                       const pugi::xml_node* root = nullptr) {
  // CRITICAL: Extract MPP from TIFF tags, not XML!
  auto x_res = tiff_file.GetXResolution();
  auto y_res = tiff_file.GetYResolution();
  uint16_t res_unit = tiff_file.GetResolutionUnit();

  if (!x_res.has_value() || !y_res.has_value()) {
    return MAKE_STATUS(absl::StatusCode::kNotFound,
                       "Missing resolution information in TIFF tags - no X/Y "
                       "resolution or resolution unit found");
  }

  // Convert resolution to microns per pixel
  double mpp_x = 0.0;
  double mpp_y = 0.0;
  switch (res_unit) {
    case RESUNIT_INCH:
      // Convert pixels per inch to microns per pixel
      mpp_x = 25400.0 / x_res.value();  // 25400 microns per inch
      mpp_y = 25400.0 / y_res.value();
      break;
    case RESUNIT_CENTIMETER:
      // Convert pixels per centimeter to microns per pixel
      mpp_x = 10000.0 / x_res.value();  // 10000 microns per cm
      mpp_y = 10000.0 / y_res.value();
      break;
    default:
      return MAKE_STATUS(
          absl::StatusCode::kInvalidArgument,
          "Unsupported resolution unit: " + std::to_string(res_unit));
  }

  // Validate if they are isotropic enough (less than 1% relative difference)
  if (std::abs(mpp_x - mpp_y) / std::max(mpp_x, mpp_y) > 0.01 || mpp_x <= 0.0 ||
      mpp_y <= 0.0) {
    return MAKE_STATUS(
        absl::StatusCode::kInvalidArgument,
        "Computed MPP values are not isotropic enough or not positive: " +
            std::to_string(mpp_x) + ", " + std::to_string(mpp_y) + " µm/px");
  }

  metadata.mpp_x = mpp_x;
  metadata.mpp_y = mpp_y;

  // Optionally validate against XML metadata if present
  if (root != nullptr) {
    auto resolution_node = root->child("ScanProfile").child("root");

    if (!resolution_node.empty()) {
      auto pixel_size_node = resolution_node.child("PixelSizeMicrons");
      if (!pixel_size_node.empty()) {
        double xml_pixel_size = pixel_size_node.text().as_double();
        // Allow 5% tolerance for floating point comparison
        // For both x and y
        double tolerance = 0.05 * (mpp_x + mpp_y) / 2.0;
        if (std::abs(mpp_y - xml_pixel_size) > tolerance ||
            std::abs(mpp_x - xml_pixel_size) > tolerance) {
          LOG(WARNING) << "TIFF resolution doesn't match XML resolution - "
                       << "TIFF: " << mpp_y << " µm/px, XML: " << xml_pixel_size
                       << " µm/px (tolerance: " << tolerance << ")";
          // Don't fail on this - just log the discrepancy
        }
      }
    }

    // Extract magnification from XML (not available in TIFF tags)
    if (!resolution_node.empty()) {
      auto magnification_node = resolution_node.child("Magnification");
      if (!magnification_node.empty()) {
        metadata.magnification = magnification_node.text().as_double();
      }

      // Extract objective name from XML (not available in TIFF tags)
      auto objective_node = resolution_node.child("ObjectiveName");
      if (!objective_node.empty()) {
        metadata.objective_name = objective_node.text().as_string();
      }
    }
  }

  return absl::OkStatus();
}

/// @brief Extract metadata from XML for a channel
/// @param root XML root node
/// @param channel_index Index of the channel for default naming
/// @return Channel info or error
absl::StatusOr<QpTiffChannelInfo> ExtractChannelFromXml(
    const pugi::xml_node& root, int channel_index) {
  QpTiffChannelInfo channel;

  // Extract basic channel information
  std::string name = QpTiffReader::GetText(&root, "Name");
  channel.name = name;

  // Extract biomarker information
  std::string biomarker = QpTiffReader::GetText(&root, "Biomarker");
  if (biomarker.empty()) {
    channel.biomarker =
        "Unknown Biomarker " + std::to_string(channel_index + 1);
  } else {
    channel.biomarker = biomarker;
  }

  // Extract exposure time
  std::string exposure_str = QpTiffReader::GetText(&root, "ExposureTime");
  channel.exposure_time = exposure_str.empty() ? 0 : std::stoul(exposure_str);

  // Extract signal units
  std::string signal_units_str = QpTiffReader::GetText(&root, "SignalUnits");
  channel.signal_units =
      signal_units_str.empty() ? 0 : std::stoul(signal_units_str);

  // Extract color
  std::string color_str = QpTiffReader::GetText(&root, "Color");
  if (!color_str.empty()) {
    auto rgb_array = ParseRgb(color_str);
    if (!rgb_array.ok()) {
      return rgb_array.status();
    }
    channel.color = ColorRGB(rgb_array.value()[0], rgb_array.value()[1],
                             rgb_array.value()[2]);
  } else {
    channel.color = GetDefaultChannelColor(channel_index + 1);
  }

  return channel;
}

}  // namespace

absl::StatusOr<std::unique_ptr<QpTiffReader>> QpTiffReader::Create(
    std::string_view filename) {
  ScopedTimer t_create("QpTiffReader::Create (total)");
  auto reader = std::unique_ptr<QpTiffReader>(new QpTiffReader(filename));

  {
    ScopedTimer t_validate("QpTiffReader::Create - ValidateTiffFile");
    RETURN_IF_ERROR(ValidateTiffFile(filename), "Failed to validate TIFF file");
  }
  {
    ScopedTimer t_pool("QpTiffReader::Create - InitializeHandlePool");
    RETURN_IF_ERROR(reader->InitializeHandlePool(),
                    "Failed to initialize handle pool");
  }
  {
    ScopedTimer t_meta("QpTiffReader::Create - ProcessMetadata (header parse)");
    RETURN_IF_ERROR(reader->ProcessMetadata(), "Failed to process metadata");
  }

  reader->PopulateSlideProperties();

  return reader;
}

QpTiffReader::QpTiffReader(std::string_view filename)
    : TiffBasedReader(std::string(filename)) {
  // Constructor is now private - initialization is done in Create()
}

// SlideReader interface implementations
int QpTiffReader::GetLevelCount() const {
  return static_cast<int>(pyramid_.size());
}

absl::StatusOr<LevelInfo> QpTiffReader::GetLevelInfo(int level) const {
  if (level < 0) {
    return absl::StatusOr<LevelInfo>(MAKE_STATUS(
        absl::StatusCode::kInvalidArgument, "Level cannot be negative"));
  }

  if (level < 0 || static_cast<size_t>(level) >= pyramid_.size()) {
    return absl::StatusOr<LevelInfo>(
        MAKE_STATUS(absl::StatusCode::kNotFound,
                    aifocore::fmt::format("Level {} not found", level)));
  }

  const auto& pyramid_level = pyramid_[level];
  LevelInfo level_info;
  level_info.dimensions = pyramid_level.size;

  // Calculate downsample factor relative to level 0
  // This should be consistent with openslide.
  if (level == 0) {
    level_info.downsample_factor = 1.0;
  } else {
    if (!pyramid_.empty()) {
      Size<double, 2> proportion =
          static_cast<Size<double, 2>>(pyramid_[0].size) /
          static_cast<Size<double, 2>>(pyramid_level.size);
      level_info.downsample_factor = (proportion[0] + proportion[1]) / 2.0;
    }
  }

  return level_info;
}

const SlideProperties& QpTiffReader::GetProperties() const {
  return properties_;
}

std::vector<ChannelMetadata> QpTiffReader::GetChannelMetadata() const {
  std::vector<ChannelMetadata> metadata;
  metadata.reserve(channels_.size());

  for (const auto& ch : channels_) {
    ChannelMetadata md;
    md.name = ch.name;
    md.biomarker = ch.biomarker;
    md.color = ch.color;
    md.exposure_time = ch.exposure_time;
    md.signal_units = ch.signal_units;
    metadata.push_back(std::move(md));
  }

  return metadata;
}

std::vector<std::string> QpTiffReader::GetAssociatedImageNames() const {
  std::vector<std::string> names;
  names.reserve(associated_images_.size());
  for (const auto& [name, info] : associated_images_) {
    names.push_back(name);
  }
  return names;
}

absl::StatusOr<ImageDimensions> QpTiffReader::GetAssociatedImageDimensions(
    std::string_view name) const {
  if (!associated_images_.contains(std::string(name))) {
    return absl::StatusOr<ImageDimensions>(MAKE_STATUS(
        absl::StatusCode::kNotFound,
        aifocore::fmt::format("Associated image '{}' not found", name)));
  }

  const auto& info = associated_images_.at(std::string(name));
  return ImageDimensions{info.size[0], info.size[1]};
}

absl::StatusOr<Image> QpTiffReader::ReadRegion(const RegionSpec& region) const {
  if (!region.IsValid()) {
    return MAKE_STATUSOR(Image, absl::StatusCode::kInvalidArgument,
                         "Invalid region specification");
  }

  if (region.level < 0 ||
      static_cast<size_t>(region.level) >= pyramid_.size()) {
    return MAKE_STATUSOR(
        Image, absl::StatusCode::kNotFound,
        "Level " + std::to_string(region.level) + " not found");
  }

  const QpTiffLevelInfo& level_info = pyramid_[region.level];

  // Validate that the region is within level bounds
  if (region.top_left[0] >= level_info.size[0] ||
      region.top_left[1] >= level_info.size[1]) {
    return MAKE_STATUSOR(
        Image, absl::StatusCode::kOutOfRange,
        "Region top-left (" + std::to_string(region.top_left[0]) + ", " +
            std::to_string(region.top_left[1]) + ") is outside level " +
            std::to_string(region.level) + " bounds (" +
            std::to_string(level_info.size[0]) + "x" +
            std::to_string(level_info.size[1]) + ")");
  }

  // Validate that the region size fits within the level
  if (region.top_left[0] + region.size[0] > level_info.size[0] ||
      region.top_left[1] + region.size[1] > level_info.size[1]) {
    return MAKE_STATUSOR(Image, absl::StatusCode::kOutOfRange,
                         "Region (" + std::to_string(region.size[0]) + "x" +
                             std::to_string(region.size[1]) + " at " +
                             std::to_string(region.top_left[0]) + "," +
                             std::to_string(region.top_left[1]) +
                             ") extends beyond level " +
                             std::to_string(region.level) + " bounds (" +
                             std::to_string(level_info.size[0]) + "x" +
                             std::to_string(level_info.size[1]) + ")");
  }

  // Create TiffFile wrapper once and reuse it for all channels
  auto tiff_file_result = TiffFile::Create(handle_pool_.get());
  if (!tiff_file_result.ok()) {
    return MAKE_STATUSOR(Image, absl::StatusCode::kInternal,
                         "Failed to create TiffFile wrapper");
  }
  auto tiff_file = std::move(tiff_file_result.value());

  // Determine which channels to load based on visibility settings
  std::vector<size_t> channels_to_load;
  if (visible_channels_.empty()) {
    // Load all channels if none specifically set
    channels_to_load.resize(level_info.pages.size());
    std::iota(channels_to_load.begin(), channels_to_load.end(), 0);
  } else {
    // Only load visible channels that exist
    for (size_t ch_idx : visible_channels_) {
      if (ch_idx < level_info.pages.size()) {
        channels_to_load.push_back(ch_idx);
      }
    }
  }

  if (channels_to_load.empty()) {
    return MAKE_STATUSOR(Image, absl::StatusCode::kInvalidArgument,
                         "No valid channels to load");
  }

  // Determine data type from the first channel's bits per sample
  RETURN_IF_ERROR(tiff_file.SetDirectory(level_info.pages[channels_to_load[0]]),
                  "Failed to set directory for first channel");

  uint16_t bits_per_sample = 0;
  ASSIGN_OR_RETURN(bits_per_sample, tiff_file.GetBitsPerSample());

  DataType data_type;
  ASSIGN_OR_RETURN(data_type, tiff_file.GetDataType());

  uint32_t bytes_per_sample = (bits_per_sample + 7) / 8;
  auto num_channels = static_cast<uint32_t>(channels_to_load.size());
  uint32_t actual_width = region.size[0];
  uint32_t actual_height = region.size[1];

  // Create spectral image with configured planar format
  Image result(region.size, num_channels, data_type, output_planar_config_);

  // Get the final destination buffer once
  uint8_t* final_buffer = result.GetData();
  size_t bytes_per_pixel = bytes_per_sample * num_channels;
  size_t pixels_per_channel = static_cast<size_t>(actual_width) * actual_height;

  // Load each channel directly into the final buffer
  for (size_t idx = 0; idx < channels_to_load.size(); ++idx) {
    size_t ch = channels_to_load[idx];

    // Set directory for this channel

    RETURN_IF_ERROR(
        tiff_file.SetDirectory(level_info.pages[ch]),
        "Failed to set directory for channel " + std::to_string(ch));

    uint32_t ch_width = 0;
    uint32_t ch_height = 0;
    std::vector<uint8_t> raw_data;

    if (level_info.allow_random_access) {
      // Use existing random access method for tiled images
      raw_data = ReadTiffRegionWithFile(tiff_file, level_info.pages[ch], region,
                                        ch_width, ch_height, bytes_per_sample);
    } else {
      // For non-tiled images, read complete image and crop
      // Get full image dimensions
      TiffImageDimensions full_dims;
      ASSIGN_OR_RETURN(
          full_dims, tiff_file.GetImageDimensions(),
          "Failed to get image dimensions for channel " + std::to_string(ch));

      // Read complete image
      uint32_t full_width = 0;
      uint32_t full_height = 0;
      RegionSpec full_region = {
          .top_left = {0, 0}, .size = {full_dims[0], full_dims[1]}, .level = 0};
      auto full_data =
          ReadTiffRegionWithFile(tiff_file, level_info.pages[ch], full_region,
                                 full_width, full_height, bytes_per_sample);

      if (full_data.empty()) {
        return MAKE_STATUSOR(
            Image, absl::StatusCode::kInternal,
            "Failed to read complete image for channel " + std::to_string(ch));
      }

      // Crop to requested region
      ch_width = region.size[0];
      ch_height = region.size[1];
      raw_data.resize(ch_width * ch_height * bytes_per_sample);

      // Copy cropped region from full image
      for (uint32_t y = 0; y < ch_height; ++y) {
        for (uint32_t x = 0; x < ch_width; ++x) {
          uint32_t src_idx = ((region.top_left[1] + y) * full_width +
                              (region.top_left[0] + x)) *
                             bytes_per_sample;
          uint32_t dst_idx = (y * ch_width + x) * bytes_per_sample;

          for (uint32_t b = 0; b < bytes_per_sample; ++b) {
            raw_data[dst_idx + b] = full_data[src_idx + b];
          }
        }
      }
    }

    // Copy this channel's data into the final buffer based on planar
    // configuration
    const uint8_t* src_bytes = raw_data.data();

    if (output_planar_config_ == PlanarConfig::kContiguous) {
      // Interleaved format:
      // pixel0_ch0, pixel0_ch1, ..., pixel1_ch0, pixel1_ch1, ...
      for (size_t pixel = 0; pixel < pixels_per_channel; ++pixel) {
        uint8_t* dst_pixel =
            final_buffer + (pixel * bytes_per_pixel) + (idx * bytes_per_sample);
        const uint8_t* src_pixel = src_bytes + (pixel * bytes_per_sample);
        std::memcpy(dst_pixel, src_pixel, bytes_per_sample);
      }
    } else {
      // Channel-separated format: ch0_all_pixels, ch1_all_pixels, ...
      size_t channel_size = pixels_per_channel * bytes_per_sample;
      uint8_t* dst_channel = final_buffer + (idx * channel_size);
      std::memcpy(dst_channel, src_bytes, channel_size);
    }
  }

  return result;
}

absl::StatusOr<RGBImage> QpTiffReader::ReadAssociatedImage(
    std::string_view name) const {
  if (!associated_images_.contains(std::string(name))) {
    return MAKE_STATUSOR(
        RGBImage, absl::StatusCode::kNotFound,
        "Associated image '" + std::string(name) + "' not found");
  }

  const QpTiffAssociatedInfo& info = associated_images_.at(std::string(name));

  // Create TiffFile wrapper once
  auto tiff_file_result = TiffFile::Create(handle_pool_.get());
  if (!tiff_file_result.ok()) {
    return MAKE_STATUSOR(RGBImage, absl::StatusCode::kInternal,
                         "Failed to create TiffFile wrapper");
  }
  auto tiff_file = std::move(tiff_file_result.value());

  return ReadAssociatedImageFromPageWithFile(tiff_file, info.page, info.size[0],
                                             info.size[1], std::string(name));
}

// TODO(jonasteuwen): This function could fail,
// make it absl::StatusOr<ImageDimensions>
ImageDimensions QpTiffReader::GetTileSize() const {
  // Try to get tile size from level 0
  if (pyramid_.empty() || pyramid_[0].pages.empty()) {
    return ImageDimensions{512, 512};  // Default for QPTIFF
  }

  // Get tile dimensions from the first page of level 0
  auto tiff_file_result = TiffFile::Create(handle_pool_.get());
  if (!tiff_file_result.ok()) {
    return ImageDimensions{512, 512};  // Default fallback
  }
  auto tiff_file = std::move(tiff_file_result.value());

  auto status = tiff_file.SetDirectory(pyramid_[0].pages[0]);

  if (!status.ok()) {
    return ImageDimensions{512, 512};  // Default fallback
  }

  if (tiff_file.IsTiled()) {
    auto tile_dims_result = tiff_file.GetTileDimensions();
    if (tile_dims_result.ok()) {
      const auto& dims = tile_dims_result.value();
      return ImageDimensions{dims[0], dims[1]};
    }
  }

  return ImageDimensions{512, 512};  // Default for QPTIFF
}

Metadata QpTiffReader::GetMetadata() const {
  Metadata metadata;
  metadata["format"] = "QPTIFF";
  metadata["mpp_x"] = metadata_.mpp_x;
  metadata["mpp_y"] = metadata_.mpp_y;
  metadata["magnification"] = metadata_.magnification;
  metadata["objective"] = metadata_.objective_name;
  metadata["levels"] = pyramid_.size();
  metadata["channels"] = channels_.size();
  metadata["associated_images"] = associated_images_.size();
  return metadata;
}

void QpTiffReader::PopulateSlideProperties() {
  properties_.mpp[0] = metadata_.mpp_x;
  properties_.mpp[1] = metadata_.mpp_y;
  properties_.objective_magnification = metadata_.magnification;
  properties_.objective_name = metadata_.objective_name;
  properties_.scanner_model = "PerkinElmer/QPTIFF";
  // scan_date is optional and not available in metadata
}

// Utility methods and implementation

std::string QpTiffReader::GetText(const void* node_ptr, const char* tag) {
  const auto* const node = static_cast<const pugi::xml_node*>(node_ptr);
  auto child = node->child(tag);
  return !child.empty() ? child.text().as_string() : std::string{};
}

absl::Status QpTiffReader::ProcessMetadata() {
  ScopedTimer t("QpTiffReader::ProcessMetadata");
  // Create TiffFile wrapper using the handle pool
  auto tiff_file_result = TiffFile::Create(handle_pool_.get());
  if (!tiff_file_result.ok()) {
    return MAKE_STATUS(absl::StatusCode::kInternal,
                       "Failed to create TiffFile wrapper");
  }

  auto tiff_file = std::move(tiff_file_result.value());

  // Get total number of directories upfront
  uint16_t total_pages;
  ASSIGN_OR_RETURN(total_pages, tiff_file.GetDirectoryCount());

  if (total_pages < 4) {
    return MAKE_STATUS(
        absl::StatusCode::kInvalidArgument,
        "QPTIFF file has too few pages: " + std::to_string(total_pages));
  }

  std::vector<QpTiffChannelInfo> full_channels;
  uint16_t thumbnail_page = 0;

  // Helper lambda to check if a page is a thumbnail
  auto isThumbnailPage = [&](uint16_t p) -> bool {
    if (tiff_file.SetDirectory(p) != absl::OkStatus()) {
      return true;  // Stop iteration on error
    }

    auto desc_result = tiff_file.GetImageDescription();
    if (desc_result.empty()) {
      return false;  // Not a thumbnail, continue
    }

    pugi::xml_document doc;
    if (!doc.load_string(desc_result.c_str())) {
      return true;  // Stop iteration on parse error
    }

    auto root = doc.child("PerkinElmer-QPI-ImageDescription");
    if (root.empty()) {
      return true;  // Stop iteration on structure error
    }

    std::string image_type = GetText(&root, "ImageType");
    return image_type == "Thumbnail";
  };

  // Phase 1: Read full resolution channels until we hit thumbnail
  for (auto page :
       std::views::iota(0u, total_pages) | std::views::take_while([&](auto p) {
         return !isThumbnailPage(p);
       })) {
    ScopedTimer t_page("QpTiffReader::ProcessMetadata - Page setup");
    RETURN_IF_ERROR(tiff_file.SetDirectory(page),
                    "Failed to set directory " + std::to_string(page));

    // Get basic directory info
    auto image_dims_result = tiff_file.GetImageDimensions();
    if (!image_dims_result.ok()) {
      return MAKE_STATUS(
          absl::StatusCode::kInternal,
          "Failed to get image dimensions for page " + std::to_string(page));
    }
    const auto& image_dims = image_dims_result.value();

    // Get XML metadata
    auto desc_result = tiff_file.GetImageDescription();
    if (desc_result.empty()) {
      continue;  // Skip pages without XML
    }

    // Parse XML
    ScopedTimer t_xml("QpTiffReader::ProcessMetadata - XML parse");
    pugi::xml_document doc;
    if (!doc.load_string(desc_result.c_str())) {
      return MAKE_STATUS(
          absl::StatusCode::kInvalidArgument,
          "Failed to parse XML metadata on page " + std::to_string(page));
    }

    auto root = doc.child("PerkinElmer-QPI-ImageDescription");
    if (root.empty()) {
      return MAKE_STATUS(
          absl::StatusCode::kInvalidArgument,
          "Invalid XML structure on page " + std::to_string(page));
    }

    std::string image_type = GetText(&root, "ImageType");

    // This should be a full resolution channel
    if (image_type == "FullResolution" || image_type.empty()) {
      QpTiffChannelInfo channel;
      ASSIGN_OR_RETURN(
          channel,
          ExtractChannelFromXml(root, static_cast<int>(full_channels.size())));

      channel.page = page;
      channel.width = image_dims[0];
      channel.height = image_dims[1];
      channel.tiled = tiff_file.IsTiled();
      channel.allow_random_access = channel.tiled;

      full_channels.push_back(std::move(channel));

      // Extract metadata from page 0
      if (page == 0) {
        RETURN_IF_ERROR(ExtractResolutionMetadata(tiff_file, metadata_, &root),
                        "Failed to extract resolution metadata from page 0");
      }
    }

    thumbnail_page = page + 1;  // Track the next page after processing
  }

  // Find and process the thumbnail page
  for (uint16_t current_page = thumbnail_page; current_page < total_pages;
       ++current_page) {
    if (isThumbnailPage(current_page)) {
      RETURN_IF_ERROR(tiff_file.SetDirectory(current_page),
                      "Failed to set directory for thumbnail");

      auto image_dims_result = tiff_file.GetImageDimensions();
      if (image_dims_result.ok()) {
        const auto& image_dims = image_dims_result.value();
        associated_images_["Thumbnail"] = QpTiffAssociatedInfo{
            .page = current_page, .size = {image_dims[0], image_dims[1]}};
      }
      thumbnail_page = current_page;
      break;
    }
  }

  if (full_channels.empty()) {
    return MAKE_STATUS(absl::StatusCode::kInvalidArgument,
                       "No full resolution channels found");
  }

  size_t num_channels = full_channels.size();
  channels_ = std::move(full_channels);

  // Phase 2: Build pyramid structure by parsing all remaining pages
  // Level 0 (full resolution)
  QpTiffLevelInfo level0;
  level0.Reserve(num_channels);
  for (size_t i = 0; i < num_channels; ++i) {
    level0.pages.push_back(channels_[i].page);
  }
  level0.size = Size<uint32_t, 2>{channels_[0].width, channels_[0].height};
  level0.tiled =
      std::ranges::all_of(channels_, [](const auto& ch) { return ch.tiled; });
  level0.allow_random_access = level0.tiled;
  pyramid_.push_back(std::move(level0));

  // Process remaining pages after thumbnail:
  // reduced levels followed by associated images
  uint16_t current_page = thumbnail_page + 1;
  std::vector<uint16_t> current_level_pages;

  while (current_page < total_pages) {
    ScopedTimer t_page2("QpTiffReader::ProcessMetadata - Reduced/Assoc page");
    RETURN_IF_ERROR(tiff_file.SetDirectory(current_page),
                    "Failed to set directory " + std::to_string(current_page));

    // Get XML metadata to determine image type
    auto desc_result = tiff_file.GetImageDescription();
    std::string image_type;

    if (!desc_result.empty()) {
      ScopedTimer t_xml2("QpTiffReader::ProcessMetadata - XML parse (rest)");
      pugi::xml_document doc;
      if (doc.load_string(desc_result.c_str())) {
        auto root = doc.child("PerkinElmer-QPI-ImageDescription");
        if (!root.empty()) {
          image_type = GetText(&root, "ImageType");
        }
      }
    }

    if (image_type == "ReducedResolution" || image_type.empty()) {
      // This is part of a reduced level
      current_level_pages.push_back(current_page);

      // If we have collected enough pages for one level
      // (num_channels), create the level
      if (current_level_pages.size() == num_channels) {
        QpTiffLevelInfo reduced_level;
        reduced_level.Reserve(num_channels);
        reduced_level.pages = current_level_pages;

        // Get dimensions from first page of this level
        RETURN_IF_ERROR(tiff_file.SetDirectory(current_level_pages[0]),
                        "Failed to set directory for reduced level");
        ASSIGN_OR_RETURN(reduced_level.size, tiff_file.GetImageDimensions());

        reduced_level.tiled = tiff_file.IsTiled();
        reduced_level.allow_random_access = reduced_level.tiled;

        pyramid_.push_back(std::move(reduced_level));
        current_level_pages.clear();
      }
    } else {
      // This is an associated image - everything after ReducedResolution
      TiffImageDimensions dims;
      ASSIGN_OR_RETURN(dims, tiff_file.GetImageDimensions(),
                       "Failed to get dimensions for associated image");

      // Use ImageType as the name, fallback to page-based naming
      std::string assoc_name =
          image_type.empty() ? "Associated_" + std::to_string(current_page)
                             : image_type;

      associated_images_[assoc_name] =
          QpTiffAssociatedInfo{.page = current_page, .size = dims};
    }

    ++current_page;
  }

  // Handle any remaining pages in current_level_pages (partial level)
  if (!current_level_pages.empty()) {
    // This shouldn't happen in well-formed QPTIFF files, but handle gracefully
    LOG(WARNING) << "Found incomplete reduced level with "
                 << current_level_pages.size() << " pages (expected "
                 << num_channels << ")";
  }

  return absl::OkStatus();
}

}  // namespace fastslide
