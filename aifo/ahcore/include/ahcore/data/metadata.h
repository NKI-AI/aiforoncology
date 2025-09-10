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
#ifndef AIFO_AHCORE_INCLUDE_AHCORE_DATA_METADATA_H_
#define AIFO_AHCORE_INCLUDE_AHCORE_DATA_METADATA_H_

#include <filesystem>
#include <fstream>
#include <iosfwd>  // Include for std::ostream forward declaration
#include <map>
#include <memory>
#include <optional>
#include <sstream>
#include <stdexcept>
#include <string>
#include <string_view>
#include <utility>
#include <variant>
#include <vector>

#include <cereal/archives/binary.hpp>
#include <cereal/types/string.hpp>
#include <cereal/types/unordered_map.hpp>
#include <cereal/types/variant.hpp>
#include <cereal/types/vector.hpp>

#include "absl/status/status.h"
#include "absl/status/statusor.h"
#include "ahcore/data/compression.h"
#include "aifocore/tiling/grid.h"
#include "aifocore/utilities/fmt.h"
#include "dlup/slide_geometry.h"

namespace fs = std::filesystem;

namespace cereal {

/**
 * @brief Saves a std::filesystem::path to a Cereal archive.
 * @param archive The archive to save to.
 * @param path The path to save.
 */
template <class Archive>
void save(Archive& archive, const std::filesystem::path& path) {
  archive(path.string());
}

/**
 * @brief Loads a std::filesystem::path from a Cereal archive.
 * @param archive The archive to load from.
 * @param path The path to load.
 */
template <class Archive>
void load(Archive& archive, std::filesystem::path& path) {
  std::string temp;
  archive(temp);
  path = std::filesystem::path(temp);
}

}  // namespace cereal

namespace aifo::data::utils {

/**
 * @brief Opens a file for reading in binary mode.
 * @param filename The name of the file to open.
 * @return A valid input file stream.
 * @throws std::runtime_error If the file does not exist or cannot be opened.
 */
inline std::ifstream OpenInputFile(const fs::path& filename) {
  if (!fs::exists(filename)) {
    throw std::runtime_error("File does not exist: " + filename.string());
  }

  std::ifstream is(filename, std::ios::binary);
  if (!is) {
    throw std::runtime_error("Failed to open file for reading: " +
                             filename.string());
  }
  return is;
}

/**
 * @brief Opens a file for writing in binary mode.
 * @param filename The name of the file to open.
 * @return A valid output file stream.
 */
inline std::ofstream OpenOutputFile(const fs::path& filename) {
  std::ofstream os(filename, std::ios::binary);
  if (!os) {
    throw std::ios::failure("Failed to open file for writing: " +
                            filename.string());
  }
  return os;
}

}  // namespace aifo::data::utils

namespace MetadataKeys {
constexpr auto Size = "size";
constexpr auto TileSize = "tile_size";
constexpr auto TileOverlap = "tile_overlap";
constexpr auto NumChannels = "num_channels";
constexpr auto NumOutputChannels = "num_output_channels";
constexpr auto Mpp = "mpp";
constexpr auto PixelFormat = "pixel_format";
constexpr auto Interpretation = "interpretation";
constexpr auto OutputPixelFormat = "output_pixel_format";
constexpr auto OutputInterpretation = "output_interpretation";
constexpr auto GridOrder = "grid_order";
constexpr auto TileIndices = "tile_indices";
constexpr auto Geometry = "geometry";
}  // namespace MetadataKeys

namespace aifo::data {

using aifocore::Size;

/**
 * @typedef MetadataValue
 * @brief Represents a variant of supported metadata types.
 */
using MetadataValue =
    std::variant<std::string, int, double, float, std::filesystem::path,
                 std::vector<std::string>, std::vector<int>,
                 std::vector<double>, std::vector<float>, Size<int, 2>,
                 aifocore::tiling::GridOrder, dlup::SlideGeometry>;

/**
 * @typedef MetadataMap
 * @brief Represents a map of metadata keys to their associated values.
 */
using MetadataMap =
    std::map<std::string,
             MetadataValue>;  // map makes it a bit easier to compare

/**
 * @class Metadata
 * @brief Represents metadata as a map of key-value pairs.
 */
class Metadata : public std::enable_shared_from_this<Metadata> {
 public:
  Metadata(Metadata&& other) noexcept = default;
  Metadata& operator=(Metadata&& other) noexcept = default;

