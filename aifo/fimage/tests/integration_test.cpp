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

#include <gtest/gtest.h>

#include <filesystem>
#include <string>
#include <utility>
#include <vector>

#include "fim/image.h"
#include "fim/operators/crop.h"
#include "fim/operators/downsample.h"
#include "fim/sinks/png_sink.h"
#include "fim/sinks/tiff_sink.h"
#include "fim/sources/png_source.h"
#include "fim/sources/tiff_source.h"
#include "fim/types.h"
#include "lodepng/lodepng.h"

namespace fim {

class IntegrationTest : public ::testing::Test {
 protected:
  void SetUp() override {
    // Create a test PNG file for testing
    CreateTestPngFile("test_integration.png", 200, 150, 3);
  }

  void TearDown() override {
    // Clean up test files
    std::filesystem::remove("test_integration.png");
    std::filesystem::remove("test_output.png");
    std::filesystem::remove("test_output.tiff");
    std::filesystem::remove("test_cropped.png");
    std::filesystem::remove("test_downsampled.png");
    std::filesystem::remove("test_memory_output.tiff");
    std::filesystem::remove("test_pyramidal.tiff");
    std::filesystem::remove("test_single.tiff");
    std::filesystem::remove("test_large_pyramidal.tiff");
  }

 private:
  void CreateTestPngFile(const std::string& filename, int width, int height,
                         int channels) {
    std::vector<uint8_t> image_data(width * height * channels);

    // Fill with a checkerboard pattern
    for (int y = 0; y < height; ++y) {
      for (int x = 0; x < width; ++x) {
        int pixel_offset = (y * width + x) * channels;
        bool is_black = ((x / 16) + (y / 16)) % 2 == 0;
        uint8_t value = is_black ? 0 : 255;

        for (int c = 0; c < channels; ++c) {
          image_data[pixel_offset + c] = value;
        }
      }
    }

    // Write PNG file using lodepng
    unsigned error = lodepng_encode24_file(filename.c_str(), image_data.data(),
                                           width, height);
    if (error != 0) {
      throw std::runtime_error("Failed to create test PNG file: " +
                               std::string(lodepng_error_text(error)));
    }
  }
};

// Test PNG source to PNG sink pipeline
TEST_F(IntegrationTest, PngToPngPipeline) {
  auto image = fim::CreatePngImage("test_integration.png");
  image.Render(fim::PngSink("test_output.png"));

  EXPECT_TRUE(std::filesystem::exists("test_output.png"));
  EXPECT_GT(std::filesystem::file_size("test_output.png"), 0);
}

// Test PNG source to TIFF sink pipeline
TEST_F(IntegrationTest, PngToTiffPipeline) {
  auto image = fim::CreatePngImage("test_integration.png");
  image.Render(fim::TiffSink("test_output.tiff"));

  EXPECT_TRUE(std::filesystem::exists("test_output.tiff"));
  EXPECT_GT(std::filesystem::file_size("test_output.tiff"), 0);
}

// Test TIFF source (if available) to PNG sink pipeline
TEST_F(IntegrationTest, TiffToPngPipeline) {
  if (!std::filesystem::exists("checkerboard.svs")) {
    GTEST_SKIP() << "Test TIFF file not found: checkerboard.svs";
  }

  auto image = fim::CreateTiffImage("checkerboard.svs");

  // Crop a small region to make processing faster
  image.Crop(100, 100, 500, 500).Render(fim::PngSink("test_output.png"));

  EXPECT_TRUE(std::filesystem::exists("test_output.png"));
  EXPECT_GT(std::filesystem::file_size("test_output.png"), 0);
}

// Test crop operator integration
TEST_F(IntegrationTest, CropOperator) {
  auto image = fim::CreatePngImage("test_integration.png");

  // Crop the center region
  image.Crop(50, 37, 100, 76).Render(fim::PngSink("test_cropped.png"));

  EXPECT_TRUE(std::filesystem::exists("test_cropped.png"));
  EXPECT_GT(std::filesystem::file_size("test_cropped.png"), 0);

  // Verify the output dimensions by loading the result
  auto cropped_image = fim::CreatePngImage("test_cropped.png");
  auto dims = cropped_image.GetSource().GetDimensions();
  EXPECT_EQ(dims.width, 100);
  EXPECT_EQ(dims.height, 76);
}

// Test downsample operator integration
TEST_F(IntegrationTest, DownsampleOperator) {
  auto image = fim::CreatePngImage("test_integration.png");

  // Downsample by factor of 2
  image.Downsample(2).Render(fim::PngSink("test_downsampled.png"));

  EXPECT_TRUE(std::filesystem::exists("test_downsampled.png"));
  EXPECT_GT(std::filesystem::file_size("test_downsampled.png"), 0);

  // Verify the output dimensions by loading the result
  auto downsampled_image = fim::CreatePngImage("test_downsampled.png");
  auto dims = downsampled_image.GetSource().GetDimensions();
  EXPECT_EQ(dims.width, 100);  // 200 / 2
  EXPECT_EQ(dims.height, 75);  // 150 / 2
}

// Test chained operations
TEST_F(IntegrationTest, ChainedOperations) {
  auto image = fim::CreatePngImage("test_integration.png");

  // Crop then downsample
  image.Crop(20, 20, 160, 110)
      .Downsample(2)
      .Render(fim::PngSink("test_output.png"));

  EXPECT_TRUE(std::filesystem::exists("test_output.png"));
  EXPECT_GT(std::filesystem::file_size("test_output.png"), 0);

  // Verify the final dimensions
  auto result_image = fim::CreatePngImage("test_output.png");
  auto dims = result_image.GetSource().GetDimensions();
  EXPECT_EQ(dims.width, 80);   // 160 / 2
  EXPECT_EQ(dims.height, 55);  // 110 / 2
}

// Test error handling
TEST_F(IntegrationTest, ErrorHandling) {
  // Test with non-existent file
  EXPECT_THROW(fim::CreatePngImage("non_existent.png"), std::runtime_error);
  EXPECT_THROW(fim::CreateTiffImage("non_existent.tiff"), std::runtime_error);

  // Test with invalid downsample factor
  auto image = fim::CreatePngImage("test_integration.png");
  EXPECT_THROW(image.Downsample(0), std::invalid_argument);
  EXPECT_THROW(image.Downsample(-1), std::invalid_argument);
}

// Test image dimensions and properties
TEST_F(IntegrationTest, ImageProperties) {
  auto image = fim::CreatePngImage("test_integration.png");

  auto dims = image.GetSource().GetDimensions();
  EXPECT_EQ(dims.width, 200);
  EXPECT_EQ(dims.height, 150);
  EXPECT_EQ(dims.channels, 4);  // PNG source always loads as RGBA

  auto tile_size = image.GetSource().GetIdealTileSize();
  EXPECT_EQ(tile_size.width, 200);
  EXPECT_EQ(tile_size.height, 150);
}

// Test tile-based processing
TEST_F(IntegrationTest, TileProcessing) {
  auto image = fim::CreatePngImage("test_integration.png");

  // Get a tile from the source
  auto tile = image.GetSource().GetTile(50, 50, 100, 75);
  EXPECT_EQ(tile.x, 50);
  EXPECT_EQ(tile.y, 50);
  EXPECT_EQ(tile.width, 100);
  EXPECT_EQ(tile.height, 75);
  EXPECT_EQ(tile.channels, 4);
  EXPECT_EQ(tile.data.size(), 100 * 75 * 4);
}

// Test large TIFF processing (if available)
TEST_F(IntegrationTest, LargeTiffProcessing) {
  if (!std::filesystem::exists("checkerboard.svs")) {
    GTEST_SKIP() << "Test TIFF file not found: checkerboard.svs";
  }

  auto image = fim::CreateTiffImage("checkerboard.svs");
  auto dims = image.GetSource().GetDimensions();

  EXPECT_GT(dims.width, 0);
  EXPECT_GT(dims.height, 0);
  EXPECT_GT(dims.channels, 0);

  // Process a small region to ensure it works
  image.Crop(1000, 1000, 512, 512)
      .Downsample(2)
      .Render(fim::TiffSink("test_output.tiff", fim::TileSize(256, 256)));

  EXPECT_TRUE(std::filesystem::exists("test_output.tiff"));
  EXPECT_GT(std::filesystem::file_size("test_output.tiff"), 0);
}

// Test MemorySource integration with operators
TEST_F(IntegrationTest, MemorySourceIntegration) {
  // Create test data
  std::vector<uint8_t> image_data(100 * 80 * 3);
  for (int y = 0; y < 80; ++y) {
    for (int x = 0; x < 100; ++x) {
      int pixel_offset = (y * 100 + x) * 3;
      uint8_t value = static_cast<uint8_t>((x + y) % 256);
      for (int c = 0; c < 3; ++c) {
        image_data[pixel_offset + c] = value;
      }
    }
  }

  // Create MemorySource
  fim::MemorySource source(std::move(image_data),
                           fim::ImageDimensions(100, 80, 3));

  // Create Image from MemorySource (we need a way to do this)
  // For now, test the source directly with operators
  auto crop_op = fim::Crop(source, 10, 10, 50, 40);
  auto downsample_op = fim::Downsample(crop_op, 2);

  // Test dimensions
  auto crop_dims = crop_op.GetDimensions();
  auto downsample_dims = downsample_op.GetDimensions();

  EXPECT_EQ(crop_dims.width, 50);
  EXPECT_EQ(crop_dims.height, 40);
  EXPECT_EQ(downsample_dims.width, 25);
  EXPECT_EQ(downsample_dims.height, 20);

  // Test rendering to TIFF
  fim::TiffSink sink("test_memory_output.tiff");
  sink.Render(downsample_op);

  EXPECT_TRUE(std::filesystem::exists("test_memory_output.tiff"));
  EXPECT_GT(std::filesystem::file_size("test_memory_output.tiff"), 0);
}

// Test pyramidal TIFF creation integration
TEST_F(IntegrationTest, PyramidalTiffIntegration) {
  auto image = fim::CreatePngImage("test_integration.png");

  // Create pyramidal TIFF with specific parameters
  image.Crop(25, 25, 150, 100)
      .Render(fim::TiffSink("test_pyramidal.tiff", fim::TileSize(64, 64),
                            true,  // pyramidal
                            10,    // 10MB memory threshold
                            2));   // downsample factor 2

  EXPECT_TRUE(std::filesystem::exists("test_pyramidal.tiff"));
  EXPECT_GT(std::filesystem::file_size("test_pyramidal.tiff"), 0);

  // Create single-page TIFF for comparison
  image.Crop(25, 25, 150, 100)
      .Render(fim::TiffSink("test_single.tiff", fim::TileSize(64, 64),
                            false));  // single-page

  EXPECT_TRUE(std::filesystem::exists("test_single.tiff"));
  EXPECT_GT(std::filesystem::file_size("test_single.tiff"), 0);

  // Pyramidal should be larger than single-page
  auto pyramidal_size = std::filesystem::file_size("test_pyramidal.tiff");
  auto single_size = std::filesystem::file_size("test_single.tiff");
  EXPECT_GT(pyramidal_size, single_size);
}

// Test pyramidal TIFF with large image (if available)
TEST_F(IntegrationTest, PyramidalLargeTiffProcessing) {
  if (!std::filesystem::exists("checkerboard.svs")) {
    GTEST_SKIP() << "Test TIFF file not found: checkerboard.svs";
  }

  auto image = fim::CreateTiffImage("checkerboard.svs");

  // Create pyramidal TIFF from large source
  image.Crop(500, 500, 1024, 1024)
      .Render(fim::TiffSink("test_large_pyramidal.tiff",
                            fim::TileSize(256, 256),
                            true,  // pyramidal
                            20,    // 20MB memory threshold
                            2));   // downsample factor 2

  EXPECT_TRUE(std::filesystem::exists("test_large_pyramidal.tiff"));
  EXPECT_GT(std::filesystem::file_size("test_large_pyramidal.tiff"), 0);
}

// Test MemorySource memory efficiency
TEST_F(IntegrationTest, MemorySourceEfficiency) {
  // Create different sized images to test memory usage
  std::vector<uint8_t> small_data(50 * 50 * 3);
  std::vector<uint8_t> large_data(200 * 200 * 3);

  // Fill with test data
  std::fill(small_data.begin(), small_data.end(), 128);
  std::fill(large_data.begin(), large_data.end(), 200);

  fim::MemorySource small_source(std::move(small_data),
                                 fim::ImageDimensions(50, 50, 3));
  fim::MemorySource large_source(std::move(large_data),
                                 fim::ImageDimensions(200, 200, 3));

  // Check memory usage
  EXPECT_EQ(small_source.GetMemoryUsage(), 50 * 50 * 3);
  EXPECT_EQ(large_source.GetMemoryUsage(), 200 * 200 * 3);

  // Test tile access doesn't increase memory usage
  auto tile1 = small_source.GetTile(0, 0, 25, 25);
  auto tile2 = large_source.GetTile(0, 0, 100, 100);

  EXPECT_EQ(small_source.GetMemoryUsage(), 50 * 50 * 3);
  EXPECT_EQ(large_source.GetMemoryUsage(), 200 * 200 * 3);

  EXPECT_EQ(tile1.width, 25);
  EXPECT_EQ(tile2.width, 100);
}

}  // namespace fim
