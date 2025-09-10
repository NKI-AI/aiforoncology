// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package services

import (
	"context"
	"fmt"

	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"aifo.dev/aifo/slideinsight/internal/server/errors"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/utils"
)

// SettingsService provides business logic for settings management
type SettingsService struct {
	db ports.Database
}

// NewSettingsService creates a new settings service
func NewSettingsService(db ports.Database) *SettingsService {
	return &SettingsService{db: db}
}

// GetSettingsGeneric retrieves settings with pagination and search support
func (s *SettingsService) GetSettingsGeneric(ctx context.Context, params utils.PaginationAndSearchParams) ([]domain.Setting, domain.PaginationInfo, error) {
	// Convert to internal search params
	search := utils.SearchParams{
		Query:   params.Query,
		SortBy:  params.SortBy,
		SortDir: params.SortDir,
		Filters: params.Filters,
	}

	// Convert to internal pagination params
	pagination := utils.PaginationParams{
		Page:  params.Page,
		Limit: params.Limit,
	}

	// Check if filtering by tenant
	tenantIDFilter, hasTenantFilter := search.Filters["tenant_id"]
	var settings []ports.Setting
	var totalCount int
	var err error

	if hasTenantFilter && tenantIDFilter != "" {
		// Parse tenant ID
		var tenantID int
		if _, err := fmt.Sscanf(tenantIDFilter, "%d", &tenantID); err != nil {
			return nil, domain.PaginationInfo{}, errors.WithDetails(errors.ErrInvalidInput, "Invalid tenant ID filter")
		}

		// Get settings for specific tenant
		settings, err = s.db.GetSettings(ctx, tenantID, search, pagination)
		if err != nil {
			return nil, domain.PaginationInfo{}, fmt.Errorf("failed to get settings for tenant %d: %w", tenantID, err)
		}

		// Get count
		if search.Query != "" || len(search.Filters) > 1 { // More than just tenant_id filter
			totalCount, err = s.db.GetSettingsCountWithSearch(ctx, tenantID, search)
		} else {
			totalCount, err = s.db.GetSettingsCount(ctx, tenantID)
		}
	} else {
		// Get settings from all tenants
		settings, err = s.db.GetAllSettings(ctx, search, pagination)
		if err != nil {
			return nil, domain.PaginationInfo{}, fmt.Errorf("failed to get all settings: %w", err)
		}

		// Get count
		if search.Query != "" || len(search.Filters) > 0 {
			totalCount, err = s.db.GetAllSettingsCountWithSearch(ctx, search)
		} else {
			totalCount, err = s.db.GetAllSettingsCount(ctx)
		}
	}

	if err != nil {
		return nil, domain.PaginationInfo{}, fmt.Errorf("failed to get settings count: %w", err)
	}

	// Convert ports.Setting to domain.Setting
	domainSettings := make([]domain.Setting, len(settings))
	for i, setting := range settings {
		domainSettings[i] = domain.Setting{
			ID:        setting.ID,
			TenantID:  setting.TenantID,
			Key:       setting.Key,
			ValueType: string(setting.ValueType),
			Value:     setting.Value,
			CreatedAt: setting.CreatedAt,
			UpdatedAt: setting.UpdatedAt,
		}
	}

	// Calculate pagination info
	paginationInfo := domain.PaginationInfo{
		Page:       pagination.Page,
		Limit:      pagination.Limit,
		Total:      totalCount,
		TotalPages: (totalCount + pagination.Limit - 1) / pagination.Limit,
		HasNext:    pagination.Page < (totalCount+pagination.Limit-1)/pagination.Limit,
		HasPrev:    pagination.Page > 1,
	}

	return domainSettings, paginationInfo, nil
}

