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
 * @file types.h
 * @brief Core type definitions for the fim image processing library.
 * @author Jonas Teuwen
 * @date 2025
 *
 * This file contains the fundamental data structures used throughout the fim
 * library for image processing, including tile representation, image
 * dimensions, and tile size specifications.
 */
#ifndef AIFO_FIMAGE_INCLUDE_FIM_TYPES_H_
#define AIFO_FIMAGE_INCLUDE_FIM_TYPES_H_

#include <cstdint>
#include <vector>

namespace fim {

/**
 * @brief Represents a rectangular tile of image data.
 *
 * A tile is a rectangular region of an image that can be processed
 * independently. This structure contains both the tile's position and
 * dimensions, as well as the actual pixel data and metadata about padding
 * requirements.
 */
struct Tile {
  int x = 0;                  ///< X coordinate of the tile's top-left corner
  int y = 0;                  ///< Y coordinate of the tile's top-left corner
  int width = 0;              ///< Width of the tile in pixels
  int height = 0;             ///< Height of the tile in pixels
  int channels = 0;           ///< Number of color channels per pixel
  std::vector<uint8_t> data;  ///< Raw pixel data in row-major order

  // Optional padding information for border tiles
  int expected_width = 0;      ///< Full tile width expected by the format
  int expected_height = 0;     ///< Full tile height expected by the format
  bool needs_padding = false;  ///< Whether this tile needs padding for output

  /**
   * @brief Default constructor creates an empty tile.
   */
  Tile() = default;

  /**
   * @brief Constructs a tile with specified dimensions and allocates data
   * storage.
   *
   * @param x X coordinate of the tile's top-left corner
   * @param y Y coordinate of the tile's top-left corner
   * @param width Width of the tile in pixels
   * @param height Height of the tile in pixels
   * @param channels Number of color channels per pixel
   */
  Tile(int x, int y, int width, int height, int channels)
      : x(x), y(y), width(width), height(height), channels(channels) {
    data.resize(width * height * channels);
  }

  /**
   * @brief Constructs a tile with padding information.
   *
   * @param x X coordinate of the tile's top-left corner
   * @param y Y coordinate of the tile's top-left corner
   * @param width Actual width of the tile in pixels
   * @param height Actual height of the tile in pixels
   * @param channels Number of color channels per pixel
   * @param expected_width Expected full tile width
   * @param expected_height Expected full tile height
   */
  Tile(int x, int y, int width, int height, int channels, int expected_width,
       int expected_height)
      : x(x),
        y(y),
        width(width),
        height(height),
        channels(channels),
        expected_width(expected_width),
        expected_height(expected_height),
        needs_padding(width < expected_width || height < expected_height) {
    data.resize(width * height * channels);
  }

  /**
   * @brief Gets the size of the actual tile data in bytes.
   *
   * @return Size of the tile data in bytes
   */
  size_t GetDataSize() const { return width * height * channels; }

  /**
   * @brief Gets the expected full tile size in bytes.
   *
   * If padding information is available, returns the size for the full
   * expected tile dimensions. Otherwise, returns the actual data size.
   *
   * @return Expected full tile size in bytes
   */
  size_t GetExpectedDataSize() const {
    if (expected_width > 0 && expected_height > 0) {
      return expected_width * expected_height * channels;
    }
    return GetDataSize();
  }

  /**
   * @brief Creates zero-padded data for full tile size.
   *
   * If padding is not needed, returns the original data. Otherwise,
   * creates a new vector with the expected dimensions and copies the
   * actual tile data to the top-left corner, filling the rest with zeros.
   *
   * @return Vector containing padded tile data
   */
  std::vector<uint8_t> CreatePaddedData() const {
    if (!needs_padding) {
      return data;
    }

    std::vector<uint8_t> padded_data(
        expected_width * expected_height * channels, 0);

    // Copy actual tile data to top-left corner of padded buffer
    for (int y = 0; y < height; ++y) {
      for (int x = 0; x < width; ++x) {
        int src_offset = (y * width + x) * channels;
        int dst_offset = (y * expected_width + x) * channels;

        for (int c = 0; c < channels; ++c) {
          padded_data[dst_offset + c] = data[src_offset + c];
        }
      }
    }

    return padded_data;
  }
};

/**
 * @brief Represents the dimensions of an image.
 *
 * This structure encapsulates the width, height, and number of channels
 * of an image, providing a convenient way to pass around image size
 * information.
 */
struct ImageDimensions {
  int width = 0;     ///< Width of the image in pixels
  int height = 0;    ///< Height of the image in pixels
  int channels = 0;  ///< Number of color channels per pixel

  /**
   * @brief Default constructor creates zero-dimensional image.
   */
  ImageDimensions() = default;

  /**
   * @brief Constructs image dimensions with specified values.
   *
   * @param w Width of the image in pixels
   * @param h Height of the image in pixels
   * @param c Number of color channels per pixel
   */
  ImageDimensions(int w, int h, int c) : width(w), height(h), channels(c) {}
};

/**
 * @brief Represents the preferred tile size for processing.
 *
 * This structure specifies the optimal tile dimensions for efficient
 * processing of an image source. Different image formats may have
 * different optimal tile sizes based on their internal structure.
 */
struct TileSize {
  int width = 0;   ///< Preferred tile width in pixels
  int height = 0;  ///< Preferred tile height in pixels

  /**
   * @brief Default constructor creates zero-sized tile.
   */
  TileSize() = default;

  /**
   * @brief Constructs tile size with specified dimensions.
   *
   * @param w Preferred tile width in pixels
   * @param h Preferred tile height in pixels
   */
  TileSize(int w, int h) : width(w), height(h) {}
};

}  // namespace fim

#endif  // AIFO_FIMAGE_INCLUDE_FIM_TYPES_H_
