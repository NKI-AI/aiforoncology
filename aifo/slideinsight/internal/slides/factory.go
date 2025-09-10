// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package slides

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Supported file formats
const (
	FormatOpenSlide = "openslide" // For formats supported by OpenSlide (SVS, NDPI, etc.)
	FormatTIFF      = "tiff"      // For single-channel TIFF files
	FormatQpTIFF    = "qptiff"    // For multi-channel QPTIFF files
)

// SlideOpenOptions contains optional settings for opening slides.
type SlideOpenOptions struct {
	FormatHint string // Format hint for slide type detection
}

// OpenSlide creates a slide using the appropriate implementation based on file format.
// It tries to determine the format from the file extension or format hint if provided.
func OpenSlide(path string, formatHint string) (Slide, error) {
	return OpenSlideWithOptions(path, &SlideOpenOptions{
		FormatHint: formatHint,
	})
}

// OpenSlideWithOptions creates a slide with advanced options including spectral settings.
func OpenSlideWithOptions(path string, options *SlideOpenOptions) (Slide, error) {
	if options == nil {
		options = &SlideOpenOptions{}
	}

	formatHint := options.FormatHint
	if formatHint == "" {
		// Try to determine format from file extension
		ext := strings.ToLower(filepath.Ext(path))
		if openslideExtensions[ext] {
			formatHint = FormatOpenSlide
		} else if qptiffExtensions[ext] {
			formatHint = FormatQpTIFF
		} else if ext == ".tif" || ext == ".tiff" {
			formatHint = FormatTIFF
		}
	}

	// Create the appropriate slide implementation
	var slide Slide
	var err error

	switch formatHint {
	case FormatOpenSlide:
		slide, err = NewOpenSlideAdapter(path)
		if err != nil {
			fmt.Printf("Factory: OpenSlideAdapter failed for %s: %v\n", path, err)
		} else {
		}
	case FormatTIFF:
		slide, err = NewTiffAdapter(path)
		if err != nil {
			fmt.Printf("Factory: TiffAdapter failed for %s: %v\n", path, err)
		} else {
		}
	case FormatQpTIFF:
		slide, err = NewFastSlideAdapter(path)
		if err != nil {
			fmt.Printf("Factory: FastSlideAdapter failed for %s: %v\n", path, err)
		}
	default:
		// If format is still unknown, try OpenSlide, then TIFF, then FastSlide
		slide, err = NewOpenSlideAdapter(path)
		if err != nil {
			fmt.Printf("Factory: OpenSlideAdapter failed for %s: %v\n", path, err)
			// If OpenSlide fails, try TIFF
			slide, err = NewTiffAdapter(path)
			if err != nil {
				fmt.Printf("Factory: TiffAdapter failed for %s: %v\n", path, err)
				// If TIFF fails, try FastSlide
				slide, err = NewFastSlideAdapter(path)
				if err != nil {
					fmt.Printf("Factory: FastSlideAdapter failed for %s: %v\n", path, err)
				} else {
				}
			} else {
			}
		} else {
		}
	}

	if err != nil {
		fmt.Printf("Factory: All adapters failed for %s\n", path)
		return nil, err
	}

	return slide, nil
}

// OpenSlide supported extensions
var openslideExtensions = map[string]bool{
	".svs":  true, // Aperio
	".ndpi": true, // Hamamatsu
	".scn":  true, // Leica
	".mrxs": true, // MIRAX
	".tif":  true, // BigTIFF
	".vms":  true, // Hamamatsu
	".vmu":  true, // Hamamatsu
	".bif":  true, // Ventana
}

// QpTiff supported extensions
// TODO: Add support for other QpTiff extensions
var qptiffExtensions = map[string]bool{
	".qptiff": true, // Multi-channel multiplex TIFF
	".qp":     true, // Alternative extension
}
