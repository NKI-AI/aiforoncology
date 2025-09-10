// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package ports

import (
	"context"
	"time"

	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"aifo.dev/aifo/slideinsight/internal/utils"
)

// VectorAnnotation represents a vector annotation in the database
type VectorAnnotation struct {
	ID        int
	ActorType string // 'user' or 'model'
	ActorID   int    // ID of the user or model
	CreatorID int
	SlideID   int    // Internal slide ID (foreign key to slides.id)
	SlideUID  string // External slide UID for API responses
	VectorUID string // External vector ID for API
	Version   int
	Name      string
	FileURI   string // URI to the GeoJSON file (optional)
	DataBlob  string // Inline data (JSON, GeoJSON, etc.) as alternative to FileURI
	Labels    string // JSON labels data
	Metadata  string // JSON metadata
	Mutable   bool   // Whether the annotation can be modified
	DeletedAt *time.Time
	DeletedBy *int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewVectorAnnotation represents a new vector annotation to be created in the database
type NewVectorAnnotation struct {
	TenantID  int
	ActorType string // 'user' or 'model'
	ActorID   int    // ID of the user or model
	CreatorID int
	SlideUID  string // External slide UID, will be converted to internal slide_id
	VectorUID string
	Version   int
	Name      string
	FileURI   string // URI to the GeoJSON file (optional)
	DataBlob  string // Inline data (JSON, GeoJSON, etc.) as alternative to FileURI
	Labels    string // JSON labels data
	Metadata  string // JSON metadata
	Mutable   bool   // Whether the annotation can be modified
}

// UpdateVectorAnnotation represents updates to an existing vector annotation
type UpdateVectorAnnotation struct {
	Name     *string // Optional name update
	FileURI  *string // Optional file URI update
	DataBlob *string // Optional inline data update
	Labels   *string // Optional labels update
	Metadata *string // Optional metadata update
	Mutable  *bool   // Optional mutable flag update
}

// VectorAnnotationsRepository defines the interface for vector annotation-related database operations
type VectorAnnotationsRepository interface {
	// LoadAllVectorAnnotations retrieves all vector annotations from the database.
	LoadAllVectorAnnotations(ctx context.Context) ([]VectorAnnotation, error)

	// GetVectorAnnotationsGeneric retrieves vector annotations with pagination and search support.
	GetVectorAnnotationsGeneric(ctx context.Context, params utils.PaginationAndSearchParams) ([]VectorAnnotation, domain.PaginationInfo, error)

	// CreateVectorAnnotation adds a new vector annotation to the database.
	CreateVectorAnnotation(ctx context.Context, newVector NewVectorAnnotation) error

	// GetVectorAnnotationByUID retrieves a specific vector annotation by its vector_id.
	GetVectorAnnotationByUID(ctx context.Context, vectorUID string) (VectorAnnotation, error)

	// UpdateVectorAnnotation updates an existing vector annotation.
	UpdateVectorAnnotation(ctx context.Context, vectorUID string, updates UpdateVectorAnnotation) error

	// SoftDeleteVectorAnnotation marks a vector annotation as deleted without removing it from the database.
	SoftDeleteVectorAnnotation(ctx context.Context, vectorUID string, deletedBy int) error

	// GetDeletedVectorAnnotations retrieves all soft-deleted vector annotations.
	GetDeletedVectorAnnotations(ctx context.Context) ([]VectorAnnotation, error)

	// RestoreVectorAnnotation restores a soft-deleted vector annotation.
	RestoreVectorAnnotation(ctx context.Context, vectorUID string) error
}
