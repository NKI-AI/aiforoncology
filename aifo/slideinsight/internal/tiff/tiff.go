// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package tiff

/*
#include <tiffio.h>
#include <stdlib.h>
#include "libtiff_wrapper.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"sync"
	"unsafe"
)

// Tiff wraps a collection of TIFF* handles (one per pyramid level) and stores metadata.
type Tiff struct {
	tifs            []*C.TIFF // One TIFF handle per pyramid level
	levels          int       // Number of pyramid levels
	baseWidth       uint32
	baseHeight      uint32
	bitsPerSample   uint16  // Bits per sample (8 or 16)
	samplesPerPixel uint16  // Must be 1 (single channel)
	baseMppX        float64 // µm/pixel at level 0
	baseMppY        float64
	downsamples     []float64    // Downsampling factor for each level (1.0, 2.0, 4.0, etc.)
	locks           []sync.Mutex // one mutex per level to guard concurrent C calls
}

func init() {
	C.InstallTiffErrorHandler()
}

// Open a pyramidal, tiled, single-channel TIFF. Errors if wrong samples-per-pixel,
// unsupported bitdepth, or TIFFOpen fails.
func Open(path string) (*Tiff, error) {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	C.ClearLastTiffError()

	// 1) Open initial handle to count directories
	mainTiff := C.TIFFOpen(cpath, C.CString("r"))
	if mainTiff == nil {
		return nil, fmt.Errorf("TIFFOpen failed: %s", C.GoString(C.GetLastTiffError()))
	}
	defer C.TIFFClose(mainTiff)

	levelCount := 1
	for C.TIFFReadDirectory(mainTiff) != 0 {
		levelCount++
	}

	// 2) Re-open one handle per directory
	tifs := make([]*C.TIFF, levelCount)
	for level := 0; level < levelCount; level++ {
		tifs[level] = C.TIFFOpen(cpath, C.CString("r"))
		if tifs[level] == nil {
			// cleanup on error
			for i := 0; i < level; i++ {
				C.TIFFClose(tifs[i])
			}
			return nil, fmt.Errorf("TIFFOpen failed for level %d: %s", level, C.GoString(C.GetLastTiffError()))
		}
		if level > 0 {
			// advance to the correct directory
			for i := 0; i < level; i++ {
				if C.TIFFReadDirectory(tifs[level]) == 0 {
					for j := 0; j <= level; j++ {
						C.TIFFClose(tifs[j])
					}
					return nil, fmt.Errorf("failed to read directory %d: %s", i, C.GoString(C.GetLastTiffError()))
				}
			}
		}
	}

	// 3) Read core tags from level 0
	tif0 := tifs[0]
	width := C.GetWidth(tif0)
	height := C.GetHeight(tif0)
	bps := C.GetBitsPerSample(tif0)
	spp := C.GetSamplesPerPixel(tif0)

	if spp != 1 {
		for _, tif := range tifs {
			C.TIFFClose(tif)
		}
		return nil, fmt.Errorf("unsupported SamplesPerPixel=%d", spp)
	}
	if bps != 8 && bps != 16 {
		for _, tif := range tifs {
			C.TIFFClose(tif)
		}
		return nil, fmt.Errorf("unsupported BitsPerSample=%d", bps)
	}

	var mppX, mppY C.double
	hasMPP := C.GetMpp(tif0, &mppX, &mppY) != 0
	if !hasMPP {
		return nil, errors.New("no resolution metadata in TIFF")
	}

	baseW := uint32(width)
	downsamples := make([]float64, levelCount)
	downsamples[0] = 1.0
	for lvl := 1; lvl < levelCount; lvl++ {
		lw := uint32(C.GetWidth(tifs[lvl]))
		if lw > 0 {
			downsamples[lvl] = float64(baseW) / float64(lw)
		} else {
			return nil, fmt.Errorf("width is 0 at level %d", lvl)
		}
	}

	// 4) Prepare per-level locks
	locks := make([]sync.Mutex, levelCount)

	t := &Tiff{
		tifs:            tifs,
		levels:          levelCount,
		baseWidth:       baseW,
		baseHeight:      uint32(height),
		bitsPerSample:   uint16(bps),
		samplesPerPixel: uint16(spp),
		baseMppX:        float64(mppX),
		baseMppY:        float64(mppY),
		downsamples:     downsamples,
		locks:           locks,
	}

	return t, nil
}

// Close all underlying TIFF handles.
func (t *Tiff) Close() {
	for i, tif := range t.tifs {
		if tif != nil {
			C.TIFFClose(tif)
			t.tifs[i] = nil
		}
	}
}

// LevelCount returns the number of pyramid levels.
func (t *Tiff) LevelCount() int { return t.levels }

// LevelSize returns width×height of a given level.
func (t *Tiff) LevelSize(level int) (uint32, uint32, error) {
	if level < 0 || level >= t.levels {
		return 0, 0, fmt.Errorf("invalid level %d", level)
	}
	var w, h C.uint32_t
	if C.GetImageSize(t.tifs[level], &w, &h) == 0 {
		return 0, 0, fmt.Errorf("tiff_get_image_size failed: %s", C.GoString(C.GetLastTiffError()))
	}
	return uint32(w), uint32(h), nil
}

// BaseResolution returns µm/pixel at level 0 (error if missing).
func (t *Tiff) BaseResolution() (xµm, yµm float64, err error) {
	if t.baseMppX <= 0 || t.baseMppY <= 0 {
		return 0, 0, errors.New("no resolution metadata in TIFF")
	}
	return t.baseMppX, t.baseMppY, nil
}

// LevelInfo returns a summary string for a level.
func (t *Tiff) LevelInfo(level int) (string, error) {
	w, h, err := t.LevelSize(level)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Level %d: %d×%d (downsample: %.2fx)",
		level, w, h, t.downsamples[level]), nil
}

// tileSize returns the native tile dimensions for a level.
func (t *Tiff) tileSize(level int) (uint32, uint32) {
	tif := t.tifs[level]
	return uint32(C.GetTileWidth(tif)), uint32(C.GetTileLength(tif))
}

// ReadRegion reads a w×h tile at (x,y) in level-0 coords, scaled to `level`.
// w,h = 0 means “full native tile”. Returns []byte for 8-bit or []uint16 for 16-bit.
func (t *Tiff) ReadRegion(
	x, y uint32, // origin in level-0 coords
	level int, // pyramid level
	w, h uint32, // tile size in level coord (0→auto)
) (interface{}, error) {
	if level < 0 || level >= t.levels {
		return nil, fmt.Errorf("invalid level %d", level)
	}

	size := w * h * uint32(t.bitsPerSample/8)
	buf := C.malloc(C.size_t(size))
	if buf == nil {
		return nil, errors.New("malloc failed")
	}
	defer C.free(buf)

	// ← serialize C call per‐level
	t.locks[level].Lock()
	defer t.locks[level].Unlock()

	C.ClearLastTiffError()
	res := C.ReadRegion(
		t.tifs[level],
		C.uint32_t(x), C.uint32_t(y),
		C.uint32_t(w), C.uint32_t(h),
		buf,
	)
	if res == 0 {
		errmsg := C.GoString(C.GetLastTiffError())
		if errmsg == "" {
			errmsg = "unknown error"
		}
		return nil, fmt.Errorf(
			"ReadRegion failed at (%d,%d) lvl=%d: %s",
			x, y, level, errmsg,
		)
	}

	data := C.GoBytes(buf, C.int(size))
	if t.bitsPerSample == 8 {
		return data, nil
	}
	return unsafe.Slice((*uint16)(unsafe.Pointer(&data[0])), int(size)/2), nil
}
