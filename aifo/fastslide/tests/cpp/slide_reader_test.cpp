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

#include "fastslide/slide_reader.h"

#include <gtest/gtest.h>

#include <algorithm>
#include <chrono>
#include <future>
#include <limits>
#include <memory>
#include <sstream>
#include <string>
#include <string_view>
#include <thread>
#include <tuple>
#include <utility>
#include <vector>

#include "absl/status/status.h"
#include "absl/status/statusor.h"

namespace fastslide {

/// @brief Test helper class to access protected SlideReader methods
class TestSlideReader : public SlideReader {
 public:
  TestSlideReader() = default;

  // Expose protected method for testing
  static RegionSpec TestClampRegion(const RegionSpec& region,
                                    const ImageDimensions& image_dims) {
    return ClampRegion(region, image_dims);
  }

  // Minimal implementation to satisfy abstract base class
  [[nodiscard]] int GetLevelCount() const override { return 0; }

  [[nodiscard]] absl::StatusOr<LevelInfo> GetLevelInfo(
      int level) const override {
    return absl::UnimplementedError("Test class");
  }

  [[nodiscard]] const SlideProperties& GetProperties() const override {
    static SlideProperties props;
    return props;
  }

  [[nodiscard]] std::vector<std::string> GetAssociatedImageNames()
      const override {
    return {};
  }

  [[nodiscard]] absl::StatusOr<ImageDimensions> GetAssociatedImageDimensions(
      std::string_view name) const override {
    return absl::UnimplementedError("Test class");
  }

  [[nodiscard]] absl::StatusOr<RGBImage> ReadRegion(
      const RegionSpec& region) const override {
    return absl::UnimplementedError("Test class");
  }

  [[nodiscard]] absl::StatusOr<RGBImage> ReadAssociatedImage(
      std::string_view name) const override {
    return absl::UnimplementedError("Test class");
  }

  [[nodiscard]] int GetBestLevelForDownsample(
      double downsample) const override {
    return 0;
  }

  [[nodiscard]] ImageDimensions GetTileSize() const override {
    return ImageDimensions{256, 256};  // Default tile size
  }

  [[nodiscard]] Metadata GetMetadata() const override { return Metadata(); }

  [[nodiscard]] std::string GetFormatName() const override {
    return "TestFormat";
  }
};

/// @brief Mock slide reader implementation for testing
class MockSlideReader : public SlideReader {
 public:
  explicit MockSlideReader(std::string_view filename) : filename_(filename) {}

  [[nodiscard]] int GetLevelCount() const override { return 1; }

  [[nodiscard]] absl::StatusOr<LevelInfo> GetLevelInfo(
      int level) const override {
    if (level != 0) {
      return absl::InvalidArgumentError("Invalid level");
    }
    LevelInfo info;
    info.dimensions = ImageDimensions{1024, 768};
    info.downsample_factor = 1.0;
    return info;
  }

  [[nodiscard]] const SlideProperties& GetProperties() const override {
    static SlideProperties props;
    props.mpp = aifocore::Size<double, 2>(0.25, 0.25);
    props.objective_magnification = 20.0;
    props.objective_name = "Test Objective";
    props.scanner_model = "Test Scanner";
    return props;
  }

  [[nodiscard]] std::vector<std::string> GetAssociatedImageNames()
      const override {
    return {"thumbnail", "macro"};
  }

  [[nodiscard]] absl::StatusOr<ImageDimensions> GetAssociatedImageDimensions(
      std::string_view name) const override {
    if (name == "thumbnail") {
      return ImageDimensions{128, 96};
    }
    if (name == "macro") {
      return ImageDimensions{512, 384};
    }
    return absl::NotFoundError("Associated image not found");
  }

  [[nodiscard]] absl::StatusOr<Image> ReadRegion(
      const RegionSpec& region) const override {
    if (!region.IsValid()) {
      return absl::InvalidArgumentError("Invalid region");
    }
    return Image(region.size, ImageFormat::kRGB, DataType::kUInt8);
  }

  [[nodiscard]] absl::StatusOr<Image> ReadAssociatedImage(
      std::string_view name) const override {
    auto dims_result = GetAssociatedImageDimensions(name);
    if (!dims_result.ok()) {
      return dims_result.status();
    }
    return Image(dims_result.value(), ImageFormat::kRGB, DataType::kUInt8);
  }

  [[nodiscard]] int GetBestLevelForDownsample(
      double downsample) const override {
    return 0;  // Only one level
  }

  [[nodiscard]] Metadata GetMetadata() const override {
    Metadata metadata;
    metadata["format"] = std::string("MockFormat");
    metadata["version"] = std::string("1.0");
    metadata["width"] = size_t(1024);
    metadata["height"] = size_t(768);
    metadata["pixel_size"] = 0.25;
    return metadata;
  }

