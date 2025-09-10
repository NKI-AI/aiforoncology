
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
 * @file png_source.cpp
 * @brief Implementation of PNG image source for the fim library.
 * @author Jonas Teuwen
 * @date 2025
 *
 * This file contains the implementation of the PngSource class which provides
 * PNG image reading capabilities using the lodepng library. It handles lazy
 * loading of PNG data and provides tile-based access to the image content.
 */
#include "fim/sources/png_source.h"

#include <algorithm>
#include <cstring>
#include <string>
#include <utility>

#include "lodepng/lodepng.h"

namespace fim {

PngSource::PngSource(const fs::path& filename)
    : filename_(filename), data_loaded_(false) {
  LoadData();
}

PngSource::PngSource(PngSource&& other) noexcept
    : filename_(std::move(other.filename_)),
      image_data_(std::move(other.image_data_)),
      dimensions_(other.dimensions_),
      data_loaded_(other.data_loaded_) {
  other.data_loaded_ = false;
}

PngSource& PngSource::operator=(PngSource&& other) noexcept {
  if (this != &other) {
    filename_ = std::move(other.filename_);
    image_data_ = std::move(other.image_data_);
    dimensions_ = other.dimensions_;
    data_loaded_ = other.data_loaded_;
    other.data_loaded_ = false;
  }
  return *this;
}

ImageDimensions PngSource::GetDimensions() const {
  if (!data_loaded_) {
    LoadData();
  }
  return dimensions_;
}

TileSize PngSource::GetIdealTileSize() const {
  if (!data_loaded_) {
    LoadData();
  }
  // PNG is not tiled, so return the full image size as ideal tile
  return TileSize(dimensions_.width, dimensions_.height);
}

void PngSource::LoadData() const {
  if (data_loaded_) {
    return;
  }

  unsigned int width, height;
  unsigned int error = lodepng::decode(image_data_, width, height, filename_);

  if (error) {
    throw std::runtime_error("PNG decode error: " +
                             std::string(lodepng_error_text(error)));
  }

  dimensions_ = ImageDimensions(static_cast<int>(width),
                                static_cast<int>(height), 4);  // RGBA
  data_loaded_ = true;
}

Tile PngSource::GetTile(int x, int y, int width, int height) const {
  if (!data_loaded_) {
    LoadData();
  }

  // Clamp tile bounds to image dimensions
  int actual_width = std::min(width, dimensions_.width - x);
  int actual_height = std::min(height, dimensions_.height - y);

  if (actual_width <= 0 || actual_height <= 0) {
    return Tile(x, y, 0, 0, dimensions_.channels);
  }

  Tile tile(x, y, actual_width, actual_height, dimensions_.channels);

  // Copy data from full image to tile
  for (int row = 0; row < actual_height; ++row) {
    int src_row_offset = (y + row) * dimensions_.width * dimensions_.channels;
    int dst_row_offset = row * actual_width * dimensions_.channels;

    std::memcpy(tile.data.data() + dst_row_offset,
                image_data_.data() + src_row_offset + x * dimensions_.channels,
                actual_width * dimensions_.channels);
  }

  return tile;
}

}  // namespace fim
