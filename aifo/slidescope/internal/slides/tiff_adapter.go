// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
package slides

import (
	"fmt"
	"image"
	"math"
	"strconv"

	"aifo.dev/aifo/slidescope/internal/tiff"
)

// TiffAdapter wraps a TIFF object and implements the Slide interface.
type TiffAdapter struct {
	tiff         *tiff.Tiff
	downsample   []float64 // cached downsamples
	properties   map[string]string
	levelWidths  []int
	levelHeights []int
}

// NewTiffAdapter creates a new adapter for a TIFF object.
func NewTiffAdapter(path string) (Slide, error) {
	tiffFile, err := tiff.Open(path)
	if err != nil {
		return nil, err
	}

	adapter := &TiffAdapter{
		tiff:       tiffFile,
		properties: make(map[string]string),
	}

	// Initialize level dimensions and downsample factors
	levels := tiffFile.LevelCount()
	adapter.downsample = make([]float64, levels)
	adapter.levelWidths = make([]int, levels)
	adapter.levelHeights = make([]int, levels)

	// Get base dimensions
	baseWidth, baseHeight, err := tiffFile.LevelSize(0)
	if err != nil {
		tiffFile.Close()
		return nil, err
	}

	// Fill level dimensions and calculate downsamples
	for i := 0; i < levels; i++ {
		w, h, err := tiffFile.LevelSize(i)
		if err != nil {
			tiffFile.Close()
			return nil, err
		}
		adapter.levelWidths[i] = int(w)
		adapter.levelHeights[i] = int(h)

		// Calculate downsample factor relative to level 0
		adapter.downsample[i] = float64(baseWidth) / float64(w)
	}

	// Add some basic properties
	adapter.properties["tiff.levels"] = strconv.Itoa(levels)
	adapter.properties["tiff.width"] = strconv.FormatUint(uint64(baseWidth), 10)
	adapter.properties["tiff.height"] = strconv.FormatUint(uint64(baseHeight), 10)

	// Try to get resolution data
	mppX, mppY, err := tiffFile.BaseResolution()
	if err == nil {
		adapter.properties["openslide.mpp-x"] = strconv.FormatFloat(mppX, 'f', -1, 64)
		adapter.properties["openslide.mpp-y"] = strconv.FormatFloat(mppY, 'f', -1, 64)
	}

	return adapter, nil
}

// Close releases resources associated with the slide.
func (a *TiffAdapter) Close() {
	a.tiff.Close()
}

// LevelCount returns the number of levels in the pyramid.
func (a *TiffAdapter) LevelCount() (int, error) {
	return a.tiff.LevelCount(), nil
}

// LargestLevelDimensions returns the dimensions of level 0 (most detailed).
func (a *TiffAdapter) LargestLevelDimensions() ([2]int, error) {
	return [2]int{a.levelWidths[0], a.levelHeights[0]}, nil
}

// LevelDimensions returns the dimensions for the specified level.
func (a *TiffAdapter) LevelDimensions(level int) ([2]int, error) {
	if level < 0 || level >= len(a.levelWidths) {
		return [2]int{0, 0}, fmt.Errorf("level %d out of bounds", level)
	}
	return [2]int{a.levelWidths[level], a.levelHeights[level]}, nil
}

// LevelDownsample returns the downsample factor of a level.
func (a *TiffAdapter) LevelDownsample(level int) (float64, error) {
	if level < 0 || level >= len(a.downsample) {
		return 1.0, fmt.Errorf("level %d out of bounds", level)
	}
	return a.downsample[level], nil
}

// LevelDownsamples returns all downsample factors.
func (a *TiffAdapter) LevelDownsamples() ([]float64, error) {
	result := make([]float64, len(a.downsample))
	copy(result, a.downsample)
	return result, nil
}

// BestLevelForDownsample returns the closest level for the requested downsample.
func (a *TiffAdapter) BestLevelForDownsample(downsample float64) (int, error) {
	if len(a.downsample) == 0 {
		return 0, fmt.Errorf("no levels available")
	}

	bestLevel := 0
	bestDist := math.Abs(downsample - a.downsample[0])

	for i := 1; i < len(a.downsample); i++ {
		dist := math.Abs(downsample - a.downsample[i])
		if dist < bestDist {
			bestDist = dist
			bestLevel = i
		}
	}
	return bestLevel, nil
}

// ReadRegion reads a region at the specified level and location.
func (a *TiffAdapter) ReadRegion(x, y, level, w, h int) (image.Image, error) {
	// Check if level is valid
	if level < 0 || level >= len(a.levelWidths) {
		return nil, fmt.Errorf("level %d out of bounds", level)
	}

	// Safety check against negative dimensions
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("invalid tile size: width=%d, height=%d", w, h)
	}

	// Get level dimensions
	levelWidth := a.levelWidths[level]
	levelHeight := a.levelHeights[level]
	downsample := a.downsample[level]

	x = int(float64(x) / downsample)
	y = int(float64(y) / downsample)

	// Check if request is completely outside level bounds
	if x >= levelWidth || y >= levelHeight {
		return nil, fmt.Errorf("tile region is outside slide bounds, x=%d, y=%d, levelWidth=%d, levelHeight=%d", x, y, levelWidth, levelHeight)
	}

	// TIFF expects uint32 for coordinates and dimensions
	data, err := a.tiff.ReadRegion(uint32(x), uint32(y), level, uint32(w), uint32(h))
	if err != nil {
		// If we can't read the region, return an empty image
		return nil, err
	}

	var img image.Image

	// Convert the raw bytes to an appropriate image type
	// This depends on the bit depth of the TIFF
	switch d := data.(type) {
	case []byte:
		// 8-bit grayscale
		img = createGrayscaleImage(d, w, h)
	case []uint16:
		// 16-bit grayscale
		img = create16bitGrayscaleImage(d, w, h)
	default:
		return nil, fmt.Errorf("unsupported pixel data format %T", d)
	}

	return img, nil
}

// PropertyValue returns the value of a specific property.
func (a *TiffAdapter) PropertyValue(name string) (string, error) {
	return a.properties[name], nil
}

// Properties returns all properties as a map.
func (a *TiffAdapter) Properties() (map[string]string, error) {
	result := make(map[string]string, len(a.properties))
	for k, v := range a.properties {
		result[k] = v
	}
	return result, nil
}

// createGrayscaleImage creates an 8-bit grayscale image from a byte slice.
func createGrayscaleImage(data []byte, width, height int) image.Image {
	img := image.NewGray(image.Rect(0, 0, width, height))
	copy(img.Pix, data)
	return img
}

// create16bitGrayscaleImage creates a 16-bit grayscale image from a uint16 slice.
func create16bitGrayscaleImage(data []uint16, width, height int) image.Image {
	img := image.NewGray16(image.Rect(0, 0, width, height))
	for i, v := range data {
		img.Pix[i*2] = uint8(v >> 8)     // High byte
		img.Pix[i*2+1] = uint8(v & 0xFF) // Low byte
	}
	return img
}
