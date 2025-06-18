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
#include "ahcore/data/readers/disk_tile_reader.h"

#include <gtest/gtest.h>
#include <spdlog/spdlog.h>

#include <iostream>
#include <memory>
#include <random>
#include <string>

#include "ahcore/data/readers/base_reader.h"
#include "ahcore/data/writers/disk_tile_writer.h"

using namespace aifo::data::readers;
using namespace aifo::data::writers;

class GlobalEnvironment : public ::testing::Environment {
 public:
  ~GlobalEnvironment() override = default;

  void SetUp() override {
    // Initialize VIPS before any tests run
    if (VIPS_INIT("")) {
      throw std::runtime_error("VIPS initialization failed");
    }

    // Setup logging
    spdlog::set_level(spdlog::level::debug);  // Set global log level to debug
    spdlog::set_pattern(
        "[%H:%M:%S.%e] [%^%l%$] [%s:%#] %v");  // Set a detailed log pattern
  }

  void TearDown() override {
    // Clean up VIPS after all tests complete
    vips_shutdown();
  }
};

class ReaderTest : public ::testing::Test {
 protected:
  std::unique_ptr<DiskTileReader> reader;
  fs::path test_dir;
  vips::VImage test_pattern;

  static fs::path CreateUniqueTestDir() {
    // TODO(jonasteuwen): Use TemporaryDirectory
    // Create a unique test directory name using timestamp and random number
    auto now = std::chrono::system_clock::now().time_since_epoch();
    auto ms =
        std::chrono::duration_cast<std::chrono::milliseconds>(now).count();

    std::random_device rd;
    std::mt19937 gen(rd());
    std::uniform_int_distribution<> dis(0, 999999);
    int random_num = dis(gen);

    fs::path temp_root = fs::temp_directory_path();
    fs::path test_dir =
        temp_root / fmt::format("reader_test_{}_{}", ms, random_num);
    fs::create_directories(test_dir);
    return test_dir;
  }

  void SetUp() override {
    // Create a test pattern using vips_eye
    VipsImage* image;
    if (vips_eye(&image, 100, 100, "uchar", TRUE, NULL)) {
      throw std::runtime_error("Could not create test image");
    }
    test_pattern = vips::VImage(image);

    // Create unique test directory
    test_dir = CreateUniqueTestDir();

    // Set up metadata
    auto metadata = aifo::data::Metadata::Create();
    metadata->SetGeometry(dlup::SlideGeometry{{100, 100}, {0, 0}, {100, 100}});
    metadata->SetTileSize(100, 100);
    metadata->SetTileOverlap(0, 0);
    metadata->SetMpp(1.0);
    metadata->Set(MetadataKeys::PixelFormat,
                  static_cast<int>(test_pattern.format()));
    metadata->Set(MetadataKeys::Interpretation,
                  static_cast<int>(test_pattern.interpretation()));
    metadata->Set(MetadataKeys::NumChannels, test_pattern.bands());
    metadata->SetGridOrder(aifocore::tiling::GridOrder::kC);

    // Create and set up the grid
    auto grid = std::make_shared<aifocore::tiling::Grid<int>>(
        aifocore::tiling::Grid<int>::FromTiling(
            {0, 0},      // offset
            {100, 100},  // size
            {100, 100},  // tile size
            {0, 0},      // overlap
            aifocore::tiling::TilingMode::kOverflow,
            aifocore::tiling::GridOrder::kC));

    // Write the test pattern to disk using a DiskTileWriter
    auto writer = std::make_unique<DiskTileWriter>(test_dir, metadata);
    writer->SetGrid(grid);
    writer->Open();

    // Use Consume to write the test pattern
    writer->Consume([this](int index) -> dlup::DatasetSample {
      dlup::DatasetSample sample;
      sample.coordinates = {0, 0};  // Single tile at origin
      sample.tile = std::make_shared<vips::VImage>(test_pattern);
      return sample;
    });

    reader = std::make_unique<DiskTileReader>(test_dir.string(),
                                              StitchingMode::kCrop);

    try {
      reader->Open();
    } catch (const std::exception& e) {
      std::cerr << "Failed to open reader: " << e.what() << std::endl;
      throw;
    }
  }

