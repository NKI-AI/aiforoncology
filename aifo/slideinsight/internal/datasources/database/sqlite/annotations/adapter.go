// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package annotations

import (
	"context"
	"database/sql"

	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/utils"
)

// Adapter provides a unified interface for all annotation operations
type Adapter struct {
	// Vector annotation services
	vectorCrud       *VectorCrudService
	vectorSearch     *VectorSearchService
	vectorSoftDelete *VectorSoftDeleteService

	// Raster annotation services
	rasterCrud       *RasterCrudService
	rasterSearch     *RasterSearchService
	rasterSoftDelete *RasterSoftDeleteService
}

// NewAdapter creates a new annotations adapter
func NewAdapter(db *sql.DB) *Adapter {
	return &Adapter{
		// Vector services
		vectorCrud:       NewVectorCrudService(db),
		vectorSearch:     NewVectorSearchService(db),
		vectorSoftDelete: NewVectorSoftDeleteService(db),

		// Raster services
		rasterCrud:       NewRasterCrudService(db),
		rasterSearch:     NewRasterSearchService(db),
		rasterSoftDelete: NewRasterSoftDeleteService(db),
	}
}

// Vector annotation methods

// LoadAllVectorAnnotations retrieves all vector annotations from the database
func (a *Adapter) LoadAllVectorAnnotations(ctx context.Context) ([]ports.VectorAnnotation, error) {
	return a.vectorSearch.LoadAllVectorAnnotations(ctx)
}

// GetVectorAnnotationsGeneric retrieves vector annotations with pagination and search support
func (a *Adapter) GetVectorAnnotationsGeneric(ctx context.Context, params utils.PaginationAndSearchParams) ([]ports.VectorAnnotation, domain.PaginationInfo, error) {
	return a.vectorSearch.GetVectorAnnotationsGeneric(ctx, params)
}

// CreateVectorAnnotation adds a new vector annotation to the database
func (a *Adapter) CreateVectorAnnotation(ctx context.Context, newVector ports.NewVectorAnnotation) error {
	return a.vectorCrud.CreateVectorAnnotation(ctx, newVector)
}

// GetVectorAnnotationByUID retrieves a specific vector annotation by its vector_id
func (a *Adapter) GetVectorAnnotationByUID(ctx context.Context, vectorUID string) (ports.VectorAnnotation, error) {
	return a.vectorCrud.GetVectorAnnotationByUID(ctx, vectorUID)
}

// UpdateVectorAnnotation updates an existing vector annotation
func (a *Adapter) UpdateVectorAnnotation(ctx context.Context, vectorUID string, updates ports.UpdateVectorAnnotation) error {
	return a.vectorCrud.UpdateVectorAnnotation(ctx, vectorUID, updates)
}

// SoftDeleteVectorAnnotation marks a vector annotation as deleted without removing it from the database
func (a *Adapter) SoftDeleteVectorAnnotation(ctx context.Context, vectorUID string, deletedBy int) error {
	return a.vectorSoftDelete.SoftDeleteVectorAnnotation(ctx, vectorUID, deletedBy)
}

// GetDeletedVectorAnnotations retrieves all soft-deleted vector annotations
func (a *Adapter) GetDeletedVectorAnnotations(ctx context.Context) ([]ports.VectorAnnotation, error) {
	return a.vectorSoftDelete.GetDeletedVectorAnnotations(ctx)
}

// RestoreVectorAnnotation restores a soft-deleted vector annotation
func (a *Adapter) RestoreVectorAnnotation(ctx context.Context, vectorUID string) error {
	return a.vectorSoftDelete.RestoreVectorAnnotation(ctx, vectorUID)
}

// Raster annotation methods

// LoadAllMasks retrieves all raster annotations from the database
func (a *Adapter) LoadAllMasks(ctx context.Context) ([]ports.Mask, error) {
	return a.rasterSearch.LoadAllMasks(ctx)
}

// GetRasterAnnotationsGeneric retrieves raster annotations with pagination and search support
func (a *Adapter) GetRasterAnnotationsGeneric(ctx context.Context, params utils.PaginationAndSearchParams) ([]ports.Mask, domain.PaginationInfo, error) {
	return a.rasterSearch.GetRasterAnnotationsGeneric(ctx, params)
}

// CreateMask adds a new raster annotation to the database
func (a *Adapter) CreateMask(ctx context.Context, newMask ports.NewMask) error {
	return a.rasterCrud.CreateMask(ctx, newMask)
}

// GetMaskByUID retrieves a specific raster annotation by its mask_id
func (a *Adapter) GetMaskByUID(ctx context.Context, maskUID string) (ports.Mask, error) {
	return a.rasterCrud.GetMaskByUID(ctx, maskUID)
}

// SoftDeleteMask marks a mask as deleted without removing it from the database
func (a *Adapter) SoftDeleteMask(ctx context.Context, maskUID string, deletedBy int) error {
	return a.rasterSoftDelete.SoftDeleteMask(ctx, maskUID, deletedBy)
}

// GetDeletedMasks retrieves all soft-deleted masks
func (a *Adapter) GetDeletedMasks(ctx context.Context) ([]ports.Mask, error) {
	return a.rasterSoftDelete.GetDeletedMasks(ctx)
}

// RestoreMask restores a soft-deleted mask
func (a *Adapter) RestoreMask(ctx context.Context, maskUID string) error {
	return a.rasterSoftDelete.RestoreMask(ctx, maskUID)
}
