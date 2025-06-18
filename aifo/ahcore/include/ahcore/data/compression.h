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
#ifndef AIFO_AHCORE_INCLUDE_AHCORE_DATA_COMPRESSION_H_
#define AIFO_AHCORE_INCLUDE_AHCORE_DATA_COMPRESSION_H_

#include <zstd.h>
#include <istream>
#include <memory>
#include <ostream>
#include <vector>
#include "absl/status/status.h"
#include "absl/status/statusor.h"

namespace aifo::data::compression {

/**
 * @class ZstdCompressor
 * @brief Streams data through Zstandard (ZSTD) in a constant‐memory fashion.
 *
 * This class never throws; errors are returned as `absl::Status`.
 *
 * \note Buffer sizes are chosen via `ZSTD_CStreamInSize()` / `ZSTD_CStreamOutSize()`.
 *
 * @code{.cpp}
 * auto compressor_or = ZstdCompressor::Create(3);
 * if (!compressor_or.ok()) {
 *   LOG(ERROR) << compressor_or.status();
 *   return;
 * }
 * auto compressor = std::move(compressor_or).value();
 * std::ifstream fin("input.bin", std::ios::binary);
 * std::ofstream fout("out.zst",  std::ios::binary);
 * if (auto st = compressor.Compress(fin, fout); !st.ok()) {
 *   LOG(ERROR) << st;
 * }
 * @endcode
 */
class ZstdCompressor {
 public:
  /**
   * @brief Create a compressor at the given compression level.
   * @param compression_level between `ZSTD_minCLevel()` and `ZSTD_maxCLevel()`.
   * @return a ready‐to‐use `ZstdCompressor` or an error.
   */
  [[nodiscard]] static absl::StatusOr<ZstdCompressor> Create(
      int compression_level = 3);

  ZstdCompressor(ZstdCompressor&&) noexcept = default;
  ZstdCompressor& operator=(ZstdCompressor&&) noexcept = default;

  ZstdCompressor(const ZstdCompressor&) = delete;
  ZstdCompressor& operator=(const ZstdCompressor&) = delete;

  ~ZstdCompressor() = default;

  /**
   * @brief Compresses all of `input` into `output`, chunk by chunk.
   * @param input  Source stream of raw bytes.
   * @param output Destination stream for ZSTD frames.
   * @returns `OkStatus()` on success, or an error.
   */
  [[nodiscard]] absl::Status Compress(std::istream& input,
                                      std::ostream& output) const;

 private:
  ZstdCompressor(ZSTD_CStream* ctx, int level) noexcept;

  std::unique_ptr<ZSTD_CStream, decltype(&ZSTD_freeCStream)> cstream_;
  int compression_level_;

  // Recommended chunk sizes (defined in compression.cpp)
  static const size_t kInputBufferSize;
  static const size_t kOutputBufferSize;
};

/**
 * @class ZstdDecompressor
 * @brief Streams Zstandard‐compressed data back to original bytes.
 *
 **/
class ZstdDecompressor {
 public:
  /**
   * @brief Create a streaming decompressor.
   * @return `ZstdDecompressor` on success, or an error.
   */
  [[nodiscard]] static absl::StatusOr<ZstdDecompressor> Create();

  ZstdDecompressor(ZstdDecompressor&&) noexcept = default;
  ZstdDecompressor& operator=(ZstdDecompressor&&) noexcept = default;

  ZstdDecompressor(const ZstdDecompressor&) = delete;
  ZstdDecompressor& operator=(const ZstdDecompressor&) = delete;

  ~ZstdDecompressor() = default;

  /**
   * @brief Reads ZSTD frames from `input`, writes raw bytes to `output`.
   */
  [[nodiscard]] absl::Status Decompress(std::istream& input,
                                        std::ostream& output) const;

 private:
  explicit ZstdDecompressor(ZSTD_DStream* ctx) noexcept;

  std::unique_ptr<ZSTD_DStream, decltype(&ZSTD_freeDStream)> dstream_;

  static const size_t kInputBufferSize;
  static const size_t kOutputBufferSize;
};

}  // namespace aifo::data::compression

#endif  // AIFO_AHCORE_INCLUDE_AHCORE_DATA_COMPRESSION_H_
