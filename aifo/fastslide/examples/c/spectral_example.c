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

#include <math.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <time.h>

#include <vips/vips.h>
#include "fastslide/c/fastslide.h"

// Configuration structure
typedef struct {
  char* slide_file;
  char* output_dir;
} Config;

// Helper function to create directory
int create_directory(const char* path) {
  struct stat st = {0};
  if (stat(path, &st) == -1) {
    if (mkdir(path, 0755) != 0) {
      perror("mkdir");
      return 0;
    }
  }
  return 1;
}

// Helper function to save RGB data as PNG using libvips
int save_rgb_as_png(const uint8_t* rgb_data, uint32_t width, uint32_t height,
                    const char* filename) {
  // Create VipsImage from RGB data
  VipsImage* image = vips_image_new_from_memory(
      (void*)rgb_data, width * height * 3, width, height, 3, VIPS_FORMAT_UCHAR);
  if (!image) {
    printf("Failed to create VipsImage from memory\n");
    return 0;
  }

  // Set interpretation to RGB
  vips_image_init_fields(image, width, height, 3, VIPS_FORMAT_UCHAR,
                         VIPS_CODING_NONE, VIPS_INTERPRETATION_sRGB, 1.0, 1.0);

  // Save as PNG
  if (vips_pngsave(image, filename, NULL) != 0) {
    printf("Failed to save PNG: %s\n", vips_error_buffer());
    g_object_unref(image);
    return 0;
  }

  g_object_unref(image);
  return 1;
}

// Helper function to get current time in milliseconds
long long get_time_ms(void) {
  struct timespec ts;
  clock_gettime(CLOCK_MONOTONIC, &ts);
  return ts.tv_sec * 1000LL + ts.tv_nsec / 1000000LL;
}

// Convert image to separate planar configuration if needed for resampling
FastSlideImage* ensure_separate_planar(FastSlideImage* image) {
  FastSlidePlanarConfig config = fastslide_image_get_planar_config(image);
  if (config == FASTSLIDE_PLANAR_CONFIG_SEPARATE) {
    // Already separate, return clone
    return fastslide_image_clone(image);
  }

  // Get image info
  FastSlideImageInfo info;
  if (!fastslide_image_get_info(image, &info)) {
    return NULL;
  }

  // Create new image with separate planar config
  FastSlideImage* separate_image = NULL;

  switch (info.format) {
    case FASTSLIDE_IMAGE_FORMAT_RGB:
      separate_image = fastslide_image_create_rgb(
          (FastSlideImageDimensions){info.width, info.height}, info.data_type);
      break;
    case FASTSLIDE_IMAGE_FORMAT_RGBA:
      separate_image = fastslide_image_create_rgba(
          (FastSlideImageDimensions){info.width, info.height}, info.data_type);
      break;
    case FASTSLIDE_IMAGE_FORMAT_GRAY:
      separate_image = fastslide_image_create_grayscale(
          (FastSlideImageDimensions){info.width, info.height}, info.data_type);
      break;
    case FASTSLIDE_IMAGE_FORMAT_SPECTRAL:
      separate_image = fastslide_image_create_spectral(
          (FastSlideImageDimensions){info.width, info.height}, info.channels,
          info.data_type);
      break;
    default:
      return NULL;
  }

  if (!separate_image) {
    return NULL;
  }

  // Copy data from original to separate image
  const uint8_t* src_data = fastslide_image_get_data(image);
  uint8_t* dst_data = fastslide_image_get_data_mutable(separate_image);

  if (!src_data || !dst_data) {
    fastslide_image_free(separate_image);
    return NULL;
  }

  // Manual conversion from interleaved to planar
  size_t pixels = (size_t)info.width * info.height;
  size_t bytes_per_sample = info.bytes_per_sample;

  for (uint32_t c = 0; c < info.channels; ++c) {
    for (size_t p = 0; p < pixels; ++p) {
      size_t src_offset = (p * info.channels + c) * bytes_per_sample;
      size_t dst_offset = (c * pixels + p) * bytes_per_sample;
      memcpy(dst_data + dst_offset, src_data + src_offset, bytes_per_sample);
    }
  }

  return separate_image;
}

