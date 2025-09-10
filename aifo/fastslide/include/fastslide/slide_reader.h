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

#ifndef AIFO_FASTSLIDE_INCLUDE_FASTSLIDE_SLIDE_READER_H_
#define AIFO_FASTSLIDE_INCLUDE_FASTSLIDE_SLIDE_READER_H_

#include <cstdint>
#include <functional>
#include <map>
#include <memory>
#include <optional>
#include <ostream>
#include <shared_mutex>
#include <string>
#include <string_view>
#include <utility>
#include <variant>
#include <vector>

#include "absl/status/statusor.h"
#include "fastslide/image.h"
#include "fastslide/utilities/cache.h"
#include "fastslide/utilities/colors.h"

namespace fastslide {

/// @brief Enhanced metadata container with printing capabilities
class Metadata {
 public:
  /// @brief Value type for metadata entries
  using Value = std::variant<std::string, size_t, double>;

  /// @brief Underlying container type
  using Container = std::map<std::string, Value>;

  /// @brief Iterator types
  using iterator = Container::iterator;
  using const_iterator = Container::const_iterator;
  using value_type = Container::value_type;
  using key_type = Container::key_type;
  using mapped_type = Container::mapped_type;

  /// @brief Default constructor
  Metadata() = default;

  /// @brief Constructor from initializer list
  /// @param init Initializer list of key-value pairs
  Metadata(std::initializer_list<value_type> init) : data_(init) {}

  /// @brief Constructor from std::map (for backward compatibility)
  /// @param data Map to initialize from
  explicit Metadata(const Container& data) : data_(data) {}

  /// @brief Constructor from std::map (move version)
  /// @param data Map to move from
  explicit Metadata(Container&& data) : data_(std::move(data)) {}

  // Map-like interface for full backward compatibility

  /// @brief Access element by key
  /// @param key Key to access
  /// @return Reference to the value
  Value& operator[](const std::string& key) { return data_[key]; }

  /// @brief Access element by key (const)
  /// @param key Key to access
  /// @return Const reference to the value
  const Value& at(const std::string& key) const { return data_.at(key); }

  /// @brief Access element by key
  /// @param key Key to access
  /// @return Reference to the value
  Value& at(const std::string& key) { return data_.at(key); }

  /// @brief Insert or assign key-value pair
  /// @param key Key to insert
  /// @param value Value to insert
  /// @return Pair of iterator and bool indicating insertion
  template <typename T>
  std::pair<iterator, bool> insert_or_assign(const std::string& key,
                                             T&& value) {
    return data_.insert_or_assign(key, std::forward<T>(value));
  }

  /// @brief Insert key-value pair
  /// @param value Key-value pair to insert
  /// @return Pair of iterator and bool indicating insertion
  std::pair<iterator, bool> insert(const value_type& value) {
    return data_.insert(value);
  }

  /// @brief Find element by key
  /// @param key Key to find
  /// @return Iterator to element or end()
  iterator find(const std::string& key) { return data_.find(key); }

  /// @brief Find element by key (const)
  /// @param key Key to find
  /// @return Const iterator to element or end()
  const_iterator find(const std::string& key) const { return data_.find(key); }

  /// @brief Check if key exists
  /// @param key Key to check
  /// @return True if key exists
  bool contains(const std::string& key) const { return data_.contains(key); }

  /// @brief Get number of elements
  /// @return Number of elements
  size_t size() const noexcept { return data_.size(); }

  /// @brief Check if empty
  /// @return True if empty
  bool empty() const noexcept { return data_.empty(); }

  /// @brief Clear all elements
  void clear() noexcept { data_.clear(); }

  /// @brief Begin iterator
  /// @return Iterator to beginning
  iterator begin() { return data_.begin(); }

  /// @brief Begin iterator (const)
  /// @return Const iterator to beginning
  const_iterator begin() const { return data_.begin(); }

  /// @brief End iterator
  /// @return Iterator to end
  iterator end() { return data_.end(); }

  /// @brief End iterator (const)
  /// @return Const iterator to end
  const_iterator end() const { return data_.end(); }

  /// @brief Const begin iterator
  /// @return Const iterator to beginning
  const_iterator cbegin() const { return data_.cbegin(); }

  /// @brief Const end iterator
  /// @return Const iterator to end
  const_iterator cend() const { return data_.cend(); }

  // Enhanced functionality

