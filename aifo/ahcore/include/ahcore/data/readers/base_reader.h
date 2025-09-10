// Copyright 2024 Jonas Teuwen. All Rights Reserved.
// Copyright 2025 Jonas Teuwen & Joren Brunekreef. All Rights Reserved.
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
#ifndef AIFO_AHCORE_INCLUDE_AHCORE_DATA_READERS_BASE_READER_H_
#define AIFO_AHCORE_INCLUDE_AHCORE_DATA_READERS_BASE_READER_H_

#include <vips/vips8>

#include <spdlog/spdlog.h>
#include <algorithm>
#include <cmath>
#include <filesystem>
#include <iostream>
#include <map>
#include <memory>
#include <stdexcept>
#include <string>
#include <utility>
#include <vector>
#include "absl/status/status.h"
#include "absl/status/statusor.h"
#include "ahcore/data/metadata.h"
#include "ahcore/utilities/torch.h"
#include "aifocore/concepts/numeric.h"
#include "aifocore/status/status_macros.h"
#include "aifocore/tiling/grid.h"
#include "aifocore/utilities/fmt.h"
#include "dlup/backends/abstract.h"

namespace aifo::data::readers {

namespace fs = std::filesystem;
using aifocore::Size;

/**
 * @class PostStitchTransform
 * @brief Abstract base class for post-stitching image transformations.
 */
class PostStitchTransform {
 public:
  PostStitchTransform() : metadata_(aifo::data::Metadata::Create()) {}

  explicit PostStitchTransform(std::shared_ptr<Metadata> metadata)
      : metadata_(std::move(metadata)) {}

  virtual ~PostStitchTransform() = default;

  // Rule of five: define copy/move operations
  PostStitchTransform(const PostStitchTransform&) = default;
  PostStitchTransform& operator=(const PostStitchTransform&) = default;
  PostStitchTransform(PostStitchTransform&&) = default;
  PostStitchTransform& operator=(PostStitchTransform&&) = default;

  [[nodiscard]] virtual vips::VImage Apply(const vips::VImage& image) const = 0;

  // Get the metadata associated with this transform
  [[nodiscard]] std::shared_ptr<Metadata> GetMetadata() const {
    return metadata_;
  }

 protected:
  std::shared_ptr<Metadata> metadata_;
};

/**
 * @class IdentityTransform
 * @brief A transform that returns the input image unchanged.
 */
class IdentityTransform : public PostStitchTransform {
 public:
  IdentityTransform() {
    metadata_->Set(MetadataKeys::NumOutputChannels, 1)->Lock();
  }

  explicit IdentityTransform(std::shared_ptr<Metadata> metadata)
      : PostStitchTransform(std::move(metadata)) {}

  [[nodiscard]] vips::VImage Apply(const vips::VImage& image) const override {
    return image;
  }
};

/**
 * @class AveragingTransform
 * @brief A transform that returns the average across bands using bandrank.
 */
class AveragingTransform : public PostStitchTransform {
 public:
  AveragingTransform() {
    metadata_->Set(MetadataKeys::NumOutputChannels, 1)
        ->Set(MetadataKeys::OutputPixelFormat, VIPS_FORMAT_UCHAR)
        ->Set(MetadataKeys::OutputInterpretation, VIPS_INTERPRETATION_B_W)
        ->Lock();
  }

  explicit AveragingTransform(std::shared_ptr<Metadata> metadata)
      : PostStitchTransform(std::move(metadata)) {}

  [[nodiscard]] vips::VImage Apply(const vips::VImage& image) const override {
    // TODO(jonasteuwen): This needs a VIPS pipeline to work efficiently.
    auto image_result = aifo::utilities::VImageToTensor(image);
    if (!image_result.ok()) {
      throw std::runtime_error(
          aifocore::fmt::format("Failed to convert image to tensor: {}",
                                image_result.status().message()));
    }
    auto tensor = image_result.value().argmax(0).unsqueeze(0).to(torch::kUInt8);

    auto tensor_result = aifo::utilities::TensorToVImage(tensor);
    if (!tensor_result.ok()) {
      throw std::runtime_error(
          aifocore::fmt::format("Failed to convert tensor to VImage: {}",
                                tensor_result.status().message()));
    }

    return tensor_result.value();
  }
};

/**
 * @enum StitchingMode
 * @brief Modes for stitching overlapping tiles.
 */
enum class StitchingMode { kCrop, kAverage, kMaximum };

/**
 * @class Reader
 * @brief Abstract base class for reading image tiles and metadata.
 */
class ImageReader : public dlup::backends::AbstractSlideBackend {
 public:
  ImageReader(const fs::path& filename, StitchingMode stitching_mode,
              std::shared_ptr<PostStitchTransform> post_transform = nullptr)
      : AbstractSlideBackend(filename),
        filename_(filename),
        stitching_mode_(stitching_mode),
        post_transform_(post_transform ? post_transform
                                       : std::make_shared<IdentityTransform>()),
        mpp_(0.0),
        input_pixel_format_(VIPS_FORMAT_UCHAR),
        input_interpretation_(VIPS_INTERPRETATION_B_W),
        output_pixel_format_(VIPS_FORMAT_UCHAR),
        output_interpretation_(VIPS_INTERPRETATION_B_W),
        num_channels_(1),
        num_cols_(0),
        num_rows_(0),
        order_(aifocore::tiling::GridOrder::kC),
        output_channels_(-1) {}