// Helper function to parse command line arguments
Config parse_command_line(int argc, char** argv) {
  Config config = {0};

  if (argc < 3) {
    printf("Usage: %s --file <slide_file> [--output <output_dir>]\n", argv[0]);
    exit(1);
  }

  // Parse --file and --output arguments
  for (int i = 1; i < argc; i++) {
    if (strcmp(argv[i], "--file") == 0 && i + 1 < argc) {
      config.slide_file = argv[++i];
    } else if (strcmp(argv[i], "--output") == 0 && i + 1 < argc) {
      config.output_dir = argv[++i];
    }
  }

  if (!config.slide_file) {
    printf("Error: --file argument is required\n");
    printf("Usage: %s --file <slide_file> [--output <output_dir>]\n", argv[0]);
    exit(1);
  }

  if (!config.output_dir) {
    config.output_dir = "output";
  }

  return config;
}

// Helper function to print slide information
void print_slide_info(const FastSlideSlideReader* reader) {
  printf("=== Slide Information ===\n");

  // Get format name
  const char* format = fastslide_slide_reader_get_format_name(reader);
  if (format) {
    printf("Format: %s\n", format);
  }

  // Get level count
  int level_count = fastslide_slide_reader_get_level_count(reader);
  printf("Level count: %d\n", level_count);

  // Get base dimensions
  FastSlideImageDimensions base_dims;
  if (fastslide_slide_reader_get_base_dimensions(reader, &base_dims)) {
    printf("Base dimensions: %ux%u\n", base_dims.width, base_dims.height);
  }

  // Get slide properties
  FastSlideSlideProperties properties;
  if (fastslide_slide_reader_get_properties(reader, &properties)) {
    printf("MPP: %.3f x %.3f\n", properties.mpp_x, properties.mpp_y);
    printf("Objective magnification: %.1fx\n",
           properties.objective_magnification);
    if (properties.objective_name) {
      printf("Objective: %s\n", properties.objective_name);
    }
    if (properties.scanner_model) {
      printf("Scanner: %s\n", properties.scanner_model);
    }
    fastslide_slide_reader_free_properties(&properties);
  }

  printf("\n");
}

// Compute histograms for all channels
FastSlideHistogram** compute_channel_histograms(
    const FastSlideSlideReader* reader, int* num_histograms) {
  printf("Computing channel histograms...\n");

  // Get channel metadata
  FastSlideChannelMetadata* channel_metadata;
  int num_channels;
  if (!fastslide_slide_reader_get_channel_metadata(reader, &channel_metadata,
                                                   &num_channels)) {
    printf("  Failed to get channel metadata\n");
    return NULL;
  }

  if (num_channels == 0) {
    printf("  No channels found\n");
    fastslide_slide_reader_free_channel_metadata(channel_metadata,
                                                 num_channels);
    return NULL;
  }

  printf("  Found %d channels\n", num_channels);

  // Use the highest level number (lowest resolution) - same as C++ version
  int level_count = fastslide_slide_reader_get_level_count(reader);
  int selected_level = level_count - 1;
  uint32_t level_width = 0;
  uint32_t level_height = 0;

  FastSlideImageDimensions dims;
  if (fastslide_slide_reader_get_level_dimensions(reader, selected_level,
                                                  &dims)) {
    level_width = dims.width;
    level_height = dims.height;
  }

  printf("  Computing histograms from level %d (%ux%u)\n", selected_level,
         level_width, level_height);

  // Read the entire level as an image
  FastSlideImage* image = fastslide_slide_reader_read_region_coords(
      reader, 0, 0, level_width, level_height, selected_level);

  if (!image) {
    printf("  Failed to read image region: %s\n", fastslide_get_last_error());
    fastslide_slide_reader_free_channel_metadata(channel_metadata,
                                                 num_channels);
    return NULL;
  }

  printf("  Read image: ");
  FastSlideImageInfo image_info;
  if (fastslide_image_get_info(image, &image_info)) {
    printf("%ux%ux%u ", image_info.width, image_info.height,
           image_info.channels);
    switch (image_info.format) {
      case FASTSLIDE_IMAGE_FORMAT_SPECTRAL:
        printf("Spectral");
        break;
      case FASTSLIDE_IMAGE_FORMAT_RGB:
        printf("RGB");
        break;
      case FASTSLIDE_IMAGE_FORMAT_RGBA:
        printf("RGBA");
        break;
      case FASTSLIDE_IMAGE_FORMAT_GRAY:
        printf("Grayscale");
        break;
    }
    printf("\n");
  }

  // Use QuPath's histogram parameters: 1024 bins, full data range
  int n_bins = 1024;
  FastSlideHistogram** histograms;
  if (!fastslide_histogram_create_from_image_channels(
          image, n_bins, &histograms, num_histograms)) {
    printf("  Failed to create histograms: %s\n", fastslide_get_last_error());
    fastslide_image_free(image);
    fastslide_slide_reader_free_channel_metadata(channel_metadata,
                                                 num_channels);
    return NULL;
  }

  printf("  Created %d histograms\n", *num_histograms);

  // Cleanup
  fastslide_image_free(image);
  fastslide_slide_reader_free_channel_metadata(channel_metadata, num_channels);

  return histograms;
}

