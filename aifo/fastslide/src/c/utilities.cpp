// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of FastSlide.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the FastSlide project root.

#include "fastslide/c/utilities.h"

#include <cstring>
#include <vips/vips8>

#include <memory>
#include <string>
#include <utility>
#include <vector>

#include "fastslide/c/histogram.h"
#include "fastslide/c/image.h"
#include "fastslide/c/slide_reader.h"
#include "fastslide/image.h"
#include "fastslide/resample/average.h"
#include "fastslide/resample/lanczos.h"
#include "fastslide/slide_reader.h"
#include "fastslide/utilities/colors.h"
#include "fastslide/utilities/combine.h"

// Forward declarations for external functions
extern "C" void fastslide_set_last_error(const char* message);
extern "C" FastSlideImage* fastslide_image_create_from_cpp(
    fastslide::Image image);

// Forward declaration for function in image.cpp
const fastslide::Image& fastslide_image_get_cpp_image(
    const FastSlideImage* image);

namespace {

void SetLastError(const char* message) {
  fastslide_set_last_error(message);
}

// Helper to get FastSlideImage internal image
const fastslide::Image& GetCppImage(const FastSlideImage* image) {
  return fastslide_image_get_cpp_image(image);
}

// Convert C channel metadata to C++ channel metadata
std::vector<fastslide::ChannelMetadata> ConvertChannelMetadata(
    const FastSlideChannelMetadata* c_metadata, int num_channels) {
  std::vector<fastslide::ChannelMetadata> cpp_metadata;
  cpp_metadata.reserve(num_channels);

  for (int i = 0; i < num_channels; ++i) {
    const auto& c_ch = c_metadata[i];
    fastslide::ChannelMetadata cpp_ch;

    if (c_ch.name) {
      cpp_ch.name = std::string(c_ch.name);
    }

    if (c_ch.biomarker) {
      cpp_ch.biomarker = std::string(c_ch.biomarker);
    }

    // Convert color
    cpp_ch.color =
        fastslide::ColorRGB{c_ch.color.r, c_ch.color.g, c_ch.color.b};

    cpp_ch.exposure_time = c_ch.exposure_time;
    cpp_ch.signal_units = c_ch.signal_units;

    cpp_metadata.push_back(std::move(cpp_ch));
  }

  return cpp_metadata;
}

// Convert C display ranges to C++ display ranges
std::vector<std::pair<double, double>> ConvertDisplayRanges(
    const FastSlideDisplayRange* c_ranges, int num_ranges) {
  std::vector<std::pair<double, double>> cpp_ranges;
  cpp_ranges.reserve(num_ranges);

  for (int i = 0; i < num_ranges; ++i) {
    cpp_ranges.emplace_back(c_ranges[i].min_value, c_ranges[i].max_value);
  }

  return cpp_ranges;
}

}  // namespace

FastSlideImage* fastslide_combine_spectral_channels_with_display_ranges(
    const FastSlideImage* image,
    const FastSlideChannelMetadata* channel_metadata, int num_channels,
    const FastSlideDisplayRange* display_ranges, int num_ranges) {

  if (!image) {
    SetLastError("image cannot be null");
    return nullptr;
  }

  if (!channel_metadata || num_channels <= 0) {
    SetLastError(
        "channel_metadata cannot be null and num_channels must be positive");
    return nullptr;
  }

  try {
    const auto& cpp_image = GetCppImage(image);

    // Check if image is spectral
    if (cpp_image.GetFormat() != fastslide::ImageFormat::kSpectral) {
      SetLastError("image must be spectral format");
      return nullptr;
    }

    // Convert channel metadata
    auto cpp_channel_metadata =
        ConvertChannelMetadata(channel_metadata, num_channels);

    // Convert display ranges if provided
    std::vector<std::pair<double, double>> cpp_display_ranges;
    if (display_ranges && num_ranges > 0) {
      cpp_display_ranges = ConvertDisplayRanges(display_ranges, num_ranges);
    }

    // Call the C++ function
    auto result = fastslide::utils::CombineSpectralChannelsWithDisplayRanges(
        cpp_image, cpp_channel_metadata, cpp_display_ranges);

    if (!result) {
      SetLastError("failed to combine spectral channels");
      return nullptr;
    }

    return fastslide_image_create_from_cpp(std::move(*result));

  } catch (const std::exception& e) {
    SetLastError(e.what());
    return nullptr;
  }
}

