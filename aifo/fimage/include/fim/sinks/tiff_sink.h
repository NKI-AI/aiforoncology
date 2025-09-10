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
 * @file tiff_sink.h
 * @brief TIFF image sink implementation for the fim library.
 * @author Jonas Teuwen
 * @date 2025
 *
 * This file contains the TiffSink class which provides TIFF image writing
 * capabilities for the fim image processing pipeline. It uses libtiff for
 * TIFF encoding and supports both tiled and strip-based TIFF output with
 * configurable compression and tile sizes. It also supports pyramidal TIFF
 * generation for efficient multi-resolution image storage.
 */
#ifndef AIFO_FIMAGE_INCLUDE_FIM_SINKS_TIFF_SINK_H_
#define AIFO_FIMAGE_INCLUDE_FIM_SINKS_TIFF_SINK_H_

#include <tiffio.h>

#include <algorithm>
#include <iostream>
#include <memory>
#include <stdexcept>
#include <string>
#include <utility>
#include <vector>

#include "fim/operators/downsample.h"
#include "fim/pipeline.h"
#include "fim/sources/memory_source.h"
#include "fim/types.h"

namespace fim {

/**
 * @brief TIFF image sink for writing single-page and pyramidal TIFF files.
 *
 * This class provides a unified sink implementation for TIFF image files using
 * libtiff. It supports both single-page and pyramidal (multi-page) TIFF output
 * and can handle various channel configurations (grayscale, RGB, RGBA).
 *
 * For pyramidal output, it creates multiple pages with progressively smaller
 * resolutions until the smallest dimension is below the tile size. The
 * implementation optimizes memory usage by switching from source-based to
 * memory-based downsampling when image size becomes manageable.
 *
 * The class follows the CRTP pattern by inheriting from SinkBase and provides
 * the required interface methods for image sinks.
 */
class TiffSink : public SinkBase<TiffSink> {
 public:
  /**
   * @brief Constructs a TIFF sink with the specified output filename.
   *
   * Uses default tile size of 256x256 pixels for tiled output.
   * Creates a single-page TIFF file.
   *
   * @param filename Path to the output TIFF file where the image will be
   * written
   */
  explicit TiffSink(const fs::path& filename);

  /**
   * @brief Constructs a TIFF sink with custom tile size.
   *
   * Creates a single-page TIFF file with the specified tile size.
   *
   * @param filename Path to the output TIFF file where the image will be
   * written
   * @param tile_size Preferred tile size for tiled output
   */
  explicit TiffSink(const fs::path& filename, const TileSize& tile_size);

  /**
   * @brief Constructs a TIFF sink with full configuration including pyramidal
   * support.
   *
   * @param filename Path to the output TIFF file where the image will be
   * written
   * @param tile_size Preferred tile size for all pyramid levels
   * @param pyramidal Whether to create a pyramidal (multi-page) TIFF
   * @param memory_threshold_mb Memory threshold in MB for switching to
   * memory-based downsampling
   * @param downsample_factor Factor by which to downsample each pyramid level
   */
  explicit TiffSink(const fs::path& filename, const TileSize& tile_size,
                    bool pyramidal, size_t memory_threshold_mb = 50,
                    int downsample_factor = 2);

  /**
   * @brief Default destructor.
   */
  ~TiffSink() = default;

  /**
   * @brief Renders the input to a TIFF file.
   *
   * This method processes the input and creates either a single-page or
   * pyramidal TIFF file based on the pyramidal setting. For pyramidal output,
   * it implements the optimized pyramid generation strategy that switches
   * between source-based and memory-based downsampling.
   *
   * @tparam InputType The type of the input source or operator
   * @param input The input source or operator to render
   * @throw std::runtime_error if the TIFF file cannot be created or written
   */
  template <typename InputType>
  void Render(const InputType& input);

 private:
  /**
   * @brief Renders a single-page TIFF file.
   *
   * @tparam InputType The type of the input source or operator
   * @param input The input source or operator to render
   */
  template <typename InputType>
  void RenderSinglePage(const InputType& input);

  /**
   * @brief Renders a pyramidal TIFF file.
   *
   * @tparam InputType The type of the input source or operator
   * @param input The input source or operator to render
   */
  template <typename InputType>
  void RenderPyramidal(const InputType& input);

