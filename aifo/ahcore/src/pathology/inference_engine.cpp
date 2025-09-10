// Copyright 2024 Jonas Teuwen. All Rights Reserved.
// Copyright 2025 Jonas Teuwen & Joren Brunekreef. All Rights Reserved.
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
#include "ahcore/pathology/inference_engine.h"

#include <spdlog/spdlog.h>
#include <algorithm>
#include <filesystem>
#include <map>
#include <memory>
#include <optional>
#include <string>
#include <utility>
#include <vector>

#include "ahcore/data/readers/disk_tile_reader.h"
#include "ahcore/data/writers/disk_tile_writer.h"
#include "ahcore/pathology/visualization/segmentation.h"
#include "aifocore/utilities/fmt.h"
#include "aifocore/utilities/temporary.h"
#include "aifocore/utilities/vips.h"
#include "aifocore/utilities/zip.h"
#include "dlup/backends/vips.h"
#include "dlup/foreground.h"
#include "dlup/utilities/filetype.h"

namespace fs = std::filesystem;

using aifo::utilities::NormalizeTensor;
using aifo::utilities::TensorToVImage;
using aifo::utilities::VImageToTensor;
using aifocore::fmt::format;

namespace aifo::pathology::inference {

constexpr auto GRID_ORDER = aifocore::tiling::GridOrder::kC;

absl::StatusOr<InferenceEngine> InferenceEngine::Create(
    const aifo::pathology::config::InferenceConfig& config) {

  auto device = config.device;

  absl::StatusOr<torch::Device> device_result;
  if (!device) {
    device_result = aifo::utilities::DeviceManager::GetAvailableDevice();
  } else {
    static const std::map<std::string, torch::DeviceType> devices = {
        {"cpu", torch::kCPU}, {"cuda", torch::kCUDA}, {"mps", torch::kMPS}};
    if (auto it = devices.find(*device); it != devices.end()) {
      device_result = torch::Device(it->second);
    } else {
      return absl::InvalidArgumentError("Unsupported device: " + *device);
    }
  }

  if (!device_result.ok()) {
    return device_result.status();
  }

  torch::Device devicex = device_result.value();
  spdlog::info("Using device: {}", torch::DeviceTypeName(devicex.type()));

  InferenceEngine processor(config, devicex);

  auto status = processor.LoadModelAndConfig();
  if (!status.ok()) {
    return status;
  }

  processor.pre_writer_transform_ = transforms::CreatePreWriterTransform(
      processor.GetModelConfig()->merge_method);
  spdlog::info("Using pre-writer transform: {}",
               processor.pre_writer_transform_->GetName());

  return processor;
}

InferenceEngine::InferenceEngine(
    const aifo::pathology::config::InferenceConfig& config,
    torch::Device device)
    : config_(config), device_(device) {
  spdlog::info("InferenceEngine initialized with model: {}",
               config_.model_path.string());
}

absl::Status InferenceEngine::LoadModelAndConfig() {
  aifocore::utilities::ZipArchive zip(config_.model_path);

  std::string model_content = zip.ReadFile("model.pt");
  model_ =
      std::make_unique<aifo::utilities::TorchModel>(model_content, device_);

  std::string config_content = zip.ReadFile("model_config.xml");
  model_config_ = ParseConfigurationFromString(config_content);
  spdlog::info("Loaded configuration:\n{}", model_config_->AsFormattedString());
  return absl::OkStatus();
}

const AifoModelConfiguration* InferenceEngine::GetModelConfig() const {
  return model_config_.get();
}

absl::Status InferenceEngine::ProcessImage(const fs::path& image_path) {
  return ProcessImageImpl(image_path, std::nullopt);
}

absl::Status InferenceEngine::ProcessImage(const fs::path& image_path,
                                           const fs::path& mask_path) {
  return ProcessImageImpl(image_path, mask_path);
}

absl::Status InferenceEngine::ProcessImageImpl(
    const fs::path& image_path, const std::optional<fs::path>& mask_path) {
  auto image_filename = image_path.filename().string();

  absl::StatusOr<SlideData> slide_data =
      mask_path ? CreateSlideData(image_path, *mask_path)
                : CreateSlideData(image_path);
  if (!slide_data.ok()) {
    return slide_data.status();
  }

  spdlog::info("Constructed dataset (mask: {})", mask_path.has_value());

  // false means "do not keep temp files"
  aifocore::utilities::TemporaryDirectory temp_dir(false);
  spdlog::info("Created temporary directory: {}", temp_dir.Path().string());

  auto status = RunInference(slide_data.value(), temp_dir.Path());
  if (!status.ok()) {
    return status;
  }

  auto output_filename =
      image_filename.substr(0, image_filename.find_last_of('.')) + ".tiff";
  fs::path output_path = config_.output_dir / output_filename;
  status = ProcessResults(slide_data.value(), temp_dir.Path(), output_path);
  if (!status.ok()) {
    return status;
  }

  if (config_.create_thumbnail) {
    auto thumbnail_filename =
        image_filename.substr(0, image_filename.find_last_of('.')) +
        ".thumbnail.png";
    fs::path thumbnail_path = config_.output_dir / thumbnail_filename;
    status = SaveThumbnail(slide_data.value(), output_path, thumbnail_path);
    if (!status.ok()) {
      return status;
    }
  }

  return absl::OkStatus();
}

absl::StatusOr<SlideData> InferenceEngine::CreateSlideData(
    const fs::path& image_path) {
  auto slide_backend = std::make_shared<dlup::backends::VipsSlide>(
      image_path, false, true, false);

  spdlog::info("Created slide backend for image: {}", image_path.string());

  absl::StatusOr<std::unique_ptr<dlup::SlideImage>> maybe_slide_image =
      dlup::SlideImage::Create(slide_backend);
  if (!maybe_slide_image.ok()) {
    return maybe_slide_image.status();
  }

  spdlog::info("Created slide image");

  std::shared_ptr<dlup::SlideImage> slide_image =
      std::move(maybe_slide_image).value();
  const double scaling = slide_image->GetScaling(GetModelConfig()->mpp);

  spdlog::info("Scaling: {}", scaling);

  auto geometry = slide_image->GetGeometry().Scaled(scaling);
  spdlog::info(
      "Slide geometry: visible region offset: {} visible region size: {}"
      "slide "
      "size: {}",
      geometry.offset, geometry.bounds, geometry.size);

  auto grid = std::make_unique<aifocore::tiling::Grid<int>>(
      aifocore::tiling::Grid<int>::FromTiling(
          geometry.offset, geometry.bounds, GetModelConfig()->tile_size,
          GetModelConfig()->tile_overlap,
          aifocore::tiling::TilingMode::kOverflow, GRID_ORDER));

  spdlog::info("Created grid");

  auto dataset = std::make_shared<dlup::SlideDataset>(
      slide_image, *grid, GetModelConfig()->mpp, GetModelConfig()->tile_size,
      false);

  spdlog::info("Created slide dataset with length: {}", dataset->Length());

  return SlideData{slide_image, geometry, std::move(grid), dataset};
}

absl::StatusOr<SlideData> InferenceEngine::CreateSlideData(
    const fs::path& image_path, const fs::path& mask_path) {
  auto slide_data = CreateSlideData(image_path);
  if (!slide_data.ok()) {
    return slide_data.status();
  }

  auto filtered_slide_data =
      ApplyForegroundFilter(slide_data.value(), mask_path);
  if (!filtered_slide_data.ok()) {
    return filtered_slide_data.status();
  }

  return filtered_slide_data;
}

absl::Status InferenceEngine::RunInference(const SlideData& slide_data,
                                           const fs::path& temp_dir) {
  spdlog::info("Running inference");

  auto writer = std::make_unique<aifo::data::writers::DiskTileWriter>(
      temp_dir, aifo::data::Metadata::Create()
                    ->SetMpp(GetModelConfig()->mpp)
                    ->SetGeometry(slide_data.geometry)
                    ->SetTileSize(GetModelConfig()->tile_size)
                    ->SetTileOverlap(GetModelConfig()->tile_overlap)
                    ->SetGridOrder(GRID_ORDER)
                    ->Lock());

  RETURN_IF_ERROR(writer->SetGrid(slide_data.grid),
                  "Failed to set grid for writer");

  spdlog::info("Set grid for writer");

  size_t total_batches =
      (slide_data.dataset->Length() + config_.batch_size - 1) /
      config_.batch_size;

  spdlog::info("Total batches: {}", total_batches);

  auto batch_generator =
      [this, total_batches, &slide_data](
          int batch_index) -> absl::StatusOr<std::vector<dlup::DatasetSample>> {
    std::vector<dlup::DatasetSample> batch_results;
    std::vector<torch::Tensor> batch_tensors;
    std::vector<aifocore::Size<int, 2>> coordinates;

    batch_tensors.reserve(config_.batch_size);
    coordinates.reserve(config_.batch_size);

    size_t start = batch_index * config_.batch_size;
    size_t end =
        std::min(start + config_.batch_size, slide_data.dataset->Length());

    for (size_t index = start; index < end; ++index) {
      dlup::DatasetSample sample;
      ASSIGN_OR_RETURN(sample, slide_data.dataset->GetTile(index),
                       "Failed to get tile " + std::to_string(index));

      coordinates.push_back(sample.coordinates);
      if (progress_cb_) {
        progress_cb_(static_cast<int>(batch_index + 1),
                     static_cast<int>(total_batches), static_cast<int>(index));
      }

      torch::Tensor image_tensor;
      ASSIGN_OR_RETURN(image_tensor, VImageToTensor(*sample.tile, device_),
                       "Failed to convert tile");
      image_tensor = image_tensor.to(torch::kFloat).unsqueeze(0);

      image_tensor = image_tensor / 255.0;

      image_tensor =
          NormalizeTensor(image_tensor, GetModelConfig()->normalization.mean,
                          GetModelConfig()->normalization.std);

      batch_tensors.push_back(image_tensor);
    }

    if (!batch_tensors.empty()) {
      torch::Tensor batched_input = torch::cat(batch_tensors, 0);
      c10::IValue raw_output = model_->Infer(batched_input);
      auto processed_output = model_->ProcessOutput(raw_output);

      torch::Tensor output_tensor = processed_output.ToTensor();

      torch::Tensor transformed_output =
          pre_writer_transform_->Forward(output_tensor);

      for (size_t i = 0; i < end - start; ++i) {
        vips::VImage output_image;
        ASSIGN_OR_RETURN(
            output_image, TensorToVImage(transformed_output[i]),
            format("Failed to convert tensor at position {} to VImage", i));

        auto output_map = std::make_shared<vips::VImage>(output_image);
        batch_results.push_back({output_map, coordinates[i], ""});
      }
    }

    return batch_results;
  };

  auto start_time = std::chrono::high_resolution_clock::now();

  RETURN_IF_ERROR(writer->Open(), "Failed to open writer");

  RETURN_IF_ERROR(
      writer->Consume([&batch_generator](
                          int batch_index) -> std::vector<dlup::DatasetSample> {
        auto batch_result = batch_generator(batch_index);
        if (!batch_result.ok()) {
          throw std::runtime_error(batch_result.status().ToString());
        }
        return std::move(batch_result).value();
      }),
      "Failed to consume batch data");

  auto total_size = writer->GetTotalSize();
  absl::Status close_status = writer->Close();
  if (!close_status.ok()) {
    spdlog::warn("Failed to close writer: {}", close_status.message());
  }

  auto duration = std::chrono::high_resolution_clock::now() - start_time;
  auto seconds = std::chrono::duration<double>(duration).count();

  spdlog::info("Inference completed in: {:.2f} seconds", seconds);
  spdlog::info("Total size of written data: {:.2f} MB",
               static_cast<double>(total_size) / (1024 * 1024));

  return absl::OkStatus();
}

absl::Status InferenceEngine::ProcessResults(const SlideData& slide_data,
                                             const fs::path& temp_dir,
                                             const fs::path& output_path) {
  spdlog::info("Stitching tiles...");

  aifo::data::readers::StitchingMode stitching_mode;
  std::shared_ptr<aifo::data::readers::PostStitchTransform> post_transform;

  const auto& merge_method = GetModelConfig()->merge_method;

  if (merge_method == "crop") {
    stitching_mode = aifo::data::readers::StitchingMode::kCrop;
    // TODO(jonasteuwen): Consider where this output stuff has to be set
    post_transform = std::make_shared<aifo::data::readers::IdentityTransform>(
        aifo::data::Metadata::Create()
            ->Set(MetadataKeys::NumOutputChannels, 1)
            ->Lock());
  } else if (merge_method == "average") {
    stitching_mode = aifo::data::readers::StitchingMode::kAverage;
    post_transform = std::make_shared<aifo::data::readers::AveragingTransform>(
        aifo::data::Metadata::Create()
            ->Set(MetadataKeys::NumOutputChannels, 1)
            ->Set(MetadataKeys::OutputPixelFormat, VIPS_FORMAT_UCHAR)
            ->Set(MetadataKeys::OutputInterpretation, VIPS_INTERPRETATION_B_W)
            ->Lock());
  } else {
    return absl::InvalidArgumentError("Unknown merge method: " + merge_method);
  }

  auto reader = std::make_unique<aifo::data::readers::DiskTileReader>(
      temp_dir, stitching_mode, post_transform);

  RETURN_IF_ERROR(reader->Open(), "Failed to open reader");

  RETURN_IF_ERROR(
      aifocore::utilities::SaveVipsImageToFile(
          reader->GetImage(), output_path,
          vips::VImage::option()
              ->set("tile_height", config_.tiff_tile_size)
              ->set("tile_width", config_.tiff_tile_size)
              ->set("xres", 1000.0 / GetModelConfig()->mpp)
              ->set("yres", 1000.0 / GetModelConfig()->mpp)
              ->set("tile", true)
              ->set("pyramid", true)
              ->set("compression", VIPS_FOREIGN_TIFF_COMPRESSION_LZW)),
      "Failed saving region to file");

  spdlog::info("Output path: {}", output_path.string());
  spdlog::info("Results saved to: {} (size: {}x{}x{} mpp: {})",
               output_path.string(), reader->GetImage().width(),
               reader->GetImage().height(), reader->GetImage().bands(),
               GetModelConfig()->mpp);

  return absl::OkStatus();
}

absl::Status InferenceEngine::SaveThumbnail(
    const SlideData& slide_data, const fs::path& tiff_path,
    const std::string& thumbnail_filename) {

  auto thumbnail_ptr =
      slide_data.dataset->GetSlide()->GetThumbnail({1024, 1024});
  if (!thumbnail_ptr) {
    return MAKE_STATUS(absl::StatusCode::kUnavailable,
                       "Failed to generate thumbnail.");
  }
  vips::VImage thumbnail = *thumbnail_ptr;

  vips::VImage segmentation_map;
  try {
    segmentation_map = vips::VImage::new_from_file(
        tiff_path.string().c_str(),
        vips::VImage::option()->set("access", VIPS_ACCESS_SEQUENTIAL));

    // Resize the segmentation map to match the thumbnail size
    segmentation_map = segmentation_map.resize(
        static_cast<double>(thumbnail.width()) / segmentation_map.width(),
        vips::VImage::option()
            ->set("vscale", static_cast<double>(thumbnail.height()) /
                                segmentation_map.height())
            ->set("kernel", VIPS_KERNEL_NEAREST));
  } catch (const vips::VError& e) {
    return MAKE_STATUS(
        absl::StatusCode::kInternal,
        aifocore::fmt::format("Failed to read segmentation map from file: {}",
                              e.what()));
  }

  vips::VImage lut = visualization::SegmentationVisualizer::CreateLut(
      GetModelConfig()->labels);

  vips::VImage overlayed_thumbnail =
      visualization::SegmentationVisualizer::OverlaySegmentation(
          thumbnail, segmentation_map, lut, 0.6);

  RETURN_IF_ERROR(aifocore::utilities::SaveVipsImageToFile(
                      overlayed_thumbnail, thumbnail_filename, nullptr),
                  "Failed saving thumbnail to file");

  spdlog::info("Thumbnail saved to: {}", thumbnail_filename);
  return absl::OkStatus();
}

absl::StatusOr<SlideData> InferenceEngine::ApplyForegroundFilter(
    const SlideData& slide_data, const fs::path& mask_path) {
  dlup::utilities::FileInfo file_info =
      dlup::utilities::DetectFileType(mask_path);

  switch (file_info.file_type) {
    case dlup::utilities::FileType::kTiff: {
      auto filtered_slide_data =
          ApplyForegroundFilterTiff(slide_data, mask_path);
      if (!filtered_slide_data.ok()) {
        return filtered_slide_data.status();
      }
      return filtered_slide_data.value();
    }
    default:
      return MAKE_STATUS(absl::StatusCode::kInvalidArgument,
                         "Unknown mask type");
  }
}

absl::StatusOr<SlideData> InferenceEngine::ApplyForegroundFilterTiff(
    const SlideData& slide_data, const fs::path& mask_path) {
  auto mask_backend = std::make_shared<dlup::backends::VipsSlide>(
      mask_path, false, false, false);
  spdlog::info("Reading mask of type TIFF from {}", mask_path.string());

  auto foreground_result = dlup::Foreground<int>::FilterGrid(
      *(slide_data.grid), *mask_backend, GetModelConfig()->tile_size,
      GetModelConfig()->mpp, config_.mask_threshold);

  slide_data.dataset->SetForegroundFilter(foreground_result);
  spdlog::info("Set foreground filter");

  return slide_data;
}

}  // namespace aifo::pathology::inference
