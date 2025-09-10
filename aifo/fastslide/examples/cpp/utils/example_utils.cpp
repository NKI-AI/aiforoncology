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

#include "./example_utils.h"

#include <vips/vips8>

#include <cstdlib>
#include <iostream>
#include <memory>
#include <string>

#include "CLI11/CLI11.hpp"
#include "fastslide/image.h"
#include "fastslide/slide_reader.h"

namespace fastslide::examples {

bool SaveAsPNG(const Image& image, const std::string& filename) {
  if (image.GetWidth() == 0 || image.GetHeight() == 0) {
    std::cerr << "Invalid image dimensions\n";
    return false;
  }

  // Assume input is already RGB(A) uint8 format
  if (image.GetDataType() != DataType::kUInt8) {
    std::cerr << "Expected uint8 RGB(A) image for PNG output\n";
    return false;
  }

  // Create VipsImage from raw data
  vips::VImage vips_image = vips::VImage::new_from_memory(
      const_cast<void*>(static_cast<const void*>(image.GetData())),
      image.SizeBytes(), image.GetWidth(), image.GetHeight(),
      image.GetChannels(), VIPS_FORMAT_UCHAR);

  try {
    vips_image.write_to_file(filename.c_str());
    return true;
  } catch (const vips::VError& e) {
    std::cerr << "Failed to save PNG: " << e.what() << "\n";
    return false;
  }
}

std::unique_ptr<Image> LoadFromPNG(const std::string& filename) {
  try {
    // Load PNG using vips
    vips::VImage vips_image = vips::VImage::new_from_file(filename.c_str());

    // Ensure it's in RGB format
    if (vips_image.bands() != 3) {
      // Convert to RGB if not already
      if (vips_image.bands() == 1) {
        // Grayscale to RGB
        vips_image = vips_image.bandjoin({vips_image, vips_image});
      } else if (vips_image.bands() == 4) {
        // RGBA to RGB (drop alpha)
        vips_image =
            vips_image.extract_band(0, vips::VImage::option()->set("n", 3));
      }
    }

    // Ensure it's uint8
    if (vips_image.format() != VIPS_FORMAT_UCHAR) {
      vips_image = vips_image.cast(VIPS_FORMAT_UCHAR);
    }

    // Get image properties
    int width = vips_image.width();
    int height = vips_image.height();
    int bands = vips_image.bands();

    // Create FastSlide image
    auto image = CreateRGBImage(
        {static_cast<uint32_t>(width), static_cast<uint32_t>(height)},
        DataType::kUInt8);

    // Copy data from vips to FastSlide image
    size_t data_size = width * height * bands;
    auto* vips_data = static_cast<const uint8_t*>(vips_image.data());
    auto* fastslide_data = image->GetDataAs<uint8_t>();

    std::memcpy(fastslide_data, vips_data, data_size);

    return image;

  } catch (const vips::VError& e) {
    std::cerr << "Failed to load PNG: " << e.what() << "\n";
    return nullptr;
  }
}

Config ParseCliAndInit(int argc, char** argv,
                       const std::string& app_description) {
  Config config;
  std::string output_dir = ".";  // Default to current directory

  CLI::App app{app_description};

  app.add_option("-f,--file", config.slide_file,
                 "Path to the slide file to test")
      ->required();

  app.add_option("-o,--output", output_dir, "Output directory for test images")
      ->default_val(".");

  try {
    app.parse(argc, argv);
  } catch (const CLI::ParseError& e) {
    std::exit(app.exit(e));
  }

  config.output_dir = std::filesystem::path(output_dir);

  // Create output directory if it doesn't exist
  if (!std::filesystem::exists(config.output_dir)) {
    std::filesystem::create_directories(config.output_dir);
  }

  return config;
}

void PrintSlideInfo(const SlideReader& reader) {
  auto channel_metadata = reader.GetChannelMetadata();
  auto level_info_or = reader.GetLevelInfo(0);

  if (level_info_or.ok()) {
    const auto& level_info = level_info_or.value();
    std::cout << reader.GetFormatName()
              << " slide: " << level_info.dimensions[0] << "x"
              << level_info.dimensions[1] << ", " << reader.GetLevelCount()
              << " levels"
              << ", " << channel_metadata.size() << " channels\n";
  }

  // Print key properties
  auto props = reader.GetProperties();
  std::cout << "MPP: " << props.mpp
            << " μm/pixel, Scanner: " << props.scanner_model << "\n";
}

void InitializeVips(char** argv) {
  if (VIPS_INIT(argv[0])) {
    vips_error_exit(nullptr);
  }
}

void CleanupVips() {
  vips_shutdown();
}

}  // namespace fastslide::examples
