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
#include <gtest/gtest.h>

#include <memory>
#include <string>

#include "ahcore/data/metadata.h"
#include "aifocore/concepts/numeric.h"

using namespace aifo::data;
using aifocore::Size;

class MetadataTest : public ::testing::Test {
 protected:
  std::shared_ptr<Metadata> metadata = Metadata::Create();

  void SetUp() override {
    metadata->SetGeometry(dlup::SlideGeometry{{400, 300}, {0, 0}, {400, 300}})
        ->SetTileSize(100, 100)
        ->SetTileOverlap(20, 20)
        ->SetMpp(0.25f)
        ->Set("format", "jpeg")
        ->Set("compression", 75);
  }
};

TEST_F(MetadataTest, InitializationAndSet) {
  ASSERT_EQ(metadata->GetGeometry().size, (Size<int, 2>(400, 300)));
  ASSERT_EQ(metadata->GetTileSize(), (Size<int, 2>(100, 100)));
  ASSERT_EQ(metadata->GetTileOverlap(), (Size<int, 2>(20, 20)));
  ASSERT_FLOAT_EQ(metadata->GetMpp(), 0.25f);
  ASSERT_EQ(metadata->Get<std::string>("format"), "jpeg");
  ASSERT_EQ(metadata->Get<int>("compression"), 75);
}

TEST_F(MetadataTest, LockingAndUnlocking) {
  ASSERT_FALSE(metadata->IsLocked());
  metadata->Lock();
  // Try to set value after locking
  ASSERT_THROW(metadata->Set("new_key", "new_value"), std::runtime_error);
  ASSERT_TRUE(metadata->IsLocked());
  metadata->Unlock();
  ASSERT_FALSE(metadata->IsLocked());
}

TEST_F(MetadataTest, SerializationAndDeserialization) {
  const std::string filename = "test_metadata.bin";

  // Serialize
  metadata->Serialize(filename);

  // Deserialize
  auto deserialized_metadata = aifo::data::Metadata::LoadFromFile(filename);

  // Validate deserialized data
  ASSERT_EQ(*metadata, *deserialized_metadata);

  // Cleanup
  std::filesystem::remove(filename);
}

TEST_F(MetadataTest, LargeMetadataHandling) {
  auto large_metadata = aifo::data::Metadata::Create();
  for (int i = 0; i < 10000; ++i) {
    large_metadata->Set("string_field_" + std::to_string(i),
                        "example_string_" + std::to_string(i));
    large_metadata->Set("int_field_" + std::to_string(i), i);
    large_metadata->Set("float_field_" + std::to_string(i), i * 0.1f);
  }

  const std::string filename = "large_metadata.bin";

  // Serialize and deserialize
  large_metadata->Serialize(filename);
  auto deserialized_large_metadata =
      aifo::data::Metadata::LoadFromFile(filename);

  // Validate
  ASSERT_EQ(*large_metadata, *deserialized_large_metadata);

  // Cleanup
  std::filesystem::remove(filename);
}

TEST_F(MetadataTest, Compression) {
  auto metadata = aifo::data::Metadata::Create();
  metadata->SetGeometry(dlup::SlideGeometry{{400, 300}, {0, 0}, {400, 300}})
      ->SetTileSize(100, 100)
      ->SetTileOverlap(20, 20)
      ->SetMpp(0.25f);

  const std::string compressed_filename = "compressed_metadata.bin";

  // Serialize with compression
  std::ofstream compressed_file(compressed_filename, std::ios::binary);
  metadata->Serialize(compressed_file, true, 5);
  compressed_file.close();

  // Deserialize
  auto deserialized_metadata =
      aifo::data::Metadata::LoadFromFile(compressed_filename);

  // Validate
  ASSERT_EQ(*metadata, *deserialized_metadata);

  // Cleanup
  std::filesystem::remove(compressed_filename);
}

TEST_F(MetadataTest, KeyExistenceAndRetrieval) {
  ASSERT_TRUE(metadata->HasKey(MetadataKeys::Geometry));
  ASSERT_TRUE(metadata->HasKey("format"));
  ASSERT_FALSE(metadata->HasKey("nonexistent_key"));

  auto tile_size_opt =
      metadata->GetOptional<Size<int, 2>>(MetadataKeys::TileSize);
  ASSERT_TRUE(tile_size_opt.has_value());
  ASSERT_EQ(*tile_size_opt, (Size<int, 2>(100, 100)));  // Use constructor

  auto nonexistent_opt = metadata->GetOptional<int>("nonexistent_key");
  ASSERT_FALSE(nonexistent_opt.has_value());
}

int main(int argc, char** argv) {
  ::testing::InitGoogleTest(&argc, argv);
  int result = RUN_ALL_TESTS();
  return result;
}
