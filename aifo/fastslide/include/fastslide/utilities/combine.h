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

#ifndef AIFO_FASTSLIDE_INCLUDE_FASTSLIDE_UTILITIES_COMBINE_H_
#define AIFO_FASTSLIDE_INCLUDE_FASTSLIDE_UTILITIES_COMBINE_H_

#include <cstdint>
#include <memory>
#include <vector>

#include "fastslide/histogram.h"
#include "fastslide/image.h"
#include "fastslide/slide_reader.h"
#include "fastslide/utilities/colors.h"

namespace fastslide::utils {

/**
 * @brief Combines multichannel data using channel colors with additive blending
 * 
 * Takes individual channel data (already converted to uint8) and combines them
 * using their respective colors with additive blending to create an RGB image.
 * 
 * @param channel_data Vector of channel data, each as vector of uint8 values
 * @param colors Vector of RGB colors for each channel
 * @param width Image width
 * @param height Image height
 * @return Combined RGB image
 */
Image CombineChannelsWithColors(
    const std::vector<std::vector<uint8_t>>& channel_data,
    const std::vector<ColorRGB>& colors, uint32_t width, uint32_t height);

/**
 * @brief Combines spectral image channels using pre-computed display ranges
 * 
 * This is the core, simplified function that does the actual channel combination
 * using additive color blending. If display_ranges is empty, uses the full data 
 * range for each channel.
 * 
 * @param image Input spectral image (should be in interleaved format for optimal performance)
 * @param channel_metadata Metadata containing channel names and colors
 * @param display_ranges Pre-computed display ranges {min, max} for each channel (empty = use full range)
 * @return Combined RGB image, or nullptr if conversion fails
 * @note This function works with both interleaved and planar formats but performs best with interleaved spectral images
 */
std::unique_ptr<Image> CombineSpectralChannelsWithDisplayRanges(
    const Image& image, const std::vector<ChannelMetadata>& channel_metadata,
    const std::vector<std::pair<double, double>>& display_ranges = {});

}  // namespace fastslide::utils

#endif  // AIFO_FASTSLIDE_INCLUDE_FASTSLIDE_UTILITIES_COMBINE_H_