FastSlideImage* fastslide_lanczos_resample(const FastSlideImage* image,
                                           uint32_t output_width,
                                           uint32_t output_height) {
  if (!image) {
    SetLastError("image cannot be null");
    return nullptr;
  }

  if (output_width == 0 || output_height == 0) {
    SetLastError("output dimensions must be positive");
    return nullptr;
  }

  try {
    const auto& cpp_image = GetCppImage(image);

    // Check if image has separate planar configuration
    if (cpp_image.GetPlanarConfig() != fastslide::PlanarConfig::kSeparate) {
      SetLastError("image must have separate planar configuration");
      return nullptr;
    }

    // Call the C++ LanczosResample function
    auto result = fastslide::resample::LanczosResample(cpp_image, output_width,
                                                       output_height);
    if (!result) {
      SetLastError("failed to resample image with Lanczos3");
      return nullptr;
    }

    return fastslide_image_create_from_cpp(std::move(*result));

  } catch (const std::exception& e) {
    SetLastError(e.what());
    return nullptr;
  }
}

FastSlideImage* fastslide_lanczos2_resample(const FastSlideImage* image,
                                            uint32_t output_width,
                                            uint32_t output_height) {
  if (!image) {
    SetLastError("image cannot be null");
    return nullptr;
  }

  if (output_width == 0 || output_height == 0) {
    SetLastError("output dimensions must be positive");
    return nullptr;
  }

  try {
    const auto& cpp_image = GetCppImage(image);

    // Check if image has separate planar configuration
    if (cpp_image.GetPlanarConfig() != fastslide::PlanarConfig::kSeparate) {
      SetLastError("image must have separate planar configuration");
      return nullptr;
    }

    // Call the C++ Lanczos2Resample function
    auto result = fastslide::resample::Lanczos2Resample(cpp_image, output_width,
                                                        output_height);
    if (!result) {
      SetLastError("failed to resample image with Lanczos2");
      return nullptr;
    }

    return fastslide_image_create_from_cpp(std::move(*result));

  } catch (const std::exception& e) {
    SetLastError(e.what());
    return nullptr;
  }
}

FastSlideImage* fastslide_cosine_resample(const FastSlideImage* image,
                                          uint32_t output_width,
                                          uint32_t output_height) {
  if (!image) {
    SetLastError("image cannot be null");
    return nullptr;
  }

  if (output_width == 0 || output_height == 0) {
    SetLastError("output dimensions must be positive");
    return nullptr;
  }

  try {
    const auto& cpp_image = GetCppImage(image);

    // Check if image has separate planar configuration
    if (cpp_image.GetPlanarConfig() != fastslide::PlanarConfig::kSeparate) {
      SetLastError("image must have separate planar configuration");
      return nullptr;
    }

    // Call the C++ CosineResample function
    auto result = fastslide::resample::CosineResample(cpp_image, output_width,
                                                      output_height);
    if (!result) {
      SetLastError("failed to resample image with Cosine windowed sinc");
      return nullptr;
    }

    return fastslide_image_create_from_cpp(std::move(*result));

  } catch (const std::exception& e) {
    SetLastError(e.what());
    return nullptr;
  }
}

FastSlideImage* fastslide_average_resample(const FastSlideImage* image,
                                           uint32_t factor) {
  if (!image) {
    SetLastError("image cannot be null");
    return nullptr;
  }

  if (factor == 0) {
    SetLastError("factor must be greater than 0");
    return nullptr;
  }

  try {
    const auto& cpp_image = GetCppImage(image);

    // Call the C++ AverageResample function
    auto result = fastslide::resample::AverageResample(cpp_image, factor);
    if (!result) {
      SetLastError("failed to resample image with average downsampling");
      return nullptr;
    }

    return fastslide_image_create_from_cpp(std::move(*result));

  } catch (const std::exception& e) {
    SetLastError(e.what());
    return nullptr;
  }
}