  /**
   * @brief Factory method to create a shared pointer to a Metadata object.
   * @return A shared pointer to a new Metadata instance.
   */
  static std::shared_ptr<Metadata> Create() {
    return std::shared_ptr<Metadata>(new Metadata());
  }

  /**
   * @brief Equality comparison operator.
   * @param other The other Metadata instance to compare to.
   * @return True if the Metadata instances are equal, false otherwise.
   */
  bool operator==(const Metadata& other) const {
    return values_ == other.values_;
  }

  /**
   * @brief Inequality comparison operator.
   * @param other The other Metadata instance to compare to.
   * @return True if the Metadata instances are not equal, false otherwise.
   */
  bool operator!=(const Metadata& other) const { return !(*this == other); }

  // Overload for aifocore::Size<int,2>
  std::shared_ptr<Metadata> Set(std::string_view key,
                                const aifocore::Size<int, 2>& value) {
    if (lock_) {
      throw std::runtime_error(
          aifocore::fmt::format("Metadata is locked. Cannot set key: {}", key));
    }
    values_.emplace(std::string(key), value);
    return shared_from_this();
  }

  // Existing generic template Set remains below
  template <typename T>
  std::shared_ptr<Metadata> Set(std::string_view key, T&& value) {
    if (lock_) {
      throw std::runtime_error(
          aifocore::fmt::format("Metadata is locked. Cannot set key: {}", key));
    }

    if constexpr (std::is_same_v<std::decay_t<T>, const char*> ||
                  std::is_same_v<std::decay_t<T>, std::string_view>) {
      values_.emplace(std::string(key), std::string(value));
    } else {
      values_.emplace(std::string(key), std::forward<T>(value));
    }
    return shared_from_this();
  }

  std::shared_ptr<Metadata> SetGeometry(const dlup::SlideGeometry& geometry) {
    return Set(MetadataKeys::Geometry, geometry);
  }

  [[deprecated("Use SetGeometry instead")]] std::shared_ptr<Metadata> SetSize(
      int width, int height) {
    return Set(MetadataKeys::Size, Size<int, 2>{width, height});
  }

  [[deprecated("Use SetGeometry instead")]] std::shared_ptr<Metadata> SetSize(
      Size<int, 2> size) {
    return Set(MetadataKeys::Size, size);
  }

  std::shared_ptr<Metadata> SetTileSize(int width, int height) {
    return Set(MetadataKeys::TileSize, Size<int, 2>{width, height});
  }

  std::shared_ptr<Metadata> SetTileSize(Size<int, 2> size) {
    return Set(MetadataKeys::TileSize, size);
  }

  std::shared_ptr<Metadata> SetTileOverlap(int width, int height) {
    return Set(MetadataKeys::TileOverlap, Size<int, 2>{width, height});
  }

  std::shared_ptr<Metadata> SetTileOverlap(Size<int, 2> size) {
    return Set(MetadataKeys::TileOverlap, size);
  }

  std::shared_ptr<Metadata> SetMpp(float mpp) {
    return Set(MetadataKeys::Mpp, mpp);
  }

  std::shared_ptr<Metadata> SetNumChannels(int num_channels) {
    return Set(MetadataKeys::NumChannels, num_channels);
  }

  std::shared_ptr<Metadata> SetPixelFormat(const int pixel_format) {
    return Set(MetadataKeys::PixelFormat, pixel_format);
  }

  std::shared_ptr<Metadata> SetInterpretation(const int interpretation) {
    return Set(MetadataKeys::Interpretation, interpretation);
  }

  std::shared_ptr<Metadata> SetGridOrder(
      const aifocore::tiling::GridOrder& order) {
    return Set(MetadataKeys::GridOrder, order);
  }

  [[deprecated("Use GetGeometry instead")]] [[nodiscard]] const Size<int, 2>&
  GetSize() const {
    return Get<Size<int, 2>>(MetadataKeys::Size);
  }

  [[nodiscard]] const Size<int, 2>& GetTileSize() const {
    return Get<Size<int, 2>>(MetadataKeys::TileSize);
  }

  [[nodiscard]] const Size<int, 2>& GetTileOverlap() const {
    return Get<Size<int, 2>>(MetadataKeys::TileOverlap);
  }

  [[nodiscard]] float GetMpp() const { return Get<float>(MetadataKeys::Mpp); }

  [[nodiscard]] int GetNumChannels() const {
    return Get<int>(MetadataKeys::NumChannels);
  }

  [[nodiscard]] int GetPixelFormat() const {
    return Get<int>(MetadataKeys::PixelFormat);
  }

