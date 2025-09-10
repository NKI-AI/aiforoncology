// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
#include "libtiff_wrapper.h"
#include <stdarg.h>
#include <stdlib.h>
#include <string.h>

// Buffer for last error message.
static char tiff_last_error[1024];

// Custom libtiff error handler that captures error messages
// for later retrieval by Go code.

void goTIFFErrorHandler(const char* module, const char* fmt, va_list ap) {
  vsnprintf(tiff_last_error, sizeof(tiff_last_error), fmt, ap);
  fprintf(stderr, "[libtiff] ERROR in %s: %s", module ? module : "<null>",
          tiff_last_error);
}

const char* GetLastTiffError(void) {
  return tiff_last_error;
}

void ClearLastTiffError(void) {
  tiff_last_error[0] = '\0';
}

void InstallTiffErrorHandler(void) {
  TIFFSetErrorHandler(goTIFFErrorHandler);
}

uint32_t GetTileWidth(TIFF* tif) {
  uint32_t w = 0;
  TIFFGetField(tif, TIFFTAG_TILEWIDTH, &w);
  return w;
}

uint32_t GetTileLength(TIFF* tif) {
  uint32_t h = 0;
  TIFFGetField(tif, TIFFTAG_TILELENGTH, &h);
  return h;
}

int ReadRegion(TIFF* tif, uint32_t x, uint32_t y, uint32_t w, uint32_t h,
               void* out_buf) {

  uint16_t bps = 0;
  TIFFGetField(tif, TIFFTAG_BITSPERSAMPLE, &bps);
  int bytespp = bps / 8;

  // Get tile size in pixels
  uint32_t tile_w = 0, tile_h = 0;
  TIFFGetField(tif, TIFFTAG_TILEWIDTH, &tile_w);
  TIFFGetField(tif, TIFFTAG_TILELENGTH, &tile_h);

  // Allocate a buffer for one tile
  tsize_t tile_size = TIFFTileSize(tif);
  void* tile_buf = _TIFFmalloc(tile_size);

  // Zero-init the output
  memset(out_buf, 0, (size_t)w * h * bytespp);

  // Which tile rows/cols cover [x..x+w) × [y..y+h)
  uint32_t tx0 = x / tile_w;
  uint32_t ty0 = y / tile_h;
  uint32_t tx1 = (x + w - 1) / tile_w;
  uint32_t ty1 = (y + h - 1) / tile_h;

  for (uint32_t ty = ty0; ty <= ty1; ++ty) {
    for (uint32_t tx = tx0; tx <= tx1; ++tx) {
      uint32_t offx = tx * tile_w;
      uint32_t offy = ty * tile_h;

      // Read the tile at its natural position
      TIFFReadTile(tif, tile_buf, offx, offy, 0, 0);

      // Compute intersection of this tile with the requested region
      uint32_t ix0 = (x > offx ? x - offx : 0);
      uint32_t iy0 = (y > offy ? y - offy : 0);
      uint32_t ix1 = ((offx + tile_w) > (x + w) ? (x + w - offx) : tile_w);
      uint32_t iy1 = ((offy + tile_h) > (y + h) ? (y + h - offy) : tile_h);

      // Copy each scanline of the overlap
      for (uint32_t row = iy0; row < iy1; ++row) {
        void* src = (char*)tile_buf + ((row * tile_w + ix0) * bytespp);
        uint32_t dst_row = offy + row - y;
        void* dst =
            (char*)out_buf + ((dst_row * w + (offx + ix0 - x)) * bytespp);
        memcpy(dst, src, (ix1 - ix0) * bytespp);
      }
    }
  }

  _TIFFfree(tile_buf);

  return 1;
}

uint32_t GetWidth(TIFF* tif) {
  uint32_t width = 0;
  TIFFGetField(tif, TIFFTAG_IMAGEWIDTH, &width);
  return width;
}

uint32_t GetHeight(TIFF* tif) {
  uint32_t height = 0;
  TIFFGetField(tif, TIFFTAG_IMAGELENGTH, &height);
  return height;
}

uint16_t GetBitsPerSample(TIFF* tif) {
  uint16_t bits_per_sample = 0;
  TIFFGetField(tif, TIFFTAG_BITSPERSAMPLE, &bits_per_sample);
  return bits_per_sample;
}

uint16_t GetSamplesPerPixel(TIFF* tif) {
  uint16_t samples_per_pixel = 0;
  TIFFGetField(tif, TIFFTAG_SAMPLESPERPIXEL, &samples_per_pixel);
  return samples_per_pixel;
}

int GetMpp(TIFF* tif, double* mpp_x, double* mpp_y) {
  float x_res = 0, y_res = 0;
  uint16_t unit = 0;

  int has_x = TIFFGetField(tif, TIFFTAG_XRESOLUTION, &x_res);
  int has_y = TIFFGetField(tif, TIFFTAG_YRESOLUTION, &y_res);
  int has_unit = TIFFGetField(tif, TIFFTAG_RESOLUTIONUNIT, &unit);

  if (!has_x || !has_y || !has_unit) {
    return 0;
  }

  switch (unit) {
    case RESUNIT_INCH:
      *mpp_x = 25400.0 / x_res;
      *mpp_y = 25400.0 / y_res;
      return 1;
    case RESUNIT_CENTIMETER:
      *mpp_x = 10000.0 / x_res;
      *mpp_y = 10000.0 / y_res;
      return 1;
    default:
      return 0;  // Unsupported unit
  }
}

int GetImageSize(TIFF* tif, uint32_t* width, uint32_t* height) {
  if (!TIFFGetField(tif, TIFFTAG_IMAGEWIDTH, width)) {
    strncpy(tiff_last_error, "TIFFTAG_IMAGEWIDTH missing",
            sizeof(tiff_last_error));
    return 0;
  }
  if (!TIFFGetField(tif, TIFFTAG_IMAGELENGTH, height)) {
    strncpy(tiff_last_error, "TIFFTAG_IMAGELENGTH missing",
            sizeof(tiff_last_error));
    return 0;
  }
  return 1;
}