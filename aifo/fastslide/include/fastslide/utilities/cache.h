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

#ifndef AIFO_FASTSLIDE_INCLUDE_FASTSLIDE_UTILITIES_CACHE_H_
#define AIFO_FASTSLIDE_INCLUDE_FASTSLIDE_UTILITIES_CACHE_H_

#include <chrono>
#include <list>
#include <memory>
#include <mutex>
#include <string>
#include <unordered_map>
#include <utility>
#include <vector>

#include "absl/status/status.h"
#include "absl/status/statusor.h"
#include "aifocore/concepts/numeric.h"

namespace fastslide {

/// @brief Cache key for tiles
///
/// This structure uniquely identifies a tile within a slide file at a specific
/// level and position. It's used as the key for the tile cache to ensure
/// efficient lookup and avoid duplicate storage.
struct TileKey {
  std::string filename;  ///< Path to the slide file
  uint16_t level;        ///< Pyramid level (0 = highest resolution)
  uint32_t tile_x;       ///< Tile X coordinate in the grid
  uint32_t tile_y;       ///< Tile Y coordinate in the grid

  /// @brief Equality comparison operator
  /// @param other Other TileKey to compare with
  /// @return true if all fields match, false otherwise
  bool operator==(const TileKey& other) const {
    return filename == other.filename && level == other.level &&
           tile_x == other.tile_x && tile_y == other.tile_y;
  }
};

/// @brief Hash function for TileKey
///
/// Provides a high-quality hash function for TileKey that minimizes collisions
struct TileKeyHash {
  /// @brief Compute hash value for a TileKey
  /// @param key The TileKey to hash
  /// @return Hash value for the key
  std::size_t operator()(const TileKey& key) const {
    std::size_t hash1 = std::hash<std::string>{}(key.filename);
    std::size_t hash2 = std::hash<uint16_t>{}(key.level);
    std::size_t hash3 = std::hash<uint32_t>{}(key.tile_x);
    std::size_t hash4 = std::hash<uint32_t>{}(key.tile_y);

    // Use a better hash combination method than simple XOR
    // This reduces clustering and improves distribution
    std::size_t result = hash1;
    result = result * 31 + hash2;
    result = result * 31 + hash3;
    result = result * 31 + hash4;

    return result;
  }
};

/// @brief Cached tile data
///
/// Contains the actual tile image data along with metadata about dimensions,
/// channels, and access time for LRU management.
struct CachedTile {
  std::vector<uint8_t> data;         ///< Raw tile image data
  aifocore::Size<uint32_t, 2> size;  ///< Tile size in pixels
  uint32_t channels;  ///< Number of color channels (e.g., 3 for RGB)
  std::chrono::steady_clock::time_point
      access_time;  ///< Last access time for LRU

  /// @brief Constructor for CachedTile
  /// @param tile_data Raw tile image data (moved for efficiency)
  /// @param size Tile size in pixels
  /// @param c Number of color channels
  CachedTile(std::vector<uint8_t>&& tile_data, aifocore::Size<uint32_t, 2> size,
             uint32_t channels)
      : data(std::move(tile_data)),
        size(size),
        channels(channels),
        access_time(std::chrono::steady_clock::now()) {}
};

/// @brief Thread-safe LRU tile cache
///
/// A high-performance, thread-safe implementation of an LRU (Least Recently
/// Used) cache for tile data. The cache uses a combination of hash map for O(1)
/// lookup and doubly-linked list for O(1) LRU operations.
///
/// Thread safety is achieved through a single mutex that protects all internal
/// data structures. While this provides coarse-grained locking, it's sufficient
/// for most use cases where cache operations are relatively infrequent compared
/// to the computational cost of loading tiles.
///
/// The cache automatically evicts the least recently used tiles when capacity
/// is exceeded, ensuring memory usage stays bounded.
class TileCache {
 public:
  /// @brief Create a new TileCache instance (preferred method)
  /// @param capacity Maximum number of tiles to cache
  /// @return StatusOr containing the TileCache instance, or an error status
  static absl::StatusOr<TileCache> Create(size_t capacity = 1000);

  /// @brief Create a shared TileCache instance for Python bindings
  /// @param capacity Maximum number of tiles to cache
  /// @return StatusOr containing shared pointer to TileCache instance
  static absl::StatusOr<std::shared_ptr<TileCache>> CreateShared(
      size_t capacity = 1000);

  /// @brief Constructor (internal use - prefer Create() method)
  /// @param capacity Maximum number of tiles to cache
  /// @note Use Create() method for better error handling
  explicit TileCache(size_t capacity);

  /// @brief Destructor
  ~TileCache() = default;

  // Non-copyable but movable
  TileCache(const TileCache&) = delete;
  TileCache& operator=(const TileCache&) = delete;

  TileCache(TileCache&&) = default;
  TileCache& operator=(TileCache&&) = default;

  /// @brief Get a tile from cache
  ///
  /// Retrieves a tile from the cache if it exists. If found, updates the
  /// access time and moves the tile to the front of the LRU list.
  ///
  /// @param key Tile key identifying the requested tile
  /// @return Cached tile if found, nullptr otherwise
  /// @note This method is thread-safe
  std::shared_ptr<CachedTile> Get(const TileKey& key);