  [[nodiscard]] int GetInterpretation() const {
    return Get<int>(MetadataKeys::Interpretation);
  }

  [[nodiscard]] aifocore::tiling::GridOrder GetGridOrder() const {
    return Get<aifocore::tiling::GridOrder>(MetadataKeys::GridOrder);
  }

  template <typename T>
  [[nodiscard]] const T& Get(const std::string& key) const {
    try {
      return std::get<T>(values_.at(key));
    } catch (const std::out_of_range& e) {
      throw std::runtime_error("Key not found: " + key);
    } catch (const std::bad_variant_access& e) {
      throw std::runtime_error("Invalid type for key: " + key);
    }
  }

  template <typename T>
  [[nodiscard]] std::optional<T> GetOptional(const std::string& key) const {
    auto it = values_.find(key);
    if (it != values_.end()) {
      return std::get<T>(it->second);
    }
    return std::nullopt;
  }

  [[nodiscard]] bool HasKey(std::string_view key) const {
    return values_.contains(std::string(key));
  }

  [[nodiscard]] size_t GetCount() const { return values_.size(); }

  [[nodiscard]] const MetadataMap& GetAll() const { return values_; }

  std::shared_ptr<Metadata> Lock() {
    lock_ = true;
    return shared_from_this();
  }

  std::shared_ptr<Metadata> Unlock() {
    lock_ = false;
    return shared_from_this();
  }

  bool IsLocked() const { return lock_; }

  /**
   * @brief Serializes the metadata to a file.
   *
   * This function writes the metadata to the specified file in a binary format.
   *
   * @param filename The name of the file to save the serialized metadata.
   * @throws std::runtime_error If the file cannot be opened or if serialization fails.
   */
  inline void Serialize(const std::string& filename) const {
    try {
      auto os = aifo::data::utils::OpenOutputFile(filename);
      Serialize(os);
    } catch (const std::exception& e) {
      throw std::runtime_error("Serialization failed: " +
                               std::string(e.what()));
    }
  }

  /**
   * @brief Deserializes the metadata from a file.
   *
   * This function loads the metadata from the specified file in a binary format.
   *
   * @param filename The name of the file containing the serialized metadata.
   * @throws std::runtime_error If the file cannot be opened or if deserialization fails.
   */
  inline void Deserialize(const fs::path& filename) {
    auto is = aifo::data::utils::OpenInputFile(filename);
    Deserialize(is);
  }

  /**
   * @brief Loads the metadata from a file.
   *
   * This function loads the metadata directly
   *
   * @param filename The name of the file containing the serialized metadata.
   * @param lock Whether to lock the metadata after loading.
   * @throws std::runtime_error If the file cannot be opened or if loading fails.
   */
  inline static std::shared_ptr<Metadata> LoadFromFile(const fs::path& filename,
                                                       bool lock = true) {
    auto metadata = Metadata::Create();
    metadata->Deserialize(filename);
    if (lock) {
      metadata->Lock();
    }
    return metadata;
  }

  /**
   * @brief Serializes the metadata to an output stream.
   *
   * This function writes the metadata to the given output stream in a binary format.
   * Optionally, the data can be compressed using Zstandard compression.
   *
   * @param os The output stream to write the serialized metadata.
   * @param use_compression If true, the metadata is compressed before being written.
   * @param compression_level The level of compression to use (default is 3).
   * @throws std::runtime_error If serialization or compression fails.
   */
  inline void Serialize(std::ostream& os, bool use_compression = false,
                        int compression_level = 3) const {
    if (use_compression) {
      SerializeCompressed(os, compression_level);
    } else {
      uint8_t is_compressed = 0;  // Uncompressed flag
      os.write(reinterpret_cast<const char*>(&is_compressed),
               sizeof(is_compressed));
      cereal::BinaryOutputArchive archive(os);
      archive(*this);
    }
  }

  /**
   * @brief Deserializes the metadata from an input stream.
   *
   * This function reads and loads the metadata from the given input stream.
   *
   * @param is The input stream to read the serialized metadata from.
   * @throws std::runtime_error If deserialization fails due to invalid or corrupted data.
   */
  inline void Deserialize(std::istream& is) {
    uint8_t is_compressed;
    is.read(reinterpret_cast<char*>(&is_compressed), sizeof(is_compressed));

    if (is_compressed) {
      DeserializeCompressed(is);
    } else {
      cereal::BinaryInputArchive archive(is);
      archive(*this);
    }
  }

