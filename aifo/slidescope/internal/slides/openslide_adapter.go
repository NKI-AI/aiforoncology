// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
package slides

import (
	"image"

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
	return a.slide.ReadRegion(x, y, level, w, h)
}

// PropertyValue returns the value of a specific property.
func (a *OpenSlideAdapter) PropertyValue(name string) (string, error) {
	return a.slide.PropertyValue(name)
}

// Properties returns all properties as a map.
func (a *OpenSlideAdapter) Properties() (map[string]string, error) {
	return a.slide.Properties()
}