  virtual ~ImageReader() = default;

  virtual absl::Status Open() = 0;
  void Close() override = 0;

  [[nodiscard]] virtual absl::StatusOr<vips::VImage> ReadTile(
      int index) const = 0;
  [[nodiscard]] vips::VImage GetImage() const;

  // Input format getters
  [[nodiscard]] VipsBandFormat GetInputPixelFormat() const {
    return input_pixel_format_;
  }

  [[nodiscard]] VipsInterpretation GetInputInterpretation() const {
    return input_interpretation_;
  }

  // Output format getters
  [[nodiscard]] VipsBandFormat GetOutputPixelFormat() const {
    return output_pixel_format_;
  }

  [[nodiscard]] VipsInterpretation GetOutputInterpretation() const {
    return output_interpretation_;
  }

  // Deprecated format getters
  [[deprecated("Use GetInputPixelFormat() instead")]] VipsBandFormat
  GetPixelFormat() const {
    return GetInputPixelFormat();
  }

  [[deprecated("Use GetInputInterpretation() instead")]] VipsInterpretation
  GetInterpretation() const {
    return GetInputInterpretation();
  }

  // Returns the number of input channels from the source data
  [[nodiscard]] int GetInputChannels() const { return num_channels_; }

  // Returns the number of output channels after transform
  [[nodiscard]] int GetNumOutputChannels() const {
    // Use output_channels_ if it's set
    if (output_channels_ > 0) {
      return output_channels_;
    }
    // Fall back to input channels if output is not explicitly set
    return GetNumInputChannels();
  }

  // Maintains backward compatibility by returning output channels
  [[deprecated(
      "Use GetNumInputChannels() instead. This method will be removed in a "
      "future release.")]] int
  GetNumChannels() const {
    return GetNumInputChannels();
  }

  int GetNumInputChannels() const { return num_channels_; }

  [[nodiscard]] std::optional<double> GetMagnification() const override {
    // Return magnification, if available.
    return std::nullopt;
  }

  [[nodiscard]] std::optional<std::string> GetVendor() const override {
    // Return vendor information, if available.
    return std::nullopt;
  }

  [[nodiscard]] std::vector<std::string> GetProperties() const override {
    // Return slide-specific properties.
    return {};
  }

  /**
   * @brief Read a region of the image. The region is extracted demand-driven.
   * @param location The top-left corner of the region.
   * @param level The pyramid level to read from. Level 0 is the only supported level.
   * @param size The size of the region.
   * @return A StatusOr containing a demand-driven image of the requested region.
   */
  [[nodiscard]] absl::StatusOr<vips::VImage> ReadRegion(
      const Size<int, 2>& coordinates, int level,
      const Size<int, 2>& size) const override {
    return ReadRegionImpl(coordinates - geometry_.offset, level, size);
  }

  /**
   * @brief Provides access to the metadata object.
   *
   * @return A pointer to the metadata object.
   */
  Metadata* GetMetadata() { return metadata_.get(); }

