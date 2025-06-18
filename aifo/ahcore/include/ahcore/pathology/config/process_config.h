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
#ifndef AIFO_AHCORE_INCLUDE_AHCORE_PATHOLOGY_CONFIG_PROCESS_CONFIG_H_
#define AIFO_AHCORE_INCLUDE_AHCORE_PATHOLOGY_CONFIG_PROCESS_CONFIG_H_

#include <filesystem>
#include <optional>
#include <string>

namespace aifo::pathology::config {

namespace fs = std::filesystem;

/**
 * @brief Configuration structure for the image processing pipeline
 */
struct ProcessConfig {
  fs::path image_file;
  fs::path output_file;
  fs::path model_file;
  fs::path mask_file;
  double mask_threshold;
  bool create_thumbnail{false};
  int batch_size{4};
  double tiff_tile_size{512};
  std::optional<std::string> device;

  /**
     * @brief Creates a ProcessConfig from flags
     */
  static ProcessConfig FromArgs(int argc, char** argv);
};

}  // namespace aifo::pathology::config

#endif  // AIFO_AHCORE_INCLUDE_AHCORE_PATHOLOGY_CONFIG_PROCESS_CONFIG_H_
