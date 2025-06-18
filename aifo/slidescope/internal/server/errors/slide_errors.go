// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
package errors

import (
	"fmt"
)

// Slide-specific error types
var (
	// ErrSlideFormatNotSupported is returned when the slide format is not supported
	ErrSlideFormatNotSupported = fmt.Errorf("%w: slide format not supported", ErrInvalidInput)

	// ErrSlideUriInvalid is returned when the slide URI is invalid
	ErrSlideUriInvalid = fmt.Errorf("%w: invalid slide URI", ErrInvalidInput)

	// ErrSlideMetadataInvalid is returned when slide metadata extraction fails
	ErrSlideMetadataInvalid = fmt.Errorf("%w: invalid slide metadata", ErrInvalidInput)

	// ErrSlideTileGeneration is returned when tile generation fails
	ErrSlideTileGeneration = fmt.Errorf("%w: error generating slide tile", ErrInternal)
)
