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

package openslide

/*
#cgo CFLAGS: -g -Wall
#cgo LDFLAGS: -lopenslide
#include <stdlib.h>
#include <openslide.h>
#include "argb_to_rgba.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"image"
	"log/slog"
	"strconv"
	"unsafe"

	"golang.org/x/image/draw"
)

// OpenSlideError represents an error that occurred in OpenSlide operations.
type OpenSlideError struct {
	Operation string // The operation that failed
	Message   string // Error message from the OpenSlide library
}

// Error implements the error interface.
func (e *OpenSlideError) Error() string {
	return fmt.Sprintf("OpenSlide %s error: %s", e.Operation, e.Message)
}

// Unwrap returns the underlying error.
func (e *OpenSlideError) Unwrap() error {
	return errors.New(e.Message)
}

// Slide wraps an OpenSlide pointer.
type Slide struct {
	ptr *C.openslide_t
}

// Open opens an OpenSlide image file.
func Open(filename string) (Slide, error) {
	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))

	ptr := C.openslide_open(cFilename)
	if ptr == nil {
		slog.Error("OpenSlide pointer is nil", "filename", filename)
		// Don't try to get an error from a NULL pointer - this causes segfault
		return Slide{}, fmt.Errorf("failed to open slide: %s", filename)
	}

	// Now check for errors with the valid pointer
	if errPtr := C.openslide_get_error(ptr); errPtr != nil {
		C.openslide_close(ptr)
		return Slide{}, fmt.Errorf("open failed: %s", C.GoString(errPtr))
	}

	return Slide{ptr: ptr}, nil
}

// Close releases the resources associated with the slide.
func (s Slide) Close() {
	C.openslide_close(s.ptr)
}

// LevelCount returns the number of levels in the image.
func (s Slide) LevelCount() (int, error) {
	count := int(C.openslide_get_level_count(s.ptr))
	if err := C.openslide_get_error(s.ptr); err != nil {
		C.openslide_close(s.ptr)
		return 0, &OpenSlideError{
			Operation: "LevelCount",
			Message:   C.GoString(err),
		}
	}
	return count, nil
}

// LargestLevelDimensions returns the dimensions of level 0.
func (s Slide) LargestLevelDimensions() ([2]int, error) {
	var w, h C.int64_t
	C.openslide_get_level0_dimensions(s.ptr, &w, &h)
	if err := C.openslide_get_error(s.ptr); err != nil {
		C.openslide_close(s.ptr)
		return [2]int{}, &OpenSlideError{
			Operation: "LargestLevelDimensions",
			Message:   C.GoString(err),
		}
	}
	return [2]int{int(w), int(h)}, nil
}

// LevelDimensions returns the dimensions for the specified level.
func (s Slide) LevelDimensions(level int) ([2]int, error) {
	var w, h C.int64_t
	C.openslide_get_level_dimensions(s.ptr, C.int32_t(level), &w, &h)
	if err := C.openslide_get_error(s.ptr); err != nil {
		C.openslide_close(s.ptr)
		return [2]int{}, &OpenSlideError{
			Operation: "LevelDimensions",
			Message:   C.GoString(err),
		}
	}
	return [2]int{int(w), int(h)}, nil
}

// LevelDownsample returns the downsample factor of a level.
func (s Slide) LevelDownsample(level int) (float64, error) {
	val := float64(C.openslide_get_level_downsample(s.ptr, C.int32_t(level)))
	if err := C.openslide_get_error(s.ptr); err != nil {
		C.openslide_close(s.ptr)
		return 0, &OpenSlideError{
			Operation: "LevelDownsample",
			Message:   C.GoString(err),
		}
	}
	return val, nil
}

// LevelDownsamples returns all downsample factors.
func (s Slide) LevelDownsamples() ([]float64, error) {
	count, err := s.LevelCount()
	if err != nil {
		return nil, err
	}

	factors := make([]float64, count)
	for i := 0; i < count; i++ {
		factor, err := s.LevelDownsample(i)
		if err != nil {
			return nil, err
		}
		factors[i] = factor
	}
	return factors, nil
}

// BestLevelForDownsample returns the closest level for the requested downsample.
func (s Slide) BestLevelForDownsample(downsample float64) (int, error) {
	val := int(C.openslide_get_best_level_for_downsample(s.ptr, C.double(downsample)))
	if err := C.openslide_get_error(s.ptr); err != nil {
		C.openslide_close(s.ptr)
		return 0, &OpenSlideError{
			Operation: "BestLevelForDownsample",
			Message:   C.GoString(err),
		}
	}
	return val, nil
}

// ReadRegion reads a region at the specified level and location, returning raw RGBA bytes.
// This is useful when you need direct access to pixel data without the overhead of creating an image.Image.
func (s Slide) ReadRegion(x, y, level, w, h int) ([]byte, error) {
	// Allocate buffer for RGBA data (4 bytes per pixel)
	pixelCount := w * h
	buffer := make([]byte, pixelCount*4)
	pixPtr := unsafe.Pointer(&buffer[0])

	C.openslide_read_region(
		s.ptr,
		(*C.uint32_t)(pixPtr),
		C.int64_t(x), C.int64_t(y), C.int32_t(level),
		C.int64_t(w), C.int64_t(h),
	)

	if errPtr := C.openslide_get_error(s.ptr); errPtr != nil {
		C.openslide_close(s.ptr)
		return nil, errors.New(C.GoString(errPtr))
	}

	// Convert from ARGB to RGBA format
	C.ArgbToRgba((*C.uint32_t)(pixPtr), C.size_t(pixelCount))
	return buffer, nil
}

// RegionToImage converts raw RGBA bytes to an image.Image.
// The bytes should be in RGBA format with 4 bytes per pixel.
func RegionToImage(data []byte, w, h int) (image.Image, error) {
	expectedSize := w * h * 4
	if len(data) != expectedSize {
		return nil, fmt.Errorf("invalid data size: expected %d bytes for %dx%d image, got %d",
			expectedSize, w, h, len(data))
	}

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	copy(img.Pix, data)
	return img, nil
}

// iccProfileSize returns the size of the ICC profile in bytes.
// Returns -1 if an error occurred, or 0 if no profile is available.
func (s Slide) iccProfileSize() (int64, error) {
	size := int64(C.openslide_get_icc_profile_size(s.ptr))
	if err := C.openslide_get_error(s.ptr); err != nil {
		C.openslide_close(s.ptr)
		return -1, &OpenSlideError{
			Operation: "ICCProfileSize",
			Message:   C.GoString(err),
		}
	}
	return size, nil
}

// ICCProfile returns the ICC profile as a byte slice.
// Returns an empty slice if no profile is available.
func (s Slide) ICCProfile() ([]byte, error) {
	size, err := s.iccProfileSize()
	if err != nil {
		return nil, err // Error already properly formatted
	}
	if size <= 0 {
		return nil, nil
	}

	buf := C.malloc(C.size_t(size))
	defer C.free(buf)

	C.openslide_read_icc_profile(s.ptr, buf)
	if err := C.openslide_get_error(s.ptr); err != nil {
		C.openslide_close(s.ptr)
		return nil, &OpenSlideError{
			Operation: "ReadICCProfile",
			Message:   C.GoString(err),
		}
	}

	goBytes := C.GoBytes(buf, C.int(size))
	return goBytes, nil
}

// DetectVendor returns the detected vendor string.
func DetectVendor(filename string) (string, error) {
	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))

	vendor := C.openslide_detect_vendor(cFilename)
	if vendor == nil {
		return "", errors.New("vendor not found: " + filename)
	}
	return C.GoString(vendor), nil
}

// cStringArrayToSlice converts a C string array to a Go slice of strings.
func cStringArrayToSlice(ptr **C.char) []string {
	var result []string
	for i := 0; ; i++ {
		p := *(**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(ptr)) + uintptr(i)*unsafe.Sizeof(*ptr)))
		if p == nil {
			break
		}
		result = append(result, C.GoString(p))
	}
	return result
}

// PropertyNames returns all available property names.
func (s Slide) PropertyNames() ([]string, error) {
	ptr := C.openslide_get_property_names(s.ptr)
	if err := C.openslide_get_error(s.ptr); err != nil {
		C.openslide_close(s.ptr)
		return nil, &OpenSlideError{
			Operation: "PropertyNames",
			Message:   C.GoString(err),
		}
	}
	return cStringArrayToSlice(ptr), nil
}

// PropertyValue returns the value of a specific property.
func (s Slide) PropertyValue(name string) (string, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	val := C.openslide_get_property_value(s.ptr, cName)
	if err := C.openslide_get_error(s.ptr); err != nil {
		C.openslide_close(s.ptr)
		return "", &OpenSlideError{
			Operation: "PropertyValue",
			Message:   C.GoString(err),
		}
	}
	if val == nil {
		return "", nil
	}
	return C.GoString(val), nil
}

// Properties returns all properties as a map.
func (s Slide) Properties() (map[string]string, error) {
	props := make(map[string]string)

	names, err := s.PropertyNames()
	if err != nil {
		return nil, err
	}

	for _, key := range names {
		val, err := s.PropertyValue(key)
		if err != nil {
			return nil, err
		}
		if val != "" {
			props[key] = val
		}
	}
	return props, nil
}

// Spacing returns the physical spacing in microns per pixel.
func (s Slide) Spacing() ([2]float64, error) {
	mppXVal, err := s.PropertyValue(PropMPPX)
	if err != nil {
		return [2]float64{}, err
	}

	mppYVal, err := s.PropertyValue(PropMPPY)
	if err != nil {
		return [2]float64{}, err
	}

	if mppXVal == "" || mppYVal == "" {
		return [2]float64{}, errors.New("missing MPP properties")
	}

	x, errX := strconv.ParseFloat(mppXVal, 64)
	y, errY := strconv.ParseFloat(mppYVal, 64)
	if errX != nil || errY != nil {
		return [2]float64{}, errors.New("invalid MPP values")
	}

	return [2]float64{x, y}, nil
}

// Thumbnail generates a thumbnail image with the specified maximum size.
func (s Slide) Thumbnail(size int) (image.Image, error) {
	dims, err := s.LargestLevelDimensions()
	if err != nil {
		return nil, err
	}

	var downsample float64
	for _, d := range dims {
		if ds := float64(d) / float64(size); ds > downsample {
			downsample = ds
		}
	}

	level, err := s.BestLevelForDownsample(downsample)
	if err != nil {
		return nil, err
	}

	levelDims, err := s.LevelDimensions(level)
	if err != nil {
		return nil, err
	}

	srcW, srcH := levelDims[0], levelDims[1]

	imgData, err := s.ReadRegion(0, 0, level, srcW, srcH)
	if err != nil {
		return nil, err
	}

	// Convert raw RGBA bytes to image.Image
	img, err := RegionToImage(imgData, srcW, srcH)
	if err != nil {
		return nil, err
	}

	// Compute output size maintaining aspect ratio
	var outW, outH int
	if srcW > srcH {
		outW = size
		outH = int(float64(srcH) / float64(srcW) * float64(size))
	} else {
		outH = size
		outW = int(float64(srcW) / float64(srcH) * float64(size))
	}

	thumb := image.NewRGBA(image.Rect(0, 0, outW, outH))
	draw.BiLinear.Scale(thumb, thumb.Bounds(), img, img.Bounds(), draw.Over, nil)
	return thumb, nil
}

// Version returns the OpenSlide library version.
func Version() string {
	return C.GoString(C.openslide_get_version())
}

// Well-known OpenSlide property keys.
const (
	PropBackgroundColor = "openslide.background-color"
	PropBoundsHeight    = "openslide.bounds-height"
	PropBoundsWidth     = "openslide.bounds-width"
	PropBoundsX         = "openslide.bounds-x"
	PropBoundsY         = "openslide.bounds-y"
	PropMPPX            = "openslide.mpp-x"
	PropMPPY            = "openslide.mpp-y"
	PropObjectivePower  = "openslide.objective-power"
)