  void TearDown() override {
    if (reader) {
      reader->Close();
    }
    try {
      if (fs::exists(test_dir)) {
        fs::remove_all(test_dir);
      }
    } catch (const std::exception& e) {
      std::cerr << "Failed to clean up test directory: " << e.what()
                << std::endl;
    }
  }
};

TEST_F(ReaderTest, ReadRegionCompletelyOutsideWithAllowOutOfBounds) {
  // Region completely outside the image bounds
  absl::StatusOr<vips::VImage> region =
      reader->ReadRegion({200, 200}, 0, {50, 50});
  ASSERT_TRUE(region.ok());
  EXPECT_EQ(region.value().width(), 50);
  EXPECT_EQ(region.value().height(), 50);

  // Check if it's completely black
  auto min_val = region.value().min();
  auto max_val = region.value().max();
  EXPECT_DOUBLE_EQ(min_val, 0.0);
  EXPECT_DOUBLE_EQ(max_val, 0.0);

  // Read negative coordinates
  absl::StatusOr<vips::VImage> region_neg =
      reader->ReadRegion({-200, -200}, 0, {50, 50});
  ASSERT_TRUE(region_neg.ok());
  EXPECT_EQ(region_neg.value().width(), 50);
  EXPECT_EQ(region_neg.value().height(), 50);

  // Check if it's completely black
  auto min_val_neg = region_neg.value().min();
  auto max_val_neg = region_neg.value().max();
  EXPECT_DOUBLE_EQ(min_val_neg, 0.0);
  EXPECT_DOUBLE_EQ(max_val_neg, 0.0);
}

TEST_F(ReaderTest, ReadRegionCompletelyInsideOfBounds) {
  // Region completely inside the image bounds
  absl::StatusOr<vips::VImage> region =
      reader->ReadRegion({0, 0}, 0, {100, 100});
  ASSERT_TRUE(region.ok());
  EXPECT_EQ(region.value().width(), 100);
  EXPECT_EQ(region.value().height(), 100);

  // Compare with original pattern
  vips::VImage diff = region.value() - test_pattern;
  EXPECT_DOUBLE_EQ(diff.max(), 0.0);
  EXPECT_DOUBLE_EQ(diff.min(), 0.0);
}

TEST_F(ReaderTest, ReadRegionPartiallyOutsideWithAllowOutOfBounds) {
  // Region partially outside the image bounds
  absl::StatusOr<vips::VImage> region =
      reader->ReadRegion({80, 80}, 0, {40, 40});
  ASSERT_TRUE(region.ok());
  EXPECT_EQ(region.value().width(), 40);
  EXPECT_EQ(region.value().height(), 40);

  // Extract and compare the valid portion
  vips::VImage valid_portion = test_pattern.extract_area(80, 80, 20, 20);
  vips::VImage region_portion = region.value().extract_area(0, 0, 20, 20);
  vips::VImage diff = valid_portion - region_portion;
  EXPECT_DOUBLE_EQ(diff.max(), 0.0);
  EXPECT_DOUBLE_EQ(diff.min(), 0.0);

  // Extract and compare the invalid portion
  vips::VImage region_portion_invalid =
      region.value().extract_area(20, 20, 20, 20);
  EXPECT_DOUBLE_EQ(region_portion_invalid.max(), 0.0);
  EXPECT_DOUBLE_EQ(region_portion_invalid.min(), 0.0);
}

int main(int argc, char** argv) {
  ::testing::InitGoogleTest(&argc, argv);
  // Add the global environment
  ::testing::AddGlobalTestEnvironment(new GlobalEnvironment);
  return RUN_ALL_TESTS();
}
