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

	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/utils"
)

// Adapter provides a unified interface for all settings operations
type Adapter struct {
	crud   *CrudService
	search *SearchService
}

// NewAdapter creates a new settings adapter
func NewAdapter(db *sql.DB) *Adapter {
	return &Adapter{
		crud:   NewCrudService(db),
		search: NewSearchService(db),
	}
}

// Settings operations

// GetSettings retrieves settings with search/filter and pagination
func (a *Adapter) GetSettings(ctx context.Context, tenantID int, search utils.SearchParams, pagination utils.PaginationParams) ([]ports.Setting, error) {
	return a.search.GetSettings(ctx, tenantID, search, pagination)
}

// GetAllSettings retrieves settings from all tenants with search/filter and pagination
func (a *Adapter) GetAllSettings(ctx context.Context, search utils.SearchParams, pagination utils.PaginationParams) ([]ports.Setting, error) {
	return a.search.GetAllSettings(ctx, search, pagination)
}

// GetSettingsCount returns the total count of settings for a tenant
func (a *Adapter) GetSettingsCount(ctx context.Context, tenantID int) (int, error) {
	return a.search.GetSettingsCount(ctx, tenantID)
}

// GetAllSettingsCount returns the total count of settings across all tenants
func (a *Adapter) GetAllSettingsCount(ctx context.Context) (int, error) {
	return a.search.GetAllSettingsCount(ctx)
}

// GetSettingsCountWithSearch returns the count of settings matching search criteria
func (a *Adapter) GetSettingsCountWithSearch(ctx context.Context, tenantID int, search utils.SearchParams) (int, error) {
	return a.search.GetSettingsCountWithSearch(ctx, tenantID, search)
}

// GetAllSettingsCountWithSearch returns the count of settings matching search criteria across all tenants
func (a *Adapter) GetAllSettingsCountWithSearch(ctx context.Context, search utils.SearchParams) (int, error) {
	return a.search.GetAllSettingsCountWithSearch(ctx, search)
}

// GetSetting retrieves a specific setting by key and tenant
func (a *Adapter) GetSetting(ctx context.Context, tenantID int, key string) (*ports.Setting, error) {
	return a.crud.GetSetting(ctx, tenantID, key)
}

// CreateSetting creates a new setting
func (a *Adapter) CreateSetting(ctx context.Context, newSetting ports.NewSetting) (*ports.Setting, error) {
	return a.crud.CreateSetting(ctx, newSetting)
}

// UpdateSetting updates an existing setting
func (a *Adapter) UpdateSetting(ctx context.Context, tenantID int, key string, updates ports.SettingUpdates) (*ports.Setting, error) {
	return a.crud.UpdateSetting(ctx, tenantID, key, updates)
}

// DeleteSetting deletes a setting
func (a *Adapter) DeleteSetting(ctx context.Context, tenantID int, key string) error {
	return a.crud.DeleteSetting(ctx, tenantID, key)
}

// SettingExists checks if a setting exists
func (a *Adapter) SettingExists(ctx context.Context, tenantID int, key string) (bool, error) {
	return a.crud.SettingExists(ctx, tenantID, key)
}
