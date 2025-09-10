// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package validation

import (
	"fmt"
	"strings"

	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type (
	// XValidator wraps the go-playground validator
	XValidator struct {
		validator *validator.Validate
	}

	// ValidationError represents a single validation error
	ValidationError struct {
		FailedField string      `json:"field"`
		Tag         string      `json:"tag"`
		Value       interface{} `json:"value"`
	}
)

// New creates a new validator instance
func New() *XValidator {
	return &XValidator{
		validator: validator.New(),
	}
}

// Validate validates a struct and returns validation errors
func (v *XValidator) Validate(data interface{}) []ValidationError {
	var validationErrors []ValidationError

	err := v.validator.Struct(data)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			validationError := ValidationError{
				FailedField: err.Field(),
				Tag:         err.Tag(),
				Value:       err.Value(),
			}
			validationErrors = append(validationErrors, validationError)
		}
	}

	return validationErrors
}

// ValidateStruct validates a struct and returns a fiber error if validation fails
func (v *XValidator) ValidateStruct(c *fiber.Ctx, data interface{}) error {
	if errs := v.Validate(data); len(errs) > 0 {
		errMsgs := make([]string, 0)
		for _, err := range errs {
			errMsgs = append(errMsgs, formatValidationError(err))
		}

		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error: strings.Join(errMsgs, " and "),
		})
	}
	return nil
}

// formatValidationError formats a validation error into a user-friendly message
func formatValidationError(err ValidationError) string {
	switch err.Tag {
	case "required":
		return fmt.Sprintf("[%s]: is required", err.FailedField)
	case "file_exists":
		return fmt.Sprintf("[%s]: file '%v' does not exist", err.FailedField, err.Value)
	case "file_ext":
		return fmt.Sprintf("[%s]: invalid file extension for '%v'", err.FailedField, err.Value)
	case "slide_uid":
		return fmt.Sprintf("[%s]: invalid slide UID format '%v'", err.FailedField, err.Value)
	case "oneof":
		return fmt.Sprintf("[%s]: '%v' must be one of the allowed values", err.FailedField, err.Value)
	default:
		return fmt.Sprintf("[%s]: '%v' | Needs to implement '%s'", err.FailedField, err.Value, err.Tag)
	}
}

// RegisterValidation registers a custom validation function
func (v *XValidator) RegisterValidation(tag string, fn validator.Func) error {
	return v.validator.RegisterValidation(tag, fn)
}

// Global validator instance with custom validators
var GlobalValidator *XValidator

// init initializes the global validator with custom validators
func init() {
	var err error
	GlobalValidator, err = InitializeWithCustomValidators()
	if err != nil {
		// Fallback to basic validator if custom validators fail
		GlobalValidator = New()
	}
}
