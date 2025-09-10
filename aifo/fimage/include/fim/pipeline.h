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
 * @file pipeline.h
 * @brief Base classes for the fim image processing pipeline.
 * @author Jonas Teuwen
 * @date 2025
 *
 * This file contains the CRTP (Curiously Recurring Template Pattern) base
 * classes that define the interface for sources, operators, and sinks in the
 * fim image processing pipeline. These classes provide a consistent interface
 * for all pipeline components.
 */
#ifndef AIFO_FIMAGE_INCLUDE_FIM_PIPELINE_H_
#define AIFO_FIMAGE_INCLUDE_FIM_PIPELINE_H_

#include <filesystem>

#include "fim/types.h"

namespace fim {

namespace fs = std::filesystem;

/**
 * @brief CRTP base class for all pipeline sources.
 *
 * This class provides a common interface for all image sources in the pipeline.
 * Image sources are responsible for providing image data, either from files
 * or other sources. The CRTP pattern ensures that all source implementations
 * provide the required methods while maintaining compile-time polymorphism.
 *
 * @tparam Derived The derived class that inherits from this base class
 */
template <typename Derived>
class SourceBase {
 public:
  /**
   * @brief Gets the dimensions of the source image.
   *
   * This method returns the complete dimensions of the source image,
   * including width, height, and number of channels.
   *
   * @return ImageDimensions containing the source image dimensions
   */
  ImageDimensions GetDimensions() const {
    return static_cast<const Derived*>(this)->GetDimensions();
  }

  /**
   * @brief Gets the ideal tile size for processing this source.
   *
   * Different image formats may have different optimal tile sizes
   * for efficient processing. This method returns the recommended
   * tile size for the specific source implementation.
   *
   * @return TileSize containing the ideal tile dimensions
   */
  TileSize GetIdealTileSize() const {
    return static_cast<const Derived*>(this)->GetIdealTileSize();
  }

  /**
   * @brief Gets a tile from the source image.
   *
   * This method retrieves a rectangular region of the source image
   * as a tile. The tile coordinates are clamped to the image bounds.
   *
   * @param x X coordinate of the top-left corner of the tile
   * @param y Y coordinate of the top-left corner of the tile
   * @param width Width of the requested tile
   * @param height Height of the requested tile
   * @return Tile containing the requested image data
   */
  Tile GetTile(int x, int y, int width, int height) const {
    return static_cast<const Derived*>(this)->GetTile(x, y, width, height);
  }
};

/**
 * @brief CRTP base class for all pipeline operators.
 *
 * This class provides a common interface for all image processing operators
 * in the pipeline. Operators transform image data from one form to another,
 * such as cropping, scaling, or filtering. Each operator takes an input
 * source and produces modified output.
 *
 * @tparam Derived The derived class that inherits from this base class
 * @tparam InputType The type of the input source or operator
 */
template <typename Derived, typename InputType>
class OperatorBase {
 public:
  /**
   * @brief Constructs an operator with the specified input.
   *
   * @param input The input source or operator that provides data to this
   * operator
   */
  explicit OperatorBase(const InputType& input) : input_(input) {}

  /**
   * @brief Gets the dimensions of the output after applying this operator.
   *
   * This method returns the dimensions of the image after the operator
   * has been applied. The dimensions may be different from the input
   * dimensions depending on the operator's function.
   *
   * @return ImageDimensions containing the output image dimensions
   */
  ImageDimensions GetDimensions() const {
    return static_cast<const Derived*>(this)->GetDimensions();
  }

  /**
   * @brief Gets the ideal tile size for processing this operator's output.
   *
   * Returns the recommended tile size for efficiently processing the
   * output of this operator. This may be influenced by the input's
   * tile size and the operator's specific requirements.
   *
   * @return TileSize containing the ideal tile dimensions
   */
  TileSize GetIdealTileSize() const {
    return static_cast<const Derived*>(this)->GetIdealTileSize();
  }

  /**
   * @brief Gets a tile from the operator's output.
   *
   * This method retrieves a rectangular region of the operator's output
   * as a tile. The operator applies its transformation to the input data
   * to produce the requested output tile.
   *
   * @param x X coordinate of the top-left corner of the tile
   * @param y Y coordinate of the top-left corner of the tile
   * @param width Width of the requested tile
   * @param height Height of the requested tile
   * @return Tile containing the processed image data
   */
  Tile GetTile(int x, int y, int width, int height) const {
    return static_cast<const Derived*>(this)->GetTile(x, y, width, height);
  }

 protected:
  const InputType& input_;  ///< Reference to the input source or operator
};

/**
 * @brief CRTP base class for all pipeline sinks.
 *
 * This class provides a common interface for all image sinks in the pipeline.
 * Sinks are responsible for outputting processed image data to various
 * destinations, such as files or memory buffers. They represent the end
 * points of the processing pipeline.
 *
 * @tparam Derived The derived class that inherits from this base class
 */
template <typename Derived>
class SinkBase {
 public:
  /**
   * @brief Constructs a sink with the specified output filename.
   *
   * @param filename Path to the output file where the image will be written
   */
  explicit SinkBase(const fs::path& filename) : filename_(filename) {}

  /**
   * @brief Renders the input to the sink's output destination.
   *
   * This method processes the entire input and writes it to the sink's
   * output destination. The specific behavior depends on the sink
   * implementation (e.g., PNG file, TIFF file, etc.).
   *
   * @tparam InputType The type of the input source or operator
   * @param input The input source or operator to render
   */
  template <typename InputType>
  void Render(const InputType& input) {
    static_cast<Derived*>(this)->Render(input);
  }

 protected:
  fs::path filename_;  ///< Path to the output file
};

}  // namespace fim

#endif  // AIFO_FIMAGE_INCLUDE_FIM_PIPELINE_H_
