// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package slides

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"aifo.dev/aifo/slideinsight/internal/server/errors"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/utils"
	"github.com/gofiber/fiber/v2/log"
)

// Adapter provides a unified interface for all slide operations
type Adapter struct {
	db *sql.DB
}

// NewAdapter creates a new slides adapter
func NewAdapter(db *sql.DB) *Adapter {
	return &Adapter{db: db}
}

// buildSlidesSearchWhereClause builds WHERE conditions and arguments for slides search queries
func (a *Adapter) buildSlidesSearchWhereClause(search utils.SearchParams, searchableFields []string) ([]string, []interface{}) {
	var whereConditions []string
	var args []interface{}

	// General search across specified searchable fields
	if search.Query != "" {
		// Build LIKE conditions for each searchable field
		var fieldConditions []string
		for _, field := range searchableFields {
			switch field {
			case "slide_name":
				fieldConditions = append(fieldConditions, "s.slide_name LIKE ?")
				args = append(args, "%"+search.Query+"%")
			case "slide_uri":
				fieldConditions = append(fieldConditions, "s.slide_uri LIKE ?")
				args = append(args, "%"+search.Query+"%")
			}
		}
		if len(fieldConditions) > 0 {
			whereConditions = append(whereConditions, "("+strings.Join(fieldConditions, " OR ")+")")
		}
	}

	// Add specific field filters
	for field, value := range search.Filters {
		switch field {
		case "slide_name":
			whereConditions = append(whereConditions, "s.slide_name LIKE ?")
			args = append(args, "%"+value+"%")
		case "has_vector_annotations":
			if value == "true" {
				whereConditions = append(whereConditions, "va.id IS NOT NULL")
			} else if value == "false" {
				whereConditions = append(whereConditions, "va.id IS NULL")
			}
		case "has_raster_annotations":
			if value == "true" {
				whereConditions = append(whereConditions, "ra.id IS NOT NULL")
			} else if value == "false" {
				whereConditions = append(whereConditions, "ra.id IS NULL")
			}
		default:
			// Skip unknown filter fields
			continue
		}
	}

	return whereConditions, args
}

// validateSortDir ensures sort direction is valid to prevent SQL injection
func (a *Adapter) validateSortDir(sortDir string) string {
	dir := strings.ToUpper(sortDir)
	if dir != "ASC" && dir != "DESC" {
		return "ASC"
	}
	return dir
}

