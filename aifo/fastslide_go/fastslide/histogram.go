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

package fastslide

/*
#include "fastslide/c/fastslide.h"
#include <stdlib.h>
#include <string.h>
*/
import "C"

import (
	"runtime"
	"unsafe"
)

// Histogram wraps a FastSlideHistogram handle.
type Histogram struct {
	handle *C.FastSlideHistogram
}

// DisplayRange represents a min/max display range pair.
type DisplayRange struct {
	Min float64
	Max float64
}

// Close closes the histogram and releases resources.
func (h *Histogram) Close() {
	if h.handle != nil {
		C.fastslide_histogram_free(h.handle)
		h.handle = nil
		runtime.SetFinalizer(h, nil)
	}
}

// finalize is called by the garbage collector if Close() wasn't called.
func (h *Histogram) finalize() {
	h.Close()
}

// ComputeDisplayRange computes the display range for the histogram.
func (h *Histogram) ComputeDisplayRange(saturation float64) (DisplayRange, error) {
	if h.handle == nil {
		return DisplayRange{}, &FastSlideError{
			Op:      "ComputeDisplayRange",
			Code:    CodeInvalidState,
			Message: "histogram is closed",
		}
	}

	var cRange C.FastSlideDisplayRange
	C.fastslide_clear_last_error()
	if C.fastslide_histogram_compute_display_range(h.handle, C.double(saturation), &cRange) == 0 {
		errorMsg := C.GoString(C.fastslide_get_last_error())
		return DisplayRange{}, &FastSlideError{
			Op:      "ComputeDisplayRange",
			Code:    CodeInternal,
			Message: errorMsg,
		}
	}

	return DisplayRange{
		Min: float64(cRange.min_value),
		Max: float64(cRange.max_value),
	}, nil
}

// GetEdgeMin returns the minimum edge value of the histogram.
func (h *Histogram) GetEdgeMin() float64 {
	if h.handle == nil {
		return 0
	}
	return float64(C.fastslide_histogram_get_edge_min(h.handle))
}

// GetEdgeMax returns the maximum edge value of the histogram.
func (h *Histogram) GetEdgeMax() float64 {
	if h.handle == nil {
		return 0
	}
	return float64(C.fastslide_histogram_get_edge_max(h.handle))
}

// GetCountSum returns the total count sum of the histogram.
func (h *Histogram) GetCountSum() int64 {
	if h.handle == nil {
		return 0
	}
	return int64(C.fastslide_histogram_get_count_sum(h.handle))
}

// CreateHistogramsFromImageChannels creates histograms for all channels of an image.
func CreateHistogramsFromImageChannels(image *Image, nBins int) ([]*Histogram, error) {
	if image.handle == nil {
		return nil, &FastSlideError{
			Op:      "CreateHistogramsFromImageChannels",
			Code:    CodeInvalidState,
			Message: "image is closed",
		}
	}

	var cHistograms **C.FastSlideHistogram
	var count C.int

	C.fastslide_clear_last_error()
	if C.fastslide_histogram_create_from_image_channels(
		image.handle,
		C.int(nBins),
		&cHistograms,
		&count,
	) == 0 {
		errorMsg := C.GoString(C.fastslide_get_last_error())
		return nil, &FastSlideError{
			Op:      "CreateHistogramsFromImageChannels",
			Code:    CodeInternal,
			Message: errorMsg,
		}
	}

	result := make([]*Histogram, int(count))
	if count > 0 {
		slice := unsafe.Slice(cHistograms, int(count))
		for i, cHist := range slice {
			hist := &Histogram{handle: cHist}
			runtime.SetFinalizer(hist, (*Histogram).finalize)
			result[i] = hist
		}
		// Free the array of histogram pointers (but not the histograms themselves)
		C.free(unsafe.Pointer(cHistograms))
	}

	return result, nil
}

// CreateHistogramFromImageChannel creates a histogram for a specific channel.
func CreateHistogramFromImageChannel(image *Image, channel uint32, nBins int) (*Histogram, error) {
	if image.handle == nil {
		return nil, &FastSlideError{
			Op:      "CreateHistogramFromImageChannel",
			Code:    CodeInvalidState,
			Message: "image is closed",
		}
	}

	C.fastslide_clear_last_error()
	handle := C.fastslide_histogram_create_from_image_channel(
		image.handle,
		C.uint32_t(channel),
		C.int(nBins),
	)

	if handle == nil {
		errorMsg := C.GoString(C.fastslide_get_last_error())
		return nil, &FastSlideError{
			Op:      "CreateHistogramFromImageChannel",
			Code:    CodeInternal,
			Message: errorMsg,
		}
	}

	hist := &Histogram{handle: handle}
	runtime.SetFinalizer(hist, (*Histogram).finalize)
	return hist, nil
}
