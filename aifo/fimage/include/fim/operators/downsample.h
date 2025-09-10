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
 * @file downsample.h
 * @brief Downsample operator implementation for the fim library.
 * @author Jonas Teuwen
 * @date 2025
 *
 * This file contains the Downsample operator class which provides image
 * downsampling capabilities for the fim image processing pipeline. It reduces
 * image size by a specified factor using average pooling for anti-aliasing.
 */
#ifndef AIFO_FIMAGE_INCLUDE_FIM_OPERATORS_DOWNSAMPLE_H_
#define AIFO_FIMAGE_INCLUDE_FIM_OPERATORS_DOWNSAMPLE_H_

#include <algorithm>

#include "fim/pipeline.h"
#include "fim/types.h"

namespace fim {

/**
 * @brief Downsample operator for reducing image size using average pooling.
 *
 * This operator reduces the size of the input image by a specified integer
 * factor using average pooling. Each output pixel is computed as the average
 * of the corresponding factor×factor region in the input image. This provides
 * anti-aliasing to reduce visual artifacts that would occur with simple
 * decimation.
 *
 * The operator works by mapping output tile coordinates to input coordinates
 * and computing the average of the corresponding pixel regions. This allows
 * for efficient processing without loading the entire image into memory.
 *
 * @tparam InputType The type of the input source or operator
 */
template <typename InputType>
class Downsample : public OperatorBase<Downsample<InputType>, InputType> {
 public:
  /**
   * @brief Constructs a downsample operator with the specified factor.
   *
   * The downsample factor determines how much the image size is reduced.
   * A factor of 2 reduces both width and height by half, a factor of 4
   * reduces them to a quarter, and so on.
   *
   * @param input The input source or operator to downsample
   * @param factor The downsampling factor (must be positive)
   * @throw std::invalid_argument if factor is not positive
   */
  Downsample(const InputType& input, int factor)
      : OperatorBase<Downsample<InputType>, InputType>(input), factor_(factor) {
    if (factor <= 0) {
      throw std::invalid_argument("Downsample factor must be positive");
    }
  }

  /**
   * @brief Gets the dimensions of the downsampled output.
   *
   * This method returns the dimensions of the image after downsampling.
   * The output dimensions are computed by dividing the input dimensions
   * by the downsample factor and rounding up.
   *
   * @return ImageDimensions containing the downsampled image dimensions
   */
  ImageDimensions GetDimensions() const {
    auto input_dims = this->input_.GetDimensions();

    int output_width = (input_dims.width + factor_ - 1) / factor_;
    int output_height = (input_dims.height + factor_ - 1) / factor_;

    return ImageDimensions(output_width, output_height, input_dims.channels);
  }

  /**
   * @brief Gets the ideal tile size for processing the downsampled output.
   *
   * This method returns the ideal tile size scaled down by the downsample
   * factor. This maintains efficient processing characteristics while
   * accounting for the reduced output size.
   *
   * @return TileSize containing the ideal tile dimensions for downsampled
   * output
   */
  TileSize GetIdealTileSize() const {
    auto input_tile_size = this->input_.GetIdealTileSize();
    return TileSize(input_tile_size.width / factor_,
                    input_tile_size.height / factor_);
  }

  /**
   * @brief Gets a tile from the downsampled output.
   *
   * This method retrieves a rectangular region of the downsampled output
   * as a tile. It computes the corresponding input region, fetches the
   * necessary input data, and applies average pooling to generate the
   * downsampled output.
   *
   * @param x X coordinate of the top-left corner of the tile in output space
   * @param y Y coordinate of the top-left corner of the tile in output space
   * @param width Width of the requested tile
   * @param height Height of the requested tile
   * @return Tile containing the requested downsampled image data
   */
  Tile GetTile(int x, int y, int width, int height) const {
    auto output_dims = GetDimensions();

    // Clamp tile bounds to output dimensions
    int actual_width = std::min(width, output_dims.width - x);
    int actual_height = std::min(height, output_dims.height - y);

    if (actual_width <= 0 || actual_height <= 0) {
      return Tile(x, y, 0, 0, output_dims.channels);
    }

    // Calculate input region needed for this output tile
    int input_x = x * factor_;
    int input_y = y * factor_;
    int input_width = actual_width * factor_;
    int input_height = actual_height * factor_;

    // Get input tile
    Tile input_tile =
        this->input_.GetTile(input_x, input_y, input_width, input_height);

    // Create output tile
    Tile output_tile(x, y, actual_width, actual_height, output_dims.channels);

    // Perform average pooling
    PerformAveragePooling(input_tile, output_tile);

    return output_tile;
  }

 private:
  int factor_;  ///< The downsampling factor

  /**
   * @brief Performs average pooling on the input tile to generate output.
   *
   * This method applies average pooling to reduce the input tile size
   * by the downsample factor. Each output pixel is computed as the
   * average of the corresponding factor×factor region in the input.
   *
   * @param input_tile The input tile to downsample
   * @param output_tile The output tile to populate with downsampled data
   */
  void PerformAveragePooling(const Tile& input_tile, Tile& output_tile) const {
    int channels = input_tile.channels;

    for (int out_y = 0; out_y < output_tile.height; ++out_y) {
      for (int out_x = 0; out_x < output_tile.width; ++out_x) {
        // Calculate the input region for this output pixel
        int in_x_start = out_x * factor_;
        int in_y_start = out_y * factor_;
        int in_x_end = std::min(in_x_start + factor_, input_tile.width);
        int in_y_end = std::min(in_y_start + factor_, input_tile.height);

        // Average the pixels in the region
        for (int c = 0; c < channels; ++c) {
          int sum = 0;
          int count = 0;

          for (int in_y = in_y_start; in_y < in_y_end; ++in_y) {
            for (int in_x = in_x_start; in_x < in_x_end; ++in_x) {
              int input_index = (in_y * input_tile.width + in_x) * channels + c;
              sum += input_tile.data[input_index];
              count++;
            }
          }

          int output_index = (out_y * output_tile.width + out_x) * channels + c;
          output_tile.data[output_index] = static_cast<uint8_t>(sum / count);
        }
      }
    }
  }
};

}  // namespace fim

#endif  // AIFO_FIMAGE_INCLUDE_FIM_OPERATORS_DOWNSAMPLE_H_
