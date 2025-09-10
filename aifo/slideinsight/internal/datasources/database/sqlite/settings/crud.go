// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package settings

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"aifo.dev/aifo/slideinsight/internal/server/errors"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
)

// CrudService provides CRUD operations for settings
type CrudService struct {
	db *sql.DB
}

// NewCrudService creates a new CRUD service
func NewCrudService(db *sql.DB) *CrudService {
	return &CrudService{db: db}
}

// GetSetting retrieves a specific setting by key and tenant
func (s *CrudService) GetSetting(ctx context.Context, tenantID int, key string) (*ports.Setting, error) {
	query := `
		SELECT id, tenant_id, key, value_type, value, created_at, updated_at
		FROM settings 
		WHERE tenant_id = ? AND key = ?
	`

	row := s.db.QueryRowContext(ctx, query, tenantID, key)

	var setting ports.Setting
	var createdAt, updatedAt string

	err := row.Scan(&setting.ID, &setting.TenantID, &setting.Key, &setting.ValueType,
		&setting.Value, &createdAt, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewSettingNotFoundError(tenantID, key)
		}
		return nil, fmt.Errorf("failed to get setting: %w", err)
	}

	if setting.CreatedAt, err = parseTimestamp(createdAt); err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}
	if setting.UpdatedAt, err = parseTimestamp(updatedAt); err != nil {
		return nil, fmt.Errorf("failed to parse updated_at: %w", err)
	}

	return &setting, nil
}

// CreateSetting creates a new setting
func (s *CrudService) CreateSetting(ctx context.Context, newSetting ports.NewSetting) (*ports.Setting, error) {
	// Validate value type
	if !isValidValueType(newSetting.ValueType) {
		return nil, fmt.Errorf("invalid value type: %s", newSetting.ValueType)
	}

	query := `
		INSERT INTO settings (tenant_id, key, value_type, value)
		VALUES (?, ?, ?, ?)
	`

	result, err := s.db.ExecContext(ctx, query,
		newSetting.TenantID, newSetting.Key, newSetting.ValueType, newSetting.Value)
	if err != nil {
		// Check for unique constraint violation
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return nil, errors.NewSettingAlreadyExistsError(newSetting.TenantID, newSetting.Key)
		}
		return nil, fmt.Errorf("failed to create setting: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get inserted setting ID: %w", err)
	}

	// Return the created setting by retrieving it
	return s.getSettingByID(ctx, int(id))
}

// UpdateSetting updates an existing setting
func (s *CrudService) UpdateSetting(ctx context.Context, tenantID int, key string, updates ports.SettingUpdates) (*ports.Setting, error) {
	setParts := []string{}
	args := []interface{}{}

	if updates.ValueType != nil {
		if !isValidValueType(*updates.ValueType) {
			return nil, fmt.Errorf("invalid value type: %s", *updates.ValueType)
		}
		setParts = append(setParts, "value_type = ?")
		args = append(args, *updates.ValueType)
	}
	if updates.Value != nil {
		setParts = append(setParts, "value = ?")
		args = append(args, *updates.Value)
	}

	if len(setParts) == 0 {
		return nil, nil // No updates to apply
	}

	query := fmt.Sprintf("UPDATE settings SET %s WHERE tenant_id = ? AND key = ?", strings.Join(setParts, ", "))
	args = append(args, tenantID, key)

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update setting: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return nil, errors.NewSettingNotFoundError(tenantID, key)
	}

	// Return the updated setting
	return s.GetSetting(ctx, tenantID, key)
}

// DeleteSetting deletes a setting
func (s *CrudService) DeleteSetting(ctx context.Context, tenantID int, key string) error {
	query := "DELETE FROM settings WHERE tenant_id = ? AND key = ?"

	result, err := s.db.ExecContext(ctx, query, tenantID, key)
	if err != nil {
		return fmt.Errorf("failed to delete setting: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.NewSettingNotFoundError(tenantID, key)
	}

	return nil
}

// SettingExists checks if a setting exists
func (s *CrudService) SettingExists(ctx context.Context, tenantID int, key string) (bool, error) {
	query := "SELECT 1 FROM settings WHERE tenant_id = ? AND key = ? LIMIT 1"

	var exists int
	err := s.db.QueryRowContext(ctx, query, tenantID, key).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check setting existence: %w", err)
	}

	return true, nil
}

// getSettingByID retrieves a setting by its ID (internal helper)
func (s *CrudService) getSettingByID(ctx context.Context, id int) (*ports.Setting, error) {
	query := `
		SELECT id, tenant_id, key, value_type, value, created_at, updated_at
		FROM settings 
		WHERE id = ?
	`

	row := s.db.QueryRowContext(ctx, query, id)

	var setting ports.Setting
	var createdAt, updatedAt string

	err := row.Scan(&setting.ID, &setting.TenantID, &setting.Key, &setting.ValueType,
		&setting.Value, &createdAt, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("setting with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get setting by ID: %w", err)
	}

	if setting.CreatedAt, err = parseTimestamp(createdAt); err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}
	if setting.UpdatedAt, err = parseTimestamp(updatedAt); err != nil {
		return nil, fmt.Errorf("failed to parse updated_at: %w", err)
	}

	return &setting, nil
}

// isValidValueType checks if the provided value type is valid
func isValidValueType(valueType ports.SettingValueType) bool {
	switch valueType {
	case ports.SettingValueTypeBoolean, ports.SettingValueTypeNumber,
		ports.SettingValueTypeString, ports.SettingValueTypeJSON:
		return true
	default:
		return false
	}
}

// parseTimestamp flexibly parses timestamps in multiple formats
func parseTimestamp(timestamp string) (time.Time, error) {
	if timestamp == "" {
		return time.Time{}, nil
	}

	// Try RFC3339 format first (2006-01-02T15:04:05Z)
	if t, err := time.Parse(time.RFC3339, timestamp); err == nil {
		return t, nil
	}

	// Try SQLite datetime format (2006-01-02 15:04:05)
	if t, err := time.Parse("2006-01-02 15:04:05", timestamp); err == nil {
		return t, nil
	}

	// Try ISO8601 with timezone offset (2006-01-02T15:04:05Z07:00)
	if t, err := time.Parse(time.RFC3339Nano, timestamp); err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("unable to parse timestamp '%s'", timestamp)
}
