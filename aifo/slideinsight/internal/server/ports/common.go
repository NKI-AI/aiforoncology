// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package ports

import (
	"context"
	"time"
)

// SoftDeletable represents an entity that supports soft deletion
type SoftDeletable interface {
	GetDeletedAt() *time.Time
	GetDeletedBy() *int
	SetDeletedAt(deletedAt *time.Time)
	SetDeletedBy(deletedBy *int)
}

// SoftDeletionRepository defines common soft deletion operations
type SoftDeletionRepository[T any] interface {
	// SoftDelete marks an entity as deleted
	SoftDelete(ctx context.Context, id string, deletedBy int) error

	// GetDeleted retrieves all soft-deleted entities
	GetDeleted(ctx context.Context) ([]T, error)

	// Restore restores a soft-deleted entity
	Restore(ctx context.Context, id string) error
}
