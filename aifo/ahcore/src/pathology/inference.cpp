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
#include <vips/vips.h>
#include <vips/vips8>

#include <spdlog/spdlog.h>

#include <torch/script.h>
#include <algorithm>
#include <chrono>
#include <filesystem>
#include <map>
#include <memory>
#include <string>
#include <utility>
#include <vector>

#include "absl/status/status.h"
#include "absl/status/statusor.h"

#include "ahcore/data/readers/disk_tile_reader.h"
#include "ahcore/data/writers/disk_tile_writer.h"
#include "ahcore/pathology/config/process_config.h"
#include "ahcore/pathology/configuration.h"
#include "ahcore/pathology/transforms/pre_writer_transforms.h"
#include "ahcore/pathology/visualization/segmentation.h"
#include "ahcore/utilities/env_config.h"
#include "aifocore/status/status_macros.h"
#include "aifocore/tiling/grid.h"
#include "aifocore/utilities/fmt.h"
#include "aifocore/utilities/spinners.h"
#include "aifocore/utilities/temporary.h"
#include "aifocore/utilities/vips.h"
#include "aifocore/utilities/zip.h"
#include "dlup/backends/vips.h"
#include "dlup/foreground.h"
#include "dlup/slide_dataset.h"
#include "dlup/slide_image.h"
#include "dlup/utilities/filetype.h"

namespace fs = std::filesystem;
using aifo::utilities::NormalizeTensor;
using aifo::utilities::TensorToVImage;
using aifo::utilities::VImageToTensor;
using aifocore::fmt::format;

namespace aifo::pathology::inference {

constexpr auto GRID_ORDER = aifocore::tiling::GridOrder::kC;

class ProcessImage {
 public:
  /**
   * @brief Factory method to create a ProcessImage instance with proper error handling.
   *
   * @param config The process configuration.
   * @return absl::StatusOr<ProcessImage> A status containing either a ProcessImage or an error.
   */
  static absl::StatusOr<ProcessImage> Create(
      aifo::pathology::config::ProcessConfig config) {
    // Create a processor instance
    ProcessImage processor(std::move(config));

    // Validate device
    auto device_result =
        processor.ParseAndValidateDevice(processor.config_.device);
    if (!device_result.ok()) {
      return device_result.status();
    }

    // Set device and log
    processor.device_ = device_result.value();
    spdlog::info("Using device: {}",
                 torch::DeviceTypeName(processor.device_.type()));

    // Load model and config
    auto status = processor.LoadModelAndConfig();
    if (!status.ok()) {
      return status;
    }

    // Set up pre-writer transform
    processor.pre_writer_transform_ = transforms::CreatePreWriterTransform(
        processor.GetConfig()->merge_method);
    spdlog::info("Using pre-writer transform: {}",
                 processor.pre_writer_transform_->GetName());

    // Return the initialized processor
    return processor;
  }

  /**
   * @brief Simple constructor that just initializes member variables.
   * @param config The process configuration.
   */
  explicit ProcessImage(aifo::pathology::config::ProcessConfig config)
      : config_(std::move(config)), device_(torch::kCPU) {}

  absl::Status InferAndSave() {
    // Check if temporary files should be kept via environment configuration
    bool keep_temp_files = aifo::utilities::EnvConfig::GetInstance().GetBool(
        aifo::utilities::EnvConfig::kKeepTemporaryFiles);

    // Create temporary directory with appropriate flag
    aifocore::utilities::TemporaryDirectory temp_directory(keep_temp_files);

    // Log the temporary directory path if it's being kept
    if (temp_directory.IsKept()) {
      spdlog::info("Temporary files are preserved at: {}",
                   temp_directory.Path().string());
    }

    RETURN_IF_ERROR(ConstructDataset(), "Failed to construct dataset");
    RETURN_IF_ERROR(RunInference(temp_directory.Path()),
                    "Failed to run inference");
    RETURN_IF_ERROR(ProcessResults(temp_directory.Path()),
                    "Failed to process results");

    return absl::OkStatus();
  }

