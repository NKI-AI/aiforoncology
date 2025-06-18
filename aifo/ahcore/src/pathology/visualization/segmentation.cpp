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
#include <vector>

#include "ahcore/pathology/visualization/segmentation.h"

namespace aifo::pathology::visualization {

vips::VImage SegmentationVisualizer::CreateLut(
    const std::vector<Label>& labels) {

  static std::vector<unsigned char> lut_data;
  // Index 0: Black (0, 0, 0)
  lut_data.insert(lut_data.end(), {0, 0, 0});  // RGB for background (black)

  // For each label, add RGB bytes
  for (const auto& label : labels) {
    auto [r, g, b] = label.RgbColor();
    lut_data.insert(lut_data.end(), {static_cast<unsigned char>(r),
                                     static_cast<unsigned char>(g),
                                     static_cast<unsigned char>(b)});
  }

  // Create and return the RGB-only color LUT
  return vips::VImage::new_from_memory(lut_data.data(), lut_data.size(),
                                       labels.size() + 1, 1, 3,
                                       VIPS_FORMAT_UCHAR);
}

vips::VImage SegmentationVisualizer::OverlaySegmentation(
    const vips::VImage& image, const vips::VImage& segmentation_map,
    const vips::VImage& color_lut, double alpha) {
  // Convert input image to RGBA
  vips::VImage rgba_image = image.colourspace(VIPS_INTERPRETATION_sRGB);
  if (rgba_image.bands() == 3) {
    rgba_image = rgba_image.bandjoin(255);  // Add alpha channel
  }

  // Apply color mapping
  vips::VImage seg_index = segmentation_map.cast(VIPS_FORMAT_UCHAR);
  // Apply the color LUT to get an RGB mask
  vips::VImage colored_mask = seg_index.maplut(color_lut);
  vips::VImage alpha_mask = (seg_index != 0) * alpha;
  colored_mask = colored_mask.bandjoin(alpha_mask);

  // Ensure segmentation mask has sRGB interpretation before blending
  if (colored_mask.interpretation() == VIPS_INTERPRETATION_MULTIBAND) {
    colored_mask = colored_mask.copy(vips::VImage::option()->set(
        "interpretation", VIPS_INTERPRETATION_sRGB));
  }

  // Resize to match input dimensions
  colored_mask = colored_mask.resize(
      static_cast<double>(rgba_image.width()) / colored_mask.width(),
      vips::VImage::option()->set(
          "vscale",
          static_cast<double>(rgba_image.height()) / colored_mask.height()));
  return rgba_image.composite2(
      colored_mask, VIPS_BLEND_MODE_OVER,
      vips::VImage::option()->set("premultiplied", false));
}

}  // namespace aifo::pathology::visualization
