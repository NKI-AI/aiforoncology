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
 * @file image.h
 * @brief Main image class and convenience functions for the fim library.
 * @author Jonas Teuwen
 * @date 2025
 *
 * This file contains the main Image class template and convenience functions
 * for creating and working with images in the fim image processing pipeline.
 * It provides a high-level interface for common image operations and source
 * creation.
 */
#ifndef AIFO_FIMAGE_INCLUDE_FIM_IMAGE_H_
#define AIFO_FIMAGE_INCLUDE_FIM_IMAGE_H_

#include <utility>

#include "fim/operators/crop.h"
#include "fim/operators/downsample.h"
#include "fim/sinks/png_sink.h"
#include "fim/sinks/tiff_sink.h"
#include "fim/sources/memory_source.h"
#include "fim/sources/png_source.h"
#include "fim/sources/tiff_source.h"

namespace fim {

/**
 * @brief Main Image class template that wraps pipeline stages.
 *
 * This class provides a high-level interface for image processing operations
 * by wrapping pipeline sources and operators. It supports method chaining
 * for building complex processing pipelines and provides convenient methods
 * for common operations like cropping, downsampling, and rendering.
 *
 * The class uses the CRTP pattern internally but provides a simpler interface
 * for end users. Each operation returns a new Image instance wrapping the
 * appropriate operator or source.
 *
 * @tparam SourceType The type of the underlying source or operator
 *
 * @code
 * // Example usage:
 * auto image = fim::CreateTiffImage("input.tiff");
 * image.Crop(100, 100, 500, 500)
 *      .Downsample(2)
 *      .Render(fim::PngSink("output.png"));
 * @endcode
 */
template <typename SourceType>
class Image {
 public:
  /**
   * @brief Constructs an Image from a source object.
   *
   * This constructor takes ownership of the source object and wraps it
   * in the Image interface. The source object is moved to avoid unnecessary
   * copying.
   *
   * @param source The source object to wrap (moved)
   */
  explicit Image(SourceType&& source) : source_(std::move(source)) {}

  /**
   * @brief Constructs an Image from a filename.
   *
   * This constructor is available for sources that support construction
   * from a filename. It creates the appropriate source type and wraps it
   * in the Image interface.
   *
   * @param filename Path to the image file
   */
  explicit Image(const fs::path& filename) : source_(filename) {}

  /**
   * @brief Applies a crop operation to the image.
   *
   * This method creates a new Image instance with a crop operator applied
   * to the current source. The crop region is specified by the top-left
   * corner and dimensions.
   *
   * @param x X coordinate of the top-left corner of the crop region
   * @param y Y coordinate of the top-left corner of the crop region
   * @param width Width of the crop region
   * @param height Height of the crop region
   * @return New Image instance with the crop operator applied
   */
  auto Crop(int x, int y, int width, int height) {
    return Image<fim::Crop<SourceType>>(
        fim::Crop<SourceType>(source_, x, y, width, height));
  }

  /**
   * @brief Applies a downsample operation to the image.
   *
   * This method creates a new Image instance with a downsample operator
   * applied to the current source. The downsample factor determines how
   * much the image size is reduced.
   *
   * @param factor The downsampling factor (must be positive)
   * @return New Image instance with the downsample operator applied
   */
  auto Downsample(int factor) {
    return Image<fim::Downsample<SourceType>>(
        fim::Downsample<SourceType>(source_, factor));
  }

  /**
   * @brief Renders the image to a sink.
   *
   * This method processes the entire image pipeline and writes the result
   * to the specified sink. The sink determines the output format and
   * destination.
   *
   * @tparam SinkType The type of the sink (e.g., PngSink, TiffSink)
   * @param sink The sink to render to
   */
  template <typename SinkType>
  void Render(SinkType sink) {
    sink.Render(source_);
  }

  /**
   * @brief Gets the underlying source for advanced use.
   *
   * This method provides access to the underlying source or operator
   * for advanced operations that are not covered by the high-level
   * Image interface.
   *
   * @return Const reference to the underlying source
   */
  const SourceType& GetSource() const { return source_; }

 private:
  SourceType source_;  ///< The underlying source or operator
};

/**
 * @brief Creates an Image from a TIFF file.
 *
 * This convenience function creates an Image instance wrapping a TiffSource
 * for the specified file. It provides a simple way to start processing
 * TIFF images.
 *
 * @param filename Path to the TIFF file
 * @return Image instance wrapping a TiffSource
 * @throw std::runtime_error if the TIFF file cannot be opened
 */
inline Image<TiffSource> CreateTiffImage(const fs::path& filename) {
  return Image<TiffSource>(TiffSource(filename));
}

/**
 * @brief Creates an Image from a PNG file.
 *
 * This convenience function creates an Image instance wrapping a PngSource
 * for the specified file. It provides a simple way to start processing
 * PNG images.
 *
 * @param filename Path to the PNG file
 * @return Image instance wrapping a PngSource
 * @throw std::runtime_error if the PNG file cannot be opened or decoded
 */
inline Image<PngSource> CreatePngImage(const fs::path& filename) {
  return Image<PngSource>(PngSource(filename));
}

}  // namespace fim

#endif  // AIFO_FIMAGE_INCLUDE_FIM_IMAGE_H_
