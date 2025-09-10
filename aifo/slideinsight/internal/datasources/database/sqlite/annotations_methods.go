// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package sqlite

import (
	"context"

	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/utils"
)

// Annotation-related methods that delegate to the annotations adapter

// Vector annotation methods

// LoadAllVectorAnnotations retrieves all vector annotations from the database
func (db *DB) LoadAllVectorAnnotations(ctx context.Context) ([]ports.VectorAnnotation, error) {
	return db.annotations.LoadAllVectorAnnotations(ctx)
}

// GetVectorAnnotationsGeneric retrieves vector annotations with pagination and search support
func (db *DB) GetVectorAnnotationsGeneric(ctx context.Context, params utils.PaginationAndSearchParams) ([]ports.VectorAnnotation, domain.PaginationInfo, error) {
	return db.annotations.GetVectorAnnotationsGeneric(ctx, params)
}

// CreateVectorAnnotation adds a new vector annotation to the database
func (db *DB) CreateVectorAnnotation(ctx context.Context, newVector ports.NewVectorAnnotation) error {
	return db.annotations.CreateVectorAnnotation(ctx, newVector)
}

// GetVectorAnnotationByUID retrieves a specific vector annotation by its vector_id
func (db *DB) GetVectorAnnotationByUID(ctx context.Context, vectorUID string) (ports.VectorAnnotation, error) {
	return db.annotations.GetVectorAnnotationByUID(ctx, vectorUID)
}

// UpdateVectorAnnotation updates an existing vector annotation
func (db *DB) UpdateVectorAnnotation(ctx context.Context, vectorUID string, updates ports.UpdateVectorAnnotation) error {
	return db.annotations.UpdateVectorAnnotation(ctx, vectorUID, updates)
}

// SoftDeleteVectorAnnotation marks a vector annotation as deleted without removing it from the database
func (db *DB) SoftDeleteVectorAnnotation(ctx context.Context, vectorUID string, deletedBy int) error {
	return db.annotations.SoftDeleteVectorAnnotation(ctx, vectorUID, deletedBy)
}

// GetDeletedVectorAnnotations retrieves all soft-deleted vector annotations
func (db *DB) GetDeletedVectorAnnotations(ctx context.Context) ([]ports.VectorAnnotation, error) {
	return db.annotations.GetDeletedVectorAnnotations(ctx)
}

// RestoreVectorAnnotation restores a soft-deleted vector annotation
func (db *DB) RestoreVectorAnnotation(ctx context.Context, vectorUID string) error {
	return db.annotations.RestoreVectorAnnotation(ctx, vectorUID)
}

// Raster annotation methods

// LoadAllMasks retrieves all raster annotations from the database
func (db *DB) LoadAllMasks(ctx context.Context) ([]ports.Mask, error) {
	return db.annotations.LoadAllMasks(ctx)
}

// GetRasterAnnotationsGeneric retrieves raster annotations with pagination and search support
func (db *DB) GetRasterAnnotationsGeneric(ctx context.Context, params utils.PaginationAndSearchParams) ([]ports.Mask, domain.PaginationInfo, error) {
	return db.annotations.GetRasterAnnotationsGeneric(ctx, params)
}

// CreateMask adds a new raster annotation to the database
func (db *DB) CreateMask(ctx context.Context, newMask ports.NewMask) error {
	return db.annotations.CreateMask(ctx, newMask)
}

// GetMaskByUID retrieves a specific raster annotation by its mask_id
func (db *DB) GetMaskByUID(ctx context.Context, maskUID string) (ports.Mask, error) {
	return db.annotations.GetMaskByUID(ctx, maskUID)
}

// SoftDeleteMask marks a mask as deleted without removing it from the database
func (db *DB) SoftDeleteMask(ctx context.Context, maskUID string, deletedBy int) error {
	return db.annotations.SoftDeleteMask(ctx, maskUID, deletedBy)
}

// GetDeletedMasks retrieves all soft-deleted masks
func (db *DB) GetDeletedMasks(ctx context.Context) ([]ports.Mask, error) {
	return db.annotations.GetDeletedMasks(ctx)
}

// RestoreMask restores a soft-deleted mask
func (db *DB) RestoreMask(ctx context.Context, maskUID string) error {
	return db.annotations.RestoreMask(ctx, maskUID)
}
