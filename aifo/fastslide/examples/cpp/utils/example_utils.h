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

#ifndef AIFO_FASTSLIDE_EXAMPLES_CPP_UTILS_EXAMPLE_UTILS_H_
#define AIFO_FASTSLIDE_EXAMPLES_CPP_UTILS_EXAMPLE_UTILS_H_

#include <filesystem>
#include <memory>
#include <string>

#include "fastslide/image.h"
#include "fastslide/slide_reader.h"

namespace fastslide::examples {

/**
 * @brief Configuration struct for CLI parameters
 */
struct Config {
  std::string slide_file;
  std::filesystem::path output_dir;
};

/**
 * @brief Save an RGB image as PNG using libvips
 * 
 * @param image The RGB image to save (must be uint8 format)
 * @param filename Output PNG filename
 * @return true if successful, false otherwise
 */
bool SaveAsPNG(const Image& image, const std::string& filename);

/**
 * @brief Load an image from PNG using libvips
 * 
 * @param filename Input PNG filename
 * @return Loaded image as RGB uint8, or nullptr if failed
 */
std::unique_ptr<Image> LoadFromPNG(const std::string& filename);

/**
 * @brief Parse CLI arguments and initialize required libraries
 * 
 * @param argc Command line argument count
 * @param argv Command line argument values
 * @param app_description Description for the CLI app
 * @return Config with parsed parameters, or exits on error
 */
Config ParseCliAndInit(int argc, char** argv,
                       const std::string& app_description);

/**
 * @brief Print basic slide information
 * 
 * @param reader The slide reader to get information from
 */
void PrintSlideInfo(const SlideReader& reader);

/**
 * @brief Initialize vips and cleanup on exit
 * 
 * @param argv Command line arguments (needed for vips init)
 */
void InitializeVips(char** argv);

/**
 * @brief Cleanup vips on exit
 */
void CleanupVips();

}  // namespace fastslide::examples

#endif  // AIFO_FASTSLIDE_EXAMPLES_CPP_UTILS_EXAMPLE_UTILS_H_
