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

#include "fastslide/readers/tiff_based_reader.h"

#include <tiffio.h>
#include <algorithm>
#include <cmath>
#include <cstdint>
#include <cstdio>
#include <cstdlib>
#include <cstring>
#include <memory>
#include <ranges>
#include <string>
#include <utility>
#include <vector>

#include "absl/status/status.h"
#include "absl/status/statusor.h"

#include "aifocore/concepts/numeric.h"
#include "aifocore/status/status_macros.h"
#include "aifocore/utilities/fmt.h"
#include "fastslide/utilities/scoped_timer.h"
#include "fastslide/utilities/tiff/tiff_file.h"
#include "fastslide/utilities/tile_cache_manager.h"

namespace fs = std::filesystem;

namespace fastslide {

using aifocore::fmt::format;

// Type alias for tile coordinates
using TileCoordinate = aifocore::Size<uint32_t, 2>;

/**
 * @brief Iterator for 2D tile coordinates within a region
 * 
 * This iterator yields tile coordinates (x, y) that intersect with a given region,
 * following C++20 iterator concepts for use with ranges and algorithms.
 */
class TileCoordinateIterator {
 public:
  using iterator_category = std::forward_iterator_tag;
  using difference_type = std::ptrdiff_t;
  using value_type = TileCoordinate;
  using pointer = const value_type*;
  using reference = const value_type&;

  TileCoordinateIterator(uint32_t start_x, uint32_t start_y, uint32_t end_x,
                         uint32_t end_y, uint32_t tile_width,
                         uint32_t tile_height)
      : current_x_(start_x),
        current_y_(start_y),
        start_x_(start_x),
        end_x_(end_x),
        end_y_(end_y),
        tile_width_(tile_width),
        tile_height_(tile_height),
        current_coord_{start_x, start_y},
        is_end_(start_y >= end_y) {}

  // End iterator constructor
  TileCoordinateIterator()
      : current_x_(0),
        current_y_(0),
        start_x_(0),
        end_x_(0),
        end_y_(0),
        tile_width_(0),
        tile_height_(0),
        current_coord_{0, 0},
        is_end_(true) {}

  reference operator*() const { return current_coord_; }

  pointer operator->() const { return &current_coord_; }

  TileCoordinateIterator& operator++() {
    if (is_end_)
      return *this;

    current_x_ += tile_width_;
    if (current_x_ >= end_x_) {
      current_x_ = start_x_;
      current_y_ += tile_height_;
      if (current_y_ >= end_y_) {
        is_end_ = true;
        return *this;
      }
    }
    current_coord_ = {current_x_, current_y_};
    return *this;
  }

  TileCoordinateIterator operator++(int) {
    TileCoordinateIterator tmp = *this;
    ++(*this);
    return tmp;
  }

  friend bool operator==(const TileCoordinateIterator& a,
                         const TileCoordinateIterator& b) {
    if (a.is_end_ && b.is_end_)
      return true;
    if (a.is_end_ || b.is_end_)
      return false;
    return a.current_x_ == b.current_x_ && a.current_y_ == b.current_y_;
  }

  friend bool operator!=(const TileCoordinateIterator& a,
                         const TileCoordinateIterator& b) {
    return !(a == b);
  }

 private:
  uint32_t current_x_, current_y_;
  uint32_t start_x_, end_x_, end_y_;
  uint32_t tile_width_, tile_height_;
  value_type current_coord_;
  bool is_end_;
};

/**
 * @brief Range view for tile coordinates within a region
 */
class TileCoordinateRange {
 public:
  TileCoordinateRange(const RegionSpec& region,
                      const aifocore::Size<uint32_t, 2>& tile_dims)
      : start_x_((region.top_left[0] / tile_dims[0]) * tile_dims[0]),
        start_y_((region.top_left[1] / tile_dims[1]) * tile_dims[1]),
        end_x_(region.top_left[0] + region.size[0]),
        end_y_(region.top_left[1] + region.size[1]),
        tile_width_(tile_dims[0]),
        tile_height_(tile_dims[1]) {}

  TileCoordinateIterator begin() const {
    return TileCoordinateIterator(start_x_, start_y_, end_x_, end_y_,
                                  tile_width_, tile_height_);
  }

  TileCoordinateIterator end() const { return TileCoordinateIterator(); }

