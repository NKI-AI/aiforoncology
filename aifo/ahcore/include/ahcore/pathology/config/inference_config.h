// Copyright 2024 Jonas Teuwen. All Rights Reserved.
// Copyright 2025 Joren Brunekreef. All Rights Reserved.
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
#ifndef AIFO_AHCORE_INCLUDE_AHCORE_PATHOLOGY_CONFIG_INFERENCE_CONFIG_H_
#define AIFO_AHCORE_INCLUDE_AHCORE_PATHOLOGY_CONFIG_INFERENCE_CONFIG_H_

#include <filesystem>
#include <optional>
#include <string>

namespace aifo::pathology::config {

namespace fs = std::filesystem;

/**
 * @brief Configuration structure for the image processing pipeline
 */
struct InferenceConfig {
  fs::path output_dir;
  fs::path model_path;
  double mask_threshold;
  int batch_size{4};
  double tiff_tile_size{512};
  std::optional<std::string> device;
  bool create_thumbnail{false};

  /**
     * @brief Creates a InferenceConfig from flags
     */
  static InferenceConfig FromArgs(int argc, char** argv);
};

}  // namespace aifo::pathology::config

#endif  // AIFO_AHCORE_INCLUDE_AHCORE_PATHOLOGY_CONFIG_INFERENCE_CONFIG_H_
