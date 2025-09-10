// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package validation

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/go-playground/validator/v10"
)

// RegisterCustomValidators registers all custom validation functions
func RegisterCustomValidators(v *validator.Validate) error {
	// Register file existence validator
	if err := v.RegisterValidation("file_exists", fileExistsValidator); err != nil {
		return err
	}

	// Register file extension validator
	if err := v.RegisterValidation("file_ext", fileExtValidator); err != nil {
		return err
	}

	// Register slide UID format validator (if you have specific format requirements)
	if err := v.RegisterValidation("slide_uid", slideUIDValidator); err != nil {
		return err
	}

	return nil
}

// fileExistsValidator validates that a file exists at the given path
func fileExistsValidator(fl validator.FieldLevel) bool {
	filePath := fl.Field().String()
	if filePath == "" {
		return true // Allow empty values (use required tag separately if needed)
	}

	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// fileExtValidator validates file extension
// Usage: validate:"file_ext=.geojson" or validate:"file_ext=.geojson;.json"
func fileExtValidator(fl validator.FieldLevel) bool {
	filePath := fl.Field().String()
	if filePath == "" {
		return true // Allow empty values
	}

	param := fl.Param()
	if param == "" {
		return false // No extensions specified
	}

	allowedExts := strings.Split(param, ";")
	fileExt := strings.ToLower(filepath.Ext(filePath))

	for _, allowedExt := range allowedExts {
		if fileExt == strings.ToLower(strings.TrimSpace(allowedExt)) {
			return true
		}
	}
	return false
}

// slideUIDValidator validates slide UID format (customize as needed)
func slideUIDValidator(fl validator.FieldLevel) bool {
	slideUID := fl.Field().String()
	if slideUID == "" {
		return true // Allow empty values
	}

	// Example validation: slide UID should be alphanumeric and between 3-50 characters
	if len(slideUID) < 3 || len(slideUID) > 50 {
		return false
	}

	for _, char := range slideUID {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_') {
			return false
		}
	}

	return true
}

// InitializeWithCustomValidators creates a new validator with custom validators registered
func InitializeWithCustomValidators() (*XValidator, error) {
	xVal := New()
	if err := RegisterCustomValidators(xVal.validator); err != nil {
		return nil, err
	}
	return xVal, nil
}
