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

#include "fastslide/utilities/combine.h"

#include <algorithm>
#include <iomanip>
#include <iostream>
#include <limits>
#include <memory>
#include <utility>
#include <vector>

#include "fastslide/histogram.h"

namespace fastslide {
namespace utils {

// Template helper to apply a function to pixel
// values regardless of underlying data type
template <typename Func>
auto ApplyToPixel(const Image& image, uint32_t y, uint32_t x, uint32_t ch,
                  Func&& func) {
  switch (image.GetDataType()) {
    case DataType::kUInt8:
      return func(static_cast<double>(image.At<uint8_t>(y, x, ch)));
    case DataType::kUInt16:
      return func(static_cast<double>(image.At<uint16_t>(y, x, ch)));
    case DataType::kInt16:
      return func(static_cast<double>(image.At<int16_t>(y, x, ch)));
    case DataType::kUInt32:
      return func(static_cast<double>(image.At<uint32_t>(y, x, ch)));
    case DataType::kInt32:
      return func(static_cast<double>(image.At<int32_t>(y, x, ch)));
    case DataType::kFloat32:
      return func(static_cast<double>(image.At<float>(y, x, ch)));
    case DataType::kFloat64:
      return func(image.At<double>(y, x, ch));
  }
  // Should never reach here if all DataType cases are covered
  return func(0.0);
}

Image CombineChannelsWithColors(
    const std::vector<std::vector<uint8_t>>& channel_data,
    const std::vector<ColorRGB>& colors, uint32_t width, uint32_t height) {

  Image result(ImageDimensions{width, height}, ImageFormat::kRGB,
               DataType::kUInt8);

  // Initialize result to black
  std::fill(result.GetDataVector().begin(), result.GetDataVector().end(), 0);

  // Combine channels using additive blending
  for (size_t ch = 0; ch < channel_data.size() && ch < colors.size(); ++ch) {
    if (channel_data[ch].empty()) {
      continue;
    }

    const auto& color = colors[ch];
    const uint8_t* data = channel_data[ch].data();

    for (uint32_t y = 0; y < height; ++y) {
      for (uint32_t x = 0; x < width; ++x) {
        size_t src_idx = y * width + x;
        if (src_idx >= channel_data[ch].size()) {
          continue;
        }

        uint8_t intensity = data[src_idx];
        if (intensity == 0) {
          continue;  // Skip black pixels for efficiency
        }

        // Blend with channel color using additive blending
        float factor = intensity / 255.0f;
        size_t pixel_idx = (y * width + x) * 3;

        // Add color contribution, clamping to prevent overflow
        result.GetDataVector()[pixel_idx + 0] = std::min(
            255, static_cast<int>(result.GetDataVector()[pixel_idx + 0] +
                                  color.r * factor));
        result.GetDataVector()[pixel_idx + 1] = std::min(
            255, static_cast<int>(result.GetDataVector()[pixel_idx + 1] +
                                  color.g * factor));
        result.GetDataVector()[pixel_idx + 2] = std::min(
            255, static_cast<int>(result.GetDataVector()[pixel_idx + 2] +
                                  color.b * factor));
      }
    }
  }

  return result;
}

std::unique_ptr<Image> CombineSpectralChannelsWithDisplayRanges(
    const Image& image, const std::vector<ChannelMetadata>& channel_metadata,
    const std::vector<std::pair<double, double>>& display_ranges) {

  if (image.GetFormat() != ImageFormat::kSpectral || channel_metadata.empty()) {
    return nullptr;
  }

  bool use_display_ranges =
      !display_ranges.empty() && display_ranges.size() >= image.GetChannels();

  // Prepare channel data and colors
  std::vector<std::vector<uint8_t>> channel_data;
  std::vector<ColorRGB> colors;
  channel_data.reserve(image.GetChannels());

  for (uint32_t ch = 0; ch < image.GetChannels(); ++ch) {
    std::vector<uint8_t> ch_data;
    ch_data.reserve(image.GetWidth() * image.GetHeight());

    // Determine range for this channel
    double channel_min = 0.0;
    double channel_max = 0.0;
    if (use_display_ranges && ch < display_ranges.size()) {
      channel_min = display_ranges[ch].first;
      channel_max = display_ranges[ch].second;
    } else {
      // Compute min/max from actual data if no display ranges provided
      channel_min = std::numeric_limits<double>::max();
      channel_max = std::numeric_limits<double>::lowest();

      for (uint32_t y = 0; y < image.GetHeight(); ++y) {
        for (uint32_t x = 0; x < image.GetWidth(); ++x) {
          double raw_value =
              ApplyToPixel(image, y, x, ch, [](double val) { return val; });
          channel_min = std::min(channel_min, raw_value);
          channel_max = std::max(channel_max, raw_value);
        }
      }
    }

    // Calculate range, avoiding division by zero
    double range = channel_max - channel_min;
    if (range == 0.0) {
      range = 1.0;
    }

    // Extract channel data with display range clipping and convert to uint8
    for (uint32_t y = 0; y < image.GetHeight(); ++y) {
      for (uint32_t x = 0; x < image.GetWidth(); ++x) {
        double raw_value =
            ApplyToPixel(image, y, x, ch, [](double val) { return val; });

        // Apply display range clipping:
        // stretch [channel_min, channel_max] to [0, 255]
        double normalized = (raw_value - channel_min) / range;
        uint8_t value =
            static_cast<uint8_t>(std::clamp(normalized * 255.0, 0.0, 255.0));
        ch_data.push_back(value);
      }
    }
    channel_data.push_back(std::move(ch_data));

    // Add channel color
    if (ch < channel_metadata.size()) {
      colors.push_back(channel_metadata[ch].color);
    } else {
      colors.push_back(GetDefaultChannelColor(static_cast<int>(ch)));
    }
  }

  // Combine channels using additive color blending
  Image combined_image = CombineChannelsWithColors(
      channel_data, colors, image.GetWidth(), image.GetHeight());

  return std::make_unique<Image>(std::move(combined_image));
}

}  // namespace utils
}  // namespace fastslide
