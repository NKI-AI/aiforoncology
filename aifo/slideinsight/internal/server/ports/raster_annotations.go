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

// Mask represents a raster annotation in the database
type Mask struct {
	ID         int
	ActorType  string // 'user' or 'model'
	ActorID    int    // ID of the user or model
	CreatorID  int    // ID of the user who created this mask
	SlideID    int    // Internal slide ID (foreign key to slides.id)
	SlideUID   string // External slide UID for API responses
	MaskUID    string // External mask ID for API
	Version    int
	Name       string
	MaskURI    string  // maps to file_uri in raster_annotations
	FileHash   *string // Hash for deduplication, nullable and unique
	TilesURL   string
	Format     string // tiff or png
	MaskWidth  int
	MaskHeight int
	MaskMpp    float64
	Labels     string // JSON labels data
	Metadata   string // JSON metadata
	Mutable    bool   // Whether the annotation can be modified
	DeletedAt  *time.Time
	DeletedBy  *int
	CreatedAt  time.Time
}

// NewMask represents a new raster annotation to be created in the database
type NewMask struct {
	TenantID   int
	ActorType  string // 'user' or 'model'
	ActorID    int    // ID of the user or model
	CreatorID  int    // ID of the user who created this mask
	SlideUID   string // External slide UID, will be converted to internal slide_id
	MaskUID    string
	Version    int
	Name       string
	MaskURI    string  // maps to file_uri in raster_annotations
	FileHash   *string // Hash for deduplication, nullable and unique
	TilesURL   string
	Format     string // tiff or png
	MaskWidth  int
	MaskHeight int
	MaskMpp    float64
	Labels     string // JSON labels data
	Metadata   string // JSON metadata
	Mutable    bool   // Whether the annotation can be modified
}

// RasterAnnotationsRepository defines the interface for raster annotation-related database operations
type RasterAnnotationsRepository interface {
	// LoadAllMasks retrieves all raster annotations from the database.
	LoadAllMasks(ctx context.Context) ([]Mask, error)

	// GetRasterAnnotationsGeneric retrieves raster annotations with pagination and search support.
	GetRasterAnnotationsGeneric(ctx context.Context, params utils.PaginationAndSearchParams) ([]Mask, domain.PaginationInfo, error)

	// CreateMask adds a new raster annotation to the database.
	CreateMask(ctx context.Context, newMask NewMask) error

	// GetMaskByUID retrieves a specific raster annotation by its mask_id.
	GetMaskByUID(ctx context.Context, maskUID string) (Mask, error)

	// SoftDeleteMask marks a mask as deleted without removing it from the database.
	SoftDeleteMask(ctx context.Context, maskUID string, deletedBy int) error

	// GetDeletedMasks retrieves all soft-deleted masks.
	GetDeletedMasks(ctx context.Context) ([]Mask, error)

	// RestoreMask restores a soft-deleted mask.
	RestoreMask(ctx context.Context, maskUID string) error
}
