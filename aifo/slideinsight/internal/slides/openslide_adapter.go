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

	"aifo.dev/aifo/fastslide_go/fastslide"
	"aifo.dev/aifo/openslide_go/openslide"
)

// OpenSlideAdapter wraps an OpenSlide object and implements the Slide interface.
type OpenSlideAdapter struct {
	slide openslide.Slide
}

// NewOpenSlideAdapter creates a new adapter for an OpenSlide object.
func NewOpenSlideAdapter(path string) (Slide, error) {
	slide, err := openslide.Open(path)
	if err != nil {
		return nil, err
	}
	return &OpenSlideAdapter{slide: slide}, nil
}

// Close releases resources associated with the slide.
func (a *OpenSlideAdapter) Close() {
	a.slide.Close()
}

// LevelCount returns the number of levels in the pyramid.
func (a *OpenSlideAdapter) LevelCount() (int, error) {
	return a.slide.LevelCount()
}

// LargestLevelDimensions returns the dimensions of level 0 (most detailed).
func (a *OpenSlideAdapter) LargestLevelDimensions() ([2]int, error) {
	return a.slide.LargestLevelDimensions()
}

// LevelDimensions returns the dimensions for the specified level.
func (a *OpenSlideAdapter) LevelDimensions(level int) ([2]int, error) {
	return a.slide.LevelDimensions(level)
}

// LevelDownsample returns the downsample factor of a level.
func (a *OpenSlideAdapter) LevelDownsample(level int) (float64, error) {
	return a.slide.LevelDownsample(level)
}

// LevelDownsamples returns all downsample factors.
func (a *OpenSlideAdapter) LevelDownsamples() ([]float64, error) {
	return a.slide.LevelDownsamples()
}

// BestLevelForDownsample returns the closest level for the requested downsample.
func (a *OpenSlideAdapter) BestLevelForDownsample(downsample float64) (int, error) {
	return a.slide.BestLevelForDownsample(downsample)
}

// ReadRegion reads a region at the specified level and location.
func (a *OpenSlideAdapter) ReadRegion(x, y, level, w, h int) (image.Image, error) {
	// Convert from level-native coordinates to level-0 coordinates (what OpenSlide expects)
	downsample, err := a.slide.LevelDownsample(level)
	if err != nil {
		return nil, fmt.Errorf("failed to get level downsample: %w", err)
	}

	l0X := int(float64(x) * downsample)
	l0Y := int(float64(y) * downsample)

	data, err := a.slide.ReadRegion(l0X, l0Y, level, w, h)
	if err != nil {
		return nil, err
	}
	return openslide.RegionToImage(data, w, h)
}

// PropertyValue returns the value of a specific property.
func (a *OpenSlideAdapter) PropertyValue(name string) (string, error) {
	// Handle generalized property names by mapping to openslide equivalents
	switch name {
	case "mpp-x":
		return a.slide.PropertyValue("openslide.mpp-x")
	case "mpp-y":
		return a.slide.PropertyValue("openslide.mpp-y")
	default:
		return a.slide.PropertyValue(name)
	}
}

// Properties returns all properties as a map.
func (a *OpenSlideAdapter) Properties() (map[string]string, error) {
	props, err := a.slide.Properties()
	if err != nil {
		return nil, err
	}

	// Add generalized property names for better compatibility
	result := make(map[string]string, len(props)+2) // Reserve space for potential additions
	for k, v := range props {
		result[k] = v
	}

	// Add generalized MPP properties if openslide-specific ones exist
	if mppX, exists := props["openslide.mpp-x"]; exists {
		result["mpp-x"] = mppX
	}
	if mppY, exists := props["openslide.mpp-y"]; exists {
		result["mpp-y"] = mppY
	}

	return result, nil
}

// Mpp returns the resolution in microns per pixel (MPP) for the slide.
func (a *OpenSlideAdapter) Mpp() (float64, float64, error) {
	mppXStr, err := a.slide.PropertyValue("openslide.mpp-x")
	if err != nil {
		return 0, 0, err
	}
	mppYStr, err := a.slide.PropertyValue("openslide.mpp-y")
	if err != nil {
		return 0, 0, err
	}

	var mppX, mppY float64
	if _, err := fmt.Sscanf(mppXStr, "%f", &mppX); err != nil {
		return 0, 0, fmt.Errorf("failed to parse mpp-x value '%s': %w", mppXStr, err)
	}
	if _, err := fmt.Sscanf(mppYStr, "%f", &mppY); err != nil {
		return 0, 0, fmt.Errorf("failed to parse mpp-y value '%s': %w", mppYStr, err)
	}

	return mppX, mppY, nil
}

// ReadRegionAsFastslideImage reads a region with optional spectral conversion parameters.
// OpenSlide doesn't support spectral conversion, so this falls back to regular ReadRegion.
func (a *OpenSlideAdapter) ReadRegionAsFastslideImage(x, y, level, w, h int) (*fastslide.Image, error) {
	goImage, err := a.ReadRegion(x, y, level, w, h)
	if err != nil {
		return nil, err
	}

	// Convert Go image to fastslide.Image
	return fastslide.FromGoImage(goImage)
}

// ===== Spectral support methods =====

// IsSpectral returns false since OpenSlide doesn't handle spectral slides.
func (a *OpenSlideAdapter) IsSpectral() bool {
	return false
}

// GetChannelMetadata returns nil since OpenSlide doesn't support spectral metadata.
func (a *OpenSlideAdapter) GetChannelMetadata() ([]ChannelMetadata, error) {
	return nil, nil
}
