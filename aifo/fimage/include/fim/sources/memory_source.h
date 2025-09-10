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
 * @file memory_source.h
 * @brief Memory-based image source implementation for the fim library.
 * @author Jonas Teuwen
 * @date 2025
 *
 * This file contains the MemorySource class which provides image reading
 * capabilities from in-memory data for the fim image processing pipeline.
 * It's particularly useful for pyramid generation where intermediate levels
 * are kept in memory for efficient downsampling.
 */
#ifndef AIFO_FIMAGE_INCLUDE_FIM_SOURCES_MEMORY_SOURCE_H_
#define AIFO_FIMAGE_INCLUDE_FIM_SOURCES_MEMORY_SOURCE_H_

#include <algorithm>
#include <vector>

#include "fim/pipeline.h"
#include "fim/types.h"

namespace fim {

/**
 * @brief Memory-based image source for in-memory image data.
 *
 * This class provides a source implementation for image data stored in memory.
 * It's designed for efficient tile-based access to image data that's already
 * loaded in memory, making it ideal for pyramid generation and processing
 * intermediate results.
 *
 * The class follows the CRTP pattern by inheriting from SourceBase and
 * provides the required interface methods for image sources.
 */
class MemorySource : public SourceBase<MemorySource> {
 public:
  /**
   * @brief Constructs a memory source from existing image data.
   *
   * @param image_data Vector containing the image data in row-major order
   * @param dimensions Image dimensions including width, height, and channels
   * @param tile_size Preferred tile size for efficient access
   */
  MemorySource(std::vector<uint8_t> image_data, ImageDimensions dimensions,
               TileSize tile_size = TileSize(256, 256));

  /**
   * @brief Move constructor.
   *
   * @param other MemorySource to move from
   */
  MemorySource(MemorySource&& other) noexcept;

  /**
   * @brief Move assignment operator.
   *
   * @param other MemorySource to move from
   * @return Reference to this instance
   */
  MemorySource& operator=(MemorySource&& other) noexcept;

  /**
   * @brief Deleted copy constructor.
   */
  MemorySource(const MemorySource&) = delete;

  /**
   * @brief Deleted copy assignment operator.
   */
  MemorySource& operator=(const MemorySource&) = delete;

  /**
   * @brief Default destructor.
   */
  ~MemorySource() = default;

  /**
   * @brief Gets the dimensions of the image.
   *
   * @return ImageDimensions containing the image dimensions
   */
  ImageDimensions GetDimensions() const;

  /**
   * @brief Gets the ideal tile size for processing.
   *
   * @return TileSize containing the ideal tile dimensions
   */
  TileSize GetIdealTileSize() const;

  /**
   * @brief Gets a tile from the image data.
   *
   * @param x X coordinate of the top-left corner of the tile
   * @param y Y coordinate of the top-left corner of the tile
   * @param width Width of the requested tile
   * @param height Height of the requested tile
   * @return Tile containing the requested image data
   */
  Tile GetTile(int x, int y, int width, int height) const;

  /**
   * @brief Gets the memory usage of this source in bytes.
   *
   * @return Memory usage in bytes
   */
  size_t GetMemoryUsage() const;

 private:
  std::vector<uint8_t> image_data_;  ///< Image data in row-major order
  ImageDimensions dimensions_;       ///< Image dimensions
  TileSize tile_size_;               ///< Preferred tile size
};

}  // namespace fim

#endif  // AIFO_FIMAGE_INCLUDE_FIM_SOURCES_MEMORY_SOURCE_H_