// LoadAllSlides retrieves slides from the database with search/filter and pagination support
func (a *Adapter) LoadAllSlides(ctx context.Context, search utils.SearchParams, pagination utils.PaginationParams) ([]ports.Slide, error) {
	baseQuery := `
		SELECT DISTINCT s.id, s.case_id, COALESCE(c.short_uid, '') as case_short_uid, s.slide_uid, s.slide_name, s.slide_uri, s.slide_width, s.slide_height, s.slide_mpp, s.image_type_id, COALESCE(s.metadata, '') as metadata, s.creator_id, COALESCE(u.short_uid, '') as creator_uid, s.deleted_at, s.deleted_by, s.created_at, s.updated_at
		FROM slides s
		LEFT JOIN cases c ON s.case_id = c.id
		LEFT JOIN users u ON s.creator_id = u.id
		LEFT JOIN vector_annotations va ON s.id = va.slide_id AND va.deleted_at IS NULL
		LEFT JOIN raster_annotations ra ON s.id = ra.slide_id AND ra.deleted_at IS NULL
		WHERE s.deleted_at IS NULL`

	// Build WHERE clause based on search parameters
	searchableFields := []string{"slide_name", "slide_uri"} // Define which fields are searchable
	whereConditions, args := a.buildSlidesSearchWhereClause(search, searchableFields)

	// Add WHERE conditions to query
	if len(whereConditions) > 0 {
		baseQuery += " AND " + strings.Join(whereConditions, " AND ")
	}

	// Add ordering with safe sort direction validation
	orderBy := "s.created_at DESC" // Default ordering (with table alias)
	if search.HasSort() {
		safeSortDir := a.validateSortDir(search.SortDir)
		switch search.SortBy {
		case "slide_name", "name":
			orderBy = "s.slide_name " + safeSortDir
		case "created_at", "createdAt":
			orderBy = "s.created_at " + safeSortDir
		case "updated_at", "updatedAt":
			orderBy = "s.updated_at " + safeSortDir
		case "slide_id", "slideId":
			orderBy = "s.slide_uid " + safeSortDir
		default:
			// Keep default ordering for unknown sort fields
		}
	}
	baseQuery += " ORDER BY " + orderBy

	// Add pagination
	if pagination.Limit > 0 {
		offset := pagination.CalculateOffset()
		baseQuery += " LIMIT ? OFFSET ?"
		args = append(args, pagination.Limit, offset)
	}

	rows, err := a.db.Query(baseQuery, args...)
	if err != nil {
		return nil, errors.NewDatabaseQueryError("slides", err)
	}
	defer rows.Close()

	var slides []ports.Slide
	for rows.Next() {
		var slide ports.Slide
		var createdAtStr, updatedAtStr string
		var metadataStr sql.NullString
		var deletedAtStr sql.NullString
		var deletedBy sql.NullInt64
		var creatorID sql.NullInt64
		if err := rows.Scan(&slide.ID, &slide.CaseID, &slide.CaseUID, &slide.SlideUID, &slide.SlideName, &slide.SlideURI, &slide.SlideWidth, &slide.SlideHeight, &slide.SlideMpp, &slide.ImageTypeID, &metadataStr, &creatorID, &slide.CreatorUID, &deletedAtStr, &deletedBy, &createdAtStr, &updatedAtStr); err != nil {
			return nil, errors.NewDatabaseScanError("slide", err)
		}

		// Convert metadata string to json.RawMessage
		if metadataStr.Valid && metadataStr.String != "" {
			slide.Metadata = json.RawMessage(metadataStr.String)
		}

		// Handle creator ID
		if creatorID.Valid {
			slide.CreatorID = int(creatorID.Int64)
		}

		// Parse the created_at timestamp
		if createdAtStr != "" {
			createdAt, err := time.Parse(time.RFC3339, createdAtStr)
			if err != nil {
				log.Error("failed to parse created_at timestamp", "error", err)
			} else {
				slide.CreatedAt = createdAt
			}
		}

		// Parse the updated_at timestamp
		if updatedAtStr != "" {
			updatedAt, err := time.Parse(time.RFC3339, updatedAtStr)
			if err != nil {
				log.Error("failed to parse updated_at timestamp", "error", err)
			} else {
				slide.UpdatedAt = updatedAt
			}
		}

		// Handle soft deletion fields
		if deletedAtStr.Valid {
			deletedAt, err := time.Parse(time.RFC3339, deletedAtStr.String)
			if err != nil {
				log.Error("failed to parse deleted_at timestamp", "error", err)
			} else {
				slide.DeletedAt = &deletedAt
			}
		}

		if deletedBy.Valid {
			deletedByInt := int(deletedBy.Int64)
			slide.DeletedBy = &deletedByInt
		}

		slides = append(slides, slide)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.NewDatabaseIterateRowsError("slides", err)
	}

	return slides, nil
}