  void ComputeParameters() {
    try {
      this->geometry_ = this->GetMetadata()->GetGeometry();
      this->actual_size_ = this->geometry_.bounds;
      this->tile_size_ = this->GetMetadata()->GetTileSize();
      this->tile_overlap_ = this->GetMetadata()->GetTileOverlap();
      this->mpp_ = this->GetMetadata()->GetMpp();
      this->input_pixel_format_ =
          static_cast<VipsBandFormat>(this->GetMetadata()->GetPixelFormat());
      this->input_interpretation_ = static_cast<VipsInterpretation>(
          this->GetMetadata()->GetInterpretation());
      this->num_channels_ = this->GetMetadata()->GetNumChannels();

      // Initialize output formats from post_transform metadata if available
      if (post_transform_ && post_transform_->GetMetadata()->HasKey(
                                 MetadataKeys::OutputPixelFormat)) {
        this->output_pixel_format_ = static_cast<VipsBandFormat>(
            post_transform_->GetMetadata()->Get<int>(
                MetadataKeys::OutputPixelFormat));
      } else {
        this->output_pixel_format_ = this->input_pixel_format_;
      }

      if (post_transform_ && post_transform_->GetMetadata()->HasKey(
                                 MetadataKeys::OutputInterpretation)) {
        this->output_interpretation_ = static_cast<VipsInterpretation>(
            post_transform_->GetMetadata()->Get<int>(
                MetadataKeys::OutputInterpretation));
      } else {
        this->output_interpretation_ = this->input_interpretation_;
      }

      // Check if the PostStitchTransform has NumOutputChannels in its metadata
      if (post_transform_ && post_transform_->GetMetadata()->HasKey(
                                 MetadataKeys::NumOutputChannels)) {
        this->output_channels_ = post_transform_->GetMetadata()->Get<int>(
            MetadataKeys::NumOutputChannels);
      } else {
        // Default to input channels if not specified
        this->output_channels_ =
            -1;  // Use negative value to indicate it's not explicitly set
      }

      if (this->GetMetadata()->HasKey(MetadataKeys::GridOrder)) {
        this->order_ = this->GetMetadata()->GetGridOrder();
      } else {
        this->order_ = aifocore::tiling::GridOrder::kC;
      }
      this->tile_indices_ =
          this->GetMetadata()->Get<std::vector<int>>(MetadataKeys::TileIndices);

      this->metadata_loaded_ = true;
    } catch (const std::exception& e) {
      throw std::runtime_error(aifocore::fmt::format(
          "Failed to compute parameters from metadata: {}", e.what()));
    }

    stride_ = Size<int, 2>{tile_size_[0] - tile_overlap_[0],
                           tile_size_[1] - tile_overlap_[1]};
    num_cols_ =
        static_cast<int>(std::ceil((geometry_.bounds[0] - tile_overlap_[0]) /
                                   static_cast<double>(stride_[0])));
    num_rows_ =
        static_cast<int>(std::ceil((geometry_.bounds[1] - tile_overlap_[1]) /
                                   static_cast<double>(stride_[1])));
  }

 protected:
  fs::path filename_;
  StitchingMode stitching_mode_;
  std::shared_ptr<PostStitchTransform> post_transform_;
  dlup::SlideGeometry geometry_;
  Size<int, 2> tile_size_;
  Size<int, 2> tile_overlap_;
  Size<int, 2> stride_;
  Size<int, 2> actual_size_;
  double mpp_;
  VipsBandFormat input_pixel_format_;
  VipsInterpretation input_interpretation_;
  VipsBandFormat output_pixel_format_;
  VipsInterpretation output_interpretation_;
  int num_channels_;
  int num_cols_;
  int num_rows_;
  aifocore::tiling::GridOrder order_;
  bool metadata_loaded_ = false;
  std::shared_ptr<Metadata> metadata_;  ///< Pointer to the metadata object.
  // The indices of the tiles actually written compared to the internal grid
  std::vector<int> tile_indices_;
  int output_channels_;

  vips::VImage StitchTiles(const Size<int, 2>& location,
                           const Size<int, 2>& region_size) const;

