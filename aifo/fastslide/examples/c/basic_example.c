// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of FastSlide.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the FastSlide project root.

#include "fastslide/c/fastslide_wrapper.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

void print_usage(const char* program_name) {
  printf("Usage: %s <slide_file>\n", program_name);
  printf("Example: %s slide.svs\n", program_name);
}

void print_supported_formats() {
  int count;
  const char** extensions = fastslide_get_supported_extensions(&count);

  if (extensions && count > 0) {
    printf("Supported formats: ");
    for (int i = 0; i < count; i++) {
      printf("%s", extensions[i]);
      if (i < count - 1)
        printf(", ");
    }
    printf("\n");
    fastslide_free_extensions(extensions);
  } else {
    printf("No supported formats found.\n");
  }
}

int main(int argc, char* argv[]) {
  if (argc != 2) {
    print_usage(argv[0]);
    return 1;
  }

  const char* slide_file = argv[1];

  printf("=== FastSlide C API Example ===\n");
  printf("Version: %s\n", fastslide_get_version());

  // Install error handler
  fastslide_install_error_handler();

  // Print supported formats
  print_supported_formats();

  // Check if file is supported
  if (!fastslide_is_supported(slide_file)) {
    printf("Error: File format not supported for '%s'\n", slide_file);
    const char* error = fastslide_get_last_error();
    if (error && strlen(error) > 0) {
      printf("Details: %s\n", error);
    }
    return 1;
  }

  printf("Opening slide: %s\n", slide_file);

  // Open the slide
  FastSlideC* slide = fastslide_open_from_file(slide_file);
  if (!slide) {
    printf("Error: Failed to open slide '%s'\n", slide_file);
    const char* error = fastslide_get_last_error();
    if (error && strlen(error) > 0) {
      printf("Details: %s\n", error);
    }
    return 1;
  }

  printf("Successfully opened slide!\n\n");

  // Get basic slide properties
  const char* format = fastslide_get_format(slide);
  printf("Format: %s\n", format ? format : "Unknown");

  const char* scanner = FastSlideGetScannerModel(slide);
  printf("Scanner: %s\n", scanner ? scanner : "Unknown");

  double mpp_x = fastslide_get_mpp_x(slide);
  double mpp_y = fastslide_get_mpp_y(slide);
  if (mpp_x > 0 && mpp_y > 0) {
    printf("Resolution: %.3f x %.3f µm/pixel\n", mpp_x, mpp_y);
  }

  double magnification = fastslide_get_objective_magnification(slide);
  if (magnification > 0) {
    printf("Magnification: %.1fx\n", magnification);
  }

  // Get level information
  int level_count = fastslide_get_level_count(slide);
  if (level_count < 0) {
    printf("Error: Failed to get level count\n");
    const char* error = fastslide_get_last_error();
    if (error && strlen(error) > 0) {
      printf("Details: %s\n", error);
    }
    fastslide_close(slide);
    return 1;
  }

  printf("\nPyramid levels: %d\n", level_count);

  for (int level = 0; level < level_count; level++) {
    uint32_t width, height;
    if (fastslide_get_level_dimensions(slide, level, &width, &height)) {
      double downsample = fastslide_get_level_downsample(slide, level);
      printf("  Level %d: %u x %u (downsample: %.2fx)\n", level, width, height,
             downsample);
    } else {
      printf("  Level %d: Failed to get dimensions\n", level);
    }
  }

  // Get base dimensions
  uint32_t base_width, base_height;
  if (fastslide_get_base_dimensions(slide, &base_width, &base_height)) {
    printf("\nBase dimensions (level 0): %u x %u\n", base_width, base_height);
  }

  // Test best level for downsample
  double test_downsample = 4.0;
  int best_level =
      fastslide_get_best_level_for_downsample(slide, test_downsample);
  if (best_level >= 0) {
    printf("Best level for %.1fx downsample: %d\n", test_downsample,
           best_level);
  }

  // Get associated image names
  char** assoc_names;
  int assoc_count;
  if (fastslide_get_associated_image_names(slide, &assoc_names, &assoc_count)) {
    if (assoc_count > 0) {
      printf("\nAssociated images (%d):\n", assoc_count);
      for (int i = 0; i < assoc_count; i++) {
        uint32_t assoc_width, assoc_height;
        if (fastslide_get_associated_image_dimensions(
                slide, assoc_names[i], &assoc_width, &assoc_height)) {
          printf("  %s: %u x %u\n", assoc_names[i], assoc_width, assoc_height);
        } else {
          printf("  %s: Failed to get dimensions\n", assoc_names[i]);
        }
      }
      fastslide_free_associated_image_names(assoc_names, assoc_count);
    } else {
      printf("\nNo associated images found.\n");
    }
  } else {
    printf("\nFailed to get associated image names\n");
  }

  // Test reading a small region (level-native coordinates)
  printf("\nTesting region reading...\n");
  uint32_t region_size = 256;
  int test_level =
      (level_count > 1) ? 1 : 0;  // Use level 1 if available, otherwise 0

  uint32_t level_width, level_height;
  if (fastslide_get_level_dimensions(slide, test_level, &level_width,
                                     &level_height)) {
    // Read from center of the level
    uint32_t x =
        (level_width > region_size) ? (level_width - region_size) / 2 : 0;
    uint32_t y =
        (level_height > region_size) ? (level_height - region_size) / 2 : 0;
    uint32_t width = (level_width > region_size) ? region_size : level_width;
    uint32_t height = (level_height > region_size) ? region_size : level_height;

    printf(
        "Reading %ux%u region at (%u, %u) from level %d (level-native "
        "coordinates)...\n",
        width, height, x, y, test_level);

    // Allocate buffer for RGB data (3 bytes per pixel)
    size_t buffer_size = width * height * 3;
    uint8_t* buffer = (uint8_t*)malloc(buffer_size);
    if (!buffer) {
      printf("Error: Failed to allocate memory for region buffer\n");
      fastslide_close(slide);
      return 1;
    }

    if (fastslide_read_region(slide, x, y, width, height, test_level, buffer)) {
      printf("Successfully read region! Buffer size: %zu bytes\n", buffer_size);

      // Print some sample pixel values
      printf("Sample pixel values (RGB):\n");
      for (int i = 0; i < 3 && i < (int)height; i++) {
        for (int j = 0; j < 3 && j < (int)width; j++) {
          size_t pixel_idx = (i * width + j) * 3;
          printf("  Pixel (%d,%d): R=%u G=%u B=%u\n", j, i, buffer[pixel_idx],
                 buffer[pixel_idx + 1], buffer[pixel_idx + 2]);
        }
      }
    } else {
      printf("Error: Failed to read region\n");
      const char* error = fastslide_get_last_error();
      if (error && strlen(error) > 0) {
        printf("Details: %s\n", error);
      }
    }

    free(buffer);
  } else {
    printf("Error: Failed to get level dimensions for region reading test\n");
  }

  // Clean up
  fastslide_close(slide);

  printf("\nExample completed successfully!\n");
  return 0;
}