// GetSetting retrieves a specific setting by tenant ID and key
func (s *SettingsService) GetSetting(ctx context.Context, tenantID int, key string) (domain.Setting, error) {
	setting, err := s.db.GetSetting(ctx, tenantID, key)
	if err != nil {
		return domain.Setting{}, fmt.Errorf("failed to get setting: %w", err)
	}

	return domain.Setting{
		ID:        setting.ID,
		TenantID:  setting.TenantID,
		Key:       setting.Key,
		ValueType: string(setting.ValueType),
		Value:     setting.Value,
		CreatedAt: setting.CreatedAt,
		UpdatedAt: setting.UpdatedAt,
	}, nil
}

// CreateSetting creates a new setting
func (s *SettingsService) CreateSetting(ctx context.Context, newSetting domain.NewSetting) (domain.Setting, error) {
	// Validate value type
	if !isValidValueType(newSetting.ValueType) {
		return domain.Setting{}, errors.WithDetails(errors.ErrInvalidInput, "Invalid value type: %s. Must be one of: boolean, number, string, json", newSetting.ValueType)
	}

	// Convert domain to ports
	portsNewSetting := ports.NewSetting{
		TenantID:  newSetting.TenantID,
		Key:       newSetting.Key,
		ValueType: ports.SettingValueType(newSetting.ValueType),
		Value:     newSetting.Value,
	}

	setting, err := s.db.CreateSetting(ctx, portsNewSetting)
	if err != nil {
		return domain.Setting{}, fmt.Errorf("failed to create setting: %w", err)
	}

	return domain.Setting{
		ID:        setting.ID,
		TenantID:  setting.TenantID,
		Key:       setting.Key,
		ValueType: string(setting.ValueType),
		Value:     setting.Value,
		CreatedAt: setting.CreatedAt,
		UpdatedAt: setting.UpdatedAt,
	}, nil
}

// UpdateSetting updates an existing setting
func (s *SettingsService) UpdateSetting(ctx context.Context, tenantID int, key string, updates domain.SettingUpdates) (domain.Setting, error) {
	// Validate value type if provided
	if updates.ValueType != nil && !isValidValueType(*updates.ValueType) {
		return domain.Setting{}, errors.WithDetails(errors.ErrInvalidInput, "Invalid value type: %s. Must be one of: boolean, number, string, json", *updates.ValueType)
	}

	// Convert domain to ports
	portsUpdates := ports.SettingUpdates{
		Value: updates.Value,
	}

	if updates.ValueType != nil {
		valueType := ports.SettingValueType(*updates.ValueType)
		portsUpdates.ValueType = &valueType
	}

	setting, err := s.db.UpdateSetting(ctx, tenantID, key, portsUpdates)
	if err != nil {
		return domain.Setting{}, fmt.Errorf("failed to update setting: %w", err)
	}

	return domain.Setting{
		ID:        setting.ID,
		TenantID:  setting.TenantID,
		Key:       setting.Key,
		ValueType: string(setting.ValueType),
		Value:     setting.Value,
		CreatedAt: setting.CreatedAt,
		UpdatedAt: setting.UpdatedAt,
	}, nil
}

// DeleteSetting deletes a setting
func (s *SettingsService) DeleteSetting(ctx context.Context, tenantID int, key string) error {
	err := s.db.DeleteSetting(ctx, tenantID, key)
	if err != nil {
		return fmt.Errorf("failed to delete setting: %w", err)
	}
	return nil
}

// SettingExists checks if a setting exists
func (s *SettingsService) SettingExists(ctx context.Context, tenantID int, key string) (bool, error) {
	exists, err := s.db.SettingExists(ctx, tenantID, key)
	if err != nil {
		return false, fmt.Errorf("failed to check setting existence: %w", err)
	}
	return exists, nil
}

// GetSettingsCount returns the total count of settings
func (s *SettingsService) GetSettingsCount(ctx context.Context) (int, error) {
	count, err := s.db.GetAllSettingsCount(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get settings count: %w", err)
	}
	return count, nil
}

// Helper function to validate value types
func isValidValueType(valueType string) bool {
	switch valueType {
	case "boolean", "number", "string", "json":
		return true
	default:
		return false
	}
}
