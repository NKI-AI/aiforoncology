// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package cases

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"github.com/gofiber/fiber/v2/log"
)

// CaseService handles basic CRUD operations for cases
type CaseService struct {
	db *sql.DB
}

// NewCaseService creates a new case service instance
func NewCaseService(db *sql.DB) *CaseService {
	return &CaseService{db: db}
}

// CreateCase adds a new case to the database
func (s *CaseService) CreateCase(ctx context.Context, newCase ports.NewCase) error {
	_, err := s.db.Exec("INSERT INTO cases (tenant_id, short_uid, creator_id, name, metadata) VALUES (?, ?, ?, ?, ?)",
		newCase.TenantID, newCase.CaseUID, newCase.CreatorID, newCase.Name, newCase.Metadata)
	if err != nil {
		return fmt.Errorf("failed to insert case: %w", err)
	}
	return nil
}

// GetCaseByUID retrieves a specific case by its ID
func (s *CaseService) GetCaseByUID(ctx context.Context, caseUID string) (ports.Case, error) {
	var case_ ports.Case
	var createdAtStr string
	var updatedAtStr string
	var deletedAtStr sql.NullString
	var deletedBy sql.NullInt64

	query := `
		SELECT c.id, c.tenant_id, COALESCE(t.short_uid, '') as tenant_uid, c.short_uid as case_uid, c.creator_id, COALESCE(u.short_uid, '') as creator_uid, c.name, c.metadata, c.deleted_at, c.deleted_by, c.created_at, c.updated_at 
		FROM cases c
		LEFT JOIN tenants t ON c.tenant_id = t.id
		LEFT JOIN users u ON c.creator_id = u.id
		WHERE c.short_uid = ? AND c.deleted_at IS NULL`

	err := s.db.QueryRow(query, caseUID).Scan(
		&case_.ID, &case_.TenantID, &case_.TenantUID, &case_.CaseUID, &case_.CreatorID, &case_.CreatorUID, &case_.Name, &case_.Metadata, &deletedAtStr, &deletedBy, &createdAtStr, &updatedAtStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return ports.Case{}, fmt.Errorf("case with UID '%s' not found", caseUID)
		}
		return ports.Case{}, fmt.Errorf("failed to get case: %w", err)
	}

	// Parse the created_at timestamp
	if createdAtStr != "" {
		createdAt, err := time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			log.Error("failed to parse created_at timestamp", "error", err)
		}
		case_.CreatedAt = createdAt
	}

	// Parse the updated_at timestamp
	if updatedAtStr != "" {
		updatedAt, err := time.Parse(time.RFC3339, updatedAtStr)
		if err != nil {
			log.Error("failed to parse updated_at timestamp", "error", err)
		}
		case_.UpdatedAt = updatedAt
	}

	// Handle soft deletion fields
	if deletedAtStr.Valid {
		deletedAt, err := time.Parse(time.RFC3339, deletedAtStr.String)
		if err != nil {
			log.Error("failed to parse deleted_at timestamp", "error", err)
		} else {
			case_.DeletedAt = &deletedAt
		}
	}

	if deletedBy.Valid {
		deletedByInt := int(deletedBy.Int64)
		case_.DeletedBy = &deletedByInt
	}

	return case_, nil
}

// GetCasesCount returns the total count of cases in the database
func (s *CaseService) GetCasesCount(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM cases WHERE deleted_at IS NULL").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count cases: %w", err)
	}
	return count, nil
}

// SoftDeleteCase marks a case as deleted without removing it from the database
func (s *CaseService) SoftDeleteCase(ctx context.Context, caseUID string, deletedBy int) error {
	result, err := s.db.Exec("UPDATE cases SET deleted_at = CURRENT_TIMESTAMP, deleted_by = ? WHERE short_uid = ? AND deleted_at IS NULL", deletedBy, caseUID)
	if err != nil {
		return fmt.Errorf("failed to soft delete case: %w", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("case with UID '%s' not found or already deleted", caseUID)
	}

	return nil
}

// GetDeletedCases retrieves all soft-deleted cases
func (s *CaseService) GetDeletedCases(ctx context.Context) ([]ports.Case, error) {
	query := `
		SELECT c.id, c.tenant_id, COALESCE(t.short_uid, '') as tenant_uid, c.short_uid as case_uid, c.creator_id, COALESCE(u.short_uid, '') as creator_uid, c.name, c.metadata, c.deleted_at, c.deleted_by, c.created_at, c.updated_at 
		FROM cases c
		LEFT JOIN tenants t ON c.tenant_id = t.id
		LEFT JOIN users u ON c.creator_id = u.id
		WHERE c.deleted_at IS NOT NULL`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query deleted cases: %w", err)
	}
	defer rows.Close()

	var cases []ports.Case
	for rows.Next() {
		var case_ ports.Case
		var createdAtStr string
		var updatedAtStr string
		var deletedAtStr sql.NullString
		var deletedBy sql.NullInt64
		if err := rows.Scan(
			&case_.ID,
			&case_.TenantID,
			&case_.TenantUID,
			&case_.CaseUID,
			&case_.CreatorID,
			&case_.CreatorUID,
			&case_.Name,
			&case_.Metadata,
			&deletedAtStr,
			&deletedBy,
			&createdAtStr,
			&updatedAtStr); err != nil {
			return nil, fmt.Errorf("failed to scan case row: %w", err)
		}

		// Parse the created_at timestamp
		if createdAtStr != "" {
			createdAt, err := time.Parse(time.RFC3339, createdAtStr)
			if err != nil {
				log.Error("failed to parse created_at timestamp", "error", err)
			} else {
				case_.CreatedAt = createdAt
			}
		}

		// Parse the updated_at timestamp
		if updatedAtStr != "" {
			updatedAt, err := time.Parse(time.RFC3339, updatedAtStr)
			if err != nil {
				log.Error("failed to parse updated_at timestamp", "error", err)
			} else {
				case_.UpdatedAt = updatedAt
			}
		}

		// Handle soft deletion fields
		if deletedAtStr.Valid {
			deletedAt, err := time.Parse(time.RFC3339, deletedAtStr.String)
			if err != nil {
				log.Error("failed to parse deleted_at timestamp", "error", err)
			} else {
				case_.DeletedAt = &deletedAt
			}
		}

		if deletedBy.Valid {
			deletedByInt := int(deletedBy.Int64)
			case_.DeletedBy = &deletedByInt
		}

		cases = append(cases, case_)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over deleted case rows: %w", err)
	}

	return cases, nil
}

// RestoreCase restores a soft-deleted case
func (s *CaseService) RestoreCase(ctx context.Context, caseUID string) error {
	result, err := s.db.Exec("UPDATE cases SET deleted_at = NULL, deleted_by = NULL WHERE short_uid = ? AND deleted_at IS NOT NULL", caseUID)
	if err != nil {
		return fmt.Errorf("failed to restore case: %w", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("case with UID '%s' not found or not deleted", caseUID)
	}

	return nil
}
