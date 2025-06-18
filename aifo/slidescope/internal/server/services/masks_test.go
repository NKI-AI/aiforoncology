// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMaskInfoString tests the MaskInfo String method
func TestMaskInfoString(t *testing.T) {
	info := MaskInfo{
		MaskURI: "path/to/mask.tiff",
		SlideID: "slide123",
	}

	serialized := info.String()
	assert.Equal(t, "path/to/mask.tiff|slide123", serialized)
}

// TestParseMaskInfo tests the ParseMaskInfo function
func TestParseMaskInfo(t *testing.T) {
	// Test with valid input
	info := ParseMaskInfo("path/to/mask.tiff|slide123")
	assert.Equal(t, "path/to/mask.tiff", info.MaskURI)
	assert.Equal(t, "slide123", info.SlideID)

	// Test with invalid input
	invalid := ParseMaskInfo("invalid")
	assert.Equal(t, "", invalid.MaskURI)
	assert.Equal(t, "", invalid.SlideID)
}
