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

#include "fim/sources/memory_source.h"

#include <gtest/gtest.h>

#include <algorithm>
#include <memory>
#include <utility>
#include <vector>

#include "fim/types.h"

namespace fim {

class MemorySourceTest : public ::testing::Test {
 protected:
  void SetUp() override {
    // Create test data for various test scenarios
    CreateTestData();
  }

  void TearDown() override {
    // No cleanup needed for memory-based tests
  }

 private:
  void CreateTestData() {
    // RGB test image 10x8x3
    rgb_data_ = CreateImageData(10, 8, 3);
    rgb_dims_ = ImageDimensions(10, 8, 3);

    // RGBA test image 6x4x4
    rgba_data_ = CreateImageData(6, 4, 4);
    rgba_dims_ = ImageDimensions(6, 4, 4);

    // Grayscale test image 20x15x1
    grayscale_data_ = CreateImageData(20, 15, 1);
    grayscale_dims_ = ImageDimensions(20, 15, 1);

    // Large test image 100x80x3
    large_data_ = CreateImageData(100, 80, 3);
    large_dims_ = ImageDimensions(100, 80, 3);
  }

  std::vector<uint8_t> CreateImageData(int width, int height, int channels) {
    std::vector<uint8_t> data(width * height * channels);

    // Fill with predictable pattern: value = (y * width + x + channel) % 256
    for (int y = 0; y < height; ++y) {
      for (int x = 0; x < width; ++x) {
        int pixel_offset = (y * width + x) * channels;
        for (int c = 0; c < channels; ++c) {
          data[pixel_offset + c] =
              static_cast<uint8_t>((y * width + x + c) % 256);
        }
      }
    }
    return data;
  }

