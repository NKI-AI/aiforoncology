// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
#ifndef AIFO_SLIDESCOPE_TIFF_LIBTIFF_WRAPPER_H
#define AIFO_SLIDESCOPE_TIFF_LIBTIFF_WRAPPER_H

#include <stdint.h>
#include <tiffio.h>

// Error handling functions
void InstallTiffErrorHandler(void);
const char* GetLastTiffError(void);
void ClearLastTiffError(void);

uint32_t GetTileWidth(TIFF* tif);
uint32_t GetTileLength(TIFF* tif);

// TIFF reading functions
int ReadRegion(TIFF* tif, uint32_t x, uint32_t y, uint32_t w, uint32_t h,
               void* out_buf);
int GetImageSize(TIFF* tif, uint32_t* width, uint32_t* height);
int GetMpp(TIFF* tif, double* mpp_x, double* mpp_y);

// Helper functions to get TIFF metadata
uint32_t GetWidth(TIFF* tif);
uint32_t GetHeight(TIFF* tif);
uint16_t GetBitsPerSample(TIFF* tif);
uint16_t GetSamplesPerPixel(TIFF* tif);

#endif  // AIFO_SLIDESCOPE_TIFF_LIBTIFF_WRAPPER_H
