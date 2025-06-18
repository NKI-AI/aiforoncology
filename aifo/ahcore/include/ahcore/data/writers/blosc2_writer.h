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
#ifndef AIFO_AHCORE_INCLUDE_AHCORE_DATA_WRITERS_BLOSC2_WRITER_H_
#define AIFO_AHCORE_INCLUDE_AHCORE_DATA_WRITERS_BLOSC2_WRITER_H_

#include <blosc2.h>

#include <filesystem>
#include <memory>
#include <string>
#include <vector>

#include "ahcore/data/writers/base_writer.h"
#include "ahcore/utilities/blosc2.h"

namespace aifo::data::writers {

namespace fs = std::filesystem;

/**
 * @class Blosc2TileWriter
 * @brief A writer for storing compressed tiles and metadata in a single file using Blosc2.
 */
class Blosc2TileWriter : public Writer {
 public:
  /**
   * @brief Constructs a `Blosc2TileWriter`.
   *
   * @param output_file The file where compressed tiles and metadata will be stored.
   * @param metadata A shared pointer to metadata (optional).
   */
  inline Blosc2TileWriter(const fs::path& output_file,
                          std::shared_ptr<Metadata> metadata = nullptr)
      : Writer(), output_file_(output_file), is_open_(false), tile_count_(0) {
    metadata_ = metadata ? metadata : Metadata::Create();
    metadata_->Validate();
    metadata_->Lock();
  }

  /**
   * @brief Destructor for `Blosc2TileWriter`.
   */
  ~Blosc2TileWriter() override;

  /**
   * @brief Opens the writer, creating or truncating the output file.
   */
  void Open() override;

  /**
   * @brief Closes the writer and finalizes the file.
   */
  void Close() override;

  /**
   * @brief Compresses and writes a tile to the file.
   *
   * @param index The index of the tile.
   * @param tile The tile data as a `vips::VImage`.
   */
  void WriteTile(int index, const vips::VImage& tile) override;

  /**
   * @brief Writes metadata to the file.
   */
  void WriteMetadata() override;

  /**
   * @brief Removes the output file.
   */
  void Unlink() override;

 private:
  std::string output_file_;  ///< The file where tiles and metadata are stored.
  std::ofstream file_stream_;  ///< Stream for writing to the file.
  bool is_open_;               ///< Indicates whether the writer is open.
  int tile_count_;             ///< The number of tiles written.

  Blosc2Context blosc2_context_;  ///< RAII-managed Blosc2 context.

  /**
   * @brief Compresses data using Blosc2.
   *
   * @param input Pointer to the input data.
   * @param input_size Size of the input data in bytes.
   * @param output Pointer to the output buffer.
   * @param output_size Size of the output buffer in bytes.
   * @return Size of the compressed data, or negative value on error.
   */
  int CompressData(const void* input, size_t input_size, void* output,
                   size_t output_size);
};

inline Blosc2TileWriter::~Blosc2TileWriter() {
  if (is_open_) {
    Close();
  }
}

inline void Blosc2TileWriter::Open() {
  file_stream_.open(output_file_, std::ios::binary | std::ios::trunc);
  if (!file_stream_) {
    throw std::runtime_error("Failed to open file for writing: " +
                             output_file_);
  }
  is_open_ = true;
}

inline void Blosc2TileWriter::Close() {
  if (is_open_) {
    file_stream_.close();
    is_open_ = false;
  }
}

inline void Blosc2TileWriter::WriteTile(int index, const vips::VImage& tile) {
  if (!is_open_) {
    throw std::runtime_error("Writer is not open");
  }

  size_t tile_size = tile.width() * tile.height() * tile.bands();
  std::vector<uint8_t> compressed_data(tile_size + BLOSC2_MAX_OVERHEAD);

  int compressed_size = CompressData(
      tile.data(), tile_size, compressed_data.data(), compressed_data.size());

  if (compressed_size < 0) {
    throw std::runtime_error("Blosc2 compression failed for tile " +
                             std::to_string(index));
  }

  // Write the tile index, compressed size, and compressed data
  file_stream_.write(reinterpret_cast<const char*>(&index), sizeof(index));
  file_stream_.write(reinterpret_cast<const char*>(&compressed_size),
                     sizeof(compressed_size));
  file_stream_.write(reinterpret_cast<const char*>(compressed_data.data()),
                     compressed_size);

  tile_count_++;
}

inline void Blosc2TileWriter::WriteMetadata() {
  if (!is_open_) {
    throw std::runtime_error("Writer is not open");
  }

  std::string serialized_metadata = metadata_->SerializeToString();
  std::vector<uint8_t> compressed_metadata(serialized_metadata.size() +
                                           BLOSC2_MAX_OVERHEAD);
  int compressed_size =
      CompressData(serialized_metadata.data(), serialized_metadata.size(),
                   compressed_metadata.data(), compressed_metadata.size());

  if (compressed_size < 0) {
    throw std::runtime_error("Blosc2 compression failed for metadata");
  }

  // Write a special marker for metadata (-1 as the tile index)
  int metadata_marker = -1;
  file_stream_.write(reinterpret_cast<const char*>(&metadata_marker),
                     sizeof(metadata_marker));
  file_stream_.write(reinterpret_cast<const char*>(&compressed_size),
                     sizeof(compressed_size));
  file_stream_.write(reinterpret_cast<const char*>(compressed_metadata.data()),
                     compressed_size);
}

inline void Blosc2TileWriter::Unlink() {
  fs::remove(output_file_);
}

inline int Blosc2TileWriter::CompressData(const void* input, size_t input_size,
                                          void* output, size_t output_size) {
  return blosc2_compress(5, BLOSC_NOSHUFFLE, sizeof(uint8_t), input, input_size,
                         output, output_size);
}

}  // namespace aifo::data::writers

#endif  // AIFO_AHCORE_INCLUDE_AHCORE_DATA_WRITERS_BLOSC2_WRITER_H_