  /**
   * @brief Serializes the metadata to an output stream with compression.
   *
   * This function writes the metadata to the given output stream in a compressed binary format.
   * The Zstandard compression algorithm is used for compressing the serialized data.
   *
   * @param os The output stream to write the compressed serialized metadata.
   * @param compression_level The level of compression to use (default is 3).
   * @throws std::runtime_error If serialization or compression fails.
   */
  inline void SerializeCompressed(std::ostream& os,
                                  int compression_level = 3) const {
    uint8_t is_compressed = 1;  // Compressed flag
    os.write(reinterpret_cast<const char*>(&is_compressed),
             sizeof(is_compressed));
    os.write(reinterpret_cast<const char*>(&compression_level),
             sizeof(compression_level));

    // Serialize to a memory buffer
    std::ostringstream temp_stream;
    {
      cereal::BinaryOutputArchive archive(temp_stream);
      archive(*this);
    }

    // Compress the serialized data
    auto comp_or =
        aifo::data::compression::ZstdCompressor::Create(compression_level);
    if (!comp_or.ok()) {
      throw std::runtime_error("Failed to create compressor: " +
                               std::string(comp_or.status().message()));
    }
    auto& compressor = *comp_or;
    std::istringstream input_stream(temp_stream.str());
    if (auto st = compressor.Compress(input_stream, os); !st.ok()) {
      throw std::runtime_error("Compression failed: " +
                               std::string(st.message()));
    }
  }

  /**
   * @brief Deserializes the metadata from an input stream with compression.
   *
   * This function reads and decompresses the metadata from the given input stream.
   * The data is expected to be in a Zstandard-compressed binary format.
   *
   * @param is The input stream to read the compressed serialized metadata from.
   * @throws std::runtime_error If decompression or deserialization fails.
   */
  inline void DeserializeCompressed(std::istream& is) {
    int compression_level;
    is.read(reinterpret_cast<char*>(&compression_level),
            sizeof(compression_level));

    // Decompress into a memory buffer
    std::ostringstream decompressed_stream;
    // 1) Create the decompressor via the factory
    auto decomp_or = aifo::data::compression::ZstdDecompressor::Create();
    if (!decomp_or.ok()) {
      throw std::runtime_error("Failed to create decompressor: " +
                               std::string(decomp_or.status().message()));
    }

    // 2) Extract the decompressor
    auto& decompressor = *decomp_or;

    // 3) Decompress and check for errors
    if (auto st = decompressor.Decompress(is, decompressed_stream); !st.ok()) {
      throw std::runtime_error("Decompression failed: " +
                               std::string(st.message()));
    }
    // Deserialize from the decompressed buffer
    std::istringstream input_stream(decompressed_stream.str());
    cereal::BinaryInputArchive archive(input_stream);
    archive(*this);
  }

  /**
   * @brief Serializes the Metadata instance using Cereal.
   * @tparam Archive The type of the Cereal archive.
   * @param archive The archive to serialize to.
   */
  template <class Archive>
  void serialize(Archive& archive) {
    archive(values_);
  }

  /**
   * @brief Gets the slide geometry if available.
   * 
   * @return The SlideGeometry object
   * @throws std::runtime_error if geometry is not set
   */
  [[nodiscard]] const dlup::SlideGeometry& GetGeometry() const {
    return Get<dlup::SlideGeometry>(MetadataKeys::Geometry);
  }

  /**
   * @brief Gets the slide geometry if available, or returns nullptr.
   * 
   * @return Optional SlideGeometry object
   */
  [[nodiscard]] std::optional<dlup::SlideGeometry> GetOptionalGeometry() const {
    return GetOptional<dlup::SlideGeometry>(MetadataKeys::Geometry);
  }

 private:
  MetadataMap values_;   ///< Stores metadata as key-value pairs.
  bool lock_;            ///< Indicates whether metadata is locked.
  Metadata() = default;  ///< Enforce factory method for creation.

  explicit Metadata(const fs::path& filename) {
    Deserialize(filename);
  }  ///< Factory method for creation.

  explicit Metadata(std::istream& stream) {
    Deserialize(stream);
  }  ///< Enforce factory method for creation.
};

/**
 * @brief Stream insertion operator for Metadata.
 *
 * @param os The output stream.
 * @param metadata The Metadata object to print.
 * @return The output stream.
 */
std::ostream& operator<<(std::ostream& os, const Metadata& metadata);

}  // namespace aifo::data

#endif  // AIFO_AHCORE_INCLUDE_AHCORE_DATA_METADATA_H_
