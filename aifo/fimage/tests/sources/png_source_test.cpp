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

#include "fim/sources/png_source.h"

#include <gtest/gtest.h>

#include <filesystem>
#include <string>
#include <utility>
#include <vector>

#include "fim/types.h"
#include "lodepng/lodepng.h"

namespace fim {

class PngSourceTest : public ::testing::Test {
 protected:
  void SetUp() override {
    // Create test PNG files with unique names to avoid conflicts
    CreateTestPngFile("png_source_test_rgb.png", 100, 50, 3);
    CreateTestPngFile("png_source_test_rgba.png", 80, 60, 4);
    CreateTestPngFile("png_source_test_grayscale.png", 40, 30, 1);
  }

  void TearDown() override {
    // Clean up test files
    std::filesystem::remove("png_source_test_rgb.png");
    std::filesystem::remove("png_source_test_rgba.png");
    std::filesystem::remove("png_source_test_grayscale.png");
  }

 private:
  void CreateTestPngFile(const std::string& filename, int width, int height,
                         int channels) {
    std::vector<uint8_t> image_data(width * height * channels);

    // Fill with test pattern
    for (int y = 0; y < height; ++y) {
      for (int x = 0; x < width; ++x) {
        int pixel_offset = (y * width + x) * channels;
        uint8_t value = static_cast<uint8_t>((y * width + x) % 256);

        for (int c = 0; c < channels; ++c) {
          image_data[pixel_offset + c] = value;
        }
      }
    }

    // Write PNG file using lodepng
    unsigned error = 0;
    if (channels == 3) {
      error = lodepng_encode24_file(filename.c_str(), image_data.data(), width,
                                    height);
    } else if (channels == 4) {
      error = lodepng_encode32_file(filename.c_str(), image_data.data(), width,
                                    height);
    } else if (channels == 1) {
      error = lodepng_encode_file(filename.c_str(), image_data.data(), width,
                                  height, LCT_GREY, 8);
    }

    if (error != 0) {
      throw std::runtime_error("Failed to create test PNG file: " +
                               std::string(lodepng_error_text(error)));
    }
  }
};

// Test PNG source construction and basic properties
TEST_F(PngSourceTest, Construction) {
  PngSource source("png_source_test_rgb.png");

  auto dims = source.GetDimensions();
  EXPECT_EQ(dims.width, 100);
  EXPECT_EQ(dims.height, 50);
  EXPECT_EQ(dims.channels, 4);  // PNG source always loads as RGBA
}

// Test PNG source with different file formats
TEST_F(PngSourceTest, DifferentFormats) {
  PngSource rgb_source("png_source_test_rgb.png");
  PngSource rgba_source("png_source_test_rgba.png");
  PngSource grayscale_source("png_source_test_grayscale.png");

  auto rgb_dims = rgb_source.GetDimensions();
  auto rgba_dims = rgba_source.GetDimensions();
  auto grayscale_dims = grayscale_source.GetDimensions();

  EXPECT_EQ(rgb_dims.width, 100);
  EXPECT_EQ(rgb_dims.height, 50);
  EXPECT_EQ(rgb_dims.channels, 4);  // All PNG sources load as RGBA

  EXPECT_EQ(rgba_dims.width, 80);
  EXPECT_EQ(rgba_dims.height, 60);
  EXPECT_EQ(rgba_dims.channels, 4);

  EXPECT_EQ(grayscale_dims.width, 40);
  EXPECT_EQ(grayscale_dims.height, 30);
  EXPECT_EQ(grayscale_dims.channels, 4);
}

// Test PNG source with non-existent file
TEST_F(PngSourceTest, NonExistentFile) {
  EXPECT_THROW(PngSource("non_existent.png"), std::runtime_error);
}

// Test ideal tile size
TEST_F(PngSourceTest, IdealTileSize) {
  PngSource source("png_source_test_rgb.png");

  auto tile_size = source.GetIdealTileSize();
  EXPECT_EQ(tile_size.width, 100);  // Should be full image width
  EXPECT_EQ(tile_size.height, 50);  // Should be full image height
}

// Test getting a tile from the full image
TEST_F(PngSourceTest, GetFullTile) {
  PngSource source("png_source_test_rgb.png");

  auto tile = source.GetTile(0, 0, 100, 50);
  EXPECT_EQ(tile.x, 0);
  EXPECT_EQ(tile.y, 0);
  EXPECT_EQ(tile.width, 100);
  EXPECT_EQ(tile.height, 50);
  EXPECT_EQ(tile.channels, 4);
  EXPECT_EQ(tile.data.size(), 100 * 50 * 4);
}

// Test getting a partial tile
TEST_F(PngSourceTest, GetPartialTile) {
  PngSource source("png_source_test_rgb.png");

  auto tile = source.GetTile(10, 20, 30, 15);
  EXPECT_EQ(tile.x, 10);
  EXPECT_EQ(tile.y, 20);
  EXPECT_EQ(tile.width, 30);
  EXPECT_EQ(tile.height, 15);
  EXPECT_EQ(tile.channels, 4);
  EXPECT_EQ(tile.data.size(), 30 * 15 * 4);
}

// Test getting a tile beyond image bounds
TEST_F(PngSourceTest, GetTileBeyondBounds) {
  PngSource source("png_source_test_rgb.png");

  auto tile = source.GetTile(80, 40, 50, 20);
  EXPECT_EQ(tile.x, 80);
  EXPECT_EQ(tile.y, 40);
  EXPECT_EQ(tile.width, 20);   // Clamped to image bounds: 100 - 80 = 20
  EXPECT_EQ(tile.height, 10);  // Clamped to image bounds: 50 - 40 = 10
  EXPECT_EQ(tile.channels, 4);
}

// Test getting a tile completely outside image bounds
TEST_F(PngSourceTest, GetTileOutsideBounds) {
  PngSource source("png_source_test_rgb.png");

  auto tile = source.GetTile(150, 100, 20, 20);
  EXPECT_EQ(tile.x, 150);
  EXPECT_EQ(tile.y, 100);
  EXPECT_EQ(tile.width, 0);
  EXPECT_EQ(tile.height, 0);
  EXPECT_EQ(tile.channels, 4);
  EXPECT_TRUE(tile.data.empty());
}

// Test getting a tile with negative coordinates (should clamp to valid region)
TEST_F(PngSourceTest, GetTileNegativeCoordinates) {
  PngSource source("png_source_test_rgb.png");  // 100x50 image

  // PNG source clamps negative coordinates to (0,0) like MemorySource
  auto tile = source.GetTile(-10, -20, 30, 40);
  EXPECT_EQ(tile.x, -10);      // Original requested x
  EXPECT_EQ(tile.y, -20);      // Original requested y
  EXPECT_EQ(tile.width, 30);   // Clamped and valid
  EXPECT_EQ(tile.height, 40);  // Clamped and valid
  EXPECT_EQ(tile.channels, 4);
}

// Test move constructor
TEST_F(PngSourceTest, MoveConstructor) {
  PngSource source1("png_source_test_rgb.png");
  auto dims1 = source1.GetDimensions();

  PngSource source2(std::move(source1));
  auto dims2 = source2.GetDimensions();

  EXPECT_EQ(dims2.width, dims1.width);
  EXPECT_EQ(dims2.height, dims1.height);
  EXPECT_EQ(dims2.channels, dims1.channels);
}

// Test move assignment
TEST_F(PngSourceTest, MoveAssignment) {
  PngSource source1("png_source_test_rgb.png");
  PngSource source2("png_source_test_rgba.png");

  auto dims1 = source1.GetDimensions();
  source2 = std::move(source1);
  auto dims2 = source2.GetDimensions();

  EXPECT_EQ(dims2.width, dims1.width);
  EXPECT_EQ(dims2.height, dims1.height);
  EXPECT_EQ(dims2.channels, dims1.channels);
}

// Test multiple tile accesses (lazy loading)
TEST_F(PngSourceTest, MultipleTileAccess) {
  PngSource source("png_source_test_rgb.png");

  // First access should load the data
  auto tile1 = source.GetTile(0, 0, 50, 25);
  EXPECT_EQ(tile1.width, 50);
  EXPECT_EQ(tile1.height, 25);

  // Second access should reuse loaded data
  auto tile2 = source.GetTile(50, 25, 50, 25);
  EXPECT_EQ(tile2.width, 50);
  EXPECT_EQ(tile2.height, 25);

  // Data should be consistent
  EXPECT_EQ(tile1.channels, tile2.channels);
}

// Test small image
TEST_F(PngSourceTest, SmallImage) {
  PngSource source("png_source_test_grayscale.png");

  auto dims = source.GetDimensions();
  EXPECT_EQ(dims.width, 40);
  EXPECT_EQ(dims.height, 30);
  EXPECT_EQ(dims.channels, 4);

  auto tile_size = source.GetIdealTileSize();
  EXPECT_EQ(tile_size.width, 40);
  EXPECT_EQ(tile_size.height, 30);
}

}  // namespace fim
