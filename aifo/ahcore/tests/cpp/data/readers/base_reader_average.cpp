// Copyright 2024 Jonas Teuwen. All Rights Reserved.
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
#include <gtest/gtest.h>
#include <filesystem>
#include <string>
#include <vector>
#include "absl/status/status.h"
#include "ahcore/data/readers/base_reader.h"
#include "vips/vips8"

using namespace aifo::data::readers;
namespace fs = std::filesystem;

// A dummy subclass of ImageReader for testing.
// It sets up a 200x200 image divided into 4 tiles (2x2 grid),
// with each tile being 100x100 and overlapping 50 pixels.
class DummyImageReader : public ImageReader {
 public:
  DummyImageReader(const std::string& filename, StitchingMode mode)
      : ImageReader(filename, mode) {
    // Set up dummy metadata:
    geometry_ = dlup::SlideGeometry{
        {150, 150},
        {0, 0},
        {150, 150}};  // Actual image size accounting for 50px overlap
    actual_size_ = {150, 150};
    tile_size_ = {100, 100};
    tile_overlap_ = {50, 50};
    stride_ = tile_size_ - tile_overlap_;  // (50, 50)
    num_cols_ = 2;
    num_rows_ = 2;
    // Use four tiles with indices 0, 1, 2, 3
    tile_indices_ = {0, 1, 2, 3};
    order_ =
        aifocore::tiling::GridOrder::kC;  // Set to C-order (row-major) layout

    // Set pixel format, interpretation, and channel count.
    input_pixel_format_ = VIPS_FORMAT_UCHAR;
    input_interpretation_ = VIPS_INTERPRETATION_sRGB;
    // Initialize output formats to match input formats
    output_pixel_format_ = input_pixel_format_;
    output_interpretation_ = input_interpretation_;
    num_channels_ = 1;
    output_channels_ = 1;

    // Initialize metadata
    metadata_ = aifo::data::Metadata::Create();
    metadata_->SetGeometry(geometry_)
        ->SetTileSize(tile_size_)
        ->SetTileOverlap(tile_overlap_)
        ->Set(MetadataKeys::PixelFormat, static_cast<int>(input_pixel_format_))
        ->Set(MetadataKeys::Interpretation,
              static_cast<int>(input_interpretation_))
        ->Set(MetadataKeys::NumChannels, num_channels_)
        ->Set(MetadataKeys::TileIndices, tile_indices_)
        ->SetGridOrder(order_);

    metadata_loaded_ = true;
  }

  absl::Status Open() override {
    // No-op for dummy.
    return absl::OkStatus();
  }

  void Close() override {
    // No cleanup necessary.
  }

  // Override ReadTile to return an image of constant value (tile index * 10).
  // For example, tile 0 returns an image of 0's, tile 1 returns 10's, etc.
  absl::StatusOr<vips::VImage> ReadTile(int index) const override {
    int width = tile_size_[0];
    int height = tile_size_[1];
    vips::VImage tile = vips::VImage::black(
        width, height,
        vips::VImage::option()->set("bands", GetNumInputChannels()));

    // Set the format (important to prevent cast errors)
    tile = tile.cast(input_pixel_format_);

    // Add the value based on tile index
    tile = tile + index * 10;
    return tile;
  }

  // Public wrapper for the protected StitchTiles function.
  vips::VImage PublicStitchTiles(const Size<int, 2>& location,
                                 const Size<int, 2>& size) const {
    return StitchTiles(location, size);
  }
};

TEST(AverageStitchingTest, CorrectAveraging) {
  // Create a dummy reader in averaging mode.
  DummyImageReader reader("dummy_path", StitchingMode::kAverage);
  reader.Open();

  // Stitch the entire image (150x150) using the public wrapper.
  vips::VImage stitched = reader.PublicStitchTiles({0, 0}, {150, 150});

  // The grid layout in C-order (row-major) with 50px overlap:
  // - Tile 0 (value 0) is at (0,0) covering (0,0) to (100,100)
  // - Tile 1 (value 10) is at (50,0) covering (50,0) to (150,100)
  // - Tile 2 (value 20) is at (0,50) covering (0,50) to (100,150)
  // - Tile 3 (value 30) is at (50,50) covering (50,50) to (150,150)

  // Test non-overlapping regions (single tile contribution)
  double top_left = stitched(25, 25)[0];  // Only tile 0
  EXPECT_NEAR(top_left, 0.0, 1e-6);

  double top_right = stitched(125, 25)[0];  // Only tile 1
  EXPECT_NEAR(top_right, 10.0, 1e-6);

  double bottom_left = stitched(25, 125)[0];  // Only tile 2
  EXPECT_NEAR(bottom_left, 20.0, 1e-6);

  double bottom_right = stitched(125, 125)[0];  // Only tile 3
  EXPECT_NEAR(bottom_right, 30.0, 1e-6);

  // Test overlapping regions
  double horizontal_overlap = stitched(75, 25)[0];  // Overlap of tiles 0 and 1
  EXPECT_NEAR(horizontal_overlap, 5.0, 1e-6);       // Average of 0 and 10

  double vertical_overlap = stitched(25, 75)[0];  // Overlap of tiles 0 and 2
  EXPECT_NEAR(vertical_overlap, 10.0, 1e-6);      // Average of 0 and 20

  double center = stitched(75, 75)[0];  // Overlap of all four tiles
  EXPECT_NEAR(center, 15.0, 1e-6);      // Average of 0, 10, 20, and 30

  reader.Close();
}