  /**
   * @brief Writes a single level of the pyramid to the TIFF file.
   *
   * @tparam InputType The type of the input source or operator
   * @param tiff Pointer to the TIFF file handle
   * @param input The input source or operator for this level
   * @param level_index The pyramid level index (0 = full resolution)
   */
  template <typename InputType>
  void WriteLevel(TIFF* tiff, const InputType& input, int level_index = 0);

  /**
   * @brief Writes the image as a tiled TIFF level.
   *
   * @tparam InputType The type of the input source or operator
   * @param tiff Pointer to the TIFF file handle
   * @param input The input source or operator to write
   */
  template <typename InputType>
  void WriteTiledLevel(TIFF* tiff, const InputType& input);

  /**
   * @brief Writes the image as a strip-based TIFF level.
   *
   * @tparam InputType The type of the input source or operator
   * @param tiff Pointer to the TIFF file handle
   * @param input The input source or operator to write
   */
  template <typename InputType>
  void WriteStripLevel(TIFF* tiff, const InputType& input);

  /**
   * @brief Configures TIFF tags for a level.
   *
   * @param tiff Pointer to the TIFF file handle
   * @param dims Image dimensions for this level
   * @param level_index The pyramid level index (0 = full resolution)
   */
  void ConfigureTiffTags(TIFF* tiff, const ImageDimensions& dims,
                         int level_index = 0);

  /**
   * @brief Assembles a complete image level into memory.
   *
   * @tparam InputType The type of the input source or operator
   * @param input The input source or operator to assemble
   * @return Vector containing the assembled image data
   */
  template <typename InputType>
  std::vector<uint8_t> AssembleImageData(const InputType& input);

  /**
   * @brief Calculates the uncompressed memory size of an image level.
   *
   * @param dims Image dimensions
   * @return Memory size in bytes
   */
  static size_t CalculateMemorySize(const ImageDimensions& dims);

  /**
   * @brief Checks if downsampling should continue for the given dimensions.
   *
   * @param dims Current level dimensions
   * @return true if downsampling should continue, false otherwise
   */
  bool ShouldContinueDownsampling(const ImageDimensions& dims) const;