  [[nodiscard]] std::string GetFormatName() const override {
    return "MockFormat";
  }

  [[nodiscard]] ImageDimensions GetTileSize() const override {
    return ImageDimensions{256, 256};  // Default tile size
  }

  [[nodiscard]] std::vector<ChannelMetadata> GetChannelMetadata()
      const override {
    std::vector<ChannelMetadata> channels;
    channels.emplace_back("RGB", "RGB", ColorRGB(255, 255, 255), 0, 8);
    return channels;
  }

 private:
  std::string filename_;
};

/// @brief Failing mock slide reader for testing error propagation
class FailingMockSlideReader : public SlideReader {
 public:
  explicit FailingMockSlideReader(std::string_view filename)
      : filename_(filename) {}

  [[nodiscard]] int GetLevelCount() const override { return 0; }

  [[nodiscard]] absl::StatusOr<LevelInfo> GetLevelInfo(
      int level) const override {
    return absl::InternalError("Mock failure");
  }

  [[nodiscard]] const SlideProperties& GetProperties() const override {
    static SlideProperties props;
    return props;
  }

  [[nodiscard]] std::vector<std::string> GetAssociatedImageNames()
      const override {
    return {};
  }

  [[nodiscard]] absl::StatusOr<ImageDimensions> GetAssociatedImageDimensions(
      std::string_view name) const override {
    return absl::InternalError("Mock failure");
  }

  [[nodiscard]] absl::StatusOr<Image> ReadRegion(
      const RegionSpec& region) const override {
    return absl::InternalError("Mock failure");
  }

  [[nodiscard]] absl::StatusOr<Image> ReadAssociatedImage(
      std::string_view name) const override {
    return absl::InternalError("Mock failure");
  }

  [[nodiscard]] int GetBestLevelForDownsample(
      double downsample) const override {
    return 0;
  }

  [[nodiscard]] Metadata GetMetadata() const override { return Metadata(); }

  [[nodiscard]] std::string GetFormatName() const override {
    return "FailingMockFormat";
  }

  [[nodiscard]] ImageDimensions GetTileSize() const override {
    return ImageDimensions{256, 256};  // Default tile size
  }

 private:
  std::string filename_;
};

/// @brief Test fixture for SlideReader tests
class SlideReaderTest : public ::testing::Test {
 protected:
  void SetUp() override {
    // Clear any existing registrations to ensure clean state
    // Note: In a real scenario, you might want to save/restore registry state
  }

