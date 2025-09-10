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
 * @file memory_source.cpp
 * @brief Implementation of memory-based image source for the fim library.
 * @author Jonas Teuwen
 * @date 2025
 *
 * This file contains the implementation of the MemorySource class which
 * provides efficient tile-based access to image data stored in memory.
 */
#include "fim/sources/memory_source.h"

#include <algorithm>
#include <cstring>
#include <utility>

namespace fim {

MemorySource::MemorySource(std::vector<uint8_t> image_data,
                           ImageDimensions dimensions, TileSize tile_size)
    : image_data_(std::move(image_data)),
      dimensions_(dimensions),
      tile_size_(tile_size) {}

MemorySource::MemorySource(MemorySource&& other) noexcept
    : image_data_(std::move(other.image_data_)),
      dimensions_(other.dimensions_),
      tile_size_(other.tile_size_) {}

MemorySource& MemorySource::operator=(MemorySource&& other) noexcept {
  if (this != &other) {
    image_data_ = std::move(other.image_data_);
    dimensions_ = other.dimensions_;
    tile_size_ = other.tile_size_;
  }
  return *this;
}

ImageDimensions MemorySource::GetDimensions() const {
  return dimensions_;
}

TileSize MemorySource::GetIdealTileSize() const {
  return tile_size_;
}

Tile MemorySource::GetTile(int x, int y, int width, int height) const {
  // Clamp tile bounds to image dimensions
  int actual_x = std::max(0, x);
  int actual_y = std::max(0, y);
  int actual_width = std::min(width, dimensions_.width - actual_x);
  int actual_height = std::min(height, dimensions_.height - actual_y);

  // Handle case where tile is completely outside image bounds
  if (actual_width <= 0 || actual_height <= 0) {
    return Tile(x, y, 0, 0, dimensions_.channels);
  }

  // Create output tile
  Tile tile(x, y, actual_width, actual_height, dimensions_.channels);

  // Copy data from image buffer to tile
  for (int tile_y = 0; tile_y < actual_height; ++tile_y) {
    for (int tile_x = 0; tile_x < actual_width; ++tile_x) {
      int src_x = actual_x + tile_x;
      int src_y = actual_y + tile_y;
      int src_offset =
          (src_y * dimensions_.width + src_x) * dimensions_.channels;
      int dst_offset = (tile_y * actual_width + tile_x) * dimensions_.channels;

      for (int c = 0; c < dimensions_.channels; ++c) {
        tile.data[dst_offset + c] = image_data_[src_offset + c];
      }
    }
  }

  return tile;
}

size_t MemorySource::GetMemoryUsage() const {
  return image_data_.size();
}

}  // namespace fim