  TileSize tile_size_;             ///< Tile size for all levels
  bool pyramidal_;                 ///< Whether to create pyramidal TIFF
  size_t memory_threshold_bytes_;  ///< Memory threshold for switching strategy
                                   ///< (pyramidal only)
  int downsample_factor_;  ///< Downsample factor between levels (pyramidal
                           ///< only)
};

template <typename InputType>
void TiffSink::Render(const InputType& input) {
  if (pyramidal_) {
    RenderPyramidal(input);
  } else {
    RenderSinglePage(input);
  }
}

template <typename InputType>
void TiffSink::RenderSinglePage(const InputType& input) {
  TIFF* tiff = TIFFOpen(filename_.c_str(), "w");
  if (!tiff) {
    throw std::runtime_error("Failed to create TIFF file: " +
                             filename_.string());
  }

  try {
    WriteLevel(tiff, input, 0);
    TIFFClose(tiff);
  } catch (...) {
    TIFFClose(tiff);
    throw;
  }
}

template <typename InputType>
void TiffSink::RenderPyramidal(const InputType& input) {
  TIFF* tiff = TIFFOpen(filename_.c_str(), "w");
  if (!tiff) {
    throw std::runtime_error("Failed to create TIFF file: " +
                             filename_.string());
  }

  try {
    int level = 0;
    std::unique_ptr<MemorySource> memory_source;

    // Write level 0 (original resolution)
    auto current_dims = input.GetDimensions();
    WriteLevel(tiff, input, level);

    // Check if level 0 should be kept in memory
    if (CalculateMemorySize(current_dims) <= memory_threshold_bytes_) {
      auto image_data = AssembleImageData(input);
      memory_source = std::make_unique<MemorySource>(std::move(image_data),
                                                     current_dims, tile_size_);
    }

    level++;

    while (ShouldContinueDownsampling(current_dims)) {
      // Create next directory for this level
      if (TIFFWriteDirectory(tiff) != 1) {
        throw std::runtime_error("Failed to write TIFF directory");
      }

      if (memory_source) {
        // Downsample from the current memory source
        auto downsampled_input = Downsample(*memory_source, downsample_factor_);
        current_dims = downsampled_input.GetDimensions();
        WriteLevel(tiff, downsampled_input, level);

        // Check if we should keep this level in memory for future levels
        if (CalculateMemorySize(current_dims) <= memory_threshold_bytes_) {
          auto image_data = AssembleImageData(downsampled_input);
          memory_source = std::make_unique<MemorySource>(
              std::move(image_data), current_dims, tile_size_);
        } else {
          // Don't keep this level in memory, but we still have the previous
          // memory source
        }
      } else {
        // Create downsampled operator from original input
        int cumulative_factor = 1;
        for (int i = 0; i < level; ++i) {
          cumulative_factor *= downsample_factor_;
        }
        auto downsampled_input = Downsample(input, cumulative_factor);
        current_dims = downsampled_input.GetDimensions();

        WriteLevel(tiff, downsampled_input, level);

        // Check if we should start keeping levels in memory
        if (CalculateMemorySize(current_dims) <= memory_threshold_bytes_) {
          auto image_data = AssembleImageData(downsampled_input);
          memory_source = std::make_unique<MemorySource>(
              std::move(image_data), current_dims, tile_size_);
        }
      }

      level++;
    }

    TIFFClose(tiff);
  } catch (...) {
    TIFFClose(tiff);
    throw;
  }
}

template <typename InputType>
void TiffSink::WriteLevel(TIFF* tiff, const InputType& input, int level_index) {
  auto dims = input.GetDimensions();
  ConfigureTiffTags(tiff, dims, level_index);

  // Choose between tiled and strip based on dimensions and tile size
  if (dims.width > tile_size_.width || dims.height > tile_size_.height) {
    WriteTiledLevel(tiff, input);
  } else {
    WriteStripLevel(tiff, input);
  }
}

template <typename InputType>
void TiffSink::WriteTiledLevel(TIFF* tiff, const InputType& input) {
  auto dims = input.GetDimensions();

  TIFFSetField(tiff, TIFFTAG_TILEWIDTH, tile_size_.width);
  TIFFSetField(tiff, TIFFTAG_TILELENGTH, tile_size_.height);

  // Write tiles
  for (int y = 0; y < dims.height; y += tile_size_.height) {
    for (int x = 0; x < dims.width; x += tile_size_.width) {
      int actual_width = std::min(tile_size_.width, dims.width - x);
      int actual_height = std::min(tile_size_.height, dims.height - y);

      Tile tile = input.GetTile(x, y, actual_width, actual_height);

      // Set padding information for border tiles
      tile.expected_width = tile_size_.width;
      tile.expected_height = tile_size_.height;
      tile.needs_padding =
          (tile.width < tile_size_.width || tile.height < tile_size_.height);

      std::vector<uint8_t> tiff_data = tile.CreatePaddedData();
      if (TIFFWriteTile(tiff, tiff_data.data(), static_cast<uint32_t>(x),
                        static_cast<uint32_t>(y), 0, 0) < 0) {
        throw std::runtime_error("Failed to write TIFF tile");
      }
    }
  }
}

template <typename InputType>
void TiffSink::WriteStripLevel(TIFF* tiff, const InputType& input) {
  auto dims = input.GetDimensions();

  TIFFSetField(tiff, TIFFTAG_ROWSPERSTRIP, tile_size_.height);

  // Write strips
  for (int y = 0; y < dims.height; y += tile_size_.height) {
    int actual_height = std::min(tile_size_.height, dims.height - y);

    Tile tile = input.GetTile(0, y, dims.width, actual_height);

    if (TIFFWriteEncodedStrip(tiff, y / tile_size_.height, tile.data.data(),
                              tile.data.size()) < 0) {
      throw std::runtime_error("Failed to write TIFF strip");
    }
  }
}

template <typename InputType>
std::vector<uint8_t> TiffSink::AssembleImageData(const InputType& input) {
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
  std::vector<uint8_t> image_data(dims.width * dims.height * dims.channels);

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

  return image_data;
}

}  // namespace fim

#endif  // AIFO_FIMAGE_INCLUDE_FIM_SINKS_TIFF_SINK_H_
