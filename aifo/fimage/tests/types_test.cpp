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

#include <fim/types.h>
#include <gtest/gtest.h>

namespace fim {

class TileTest : public ::testing::Test {
 protected:
  void SetUp() override {}

  void TearDown() override {}
};

// Test Tile default construction
TEST_F(TileTest, DefaultConstruction) {
  Tile tile;
  EXPECT_EQ(tile.x, 0);
  EXPECT_EQ(tile.y, 0);
  EXPECT_EQ(tile.width, 0);
  EXPECT_EQ(tile.height, 0);
  EXPECT_EQ(tile.channels, 0);
  EXPECT_EQ(tile.expected_width, 0);
  EXPECT_EQ(tile.expected_height, 0);
  EXPECT_FALSE(tile.needs_padding);
  EXPECT_TRUE(tile.data.empty());
}

// Test Tile construction with basic parameters
TEST_F(TileTest, BasicConstruction) {
  Tile tile(10, 20, 100, 200, 3);
  EXPECT_EQ(tile.x, 10);
  EXPECT_EQ(tile.y, 20);
  EXPECT_EQ(tile.width, 100);
  EXPECT_EQ(tile.height, 200);
  EXPECT_EQ(tile.channels, 3);
  EXPECT_EQ(tile.expected_width, 0);
  EXPECT_EQ(tile.expected_height, 0);
  EXPECT_FALSE(tile.needs_padding);
  EXPECT_EQ(tile.data.size(), 100 * 200 * 3);
}

// Test Tile construction with padding parameters
TEST_F(TileTest, PaddingConstruction) {
  Tile tile(10, 20, 80, 90, 4, 100, 100);
  EXPECT_EQ(tile.x, 10);
  EXPECT_EQ(tile.y, 20);
  EXPECT_EQ(tile.width, 80);
  EXPECT_EQ(tile.height, 90);
  EXPECT_EQ(tile.channels, 4);
  EXPECT_EQ(tile.expected_width, 100);
  EXPECT_EQ(tile.expected_height, 100);
  EXPECT_TRUE(tile.needs_padding);
  EXPECT_EQ(tile.data.size(), 80 * 90 * 4);
}

// Test Tile construction with no padding needed
TEST_F(TileTest, NoPaddingConstruction) {
  Tile tile(10, 20, 100, 100, 3, 100, 100);
  EXPECT_EQ(tile.x, 10);
  EXPECT_EQ(tile.y, 20);
  EXPECT_EQ(tile.width, 100);
  EXPECT_EQ(tile.height, 100);
  EXPECT_EQ(tile.channels, 3);
  EXPECT_EQ(tile.expected_width, 100);
  EXPECT_EQ(tile.expected_height, 100);
  EXPECT_FALSE(tile.needs_padding);
  EXPECT_EQ(tile.data.size(), 100 * 100 * 3);
}

// Test GetDataSize method
TEST_F(TileTest, GetDataSize) {
  Tile tile(0, 0, 10, 20, 3);
  EXPECT_EQ(tile.GetDataSize(), 10 * 20 * 3);
}

// Test GetExpectedDataSize method without padding
TEST_F(TileTest, GetExpectedDataSizeWithoutPadding) {
  Tile tile(0, 0, 10, 20, 3);
  EXPECT_EQ(tile.GetExpectedDataSize(), 10 * 20 * 3);
}

// Test GetExpectedDataSize method with padding
TEST_F(TileTest, GetExpectedDataSizeWithPadding) {
  Tile tile(0, 0, 10, 20, 3, 25, 30);
  EXPECT_EQ(tile.GetExpectedDataSize(), 25 * 30 * 3);
}

// Test CreatePaddedData method without padding
TEST_F(TileTest, CreatePaddedDataWithoutPadding) {
  Tile tile(0, 0, 2, 2, 1);
  // Fill with test data
  tile.data = {1, 2, 3, 4};

  auto padded = tile.CreatePaddedData();
  EXPECT_EQ(padded.size(), 4);
  EXPECT_EQ(padded[0], 1);
  EXPECT_EQ(padded[1], 2);
  EXPECT_EQ(padded[2], 3);
  EXPECT_EQ(padded[3], 4);
}

// Test CreatePaddedData method with padding
TEST_F(TileTest, CreatePaddedDataWithPadding) {
  Tile tile(0, 0, 2, 2, 1, 3, 3);
  // Fill with test data
  tile.data = {1, 2, 3, 4};

  auto padded = tile.CreatePaddedData();
  EXPECT_EQ(padded.size(), 9);

  // Check that the original data is in the top-left corner
  EXPECT_EQ(padded[0], 1);  // (0,0)
  EXPECT_EQ(padded[1], 2);  // (0,1)
  EXPECT_EQ(padded[2], 0);  // (0,2) - padding
  EXPECT_EQ(padded[3], 3);  // (1,0)
  EXPECT_EQ(padded[4], 4);  // (1,1)
  EXPECT_EQ(padded[5], 0);  // (1,2) - padding
  EXPECT_EQ(padded[6], 0);  // (2,0) - padding
  EXPECT_EQ(padded[7], 0);  // (2,1) - padding
  EXPECT_EQ(padded[8], 0);  // (2,2) - padding
}

// Test CreatePaddedData method with multi-channel data
TEST_F(TileTest, CreatePaddedDataMultiChannel) {
  Tile tile(0, 0, 2, 2, 2, 3, 3);
  // Fill with test data (2x2 image with 2 channels)
  tile.data = {1, 2, 3, 4, 5, 6, 7, 8};

  auto padded = tile.CreatePaddedData();
  EXPECT_EQ(padded.size(), 18);  // 3x3x2

  // Check that the original data is in the top-left corner
  EXPECT_EQ(padded[0], 1);   // (0,0) channel 0
  EXPECT_EQ(padded[1], 2);   // (0,0) channel 1
  EXPECT_EQ(padded[2], 3);   // (0,1) channel 0
  EXPECT_EQ(padded[3], 4);   // (0,1) channel 1
  EXPECT_EQ(padded[4], 0);   // (0,2) channel 0 - padding
  EXPECT_EQ(padded[5], 0);   // (0,2) channel 1 - padding
  EXPECT_EQ(padded[6], 5);   // (1,0) channel 0
  EXPECT_EQ(padded[7], 6);   // (1,0) channel 1
  EXPECT_EQ(padded[8], 7);   // (1,1) channel 0
  EXPECT_EQ(padded[9], 8);   // (1,1) channel 1
  EXPECT_EQ(padded[10], 0);  // (1,2) channel 0 - padding
  EXPECT_EQ(padded[11], 0);  // (1,2) channel 1 - padding
  // Rest should be padding (zeros)
  for (int i = 12; i < 18; ++i) {
    EXPECT_EQ(padded[i], 0);
  }
}

class ImageDimensionsTest : public ::testing::Test {
 protected:
  void SetUp() override {}

