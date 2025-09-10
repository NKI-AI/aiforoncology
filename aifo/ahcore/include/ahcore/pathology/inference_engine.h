// Copyright 2024 Jonas Teuwen. All Rights Reserved.
// Copyright 2025 Joren Brunekreef. All Rights Reserved.
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
#pragma once

#include <spdlog/spdlog.h>

#include <torch/script.h>
#include <filesystem>
#include <functional>
#include <map>
#include <memory>
#include <optional>
#include <string>
#include <utility>
#include <vector>

#include "absl/status/status.h"
#include "absl/status/statusor.h"

#include "ahcore/utilities/torch.h"

#include "ahcore/pathology/config/inference_config.h"
#include "ahcore/pathology/configuration.h"
#include "ahcore/pathology/transforms/pre_writer_transforms.h"
#include "aifocore/tiling/grid.h"
#include "aifocore/utilities/temporary.h"
#include "aifocore/utilities/zip.h"
#include "dlup/slide_dataset.h"
#include "dlup/slide_image.h"

namespace dlup::backends {
class VipsSlideBackend;
}

namespace fs = std::filesystem;

namespace aifo::pathology::inference {

struct SlideData {
  std::shared_ptr<dlup::SlideImage> slide_image;
  dlup::SlideGeometry geometry;
  std::shared_ptr<aifocore::tiling::Grid<int>> grid;
  std::shared_ptr<dlup::SlideDataset> dataset;
};

class InferenceEngine {
 public:
  static absl::StatusOr<InferenceEngine> Create(
      const aifo::pathology::config::InferenceConfig& config);

  explicit InferenceEngine(
      const aifo::pathology::config::InferenceConfig& config,
      torch::Device device);

  absl::Status ProcessImage(const fs::path& image_path);
  absl::Status ProcessImage(const fs::path& image_path,
                            const fs::path& mask_path);

  using ProgressCallback = std::function<void(
      int /*current_batch*/, int /*total_batches*/, int /*tile_index*/)>;

  // Set a progress callback to be invoked from C++ during processing.
  void SetProgressCallback(ProgressCallback cb) {
    progress_cb_ = std::move(cb);
  }

 private:
  // Unified implementation to avoid code duplication between overloads
  absl::Status ProcessImageImpl(const fs::path& image_path,
                                const std::optional<fs::path>& mask_path);

  absl::Status LoadModelAndConfig();

  const AifoModelConfiguration* GetModelConfig() const;

  absl::StatusOr<SlideData> CreateSlideData(const fs::path& image_path);
  absl::StatusOr<SlideData> CreateSlideData(const fs::path& image_path,
                                            const fs::path& mask_path);

  absl::Status RunInference(const SlideData& slide_data,
                            const fs::path& temp_dir);
  absl::Status ProcessResults(const SlideData& slide_data,
                              const fs::path& temp_dir,
                              const fs::path& output_path);
  absl::Status SaveThumbnail(const SlideData& slide_data,
                             const fs::path& tiff_path,
                             const std::string& thumbnail_filename);

  absl::StatusOr<SlideData> ApplyForegroundFilter(const SlideData& slide_data,
                                                  const fs::path& mask_path);
  absl::StatusOr<SlideData> ApplyForegroundFilterTiff(
      const SlideData& slide_data, const fs::path& mask_path);

  aifo::pathology::config::InferenceConfig config_;

  std::unique_ptr<aifo::utilities::TorchModel> model_;
  std::unique_ptr<AifoModelConfiguration> model_config_;
  torch::Device device_;
  std::shared_ptr<transforms::PreWriterTransform> pre_writer_transform_;
  ProgressCallback progress_cb_;
};

}  // namespace aifo::pathology::inference