// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package services

import (
	"encoding/json"
	"time"
)

// ConversionHelpers provides reusable utilities for database to domain model conversions
type ConversionHelpers struct {
	timeFormat string
}

// DefaultConversionHelpers creates conversion helpers with RFC3339 time format
func DefaultConversionHelpers() *ConversionHelpers {
	return &ConversionHelpers{
		timeFormat: time.RFC3339,
	}
}

// WithTimeFormat creates conversion helpers with a custom time format
func WithTimeFormat(format string) *ConversionHelpers {
	return &ConversionHelpers{
		timeFormat: format,
	}
}

// FormatTime formats a time using the helper's configured format
func (h *ConversionHelpers) FormatTime(t time.Time) string {
	return t.Format(h.timeFormat)
}

// FormatOptionalTime formats an optional time pointer, returning nil if input is nil
func (h *ConversionHelpers) FormatOptionalTime(t *time.Time) *string {
	if t == nil {
		return nil
	}
	formatted := h.FormatTime(*t)
	return &formatted
}

// ConvertMetadata converts byte slice metadata to json.RawMessage
func (h *ConversionHelpers) ConvertMetadata(metadata []byte) json.RawMessage {
	if len(metadata) == 0 {
		return nil
	}
	return json.RawMessage(metadata)
}

// SoftDeletionFields represents the common soft deletion fields
type SoftDeletionFields struct {
	DeletedAt *time.Time
	DeletedBy *int
}

// ConvertedSoftDeletionFields represents the converted soft deletion fields
type ConvertedSoftDeletionFields struct {
	DeletedAt *string
	DeletedBy *int
}

// ConvertSoftDeletion converts soft deletion fields using the helper's time format
func (h *ConversionHelpers) ConvertSoftDeletion(fields SoftDeletionFields) ConvertedSoftDeletionFields {
	return ConvertedSoftDeletionFields{
		DeletedAt: h.FormatOptionalTime(fields.DeletedAt),
		DeletedBy: fields.DeletedBy,
	}
}

// ApplySoftDeletion applies converted soft deletion fields to a domain model
// This uses a functional approach to work with any domain type
func ApplySoftDeletion[T any](
	domain T,
	converted ConvertedSoftDeletionFields,
	setDeletedAt func(*T, *string),
	setDeletedBy func(*T, *int),
) T {
	if converted.DeletedAt != nil {
		setDeletedAt(&domain, converted.DeletedAt)
	}
	if converted.DeletedBy != nil {
		setDeletedBy(&domain, converted.DeletedBy)
	}
	return domain
}

// CommonEntityFields represents fields that are common across many entities
type CommonEntityFields struct {
	TenantID   int
	TenantUID  string
	CreatorID  int
	CreatorUID string
	Name       string
	CreatedAt  time.Time
}

// ConvertedCommonFields represents the converted common fields
type ConvertedCommonFields struct {
	TenantID   int
	TenantUID  string
	CreatorID  int
	CreatorUID string
	Name       string
	CreatedAt  string
}

// ConvertCommonFields converts common entity fields using the helper's time format
func (h *ConversionHelpers) ConvertCommonFields(fields CommonEntityFields) ConvertedCommonFields {
	return ConvertedCommonFields{
		TenantID:   fields.TenantID,
		TenantUID:  fields.TenantUID,
		CreatorID:  fields.CreatorID,
		CreatorUID: fields.CreatorUID,
		Name:       fields.Name,
		CreatedAt:  h.FormatTime(fields.CreatedAt),
	}
}

// ConvertDBToDomainWithSoftDeletion is a generic helper for entities that support soft deletion
func ConvertDBToDomainWithSoftDeletion[TDB, TDomain any](
	dbRecord TDB,
	helpers *ConversionHelpers,
	baseConverter func(TDB, *ConversionHelpers) TDomain,
	getSoftDeletion func(TDB) SoftDeletionFields,
	applySoftDeletion func(TDomain, ConvertedSoftDeletionFields) TDomain,
) TDomain {
	// Convert the base entity
	domain := baseConverter(dbRecord, helpers)

	// Apply soft deletion if present
	softDeletion := getSoftDeletion(dbRecord)
	convertedSoftDeletion := helpers.ConvertSoftDeletion(softDeletion)

	return applySoftDeletion(domain, convertedSoftDeletion)
}

// ConvertDBToDomain is a generic helper for simple conversions without soft deletion
func ConvertDBToDomain[TDB, TDomain any](
	dbRecord TDB,
	helpers *ConversionHelpers,
	converter func(TDB, *ConversionHelpers) TDomain,
) TDomain {
	return converter(dbRecord, helpers)
}
