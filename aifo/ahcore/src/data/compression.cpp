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
#include "ahcore/data/compression.h"

#include <spdlog/spdlog.h>
#include <zstd.h>
#include <vector>
#include "absl/status/status.h"
#include "absl/status/statusor.h"
#include "aifocore/status/status_macros.h"

// Define the buffer sizes at runtime:
const size_t aifo::data::compression::ZstdCompressor::kInputBufferSize =
    ZSTD_CStreamInSize();
const size_t aifo::data::compression::ZstdCompressor::kOutputBufferSize =
    ZSTD_CStreamOutSize();

const size_t aifo::data::compression::ZstdDecompressor::kInputBufferSize =
    ZSTD_DStreamInSize();
const size_t aifo::data::compression::ZstdDecompressor::kOutputBufferSize =
    ZSTD_DStreamOutSize();

namespace aifo::data::compression {

absl::StatusOr<ZstdCompressor> ZstdCompressor::Create(int compression_level) {
  if (compression_level < ZSTD_minCLevel() ||
      compression_level > ZSTD_maxCLevel()) {
    return MAKE_STATUS(absl::StatusCode::kInvalidArgument,
                       "Compression level out of range");
  }
  ZSTD_CStream* ctx = ZSTD_createCStream();
  if (ctx == nullptr) {
    return MAKE_STATUS(absl::StatusCode::kInternal,
                       "ZSTD_createCStream() failed");
  }
  size_t ret = ZSTD_initCStream(ctx, compression_level);
  if (ZSTD_isError(ret) != 0) {
    ZSTD_freeCStream(ctx);
    return MAKE_STATUS(absl::StatusCode::kInternal, ZSTD_getErrorName(ret));
  }
  return ZstdCompressor(ctx, compression_level);
}

ZstdCompressor::ZstdCompressor(ZSTD_CStream* ctx, int level) noexcept
    : cstream_(ctx, &ZSTD_freeCStream), compression_level_(level) {}

absl::Status ZstdCompressor::Compress(std::istream& input,
                                      std::ostream& output) const {
  spdlog::debug("[Compress] start");
  std::vector<char> in_buf(kInputBufferSize);
  std::vector<char> out_buf(kOutputBufferSize);
  ZSTD_inBuffer input_zstd = {.src = in_buf.data(), .size = 0, .pos = 0};
  ZSTD_outBuffer output_zstd = {
      .dst = out_buf.data(), .size = kOutputBufferSize, .pos = 0};

  while (true) {
    input.read(in_buf.data(), static_cast<std::streamsize>(kInputBufferSize));
    auto got = input.gcount();
    spdlog::debug("[Compress] read {} bytes (eof={}, fail={})", got,
                  input.eof(), input.fail());
    if (got == 0) {
      spdlog::debug("[Compress] no more input, break");
      break;
    }
    input_zstd.src = in_buf.data();
    input_zstd.size = got;
    input_zstd.pos = 0;

    while (input_zstd.pos < input_zstd.size) {
      output_zstd.pos = 0;
      size_t ret =
          ZSTD_compressStream(cstream_.get(), &output_zstd, &input_zstd);
      spdlog::debug(
          "[Compress] ZSTD_compressStream -> ret={}, output_zstd.pos={}", ret,
          output_zstd.pos);
      if (ZSTD_isError(ret) != 0) {
        return MAKE_STATUS(absl::StatusCode::kInternal, ZSTD_getErrorName(ret));
      }
      output.write(static_cast<const char*>(output_zstd.dst),
                   static_cast<std::streamsize>(output_zstd.pos));
    }
  }

  {
    while (true) {
      output_zstd.pos = 0;
      size_t remaining = ZSTD_endStream(cstream_.get(), &output_zstd);
      if (ZSTD_isError(remaining) != 0) {
        return MAKE_STATUS(absl::StatusCode::kInternal,
                           ZSTD_getErrorName(remaining));
      }

      // Write whatever this iteration produced
      output.write(static_cast<const char*>(output_zstd.dst),
                   static_cast<std::streamsize>(output_zstd.pos));

      if (remaining == 0) {
        // no more data buffered inside ZSTD → we're done
        break;
      }
    }
  }

  return absl::OkStatus();
}

absl::StatusOr<ZstdDecompressor> ZstdDecompressor::Create() {
  ZSTD_DStream* ctx = ZSTD_createDStream();
  if (ctx == nullptr) {
    return MAKE_STATUS(absl::StatusCode::kInternal,
                       "ZSTD_createDStream() failed");
  }
  size_t ret = ZSTD_initDStream(ctx);
  if (ZSTD_isError(ret) != 0) {
    ZSTD_freeDStream(ctx);
    return MAKE_STATUS(absl::StatusCode::kInternal, ZSTD_getErrorName(ret));
  }
  return ZstdDecompressor(ctx);
}

ZstdDecompressor::ZstdDecompressor(ZSTD_DStream* ctx) noexcept
    : dstream_(ctx, &ZSTD_freeDStream) {}

absl::Status ZstdDecompressor::Decompress(std::istream& input,
                                          std::ostream& output) const {
  std::vector<char> in_buf(kInputBufferSize);
  std::vector<char> out_buf(kOutputBufferSize);
  ZSTD_inBuffer input_zstd = {in_buf.data(), 0, 0};
  ZSTD_outBuffer output_zstd = {out_buf.data(), kOutputBufferSize, 0};

  while (input) {
    input.read(in_buf.data(), static_cast<std::streamsize>(kInputBufferSize));
    input_zstd.size = input.gcount();
    input_zstd.pos = 0;
    while (input_zstd.pos < input_zstd.size) {
      output_zstd.pos = 0;
      size_t ret =
          ZSTD_decompressStream(dstream_.get(), &output_zstd, &input_zstd);
      if (ZSTD_isError(ret) != 0) {
        return MAKE_STATUS(absl::StatusCode::kInternal, ZSTD_getErrorName(ret));
      }
      output.write(static_cast<const char*>(output_zstd.dst),
                   static_cast<std::streamsize>(output_zstd.pos));
    }
  }
  return absl::OkStatus();
}

}  // namespace aifo::data::compression
