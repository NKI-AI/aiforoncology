// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package errors

import (
	"fmt"
)

// Resource-specific error types for different domain entities
var (
	// Study errors
	ErrStudyNotFound       = fmt.Errorf("%w: study not found", ErrNotFound)
	ErrStudyIDEmpty        = fmt.Errorf("%w: study ID cannot be empty", ErrInvalidInput)
	ErrStudyAlreadyDeleted = fmt.Errorf("%w: study already deleted", ErrInvalidInput)
	ErrStudyNotDeleted     = fmt.Errorf("%w: study not deleted", ErrInvalidInput)

	// Case errors
	ErrCaseNotFound       = fmt.Errorf("%w: case not found", ErrNotFound)
	ErrCaseIDEmpty        = fmt.Errorf("%w: case ID cannot be empty", ErrInvalidInput)
	ErrCaseAlreadyDeleted = fmt.Errorf("%w: case already deleted", ErrInvalidInput)
	ErrCaseNotDeleted     = fmt.Errorf("%w: case not deleted", ErrInvalidInput)

	// Slide errors
	ErrSlideIDEmpty        = fmt.Errorf("%w: slide ID cannot be empty", ErrInvalidInput)
	ErrSlideAlreadyDeleted = fmt.Errorf("%w: slide already deleted", ErrInvalidInput)
	ErrSlideNotDeleted     = fmt.Errorf("%w: slide not deleted", ErrInvalidInput)

	// Tenant errors
	ErrTenantUIDEmpty = fmt.Errorf("%w: tenant UID cannot be empty", ErrInvalidInput)

	// Mask/Raster annotation errors (Note: ErrMaskNotFound is defined in errors.go)
	ErrMaskAlreadyDeleted = fmt.Errorf("%w: mask already deleted", ErrInvalidInput)
	ErrMaskNotDeleted     = fmt.Errorf("%w: mask not deleted", ErrInvalidInput)

	// Vector annotation errors
	ErrVectorAnnotationNotFound       = fmt.Errorf("%w: vector annotation not found", ErrNotFound)
	ErrVectorAnnotationAlreadyDeleted = fmt.Errorf("%w: vector annotation already deleted", ErrInvalidInput)
	ErrVectorAnnotationNotDeleted     = fmt.Errorf("%w: vector annotation not deleted", ErrInvalidInput)

	// Domain errors
	ErrDomainNotFound = fmt.Errorf("%w: domain not found", ErrNotFound)

	// Setting errors
	ErrSettingNotFound      = fmt.Errorf("%w: setting not found", ErrNotFound)
	ErrSettingKeyEmpty      = fmt.Errorf("%w: setting key cannot be empty", ErrInvalidInput)
	ErrSettingAlreadyExists = fmt.Errorf("%w: setting already exists", ErrAlreadyExists)
)

// Generic resource error helper functions

// NewResourceNotFoundError creates a generic resource not found error with context
func NewResourceNotFoundError(resourceType, identifier string) error {
	return WithDetails(ErrNotFound, "%s with ID '%s' not found", resourceType, identifier)
}

// NewStudyNotFoundError creates a study not found error with ID context
func NewStudyNotFoundError(studyUID string) error {
	return WithDetails(ErrStudyNotFound, "ID '%s'", studyUID)
}

// NewStudyAlreadyDeletedError creates a study already deleted error with ID context
func NewStudyAlreadyDeletedError(studyUID string) error {
	return WithDetails(ErrStudyAlreadyDeleted, "ID '%s'", studyUID)
}

// NewStudyNotDeletedError creates a study not deleted error with ID context
func NewStudyNotDeletedError(studyUID string) error {
	return WithDetails(ErrStudyNotDeleted, "ID '%s'", studyUID)
}

// NewCaseNotFoundError creates a case not found error with ID context
func NewCaseNotFoundError(caseUID string) error {
	return WithDetails(ErrCaseNotFound, "ID '%s'", caseUID)
}

// NewCaseAlreadyDeletedError creates a case already deleted error with ID context
func NewCaseAlreadyDeletedError(caseUID string) error {
	return WithDetails(ErrCaseAlreadyDeleted, "ID '%s'", caseUID)
}

// NewCaseNotDeletedError creates a case not deleted error with ID context
func NewCaseNotDeletedError(caseUID string) error {
	return WithDetails(ErrCaseNotDeleted, "ID '%s'", caseUID)
}

// NewSlideNotFoundError creates a slide not found error with ID context
func NewSlideNotFoundError(slideUID string) error {
	return WithDetails(ErrSlideNotFound, "ID '%s'", slideUID)
}

// NewSlideAlreadyDeletedError creates a slide already deleted error with ID context
func NewSlideAlreadyDeletedError(slideUID string) error {
	return WithDetails(ErrSlideAlreadyDeleted, "ID '%s'", slideUID)
}

// NewSlideNotDeletedError creates a slide not deleted error with ID context
func NewSlideNotDeletedError(slideUID string) error {
	return WithDetails(ErrSlideNotDeleted, "ID '%s'", slideUID)
}

// NewTenantNotFoundError creates a tenant not found error with ID context
func NewTenantNotFoundError(tenantUID string) error {
	return WithDetails(ErrTenantNotFound, "ID '%s'", tenantUID)
}

// NewMaskNotFoundError creates a mask not found error with ID context
func NewMaskNotFoundError(maskUID string) error {
	return WithDetails(ErrMaskNotFound, "ID '%s'", maskUID)
}

// NewMaskAlreadyDeletedError creates a mask already deleted error with ID context
func NewMaskAlreadyDeletedError(maskUID string) error {
	return WithDetails(ErrMaskAlreadyDeleted, "ID '%s'", maskUID)
}

// NewMaskNotDeletedError creates a mask not deleted error with ID context
func NewMaskNotDeletedError(maskUID string) error {
	return WithDetails(ErrMaskNotDeleted, "ID '%s'", maskUID)
}

// NewVectorAnnotationNotFoundError creates a vector annotation not found error with ID context
func NewVectorAnnotationNotFoundError(vectorUID string) error {
	return WithDetails(ErrVectorAnnotationNotFound, "ID '%s'", vectorUID)
}

// NewVectorAnnotationAlreadyDeletedError creates a vector annotation already deleted error with ID context
func NewVectorAnnotationAlreadyDeletedError(vectorUID string) error {
	return WithDetails(ErrVectorAnnotationAlreadyDeleted, "ID '%s'", vectorUID)
}

// NewVectorAnnotationNotDeletedError creates a vector annotation not deleted error with ID context
func NewVectorAnnotationNotDeletedError(vectorUID string) error {
	return WithDetails(ErrVectorAnnotationNotDeleted, "ID '%s'", vectorUID)
}

// NewDomainNotFoundError creates a domain not found error with ID context
func NewDomainNotFoundError(domainID interface{}) error {
	return WithDetails(ErrDomainNotFound, "ID '%v'", domainID)
}

// NewSettingNotFoundError creates a setting not found error with tenant and key context
func NewSettingNotFoundError(tenantID int, key string) error {
	return WithDetails(ErrSettingNotFound, "tenant %d, key '%s'", tenantID, key)
}

// NewSettingAlreadyExistsError creates a setting already exists error with tenant and key context
func NewSettingAlreadyExistsError(tenantID int, key string) error {
	return WithDetails(ErrSettingAlreadyExists, "tenant %d, key '%s'", tenantID, key)
}