  /// @brief Put a tile in cache
  ///
  /// Adds a tile to the cache. If the tile already exists, updates it and
  /// moves it to the front of the LRU list. If the cache is at capacity,
  /// evicts the least recently used tile first.
  ///
  /// @param key Tile key identifying the tile
  /// @param tile Shared pointer to the tile data
  /// @note This method is thread-safe
  void Put(const TileKey& key, std::shared_ptr<CachedTile> tile);

  /// @brief Clear all cached tiles
  ///
  /// Removes all tiles from the cache and resets statistics.
  ///
  /// @note This method is thread-safe
  void Clear();

  /// @brief Get cache statistics
  ///
  /// Returns current cache statistics including capacity, size, hit/miss
  /// counts, and hit ratio.
  struct Stats {
    size_t capacity;            ///< Maximum cache capacity
    size_t size;                ///< Current number of cached tiles
    size_t hits;                ///< Number of cache hits
    size_t misses;              ///< Number of cache misses
    double hit_ratio;           ///< Hit ratio (hits / (hits + misses))
    size_t memory_usage_bytes;  ///< Total memory usage in bytes
  };

  /// @brief Get current cache statistics
  /// @return Current cache statistics
  /// @note This method is thread-safe
  Stats GetStats() const;

  /// @brief Get current cache capacity
  /// @return Maximum number of tiles that can be cached
  /// @note This method is thread-safe
  size_t GetCapacity() const;

  /// @brief Get current cache size
  /// @return Current number of tiles in cache
  /// @note This method is thread-safe
  size_t GetSize() const;

  /// @brief Get total memory usage in bytes
  /// @return Total memory usage of all cached tiles in bytes
  /// @note This method is thread-safe
  size_t GetMemoryUsage() const;

  /// @brief Set cache capacity
  ///
  /// Creates a new cache with the specified capacity, replacing the existing
  /// cache. All existing cached tiles will be lost.
  ///
  /// @param capacity New cache capacity
  /// @return Status indicating success or failure
  /// @note This method is thread-safe
  absl::Status SetCapacity(size_t capacity);

 private:
  /// @brief Internal cache entry structure
  ///
  /// Combines the tile data with an iterator to its position in the LRU list
  /// for efficient LRU management.
  struct CacheEntry {
    TileKey key;                          ///< Cache key
    std::shared_ptr<CachedTile> tile;     ///< Cached tile data
    std::list<TileKey>::iterator lru_it;  ///< Iterator to LRU list position
  };

  /// @brief Evict the least recently used tile
  ///
  /// Removes the tile at the back of the LRU list and from the cache map.
  /// Called when the cache is at capacity and a new tile needs to be added.
  ///
  /// @note Assumes mutex is already held
  void EvictLru();

  size_t capacity_;           ///< Maximum cache capacity
  mutable std::mutex mutex_;  ///< Mutex for thread safety
  std::unordered_map<TileKey, CacheEntry, TileKeyHash>
      cache_;                    ///< Main cache storage
  std::list<TileKey> lru_list_;  ///< LRU ordering list

  // Statistics (mutable to allow const access)
  mutable size_t hits_;    ///< Number of cache hits
  mutable size_t misses_;  ///< Number of cache misses
};

/// @brief Global tile cache instance
///
/// Provides a singleton global cache instance that can be shared across
/// all slide readers. This is useful for applications that want to share
/// cached tiles between multiple readers or maintain a global cache policy.
///
/// The global cache is thread-safe and can be safely accessed from multiple
/// threads simultaneously.
class GlobalTileCache {
 public:
  /// @brief Get singleton instance
  ///
  /// Returns the global singleton instance of the tile cache. The instance
  /// is created on first access and destroyed when the program exits.
  ///
  /// @return Reference to the global cache instance
  /// @note This method is thread-safe
  static GlobalTileCache& Instance();

  /// @brief Get the tile cache
  ///
  /// Returns a reference to the underlying tile cache. This allows direct
  /// access to all cache operations.
  ///
  /// @return Reference to the tile cache
  /// @note This method is thread-safe
  TileCache& GetCache();

  /// @brief Get the tile cache (const version)
  ///
  /// Returns a const reference to the underlying tile cache for read-only
  /// operations.
  ///
  /// @return Const reference to the tile cache
  /// @note This method is thread-safe
  const TileCache& GetCache() const;

  /// @brief Set cache capacity
  ///
  /// Creates a new cache with the specified capacity, replacing the existing
  /// cache. All existing cached tiles will be lost.
  ///
  /// @param capacity New cache capacity
  /// @return Status indicating success or failure
  /// @note This method is thread-safe
  absl::Status SetCapacity(size_t capacity);

  /// @brief Get current cache capacity
  /// @return Current cache capacity
  /// @note This method is thread-safe
  size_t GetCapacity() const;

 private:
  /// @brief Private constructor for singleton
  GlobalTileCache();

  mutable std::mutex mutex_;          ///< Mutex for thread safety
  std::unique_ptr<TileCache> cache_;  ///< The actual cache instance
};

}  // namespace fastslide

#endif  // AIFO_FASTSLIDE_INCLUDE_FASTSLIDE_UTILITIES_CACHE_H_