// Extract display ranges from histograms
void extract_display_ranges(FastSlideHistogram** histograms, int num_histograms,
                            double* min_values, double* max_values,
                            double saturation) {
  for (int i = 0; i < num_histograms; ++i) {
    FastSlideDisplayRange range;
    if (fastslide_histogram_compute_display_range(histograms[i], saturation,
                                                  &range)) {
      min_values[i] = range.min_value;
      max_values[i] = range.max_value;
    } else {
      // Fallback to edge values
      min_values[i] = fastslide_histogram_get_edge_min(histograms[i]);
      max_values[i] = fastslide_histogram_get_edge_max(histograms[i]);
    }
  }
}

// Print channel information
void print_channel_info(const FastSlideChannelMetadata* channel_metadata,
                        int num_channels, const double* min_values,
                        const double* max_values) {
  if (num_channels == 0) {
    printf("No channel information available\n");
    return;
  }

  printf("\nChannels:\n");
  for (int i = 0; i < num_channels; ++i) {
    const FastSlideChannelMetadata* ch = &channel_metadata[i];

    printf("  %2d: ", i + 1);

    // Show biomarker if available, otherwise use channel name
    if (ch->biomarker && strlen(ch->biomarker) > 0) {
      printf("%s", ch->biomarker);
      if (ch->name && strlen(ch->name) > 0 &&
          strcmp(ch->name, ch->biomarker) != 0) {
        printf(" (%s)", ch->name);
      }
    } else if (ch->name && strlen(ch->name) > 0) {
      printf("%s", ch->name);
    } else {
      printf("Channel_%d", i);
    }

    // Display range with nice formatting
    printf(" [%.1f-%.1f]\n", min_values[i], max_values[i]);
  }
}

