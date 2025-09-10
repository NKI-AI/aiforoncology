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

#ifndef AIFO_FASTSLIDE_INCLUDE_FASTSLIDE_UTILITIES_TIFF_TIFF_DIRECTORY_PROCESSOR_H_
#define AIFO_FASTSLIDE_INCLUDE_FASTSLIDE_UTILITIES_TIFF_TIFF_DIRECTORY_PROCESSOR_H_

#include <cstdint>
#include <functional>
#include <memory>
#include <vector>

#include "absl/status/status.h"
#include "absl/status/statusor.h"
#include "fastslide/utilities/tiff_file.h"

namespace fastslide {

/// @brief Information about a TIFF directory with page number
struct TiffDirectoryInfoWithPage {
  uint16_t page;           ///< TIFF page/directory number
  TiffDirectoryInfo info;  ///< Directory information from TiffFile
};

/// @brief Callback interface for format-specific directory processing
class TiffDirectoryCallback {
 public:
  virtual ~TiffDirectoryCallback() = default;

  /// @brief Process a single TIFF directory
  /// @param dir_info Directory information with page number
  /// @return Status indicating success or failure
  virtual absl::Status ProcessDirectory(
      const TiffDirectoryInfoWithPage& dir_info) = 0;

  /// @brief Called after all directories have been processed
  /// @return Status indicating success or failure
  virtual absl::Status Finalize() = 0;
};

/// @brief Helper class for common TIFF directory traversal and processing
class TiffDirectoryProcessor {
 public:
  /// @brief Constructor
  /// @param tiff_file TiffFile wrapper instance
  explicit TiffDirectoryProcessor(TiffFile tiff_file);

  /// @brief Process all directories using the provided callback
  /// @param callback Format-specific processing callback
  /// @return Status indicating success or failure
  absl::Status ProcessAllDirectories(TiffDirectoryCallback& callback);

  /// @brief Get the number of directories in the TIFF file
  /// @return Number of directories or error
  absl::StatusOr<uint16_t> GetDirectoryCount() const;

 private:
  TiffFile tiff_file_;  ///< TIFF file wrapper

  /// @brief Extract directory information from current directory
  /// @param page Page number
  /// @return Directory information with page number or error
  absl::StatusOr<TiffDirectoryInfoWithPage> ExtractDirectoryInfo(uint16_t page);
};

}  // namespace fastslide

#endif  // AIFO_FASTSLIDE_INCLUDE_FASTSLIDE_UTILITIES_TIFF_TIFF_DIRECTORY_PROCESSOR_H_