 private:
  absl::Status LoadModelAndConfig() {
    aifocore::utilities::ZipArchive zip(
        config_.model_file);  // RAII ensures the zip file is properly closed

    // Load the model
    std::string model_content = zip.ReadFile("model.pt");
    model_ =
        std::make_unique<aifo::utilities::TorchModel>(model_content, device_);

    // Load the configuration
    std::string config_content = zip.ReadFile("model_config.xml");
    config_impl_ = ParseConfigurationFromString(config_content);
    spdlog::info("Loaded configuration:\n{}",
                 config_impl_->AsFormattedString());
    return absl::OkStatus();
  }

  absl::StatusOr<torch::Device> ParseAndValidateDevice(
      const std::optional<std::string>& device) {
    static const std::map<std::string, torch::DeviceType> devices = {
        {"cpu", torch::kCPU}, {"cuda", torch::kCUDA}, {"mps", torch::kMPS}};
    if (!device) {
      return aifo::utilities::DeviceManager::GetAvailableDevice();
    }
    if (auto it = devices.find(*device); it != devices.end()) {
      if (it->second == torch::kCUDA && !torch::cuda::is_available()) {
        return absl::UnavailableError("CUDA is not available on this system.");
      }
      if (it->second == torch::kMPS && !torch::mps::is_available()) {
        return absl::UnavailableError("MPS is not available on this system.");
      }
      return torch::Device(it->second);
    }
    return absl::InvalidArgumentError("Unsupported device: " + *device);
  }

  const AifoModelConfiguration* GetConfig() const { return config_impl_.get(); }

  absl::Status ConstructDataset() {
    auto slide_backend = std::make_shared<dlup::backends::VipsSlide>(
        config_.image_file, false, true, false);

    // TODO(jonasteuwen): Figure out a way to be able
    // to construct the variable inside the ASSIGN_OR_RETURN
    // TODO(jonasteuwen): Does this need to be a shared_ptr?
    absl::StatusOr<std::unique_ptr<dlup::SlideImage>> maybe_slide_image =
        dlup::SlideImage::Create(slide_backend);

    if (!maybe_slide_image.ok()) {
      return maybe_slide_image.status();
    }

    std::shared_ptr<dlup::SlideImage> slide_image =
        std::move(maybe_slide_image.value());
    const double scaling = slide_image->GetScaling(GetConfig()->mpp);

    geometry_ = slide_image->GetGeometry().Scaled(scaling);
    spdlog::info(
        "Slide geometry: visible region offset: {} visible region size: {} and "
        "slide "
        "size: {}",
        geometry_.offset, geometry_.bounds, geometry_.size);

    grid_ = std::make_unique<aifocore::tiling::Grid<int>>(
        aifocore::tiling::Grid<int>::FromTiling(
            geometry_.offset, geometry_.bounds, GetConfig()->tile_size,
            GetConfig()->tile_overlap, aifocore::tiling::TilingMode::kOverflow,
            GRID_ORDER));

    dataset_ = std::make_shared<dlup::SlideDataset>(
        slide_image, *grid_, GetConfig()->mpp, GetConfig()->tile_size, false);

    if (!config_.mask_file.empty()) {
      RETURN_IF_ERROR(ApplyForegroundFilter(),
                      "Failed to apply foreground filter");
    } else {
      spdlog::info("No mask file provided. Using the full slide for inference");
    }
    return absl::OkStatus();
  }

  /**
   * @brief Apply foreground filtering based on the mask file
   */
  absl::Status ApplyForegroundFilter() {
    // Detect the type of backend to use
    dlup::utilities::FileInfo file_info =
        dlup::utilities::DetectFileType(config_.mask_file);

    switch (file_info.file_type) {
      case dlup::utilities::FileType::kTiff:
        RETURN_IF_ERROR(ApplyTiffForegroundFilter(),
                        "Failed to apply TIFF foreground filter");
        break;
      case dlup::utilities::FileType::kXml:
        RETURN_IF_ERROR(ApplyXmlForegroundFilter(file_info),
                        "Failed to apply XML foreground filter");
        break;
      default:
        spdlog::error("Unknown mask type for {}", config_.mask_file.string());
        return MAKE_STATUS(absl::StatusCode::kInvalidArgument,
                           "Unknown mask type");
    }

    return absl::OkStatus();
  }