 protected:
  std::vector<uint8_t> rgb_data_;
  ImageDimensions rgb_dims_;
  std::vector<uint8_t> rgba_data_;
  ImageDimensions rgba_dims_;
  std::vector<uint8_t> grayscale_data_;
  ImageDimensions grayscale_dims_;
  std::vector<uint8_t> large_data_;
  ImageDimensions large_dims_;
};

// Test basic construction
TEST_F(MemorySourceTest, BasicConstruction) {
  MemorySource source(std::vector<uint8_t>(rgb_data_), rgb_dims_);

  auto dims = source.GetDimensions();
  EXPECT_EQ(dims.width, 10);
  EXPECT_EQ(dims.height, 8);
  EXPECT_EQ(dims.channels, 3);
}

// Test construction with custom tile size
TEST_F(MemorySourceTest, ConstructionWithCustomTileSize) {
  TileSize custom_tile_size(64, 64);
  MemorySource source(std::vector<uint8_t>(rgb_data_), rgb_dims_,
                      custom_tile_size);

  auto tile_size = source.GetIdealTileSize();
  EXPECT_EQ(tile_size.width, 64);
  EXPECT_EQ(tile_size.height, 64);
}

// Test dimensions with different channel counts
TEST_F(MemorySourceTest, DifferentChannelCounts) {
  MemorySource rgb_source(std::vector<uint8_t>(rgb_data_), rgb_dims_);
  MemorySource rgba_source(std::vector<uint8_t>(rgba_data_), rgba_dims_);
  MemorySource grayscale_source(std::vector<uint8_t>(grayscale_data_),
                                grayscale_dims_);

  auto rgb_dims = rgb_source.GetDimensions();
  auto rgba_dims = rgba_source.GetDimensions();
  auto grayscale_dims = grayscale_source.GetDimensions();

  EXPECT_EQ(rgb_dims.channels, 3);
  EXPECT_EQ(rgba_dims.channels, 4);
  EXPECT_EQ(grayscale_dims.channels, 1);
}

// Test ideal tile size
TEST_F(MemorySourceTest, IdealTileSize) {
  MemorySource source(std::vector<uint8_t>(rgb_data_), rgb_dims_);

  auto tile_size = source.GetIdealTileSize();
  EXPECT_EQ(tile_size.width, 256);  // Default tile size
  EXPECT_EQ(tile_size.height, 256);
}

// Test getting a full tile
TEST_F(MemorySourceTest, GetFullTile) {
  MemorySource source(std::vector<uint8_t>(rgb_data_), rgb_dims_);

  auto tile = source.GetTile(0, 0, 10, 8);
  EXPECT_EQ(tile.x, 0);
  EXPECT_EQ(tile.y, 0);
  EXPECT_EQ(tile.width, 10);
  EXPECT_EQ(tile.height, 8);
  EXPECT_EQ(tile.channels, 3);
  EXPECT_EQ(tile.data.size(), 10 * 8 * 3);
}

// Test getting a partial tile
TEST_F(MemorySourceTest, GetPartialTile) {
  MemorySource source(std::vector<uint8_t>(rgb_data_), rgb_dims_);

  auto tile = source.GetTile(2, 3, 4, 3);
  EXPECT_EQ(tile.x, 2);
  EXPECT_EQ(tile.y, 3);
  EXPECT_EQ(tile.width, 4);
  EXPECT_EQ(tile.height, 3);
  EXPECT_EQ(tile.channels, 3);
  EXPECT_EQ(tile.data.size(), 4 * 3 * 3);
}

// Test tile data correctness
TEST_F(MemorySourceTest, TileDataCorrectness) {
  MemorySource source(std::vector<uint8_t>(rgb_data_), rgb_dims_);

  // Get a 2x2 tile from position (1, 1)
  auto tile = source.GetTile(1, 1, 2, 2);

  // Verify the data matches expected pattern
  for (int y = 0; y < 2; ++y) {
    for (int x = 0; x < 2; ++x) {
      int source_x = 1 + x;
      int source_y = 1 + y;
      int tile_offset = (y * 2 + x) * 3;

      for (int c = 0; c < 3; ++c) {
        uint8_t expected =
            static_cast<uint8_t>((source_y * 10 + source_x + c) % 256);
        EXPECT_EQ(tile.data[tile_offset + c], expected);
      }
    }
  }
}

// Test getting a tile beyond image bounds
TEST_F(MemorySourceTest, GetTileBeyondBounds) {
  MemorySource source(std::vector<uint8_t>(rgb_data_), rgb_dims_);

  auto tile = source.GetTile(8, 6, 5, 5);
  EXPECT_EQ(tile.x, 8);
  EXPECT_EQ(tile.y, 6);
  EXPECT_EQ(tile.width, 2);   // Clamped to image bounds: 10 - 8 = 2
  EXPECT_EQ(tile.height, 2);  // Clamped to image bounds: 8 - 6 = 2
  EXPECT_EQ(tile.channels, 3);
}

// Test getting a tile completely outside image bounds
TEST_F(MemorySourceTest, GetTileOutsideBounds) {
  MemorySource source(std::vector<uint8_t>(rgb_data_), rgb_dims_);

  auto tile = source.GetTile(15, 10, 5, 5);
  EXPECT_EQ(tile.x, 15);
  EXPECT_EQ(tile.y, 10);
  EXPECT_EQ(tile.width, 0);
  EXPECT_EQ(tile.height, 0);
  EXPECT_EQ(tile.channels, 3);
  EXPECT_TRUE(tile.data.empty());
}

// Test getting a tile with negative coordinates (should clamp to 0,0)
TEST_F(MemorySourceTest, GetTileNegativeCoordinates) {
  MemorySource source(std::vector<uint8_t>(rgb_data_), rgb_dims_);

  // When given negative coordinates, MemorySource clamps to (0,0)
  auto tile = source.GetTile(-5, -3, 10, 8);
  EXPECT_EQ(tile.x, -5);      // Original requested x
  EXPECT_EQ(tile.y, -3);      // Original requested y
  EXPECT_EQ(tile.width, 10);  // Full width since clamped coordinates allow it
  EXPECT_EQ(tile.height, 8);  // Full height since clamped coordinates allow it
  EXPECT_EQ(tile.channels, 3);

  // Should contain data from the top-left corner (0,0) of the source
  for (int y = 0; y < 8; ++y) {
    for (int x = 0; x < 10; ++x) {
      int tile_offset = (y * 10 + x) * 3;
      for (int c = 0; c < 3; ++c) {
        uint8_t expected = static_cast<uint8_t>((y * 10 + x + c) % 256);
        EXPECT_EQ(tile.data[tile_offset + c], expected);
      }
    }
  }
}

// Test move constructor
TEST_F(MemorySourceTest, MoveConstructor) {
  MemorySource source1(std::vector<uint8_t>(rgb_data_), rgb_dims_);
  auto dims1 = source1.GetDimensions();
  auto memory_usage1 = source1.GetMemoryUsage();

  MemorySource source2(std::move(source1));
  auto dims2 = source2.GetDimensions();
  auto memory_usage2 = source2.GetMemoryUsage();

  EXPECT_EQ(dims2.width, dims1.width);
  EXPECT_EQ(dims2.height, dims1.height);
  EXPECT_EQ(dims2.channels, dims1.channels);
  EXPECT_EQ(memory_usage2, memory_usage1);

  // Verify source2 is functional
  auto tile = source2.GetTile(0, 0, 2, 2);
  EXPECT_EQ(tile.width, 2);
  EXPECT_EQ(tile.height, 2);
}

// Test move assignment
TEST_F(MemorySourceTest, MoveAssignment) {
  MemorySource source1(std::vector<uint8_t>(rgb_data_), rgb_dims_);
  MemorySource source2(std::vector<uint8_t>(rgba_data_), rgba_dims_);

  auto dims1 = source1.GetDimensions();
  auto memory_usage1 = source1.GetMemoryUsage();

  source2 = std::move(source1);
  auto dims2 = source2.GetDimensions();
  auto memory_usage2 = source2.GetMemoryUsage();

  EXPECT_EQ(dims2.width, dims1.width);
  EXPECT_EQ(dims2.height, dims1.height);
  EXPECT_EQ(dims2.channels, dims1.channels);
  EXPECT_EQ(memory_usage2, memory_usage1);
}

// Test memory usage calculation
TEST_F(MemorySourceTest, MemoryUsage) {
  MemorySource rgb_source(std::vector<uint8_t>(rgb_data_), rgb_dims_);
  MemorySource rgba_source(std::vector<uint8_t>(rgba_data_), rgba_dims_);
  MemorySource grayscale_source(std::vector<uint8_t>(grayscale_data_),
                                grayscale_dims_);

  EXPECT_EQ(rgb_source.GetMemoryUsage(), 10 * 8 * 3);         // 240 bytes
  EXPECT_EQ(rgba_source.GetMemoryUsage(), 6 * 4 * 4);         // 96 bytes
  EXPECT_EQ(grayscale_source.GetMemoryUsage(), 20 * 15 * 1);  // 300 bytes
}

// Test large image handling
TEST_F(MemorySourceTest, LargeImage) {
  MemorySource source(std::vector<uint8_t>(large_data_), large_dims_);

  auto dims = source.GetDimensions();
  EXPECT_EQ(dims.width, 100);
  EXPECT_EQ(dims.height, 80);
  EXPECT_EQ(dims.channels, 3);

  // Test getting various tiles from large image
  auto corner_tile = source.GetTile(0, 0, 32, 32);
  EXPECT_EQ(corner_tile.width, 32);
  EXPECT_EQ(corner_tile.height, 32);

  auto center_tile = source.GetTile(40, 30, 20, 20);
  EXPECT_EQ(center_tile.width, 20);
  EXPECT_EQ(center_tile.height, 20);

  auto edge_tile = source.GetTile(90, 70, 20, 20);
  EXPECT_EQ(edge_tile.width, 10);   // Clamped
  EXPECT_EQ(edge_tile.height, 10);  // Clamped

  EXPECT_EQ(source.GetMemoryUsage(), 100 * 80 * 3);
}

// Test with empty image
TEST_F(MemorySourceTest, EmptyImage) {
  ImageDimensions empty_dims(0, 0, 3);
  std::vector<uint8_t> empty_data;
  MemorySource source(std::move(empty_data), empty_dims);

  auto dims = source.GetDimensions();
  EXPECT_EQ(dims.width, 0);
  EXPECT_EQ(dims.height, 0);
  EXPECT_EQ(dims.channels, 3);

  auto tile = source.GetTile(0, 0, 10, 10);
  EXPECT_EQ(tile.width, 0);
  EXPECT_EQ(tile.height, 0);
  EXPECT_TRUE(tile.data.empty());

  EXPECT_EQ(source.GetMemoryUsage(), 0);
}

// Test single pixel image
TEST_F(MemorySourceTest, SinglePixelImage) {
  ImageDimensions single_dims(1, 1, 3);
  std::vector<uint8_t> single_data = {100, 150, 200};
  MemorySource source(std::move(single_data), single_dims);

  auto dims = source.GetDimensions();
  EXPECT_EQ(dims.width, 1);
  EXPECT_EQ(dims.height, 1);
  EXPECT_EQ(dims.channels, 3);

  auto tile = source.GetTile(0, 0, 1, 1);
  EXPECT_EQ(tile.width, 1);
  EXPECT_EQ(tile.height, 1);
  EXPECT_EQ(tile.data.size(), 3);
  EXPECT_EQ(tile.data[0], 100);
  EXPECT_EQ(tile.data[1], 150);
  EXPECT_EQ(tile.data[2], 200);
}

// Test grayscale image handling
TEST_F(MemorySourceTest, GrayscaleImage) {
  MemorySource source(std::vector<uint8_t>(grayscale_data_), grayscale_dims_);

  auto dims = source.GetDimensions();
  EXPECT_EQ(dims.channels, 1);

  auto tile = source.GetTile(5, 5, 3, 3);
  EXPECT_EQ(tile.channels, 1);
  EXPECT_EQ(tile.data.size(), 3 * 3 * 1);

  // Verify grayscale data pattern
  for (int y = 0; y < 3; ++y) {
    for (int x = 0; x < 3; ++x) {
      int source_x = 5 + x;
      int source_y = 5 + y;
      int tile_offset = y * 3 + x;
      uint8_t expected = static_cast<uint8_t>((source_y * 20 + source_x) % 256);
      EXPECT_EQ(tile.data[tile_offset], expected);
    }
  }
}

// Test RGBA image handling
TEST_F(MemorySourceTest, RgbaImage) {
  MemorySource source(std::vector<uint8_t>(rgba_data_), rgba_dims_);

  auto dims = source.GetDimensions();
  EXPECT_EQ(dims.channels, 4);

  auto tile = source.GetTile(1, 1, 2, 2);
  EXPECT_EQ(tile.channels, 4);
  EXPECT_EQ(tile.data.size(), 2 * 2 * 4);

  // Verify RGBA data pattern
  for (int y = 0; y < 2; ++y) {
    for (int x = 0; x < 2; ++x) {
      int source_x = 1 + x;
      int source_y = 1 + y;
      int tile_offset = (y * 2 + x) * 4;

      for (int c = 0; c < 4; ++c) {
        uint8_t expected =
            static_cast<uint8_t>((source_y * 6 + source_x + c) % 256);
        EXPECT_EQ(tile.data[tile_offset + c], expected);
      }
    }
  }
}

// Test multiple tile accesses
TEST_F(MemorySourceTest, MultipleTileAccess) {
  MemorySource source(std::vector<uint8_t>(rgb_data_), rgb_dims_);

  // Access the same tile multiple times
  auto tile1 = source.GetTile(2, 2, 3, 3);
  auto tile2 = source.GetTile(2, 2, 3, 3);

  EXPECT_EQ(tile1.width, tile2.width);
  EXPECT_EQ(tile1.height, tile2.height);
  EXPECT_EQ(tile1.data, tile2.data);

  // Access different tiles
  auto tile3 = source.GetTile(0, 0, 2, 2);
  auto tile4 = source.GetTile(5, 5, 2, 2);

  EXPECT_EQ(tile3.width, 2);
  EXPECT_EQ(tile4.width, 2);
  EXPECT_NE(tile3.data, tile4.data);  // Should have different data
}

}  // namespace fim