// GetSlidesCount returns the total count of slides matching search criteria
func (a *Adapter) GetSlidesCount(ctx context.Context, search utils.SearchParams) (int, error) {
	baseQuery := `
		SELECT COUNT(DISTINCT s.id) 
		FROM slides s
		LEFT JOIN cases c ON s.case_id = c.id
		LEFT JOIN vector_annotations va ON s.id = va.slide_id AND va.deleted_at IS NULL
		LEFT JOIN raster_annotations ra ON s.id = ra.slide_id AND ra.deleted_at IS NULL
		WHERE s.deleted_at IS NULL`

	// Build WHERE clause based on search parameters (same logic as LoadAllSlides)
	searchableFields := []string{"slide_name", "slide_uri"} // Define which fields are searchable
	whereConditions, args := a.buildSlidesSearchWhereClause(search, searchableFields)

	// Add WHERE conditions to query
	if len(whereConditions) > 0 {
		baseQuery += " AND " + strings.Join(whereConditions, " AND ")
	}

	var count int
	err := a.db.QueryRow(baseQuery, args...).Scan(&count)
	if err != nil {
		return 0, errors.NewDatabaseQueryError("slide count", err)
	}
	return count, nil
}

func (a *Adapter) GetSlidesByCaseUID(ctx context.Context, caseUID string) ([]ports.Slide, error) {
	// First, get the internal case ID from the case's short_uid
	var internalCaseID int
	err := a.db.QueryRow("SELECT id FROM cases WHERE short_uid = ? AND deleted_at IS NULL", caseUID).Scan(&internalCaseID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewCaseNotFoundError(caseUID)
		}
		return nil, errors.NewDatabaseQueryError("case ID lookup", err)
	}

	// Now query slides using the internal case ID, with join to get case short_uid and creator info
	query := `
		SELECT s.id, s.case_id, COALESCE(c.short_uid, '') as case_short_uid, s.slide_uid, s.slide_name, s.slide_uri, s.slide_width, s.slide_height, s.slide_mpp, s.image_type_id, s.creator_id, COALESCE(u.short_uid, '') as creator_uid, s.deleted_at, s.deleted_by, s.created_at, s.updated_at
		FROM slides s
		LEFT JOIN cases c ON s.case_id = c.id
		LEFT JOIN users u ON s.creator_id = u.id
		WHERE s.case_id = ? AND s.deleted_at IS NULL`

	rows, err := a.db.Query(query, internalCaseID)
	if err != nil {
		return nil, errors.NewDatabaseQueryError("slides by case", err)
	}
	defer rows.Close()

	var slides []ports.Slide
	for rows.Next() {
		var slide ports.Slide
		var createdAtStr, updatedAtStr string
		var deletedAtStr sql.NullString
		var deletedBy sql.NullInt64
		var creatorID sql.NullInt64
		if err := rows.Scan(&slide.ID, &slide.CaseID, &slide.CaseUID, &slide.SlideUID, &slide.SlideName, &slide.SlideURI, &slide.SlideWidth, &slide.SlideHeight, &slide.SlideMpp, &slide.ImageTypeID, &creatorID, &slide.CreatorUID, &deletedAtStr, &deletedBy, &createdAtStr, &updatedAtStr); err != nil {
			return nil, errors.NewDatabaseScanError("slide", err)
		}

		// Handle creator ID
		if creatorID.Valid {
			slide.CreatorID = int(creatorID.Int64)
		}

		// Parse the created_at timestamp
		if createdAtStr != "" {
			createdAt, err := time.Parse(time.RFC3339, createdAtStr)
			if err != nil {
				log.Error("failed to parse created_at timestamp", "error", err)
			} else {
				slide.CreatedAt = createdAt
			}
		}

		// Parse the updated_at timestamp
		if updatedAtStr != "" {
			updatedAt, err := time.Parse(time.RFC3339, updatedAtStr)
			if err != nil {
				log.Error("failed to parse updated_at timestamp", "error", err)
			} else {
				slide.UpdatedAt = updatedAt
			}
		}

		// Handle soft deletion fields
		if deletedAtStr.Valid {
			deletedAt, err := time.Parse(time.RFC3339, deletedAtStr.String)
			if err != nil {
				log.Error("failed to parse deleted_at timestamp", "error", err)
			} else {
				slide.DeletedAt = &deletedAt
			}
		}

		if deletedBy.Valid {
			deletedByInt := int(deletedBy.Int64)
			slide.DeletedBy = &deletedByInt
		}

		slides = append(slides, slide)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.NewDatabaseIterateRowsError("slides", err)
	}

	return slides, nil
}

