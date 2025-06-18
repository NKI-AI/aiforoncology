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
#ifndef AIFO_AHCORE_INCLUDE_AHCORE_PATHOLOGY_CONFIGURATION_H_
#define AIFO_AHCORE_INCLUDE_AHCORE_PATHOLOGY_CONFIGURATION_H_

#include <libxml/parser.h>
#include <libxml/tree.h>

#include <iostream>
#include <memory>
#include <regex>
#include <sstream>
#include <stdexcept>
#include <string>
#include <tuple>
#include <utility>
#include <vector>

#include "aifocore/concepts/numeric.h"

// Struct for a label
struct Label {
  std::string name;
  std::string hex_color;
  int index;

  [[nodiscard]] std::tuple<int, int, int> RgbColor() const {
    return HexToRgb(hex_color);
  }

  static std::tuple<int, int, int> HexToRgb(const std::string& hex_color) {
    if (!std::regex_match(hex_color, std::regex("^#[0-9a-fA-F]{6}$"))) {
      throw std::invalid_argument("Invalid hex color: " + hex_color);
    }

    int r = std::stoi(hex_color.substr(1, 2), nullptr, 16);
    int g = std::stoi(hex_color.substr(3, 2), nullptr, 16);
    int b = std::stoi(hex_color.substr(5, 2), nullptr, 16);
    return {r, g, b};
  }
};

// Struct for normalization values
struct Normalization {
  std::vector<float> mean;  // Mean values for normalization
  std::vector<float> std;   // Standard deviation values for normalization

  [[nodiscard]] std::string AsFormattedString() const {
    std::ostringstream oss;
    oss << "Normalization:\n  Mean: ";
    for (const auto& val : mean) {
      oss << val << " ";
    }
    oss << "\n  Std: ";
    for (const auto& val : std) {
      oss << val << " ";
    }
    return oss.str();
  }
};

// Struct for the full configuration
struct AifoModelConfiguration {
  std::string model_name;
  std::string version;
  std::string task_type;
  double mpp;
  aifocore::Size<int, 2> tile_size;
  aifocore::Size<int, 2> tile_overlap;
  std::string merge_method;
  std::vector<Label> labels;
  Normalization normalization;

  [[nodiscard]] std::string AsFormattedString() const {
    std::ostringstream oss;
    oss << "Model Name: " << model_name << "\n"
        << "Version: " << version << "\n"
        << "Task Type: " << task_type << "\n"
        << "mpp: " << mpp << "\n"
        << "Tile Size: " << tile_size << "\n"
        << "Tile Overlap: " << tile_overlap << "\n"
        << "Merge Method: " << merge_method << "\n"
        << "Labels:\n";
    for (const auto& label : labels) {
      const auto [r, g, b] = label.RgbColor();
      oss << "  Name: " << label.name << ", Color: " << label.hex_color
          << " (RGB: " << r << ", " << g << ", " << b
          << "), Index: " << label.index << "\n";
    }
    oss << normalization.AsFormattedString();
    return oss.str();
  }

  void Print() const { std::cout << AsFormattedString(); }
};

// Helper to get text content of an XML node
std::string GetNodeContent(const xmlNode* node);

// Core parsing function
std::unique_ptr<AifoModelConfiguration> ParseConfiguration(xmlDoc* doc);

// Parsing from file
std::unique_ptr<AifoModelConfiguration> ParseConfigurationFromFile(
    const std::string& filename);

// Parsing from string
std::unique_ptr<AifoModelConfiguration> ParseConfigurationFromString(
    const std::string& xml_content);

#endif  // AIFO_AHCORE_INCLUDE_AHCORE_PATHOLOGY_CONFIGURATION_H_