 private:
  [[nodiscard]] vips::VImage ReadRegionImpl(const Size<int, 2>& coordinates,
                                            int level,
                                            const Size<int, 2>& size) const {
    bool allow_out_of_bounds = true;

    if (level > 0) {
      throw std::runtime_error(
          "Reading regions at levels different from 0 is not supported.");
    }

    if (!allow_out_of_bounds && (coordinates[0] < 0 || coordinates[1] < 0 ||
                                 coordinates[0] + size[0] > actual_size_[0] ||
                                 coordinates[1] + size[1] > actual_size_[1])) {
      throw std::runtime_error("Requested region is out of bounds.");
    }

    // If we allow out of bounds and the region
    // is completely outside, return black image
    if (allow_out_of_bounds &&
        (coordinates[0] >= actual_size_[0] ||  // Entirely to the right
         coordinates[1] >= actual_size_[1] ||  // Entirely below
         coordinates[0] + size[0] <= 0 ||      // Entirely to the left
         coordinates[1] + size[1] <= 0)) {     // Entirely above
      auto black_image =
          vips::VImage::black(
              size[0], size[1],
              vips::VImage::option()->set("bands", GetNumOutputChannels()))
              .cast(GetOutputPixelFormat());

      return black_image;
    }

    // If we're partially out of bounds,
    // create a black image and insert the valid region
    if (allow_out_of_bounds && (coordinates[0] < 0 || coordinates[1] < 0 ||
                                coordinates[0] + size[0] > actual_size_[0] ||
                                coordinates[1] + size[1] > actual_size_[1])) {
      vips::VImage black =
          vips::VImage::black(
              size[0], size[1],
              vips::VImage::option()->set("bands", GetNumOutputChannels()))
              .cast(GetOutputPixelFormat());

      // Calculate the valid region coordinates
      int valid_x = std::max(0, coordinates[0]);
      int valid_y = std::max(0, coordinates[1]);

      int offset_x = valid_x - coordinates[0];
      int offset_y = valid_y - coordinates[1];

      int valid_width = std::min(actual_size_[0] - valid_x, size[0] - offset_x);
      int valid_height =
          std::min(actual_size_[1] - valid_y, size[1] - offset_y);

      if (valid_width > 0 && valid_height > 0) {
        vips::VImage valid_region =
            StitchTiles(Size<int, 2>{valid_x, valid_y},
                        Size<int, 2>{valid_width, valid_height});
        // Calculate where to insert in the black image
        int insert_x = valid_x - coordinates[0];
        int insert_y = valid_y - coordinates[1];
        black = black.insert(valid_region, insert_x, insert_y);
      } else {
        spdlog::warn("No valid region to extract (width={}, height={})",
                     valid_width, valid_height);
      }
      return black;
    }
    return StitchTiles(coordinates, size);
  }
};

vips::VImage ImageReader::StitchTiles(const Size<int, 2>& location,
                                      const Size<int, 2>& region_size) const {
  int x = location[0];
  int y = location[1];
  int width = region_size[0];
  int height = region_size[1];

  if (x < 0 || y < 0 || x + width > actual_size_[0] ||
      y + height > actual_size_[1]) {
    throw std::runtime_error(aifocore::fmt::format(
        "Requested region out of bounds. Only request regions within the "
        "visible region of the slide. Got ({}, {}) with size ({}, {}) for "
        "slide with size ({}, {})",
        x, y, width, height, actual_size_[0], actual_size_[1]));
  }

  vips::VImage stitched_image =
      vips::VImage::black(
          width, height,
          vips::VImage::option()->set("bands", GetNumInputChannels()))
          .cast(GetOutputPixelFormat());

  vips::VImage count_mask = (stitching_mode_ == StitchingMode::kAverage)
                                ? vips::VImage::black(width, height)
                                : vips::VImage();

  int start_col = std::min(x / stride_[0], num_cols_ - 1);
  int start_row = std::min(y / stride_[1], num_rows_ - 1);
  int end_row = std::min((y + height - 1) / stride_[1] + 1, num_rows_);
  int end_col = std::min((x + width - 1) / stride_[0] + 1, num_cols_);

  for (int row = start_row; row < end_row; ++row) {
    for (int col = start_col; col < end_col; ++col) {
      int tile_index = (order_ == aifocore::tiling::GridOrder::kC)
                           ? (row * num_cols_ + col)
                           : (col * num_rows_ + row);

      int actual_index = tile_indices_[tile_index];
      if (actual_index == -1) {
        continue;
      }

      absl::StatusOr<vips::VImage> tile = ReadTile(actual_index);
      if (!tile.ok()) {
        throw std::runtime_error(std::string(tile.status().message()));
      }

      int start_x = col * stride_[0] - x;
      int end_x = start_x + tile_size_[0];
      int start_y = row * stride_[1] - y;
      int end_y = start_y + tile_size_[1];

      int img_start_x = std::max(0, start_x);
      int img_end_x = std::min(width, end_x);
      int img_start_y = std::max(0, start_y);
      int img_end_y = std::min(height, end_y);

      int crop_start_x = img_start_x - start_x;
      int crop_end_x = img_end_x - start_x;
      int crop_start_y = img_start_y - start_y;
      int crop_end_y = img_end_y - start_y;

      int crop_width = crop_end_x - crop_start_x;
      int crop_height = crop_end_y - crop_start_y;

      try {
        vips::VImage cropped_tile = tile.value().crop(
            crop_start_x, crop_start_y, crop_width, crop_height);

        if (stitching_mode_ == StitchingMode::kCrop) {
          stitched_image =
              stitched_image.insert(cropped_tile, img_start_x, img_start_y);
        } else if (stitching_mode_ == StitchingMode::kAverage) {
          vips::VImage existing_region = stitched_image.extract_area(
              img_start_x, img_start_y, crop_width, crop_height);
          vips::VImage existing_counts = count_mask.extract_area(
              img_start_x, img_start_y, crop_width, crop_height);

          vips::VImage updated_region = existing_region + cropped_tile;
          vips::VImage updated_counts = existing_counts + 1;

          stitched_image =
              stitched_image.insert(updated_region, img_start_x, img_start_y);
          count_mask =
              count_mask.insert(updated_counts, img_start_x, img_start_y);

        } else if (stitching_mode_ == StitchingMode::kMaximum) {
          vips::VImage existing_region = stitched_image.extract_area(
              img_start_x, img_start_y, crop_width, crop_height);
          vips::VImage mask = existing_region > cropped_tile;
          vips::VImage max_region =
              mask.ifthenelse(existing_region, cropped_tile);
          stitched_image =
              stitched_image.insert(max_region, img_start_x, img_start_y);
        }
      } catch (const std::exception& e) {
        spdlog::error("Failed to read tile at index {}: {}", actual_index,
                      e.what());
        // Continue with the next tile
        continue;
      }
    }
  }

  if (stitching_mode_ == StitchingMode::kAverage) {
    // Replace any zeros in the count_mask with ones to avoid division by zero
    count_mask = count_mask.ifthenelse(count_mask, 1.0);

    // Convert both images to floating point for accurate division
    stitched_image = stitched_image.cast(VIPS_FORMAT_FLOAT);
    count_mask = count_mask.cast(VIPS_FORMAT_FLOAT);

    // Divide to compute the average, then cast back to the original format
    stitched_image = stitched_image / count_mask;
  }

  vips::VImage transformed_image = post_transform_->Apply(stitched_image);
  return transformed_image;
}

vips::VImage ImageReader::GetImage() const {
  const auto [width, height] = geometry_.size;

  // Step 1: Create a new VIPS pipeline
  VipsImage* pipeline = vips_image_new();
  if (pipeline == nullptr) {
    throw std::runtime_error("Failed to create VIPS image pipeline");
  }

  // Step 2: Initialize pipeline fields
  vips_image_init_fields(pipeline, width, height, GetNumOutputChannels(),
                         GetOutputPixelFormat(), VIPS_CODING_NONE,
                         GetOutputInterpretation(), 1.0, 1.0);

  // Set resolution if mpp_ is available
  if (mpp_) {
    pipeline->Xres = 1000.0 / mpp_;
    pipeline->Yres = 1000.0 / mpp_;
  }

  // Step 3: Configure demand-driven pipeline style
  if (vips_image_pipelinev(pipeline, VIPS_DEMAND_STYLE_SMALLTILE, nullptr)) {
    g_object_unref(pipeline);
    throw std::runtime_error("Pipeline setup failed");
  }

  // Step 4: Connect pipeline to the reader's tile generation logic
  auto generate_region = [](VipsRegion* region, void* seq, void* state,
                            void* unused, gboolean* stop) -> int {
    (void)seq;     // Suppress unused parameter warning
    (void)unused;  // Suppress unused parameter warning

    auto* reader = static_cast<ImageReader*>(state);
    try {
      // Compute the location and size of the region
      Size<int, 2> location = {region->valid.left, region->valid.top};
      Size<int, 2> size = {region->valid.width, region->valid.height};

      // Read the requested region
      absl::StatusOr<vips::VImage> region_data =
          reader->ReadRegion(location, 0, size);
      // TODO(jonasteuwen): This needs to be handled with the macros
      if (!region_data.ok()) {
        throw std::runtime_error("Could not read region.");
      }

      // Directly map the memory of region_data into the VIPS region
      unsigned char* memory = const_cast<unsigned char*>(
          static_cast<const unsigned char*>(region_data.value().data()));
      region->data = memory;  // Assign raw memory directly
    } catch (const std::exception& e) {
      std::cerr << "Error during region generation: " << e.what() << '\n';
      *stop = true;  // Stop pipeline on error
      return -1;
    }
    return 0;
  };

  // Step 5: Generate image using the pipeline
  if (vips_image_generate(pipeline, nullptr, generate_region, nullptr,
                          const_cast<ImageReader*>(this), nullptr)) {
    g_object_unref(pipeline);
    throw std::runtime_error("Failed to generate image pipeline");
  }

  // Step 6: Wrap and return the pipeline as a VImage
  return vips::VImage(pipeline);
}

}  // namespace aifo::data::readers

#endif  // AIFO_AHCORE_INCLUDE_AHCORE_DATA_READERS_BASE_READER_H_