// CreateSlide adds a new slide to the database
func (a *Adapter) CreateSlide(ctx context.Context, newSlide ports.NewSlide) error {
	// Convert json.RawMessage to string for database storage
	var metadataStr string
	if len(newSlide.Metadata) > 0 {
		metadataStr = string(newSlide.Metadata)
	}

	_, err := a.db.Exec("INSERT INTO slides (case_id, slide_uid, slide_name, slide_uri, slide_width, slide_height, slide_mpp, image_type_id, metadata, creator_id, tenant_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		newSlide.CaseID, newSlide.SlideID, newSlide.SlideName, newSlide.SlideURI, newSlide.SlideWidth, newSlide.SlideHeight, newSlide.SlideMpp, newSlide.ImageTypeID, metadataStr, newSlide.CreatorID, newSlide.TenantID)
	if err != nil {
		return errors.NewDatabaseInsertError("slide", err)
	}
	return nil
}

// SlideExists checks if a slide with the given ID already exists
func (a *Adapter) SlideExists(ctx context.Context, slideUID string) (bool, error) {
	var exists bool
	err := a.db.QueryRow("SELECT EXISTS(SELECT 1 FROM slides WHERE slide_uid = ? AND deleted_at IS NULL)", slideUID).Scan(&exists)
	if err != nil {
		return false, errors.NewDatabaseQueryError("slide existence check", err)
	}
	return exists, nil
}

// GetSlideByUID retrieves a specific slide by its slide_uid
func (a *Adapter) GetSlideByUID(ctx context.Context, slideUID string) (ports.Slide, error) {
	var slide ports.Slide
	var createdAtStr, updatedAtStr string
	var metadataStr sql.NullString
	var deletedAtStr sql.NullString
	var deletedBy sql.NullInt64
	var creatorID sql.NullInt64

	// Join with cases table to get the case's short_uid and users table to get creator info
	query := `
		SELECT s.id, s.case_id, COALESCE(c.short_uid, '') as case_short_uid, s.slide_uid, s.slide_name, s.slide_uri, s.slide_width, s.slide_height, s.slide_mpp, s.image_type_id, COALESCE(s.metadata, '') as metadata, s.creator_id, COALESCE(u.short_uid, '') as creator_uid, s.deleted_at, s.deleted_by, s.created_at, s.updated_at
		FROM slides s
		LEFT JOIN cases c ON s.case_id = c.id
		LEFT JOIN users u ON s.creator_id = u.id
		WHERE s.slide_uid = ? AND s.deleted_at IS NULL`

	err := a.db.QueryRow(query, slideUID).Scan(
		&slide.ID, &slide.CaseID, &slide.CaseUID, &slide.SlideUID, &slide.SlideName, &slide.SlideURI, &slide.SlideWidth, &slide.SlideHeight, &slide.SlideMpp, &slide.ImageTypeID, &metadataStr, &creatorID, &slide.CreatorUID, &deletedAtStr, &deletedBy, &createdAtStr, &updatedAtStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return ports.Slide{}, errors.NewSlideNotFoundError(slideUID)
		}
		return ports.Slide{}, errors.NewDatabaseQueryError("slide", err)
	}

	// Convert metadata string to json.RawMessage
	if metadataStr.Valid && metadataStr.String != "" {
		slide.Metadata = json.RawMessage(metadataStr.String)
	}

	// Handle creator ID
	if creatorID.Valid {
		slide.CreatorID = int(creatorID.Int64)
	}

	// Parse the created_at timestamp
	if createdAtStr != "" {
		createdAt, err := time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			log.Error("failed to parse created_at timestamp", "error", err)
		} else {
			slide.CreatedAt = createdAt
		}
	}

	// Parse the updated_at timestamp
	if updatedAtStr != "" {
		updatedAt, err := time.Parse(time.RFC3339, updatedAtStr)
		if err != nil {
			log.Error("failed to parse updated_at timestamp", "error", err)
		} else {
			slide.UpdatedAt = updatedAt
		}
	}

	// Handle soft deletion fields
	if deletedAtStr.Valid {
		deletedAt, err := time.Parse(time.RFC3339, deletedAtStr.String)
		if err != nil {
			log.Error("failed to parse deleted_at timestamp", "error", err)
		} else {
			slide.DeletedAt = &deletedAt
		}
	}

	if deletedBy.Valid {
		deletedByInt := int(deletedBy.Int64)
		slide.DeletedBy = &deletedByInt
	}

	return slide, nil
}