  /// @brief Get value as string with optional default
  /// @param key Key to look up
  /// @param default_value Default value if key not found
  /// @return String value or default
  std::string GetString(const std::string& key,
                        const std::string& default_value = "") const {
    auto it = data_.find(key);
    if (it == data_.end()) {
      return default_value;
    }

    return std::visit(
        [&default_value](const auto& value) -> std::string {
          if constexpr (std::is_same_v<std::decay_t<decltype(value)>,
                                       std::string>) {
            return value;
          } else if constexpr (std::is_same_v<std::decay_t<decltype(value)>,
                                              size_t>) {
            return std::to_string(value);
          } else if constexpr (std::is_same_v<std::decay_t<decltype(value)>,
                                              double>) {
            return std::to_string(value);
          } else {
            return default_value;
          }
        },
        it->second);
  }

  /// @brief Get value as double with optional default
  /// @param key Key to look up
  /// @param default_value Default value if key not found or cannot convert
  /// @return Double value or default
  double GetDouble(const std::string& key, double default_value = 0.0) const {
    auto it = data_.find(key);
    if (it == data_.end()) {
      return default_value;
    }

    return std::visit(
        [default_value](const auto& value) -> double {
          if constexpr (std::is_same_v<std::decay_t<decltype(value)>, double>) {
            return value;
          } else if constexpr (std::is_same_v<std::decay_t<decltype(value)>,
                                              size_t>) {
            return static_cast<double>(value);
          } else if constexpr (std::is_same_v<std::decay_t<decltype(value)>,
                                              std::string>) {
            try {
              return std::stod(value);
            } catch (...) {
              return default_value;
            }
          } else {
            return default_value;
          }
        },
        it->second);
  }

  /// @brief Get value as size_t with optional default
  /// @param key Key to look up
  /// @param default_value Default value if key not found or cannot convert
  /// @return size_t value or default
  size_t GetSize(const std::string& key, size_t default_value = 0) const {
    auto it = data_.find(key);
    if (it == data_.end()) {
      return default_value;
    }

    return std::visit(
        [default_value](const auto& value) -> size_t {
          if constexpr (std::is_same_v<std::decay_t<decltype(value)>, size_t>) {
            return value;
          } else if constexpr (std::is_same_v<std::decay_t<decltype(value)>,
                                              double>) {
            return static_cast<size_t>(value);
          } else if constexpr (std::is_same_v<std::decay_t<decltype(value)>,
                                              std::string>) {
            try {
              return std::stoull(value);
            } catch (...) {
              return default_value;
            }
          } else {
            return default_value;
          }
        },
        it->second);
  }

  /// @brief Convert to formatted string
  /// @param indent Number of spaces to indent each line
  /// @return Formatted string representation
  std::string ToString(size_t indent = 0) const {
    std::string result;
    std::string indent_str(indent, ' ');

    for (const auto& [key, value] : data_) {
      result += indent_str + key + ": ";

      std::visit(
          [&result](const auto& v) {
            if constexpr (std::is_same_v<std::decay_t<decltype(v)>,
                                         std::string>) {
              result += "\"" + v + "\"";
            } else {
              result += std::to_string(v);
            }
          },
          value);

      result += "\n";
    }

    return result;
  }

  /// @brief Get underlying container (for compatibility)
  /// @return Reference to underlying container
  const Container& GetContainer() const { return data_; }

  /// @brief Get underlying container (for compatibility)
  /// @return Reference to underlying container
  Container& GetContainer() { return data_; }

  /// @brief Implicit conversion to underlying container for backward
  /// compatibility
  /// @return Reference to underlying container
  operator const Container&() const { return data_; }

  /// @brief Implicit conversion to underlying container for backward
  /// compatibility
  /// @return Reference to underlying container
  operator Container&() { return data_; }

 private:
  Container data_;  ///< Underlying metadata storage
};

/// @brief Stream output operator for Metadata
/// @param os Output stream
/// @param metadata Metadata to output
/// @return Output stream reference
inline std::ostream& operator<<(std::ostream& os, const Metadata& metadata) {
  os << "Metadata {\n";
  for (const auto& [key, value] : metadata) {
    os << "  " << key << ": ";

    std::visit(
        [&os](const auto& v) {
          if constexpr (std::is_same_v<std::decay_t<decltype(v)>,
                                       std::string>) {
            os << "\"" << v << "\"";
          } else {
            os << v;
          }
        },
        value);

    os << "\n";
  }
  os << "}";
  return os;
}

// ImageCoordinate and ImageDimensions are now defined in fastslide/image.h

/// @brief Channel metadata structure for microscopy channels
struct ChannelMetadata {
  std::string name;        ///< Channel name (e.g., "DAPI", "ATTO 550")
  std::string biomarker;   ///< Biomarker information (e.g., "Ki-67", "CD20")
  ColorRGB color;          ///< Display color for visualization
  uint32_t exposure_time;  ///< Exposure time in microseconds
  uint32_t signal_units;   ///< Signal units (bit depth related)
  std::map<std::string, std::string>
      additional;  ///< Format-specific additional metadata

