// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
package slides

import (
	"testing"
)

// TestOpenSlideFactoryExtensionDetection tests that the factory correctly detects slide formats by extension
func TestOpenSlideFactoryExtensionDetection(t *testing.T) {
	// These tests just verify the extension detection logic in the factory
	testCases := []struct {
		filename    string
		expectedFmt string
	}{
		{"sample.svs", FormatOpenSlide},
		{"sample.ndpi", FormatOpenSlide},
		{"sample.scn", FormatOpenSlide},
		{"sample.mrxs", FormatOpenSlide},
		{"sample.tif", FormatOpenSlide}, // Note: tif can be both
		{"sample.vms", FormatOpenSlide},
		{"sample.vmu", FormatOpenSlide},
		{"sample.bif", FormatOpenSlide},
		{"sample.tiff", FormatTIFF},
		{"sample.unknown", ""}, // should use default fallback behavior
	}

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			// Manually determine what format would be selected
			var format string
			ext := getExtension(tc.filename)
			if openslideExtensions[ext] {
				format = FormatOpenSlide
			} else if ext == ".tif" || ext == ".tiff" {
				format = FormatTIFF
			}

			if format != tc.expectedFmt {
				t.Errorf("Expected format %s for %s, got %s", tc.expectedFmt, tc.filename, format)
			}
		})
	}
}

// Helper function mimicking the extension extraction logic in OpenSlide
func getExtension(filename string) string {
	// Find last dot
	lastDot := -1
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			lastDot = i
			break
		}
	}
	if lastDot == -1 {
		return ""
	}
	return filename[lastDot:]
}
