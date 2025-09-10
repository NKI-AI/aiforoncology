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
#include "ahcore/data/metadata.h"

#include <dlup/slide_geometry.h>
#include <algorithm>
#include <filesystem>
#include <iostream>
#include <ostream>
#include <string>
#include <type_traits>
#include <vector>

#include "aifocore/concepts/numeric.h"

namespace aifo::data {

std::ostream& operator<<(std::ostream& os, const Metadata& metadata) {
  os << "Metadata Contents:\n";

  // First pass: determine max key length for proper alignment
  size_t max_key_length = 0;
  for (const auto& [key, _] : metadata.GetAll()) {
    max_key_length = std::max(max_key_length, key.length());
  }

  // Second pass: print with proper alignment
  for (const auto& [key, value] : metadata.GetAll()) {
    // Print key and colon with consistent spacing
    os << key << ":";

    // Add spacing after the colon for alignment
    int spaces_needed = max_key_length - key.length() + 1;
    os << std::string(spaces_needed, ' ');

    std::visit(
        [&os](const auto& v) {
          using T = std::decay_t<decltype(v)>;
          if constexpr (std::is_same_v<T, aifocore::Size<int, 2>>) {
            os << "Size(" << v[0] << ", " << v[1] << ")";
          } else if constexpr (std::is_same_v<T, std::filesystem::path>) {
            os << "Path(" << v.string() << ")";
          } else if constexpr (std::is_same_v<T, std::string>) {
            os << "String(\"" << v << "\")";
          } else if constexpr (std::is_same_v<T, std::vector<std::string>>) {
            os << "Vector<string>[";
            for (size_t i = 0; i < v.size(); ++i) {
              os << "\"" << v[i] << "\"";
              if (i < v.size() - 1) {
                os << ", ";
              }
            }
            os << "]";
          } else if constexpr (std::is_same_v<T, std::vector<int>> ||
                               std::is_same_v<T, std::vector<double>> ||
                               std::is_same_v<T, std::vector<float>>) {
            os << "Vector<";
            if constexpr (std::is_same_v<T, std::vector<int>>) {
              os << "int";
            } else if constexpr (std::is_same_v<T, std::vector<double>>) {
              os << "double";
            } else if constexpr (std::is_same_v<T, std::vector<float>>) {
              os << "float";
            }
            os << ">[";
            for (size_t i = 0; i < v.size(); ++i) {
              os << v[i];
              if (i < v.size() - 1) {
                os << ", ";
              }
            }
            os << "]";
          } else if constexpr (std::is_same_v<T, dlup::SlideGeometry>) {
            os << "SlideGeometry(size={" << v.size[0] << ", " << v.size[1]
               << "}, offset={" << v.offset[0] << ", " << v.offset[1]
               << "}, bounds={" << v.bounds[0] << ", " << v.bounds[1] << "})";
          } else if constexpr (std::is_integral_v<T>) {
            os << "Int(" << v << ")";
          } else if constexpr (std::is_floating_point_v<T>) {
            os << "Float(" << v << ")";
          } else if constexpr (std::is_same_v<T, aifocore::tiling::GridOrder>) {
            os << "GridOrder(";
            switch (v) {
              case aifocore::tiling::GridOrder::kC:
                os << "C-order";
                break;
              case aifocore::tiling::GridOrder::kF:
                os << "Fortran-order";
                break;
              default:
                os << "Unknown";
            }
            os << ")";
          } else {
            os << "UnknownType";
          }
        },
        value);
    os << "\n";
  }

  os << "Total fields: " << metadata.GetCount();
  return os;
}

}  // namespace aifo::data
