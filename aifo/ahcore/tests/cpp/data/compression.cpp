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
#include <gtest/gtest.h>
#include <sstream>
#include <string>
#include <utility>
#include "absl/status/status.h"
#include "absl/status/statusor.h"

using namespace aifo::data::compression;

TEST(ZstdCodecTest, RoundTripSmallString) {
  std::string original = "The quick brown fox jumps over the lazy dog";
  std::stringstream in{original};
  std::stringstream compressed;
  std::stringstream out;

  // Create compressor @ level 2
  auto comp_or = ZstdCompressor::Create(2);
  ASSERT_TRUE(comp_or.ok()) << comp_or.status();
  auto compressor = std::move(comp_or).value();
  ASSERT_TRUE(compressor.Compress(in, compressed).ok());

  // Decompress
  auto decomp_or = ZstdDecompressor::Create();
  ASSERT_TRUE(decomp_or.ok()) << decomp_or.status();
  auto decompressor = std::move(decomp_or).value();
  ASSERT_TRUE(decompressor.Decompress(compressed, out).ok());

  EXPECT_EQ(out.str(), original);
}

TEST(ZstdCodecTest, EmptyInput) {
  std::stringstream in;
  std::stringstream compressed;
  std::stringstream out;

  auto comp_or = ZstdCompressor::Create();
  ASSERT_TRUE(comp_or.ok());
  ASSERT_TRUE(comp_or->Compress(in, compressed).ok());

  auto decomp_or = ZstdDecompressor::Create();
  ASSERT_TRUE(decomp_or.ok());
  ASSERT_TRUE(decomp_or->Decompress(compressed, out).ok());
  EXPECT_TRUE(out.str().empty());
}