  /**
   * @brief Apply foreground filtering using a TIFF mask
   */
  absl::Status ApplyTiffForegroundFilter() {
    auto mask_backend = std::make_shared<dlup::backends::VipsSlide>(
        config_.mask_file, false, false, false);
    spdlog::info("Reading mask of type TIFF from {}",
                 config_.mask_file.string());

    foreground_result_ = dlup::Foreground<int>::FilterGrid(
        *grid_, *mask_backend, GetConfig()->tile_size, GetConfig()->mpp,
        config_.mask_threshold);

    dataset_->SetForegroundFilter(*foreground_result_);
    return absl::OkStatus();
  }

  /**
   * @brief Apply foreground filtering using an XML mask
   */
  absl::Status ApplyXmlForegroundFilter(
      const dlup::utilities::FileInfo& file_info) {
    if (file_info.metadata.at("root_tag") == "DlupAnnotations") {
      // TODO(jonasteuwen): For this we first need to check if
      // indeed Foreground can be filtered.
      // For this likely you need to make a circle in QuPath
      // and see if it actually works with the dummy models.
      std::string error_message = format(
          "Reading mask of type from {} failed. Not implemented. If this "
          "is an acute issue please open an issue at "
          "https://github.com/NKI-AI/aiforoncology-internal/ so we can "
          "give it more priority",
          config_.mask_file.string());
      spdlog::error(error_message);
      return MAKE_STATUS(absl::StatusCode::kUnimplemented, error_message);
    }
    spdlog::error(
        "Unknown XML mask type for {}. Got root tag {}, but expected "
        "DlupAnnotations.",
        config_.mask_file.string(), file_info.metadata.at("root_tag"));
    return MAKE_STATUS(absl::StatusCode::kInvalidArgument,
                       "Unknown XML mask type");
  }