// SoftDeleteSlide marks a slide as deleted without removing it from the database
func (a *Adapter) SoftDeleteSlide(ctx context.Context, slideUID string, deletedBy int) error {
	result, err := a.db.Exec("UPDATE slides SET deleted_at = CURRENT_TIMESTAMP, deleted_by = ? WHERE slide_uid = ? AND deleted_at IS NULL", deletedBy, slideUID)
	if err != nil {
		return errors.NewDatabaseUpdateError("slide soft delete", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseCheckRowsError(err)
	}

	if rowsAffected == 0 {
		return errors.NewSlideAlreadyDeletedError(slideUID)
	}

	return nil
}

// GetDeletedSlides retrieves all soft-deleted slides
func (a *Adapter) GetDeletedSlides(ctx context.Context) ([]ports.Slide, error) {
	rows, err := a.db.Query("SELECT id, case_id, slide_uid, slide_name, slide_uri, slide_width, slide_height, slide_mpp, deleted_at, deleted_by, created_at FROM slides WHERE deleted_at IS NOT NULL")
	if err != nil {
		return nil, errors.NewDatabaseQueryError("deleted slides", err)
	}
	defer rows.Close()

	var slides []ports.Slide
	for rows.Next() {
		var slide ports.Slide
		var createdAtStr string
		var deletedAtStr sql.NullString
		var deletedBy sql.NullInt64
		if err := rows.Scan(&slide.ID, &slide.CaseID, &slide.SlideUID, &slide.SlideName, &slide.SlideURI, &slide.SlideWidth, &slide.SlideHeight, &slide.SlideMpp, &deletedAtStr, &deletedBy, &createdAtStr); err != nil {
			return nil, errors.NewDatabaseScanError("deleted slide", err)
		}

		// Parse the created_at timestamp
		if createdAtStr != "" {
			createdAt, err := time.Parse(time.RFC3339, createdAtStr)
			if err != nil {
				log.Error("failed to parse created_at timestamp", "error", err)
			} else {
				slide.CreatedAt = createdAt
			}
		}

		// Handle soft deletion fields
		if deletedAtStr.Valid {
			deletedAt, err := time.Parse(time.RFC3339, deletedAtStr.String)
			if err != nil {
				log.Error("failed to parse deleted_at timestamp", "error", err)
			} else {
				slide.DeletedAt = &deletedAt
			}
		}

		if deletedBy.Valid {
			deletedByInt := int(deletedBy.Int64)
			slide.DeletedBy = &deletedByInt
		}

		slides = append(slides, slide)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.NewDatabaseIterateRowsError("deleted slides", err)
	}

	return slides, nil
}

// RestoreSlide restores a soft-deleted slide
func (a *Adapter) RestoreSlide(ctx context.Context, slideUID string) error {
	result, err := a.db.Exec("UPDATE slides SET deleted_at = NULL, deleted_by = NULL WHERE slide_uid = ? AND deleted_at IS NOT NULL", slideUID)
	if err != nil {
		return errors.NewDatabaseUpdateError("slide restore", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseCheckRowsError(err)
	}

	if rowsAffected == 0 {
		return errors.NewSlideNotDeletedError(slideUID)
	}

	return nil
}