 private:
  uint32_t start_x_, start_y_, end_x_, end_y_;
  uint32_t tile_width_, tile_height_;
};

TiffBasedReader::TiffBasedReader(fs::path filename)
    : filename_(std::move(filename)),
      tile_cache_manager_(std::make_unique<TileCacheManager>()) {}

void TiffBasedReader::SetCache(std::shared_ptr<TileCache> cache) {
  SlideReader::SetCache(cache);
  tile_cache_manager_->SetCache(cache);
}

absl::Status TiffBasedReader::ValidateTiffFile(const fs::path& filename) {
  return TiffFile::ValidateFile(filename);
}

absl::Status TiffBasedReader::InitializeHandlePool() {
  auto handle_pool_result = TIFFHandlePool::Create(filename_);
  if (!handle_pool_result.ok()) {
    return handle_pool_result.status();
  }
  handle_pool_ = std::move(handle_pool_result.value());
  return absl::OkStatus();
}

std::vector<uint8_t> TiffBasedReader::ReadTiffRegion(
    uint16_t page, const RegionSpec& region, uint32_t& actual_width,
    uint32_t& actual_height, uint32_t bytes_per_pixel) const {

  // Create TiffFile wrapper using the handle pool
  auto tiff_file_result = TiffFile::Create(handle_pool_.get());
  if (!tiff_file_result.ok()) {
    return {};
  }
  auto tiff_file = std::move(tiff_file_result.value());

  return ReadTiffRegionWithFile(tiff_file, page, region, actual_width,
                                actual_height, bytes_per_pixel);
}

std::vector<uint8_t> TiffBasedReader::ReadTiffRegionWithFile(
    TiffFile& tiff_file, uint16_t page, const RegionSpec& region,
    uint32_t& actual_width, uint32_t& actual_height,
    uint32_t bytes_per_pixel) const {
  ScopedTimer t_region("TiffBasedReader::ReadTiffRegionWithFile");

  // Set directory
  auto status = tiff_file.SetDirectory(page);

  if (!status.ok()) {
    return {};
  }

  // Get image dimensions
  auto image_dims_result = tiff_file.GetImageDimensions();
  if (!image_dims_result.ok()) {
    return {};
  }
  const auto& image_dims = image_dims_result.value();

  // Clamp region to image bounds using elementwise operations
  auto remaining_space = image_dims - region.top_left;
  auto actual_size = Min(region.size, remaining_space);
  actual_width = actual_size[0];
  actual_height = actual_size[1];

  if (actual_width == 0 || actual_height == 0) {
    return {};
  }

  // Allocate vector for region
  std::vector<uint8_t> buffer(actual_width * actual_height * bytes_per_pixel);

  // Check if image is tiled
  bool is_tiled = tiff_file.IsTiled();

  absl::Status read_status;
  if (is_tiled) {
    read_status = ReadTiledRegion(tiff_file, buffer.data(), page, region,
                                  bytes_per_pixel);
  } else {
    read_status = ReadStripRegion(tiff_file, buffer.data(), page, region,
                                  bytes_per_pixel);
  }

  if (!read_status.ok()) {
    return {};
  }

  return buffer;
}

absl::Status TiffBasedReader::ReadTiledRegion(TiffFile& tiff_file,
                                              uint8_t* buffer, uint16_t page,
                                              const RegionSpec& region,
                                              uint32_t bytes_per_pixel) const {
  ScopedTimer t_tiled("TiffBasedReader::ReadTiledRegion");
  // Get tile dimensions
  TiffTileDimensions tile_dims;
  ASSIGN_OR_RETURN(tile_dims, tiff_file.GetTileDimensions(),
                   "Failed to get tile dimensions");

  // Get tile size
  size_t tile_size = 0;
  ASSIGN_OR_RETURN(tile_size, tiff_file.GetTileSize(),
                   "Failed to get tile size");

  // Check planar configuration
  TiffPlanarConfig planar_config;
  ASSIGN_OR_RETURN(planar_config, tiff_file.GetPlanarConfig(),
                   "Failed to get planar configuration");

  // Get number of samples per pixel
  uint16_t samples_per_pixel = 0;
  ASSIGN_OR_RETURN(samples_per_pixel, tiff_file.GetSamplesPerPixel(),
                   "Failed to get samples per pixel");

  uint32_t bytes_per_sample = bytes_per_pixel / samples_per_pixel;
  std::vector<uint8_t> tile_buffer(tile_size);
  std::vector<uint8_t> planar_tile_buffer;

  if (planar_config == TiffPlanarConfig::Separate) {
    planar_tile_buffer.resize(tile_size * samples_per_pixel);
  }

  // Iterate through tiles that intersect with our region
  for (const auto& tile_coords : TileCoordinateRange(region, tile_dims)) {
    // Calculate tile indices for cache key using Size elementwise division
    auto cache_tile_coords = tile_coords / tile_dims;

    // Create tile loader function for cache miss handling
    auto tile_loader = [&]() -> absl::StatusOr<std::shared_ptr<CachedTile>> {
      ScopedTimer t_load("TileLoader miss: ReadTile + interleave");
      std::vector<uint8_t> final_tile_data;

      if (planar_config == TiffPlanarConfig::Separate) {
        // Read each sample separately and interleave
        final_tile_data.resize(tile_size * samples_per_pixel);

        for (uint16_t sample = 0; sample < samples_per_pixel; ++sample) {
          RETURN_IF_ERROR(
              tiff_file.ReadTile(tile_buffer.data(), tile_coords, sample),
              "Failed to read tile for sample " + std::to_string(sample));

          // Interleave this sample's data
          size_t pixels_in_tile = tile_size / bytes_per_sample;
          for (size_t pixel = 0; pixel < pixels_in_tile; ++pixel) {
            size_t src_offset = pixel * bytes_per_sample;
            size_t dst_offset =
                pixel * bytes_per_pixel + sample * bytes_per_sample;
            std::memcpy(&final_tile_data[dst_offset], &tile_buffer[src_offset],
                        bytes_per_sample);
          }
        }
      } else {
        // Read interleaved tile directly
        RETURN_IF_ERROR(tiff_file.ReadTile(tile_buffer.data(), tile_coords),
                        "Failed to read tile");
        final_tile_data.assign(tile_buffer.begin(),
                               tile_buffer.begin() + tile_size);
      }

      return std::make_shared<CachedTile>(std::move(final_tile_data), tile_dims,
                                          bytes_per_pixel);
    };

    // Get tile using cache manager
    auto tile_result = tile_cache_manager_->GetTile(
        filename_.string(), page, cache_tile_coords, tile_loader);

    if (!tile_result.ok()) {
      continue;  // Skip failed tiles
    }

    // Copy cached data to tile buffer
    auto cached_tile = tile_result.value();
    if (cached_tile && cached_tile->data.size() <= tile_buffer.size()) {
      std::copy(cached_tile->data.begin(), cached_tile->data.end(),
                tile_buffer.begin());
    } else {
      continue;  // Skip if tile data is invalid
    }

    {
      ScopedTimer t_copy("CopyTileToBuffer");
      // Copy relevant portion of tile to output buffer
      CopyTileToBuffer(tile_buffer.data(), buffer, tile_dims[0], tile_dims[1],
                       tile_coords[0], tile_coords[1], region.top_left[0],
                       region.top_left[1], region.size[0], region.size[1],
                       bytes_per_pixel);
    }
  }

  return absl::OkStatus();
}

absl::Status TiffBasedReader::ReadStripRegion(TiffFile& tiff_file,
                                              uint8_t* buffer, uint16_t page,
                                              const RegionSpec& region,
                                              uint32_t bytes_per_pixel) const {
  ScopedTimer t_strip("TiffBasedReader::ReadStripRegion");
  // Get image dimensions to determine width for scanline caching
  TiffImageDimensions image_dims;
  ASSIGN_OR_RETURN(image_dims, tiff_file.GetImageDimensions(),
                   "Failed to get image dimensions");
  uint32_t image_width = image_dims[0];

  // Get scanline size
  size_t scanline_size;
  ASSIGN_OR_RETURN(scanline_size, tiff_file.GetScanlineSize(),
                   "Failed to get scanline size");

  // Check planar configuration
  TiffPlanarConfig planar_config;
  ASSIGN_OR_RETURN(planar_config, tiff_file.GetPlanarConfig(),
                   "Failed to get planar configuration");

  // Get number of samples per pixel
  uint16_t samples_per_pixel = 0;
  ASSIGN_OR_RETURN(samples_per_pixel, tiff_file.GetSamplesPerPixel(),
                   "Failed to get samples per pixel");

  uint32_t bytes_per_sample = bytes_per_pixel / samples_per_pixel;
  std::vector<uint8_t> scanline_buffer(scanline_size);

  for (uint32_t row = 0; row < region.size[1]; ++row) {
    uint32_t image_row = region.top_left[1] + row;

    // Create scanline loader function for cache miss handling
    auto scanline_loader =
        [&]() -> absl::StatusOr<std::shared_ptr<CachedTile>> {
      // Timer disabled, it takes about ~20us, so not a bottleneck.
      // ScopedTimer t_load("ScanlineLoader miss: ReadScanline + interleave");
      std::vector<uint8_t> final_scanline_data;

      if (planar_config == TiffPlanarConfig::Separate) {
        // Read each sample separately and interleave
        final_scanline_data.resize(scanline_size * samples_per_pixel);

        for (uint16_t sample = 0; sample < samples_per_pixel; ++sample) {
          RETURN_IF_ERROR(
              tiff_file.ReadScanline(scanline_buffer.data(), image_row, sample),
              "Failed to read scanline for sample " + std::to_string(sample));

          // Interleave this sample's data
          size_t pixels_in_scanline = scanline_size / bytes_per_sample;
          for (size_t pixel = 0; pixel < pixels_in_scanline; ++pixel) {
            size_t src_offset = pixel * bytes_per_sample;
            size_t dst_offset =
                pixel * bytes_per_pixel + sample * bytes_per_sample;
            std::memcpy(&final_scanline_data[dst_offset],
                        &scanline_buffer[src_offset], bytes_per_sample);
          }
        }
      } else {
        // Read interleaved scanline directly
        RETURN_IF_ERROR(
            tiff_file.ReadScanline(scanline_buffer.data(), image_row, 0),
            "Failed to read scanline");
        final_scanline_data.assign(scanline_buffer.begin(),
                                   scanline_buffer.begin() + scanline_size);
      }

      return std::make_shared<CachedTile>(
          std::move(final_scanline_data),
          aifocore::Size<uint32_t, 2>{image_width,
                                      1},  // height = 1 for scanline
          bytes_per_pixel);
    };

    // Get scanline using cache manager (treat as a "tile")
    auto scanline_result = tile_cache_manager_->GetTile(
        filename_.string(), static_cast<uint16_t>(page),
        TileCoordinate{0, image_row}, scanline_loader);

    if (!scanline_result.ok()) {
      return absl::InternalError("Failed to read scanline");
    }

    // Copy cached data to scanline buffer
    auto cached_scanline = scanline_result.value();
    if (cached_scanline &&
        cached_scanline->data.size() <= scanline_buffer.size()) {
      std::copy(cached_scanline->data.begin(), cached_scanline->data.end(),
                scanline_buffer.begin());
    } else {
      return absl::InternalError("Invalid scanline data");
    }

    // Copy the relevant portion of the scanline
    std::copy(scanline_buffer.data() + region.top_left[0] * bytes_per_pixel,
              scanline_buffer.data() +
                  (region.top_left[0] + region.size[0]) * bytes_per_pixel,
              buffer + row * region.size[0] * bytes_per_pixel);
  }

  return absl::OkStatus();
}

void TiffBasedReader::CopyTileToBuffer(const uint8_t* tile, uint8_t* out,
                                       uint32_t tile_w, uint32_t tile_h,
                                       uint32_t tile_x, uint32_t tile_y,
                                       uint32_t rx, uint32_t ry, uint32_t rw,
                                       uint32_t rh, uint32_t bpp) const {
  const uint32_t x0 = std::max(rx, tile_x);
  const uint32_t y0 = std::max(ry, tile_y);
  const uint32_t x1 = std::min(rx + rw, tile_x + tile_w);
  const uint32_t y1 = std::min(ry + rh, tile_y + tile_h);
  if (x0 >= x1 || y0 >= y1)
    return;

  const uint32_t rows = y1 - y0;
  const uint32_t cols = x1 - x0;

  const size_t copy_bytes = size_t(cols) * bpp;
  const size_t tile_row_stride = size_t(tile_w) * bpp;
  const size_t out_row_stride = size_t(rw) * bpp;

  const size_t tile_xoff = size_t(x0 - tile_x) * bpp;
  const size_t out_xoff = size_t(x0 - rx) * bpp;

  const uint8_t* src = tile + size_t(y0 - tile_y) * tile_row_stride + tile_xoff;
  uint8_t* dst = out + size_t(y0 - ry) * out_row_stride + out_xoff;

  for (uint32_t r = 0; r < rows; ++r) {
    std::memcpy(dst, src, copy_bytes);
    src += tile_row_stride;
    dst += out_row_stride;
  }
}

// void TiffBasedReader::CopyTileToBuffer(
//     const uint8_t* tile_buffer, uint8_t* output_buffer, uint32_t tile_width,
//     uint32_t tile_height, uint32_t tile_x, uint32_t tile_y, uint32_t region_x,
//     uint32_t region_y, uint32_t region_width, uint32_t region_height,
//     uint32_t bytes_per_pixel) const {
//   // Calculate intersection of tile and region
//   uint32_t copy_start_x = std::max(region_x, tile_x);
//   uint32_t copy_end_x = std::min(region_x + region_width, tile_x + tile_width);
//   uint32_t copy_start_y = std::max(region_y, tile_y);
//   uint32_t copy_end_y =
//       std::min(region_y + region_height, tile_y + tile_height);

//   for (uint32_t row = copy_start_y; row < copy_end_y; ++row) {
//     uint32_t tile_row = row - tile_y;
//     uint32_t out_row = row - region_y;

//     // Source offset in tile
//     uint32_t src_offset =
//         (tile_row * tile_width + (copy_start_x - tile_x)) * bytes_per_pixel;
//     // Destination offset in output buffer
//     uint32_t dst_offset =
//         (out_row * region_width + (copy_start_x - region_x)) * bytes_per_pixel;
//     // Number of bytes to copy
//     uint32_t copy_bytes = (copy_end_x - copy_start_x) * bytes_per_pixel;

//     std::copy(tile_buffer + src_offset, tile_buffer + src_offset + copy_bytes,
//               output_buffer + dst_offset);
//   }
// }

absl::StatusOr<RGBImage> TiffBasedReader::ReadAssociatedImageFromPage(
    uint16_t page, uint32_t width, uint32_t height,
    const std::string& name) const {

  // Create TiffFile wrapper using the handle pool
  auto tiff_file_result = TiffFile::Create(handle_pool_.get());
  if (!tiff_file_result.ok()) {
    return MAKE_STATUSOR(RGBImage, absl::StatusCode::kInternal,
                         "Failed to create TiffFile wrapper");
  }
  auto tiff_file = std::move(tiff_file_result.value());

  return ReadAssociatedImageFromPageWithFile(tiff_file, page, width, height,
                                             name);
}

absl::StatusOr<RGBImage> TiffBasedReader::ReadAssociatedImageFromPageWithFile(
    TiffFile& tiff_file, uint16_t page, uint32_t width, uint32_t height,
    const std::string& name) const {

  // Set directory
  RETURN_IF_ERROR(tiff_file.SetDirectory(page),
                  format("Failed to set directory for {}", name));

  // Allocate RGBA buffer using standard C++ memory management
  std::vector<uint32_t> raster(width * height);

  // Read RGBA image, these are usually the associated images in the TIFF file
  RETURN_IF_ERROR(tiff_file.ReadRGBAImage(raster.data(), width, height, true),
                  "Failed to read RGBA data");

  ImageDimensions dims{width, height};
  RGBImage rgb_image(dims, ImageFormat::kRGB, DataType::kUInt8);

  for (uint32_t y = 0; y < height; ++y) {
    for (uint32_t x = 0; x < width; ++x) {
      uint32_t flipped_y = height - 1 - y;
      uint32_t rgba = raster[flipped_y * width + x];

      rgb_image.At<uint8_t>(y, x, 0) = TIFFGetR(rgba);
      rgb_image.At<uint8_t>(y, x, 1) = TIFFGetG(rgba);
      rgb_image.At<uint8_t>(y, x, 2) = TIFFGetB(rgba);
    }
  }

  return rgb_image;
}

int TiffBasedReader::GetBestLevelForDownsampleImpl(
    double downsample, int level_count,
    std::function<double(int)> get_level_downsample) const {
  if (level_count == 0) {
    return 0;
  }

  int best_level = 0;
  double best_diff = std::abs(1.0 - downsample);

  for (int level = 0; level < level_count; ++level) {
    double level_downsample = get_level_downsample(level);
    double diff = std::abs(level_downsample - downsample);
    if (diff < best_diff) {
      best_diff = diff;
      best_level = level;
    }
  }

  return best_level;
}

}  // namespace fastslide
