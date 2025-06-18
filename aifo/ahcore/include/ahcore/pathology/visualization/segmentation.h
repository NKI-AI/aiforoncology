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
#ifndef AIFO_AHCORE_INCLUDE_AHCORE_PATHOLOGY_VISUALIZATION_SEGMENTATION_H_
#define AIFO_AHCORE_INCLUDE_AHCORE_PATHOLOGY_VISUALIZATION_SEGMENTATION_H_

#include <vips/vips8>

#include <vector>
#include "ahcore/pathology/configuration.h"

namespace aifo::pathology::visualization {

/**
 * @brief Class handling visualization of segmentation results
 */
class SegmentationVisualizer {
 public:
  /**
     * @brief Creates a lookup table for segmentation visualization
     * 
     * @param labels Vector of labels with their associated colors
     * @return VImage containing the RGB lookup table
     */
  static vips::VImage CreateLut(const std::vector<Label>& labels);

  /**
     * @brief Overlays segmentation map on the original image
     * 
     * @param image Original image
     * @param segmentation_map Segmentation map to overlay
     * @return VImage with the segmentation overlay
     */
  static vips::VImage OverlaySegmentation(const vips::VImage& image,
                                          const vips::VImage& segmentation_map);

  /**
     * @brief Overlays segmentation map on the original image with specified alpha
     * 
     * @param image Original image
     * @param segmentation_map Segmentation map to overlay
     * @param color_lut Color lookup table to use
     * @param alpha Transparency value for the overlay
     * @return VImage with the segmentation overlay
     */
  static vips::VImage OverlaySegmentation(const vips::VImage& image,
                                          const vips::VImage& segmentation_map,
                                          const vips::VImage& color_lut,
                                          double alpha);
};

}  // namespace aifo::pathology::visualization

#endif  // AIFO_AHCORE_INCLUDE_AHCORE_PATHOLOGY_VISUALIZATION_SEGMENTATION_H_