FastSlideImage* fastslide_average_2x2_resample(const FastSlideImage* image) {
  if (!image) {
    SetLastError("image cannot be null");
    return nullptr;
  }

  try {
    const auto& cpp_image = GetCppImage(image);

    // Check if image has separate planar configuration
    if (cpp_image.GetPlanarConfig() != fastslide::PlanarConfig::kSeparate) {
      SetLastError("image must have separate planar configuration");
      return nullptr;
    }

    // Call the C++ Average2x2Resample function
    auto result = fastslide::resample::Average2x2Resample(cpp_image);
    if (!result) {
      SetLastError("failed to resample image with 2x2 average downsampling");
      return nullptr;
    }

    return fastslide_image_create_from_cpp(std::move(*result));

  } catch (const std::exception& e) {
    SetLastError(e.what());
    return nullptr;
  }
}

FastSlideImage* fastslide_average_4x4_resample(const FastSlideImage* image) {
  if (!image) {
    SetLastError("image cannot be null");
    return nullptr;
  }

  try {
    const auto& cpp_image = GetCppImage(image);

    // Check if image has separate planar configuration
    if (cpp_image.GetPlanarConfig() != fastslide::PlanarConfig::kSeparate) {
      SetLastError("image must have separate planar configuration");
      return nullptr;
    }

    // Call the C++ Average4x4Resample function
    auto result = fastslide::resample::Average4x4Resample(cpp_image);
    if (!result) {
      SetLastError("failed to resample image with 4x4 average downsampling");
      return nullptr;
    }

    return fastslide_image_create_from_cpp(std::move(*result));

  } catch (const std::exception& e) {
    SetLastError(e.what());
    return nullptr;
  }
}

FastSlideImage* fastslide_average_8x8_resample(const FastSlideImage* image) {
  if (!image) {
    SetLastError("image cannot be null");
    return nullptr;
  }

  try {
    const auto& cpp_image = GetCppImage(image);

    // Check if image has separate planar configuration
    if (cpp_image.GetPlanarConfig() != fastslide::PlanarConfig::kSeparate) {
      SetLastError("image must have separate planar configuration");
      return nullptr;
    }

    // Call the C++ Average8x8Resample function
    auto result = fastslide::resample::Average8x8Resample(cpp_image);
    if (!result) {
      SetLastError("failed to resample image with 8x8 average downsampling");
      return nullptr;
    }

    return fastslide_image_create_from_cpp(std::move(*result));

  } catch (const std::exception& e) {
    SetLastError(e.what());
    return nullptr;
  }
}

FastSlideImage* fastslide_combine_spectral_channels_with_histograms(
    const FastSlideImage* image,
    const FastSlideChannelMetadata* channel_metadata, int num_channels,
    FastSlideHistogram** histograms, int num_histograms, double saturation) {

  if (!image) {
    SetLastError("image cannot be null");
    return nullptr;
  }

  if (!channel_metadata || num_channels <= 0) {
    SetLastError(
        "channel_metadata cannot be null and num_channels must be positive");
    return nullptr;
  }

  if (!histograms || num_histograms <= 0) {
    SetLastError(
        "histograms cannot be null and num_histograms must be positive");
    return nullptr;
  }

  try {
    // Compute display ranges from histograms
    std::vector<FastSlideDisplayRange> c_display_ranges;
    c_display_ranges.reserve(num_histograms);

    for (int i = 0; i < num_histograms; ++i) {
      FastSlideDisplayRange range;
      if (!fastslide_histogram_compute_display_range(histograms[i], saturation,
                                                     &range)) {
        SetLastError("failed to compute display range from histogram");
        return nullptr;
      }
      c_display_ranges.push_back(range);
    }

    // Use the first function with the computed display ranges
    return fastslide_combine_spectral_channels_with_display_ranges(
        image, channel_metadata, num_channels, c_display_ranges.data(),
        static_cast<int>(c_display_ranges.size()));

  } catch (const std::exception& e) {
    SetLastError(e.what());
    return nullptr;
  }
}