// Save complete level for spectral analysis
int save_complete_level_for_spectral(
    const FastSlideSlideReader* reader, const char* output_dir,
    const FastSlideChannelMetadata* channel_metadata, int num_channels,
    const double* min_values, const double* max_values) {
  printf("Analyzing complete slide level...\n");

  // Find the first level under 5000 pixels on each side
  int level_count = fastslide_slide_reader_get_level_count(reader);
  int target_level = -1;
  uint32_t target_width = 0, target_height = 0;

  for (int level = 0; level < level_count; ++level) {
    FastSlideImageDimensions dims;
    if (fastslide_slide_reader_get_level_dimensions(reader, level, &dims)) {
      if (dims.width <= 5000 && dims.height <= 5000) {
        target_level = level;
        target_width = dims.width;
        target_height = dims.height;
        break;
      }
    }
  }

  if (target_level == -1) {
    target_level = level_count - 1;
    FastSlideImageDimensions dims;
    if (fastslide_slide_reader_get_level_dimensions(reader, target_level,
                                                    &dims)) {
      target_width = dims.width;
      target_height = dims.height;
    } else {
      printf("  Failed to get dimensions for fallback level\n");
      return 0;
    }
  }

  printf("  Level %d (%ux%u, ", target_level, target_width, target_height);

  // Read the region
  FastSlideImage* image = fastslide_slide_reader_read_region_coords(
      reader, 0, 0, target_width, target_height, target_level);

  if (!image) {
    printf("FAILED)\n");
    printf("  Failed to read region: %s\n", fastslide_get_last_error());
    return 0;
  }

  // Get channel count
  uint32_t channels = fastslide_image_get_channels(image);
  printf("%u channels) ", channels);

  // Check if the image is spectral
  FastSlideImageFormat format = fastslide_image_get_format(image);
  FastSlideImage* rgb_image = NULL;

  if (format != FASTSLIDE_IMAGE_FORMAT_SPECTRAL) {
    printf("→ RGB conversion");
    rgb_image = fastslide_image_to_rgb(image);
  } else {
    printf("→ spectral blending");

    // Use proper spectral blending if we have channel metadata and display ranges
    if (channel_metadata && num_channels > 0 && min_values && max_values) {
      // Create display ranges array
      FastSlideDisplayRange* display_ranges =
          malloc(num_channels * sizeof(FastSlideDisplayRange));
      if (display_ranges) {
        for (int i = 0; i < num_channels; ++i) {
          display_ranges[i].min_value = min_values[i];
          display_ranges[i].max_value = max_values[i];
        }

        // Use spectral blending
        rgb_image = fastslide_combine_spectral_channels_with_display_ranges(
            image, channel_metadata, num_channels, display_ranges,
            num_channels);

        free(display_ranges);

        if (!rgb_image) {
          printf(" (fallback)");
          // Fallback to basic RGB conversion
          rgb_image = fastslide_image_to_rgb(image);
        }
      } else {
        printf(" (fallback - malloc failed)");
        rgb_image = fastslide_image_to_rgb(image);
      }
    } else {
      printf(" (fallback - no metadata)");
      rgb_image = fastslide_image_to_rgb(image);
    }
  }

  if (!rgb_image) {
    printf(" → FAILED (RGB conversion)\n");
    fastslide_image_free(image);
    return 0;
  }

  // Get RGB image data
  FastSlideImageInfo rgb_info;
  if (!fastslide_image_get_info(rgb_image, &rgb_info)) {
    printf(" → FAILED (get RGB info)\n");
    fastslide_image_free(rgb_image);
    fastslide_image_free(image);
    return 0;
  }

  // Get RGB data from image
  const uint8_t* rgb_data = fastslide_image_get_data(rgb_image);
  if (!rgb_data) {
    printf(" → FAILED (get RGB data)\n");
    fastslide_image_free(rgb_image);
    fastslide_image_free(image);
    return 0;
  }

  // Create output filename
  char filename[512];
  snprintf(filename, sizeof(filename), "%s/level_%d.png", output_dir,
           target_level);

  // Save as PNG
  if (save_rgb_as_png(rgb_data, target_width, target_height, filename)) {
    printf(" → level_%d.png ✓\n", target_level);
  } else {
    printf(" → FAILED (PNG save)\n");
    fastslide_image_free(rgb_image);
    fastslide_image_free(image);
    return 0;
  }

  // Now work on the smaller version (like C++ example)
  printf("  Creating smaller version:\n");
  uint32_t new_width = target_width / 2;
  uint32_t new_height = target_height / 2;
  printf("    Resampling spectral data (%ux%u → %ux%u)", target_width,
         target_height, new_width, new_height);

  // Time the resampling operation - run 20 times for average like C++ version
  const int num_runs = 20;
  long long durations[20];
  FastSlideImage* resampled_spectral = NULL;

  // Ensure image has separate planar configuration for resampling
  FastSlideImage* separate_image = ensure_separate_planar(image);
  if (!separate_image) {
    printf(" → FAILED (separate planar conversion)\n");
    fastslide_image_free(rgb_image);
    fastslide_image_free(image);
    return 0;
  }

  for (int run = 0; run < num_runs; ++run) {
    long long start_time = get_time_ms();

    // Perform Lanczos resampling
    FastSlideImage* temp_resampled =
        fastslide_lanczos_resample(separate_image, new_width, new_height);

    long long end_time = get_time_ms();
    durations[run] = end_time - start_time;

    if (!temp_resampled) {
      printf(" → FAILED (resampling error: %s)\n", fastslide_get_last_error());
      fastslide_image_free(separate_image);
      fastslide_image_free(rgb_image);
      fastslide_image_free(image);
      return 0;
    }

    // Keep the last result for subsequent processing
    if (run == num_runs - 1) {
      resampled_spectral = temp_resampled;
    } else {
      fastslide_image_free(temp_resampled);
    }
  }

  // Compute average duration
  long long total_duration = 0;
  for (int i = 0; i < num_runs; ++i) {
    total_duration += durations[i];
  }
  long long average_duration = total_duration / num_runs;

  if (!resampled_spectral) {
    printf(" → FAILED (spectral resampling)\n");
    fastslide_image_free(separate_image);
    fastslide_image_free(rgb_image);
    fastslide_image_free(image);
    return 0;
  }

  printf(" ✓ (avg: %lld ms over %d runs)\n", average_duration, num_runs);

  printf("    Converting resampled spectral to RGB");
  // Convert resampled spectral to RGB
  FastSlideImage* resampled_image = NULL;
  if (channel_metadata && num_channels > 0 && min_values && max_values) {
    // Create display ranges array
    FastSlideDisplayRange* display_ranges =
        malloc(num_channels * sizeof(FastSlideDisplayRange));
    if (display_ranges) {
      for (int i = 0; i < num_channels; ++i) {
        display_ranges[i].min_value = min_values[i];
        display_ranges[i].max_value = max_values[i];
      }

      resampled_image = fastslide_combine_spectral_channels_with_display_ranges(
          resampled_spectral, channel_metadata, num_channels, display_ranges,
          num_channels);

      free(display_ranges);
    }

    if (!resampled_image) {
      printf(" (fallback)");
      resampled_image = fastslide_image_to_rgb(resampled_spectral);
    }
  } else {
    resampled_image = fastslide_image_to_rgb(resampled_spectral);
  }

  if (!resampled_image) {
    printf(" → FAILED\n");
    fastslide_image_free(resampled_spectral);
    fastslide_image_free(separate_image);
    fastslide_image_free(rgb_image);
    fastslide_image_free(image);
    return 0;
  }
  printf(" ✓\n");

  // Save smaller resampled PNG
  char smaller_filename[512];
  snprintf(smaller_filename, sizeof(smaller_filename),
           "%s/level_%d_smaller.png", output_dir, target_level);

  const uint8_t* resampled_rgb_data = fastslide_image_get_data(resampled_image);
  printf("    Saving %s", strrchr(smaller_filename, '/') + 1);
  if (save_rgb_as_png(resampled_rgb_data, new_width, new_height,
                      smaller_filename)) {
    printf(" ✓\n");
  } else {
    printf(" → FAILED\n");
    fastslide_image_free(resampled_image);
    fastslide_image_free(resampled_spectral);
    fastslide_image_free(separate_image);
    fastslide_image_free(rgb_image);
    fastslide_image_free(image);
    return 0;
  }

  // Cleanup all images
  fastslide_image_free(resampled_image);
  fastslide_image_free(resampled_spectral);
  fastslide_image_free(separate_image);
  fastslide_image_free(rgb_image);
  fastslide_image_free(image);
  return 1;
}

