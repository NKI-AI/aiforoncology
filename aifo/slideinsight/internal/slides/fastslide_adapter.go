// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package slides

import (
	"fmt"
	"image"
	"strconv"

	"aifo.dev/aifo/fastslide_go/fastslide"
)

// FastSlideAdapter wraps a FastSlide SlideReader and implements the Slide interface.
type FastSlideAdapter struct {
	reader           *fastslide.SlideReader
	downsample       []float64 // cached downsamples
	properties       map[string]string
	levelWidths      []int
	levelHeights     []int
	channelMetadata  []ChannelMetadata // cached channel metadata
	isSpectralSlide  bool              // whether this is a spectral slide
	spectralComputed bool              // whether spectral metadata has been computed
}

// NewFastSlideAdapter creates a new adapter for a FastSlide SlideReader.
func NewFastSlideAdapter(path string) (Slide, error) {
	reader, err := fastslide.Open(path)
	if err != nil {
		return nil, err
	}

	adapter := &FastSlideAdapter{
		reader:     reader,
		properties: make(map[string]string),
	}

	// Initialize level dimensions and downsample factors
	levels, err := reader.LevelCount()
	if err != nil {
		reader.Close()
		return nil, err
	}

	adapter.downsample = make([]float64, levels)
	adapter.levelWidths = make([]int, levels)
	adapter.levelHeights = make([]int, levels)

	// Get base dimensions
	baseWidth, baseHeight, err := reader.BaseDimensions()
	if err != nil {
		reader.Close()
		return nil, err
	}

	// Fill level dimensions and calculate downsamples
	for i := 0; i < levels; i++ {
		w, h, err := reader.LevelDimensions(i)
		if err != nil {
			reader.Close()
			return nil, err
		}
		adapter.levelWidths[i] = int(w)
		adapter.levelHeights[i] = int(h)

		// Get the downsample factor from FastSlide
		downsample, err := reader.LevelDownsample(i)
		if err != nil {
			reader.Close()
			return nil, err
		}
		adapter.downsample[i] = downsample
	}

	// Add some basic properties
	adapter.properties["fastslide.levels"] = strconv.Itoa(levels)
	adapter.properties["fastslide.width"] = strconv.FormatUint(uint64(baseWidth), 10)
	adapter.properties["fastslide.height"] = strconv.FormatUint(uint64(baseHeight), 10)

	// Try to get format and scanner info
	if format, err := reader.FormatName(); err == nil && format != "" {
		adapter.properties["fastslide.format"] = format
	}

	// Try to get resolution data and set generalized property names
	props, err := reader.GetProperties()
	if err == nil {
		if props.MppX > 0 && props.MppY > 0 {
			adapter.properties["mpp-x"] = strconv.FormatFloat(props.MppX, 'f', -1, 64)
			adapter.properties["mpp-y"] = strconv.FormatFloat(props.MppY, 'f', -1, 64)
		}
		if props.ScannerModel != "" {
			adapter.properties["fastslide.scanner-model"] = props.ScannerModel
		}
		if props.ScanDate != "" {
			adapter.properties["fastslide.scan-date"] = props.ScanDate
		}
		if props.ObjectiveMagnification > 0 {
			adapter.properties["fastslide.objective-magnification"] = strconv.FormatFloat(props.ObjectiveMagnification, 'f', -1, 64)
		}
	}

	// Initialize spectral metadata (lazy loading)
	adapter.initSpectralInfo()

	return adapter, nil
}

// initSpectralInfo initializes spectral information for the slide.
func (a *FastSlideAdapter) initSpectralInfo() {
	if a.spectralComputed {
		return
	}

	// Try to get channel metadata to determine if this is a spectral slide
	channelMetadata, err := a.reader.GetChannelMetadata()
	if err != nil || len(channelMetadata) == 0 {
		a.isSpectralSlide = false
		a.spectralComputed = true
		return
	}

	// Convert FastSlide channel metadata to our format
	a.channelMetadata = make([]ChannelMetadata, len(channelMetadata))
	for i, ch := range channelMetadata {
		a.channelMetadata[i] = ChannelMetadata{
			Name:      ch.Name,
			Biomarker: ch.Biomarker,
			Color:     [3]uint8{ch.Color.R, ch.Color.G, ch.Color.B},
		}
	}

	a.isSpectralSlide = true
	a.spectralComputed = true
}

// Close releases resources associated with the slide.
func (a *FastSlideAdapter) Close() {
	a.reader.Close()
}

// LevelCount returns the number of levels in the pyramid.
func (a *FastSlideAdapter) LevelCount() (int, error) {
	return a.reader.LevelCount()
}

// LargestLevelDimensions returns the dimensions of level 0 (most detailed).
func (a *FastSlideAdapter) LargestLevelDimensions() ([2]int, error) {
	return [2]int{a.levelWidths[0], a.levelHeights[0]}, nil
}