extern "C" void fastslide_examples_initialize_vips(void) {
  static const char* dummy_argv = "fastslide_c";
  if (VIPS_INIT(dummy_argv)) {
    vips_error_exit(nullptr);
  }
}

extern "C" void fastslide_examples_cleanup_vips(void) {
  vips_shutdown();
}

extern "C" int fastslide_examples_save_as_png(const FastSlideImage* image,
                                              const char* filename) {
  if (!image || !filename) {
    SetLastError("image and filename cannot be null");
    return 0;
  }

  try {
    const auto& cpp_image = GetCppImage(image);

    if (cpp_image.GetWidth() == 0 || cpp_image.GetHeight() == 0) {
      SetLastError("Invalid image dimensions");
      return 0;
    }

    // Ensure it's uint8 RGB format
    if (cpp_image.GetDataType() != fastslide::DataType::kUInt8) {
      SetLastError("Expected uint8 RGB(A) image for PNG output");
      return 0;
    }

    if (cpp_image.GetFormat() != fastslide::ImageFormat::kRGB &&
        cpp_image.GetFormat() != fastslide::ImageFormat::kRGBA) {
      SetLastError("Expected RGB or RGBA image for PNG output");
      return 0;
    }

    // Create VipsImage from raw data
    vips::VImage vips_image = vips::VImage::new_from_memory(
        const_cast<void*>(static_cast<const void*>(cpp_image.GetData())),
        cpp_image.SizeBytes(), cpp_image.GetWidth(), cpp_image.GetHeight(),
        cpp_image.GetChannels(), VIPS_FORMAT_UCHAR);

    vips_image.write_to_file(filename);
    return 1;

  } catch (const vips::VError& e) {
    std::string error_msg = "Failed to save PNG: " + std::string(e.what());
    SetLastError(error_msg.c_str());
    return 0;
  } catch (const std::exception& e) {
    SetLastError(e.what());
    return 0;
  }
}

extern "C" FastSlideImage* fastslide_examples_load_from_png(
    const char* filename) {
  if (!filename) {
    SetLastError("filename cannot be null");
    return nullptr;
  }

  try {
    // Load PNG using vips
    vips::VImage vips_image = vips::VImage::new_from_file(filename);

    // Ensure it's in RGB format
    if (vips_image.bands() != 3) {
      // Convert to RGB if not already
      if (vips_image.bands() == 1) {
        // Grayscale to RGB
        vips_image = vips_image.bandjoin({vips_image, vips_image});
      } else if (vips_image.bands() == 4) {
        // RGBA to RGB (drop alpha)
        vips_image =
            vips_image.extract_band(0, vips::VImage::option()->set("n", 3));
      }
    }

    // Ensure it's uint8
    if (vips_image.format() != VIPS_FORMAT_UCHAR) {
      vips_image = vips_image.cast(VIPS_FORMAT_UCHAR);
    }

    // Get image properties
    int width = vips_image.width();
    int height = vips_image.height();
    int bands = vips_image.bands();

    // Create FastSlide image
    auto cpp_image = fastslide::CreateRGBImage(
        {static_cast<uint32_t>(width), static_cast<uint32_t>(height)},
        fastslide::DataType::kUInt8);

    // Copy data from vips to FastSlide image
    size_t data_size = width * height * bands;
    auto* vips_data = static_cast<const uint8_t*>(vips_image.data());
    auto* fastslide_data = cpp_image->GetDataAs<uint8_t>();

    std::memcpy(fastslide_data, vips_data, data_size);

    return fastslide_image_create_from_cpp(std::move(*cpp_image));

  } catch (const vips::VError& e) {
    std::string error_msg = "Failed to load PNG: " + std::string(e.what());
    SetLastError(error_msg.c_str());
    return nullptr;
  } catch (const std::exception& e) {
    SetLastError(e.what());
    return nullptr;
  }
}