  /// @brief Default constructor
  ChannelMetadata() : exposure_time(0), signal_units(0) {}

  /// @brief Constructor with basic fields
  ChannelMetadata(std::string name, std::string biomarker, ColorRGB color,
                  uint32_t exposure_time = 0, uint32_t signal_units = 0)
      : name(std::move(name)),
        biomarker(std::move(biomarker)),
        color(color),
        exposure_time(exposure_time),
        signal_units(signal_units) {}
};

/// @brief Pyramid level metadata
struct LevelInfo {
  ImageDimensions dimensions;  ///< Level dimensions
  double downsample_factor;    ///< Downsample factor relative to level 0

  /// @brief Get number of channels (from global channel metadata)
  [[nodiscard]] size_t GetChannelCount() const noexcept {
    // Note: Channel count should be retrieved from slide's channel metadata
    // This is kept for compatibility but should use
    // slide.GetChannelMetadata().size()
    return 0;
  }
};

/// @brief Physical slide properties
struct SlideProperties {
  aifocore::Size<double, 2> mpp;   ///< Microns per pixel in X, Y direction
  double objective_magnification;  ///< Objective magnification (e.g., 20.0)
  std::string objective_name;      ///< Objective name (e.g., "Plan Apo 20x")
  std::string scanner_model;       ///< Scanner model/manufacturer
  std::optional<std::string> scan_date;  ///< Scan date/time if available
};

/// @brief Region of interest specification
struct RegionSpec {
  ImageCoordinate top_left;  ///< Top-left coordinate
  ImageDimensions size;      ///< Desired region size
  int level;                 ///< Pyramid level (0 = full resolution)

  /// @brief Check if region is valid
  [[nodiscard]] bool IsValid() const noexcept {
    return size[0] > 0 && size[1] > 0 && level >= 0;
  }
};

/// @brief Abstract base class for slide readers
class SlideReader {
 public:
  /// @brief Virtual destructor
  virtual ~SlideReader() = default;

  /// @brief Delete copy constructor and assignment
  SlideReader(const SlideReader&) = delete;
  SlideReader& operator=(const SlideReader&) = delete;

  /// @brief Delete move constructor and assignment
  SlideReader(SlideReader&&) = delete;
  SlideReader& operator=(SlideReader&&) = delete;

  /// @brief Get number of pyramid levels
  /// @return Number of levels (level 0 is full resolution)
  [[nodiscard]] virtual int GetLevelCount() const = 0;

  /// @brief Get level information
  /// @param level Pyramid level
  /// @return Level information or error status
  [[nodiscard]] virtual absl::StatusOr<LevelInfo> GetLevelInfo(
      int level) const = 0;

  /// @brief Get slide physical properties
  /// @return Slide properties
  [[nodiscard]] virtual const SlideProperties& GetProperties() const = 0;

  /// @brief Get channel metadata for all channels
  /// @return Vector of channel metadata (index corresponds to channel index)
  [[nodiscard]] virtual std::vector<ChannelMetadata> GetChannelMetadata()
      const = 0;

  /// @brief Get available associated image names
  /// @return Vector of associated image names (e.g., "thumbnail", "macro")
  [[nodiscard]] virtual std::vector<std::string> GetAssociatedImageNames()
      const = 0;

  /// @brief Get dimensions of an associated image
  /// @param name Associated image name
  /// @return Image dimensions or error status
  [[nodiscard]] virtual absl::StatusOr<ImageDimensions>
  GetAssociatedImageDimensions(std::string_view name) const = 0;

  /// @brief Get best level for a given downsample factor
  /// @param downsample Desired downsample factor
  /// @return Best matching level
  [[nodiscard]] virtual int GetBestLevelForDownsample(double downsample) const;

  /// @brief Read a region from the slide
  /// @param region Region specification
  /// @return Image or error status
  [[nodiscard]] virtual absl::StatusOr<Image> ReadRegion(
      const RegionSpec& region) const = 0;

  /// @brief Read an associated image
  /// @param name Associated image name
  /// @return Image or error status
  [[nodiscard]] virtual absl::StatusOr<Image> ReadAssociatedImage(
      std::string_view name) const = 0;