// Save associated images
void save_associated_images(const FastSlideSlideReader* reader,
                            const char* output_dir) {
  char** names;
  int count;

  if (!fastslide_slide_reader_get_associated_image_names(reader, &names,
                                                         &count)) {
    printf("  No associated images found\n");
    return;
  }

  if (count == 0) {
    printf("  No associated images found\n");
    fastslide_slide_reader_free_associated_image_names(names, count);
    return;
  }

  for (int i = 0; i < count; i++) {
    printf("  Processing %s", names[i]);

    // Read associated image
    FastSlideImage* image =
        fastslide_slide_reader_read_associated_image(reader, names[i]);
    if (!image) {
      printf(" → FAILED (read error)\n");
      continue;
    }

    // Convert to RGB
    FastSlideImage* rgb_image = fastslide_image_to_rgb(image);
    if (!rgb_image) {
      printf(" → FAILED (RGB conversion)\n");
      fastslide_image_free(image);
      continue;
    }

    // Get image info
    FastSlideImageInfo info;
    if (!fastslide_image_get_info(rgb_image, &info)) {
      printf(" → FAILED (get info)\n");
      fastslide_image_free(rgb_image);
      fastslide_image_free(image);
      continue;
    }

    // Get RGB data
    const uint8_t* rgb_data = fastslide_image_get_data(rgb_image);
    if (!rgb_data) {
      printf(" → FAILED (get data)\n");
      fastslide_image_free(rgb_image);
      fastslide_image_free(image);
      continue;
    }

    // Create output filename
    char filename[512];
    snprintf(filename, sizeof(filename), "%s/%s.png", output_dir, names[i]);

    // Save as PNG
    if (save_rgb_as_png(rgb_data, info.width, info.height, filename)) {
      printf(" → %s.png\n", names[i]);
    } else {
      printf(" → FAILED (PNG save)\n");
    }

    // Cleanup
    fastslide_image_free(rgb_image);
    fastslide_image_free(image);
  }

  fastslide_slide_reader_free_associated_image_names(names, count);
}

