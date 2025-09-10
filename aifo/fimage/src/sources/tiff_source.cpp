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
/**
 * @file tiff_source.cpp
 * @brief Implementation of TIFF image source for the fim library.
 * @author Jonas Teuwen
 * @date 2025
 *
 * This file contains the implementation of the TiffSource class which provides
 * TIFF image reading capabilities using libtiff. It handles both tiled and
 * strip-based TIFF files with efficient tile-based access and lazy loading
 * of image metadata.
 */
#include "fim/sources/tiff_source.h"

#include <tiffio.h>

#include <algorithm>
#include <cstring>
#include <filesystem>
#include <stdexcept>
#include <string>
#include <utility>
#include <vector>

namespace fim {

namespace fs = std::filesystem;

TiffSource::TiffSource(const fs::path& filename)
    : filename_(filename), tiff_handle_(nullptr), dimensions_loaded_(false) {
  OpenTiff(filename);
}

TiffSource::~TiffSource() {
  CloseTiff();
}

TiffSource::TiffSource(TiffSource&& other) noexcept
    : filename_(std::move(other.filename_)),
      tiff_handle_(other.tiff_handle_),
      dimensions_(other.dimensions_),
      ideal_tile_size_(other.ideal_tile_size_),
      dimensions_loaded_(other.dimensions_loaded_) {
  other.tiff_handle_ = nullptr;
  other.dimensions_loaded_ = false;
}

TiffSource& TiffSource::operator=(TiffSource&& other) noexcept {
  if (this != &other) {
    CloseTiff();
    filename_ = std::move(other.filename_);
    tiff_handle_ = other.tiff_handle_;
    dimensions_ = other.dimensions_;
    ideal_tile_size_ = other.ideal_tile_size_;
    dimensions_loaded_ = other.dimensions_loaded_;
    other.tiff_handle_ = nullptr;
    other.dimensions_loaded_ = false;
  }
  return *this;
}

void TiffSource::OpenTiff(const fs::path& filename) {
  tiff_handle_ = TIFFOpen(filename.c_str(), "r");
  if (!tiff_handle_) {
    throw std::runtime_error("Failed to open TIFF file: " + filename.string());
  }
}

void TiffSource::CloseTiff() {
  if (tiff_handle_) {
    TIFFClose(tiff_handle_);
    tiff_handle_ = nullptr;
  }
}

ImageDimensions TiffSource::GetDimensions() const {
  if (!dimensions_loaded_) {
    LoadDimensions();
  }
  return dimensions_;
}

TileSize TiffSource::GetIdealTileSize() const {
  if (!dimensions_loaded_) {
    LoadDimensions();
  }
  return ideal_tile_size_;
}

void TiffSource::LoadDimensions() const {
  if (!tiff_handle_) {
    throw std::runtime_error("TIFF file not opened");
  }

  uint32_t width, height;
  uint16_t samples_per_pixel;

  TIFFGetField(tiff_handle_, TIFFTAG_IMAGEWIDTH, &width);
  TIFFGetField(tiff_handle_, TIFFTAG_IMAGELENGTH, &height);
  TIFFGetField(tiff_handle_, TIFFTAG_SAMPLESPERPIXEL, &samples_per_pixel);

  dimensions_ =
      ImageDimensions(static_cast<int>(width), static_cast<int>(height),
                      static_cast<int>(samples_per_pixel));

  // Check if TIFF is tiled and get tile dimensions
  uint32_t tile_width, tile_height;
  if (TIFFIsTiled(tiff_handle_)) {
    TIFFGetField(tiff_handle_, TIFFTAG_TILEWIDTH, &tile_width);
    TIFFGetField(tiff_handle_, TIFFTAG_TILELENGTH, &tile_height);
    ideal_tile_size_ =
        TileSize(static_cast<int>(tile_width), static_cast<int>(tile_height));
  } else {
    // For non-tiled TIFF, use strip size as ideal tile height
    uint32_t rows_per_strip;
    TIFFGetField(tiff_handle_, TIFFTAG_ROWSPERSTRIP, &rows_per_strip);
    ideal_tile_size_ =
        TileSize(static_cast<int>(width), static_cast<int>(rows_per_strip));
  }

  dimensions_loaded_ = true;
}

Tile TiffSource::GetTile(int x, int y, int width, int height) const {
  if (!tiff_handle_) {
    throw std::runtime_error("TIFF file not opened");
  }

  if (!dimensions_loaded_) {
    LoadDimensions();
  }

  // Clamp tile bounds to image dimensions
  int actual_width = std::min(width, dimensions_.width - x);
  int actual_height = std::min(height, dimensions_.height - y);

  if (actual_width <= 0 || actual_height <= 0) {
    return Tile(x, y, 0, 0, dimensions_.channels);
  }

  Tile tile(x, y, actual_width, actual_height, dimensions_.channels);

  // Read tile data
  if (TIFFIsTiled(tiff_handle_)) {
    ReadTiledData(tile);
  } else {
    ReadStripData(tile);
  }

  return tile;
}

void TiffSource::ReadTiledData(Tile& tile) const {
  uint32_t tile_width, tile_height;
  TIFFGetField(tiff_handle_, TIFFTAG_TILEWIDTH, &tile_width);
  TIFFGetField(tiff_handle_, TIFFTAG_TILELENGTH, &tile_height);

  size_t tile_size = TIFFTileSize(tiff_handle_);
  std::vector<uint8_t> tiff_tile_data(tile_size);

  // Calculate tile-aligned bounds
  int tile_x_start = (tile.x / tile_width) * tile_width;
  int tile_y_start = (tile.y / tile_height) * tile_height;
  int tile_x_end =
      ((tile.x + tile.width + tile_width - 1) / tile_width) * tile_width;
  int tile_y_end =
      ((tile.y + tile.height + tile_height - 1) / tile_height) * tile_height;

  // Read overlapping tiles using tile-aligned coordinates
  for (int ty = tile_y_start; ty < tile_y_end; ty += tile_height) {
    for (int tx = tile_x_start; tx < tile_x_end; tx += tile_width) {
      if (TIFFReadTile(tiff_handle_, tiff_tile_data.data(), tx, ty, 0, 0) < 0) {
        throw std::runtime_error("Failed to read TIFF tile");
      }

      // Copy relevant portion of tile data
      CopyTileData(tiff_tile_data, tile, tx, ty, tile_width, tile_height);
    }
  }
}

void TiffSource::ReadStripData(Tile& tile) const {
  uint32_t rows_per_strip;
  TIFFGetField(tiff_handle_, TIFFTAG_ROWSPERSTRIP, &rows_per_strip);

  size_t strip_size = TIFFStripSize(tiff_handle_);
  std::vector<uint8_t> strip_data(strip_size);

  // Read overlapping strips
  for (int row = tile.y; row < tile.y + tile.height; row += rows_per_strip) {
    int strip_num = row / rows_per_strip;
    if (TIFFReadEncodedStrip(tiff_handle_, strip_num, strip_data.data(),
                             strip_size) < 0) {
      throw std::runtime_error("Failed to read TIFF strip");
    }

    // Copy relevant portion of strip data
    CopyStripData(strip_data, tile, row, rows_per_strip);
  }
}

void TiffSource::CopyTileData(const std::vector<uint8_t>& src_data, Tile& tile,
                              int tile_x, int tile_y, int tile_width,
                              int tile_height) const {
  // Calculate intersection of source tile and requested tile
  int src_x_start = std::max(0, tile.x - tile_x);
  int src_y_start = std::max(0, tile.y - tile_y);
  int src_x_end = std::min(tile_width, tile.x + tile.width - tile_x);
  int src_y_end = std::min(tile_height, tile.y + tile.height - tile_y);

  int dst_x_start = std::max(0, tile_x - tile.x);
  int dst_y_start = std::max(0, tile_y - tile.y);

  // Copy data row by row
  for (int y = src_y_start; y < src_y_end; ++y) {
    int src_row_offset = y * tile_width * tile.channels;
    int dst_row_offset =
        (dst_y_start + y - src_y_start) * tile.width * tile.channels;

    int copy_width = (src_x_end - src_x_start) * tile.channels;
    std::memcpy(tile.data.data() + dst_row_offset + dst_x_start * tile.channels,
                src_data.data() + src_row_offset + src_x_start * tile.channels,
                copy_width);
  }
}

void TiffSource::CopyStripData(const std::vector<uint8_t>& src_data, Tile& tile,
                               int strip_row, int rows_per_strip) const {
  // Calculate intersection of source strip and requested tile
  int src_y_start = std::max(0, tile.y - strip_row);
  int src_y_end = std::min(rows_per_strip, tile.y + tile.height - strip_row);

  int dst_y_start = std::max(0, strip_row - tile.y);

  // Copy data row by row
  for (int y = src_y_start; y < src_y_end; ++y) {
    int src_row_offset = y * dimensions_.width * tile.channels;
    int dst_row_offset =
        (dst_y_start + y - src_y_start) * tile.width * tile.channels;

    std::memcpy(tile.data.data() + dst_row_offset,
                src_data.data() + src_row_offset + tile.x * tile.channels,
                tile.width * tile.channels);
  }
}

}  // namespace fim