  absl::Status RunInference(const fs::path& temp_directory) {
    auto writer = std::make_unique<aifo::data::writers::DiskTileWriter>(
        temp_directory, aifo::data::Metadata::Create()
                            ->SetMpp(GetConfig()->mpp)
                            ->SetGeometry(geometry_)
                            ->SetTileSize(GetConfig()->tile_size)
                            ->SetTileOverlap(GetConfig()->tile_overlap)
                            ->SetGridOrder(GRID_ORDER)
                            ->Lock());

    RETURN_IF_ERROR(writer->SetGrid(grid_), "Failed to set grid for writer");

    size_t total_batches =
        (dataset_->Length() + config_.batch_size - 1) / config_.batch_size;
    auto spinner = std::make_unique<aifocore::utilities::Spinner>();
    spinner->Start();

    auto batch_generator = [this, &spinner, total_batches](int batch_index)
        -> absl::StatusOr<std::vector<dlup::DatasetSample>> {
      std::vector<dlup::DatasetSample> batch_results;
      std::vector<torch::Tensor> batch_tensors;
      std::vector<aifocore::Size<int, 2>> coordinates;

      batch_tensors.reserve(config_.batch_size);
      coordinates.reserve(config_.batch_size);

      size_t start = batch_index * config_.batch_size;
      size_t end = std::min(start + config_.batch_size, dataset_->Length());

      // Collect batch inputs
      for (size_t index = start; index < end; ++index) {
        dlup::DatasetSample sample;
        ASSIGN_OR_RETURN(sample, dataset_->GetTile(index),
                         "Failed to get tile " + std::to_string(index));

        coordinates.push_back(sample.coordinates);
        spinner->SetText(format("Processing batch {}/{} (tile {})",
                                batch_index + 1, total_batches, start));

        torch::Tensor image_tensor;
        ASSIGN_OR_RETURN(image_tensor, VImageToTensor(*sample.tile, device_),
                         "Failed to convert tile");
        image_tensor = image_tensor.to(torch::kFloat).unsqueeze(0);

        image_tensor =
            NormalizeTensor(image_tensor, GetConfig()->normalization.mean,
                            GetConfig()->normalization.std);

        batch_tensors.push_back(image_tensor);
      }

      // Process batch
      if (!batch_tensors.empty()) {
        torch::Tensor batched_input = torch::cat(batch_tensors, 0);
        c10::IValue raw_output = model_->Infer(batched_input);
        auto processed_output = model_->ProcessOutput(raw_output);

        torch::Tensor output_tensor = processed_output.ToTensor();

        torch::Tensor transformed_output =
            pre_writer_transform_->Forward(output_tensor);

        // Convert results back to samples
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

    // Wrap the batch generator to handle StatusOr
    RETURN_IF_ERROR(writer->Consume([&batch_generator](int batch_index)
                                        -> std::vector<dlup::DatasetSample> {
      auto batch_result = batch_generator(batch_index);
      if (!batch_result.ok()) {
        // Propagate the error by throwing an exception that will be caught
        // by the Consume implementation
        throw std::runtime_error(batch_result.status().ToString());
      }
      return std::move(batch_result).value();
    }),
                    "Failed to consume batch data");

    // Get the size before closing
    auto total_size = writer->GetTotalSize();
    absl::Status close_status = writer->Close();
    if (!close_status.ok()) {
      spdlog::warn("Failed to close writer: {}", close_status.message());
    }

    spinner->Stop();

    auto duration = std::chrono::high_resolution_clock::now() - start_time;
    auto seconds = std::chrono::duration<double>(duration).count();

    spdlog::info("Inference completed in: {:.2f} seconds", seconds);
    spdlog::info("Total size of written data: {:.2f} MB",
                 static_cast<double>(total_size) / (1024 * 1024));
    return absl::OkStatus();
  }

  absl::Status ProcessResults(const fs::path& temp_directory) {
    spdlog::info("Stitching tiles...");
    aifo::data::readers::StitchingMode stitching_mode;
    std::shared_ptr<aifo::data::readers::PostStitchTransform> post_transform;

    const auto& merge_method = GetConfig()->merge_method;

    if (merge_method == "crop") {
      stitching_mode = aifo::data::readers::StitchingMode::kCrop;
      // TODO(jonasteuwen): Consider where this output stuff has to be set
      post_transform = std::make_shared<aifo::data::readers::IdentityTransform>(
          aifo::data::Metadata::Create()
              ->Set(MetadataKeys::NumOutputChannels, 1)
              ->Lock());
    } else if (merge_method == "average") {
      stitching_mode = aifo::data::readers::StitchingMode::kAverage;
      post_transform =
          std::make_shared<aifo::data::readers::AveragingTransform>(
              aifo::data::Metadata::Create()
                  ->Set(MetadataKeys::NumOutputChannels, 1)
                  ->Set(MetadataKeys::OutputPixelFormat, VIPS_FORMAT_UCHAR)
                  ->Set(MetadataKeys::OutputInterpretation,
                        VIPS_INTERPRETATION_B_W)
                  ->Lock());
    } else if (merge_method == "maximum") {
      return MAKE_STATUS(
          absl::StatusCode::kUnimplemented,
          "Maximum merge method is not implemented. Consider applying a filter "
          "to "
          "the tiles before merging to reduce noise and improve stability.");
    } else {
      return MAKE_STATUS(absl::StatusCode::kInvalidArgument,
                         "Unsupported merge method: " + merge_method);
    }

    auto reader = std::make_unique<aifo::data::readers::DiskTileReader>(
        temp_directory, stitching_mode, post_transform);
    RETURN_IF_ERROR(reader->Open(), "Failed to open reader");
    RETURN_IF_ERROR(
        aifocore::utilities::SaveVipsImageToFile(
            reader->GetImage(), config_.output_file,
            vips::VImage::option()
                ->set("tile_height", config_.tiff_tile_size)
                ->set("tile_width", config_.tiff_tile_size)
                ->set("xres", 1000.0 / GetConfig()->mpp)
                ->set("yres", 1000.0 / GetConfig()->mpp)
                ->set("tile", true)
                ->set("pyramid", true)
                ->set("compression", VIPS_FOREIGN_TIFF_COMPRESSION_LZW)),
        "Failed saving region to file");

    spdlog::info("Results saved to: {} (size: {}x{}x{} mpp: {})",
                 config_.output_file.string(), reader->GetImage().width(),
                 reader->GetImage().height(), reader->GetImage().bands(),
                 GetConfig()->mpp);

    // Generate and save the thumbnail if requested
    if (config_.create_thumbnail) {
      auto thumbnail_ptr = dataset_->GetSlide()->GetThumbnail({1024, 1024});
      if (!thumbnail_ptr) {
        return MAKE_STATUS(absl::StatusCode::kUnavailable,
                           "Failed to generate thumbnail.");
      }
      vips::VImage thumbnail = *thumbnail_ptr;

      // Read the segmentation map from the output file since the current reader
      // does a lot of reading of 1024 tiles and we don't
      // want to do that without caching.
      vips::VImage segmentation_map;
      try {
        segmentation_map = vips::VImage::new_from_file(
            config_.output_file.string().c_str(),
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
            aifocore::fmt::format(
                "Failed to read segmentation map from file: {}", e.what()));
      }

      vips::VImage lut = visualization::SegmentationVisualizer::CreateLut(
          config_impl_->labels);

      vips::VImage overlayed_thumbnail =
          visualization::SegmentationVisualizer::OverlaySegmentation(
              thumbnail, segmentation_map, lut, 0.6);

      // Save the thumbnail
      fs::path thumbnail_file =
          config_.output_file.replace_extension(".thumbnail.png");

      RETURN_IF_ERROR(aifocore::utilities::SaveVipsImageToFile(
                          overlayed_thumbnail, thumbnail_file, nullptr),
                      "Failed saving thumbnail to file");

      spdlog::info("Thumbnail saved to: {}", thumbnail_file.string());
    }

    return absl::OkStatus();
  }

  aifo::pathology::config::ProcessConfig config_;
  std::unique_ptr<aifo::utilities::TorchModel> model_;
  std::shared_ptr<dlup::SlideDataset> dataset_;
  std::shared_ptr<aifocore::tiling::Grid<int>> grid_;
  std::unique_ptr<AifoModelConfiguration> config_impl_;
  dlup::SlideGeometry geometry_;
  torch::Device device_;
  std::optional<dlup::ForegroundResult<int>> foreground_result_;
  std::shared_ptr<transforms::PreWriterTransform> pre_writer_transform_;
};

}  // namespace aifo::pathology::inference

using aifo::pathology::inference::ProcessImage;

int main(int argc, char* argv[]) {
  // TODO(jonasteuwen): There is a config module
  // that should be used for this
  spdlog::set_level(spdlog::level::info);  // Set global log level to info
  spdlog::set_pattern("[%Y-%m-%d %H:%M:%S.%e][%^%l%$] %v");  // Format logs

  auto config = aifo::pathology::config::ProcessConfig::FromArgs(argc, argv);

  aifocore::utilities::VipsInitializer vips(argv[0]);  // Initialize libvips

  // Use the factory method to create the processor
  auto processor_result = ProcessImage::Create(config);
  if (!processor_result.ok()) {
    spdlog::error("Error initializing processor: {}",
                  processor_result.status().message());
    return 1;
  }

  // Run the processing pipeline
  absl::Status status = processor_result.value().InferAndSave();
  if (!status.ok()) {
    spdlog::error("Error running inference: {}", status.message());
    return 1;
  }

  return 0;
}