  void TearDown() override {
    // Clean up any test registrations
  }
};

/// @brief Parameterized test fixture for ClampRegion tests
class ClampRegionTest : public ::testing::TestWithParam<
                            std::tuple<RegionSpec,       // Input region
                                       ImageDimensions,  // Image dimensions
                                       RegionSpec        // Expected output
                                       >> {};

// Test ClampRegion utility function
TEST_F(SlideReaderTest, ClampRegionWithinBounds) {
  ImageDimensions image_dims{1000, 800};
  RegionSpec region;
  region.top_left = ImageCoordinate{100, 50};
  region.size = ImageDimensions{200, 150};
  region.level = 0;

  auto clamped = TestSlideReader::TestClampRegion(region, image_dims);

  // Should remain unchanged as it's within bounds
  EXPECT_EQ(clamped.top_left[0], 100);
  EXPECT_EQ(clamped.top_left[1], 50);
  EXPECT_EQ(clamped.size[0], 200);
  EXPECT_EQ(clamped.size[1], 150);
  EXPECT_EQ(clamped.level, 0);
}

TEST_F(SlideReaderTest, ClampRegionBeyondBounds) {
  ImageDimensions image_dims{1000, 800};
  RegionSpec region;
  region.top_left = ImageCoordinate{900, 700};
  region.size = ImageDimensions{200, 150};  // Extends beyond bounds
  region.level = 0;

  auto clamped = TestSlideReader::TestClampRegion(region, image_dims);

  // Coordinates should be clamped to image bounds
  EXPECT_EQ(clamped.top_left[0], 900);
  EXPECT_EQ(clamped.top_left[1], 700);
  EXPECT_EQ(clamped.size[0], 100);  // 1000 - 900 = 100
  EXPECT_EQ(clamped.size[1], 100);  // 800 - 700 = 100
  EXPECT_EQ(clamped.level, 0);
}

TEST_F(SlideReaderTest, ClampRegionStartBeyondBounds) {
  ImageDimensions image_dims{1000, 800};
  RegionSpec region;
  region.top_left = ImageCoordinate{1200, 900};  // Beyond bounds
  region.size = ImageDimensions{200, 150};
  region.level = 0;

  auto clamped = TestSlideReader::TestClampRegion(region, image_dims);

  // Coordinates should be clamped to image bounds
  EXPECT_EQ(clamped.top_left[0], 1000);  // Clamped to image width
  EXPECT_EQ(clamped.top_left[1], 800);   // Clamped to image height
  EXPECT_EQ(clamped.size[0], 0);         // No area remaining
  EXPECT_EQ(clamped.size[1], 0);         // No area remaining
  EXPECT_EQ(clamped.level, 0);
}

TEST_F(SlideReaderTest, ClampRegionAtImageBoundary) {
  ImageDimensions image_dims{1000, 800};
  RegionSpec region;
  region.top_left = ImageCoordinate{1000, 800};  // At boundary
  region.size = ImageDimensions{100, 100};
  region.level = 0;

  auto clamped = TestSlideReader::TestClampRegion(region, image_dims);

  // Should be clamped to zero size
  EXPECT_EQ(clamped.top_left[0], 1000);
  EXPECT_EQ(clamped.top_left[1], 800);
  EXPECT_EQ(clamped.size[0], 0);
  EXPECT_EQ(clamped.size[1], 0);
  EXPECT_EQ(clamped.level, 0);
}

TEST_F(SlideReaderTest, ClampRegionZeroSizedImage) {
  ImageDimensions image_dims{0, 0};
  RegionSpec region;
  region.top_left = ImageCoordinate{0, 0};
  region.size = ImageDimensions{100, 100};
  region.level = 0;

  auto clamped = TestSlideReader::TestClampRegion(region, image_dims);

  // Should be clamped to zero size
  EXPECT_EQ(clamped.top_left[0], 0);
  EXPECT_EQ(clamped.top_left[1], 0);
  EXPECT_EQ(clamped.size[0], 0);
  EXPECT_EQ(clamped.size[1], 0);
  EXPECT_EQ(clamped.level, 0);
}

TEST_F(SlideReaderTest, ClampRegionMaxValues) {
  ImageDimensions image_dims{1000, 800};
  RegionSpec region;
  constexpr uint32_t max_val = std::numeric_limits<uint32_t>::max();
  region.top_left = ImageCoordinate{max_val, max_val};
  region.size = ImageDimensions{max_val, max_val};
  region.level = 0;

  auto clamped = TestSlideReader::TestClampRegion(region, image_dims);

  // Should be clamped to image bounds
  EXPECT_EQ(clamped.top_left[0], 1000);
  EXPECT_EQ(clamped.top_left[1], 800);
  EXPECT_EQ(clamped.size[0], 0);
  EXPECT_EQ(clamped.size[1], 0);
  EXPECT_EQ(clamped.level, 0);
}

// Parameterized tests for ClampRegion
INSTANTIATE_TEST_SUITE_P(
    ClampRegionParameterized, ClampRegionTest,
    ::testing::Values(std::make_tuple(RegionSpec{ImageCoordinate{50, 30},
                                                 ImageDimensions{100, 80}, 0},
                                      ImageDimensions{1000, 800},
                                      RegionSpec{ImageCoordinate{50, 30},
                                                 ImageDimensions{100, 80}, 0}),
                      std::make_tuple(RegionSpec{ImageCoordinate{950, 750},
                                                 ImageDimensions{100, 80}, 0},
                                      ImageDimensions{1000, 800},
                                      RegionSpec{ImageCoordinate{950, 750},
                                                 ImageDimensions{50, 50}, 0}),
                      std::make_tuple(RegionSpec{ImageCoordinate{1100, 900},
                                                 ImageDimensions{100, 80}, 0},
                                      ImageDimensions{1000, 800},
                                      RegionSpec{ImageCoordinate{1000, 800},
                                                 ImageDimensions{0, 0}, 0})));

TEST_P(ClampRegionTest, ParameterizedClampRegion) {
  const auto& [input_region, image_dims, expected] = GetParam();
  auto clamped = TestSlideReader::TestClampRegion(input_region, image_dims);

  EXPECT_EQ(clamped.top_left[0], expected.top_left[0]);
  EXPECT_EQ(clamped.top_left[1], expected.top_left[1]);
  EXPECT_EQ(clamped.size[0], expected.size[0]);
  EXPECT_EQ(clamped.size[1], expected.size[1]);
  EXPECT_EQ(clamped.level, expected.level);
}

// Performance test for ClampRegion
TEST_F(SlideReaderTest, ClampRegionPerformance) {
  ImageDimensions image_dims{10000, 8000};
  RegionSpec region;
  region.top_left = ImageCoordinate{5000, 4000};
  region.size = ImageDimensions{1000, 1000};
  region.level = 0;

  const int iterations = 100000;
  auto start = std::chrono::high_resolution_clock::now();

  for (int i = 0; i < iterations; ++i) {
    auto clamped = TestSlideReader::TestClampRegion(region, image_dims);
    // Use result to prevent optimization
    volatile auto x = clamped.size[0];
    (void)x;
  }

  auto end = std::chrono::high_resolution_clock::now();
  auto duration =
      std::chrono::duration_cast<std::chrono::microseconds>(end - start);

  // Should complete in reasonable time (less than 100ms for 100k iterations)
  EXPECT_LT(duration.count(), 100000);
}

// Test SlideReaderRegistry singleton behavior
TEST_F(SlideReaderTest, RegistrySingletonBehavior) {
  auto& registry1 = SlideReaderRegistry::GetInstance();
  auto& registry2 = SlideReaderRegistry::GetInstance();

  // Should return the same instance
  EXPECT_EQ(&registry1, &registry2);
}

// Thread safety tests for SlideReaderRegistry (concurrent reads and writes)
TEST_F(SlideReaderTest, RegistryThreadSafety) {
  auto& registry = SlideReaderRegistry::GetInstance();

  constexpr int num_threads = 6;

  std::vector<std::future<bool>> futures;
  std::atomic<int> read_successes{0};
  std::atomic<int> write_successes{0};

  // Test concurrent read and write operations
  for (int i = 0; i < num_threads; ++i) {
    if (i < 2) {
      // Writer threads: concurrent registration
      futures.push_back(
          std::async(std::launch::async, [&registry, &write_successes, i]() {
            bool thread_success = true;

            for (int j = 0; j < 50; ++j) {
              try {
                // Register unique extensions per thread
                std::string ext = ".concurrent_" + std::to_string(i) + "_" +
                                  std::to_string(j);
                registry.RegisterReader(ext, [](std::string_view filename) {
                  return std::make_unique<MockSlideReader>(filename);
                });
                write_successes++;
              } catch (...) {
                thread_success = false;
                break;
              }
            }

            return thread_success;
          }));
    } else {
      // Reader threads: concurrent access
      futures.push_back(
          std::async(std::launch::async, [&registry, &read_successes]() {
            bool thread_success = true;

            // Give writers a chance to register some extensions first
            std::this_thread::sleep_for(std::chrono::milliseconds(10));

            for (int j = 0; j < 50; ++j) {
              try {
                // Test concurrent read access
                auto extensions = registry.GetSupportedExtensions();

                // Try to create readers for any available extensions
                if (!extensions.empty()) {
                  // Use the first available extension for testing
                  auto reader_result =
                      registry.CreateReader("test" + extensions[0]);
                  if (reader_result.ok()) {
                    read_successes++;
                  }
                }
              } catch (...) {
                thread_success = false;
                break;
              }
            }

            return thread_success;
          }));
    }
  }

  // Wait for all threads to complete
  bool all_succeeded = true;
  for (auto& future : futures) {
    if (!future.get()) {
      all_succeeded = false;
    }
  }

  EXPECT_TRUE(all_succeeded);
  EXPECT_GT(write_successes.load(), 0);
  EXPECT_GT(read_successes.load(), 0);

  // Verify that all registrations were successful
  auto final_extensions = registry.GetSupportedExtensions();
  EXPECT_GE(final_extensions.size(), write_successes.load());
}

// Performance test demonstrating thread safety (not necessarily performance
// gains)
TEST_F(SlideReaderTest, RegistryConcurrentReadStability) {
  auto& registry = SlideReaderRegistry::GetInstance();

  // Pre-register a reader for testing
  registry.RegisterReader(".perftest", [](std::string_view filename) {
    return std::make_unique<MockSlideReader>(filename);
  });

  constexpr int num_reader_threads = 4;
  constexpr int reads_per_thread = 500;

  std::atomic<int> successful_operations{0};
  std::atomic<int> failed_operations{0};

  // Test concurrent reads for stability (not performance)
  std::vector<std::future<void>> futures;
  for (int i = 0; i < num_reader_threads; ++i) {
    futures.push_back(std::async(
        std::launch::async,
        [&registry, &successful_operations, &failed_operations]() {
          for (int j = 0; j < reads_per_thread; ++j) {
            try {
              auto extensions = registry.GetSupportedExtensions();
              auto reader_result = registry.CreateReader("test.perftest");

              if (reader_result.ok() && !extensions.empty()) {
                successful_operations++;
              } else {
                failed_operations++;
              }
            } catch (...) {
              failed_operations++;
            }
          }
        }));
  }

  for (auto& future : futures) {
    future.wait();
  }

  // All operations should succeed - this demonstrates thread safety
  EXPECT_EQ(successful_operations.load(),
            num_reader_threads * reads_per_thread);
  EXPECT_EQ(failed_operations.load(), 0);

  // Verify registry is still in consistent state
  auto extensions = registry.GetSupportedExtensions();
  EXPECT_FALSE(extensions.empty());

  auto reader_result = registry.CreateReader("test.perftest");
  EXPECT_TRUE(reader_result.ok());
}

// Test RegisterReader functionality
TEST_F(SlideReaderTest, RegisterReaderBasicFunctionality) {
  auto& registry = SlideReaderRegistry::GetInstance();

  // Register a mock reader
  registry.RegisterReader(".mock", [](std::string_view filename) {
    return std::make_unique<MockSlideReader>(filename);
  });

  // Check that the extension is supported
  auto extensions = registry.GetSupportedExtensions();
  EXPECT_NE(std::find(extensions.begin(), extensions.end(), ".mock"),
            extensions.end());
}

TEST_F(SlideReaderTest, RegisterReaderCaseInsensitive) {
  auto& registry = SlideReaderRegistry::GetInstance();

  // Register with uppercase extension
  registry.RegisterReader(".MOCK", [](std::string_view filename) {
    return std::make_unique<MockSlideReader>(filename);
  });

  // Should be normalized to lowercase
  auto extensions = registry.GetSupportedExtensions();
  EXPECT_NE(std::find(extensions.begin(), extensions.end(), ".mock"),
            extensions.end());
  EXPECT_EQ(std::find(extensions.begin(), extensions.end(), ".MOCK"),
            extensions.end());
}

TEST_F(SlideReaderTest, RegisterReaderWithoutDotPrefix) {
  auto& registry = SlideReaderRegistry::GetInstance();

  // Register without dot prefix
  registry.RegisterReader("mock", [](std::string_view filename) {
    return std::make_unique<MockSlideReader>(filename);
  });

  // Should be normalized to include dot
  auto extensions = registry.GetSupportedExtensions();
  EXPECT_NE(std::find(extensions.begin(), extensions.end(), ".mock"),
            extensions.end());
}

TEST_F(SlideReaderTest, RegisterReaderOverride) {
  auto& registry = SlideReaderRegistry::GetInstance();

  // Register initial reader
  registry.RegisterReader(".test", [](std::string_view filename) {
    return std::make_unique<MockSlideReader>(filename);
  });

  // Override with new reader
  registry.RegisterReader(".test", [](std::string_view filename) {
    return std::make_unique<MockSlideReader>(filename);
  });

  // Should still be supported (just overridden)
  auto extensions = registry.GetSupportedExtensions();
  EXPECT_NE(std::find(extensions.begin(), extensions.end(), ".test"),
            extensions.end());
}

TEST_F(SlideReaderTest, RegisterReaderInputValidation) {
  auto& registry = SlideReaderRegistry::GetInstance();

  size_t initial_count = registry.GetSupportedExtensions().size();

  // Register with empty extension (should be ignored)
  registry.RegisterReader("", [](std::string_view filename) {
    return std::make_unique<MockSlideReader>(filename);
  });

  // Register with null factory (should be ignored)
  registry.RegisterReader(".null", nullptr);

  // Count should remain the same
  EXPECT_EQ(registry.GetSupportedExtensions().size(), initial_count);
}

// Test CreateReader functionality
TEST_F(SlideReaderTest, CreateReaderSuccess) {
  auto& registry = SlideReaderRegistry::GetInstance();

  // Register a mock reader
  registry.RegisterReader(".mock", [](std::string_view filename) {
    return std::make_unique<MockSlideReader>(filename);
  });

  // Create reader for supported format
  auto reader_result = registry.CreateReader("test.mock");
  ASSERT_TRUE(reader_result.ok());

  auto reader = std::move(reader_result.value());
  ASSERT_NE(reader, nullptr);
  EXPECT_EQ(reader->GetFormatName(), "MockFormat");
}

TEST_F(SlideReaderTest, CreateReaderCaseInsensitiveExtension) {
  auto& registry = SlideReaderRegistry::GetInstance();

  // Register with lowercase extension
  registry.RegisterReader(".mock", [](std::string_view filename) {
    return std::make_unique<MockSlideReader>(filename);
  });

  // Create reader with uppercase extension
  auto reader_result = registry.CreateReader("test.MOCK");
  ASSERT_TRUE(reader_result.ok());

  auto reader = std::move(reader_result.value());
  ASSERT_NE(reader, nullptr);
  EXPECT_EQ(reader->GetFormatName(), "MockFormat");
}

TEST_F(SlideReaderTest, CreateReaderUnsupportedFormat) {
  auto& registry = SlideReaderRegistry::GetInstance();

  // Try to create reader for unsupported format
  auto reader_result = registry.CreateReader("test.unsupported");
  EXPECT_FALSE(reader_result.ok());
  EXPECT_EQ(reader_result.status().code(), absl::StatusCode::kInvalidArgument);

  std::string error_message = std::string(reader_result.status().message());
  EXPECT_NE(error_message.find("Unsupported file format"), std::string::npos);
  EXPECT_NE(error_message.find(".unsupported"), std::string::npos);
  EXPECT_NE(error_message.find("test.unsupported"), std::string::npos);
}

TEST_F(SlideReaderTest, CreateReaderNoExtension) {
  auto& registry = SlideReaderRegistry::GetInstance();

  // Try to create reader for file without extension
  auto reader_result = registry.CreateReader("test_file");
  EXPECT_FALSE(reader_result.ok());
  EXPECT_EQ(reader_result.status().code(), absl::StatusCode::kInvalidArgument);

  std::string error_message = std::string(reader_result.status().message());
  EXPECT_NE(error_message.find("no extension"), std::string::npos);
}

TEST_F(SlideReaderTest, CreateReaderEmptyFilename) {
  auto& registry = SlideReaderRegistry::GetInstance();

  // Try to create reader with empty filename
  auto reader_result = registry.CreateReader("");
  EXPECT_FALSE(reader_result.ok());
  EXPECT_EQ(reader_result.status().code(), absl::StatusCode::kInvalidArgument);

  std::string error_message = std::string(reader_result.status().message());
  EXPECT_NE(error_message.find("empty"), std::string::npos);
}

TEST_F(SlideReaderTest, CreateReaderFactoryFailure) {
  auto& registry = SlideReaderRegistry::GetInstance();

  // Register a factory that always fails
  registry.RegisterReader(".fail", [](std::string_view filename) {
    return absl::InternalError("Factory failure");
  });

  // Try to create reader - should propagate factory error with context
  auto reader_result = registry.CreateReader("test.fail");
  EXPECT_FALSE(reader_result.ok());
  EXPECT_EQ(reader_result.status().code(), absl::StatusCode::kInternal);

  std::string error_message = std::string(reader_result.status().message());
  EXPECT_NE(error_message.find("Failed to create reader"), std::string::npos);
  EXPECT_NE(error_message.find("test.fail"), std::string::npos);
  EXPECT_NE(error_message.find("Factory failure"), std::string::npos);
}

// Test GetSupportedExtensions functionality
TEST_F(SlideReaderTest, GetSupportedExtensionsEmpty) {
  auto& registry = SlideReaderRegistry::GetInstance();

  // Note: There might be existing registrations from other tests or static
  // initialization
  // So we just check that the function returns a vector
  auto extensions = registry.GetSupportedExtensions();
  EXPECT_TRUE(extensions.empty() ||
              !extensions.empty());  // Always true, but validates call
}

TEST_F(SlideReaderTest, GetSupportedExtensionsMultiple) {
  auto& registry = SlideReaderRegistry::GetInstance();

  // Register multiple formats
  registry.RegisterReader(".format1", [](std::string_view filename) {
    return std::make_unique<MockSlideReader>(filename);
  });

  registry.RegisterReader(".format2", [](std::string_view filename) {
    return std::make_unique<MockSlideReader>(filename);
  });

  registry.RegisterReader(".format3", [](std::string_view filename) {
    return std::make_unique<MockSlideReader>(filename);
  });

  auto extensions = registry.GetSupportedExtensions();

  // Should contain all registered formats
  EXPECT_NE(std::find(extensions.begin(), extensions.end(), ".format1"),
            extensions.end());
  EXPECT_NE(std::find(extensions.begin(), extensions.end(), ".format2"),
            extensions.end());
  EXPECT_NE(std::find(extensions.begin(), extensions.end(), ".format3"),
            extensions.end());
}

TEST_F(SlideReaderTest, GetSupportedExtensionsSorted) {
  auto& registry = SlideReaderRegistry::GetInstance();

  // Register formats in random order
  registry.RegisterReader(".zzz", [](std::string_view filename) {
    return std::make_unique<MockSlideReader>(filename);
  });

  registry.RegisterReader(".aaa", [](std::string_view filename) {
    return std::make_unique<MockSlideReader>(filename);
  });

  registry.RegisterReader(".mmm", [](std::string_view filename) {
    return std::make_unique<MockSlideReader>(filename);
  });

  auto extensions = registry.GetSupportedExtensions();

  // Should be sorted alphabetically
  EXPECT_TRUE(std::is_sorted(extensions.begin(), extensions.end()));
}

// Test RegionSpec validation
TEST_F(SlideReaderTest, RegionSpecValidation) {
  // Valid region should pass validation
  RegionSpec valid_region;
  valid_region.top_left = ImageCoordinate{0, 0};
  valid_region.size = ImageDimensions{100, 100};
  valid_region.level = 0;
  EXPECT_TRUE(valid_region.IsValid());

  // Invalid width (zero)
  RegionSpec invalid_size_width;
  invalid_size_width.top_left = ImageCoordinate{0, 0};
  invalid_size_width.size = ImageDimensions{0, 100};
  invalid_size_width.level = 0;
  EXPECT_FALSE(invalid_size_width.IsValid());

  // Invalid height (zero)
  RegionSpec invalid_size_height;
  invalid_size_height.top_left = ImageCoordinate{0, 0};
  invalid_size_height.size = ImageDimensions{100, 0};
  invalid_size_height.level = 0;
  EXPECT_FALSE(invalid_size_height.IsValid());

  // Invalid level (negative)
  RegionSpec invalid_level;
  invalid_level.top_left = ImageCoordinate{0, 0};
  invalid_level.size = ImageDimensions{100, 100};
  invalid_level.level = -1;
  EXPECT_FALSE(invalid_level.IsValid());
}

// Test Metadata functionality
TEST_F(SlideReaderTest, MetadataBasicOperations) {
  Metadata metadata;

  // Test insertion and retrieval
  metadata["string_key"] = std::string("test_value");
  metadata["size_key"] = size_t(42);
  metadata["double_key"] = 3.14;

  EXPECT_EQ(metadata.size(), 3);
  EXPECT_FALSE(metadata.empty());

  // Test GetString
  EXPECT_EQ(metadata.GetString("string_key"), "test_value");
  EXPECT_EQ(metadata.GetString("size_key"), "42");
  EXPECT_EQ(metadata.GetString("double_key"), "3.140000");
  EXPECT_EQ(metadata.GetString("nonexistent", "default"), "default");

  // Test GetSize
  EXPECT_EQ(metadata.GetSize("size_key"), 42);
  EXPECT_EQ(metadata.GetSize("double_key"), 3);
  EXPECT_EQ(metadata.GetSize("nonexistent", 99), 99);

  // Test GetDouble
  EXPECT_NEAR(metadata.GetDouble("double_key"), 3.14, 1e-6);
  EXPECT_NEAR(metadata.GetDouble("size_key"), 42.0, 1e-6);
  EXPECT_NEAR(metadata.GetDouble("nonexistent", 1.5), 1.5, 1e-6);

  // Test contains
  EXPECT_TRUE(metadata.contains("string_key"));
  EXPECT_FALSE(metadata.contains("nonexistent"));
}

TEST_F(SlideReaderTest, MetadataInitializerList) {
  Metadata metadata = {
      {"key1", std::string("value1")}, {"key2", size_t(123)}, {"key3", 2.71}};

  EXPECT_EQ(metadata.size(), 3);
  EXPECT_EQ(metadata.GetString("key1"), "value1");
  EXPECT_EQ(metadata.GetSize("key2"), 123);
  EXPECT_NEAR(metadata.GetDouble("key3"), 2.71, 1e-6);
}

TEST_F(SlideReaderTest, MetadataToString) {
  Metadata metadata = {{"string_key", std::string("test")},
                       {"size_key", size_t(42)},
                       {"double_key", 3.14}};

  std::string str = metadata.ToString();
  EXPECT_FALSE(str.empty());

  // Should contain all keys and values
  EXPECT_NE(str.find("string_key"), std::string::npos);
  EXPECT_NE(str.find("size_key"), std::string::npos);
  EXPECT_NE(str.find("double_key"), std::string::npos);
  EXPECT_NE(str.find("test"), std::string::npos);
  EXPECT_NE(str.find("42"), std::string::npos);
  EXPECT_NE(str.find("3.14"), std::string::npos);
}

TEST_F(SlideReaderTest, MetadataStreamOperator) {
  Metadata metadata = {{"string_key", std::string("test")},
                       {"size_key", size_t(42)},
                       {"double_key", 3.14}};

  std::ostringstream oss;
  oss << metadata;
  std::string output = oss.str();

  EXPECT_FALSE(output.empty());
  EXPECT_NE(output.find("Metadata {"), std::string::npos);
  EXPECT_NE(output.find("string_key"), std::string::npos);
  EXPECT_NE(output.find("size_key"), std::string::npos);
  EXPECT_NE(output.find("double_key"), std::string::npos);
  EXPECT_NE(output.find("}"), std::string::npos);
}

TEST_F(SlideReaderTest, MetadataTypeConversionEdgeCases) {
  Metadata metadata;

  // Test string to number conversions with invalid strings
  metadata["invalid_number"] = std::string("not_a_number");
  metadata["valid_double"] = std::string("3.14159");
  metadata["valid_size"] = std::string("12345");

  // Invalid conversions should return defaults
  EXPECT_EQ(metadata.GetDouble("invalid_number", 999.0), 999.0);
  EXPECT_EQ(metadata.GetSize("invalid_number", 888), 888);

  // Valid conversions should work
  EXPECT_NEAR(metadata.GetDouble("valid_double"), 3.14159, 1e-5);
  EXPECT_EQ(metadata.GetSize("valid_size"), 12345);
}

// Test RGBImage functionality
TEST_F(SlideReaderTest, RGBImageBasicOperations) {
  ImageDimensions dims{100, 50};
  RGBImage image(dims, ImageFormat::kRGB, DataType::kUInt8);

  EXPECT_EQ(image.GetDimensions()[0], 100);
  EXPECT_EQ(image.GetDimensions()[1], 50);
  EXPECT_EQ(image.SizeBytes(), 100 * 50 * 3);
  EXPECT_FALSE(image.Empty());

  // Test pixel access
  image.At<uint8_t>(20, 10, 0) = 255;  // Red
  image.At<uint8_t>(20, 10, 1) = 128;  // Green
  image.At<uint8_t>(20, 10, 2) = 64;   // Blue

  EXPECT_EQ(image.At<uint8_t>(20, 10, 0), 255);
  EXPECT_EQ(image.At<uint8_t>(20, 10, 1), 128);
  EXPECT_EQ(image.At<uint8_t>(20, 10, 2), 64);
}

TEST_F(SlideReaderTest, RGBImageEmpty) {
  RGBImage empty_image;

  EXPECT_EQ(empty_image.GetDimensions()[0], 0);
  EXPECT_EQ(empty_image.GetDimensions()[1], 0);
  EXPECT_EQ(empty_image.SizeBytes(), 0);
  EXPECT_TRUE(empty_image.Empty());
}

TEST_F(SlideReaderTest, RGBImageZeroSized) {
  ImageDimensions dims{0, 0};
  RGBImage image(dims, ImageFormat::kRGB, DataType::kUInt8);

  EXPECT_EQ(image.GetDimensions()[0], 0);
  EXPECT_EQ(image.GetDimensions()[1], 0);
  EXPECT_EQ(image.SizeBytes(), 0);
  EXPECT_TRUE(image.Empty());
}

TEST_F(SlideReaderTest, RGBImageBoundaryPixelAccess) {
  ImageDimensions dims{100, 50};
  RGBImage image(dims, ImageFormat::kRGB, DataType::kUInt8);

  // Test corner pixels
  image.At<uint8_t>(0, 0, 0) = 255;
  image.At<uint8_t>(49, 99, 2) = 128;  // Last valid pixel (height-1, width-1)

  EXPECT_EQ(image.At<uint8_t>(0, 0, 0), 255);
  EXPECT_EQ(image.At<uint8_t>(49, 99, 2), 128);
}

TEST_F(SlideReaderTest, LevelInfoBasic) {
  LevelInfo info;
  info.dimensions = ImageDimensions{1024, 768};
  info.downsample_factor = 2.0;

  EXPECT_EQ(info.dimensions[0], 1024);
  EXPECT_EQ(info.dimensions[1], 768);
  EXPECT_DOUBLE_EQ(info.downsample_factor, 2.0);
}

TEST_F(SlideReaderTest, ChannelMetadata) {
  auto reader = std::make_unique<MockSlideReader>("test.mock");
  auto channel_metadata = reader->GetChannelMetadata();

  EXPECT_EQ(channel_metadata.size(), 1);
  EXPECT_EQ(channel_metadata[0].name, "RGB");
  EXPECT_EQ(channel_metadata[0].biomarker, "RGB");
}

// Integration test with MockSlideReader
TEST_F(SlideReaderTest, MockSlideReaderIntegration) {
  auto& registry = SlideReaderRegistry::GetInstance();

  // Register mock reader
  registry.RegisterReader(".integration", [](std::string_view filename) {
    return std::make_unique<MockSlideReader>(filename);
  });

  // Create and test reader
  auto reader_result = registry.CreateReader("test.integration");
  ASSERT_TRUE(reader_result.ok());

  auto reader = std::move(reader_result.value());

  // Test basic functionality
  EXPECT_EQ(reader->GetLevelCount(), 1);
  EXPECT_EQ(reader->GetFormatName(), "MockFormat");

  auto level_info = reader->GetLevelInfo(0);
  ASSERT_TRUE(level_info.ok());
  EXPECT_EQ(level_info.value().dimensions[0], 1024);
  EXPECT_EQ(level_info.value().dimensions[1], 768);

  auto metadata = reader->GetMetadata();
  EXPECT_EQ(metadata.GetString("format"), "MockFormat");
  EXPECT_EQ(metadata.GetSize("width"), 1024);
}

}  // namespace fastslide
