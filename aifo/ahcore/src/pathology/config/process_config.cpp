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
#include "ahcore/pathology/config/process_config.h"
#include "CLI11/CLI11.hpp"

namespace aifo::pathology::config {

ProcessConfig ProcessConfig::FromArgs(int argc, char** argv) {
  ProcessConfig config;

  CLI::App app{"Ahcore inference"};

  app.add_option("--image", config.image_file,
                 "Path to input image file (required)")
      ->required();
  app.add_option("--output", config.output_file,
                 "Path to output file (required)")
      ->required();
  app.add_option("--model", config.model_file,
                 "Path to model pack file (required)")
      ->required();
  app.add_option("--mask", config.mask_file,
                 "Optional mask file (TIFF or XML)");
  app.add_option("--mask-threshold", config.mask_threshold,
                 "Threshold for the mask");
  app.add_flag("--create-thumbnail", config.create_thumbnail,
               "Create a thumbnail image");
  app.add_option("--batch-size", config.batch_size, "Batch size for inference");
  app.add_option("--tile-size", config.tiff_tile_size,
                 "Tile size to write TIFF");
  app.add_option("--device", config.device, "Device to use: cpu, mps, or cuda");

  try {
    app.parse(argc, argv);
  } catch (const CLI::ParseError& e) {
    std::exit(app.exit(e));
  }

  return config;
}

}  // namespace aifo::pathology::config
