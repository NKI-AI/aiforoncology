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

#include <algorithm>
#include <chrono>
#include <filesystem>
#include <iomanip>
#include <iostream>
#include <memory>
#include <string>
#include <utility>
#include <vector>

#include "absl/status/status.h"
#include "fastslide/histogram.h"
#include "fastslide/image.h"
#include "fastslide/readers/readers.h"
#include "fastslide/resample/lanczos.h"
#include "fastslide/slide_reader.h"
#include "fastslide/utilities/combine.h"
#include "utils/example_utils.h"

namespace {

using fastslide::Image;
using fastslide::ImageFormat;
using fastslide::RegionSpec;
using fastslide::SlideReader;
using fastslide::SlideReaderRegistry;
using fastslide::examples::Config;
using fastslide::examples::SaveAsPNG;

/**
 * @brief Computes and exports histograms for each channel using the highest level (like QuPath)
 * 
 * This function creates histograms for each channel of the slide image using the same
 * approach as the statistics computation - analyzing the highest level (lowest resolution)
 * for efficiency while maintaining QuPath compatibility.
 * 
 * @param reader The slide reader to analyze
 * @param output_dir Output directory for CSV exports
 * @return Vector of histograms for each channel, or empty vector on failure
 */
std::vector<fastslide::Histogram> ComputeChannelHistograms(
    const SlideReader& reader, const std::filesystem::path& output_dir) {
  // Get channel metadata
  auto channel_metadata = reader.GetChannelMetadata();
  if (channel_metadata.empty()) {
    std::cout << "  No channel metadata found\n";
    return {};
  }

  // Use the highest level number (lowest resolution) - same as statistics
  int selected_level = reader.GetLevelCount() - 1;

  auto level_info_or = reader.GetLevelInfo(selected_level);
  if (!level_info_or.ok()) {
    std::cout << "  Failed to get level info for histogram computation\n";
    return {};
  }

  const auto& level_info = level_info_or.value();
  uint32_t level_width = level_info.dimensions[0];
  uint32_t level_height = level_info.dimensions[1];

  std::cout << "  Computing histograms from level " << selected_level << " ("
            << level_width << "x" << level_height << ")\n";

  // Read the entire level as an image
  RegionSpec region{.top_left = {0, 0},
                    .size = {level_width, level_height},
                    .level = selected_level};

  auto image_result = reader.ReadRegion(region);
  if (!image_result.ok()) {
    std::cout << "  Failed to read region for histogram computation\n";
    return {};
  }

  const auto& image = image_result.value();

  // Use QuPath's histogram parameters: 1024 bins, full data range
  const int n_bins = 1024;

  // Create histograms for each channel
  auto histograms_result =
      fastslide::CreateHistogramsFromImageChannels(image, n_bins);

  if (!histograms_result.ok()) {
    std::cout << "  Failed to create histograms\n";
    return {};
  }

  const auto& histograms = histograms_result.value();
  std::cout << "  Created " << histograms.size() << " histograms\n";

  return histograms;
}

/**
 * @brief Extract QuPath display ranges from computed histograms
 * 
 * @param histograms Vector of pre-computed histograms
 * @param saturation Saturation percentage (default: 0.001 = 0.1%)
 * @return Vector of {min, max} display range pairs
 */
std::vector<std::pair<double, double>> ExtractDisplayRanges(
    const std::vector<fastslide::Histogram>& histograms,
    double saturation = 0.001) {
  std::vector<std::pair<double, double>> ranges;
  ranges.reserve(histograms.size());

  for (const auto& histogram : histograms) {
    ranges.push_back(histogram.ComputeDisplayRange(saturation));
  }

  return ranges;
}

/**
 * @brief Print detailed channel information with biomarkers and display ranges
 * 
 * @param channel_metadata Channel metadata from the slide
 * @param display_ranges Display ranges for each channel
 */
void PrintChannelInfo(
    const std::vector<fastslide::ChannelMetadata>& channel_metadata,
    const std::vector<std::pair<double, double>>& display_ranges) {
  if (channel_metadata.empty() || display_ranges.empty()) {
    std::cout << "No channel information available\n";
    return;
  }

  std::cout << "\nChannels:\n";
  for (size_t i = 0; i < channel_metadata.size() && i < display_ranges.size();
       ++i) {
    const auto& meta = channel_metadata[i];
    const auto& range = display_ranges[i];

    std::cout << "  " << std::setw(2) << i + 1 << ": ";

    // Show biomarker if available, otherwise use channel name
    if (!meta.biomarker.empty()) {
      std::cout << meta.biomarker;
      if (!meta.name.empty() && meta.name != meta.biomarker) {
        std::cout << " (" << meta.name << ")";
      }
    } else if (!meta.name.empty()) {
      std::cout << meta.name;
    } else {
      std::cout << "Channel_" << i;
    }

    // Display range with nice formatting
    std::cout << " [" << std::fixed << std::setprecision(1) << range.first
              << "-" << range.second << "]\n";
  }
}

/**
 * @brief Save the highest resolution level that fits under 5000 pixels for spectral slides
 * 
 * @param reader The slide reader
 * @param output_dir Output directory for the PNG file
 * @param channel_metadata Channel metadata for spectral conversion
 * @param display_ranges Pre-computed display ranges for spectral channels
 * @return absl::Status indicating success or failure
 */
absl::Status SaveCompleteLevelForSpectral(
    const SlideReader& reader, const std::filesystem::path& output_dir,
    const std::vector<fastslide::ChannelMetadata>& channel_metadata,
    const std::vector<std::pair<double, double>>& display_ranges) {
  // Find the first level under 5000 pixels on each side
  int target_level = -1;
  uint32_t target_width = 0, target_height = 0;

  for (int level = 0; level < reader.GetLevelCount(); ++level) {
    auto level_info_or = reader.GetLevelInfo(level);
    if (!level_info_or.ok())
      continue;

    const auto& level_info = level_info_or.value();
    uint32_t width = level_info.dimensions[0];
    uint32_t height = level_info.dimensions[1];

    if (width < 3000 && height < 3000 && target_level == -1) {
      target_level = level;
      target_width = width;
      target_height = height;
      break;
    }
  }

  if (target_level == -1) {
    target_level = reader.GetLevelCount() - 1;
    auto level_info_or = reader.GetLevelInfo(target_level);
    if (level_info_or.ok()) {
      target_width = level_info_or.value().dimensions[0];
      target_height = level_info_or.value().dimensions[1];
    }
  }

  std::cout << "  Level " << target_level << " (" << target_width << "x"
            << target_height << ", " << std::flush;

  // Read the full region
  RegionSpec region{.top_left = {0, 0},
                    .size = {target_width, target_height},
                    .level = target_level};

  auto image_result = reader.ReadRegion(region);
  if (!image_result.ok()) {
    std::cout << "FAILED)\n";
    return absl::Status(image_result.status().code(),
                        "Failed to read region: " +
                            std::string(image_result.status().message()));
  }

  const auto& image = image_result.value();

  // Show image properties
  std::cout << image.GetChannels() << " channels) ["
            << fastslide::GetName(image.GetFormat()) << ", "
            << fastslide::GetName(image.GetDataType()) << ", "
            << fastslide::GetName(image.GetPlanarConfig()) << "] ";

  // Exit if the image is not spectral
  if (image.GetFormat() != ImageFormat::kSpectral) {
    std::cout << " → FAILED (not spectral)\n";
    return absl::Status(absl::StatusCode::kInternal, "Image is not spectral");
  }

  // Note: Spectral images are automatically created with interleaved format
  // (PlanarConfig::kContiguous) for optimal performance in pixel-wise
  // operations, regardless of the TIFF file's internal layout

  if (channel_metadata.empty() || display_ranges.empty()) {
    std::cout << " → FAILED (no channel metadata or display ranges)\n";
    return absl::Status(absl::StatusCode::kInternal,
                        "No channel metadata or display ranges");
  }

  std::cout << "→ spectral blending (original)";
  // Convert original spectral to RGB first
  std::unique_ptr<Image> rgb_image =
      fastslide::utils::CombineSpectralChannelsWithDisplayRanges(
          image, channel_metadata, display_ranges);
  if (!rgb_image) {
    std::cout << " (fallback)";
    rgb_image = image.ToRGB();
  }

  if (!rgb_image) {
    std::cout << " → FAILED\n";
    return absl::Status(absl::StatusCode::kInternal,
                        "Failed to convert original spectral image to RGB");
  }

  // Save original size PNG first
  std::string level_name = "level_" + std::to_string(target_level);
  std::filesystem::path filename = output_dir / (level_name + ".png");

  std::cout << " → saving " << filename.filename().string();
  if (!SaveAsPNG(*rgb_image, filename.string())) {
    std::cout << " → FAILED (original)\n";
    return absl::Status(absl::StatusCode::kInternal,
                        "Failed to save original PNG");
  }
  std::cout << " ✓\n";

  // Now work on the smaller version
  std::cout << "  Creating smaller version:\n";
  uint32_t new_width = image.GetWidth() / 2;
  uint32_t new_height = image.GetHeight() / 2;
  std::cout << "    Resampling spectral data (" << image.GetWidth() << "x"
            << image.GetHeight() << " → " << new_width << "x" << new_height
            << ")";

  // Time the resampling operation specifically - run 10 times for average
  constexpr int num_runs = 20;
  std::vector<std::chrono::milliseconds> durations;
  durations.reserve(num_runs);

  std::unique_ptr<Image> resampled_spectral;

  for (int run = 0; run < num_runs; ++run) {
    auto start_time = std::chrono::high_resolution_clock::now();

    try {
      auto temp_resampled =
          fastslide::resample::LanczosResample(image, new_width, new_height);

      // Keep the last result for subsequent processing
      if (run == num_runs - 1) {
        resampled_spectral = std::move(temp_resampled);
      }
    } catch (const std::invalid_argument& e) {
      std::cout << " → FAILED (resampling error: " << e.what() << ")\n";
      return absl::Status(absl::StatusCode::kInvalidArgument,
                          "Resampling failed: " + std::string(e.what()));
    }

    auto end_time = std::chrono::high_resolution_clock::now();
    auto duration = std::chrono::duration_cast<std::chrono::milliseconds>(
        end_time - start_time);
    durations.push_back(duration);
  }

  // Compute average duration
  auto total_duration = std::chrono::milliseconds::zero();
  for (const auto& d : durations) {
    total_duration += d;
  }
  auto average_duration = total_duration / num_runs;

  if (!resampled_spectral) {
    std::cout << " → FAILED (spectral resampling)\n";
    return absl::Status(absl::StatusCode::kInternal,
                        "Failed to resample spectral image");
  }

  std::cout << " ✓ (avg: " << average_duration.count() << " ms over "
            << num_runs << " runs)\n";

  std::cout << "    Converting resampled spectral to RGB";
  // Convert resampled spectral to RGB
  std::unique_ptr<Image> resampled_image =
      fastslide::utils::CombineSpectralChannelsWithDisplayRanges(
          *resampled_spectral, channel_metadata, display_ranges);
  if (!resampled_image) {
    std::cout << " (fallback)";
    resampled_image = resampled_spectral->ToRGB();
  }

  if (!resampled_image) {
    std::cout << " → FAILED\n";
    return absl::Status(absl::StatusCode::kInternal,
                        "Failed to convert resampled spectral image to RGB");
  }
  std::cout << " ✓\n";

  // Save smaller resampled PNG
  std::string smaller_level_name =
      "level_" + std::to_string(target_level) + "_smaller";
  std::filesystem::path smaller_filename =
      output_dir / (smaller_level_name + ".png");

  std::cout << "    Saving " << smaller_filename.filename().string();
  if (SaveAsPNG(*resampled_image, smaller_filename.string())) {
    std::cout << " ✓\n";
    return absl::OkStatus();
  } else {
    std::cout << " → FAILED\n";
    return absl::Status(absl::StatusCode::kInternal,
                        "Failed to save smaller PNG");
  }
}

/**
 * @brief Save all associated images for spectral slides with detailed information
 * 
 * @param reader The slide reader
 * @param output_dir Output directory for PNG files
 */
void SaveAssociatedImages(const SlideReader& reader,
                          const std::filesystem::path& output_dir) {
  auto associated_names = reader.GetAssociatedImageNames();

  if (associated_names.empty()) {
    std::cout << "  No associated images found\n";
    return;
  }

  std::cout << "  Found " << associated_names.size()
            << " associated image(s):\n";

  for (const auto& name : associated_names) {
    std::cout << "    " << name;

    // Get dimensions first
    auto dims_result = reader.GetAssociatedImageDimensions(name);
    if (!dims_result.ok()) {
      std::cout << " → FAILED (dimensions error)\n";
      continue;
    }

    const auto& dims = dims_result.value();
    std::cout << " (" << dims[0] << "x" << dims[1] << ")";

    // Read the image
    auto image_result = reader.ReadAssociatedImage(name);
    if (!image_result.ok()) {
      std::cout << " → FAILED (read error: " << image_result.status().message()
                << ")\n";
      continue;
    }

    const auto& image = image_result.value();

    // Show image properties
    std::cout << " [" << fastslide::GetName(image.GetFormat()) << ", "
              << fastslide::GetName(image.GetDataType()) << ", "
              << fastslide::GetName(image.GetPlanarConfig()) << "]";

    // Convert to RGB if needed
    std::unique_ptr<Image> rgb_image;
    if (image.GetFormat() == ImageFormat::kRGB) {
      rgb_image = image.Clone();
    } else {
      rgb_image = image.ToRGB();
    }

    if (!rgb_image) {
      std::cout << " → FAILED (RGB conversion)\n";
      continue;
    }

    // Create filename with format info
    std::string safe_name = name;
    std::replace(safe_name.begin(), safe_name.end(), ' ', '_');
    std::replace(safe_name.begin(), safe_name.end(), '/', '_');

    std::filesystem::path filename = output_dir / (safe_name + ".png");

    if (SaveAsPNG(*rgb_image, filename.string())) {
      std::cout << " → " << filename.filename().string() << "\n";
    } else {
      std::cout << " → FAILED (save error)\n";
    }
  }
}

}  // namespace