  void TearDown() override {}
};

// Test ImageDimensions default construction
TEST_F(ImageDimensionsTest, DefaultConstruction) {
  ImageDimensions dims;
  EXPECT_EQ(dims.width, 0);
  EXPECT_EQ(dims.height, 0);
  EXPECT_EQ(dims.channels, 0);
}

// Test ImageDimensions construction with parameters
TEST_F(ImageDimensionsTest, ParameterizedConstruction) {
  ImageDimensions dims(800, 600, 3);
  EXPECT_EQ(dims.width, 800);
  EXPECT_EQ(dims.height, 600);
  EXPECT_EQ(dims.channels, 3);
}

// Test ImageDimensions with various channel counts
TEST_F(ImageDimensionsTest, VariousChannelCounts) {
  ImageDimensions grayscale(100, 100, 1);
  ImageDimensions rgb(100, 100, 3);
  ImageDimensions rgba(100, 100, 4);

  EXPECT_EQ(grayscale.channels, 1);
  EXPECT_EQ(rgb.channels, 3);
  EXPECT_EQ(rgba.channels, 4);
}

class TileSizeTest : public ::testing::Test {
 protected:
  void SetUp() override {}

  void TearDown() override {}
};

// Test TileSize default construction
TEST_F(TileSizeTest, DefaultConstruction) {
  TileSize tile_size;
  EXPECT_EQ(tile_size.width, 0);
  EXPECT_EQ(tile_size.height, 0);
}

// Test TileSize construction with parameters
TEST_F(TileSizeTest, ParameterizedConstruction) {
  TileSize tile_size(256, 256);
  EXPECT_EQ(tile_size.width, 256);
  EXPECT_EQ(tile_size.height, 256);
}

// Test TileSize with different dimensions
TEST_F(TileSizeTest, DifferentDimensions) {
  TileSize tile_size(512, 256);
  EXPECT_EQ(tile_size.width, 512);
  EXPECT_EQ(tile_size.height, 256);
}

// Test TileSize with zero dimensions
TEST_F(TileSizeTest, ZeroDimensions) {
  TileSize tile_size(0, 0);
  EXPECT_EQ(tile_size.width, 0);
  EXPECT_EQ(tile_size.height, 0);
}

}  // namespace fim
