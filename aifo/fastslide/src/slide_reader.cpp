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

#include "fastslide/slide_reader.h"

#include <algorithm>
#include <filesystem>
#include <shared_mutex>
#include <string>
#include <utility>
#include <vector>

#include "absl/status/status.h"
#include "absl/strings/ascii.h"

#include "aifocore/status/status_macros.h"
#include "aifocore/utilities/fmt.h"
#include "fastslide/utilities/colors.h"

namespace fastslide {

using aifocore::fmt::format;

RegionSpec SlideReader::ClampRegion(const RegionSpec& region,
                                    const ImageDimensions& image_dims) {
  // Handle edge case of zero-sized image
  if (image_dims[0] == 0 || image_dims[1] == 0) {
    RegionSpec clamped = region;
    clamped.top_left[0] = 0;
    clamped.top_left[1] = 0;
    clamped.size[0] = 0;
    clamped.size[1] = 0;
    return clamped;
  }

  RegionSpec clamped = region;

  // Clamp coordinates to image bounds using std::clamp for safety
  clamped.top_left[0] =
      std::clamp(clamped.top_left[0], uint32_t{0}, image_dims[0]);
  clamped.top_left[1] =
      std::clamp(clamped.top_left[1], uint32_t{0}, image_dims[1]);

  // Calculate remaining image area with overflow protection
  const uint32_t remaining_width = image_dims[0] - clamped.top_left[0];
  const uint32_t remaining_height = image_dims[1] - clamped.top_left[1];

  // Clamp size to remaining image area
  clamped.size[0] = std::min(clamped.size[0], remaining_width);
  clamped.size[1] = std::min(clamped.size[1], remaining_height);

  return clamped;
}

// SlideReaderRegistry implementation
SlideReaderRegistry& SlideReaderRegistry::GetInstance() {
  static SlideReaderRegistry instance;
  return instance;
}

void SlideReaderRegistry::RegisterReader(std::string_view extension,
                                         SlideReaderFactory factory) {
  // Validate inputs
  if (extension.empty()) {
    return;  // Silently ignore empty extensions
  }

  if (!factory) {
    return;  // Silently ignore null factories
  }

  std::string ext_lower{extension};
  absl::AsciiStrToLower(&ext_lower);

  // Ensure extension starts with a dot
  if (ext_lower[0] != '.') {
    ext_lower = "." + ext_lower;
  }

  // Thread-safe write operation: exclusive lock for modification
  std::unique_lock<std::shared_mutex> lock(mutex_);
  factories_[std::move(ext_lower)] = std::move(factory);
}

absl::StatusOr<std::unique_ptr<SlideReader>> SlideReaderRegistry::CreateReader(
    std::string_view filename) const {
  if (filename.empty()) {
    return MAKE_STATUS(absl::StatusCode::kInvalidArgument,
                       "Filename cannot be empty");
  }

  const std::filesystem::path file_path{filename};
  std::string extension = file_path.extension().string();

  if (extension.empty()) {
    return MAKE_STATUS(
        absl::StatusCode::kInvalidArgument,
        aifocore::fmt::format("File {} has no extension", filename));
  }

  absl::AsciiStrToLower(&extension);

  // Thread-safe read operation: shared lock allows concurrent readers
  std::shared_lock<std::shared_mutex> lock(mutex_);

  const auto iter = factories_.find(extension);
  if (iter == factories_.end()) {
    return MAKE_STATUS(absl::StatusCode::kInvalidArgument,
                       format("Unsupported file format '{}' for file '{}'",
                              extension, filename));
  }

  // Call the factory function (still under shared lock to
  // ensure factory doesn't change)
  auto reader_result = iter->second(filename);
  if (!reader_result.ok()) {
    return MAKE_STATUS(reader_result.status().code(),
                       format("Failed to create reader for '{}': {}", filename,
                              reader_result.status().message()));
  }

  return reader_result;
}

std::vector<std::string> SlideReaderRegistry::GetSupportedExtensions() const {
  // Thread-safe read operation: shared lock allows concurrent readers
  std::shared_lock<std::shared_mutex> lock(mutex_);

  std::vector<std::string> extensions;
  extensions.reserve(factories_.size());

  for (const auto& [extension, factory] : factories_) {
    extensions.push_back(extension);
  }

  std::ranges::sort(extensions);
  return extensions;
}

int SlideReader::GetBestLevelForDownsample(double downsample) const {
  if (downsample <= 1.0) {
    return 0;
  }

  const int level_count = GetLevelCount();
  if (level_count == 0) {
    return 0;
  }

  int best_level = 0;
  double best_diff = std::abs(1.0 - downsample);

  for (int level = 0; level < level_count; ++level) {
    auto level_info_result = GetLevelInfo(level);
    if (!level_info_result.ok()) {

      continue;  // Skip invalid levels
    }

    double level_downsample = level_info_result.value().downsample_factor;
    double diff = std::abs(level_downsample - downsample);
    if (diff < best_diff) {
      best_diff = diff;
      best_level = level;
    }
  }

  return best_level;
}

}  // namespace fastslide