  /// @brief Get format-specific metadata as key-value pairs
  /// @return Metadata map
  [[nodiscard]] virtual Metadata GetMetadata() const = 0;

  /// @brief Get file format name
  /// @return Format name (e.g., "QPTIFF", "SVS", "NDPI")
  [[nodiscard]] virtual std::string GetFormatName() const = 0;

  /// @brief Get optimal tile size for efficient reading
  /// @return Tile size (width, height) in pixels, or {0, 0} if not tiled
  [[nodiscard]] virtual ImageDimensions GetTileSize() const = 0;

  /// @brief Set tile cache for caching decoded internal tiles
  /// @param cache Shared pointer to tile cache (nullptr to disable caching)
  virtual void SetCache(std::shared_ptr<TileCache> cache) {
    cache_ = std::move(cache);
  }

  /// @brief Get current tile cache
  /// @return Shared pointer to current cache (nullptr if disabled)
  [[nodiscard]] virtual std::shared_ptr<TileCache> GetCache() const {
    return cache_;
  }

  /// @brief Check if caching is enabled
  /// @return True if cache is set and enabled
  [[nodiscard]] virtual bool IsCacheEnabled() const {
    return cache_ != nullptr;
  }

  /// @brief Set which channels are visible during ReadRegion operations
  /// @param channel_indices Vector of channel indices to load
  /// (empty = all channels)
  /// @details Only the specified channels will be loaded and combined in
  /// ReadRegion. This can significantly improve performance for multichannel
  /// formats when only a subset of channels is needed for visualization.
  virtual void SetVisibleChannels(const std::vector<size_t>& channel_indices) {
    visible_channels_ = channel_indices;
  }

  /// @brief Get currently visible channel indices
  /// @return Vector of visible channel indices (empty = all channels visible)
  [[nodiscard]] virtual const std::vector<size_t>& GetVisibleChannels() const {
    return visible_channels_;
  }

  /// @brief Reset to show all channels
  virtual void ShowAllChannels() { visible_channels_.clear(); }

 protected:
  /// @brief Protected constructor (only derived classes can instantiate)
  SlideReader() = default;

  /// @brief Utility function to clamp region to image bounds
  /// @param region Input region specification
  /// @param image_dims Image dimensions to clamp against
  /// @return Clamped region specification
  static RegionSpec ClampRegion(const RegionSpec& region,
                                const ImageDimensions& image_dims);

  /// @brief Channel indices to load (empty = all channels)
  /// @details Protected so derived classes can access for implementing
  /// selective loading
  std::vector<size_t> visible_channels_;

 private:
  /// @brief Optional tile cache for decoded internal tiles
  std::shared_ptr<TileCache> cache_;
};

/// @brief Factory function type for creating slide readers
using SlideReaderFactory =
    std::function<absl::StatusOr<std::unique_ptr<SlideReader>>(
        std::string_view filename)>;

/// @brief Slide reader registry for different formats
class SlideReaderRegistry {
 public:
  /// @brief Get singleton instance
  static SlideReaderRegistry& GetInstance();

  /// @brief Register a reader factory for a file extension
  /// @param extension File extension (e.g., ".qptiff", ".svs")
  /// @param factory Factory function
  void RegisterReader(std::string_view extension, SlideReaderFactory factory);

  /// @brief Create a reader for the given file
  /// @param filename Path to slide file
  /// @return Slide reader instance or error status
  absl::StatusOr<std::unique_ptr<SlideReader>> CreateReader(
      std::string_view filename) const;

  /// @brief Get list of supported extensions
  /// @return Vector of supported file extensions
  [[nodiscard]] std::vector<std::string> GetSupportedExtensions() const;

 private:
  std::map<std::string, SlideReaderFactory> factories_;
  mutable std::shared_mutex
      mutex_;  ///< Thread synchronization for registry access
};

/// @brief Auto-registration helper for slide readers
template <typename ReaderType>
class SlideReaderRegistrar {
 public:
  explicit SlideReaderRegistrar(std::string_view extension) {
    SlideReaderRegistry::GetInstance().RegisterReader(
        extension,
        [](std::string_view filename)
            -> absl::StatusOr<std::unique_ptr<SlideReader>> {
          return std::make_unique<ReaderType>(filename);
        });
  }
};

/// @brief Macro for easy reader registration
#define REGISTER_SLIDE_READER(ReaderClass, Extension)                   \
  static SlideReaderRegistrar<ReaderClass> g_##ReaderClass##_registrar( \
      Extension)

}  // namespace fastslide

#endif  // AIFO_FASTSLIDE_INCLUDE_FASTSLIDE_SLIDE_READER_H_