// LevelDimensions returns the dimensions for the specified level.
func (a *FastSlideAdapter) LevelDimensions(level int) ([2]int, error) {
	if level < 0 || level >= len(a.levelWidths) {
		return [2]int{0, 0}, fmt.Errorf("level %d out of bounds", level)
	}
	return [2]int{a.levelWidths[level], a.levelHeights[level]}, nil
}

// LevelDownsample returns the downsample factor of a level.
func (a *FastSlideAdapter) LevelDownsample(level int) (float64, error) {
	if level < 0 || level >= len(a.downsample) {
		return 1.0, fmt.Errorf("level %d out of bounds", level)
	}
	return a.downsample[level], nil
}

// LevelDownsamples returns all downsample factors.
func (a *FastSlideAdapter) LevelDownsamples() ([]float64, error) {
	result := make([]float64, len(a.downsample))
	copy(result, a.downsample)
	return result, nil
}

// BestLevelForDownsample returns the closest level for the requested downsample.
func (a *FastSlideAdapter) BestLevelForDownsample(downsample float64) (int, error) {
	return a.reader.BestLevelForDownsample(downsample)
}

// ReadRegion reads a region at the specified level and location using basic RGB conversion.
func (a *FastSlideAdapter) ReadRegion(x, y, level, w, h int) (image.Image, error) {
	fastslideImage, err := a.ReadRegionAsFastslideImage(x, y, level, w, h)
	if err != nil {
		return nil, err
	}
	defer fastslideImage.Close()

	return fastslideImage.ToGoImage()
}

// ReadRegionAsFastslideImage reads a region with optional spectral conversion parameters.
func (a *FastSlideAdapter) ReadRegionAsFastslideImage(x, y, level, w, h int) (*fastslide.Image, error) {
	// Check if level is valid
	if level < 0 || level >= len(a.levelWidths) {
		return nil, fmt.Errorf("level %d out of bounds (available levels: 0-%d)", level, len(a.levelWidths)-1)
	}

	// Safety check against negative dimensions
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("invalid tile size: width=%d, height=%d", w, h)
	}

	// Get level dimensions
	levelWidth := a.levelWidths[level]
	levelHeight := a.levelHeights[level]

	// Coordinates are already in level-native coordinate system
	levelX := x
	levelY := y

	// Check if request is completely outside level bounds
	if levelX >= levelWidth || levelY >= levelHeight {
		return nil, fmt.Errorf("tile region is outside slide bounds, x=%d, y=%d, levelWidth=%d, levelHeight=%d", levelX, levelY, levelWidth, levelHeight)
	}

	// Read the region from FastSlide (returns *fastslide.Image)
	fastslideImage, err := a.reader.ReadRegion(uint32(levelX), uint32(levelY), uint32(w), uint32(h), level)
	if err != nil {
		return nil, err
	}

	return fastslideImage, nil
}

// Mpp returns the resolution in microns per pixel (MPP) for the slide.
func (a *FastSlideAdapter) Mpp() (float64, float64, error) {
	props, err := a.reader.GetProperties()
	if err != nil {
		return 0, 0, err
	}
	return props.MppX, props.MppY, nil
}

// PropertyValue returns the value of a specific property.
func (a *FastSlideAdapter) PropertyValue(name string) (string, error) {
	if value, exists := a.properties[name]; exists {
		return value, nil
	}
	return "", nil
}

// Properties returns all properties as a map.
func (a *FastSlideAdapter) Properties() (map[string]string, error) {
	result := make(map[string]string, len(a.properties))
	for k, v := range a.properties {
		result[k] = v
	}
	return result, nil
}

// ===== Spectral support methods =====

// IsSpectral returns true if this slide contains spectral/multiplex data.
func (a *FastSlideAdapter) IsSpectral() bool {
	a.initSpectralInfo()
	return a.isSpectralSlide
}

// GetChannelMetadata returns channel metadata for spectral slides.
func (a *FastSlideAdapter) GetChannelMetadata() ([]ChannelMetadata, error) {
	a.initSpectralInfo()
	if !a.isSpectralSlide {
		return nil, nil
	}

	// Return a copy to prevent external modification
	result := make([]ChannelMetadata, len(a.channelMetadata))
	copy(result, a.channelMetadata)
	return result, nil
}

// createRGBImageFromData creates an RGB image from a byte slice containing RGB data.
func createRGBImageFromData(data []uint8, width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Convert RGB to RGBA (add alpha channel)
	for i := 0; i < width*height; i++ {
		srcIdx := i * 3
		dstIdx := i * 4

		if srcIdx+2 < len(data) {
			img.Pix[dstIdx] = data[srcIdx]     // R
			img.Pix[dstIdx+1] = data[srcIdx+1] // G
			img.Pix[dstIdx+2] = data[srcIdx+2] // B
			img.Pix[dstIdx+3] = 255            // A (fully opaque)
		}
	}

	return img
}
