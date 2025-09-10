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
#include <fstream>
#include <iomanip>
#include <iostream>
#include <iterator>
#include <string>
#include <vector>
#include "ahcore/data/metadata.h"
#include "dlup/slide_geometry.h"

// Hex dump utility
void HexDump(const std::vector<uint8_t>& data, size_t max_lines = 8) {
  constexpr size_t bytes_per_line = 16;  // Number of bytes per line
  size_t lines = 0;

  std::cout << "Hex Dump (first " << max_lines << " lines):\n";
  std::cout << std::string(76, '-') << "\n";  // Header line

  for (size_t i = 0; i < data.size(); i += bytes_per_line) {
    if (lines++ >= max_lines) {
      std::cout << "... (output truncated)\n";
      break;
    }

    // Print offset
    std::cout << std::setw(6) << std::setfill('0') << std::hex << i << ": ";

    // Print hex values
    for (size_t j = 0; j < bytes_per_line; ++j) {
      if (i + j < data.size()) {
        std::cout << std::setw(2) << std::setfill('0')
                  << static_cast<int>(data[i + j]) << " ";
      } else {
        std::cout << "   ";  // Padding for missing bytes
      }
    }

    // Separate hex values from ASCII
    std::cout << " |";

    // Print ASCII representation
    for (size_t j = 0; j < bytes_per_line; ++j) {
      if (i + j < data.size()) {
        unsigned char c = data[i + j];
        std::cout << (std::isprint(c) ? static_cast<char>(c) : '.');
      } else {
        std::cout << " ";
      }
    }
    std::cout << "|\n";
  }

  std::cout << std::string(76, '-') << "\n\n";  // Footer line
  std::cout << std::dec;                        // Reset to decimal output
}

int main() {
  // Small example
  auto small_metadata = aifo::data::Metadata::Create();
  small_metadata
      ->SetGeometry(dlup::SlideGeometry{{400, 300}, {0, 0}, {400, 300}})
      ->SetTileSize(100, 100)
      ->SetTileOverlap(20, 20)
      ->SetMpp(0.25f)
      ->Set("format", "jpeg")
      ->Set("compression", 75);

  std::cout << "Small Example (Uncompressed):\n";
  std::cout << *small_metadata << "\n\n";

  // Serialize to a file
  std::string small_filename = "small_metadata.bin";
  small_metadata->Serialize(small_filename);
  std::cout << "Small metadata serialized to: " << small_filename << "\n";

  // Read serialized data for HexDump
  std::ifstream small_file_stream(small_filename, std::ios::binary);
  std::vector<uint8_t> small_file_data(
      (std::istreambuf_iterator<char>(small_file_stream)),
      std::istreambuf_iterator<char>());
  small_file_stream.close();

  // Show HexDump
  HexDump(small_file_data);

  // Deserialize from the file
  auto small_deserialized_metadata =
      aifo::data::Metadata::LoadFromFile(small_filename);
  std::cout << "\nDeserialized Small Metadata:\n";
  std::cout << *small_deserialized_metadata << "\n\n";

  // Validate
  if (*small_metadata == *small_deserialized_metadata) {
    std::cout << "Validation: Small metadata deserialized correctly.\n";
  } else {
    std::cout << "Validation: Small metadata deserialization failed.\n";
  }

  // Large example
  auto large_metadata = aifo::data::Metadata::Create();
  for (int i = 0; i < 10000; ++i) {
    large_metadata->Set("string_field_" + std::to_string(i),
                        "example_string_" + std::to_string(i));
    large_metadata->Set("int_field_" + std::to_string(i), i);
    large_metadata->Set("float_field_" + std::to_string(i), i * 0.1f);
  }

  // Serialize large metadata uncompressed to a file
  std::string large_uncompressed_filename = "large_metadata_uncompressed.bin";
  large_metadata->Serialize(large_uncompressed_filename);
  std::cout << "\nLarge metadata (uncompressed) serialized to: "
            << large_uncompressed_filename << "\n";

  // Serialize large metadata compressed to a file
  std::string large_compressed_filename = "large_metadata_compressed.bin";
  std::ofstream compressed_file(large_compressed_filename, std::ios::binary);
  large_metadata->Serialize(compressed_file, true, 5);
  compressed_file.close();
  std::cout << "Large metadata (compressed) serialized to: "
            << large_compressed_filename << "\n";

  // Deserialize large metadata from compressed file
  auto large_deserialized_metadata =
      aifo::data::Metadata::LoadFromFile(large_compressed_filename);

  std::cout << "\nValidation for Large Metadata:\n";
  if (*large_metadata == *large_deserialized_metadata) {
    std::cout << "Validation: Large metadata deserialized correctly.\n";
  } else {
    std::cout << "Validation: Large metadata deserialization failed.\n";
  }

  // Compare file sizes
  auto GetFileSize = [](const std::string& filename) -> size_t {
    std::ifstream file(filename, std::ios::binary | std::ios::ate);
    return file.tellg();
  };

  size_t uncompressed_size = GetFileSize(large_uncompressed_filename);
  size_t compressed_size = GetFileSize(large_compressed_filename);

  std::cout << "\nFile Size Comparison:\n";
  std::cout << "Uncompressed file size: " << uncompressed_size << " bytes\n";
  std::cout << "Compressed file size: " << compressed_size << " bytes\n";

  double compression_ratio =
      static_cast<double>(compressed_size) / uncompressed_size;
  std::cout << "Compression ratio: " << compression_ratio << "\n";

  return 0;
}