int main(int argc, char** argv) {
  auto config = fastslide::examples::ParseCliAndInit(
      argc, argv,
      "FastSlide Spectral Example - Advanced multiplex/spectral slide "
      "processing");

  fastslide::examples::InitializeVips(argv);

  // Initialize slide readers
  fastslide::InitializeReaders();

  // Create reader
  auto reader_or =
      SlideReaderRegistry::GetInstance().CreateReader(config.slide_file);
  if (!reader_or.ok()) {
    std::cerr << "Failed to create reader: " << reader_or.status().message()
              << "\n";
    fastslide::examples::CleanupVips();
    return 1;
  }

  auto reader = std::move(reader_or.value());

  // Print basic slide info
  fastslide::examples::PrintSlideInfo(*reader);

  // Compute histograms and display ranges
  std::cout << "\nAnalyzing spectral channels:\n";
  auto histograms = ComputeChannelHistograms(*reader, config.output_dir);
  auto display_ranges = ExtractDisplayRanges(histograms);

  // Get channel metadata
  auto channel_metadata = reader->GetChannelMetadata();

  // Show detailed channel information
  PrintChannelInfo(channel_metadata, display_ranges);

  std::cout << "\nSaving images:\n";

  // Save main level with spectral processing
  auto status = SaveCompleteLevelForSpectral(*reader, config.output_dir,
                                             channel_metadata, display_ranges);
  if (!status.ok()) {
    std::cout << "  ✗ " << status.message() << "\n";
  }

  // Save all associated images with detailed information
  std::cout << "\nProcessing associated images:\n";
  SaveAssociatedImages(*reader, config.output_dir);

  std::cout << "\n✓ All images saved to " << config.output_dir << "\n";

  fastslide::examples::CleanupVips();
  return 0;
}