int main(int argc, char** argv) {
  Config config = parse_command_line(argc, argv);

  printf("FastSlide Spectral Example - C API\n");
  printf("Slide file: %s\n", config.slide_file);
  printf("Output directory: %s\n\n", config.output_dir);

  // Initialize libvips
  if (vips_init(argv[0]) != 0) {
    printf("Failed to initialize libvips\n");
    return 1;
  }

  // Create output directory
  if (!create_directory(config.output_dir)) {
    printf("Failed to create output directory: %s\n", config.output_dir);
    vips_shutdown();
    return 1;
  }

  // Initialize FastSlide
  if (!fastslide_initialize()) {
    printf("Failed to initialize FastSlide: %s\n", fastslide_get_last_error());
    vips_shutdown();
    return 1;
  }

  // Create slide reader
  FastSlideSlideReader* reader = fastslide_create_reader(config.slide_file);
  if (!reader) {
    printf("Failed to create reader: %s\n", fastslide_get_last_error());
    fastslide_cleanup();
    vips_shutdown();
    return 1;
  }

  // Print basic slide info
  print_slide_info(reader);

  // Compute histograms and display ranges
  printf("Analyzing spectral channels:\n");
  int num_histograms;
  FastSlideHistogram** histograms =
      compute_channel_histograms(reader, &num_histograms);

  // Declare variables for later use
  FastSlideChannelMetadata* channel_metadata = NULL;
  int num_channels = 0;
  double* min_values = NULL;
  double* max_values = NULL;

  if (histograms && num_histograms > 0) {
    // Extract display ranges
    double saturation = 0.001;
    min_values = malloc(num_histograms * sizeof(double));
    max_values = malloc(num_histograms * sizeof(double));

    extract_display_ranges(histograms, num_histograms, min_values, max_values,
                           saturation);

    // Get channel metadata for display
    if (fastslide_slide_reader_get_channel_metadata(reader, &channel_metadata,
                                                    &num_channels)) {
      print_channel_info(channel_metadata, num_channels, min_values,
                         max_values);
    }

    // Cleanup histograms (but keep min_values, max_values, and channel_metadata)
    fastslide_histogram_free_array(histograms, num_histograms);
  }

  printf("\nSaving images:\n");

  // Save main level
  if (!save_complete_level_for_spectral(reader, config.output_dir,
                                        channel_metadata, num_channels,
                                        min_values, max_values)) {
    printf("  ✗ Failed to save main level\n");
  }

  // Cleanup the saved variables
  if (channel_metadata) {
    fastslide_slide_reader_free_channel_metadata(channel_metadata,
                                                 num_channels);
  }
  free(min_values);
  free(max_values);

  // Save associated images
  save_associated_images(reader, config.output_dir);

  // Cleanup
  fastslide_slide_reader_free(reader);
  fastslide_cleanup();
  vips_shutdown();

  printf("\n✓ All images saved to %s\n", config.output_dir);
  return 0;
}