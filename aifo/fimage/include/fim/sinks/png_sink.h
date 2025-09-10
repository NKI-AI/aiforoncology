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
 * @file png_sink.h
 * @brief PNG image sink implementation for the fim library.
 * @author Jonas Teuwen
 * @date 2025
 *
 * This file contains the PngSink class which provides PNG image writing
 * capabilities for the fim image processing pipeline. It uses the lodepng
 * library for PNG encoding and supports grayscale, RGB, and RGBA formats.
 */
#ifndef AIFO_FIMAGE_INCLUDE_FIM_SINKS_PNG_SINK_H_
#define AIFO_FIMAGE_INCLUDE_FIM_SINKS_PNG_SINK_H_

#include <algorithm>
#include <iostream>
#include <stdexcept>
#include <string>
#include <vector>

#include "fim/pipeline.h"
#include "fim/types.h"
#include "lodepng/lodepng.h"

namespace fim {

/**
 * @brief PNG image sink for writing PNG files.
 *
 * This class provides a sink implementation for PNG image files using
 * the lodepng library. It supports grayscale (1 channel), RGB (3 channels),
 * and RGBA (4 channels) formats. The sink assembles the complete image
 * from tiles and writes it to a PNG file.
 *
 * The class follows the CRTP pattern by inheriting from SinkBase and
 * provides the required interface methods for image sinks.
 */
class PngSink : public SinkBase<PngSink> {
 public:
  /**
   * @brief Constructs a PNG sink with the specified output filename.
   *
   * @param filename Path to the output PNG file where the image will be written
   */
  explicit PngSink(const fs::path& filename);

  /**
   * @brief Default destructor.
   */
  ~PngSink() = default;

  /**
   * @brief Renders the input to a PNG file.
   *
   * This method processes the entire input by fetching tiles and assembling
   * them into a complete image, then writes the result to a PNG file using
   * lodepng. The method supports 1, 3, and 4 channel images.
   *
   * @tparam InputType The type of the input source or operator
   * @param input The input source or operator to render
   * @throw std::runtime_error if the image dimensions are invalid or PNG
   * writing fails
   */
  template <typename InputType>
  void Render(const InputType& input);

 private:
  /**
   * @brief Assembles the complete image from tiles.
   *
   * This method fetches tiles from the input source and assembles them
   * into a complete image buffer. It uses the input's ideal tile size
   * to optimize the tiling process.
   *
   * @tparam InputType The type of the input source or operator
   * @param input The input source or operator to assemble
   * @param image_data Output vector to store the assembled image data
   */
  template <typename InputType>
  void AssembleImage(const InputType& input, std::vector<uint8_t>& image_data);
};

template <typename InputType>
void PngSink::Render(const InputType& input) {
  auto dims = input.GetDimensions();

  if (dims.width <= 0 || dims.height <= 0) {
    throw std::runtime_error("Invalid image dimensions for PNG output");
  }

  // Assemble full image data
  std::vector<uint8_t> image_data;
  AssembleImage(input, image_data);

  // Write PNG using lodepng
  unsigned png_error = 0;
  if (dims.channels == 3) {
    // RGB data
    png_error = lodepng_encode24_file(filename_.c_str(), image_data.data(),
                                      dims.width, dims.height);
  } else if (dims.channels == 4) {
    // RGBA data
    png_error = lodepng_encode32_file(filename_.c_str(), image_data.data(),
                                      dims.width, dims.height);
  } else if (dims.channels == 1) {
    // Grayscale data
    png_error = lodepng_encode_file(filename_.c_str(), image_data.data(),
                                    dims.width, dims.height, LCT_GREY, 8);
  } else {
    throw std::runtime_error("Unsupported channel count for PNG output: " +
                             std::to_string(dims.channels));
  }

  if (png_error != 0) {
    throw std::runtime_error("Failed to write PNG file " + filename_.string() +
                             " (error " + std::to_string(png_error) +
                             "): " + lodepng_error_text(png_error));
  }
}

template <typename InputType>
void PngSink::AssembleImage(const InputType& input,
                            std::vector<uint8_t>& image_data) {
  auto dims = input.GetDimensions();
  auto ideal_tile_size = input.GetIdealTileSize();

  // Use ideal tile size, but ensure we don't exceed image dimensions
  int tile_width = std::min(ideal_tile_size.width, dims.width);
  int tile_height = std::min(ideal_tile_size.height, dims.height);

  // If ideal tile size is 0, use reasonable defaults
  if (tile_width <= 0)
    tile_width = 256;
  if (tile_height <= 0)
    tile_height = 256;

  // Resize image data to hold the full image
  image_data.resize(dims.width * dims.height * dims.channels);

  // Process image in tiles
  for (int y = 0; y < dims.height; y += tile_height) {
    for (int x = 0; x < dims.width; x += tile_width) {
      int actual_width = std::min(tile_width, dims.width - x);
      int actual_height = std::min(tile_height, dims.height - y);

      // Get tile from input
      Tile tile = input.GetTile(x, y, actual_width, actual_height);

      // Copy tile data to appropriate position in full image
      for (int tile_y = 0; tile_y < tile.height; ++tile_y) {
        for (int tile_x = 0; tile_x < tile.width; ++tile_x) {
          int src_offset = (tile_y * tile.width + tile_x) * dims.channels;
          int dst_offset =
              ((y + tile_y) * dims.width + (x + tile_x)) * dims.channels;

          for (int c = 0; c < dims.channels; ++c) {
            image_data[dst_offset + c] = tile.data[src_offset + c];
          }
        }
      }
    }
  }
}

}  // namespace fim

#endif  // AIFO_FIMAGE_INCLUDE_FIM_SINKS_PNG_SINK_H_
