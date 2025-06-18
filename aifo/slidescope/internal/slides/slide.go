// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
package slides

import (
	"image"
)

// Slide defines the common interface that all slide implementations must satisfy.
// This allows the XYZ tile generator to work with different slide backends.
type Slide interface {
	// Close releases resources associated with the slide.
	Close()

	// LevelCount returns the number of levels in the pyramid.
	LevelCount() (int, error)

	// LargestLevelDimensions returns the dimensions of level 0 (most detailed).
	LargestLevelDimensions() ([2]int, error)

	// LevelDimensions returns the dimensions for the specified level.
	LevelDimensions(level int) ([2]int, error)

	// LevelDownsample returns the downsample factor of a level.
	LevelDownsample(level int) (float64, error)

	// LevelDownsamples returns all downsample factors.
	LevelDownsamples() ([]float64, error)

	// BestLevelForDownsample returns the closest level for the requested downsample.
	BestLevelForDownsample(downsample float64) (int, error)

	// ReadRegion reads a region at the specified level and location.
	ReadRegion(x, y, level, w, h int) (image.Image, error)

	// PropertyValue returns the value of a specific property.
	PropertyValue(name string) (string, error)

	// Properties returns all properties as a map.
	Properties() (map[string]string, error)
}

// SlideOpener is a function type that can open a slide file.
type SlideOpener func(path string) (Slide, error)
