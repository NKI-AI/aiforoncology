// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package slides

import (
	"context"
	"fmt"

	"aifo.dev/aifo/openslide_go/openslide"
	"aifo.dev/aifo/slideinsight/internal/server/domain"
)

// GetSlideMetadata retrieves detailed metadata for a specific slide
func (s *slidesService) GetSlideMetadata(ctx context.Context, slideUID string) (domain.SlideMetadata, error) {
	// First get the basic slide info
	slide, err := s.GetSlideByUID(ctx, slideUID)
	if err != nil {
		return domain.SlideMetadata{}, err
	}

	// Get cached slide or open it
	slideObj, err := openslide.Open(slide.SlideURI)
	if err != nil {
		return domain.SlideMetadata{}, err
	}

	// Get vendor information
	vendor := "Unknown"
	properties, err := slideObj.Properties()
	if err != nil {
		return domain.SlideMetadata{}, fmt.Errorf("failed to get slide properties: %w", err)
	}

	if v, ok := properties["openslide.vendor"]; ok {
		vendor = v
	}

	// Get magnification
	magnification := "Unknown"
	if m, ok := properties["openslide.objective-power"]; ok {
		magnification = m
	}

	// Determine zoom levels
	// This is a simplification - you may need to adjust this based on actual slide dimensions
	maxLevel := 0
	if slide.SlideWidth > 0 && slide.SlideHeight > 0 {
		maxDimension := slide.SlideWidth
		if slide.SlideHeight > maxDimension {
			maxDimension = slide.SlideHeight
		}

		// TODO: Make tile size configurable
		// Calculate how many times we can divide by 2 until we reach a reasonable minimum size
		for size := maxDimension; size > 256; size /= 2 {
			maxLevel++
		}
	}

	return domain.SlideMetadata{
		SlideUID:      slideUID,
		SlideName:     slide.SlideName,
		MinLevel:      0,
		MaxLevel:      maxLevel,
		TileSize:      256, // Standard tile size, TODO: Make configurable
		Format:        "jpg",
		SlideMpp:      slide.SlideMpp,
		SlideWidth:    slide.SlideWidth,
		SlideHeight:   slide.SlideHeight,
		Vendor:        vendor,
		Magnification: magnification,
	}, nil
}
