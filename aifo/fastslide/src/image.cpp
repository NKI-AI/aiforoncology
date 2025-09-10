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

#include "fastslide/image.h"

#include <algorithm>
#include <cstring>
#include <memory>
#include <type_traits>
#include <vector>

namespace fastslide {

namespace {

/// @brief Convert single pixel to RGB using weighted average for grayscale
/// @tparam T Data type
/// @param src_data Source pixel data
/// @param dst_data Destination pixel data
/// @param src_channels Number of source channels
/// @param pixel_index Pixel index
/// @param src_config Source planar configuration
/// @param dst_config Destination planar configuration
/// @param image_dims Image dimensions [width, height]
template <typename T>
void ConvertPixelToRGB(const T* src_data, T* dst_data, uint32_t src_channels,
                       size_t pixel_index, PlanarConfig src_config,
                       PlanarConfig dst_config,
                       const ImageDimensions& image_dims) {
  // Helper function to get pixel value based on planar config
  auto get_pixel = [&](size_t channel) -> T {
    if (src_config == PlanarConfig::kContiguous) {
      return src_data[pixel_index * src_channels + channel];
    } else {
      size_t pixels_per_channel =
          static_cast<size_t>(image_dims[0]) * image_dims[1];
      return src_data[channel * pixels_per_channel + pixel_index];
    }
  };

  // Helper function to set pixel value based on planar config
  auto set_pixel = [&](size_t channel, T value) {
    if (dst_config == PlanarConfig::kContiguous) {
      dst_data[pixel_index * 3 + channel] = value;
    } else {
      size_t pixels_per_channel =
          static_cast<size_t>(image_dims[0]) * image_dims[1];
      dst_data[channel * pixels_per_channel + pixel_index] = value;
    }
  };

  if (src_channels == 1) {
    // Grayscale to RGB
    T gray_value = get_pixel(0);
    set_pixel(0, gray_value);  // R
    set_pixel(1, gray_value);  // G
    set_pixel(2, gray_value);  // B
  } else if (src_channels == 3) {
    // RGB to RGB (copy)
    set_pixel(0, get_pixel(0));  // R
    set_pixel(1, get_pixel(1));  // G
    set_pixel(2, get_pixel(2));  // B
  } else if (src_channels == 4) {
    // RGBA to RGB (drop alpha)
    set_pixel(0, get_pixel(0));  // R
    set_pixel(1, get_pixel(1));  // G
    set_pixel(2, get_pixel(2));  // B
  } else {
    // Spectral/hyperspectral to RGB - use first 3 channels or average
    if (src_channels >= 3) {
      set_pixel(0, get_pixel(0));  // R
      set_pixel(1, get_pixel(1));  // G
      set_pixel(2, get_pixel(2));  // B
    } else {
      // Less than 3 channels - replicate first channel
      T value = get_pixel(0);
      set_pixel(0, value);  // R
      set_pixel(1, value);  // G
      set_pixel(2, value);  // B
    }
  }
}

/// @brief Convert single pixel to grayscale using luminance formula
/// @tparam T Data type
/// @param src_data Source pixel data
/// @param dst_data Destination pixel data
/// @param src_channels Number of source channels
/// @param pixel_index Pixel index
/// @param src_config Source planar configuration
/// @param dst_config Destination planar configuration
/// @param image_dims Image dimensions [width, height]
template <typename T>
void ConvertPixelToGrayscale(const T* src_data, T* dst_data,
                             uint32_t src_channels, size_t pixel_index,
                             PlanarConfig src_config, PlanarConfig dst_config,
                             const ImageDimensions& image_dims) {
  // Helper function to get pixel value based on planar config
  auto get_pixel = [&](size_t channel) -> T {
    if (src_config == PlanarConfig::kContiguous) {
      return src_data[pixel_index * src_channels + channel];
    } else {
      size_t pixels_per_channel =
          static_cast<size_t>(image_dims[0]) * image_dims[1];
      return src_data[channel * pixels_per_channel + pixel_index];
    }
  };

  // Helper function to set pixel value based on planar config
  auto set_pixel = [&](size_t channel, T value) {
    if (dst_config == PlanarConfig::kContiguous) {
      dst_data[pixel_index] = value;
    } else {
      size_t pixels_per_channel =
          static_cast<size_t>(image_dims[0]) * image_dims[1];
      dst_data[channel * pixels_per_channel + pixel_index] = value;
    }
  };

  if (src_channels == 1) {
    // Already grayscale
    set_pixel(0, get_pixel(0));
  } else if (src_channels == 3) {
    // RGB to grayscale using luminance formula: 0.299*R + 0.587*G + 0.114*B
    T r = get_pixel(0);
    T g = get_pixel(1);
    T b = get_pixel(2);

    if constexpr (std::is_floating_point_v<T>) {
      set_pixel(0, static_cast<T>(0.299 * r + 0.587 * g + 0.114 * b));
    } else {
      // For integer types, use integer arithmetic to avoid precision loss
      set_pixel(0, static_cast<T>((299 * r + 587 * g + 114 * b) / 1000));
    }
  } else if (src_channels == 4) {
    // RGBA to grayscale (ignore alpha)
    T r = get_pixel(0);
    T g = get_pixel(1);
    T b = get_pixel(2);

    if constexpr (std::is_floating_point_v<T>) {
      set_pixel(0, static_cast<T>(0.299 * r + 0.587 * g + 0.114 * b));
    } else {
      set_pixel(0, static_cast<T>((299 * r + 587 * g + 114 * b) / 1000));
    }
  } else {
    // Spectral/hyperspectral to grayscale - take average of all channels
    T sum = T{0};
    for (uint32_t c = 0; c < src_channels; ++c) {
      sum += get_pixel(c);
    }
    set_pixel(0, sum / static_cast<T>(src_channels));
  }
}

/// @brief Template helper for data type conversion dispatch
/// @tparam Function Function type
/// @param src_image Source image
/// @param dst_image Destination image
/// @param func Function to call with typed pointers
template <typename Function>
void DispatchDataType(const Image& src_image, Image& dst_image,
                      Function&& func) {
  DataType dtype = src_image.GetDataType();

  DispatchByDataType(dtype, [&]<typename T>() {
    func(src_image.GetDataAs<T>(), dst_image.GetDataAs<T>());
  });
}

}  // namespace

std::unique_ptr<Image> Image::ToRGB() const {
  if (Empty()) {
    return nullptr;
  }

  // If already RGB, return a copy
  if (format_ == ImageFormat::kRGB) {
    return Clone();
  }

  // Create new RGB image with same dimensions and data type
  auto rgb_image =
      std::make_unique<Image>(dimensions_, ImageFormat::kRGB, dtype_);

  size_t pixel_count = GetPixelCount();

  DispatchDataType(
      *this, *rgb_image, [&](const auto* src_data, auto* dst_data) {
        for (size_t i = 0; i < pixel_count; ++i) {
          ConvertPixelToRGB(src_data, dst_data, channels_, i, planar_config_,
                            rgb_image->planar_config_, dimensions_);
        }
      });

  return rgb_image;
}

std::unique_ptr<Image> Image::ToGrayscale() const {
  if (Empty()) {
    return nullptr;
  }

  // If already grayscale, return a copy
  if (format_ == ImageFormat::kGray) {
    return Clone();
  }

  // Create new grayscale image with same dimensions and data type
  auto gray_image =
      std::make_unique<Image>(dimensions_, ImageFormat::kGray, dtype_);

  size_t pixel_count = GetPixelCount();

  DispatchDataType(
      *this, *gray_image, [&](const auto* src_data, auto* dst_data) {
        for (size_t i = 0; i < pixel_count; ++i) {
          ConvertPixelToGrayscale(src_data, dst_data, channels_, i,
                                  planar_config_, gray_image->planar_config_,
                                  dimensions_);
        }
      });

  return gray_image;
}

std::unique_ptr<Image> Image::ExtractChannels(
    const std::vector<uint32_t>& channel_indices) const {
  if (Empty() || channel_indices.empty()) {
    return nullptr;
  }

  // Validate channel indices
  for (uint32_t idx : channel_indices) {
    if (idx >= channels_) {
      throw std::out_of_range("Channel index out of bounds");
    }
  }

  uint32_t new_channels = static_cast<uint32_t>(channel_indices.size());

  // Determine new format
  ImageFormat new_format;
  if (new_channels == 1) {
    new_format = ImageFormat::kGray;
  } else if (new_channels == 3) {
    new_format = ImageFormat::kRGB;
  } else if (new_channels == 4) {
    new_format = ImageFormat::kRGBA;
  } else {
    new_format = ImageFormat::kSpectral;
  }

  // Create new image
  std::unique_ptr<Image> extracted_image;
  if (new_format == ImageFormat::kSpectral) {
    extracted_image = std::make_unique<Image>(dimensions_, new_channels, dtype_,
                                              planar_config_);
  } else {
    extracted_image = std::make_unique<Image>(dimensions_, new_format, dtype_,
                                              planar_config_);
  }

  size_t pixel_count = GetPixelCount();

  DispatchDataType(
      *this, *extracted_image, [&](const auto* src_data, auto* dst_data) {
        for (size_t pixel = 0; pixel < pixel_count; ++pixel) {
          for (size_t new_ch = 0; new_ch < channel_indices.size(); ++new_ch) {
            uint32_t src_ch = channel_indices[new_ch];

            size_t src_idx, dst_idx;
            if (planar_config_ == PlanarConfig::kContiguous) {
              // Interleaved layout
              src_idx = pixel * channels_ + src_ch;
              dst_idx = pixel * new_channels + new_ch;
            } else {
              // Planar layout
              src_idx = src_ch * pixel_count + pixel;
              dst_idx = new_ch * pixel_count + pixel;
            }

            dst_data[dst_idx] = src_data[src_idx];
          }
        }
      });

  return extracted_image;
}

std::unique_ptr<Image> Image::ToInterleaved() const {
  return ConvertMemoryFormat(PlanarConfig::kContiguous);
}

std::unique_ptr<Image> Image::ToPlanar() const {
  return ConvertMemoryFormat(PlanarConfig::kSeparate);
}

std::unique_ptr<Image> Image::ConvertMemoryFormat(
    PlanarConfig target_config) const {
  if (Empty()) {
    return nullptr;
  }

  // If already in target format, return a copy (Google C++ way)
  if (planar_config_ == target_config) {
    return Clone();
  }

  // Create new image with target layout
  std::unique_ptr<Image> converted_image;
  if (format_ == ImageFormat::kSpectral) {
    converted_image =
        std::make_unique<Image>(dimensions_, channels_, dtype_, target_config);
  } else {
    converted_image =
        std::make_unique<Image>(dimensions_, format_, dtype_, target_config);
  }

  const size_t pixel_count = GetPixelCount();

  // Special case: single channel - just copy data (already optimal layout)
  if (channels_ == 1) {
    std::memcpy(converted_image->GetData(), GetData(), SizeBytes());
    return converted_image;
  }

  // Dispatch based on data type for optimal vectorization
  DispatchDataType(
      *this, *converted_image, [&](const auto* src_data, auto* dst_data) {
        if (target_config == PlanarConfig::kContiguous) {
          // Convert from Separate to Contig (planar to interleaved)
          // Vectorization-friendly: process each channel as contiguous block
          for (uint32_t ch = 0; ch < channels_; ++ch) {
            const auto* src_channel_start = src_data + ch * pixel_count;
            for (size_t pixel = 0; pixel < pixel_count; ++pixel) {
              dst_data[pixel * channels_ + ch] = src_channel_start[pixel];
            }
          }
        } else {
          // Convert from Contig to Separate (interleaved to planar)
          // Vectorization-friendly: write each channel as contiguous block
          for (uint32_t ch = 0; ch < channels_; ++ch) {
            auto* dst_channel_start = dst_data + ch * pixel_count;
            for (size_t pixel = 0; pixel < pixel_count; ++pixel) {
              dst_channel_start[pixel] = src_data[pixel * channels_ + ch];
            }
          }
        }
      });

  return converted_image;
}

void Image::Paste(const Image& source_image, uint32_t dest_x, uint32_t dest_y,
                  uint32_t source_x, uint32_t source_y, uint32_t source_width,
                  uint32_t source_height) {
  if (source_image.Empty()) {
    return;  // Nothing to paste
  }

  // Handle auto-initialization for blank images
  if (!initialized_) {
    if (dimensions_[0] == 0 || dimensions_[1] == 0) {
      throw std::invalid_argument(
          "Cannot paste into image with zero dimensions");
    }

    // Auto-initialize with source image properties
    format_ = source_image.format_;
    dtype_ = source_image.dtype_;
    channels_ = source_image.channels_;
    planar_config_ = source_image.planar_config_;
    bytes_per_sample_ = source_image.bytes_per_sample_;

    // Allocate memory
    size_t total_size =
        dimensions_[0] * dimensions_[1] * channels_ * bytes_per_sample_;
    data_.resize(total_size, 0);  // Zero-initialize for clean background
    initialized_ = true;
  }

  if (Empty()) {
    return;  // Still empty after potential initialization
  }

  // Validate compatibility
  if (dtype_ != source_image.dtype_) {
    throw std::invalid_argument("Data types must match for pasting");
  }
  if (channels_ != source_image.channels_) {
    throw std::invalid_argument("Channel counts must match for pasting");
  }
  if (planar_config_ != source_image.planar_config_) {
    throw std::invalid_argument("Planar configurations must match for pasting");
  }

  // Determine source region dimensions
  uint32_t src_width =
      (source_width == 0) ? source_image.dimensions_[0] : source_width;
  uint32_t src_height =
      (source_height == 0) ? source_image.dimensions_[1] : source_height;

  // Validate source region bounds
  if (source_x >= source_image.dimensions_[0] ||
      source_y >= source_image.dimensions_[1]) {
    return;  // Source coordinates out of bounds
  }

  // Clip source region to source image bounds
  src_width = std::min(src_width, source_image.dimensions_[0] - source_x);
  src_height = std::min(src_height, source_image.dimensions_[1] - source_y);

  // Early exit if destination is completely outside this image
  if (dest_x >= dimensions_[0] || dest_y >= dimensions_[1]) {
    return;
  }

  // Clip to destination image bounds
  uint32_t copy_width = std::min(src_width, dimensions_[0] - dest_x);
  uint32_t copy_height = std::min(src_height, dimensions_[1] - dest_y);

  if (copy_width == 0 || copy_height == 0) {
    return;  // Nothing to copy after clipping
  }

  DispatchDataType(
      /* source = */ source_image,
      /* dest   = */ *this, [&](const auto* src_data, auto* dst_data) {
        // T is the per-sample type
        using T =
            std::remove_const_t<std::remove_pointer_t<decltype(src_data)>>;
        constexpr size_t sample_size = sizeof(T);

        const uint32_t source_width = source_image.GetWidth();
        const uint32_t source_height = source_image.GetHeight();
        const uint32_t dest_width = dimensions_[0];
        // planar_config_, channels_ are members of *this (the dest)

        if (planar_config_ == PlanarConfig::kContiguous) {
          // interleaved: copy each scanline in one go
          size_t row_bytes = copy_width * channels_ * sample_size;
          for (uint32_t row = 0; row < copy_height; ++row) {
            const T* src_ptr =
                src_data +
                ((source_y + row) * source_width + source_x) * channels_;
            T* dst_ptr =
                dst_data + ((dest_y + row) * dest_width + dest_x) * channels_;
            std::memcpy(dst_ptr, src_ptr, row_bytes);
          }

        } else {
          // planar: copy each channel/scanline
          size_t src_plane_pixels =
              static_cast<size_t>(source_width) * source_height;
          size_t dst_plane_pixels =
              static_cast<size_t>(dest_width) * dimensions_[1];
          size_t row_bytes = copy_width * sample_size;

          for (uint32_t ch = 0; ch < channels_; ++ch) {
            const T* src_plane = src_data + ch * src_plane_pixels +
                                 source_y * source_width + source_x;
            T* dst_plane =
                dst_data + ch * dst_plane_pixels + dest_y * dest_width + dest_x;

            for (uint32_t row = 0; row < copy_height; ++row) {
              std::memcpy(dst_plane + row * dest_width,
                          src_plane + row * source_width, row_bytes);
            }
          }
        }
      });
}

}  // namespace fastslide
