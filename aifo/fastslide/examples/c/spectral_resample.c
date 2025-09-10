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
  int num_runs;
  double downsample_factor;
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

// Helper function to parse command line arguments
Config parse_command_line(int argc, char** argv) {
  Config config = {0};
  config.num_runs = 20;            // Default number of timing runs
  config.downsample_factor = 2.0;  // Default downsample factor

  if (argc < 3) {
    printf(
        "Usage: %s --file <slide_file> [--output <output_dir>] [--runs <num>] "
        "[--downsample <factor>]\n",
        argv[0]);
    exit(1);
  }

  // Parse arguments
  for (int i = 1; i < argc; i++) {
    if (strcmp(argv[i], "--file") == 0 && i + 1 < argc) {
      config.slide_file = argv[++i];
    } else if (strcmp(argv[i], "--output") == 0 && i + 1 < argc) {
      config.output_dir = argv[++i];
    } else if (strcmp(argv[i], "--runs") == 0 && i + 1 < argc) {
      config.num_runs = atoi(argv[++i]);
      if (config.num_runs <= 0)
        config.num_runs = 20;
    } else if (strcmp(argv[i], "--downsample") == 0 && i + 1 < argc) {
      config.downsample_factor = atof(argv[++i]);
      if (config.downsample_factor <= 1.0)
        config.downsample_factor = 2.0;
    }
  }

  if (!config.slide_file) {
    printf("Error: --file argument is required\n");
    printf(
        "Usage: %s --file <slide_file> [--output <output_dir>] [--runs <num>] "
        "[--downsample <factor>]\n",
        argv[0]);
    exit(1);
  }

  if (!config.output_dir) {
    config.output_dir = "output_resample";
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

// Convert image to separate planar configuration if needed
FastSlideImage* ensure_separate_planar(FastSlideImage* image) {
  FastSlidePlanarConfig config = fastslide_image_get_planar_config(image);
  if (config == FASTSLIDE_PLANAR_CONFIG_SEPARATE) {
    // Already separate, return clone
    return fastslide_image_clone(image);
  }

  printf("    Converting to separate planar configuration...");

  // Get image info
  FastSlideImageInfo info;
  if (!fastslide_image_get_info(image, &info)) {
    printf(" FAILED (get info)\n");
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
      printf(" FAILED (unsupported format)\n");
      return NULL;
  }

  if (!separate_image) {
    printf(" FAILED (create separate image)\n");
    return NULL;
  }

  // Copy data from original to separate image
  const uint8_t* src_data = fastslide_image_get_data(image);
  uint8_t* dst_data = fastslide_image_get_data_mutable(separate_image);

  if (!src_data || !dst_data) {
    printf(" FAILED (get data pointers)\n");
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

  printf(" ✓\n");
  return separate_image;
}

// Compute histograms and display ranges for spectral images
int compute_spectral_display_ranges(const FastSlideSlideReader* reader,
                                    FastSlideChannelMetadata** channel_metadata,
                                    int* num_channels, double** min_values,
                                    double** max_values) {
  printf("Computing spectral channel display ranges...\n");

  // Get channel metadata
  if (!fastslide_slide_reader_get_channel_metadata(reader, channel_metadata,
                                                   num_channels)) {
    printf("  No channel metadata found\n");
    return 0;
  }

  if (*num_channels == 0) {
    printf("  No channels found\n");
    return 0;
  }

  printf("  Found %d spectral channels\n", *num_channels);

  // Use highest level for histogram computation (like QuPath)
  int level_count = fastslide_slide_reader_get_level_count(reader);
  int selected_level = level_count - 1;

  FastSlideImageDimensions dims;
  if (!fastslide_slide_reader_get_level_dimensions(reader, selected_level,
                                                   &dims)) {
    printf("  Failed to get level dimensions\n");
    return 0;
  }

  printf("  Computing histograms from level %d (%ux%u)\n", selected_level,
         dims.width, dims.height);

  // Read the entire level
  FastSlideImage* image = fastslide_slide_reader_read_region_coords(
      reader, 0, 0, dims.width, dims.height, selected_level);

  if (!image) {
    printf("  Failed to read image region: %s\n", fastslide_get_last_error());
    return 0;
  }

  // Create histograms
  FastSlideHistogram** histograms;
  int num_histograms;
  if (!fastslide_histogram_create_from_image_channels(image, 1024, &histograms,
                                                      &num_histograms)) {
    printf("  Failed to create histograms: %s\n", fastslide_get_last_error());
    fastslide_image_free(image);
    return 0;
  }

  // Allocate display range arrays
  *min_values = malloc(*num_channels * sizeof(double));
  *max_values = malloc(*num_channels * sizeof(double));

  if (!*min_values || !*max_values) {
    printf("  Failed to allocate display range arrays\n");
    fastslide_histogram_free_array(histograms, num_histograms);
    fastslide_image_free(image);
    return 0;
  }

  // Compute display ranges with 0.1% saturation
  double saturation = 0.001;
  for (int i = 0; i < *num_channels && i < num_histograms; ++i) {
    FastSlideDisplayRange range;
    if (fastslide_histogram_compute_display_range(histograms[i], saturation,
                                                  &range)) {
      (*min_values)[i] = range.min_value;
      (*max_values)[i] = range.max_value;
    } else {
      // Fallback to edge values
      (*min_values)[i] = fastslide_histogram_get_edge_min(histograms[i]);
      (*max_values)[i] = fastslide_histogram_get_edge_max(histograms[i]);
    }
  }

  printf("  Computed display ranges for %d channels\n", *num_channels);

  // Cleanup
  fastslide_histogram_free_array(histograms, num_histograms);
  fastslide_image_free(image);

  return 1;
}

// Save original and resampled versions of a spectral image with timing
int process_and_resample_spectral_image(
    const FastSlideSlideReader* reader, const char* output_dir,
    const Config* config, const FastSlideChannelMetadata* channel_metadata,
    int num_channels, const double* min_values, const double* max_values) {
  printf("\nProcessing spectral image for resampling demonstration...\n");

  // Find suitable level (under 3000 pixels for efficiency)
  int level_count = fastslide_slide_reader_get_level_count(reader);
  int target_level = -1;
  uint32_t target_width = 0, target_height = 0;

  for (int level = 0; level < level_count; ++level) {
    FastSlideImageDimensions dims;
    if (fastslide_slide_reader_get_level_dimensions(reader, level, &dims)) {
      if (dims.width < 3000 && dims.height < 3000) {
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
    }
  }

  printf("  Using level %d (%ux%u)\n", target_level, target_width,
         target_height);

  // Read the full region
  FastSlideImage* original_image = fastslide_slide_reader_read_region_coords(
      reader, 0, 0, target_width, target_height, target_level);

  if (!original_image) {
    printf("  Failed to read region: %s\n", fastslide_get_last_error());
    return 0;
  }

  // Check if image is spectral
  FastSlideImageFormat format = fastslide_image_get_format(original_image);
  if (format != FASTSLIDE_IMAGE_FORMAT_SPECTRAL) {
    printf("  Image is not spectral format, converting to RGB directly\n");
    FastSlideImage* rgb_image = fastslide_image_to_rgb(original_image);
    if (rgb_image) {
      // Save RGB image
      const uint8_t* rgb_data = fastslide_image_get_data(rgb_image);
      char filename[512];
      snprintf(filename, sizeof(filename), "%s/original_rgb.png", output_dir);
      save_rgb_as_png(rgb_data, target_width, target_height, filename);
      printf("  Saved RGB version: original_rgb.png\n");
      fastslide_image_free(rgb_image);
    }
    fastslide_image_free(original_image);
    return 1;
  }

  printf("  Processing spectral image with %u channels\n",
         fastslide_image_get_channels(original_image));

  // Convert original spectral to RGB for saving
  FastSlideImage* original_rgb = NULL;
  if (channel_metadata && num_channels > 0 && min_values && max_values) {
    printf("  Converting original using spectral blending...");

    // Create display ranges array
    FastSlideDisplayRange* display_ranges =
        malloc(num_channels * sizeof(FastSlideDisplayRange));
    if (display_ranges) {
      for (int i = 0; i < num_channels; ++i) {
        display_ranges[i].min_value = min_values[i];
        display_ranges[i].max_value = max_values[i];
      }

      original_rgb = fastslide_combine_spectral_channels_with_display_ranges(
          original_image, channel_metadata, num_channels, display_ranges,
          num_channels);

      free(display_ranges);
    }

    if (!original_rgb) {
      printf(" (fallback)");
      original_rgb = fastslide_image_to_rgb(original_image);
    }
    printf(" ✓\n");
  } else {
    printf("  Converting original using basic RGB conversion...");
    original_rgb = fastslide_image_to_rgb(original_image);
    printf(" ✓\n");
  }

  if (!original_rgb) {
    printf("  Failed to convert original to RGB\n");
    fastslide_image_free(original_image);
    return 0;
  }

  // Save original RGB
  const uint8_t* original_rgb_data = fastslide_image_get_data(original_rgb);
  char filename[512];
  snprintf(filename, sizeof(filename), "%s/original_spectral.png", output_dir);
  if (save_rgb_as_png(original_rgb_data, target_width, target_height,
                      filename)) {
    printf("  Saved original: original_spectral.png\n");
  }

  // Now perform resampling with timing
  uint32_t new_width = (uint32_t)(target_width / config->downsample_factor);
  uint32_t new_height = (uint32_t)(target_height / config->downsample_factor);

  printf("  Resampling spectral data (%ux%u → %ux%u, downsample %.1fx)\n",
         target_width, target_height, new_width, new_height,
         config->downsample_factor);

  // Ensure image has separate planar configuration for resampling
  FastSlideImage* separate_image = ensure_separate_planar(original_image);
  if (!separate_image) {
    printf("  Failed to convert to separate planar configuration\n");
    fastslide_image_free(original_rgb);
    fastslide_image_free(original_image);
    return 0;
  }

  // Time the resampling operation
  printf("  Running %d timing iterations", config->num_runs);
  fflush(stdout);

  long long* durations = malloc(config->num_runs * sizeof(long long));
  FastSlideImage* resampled_spectral = NULL;

  for (int run = 0; run < config->num_runs; ++run) {
    if (run % 5 == 0) {
      printf(".");
      fflush(stdout);
    }

    long long start_time = get_time_ms();

    // Perform Lanczos resampling
    FastSlideImage* temp_resampled =
        fastslide_lanczos_resample(separate_image, new_width, new_height);

    long long end_time = get_time_ms();
    durations[run] = end_time - start_time;

    if (!temp_resampled) {
      printf("\n  Resampling failed on run %d: %s\n", run + 1,
             fastslide_get_last_error());
      free(durations);
      fastslide_image_free(separate_image);
      fastslide_image_free(original_rgb);
      fastslide_image_free(original_image);
      return 0;
    }

    // Keep the last result for saving
    if (run == config->num_runs - 1) {
      resampled_spectral = temp_resampled;
    } else {
      fastslide_image_free(temp_resampled);
    }
  }

  // Compute average timing
  long long total_duration = 0;
  for (int i = 0; i < config->num_runs; ++i) {
    total_duration += durations[i];
  }
  long long average_duration = total_duration / config->num_runs;

  printf(" ✓ (avg: %lld ms over %d runs)\n", average_duration,
         config->num_runs);
  free(durations);

  if (!resampled_spectral) {
    printf("  Failed to get resampled result\n");
    fastslide_image_free(separate_image);
    fastslide_image_free(original_rgb);
    fastslide_image_free(original_image);
    return 0;
  }

  printf("  Converting resampled spectral to RGB...");

  // Convert resampled spectral to RGB
  FastSlideImage* resampled_rgb = NULL;
  if (channel_metadata && num_channels > 0 && min_values && max_values) {
    // Use spectral blending
    FastSlideDisplayRange* display_ranges =
        malloc(num_channels * sizeof(FastSlideDisplayRange));
    if (display_ranges) {
      for (int i = 0; i < num_channels; ++i) {
        display_ranges[i].min_value = min_values[i];
        display_ranges[i].max_value = max_values[i];
      }

      resampled_rgb = fastslide_combine_spectral_channels_with_display_ranges(
          resampled_spectral, channel_metadata, num_channels, display_ranges,
          num_channels);

      free(display_ranges);
    }

    if (!resampled_rgb) {
      resampled_rgb = fastslide_image_to_rgb(resampled_spectral);
    }
  } else {
    resampled_rgb = fastslide_image_to_rgb(resampled_spectral);
  }

  if (!resampled_rgb) {
    printf(" FAILED\n");
    fastslide_image_free(resampled_spectral);
    fastslide_image_free(separate_image);
    fastslide_image_free(original_rgb);
    fastslide_image_free(original_image);
    return 0;
  }
  printf(" ✓\n");

  // Save resampled RGB
  const uint8_t* resampled_rgb_data = fastslide_image_get_data(resampled_rgb);
  snprintf(filename, sizeof(filename), "%s/resampled_spectral_%dx%d.png",
           output_dir, new_width, new_height);
  if (save_rgb_as_png(resampled_rgb_data, new_width, new_height, filename)) {
    printf("  Saved resampled: resampled_spectral_%dx%d.png\n", new_width,
           new_height);
  }

  // Cleanup
  fastslide_image_free(resampled_rgb);
  fastslide_image_free(resampled_spectral);
  fastslide_image_free(separate_image);
  fastslide_image_free(original_rgb);
  fastslide_image_free(original_image);

  return 1;
}

// Test different resampling algorithms
void test_resampling_algorithms(const FastSlideSlideReader* reader,
                                const char* output_dir) {
  printf("\nTesting different resampling algorithms...\n");

  // Find a suitable level for testing
  int level_count = fastslide_slide_reader_get_level_count(reader);
  int test_level = level_count - 1;  // Use lowest resolution for speed

  FastSlideImageDimensions dims;
  if (!fastslide_slide_reader_get_level_dimensions(reader, test_level, &dims)) {
    printf("  Failed to get test level dimensions\n");
    return;
  }

  // Use a smaller region for algorithm comparison
  uint32_t test_width = dims.width > 1000 ? 1000 : dims.width;
  uint32_t test_height = dims.height > 1000 ? 1000 : dims.height;

  printf("  Using test region: %ux%u from level %d\n", test_width, test_height,
         test_level);

  FastSlideImage* test_image = fastslide_slide_reader_read_region_coords(
      reader, 0, 0, test_width, test_height, test_level);

  if (!test_image) {
    printf("  Failed to read test region\n");
    return;
  }

  // Ensure separate planar configuration
  FastSlideImage* separate_test = ensure_separate_planar(test_image);
  if (!separate_test) {
    printf("  Failed to convert test image to separate planar\n");
    fastslide_image_free(test_image);
    return;
  }

  // Test dimensions (half size)
  uint32_t target_width = test_width / 2;
  uint32_t target_height = test_height / 2;

  // Test Lanczos3
  printf("  Testing Lanczos3 resampling...");
  long long start_time = get_time_ms();
  FastSlideImage* lanczos3_result =
      fastslide_lanczos_resample(separate_test, target_width, target_height);
  long long lanczos3_time = get_time_ms() - start_time;

  if (lanczos3_result) {
    printf(" ✓ (%lld ms)\n", lanczos3_time);

    // Convert to RGB and save
    FastSlideImage* lanczos3_rgb = fastslide_image_to_rgb(lanczos3_result);
    if (lanczos3_rgb) {
      const uint8_t* rgb_data = fastslide_image_get_data(lanczos3_rgb);
      char filename[512];
      snprintf(filename, sizeof(filename), "%s/test_lanczos3_%dx%d.png",
               output_dir, target_width, target_height);
      save_rgb_as_png(rgb_data, target_width, target_height, filename);
      fastslide_image_free(lanczos3_rgb);
    }
    fastslide_image_free(lanczos3_result);
  } else {
    printf(" FAILED (%s)\n", fastslide_get_last_error());
  }

  // Test Lanczos2
  printf("  Testing Lanczos2 resampling...");
  start_time = get_time_ms();
  FastSlideImage* lanczos2_result =
      fastslide_lanczos2_resample(separate_test, target_width, target_height);
  long long lanczos2_time = get_time_ms() - start_time;

  if (lanczos2_result) {
    printf(" ✓ (%lld ms)\n", lanczos2_time);

    // Convert to RGB and save
    FastSlideImage* lanczos2_rgb = fastslide_image_to_rgb(lanczos2_result);
    if (lanczos2_rgb) {
      const uint8_t* rgb_data = fastslide_image_get_data(lanczos2_rgb);
      char filename[512];
      snprintf(filename, sizeof(filename), "%s/test_lanczos2_%dx%d.png",
               output_dir, target_width, target_height);
      save_rgb_as_png(rgb_data, target_width, target_height, filename);
      fastslide_image_free(lanczos2_rgb);
    }
    fastslide_image_free(lanczos2_result);
  } else {
    printf(" FAILED (%s)\n", fastslide_get_last_error());
  }

  // Test Cosine-windowed sinc
  printf("  Testing Cosine-windowed sinc resampling...");
  start_time = get_time_ms();
  FastSlideImage* cosine_result =
      fastslide_cosine_resample(separate_test, target_width, target_height);
  long long cosine_time = get_time_ms() - start_time;

  if (cosine_result) {
    printf(" ✓ (%lld ms)\n", cosine_time);

    // Convert to RGB and save
    FastSlideImage* cosine_rgb = fastslide_image_to_rgb(cosine_result);
    if (cosine_rgb) {
      const uint8_t* rgb_data = fastslide_image_get_data(cosine_rgb);
      char filename[512];
      snprintf(filename, sizeof(filename), "%s/test_cosine_%dx%d.png",
               output_dir, target_width, target_height);
      save_rgb_as_png(rgb_data, target_width, target_height, filename);
      fastslide_image_free(cosine_rgb);
    }
    fastslide_image_free(cosine_result);
  } else {
    printf(" FAILED (%s)\n", fastslide_get_last_error());
  }

  // Cleanup
  fastslide_image_free(separate_test);
  fastslide_image_free(test_image);

  printf("  Algorithm comparison complete\n");
}

int main(int argc, char** argv) {
  Config config = parse_command_line(argc, argv);

  printf("FastSlide Spectral Resampling Example - C API\n");
  printf("Slide file: %s\n", config.slide_file);
  printf("Output directory: %s\n", config.output_dir);
  printf("Timing runs: %d\n", config.num_runs);
  printf("Downsample factor: %.1fx\n\n", config.downsample_factor);

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

  // Compute display ranges for spectral channels
  FastSlideChannelMetadata* channel_metadata = NULL;
  int num_channels = 0;
  double* min_values = NULL;
  double* max_values = NULL;

  // Process and resample the main spectral image
  if (!process_and_resample_spectral_image(reader, config.output_dir, &config,
                                           channel_metadata, num_channels,
                                           min_values, max_values)) {
    printf("Failed to process and resample spectral image\n");
  }

  // Test different resampling algorithms
  test_resampling_algorithms(reader, config.output_dir);

  // Cleanup spectral data
  if (channel_metadata) {
    fastslide_slide_reader_free_channel_metadata(channel_metadata,
                                                 num_channels);
  }
  free(min_values);
  free(max_values);

  // Cleanup
  fastslide_slide_reader_free(reader);
  fastslide_cleanup();
  vips_shutdown();

  printf("\n✓ All resampled images saved to %s\n", config.output_dir);
  printf("✓ Spectral resampling demonstration complete\n");

  return 0;
}