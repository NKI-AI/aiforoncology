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

// Mask-specific error types
var (
	// ErrMaskFormatNotSupported is returned when the mask file format is not supported
	ErrMaskFormatNotSupported = fmt.Errorf("%w: mask format not supported", ErrInvalidInput)

	// ErrMaskUriInvalid is returned when the mask URI is invalid
	ErrMaskUriInvalid = fmt.Errorf("%w: invalid mask URI", ErrInvalidInput)

	// ErrMaskSlideIDInvalid is returned when mask's slide ID is invalid
	ErrMaskSlideIDInvalid = fmt.Errorf("%w: invalid slide ID for mask", ErrInvalidInput)

	// ErrMaskSlideNotFound is returned when the slide referenced by a mask doesn't exist
	ErrMaskSlideNotFound = fmt.Errorf("%w: slide for mask not found", ErrNotFound)

	// ErrMaskTileGeneration is returned when mask tile generation fails
	ErrMaskTileGeneration = fmt.Errorf("%w: error generating mask tile", ErrInternal)

	// ErrMaskMetadataInvalid is returned when mask metadata extraction fails
	ErrMaskMetadataInvalid = fmt.Errorf("%w: invalid mask metadata", ErrInvalidInput)
)
