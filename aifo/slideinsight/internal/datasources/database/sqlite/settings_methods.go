// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package sqlite

import (
	"context"

	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/utils"
)

// Settings operations
// TODO: There are just simple wrappers around the settings service. The service can be embedded.
// GetSettings retrieves settings with search/filter and pagination
func (db *DB) GetSettings(ctx context.Context, tenantID int, search utils.SearchParams, pagination utils.PaginationParams) ([]ports.Setting, error) {
	return db.settings.GetSettings(ctx, tenantID, search, pagination)
}

// GetAllSettings retrieves settings from all tenants with search/filter and pagination
func (db *DB) GetAllSettings(ctx context.Context, search utils.SearchParams, pagination utils.PaginationParams) ([]ports.Setting, error) {
	return db.settings.GetAllSettings(ctx, search, pagination)
}

// GetSettingsCount returns the total count of settings for a tenant
func (db *DB) GetSettingsCount(ctx context.Context, tenantID int) (int, error) {
	return db.settings.GetSettingsCount(ctx, tenantID)
}

// GetAllSettingsCount returns the total count of settings across all tenants
func (db *DB) GetAllSettingsCount(ctx context.Context) (int, error) {
	return db.settings.GetAllSettingsCount(ctx)
}

// GetSettingsCountWithSearch returns the count of settings matching search criteria
func (db *DB) GetSettingsCountWithSearch(ctx context.Context, tenantID int, search utils.SearchParams) (int, error) {
	return db.settings.GetSettingsCountWithSearch(ctx, tenantID, search)
}

// GetAllSettingsCountWithSearch returns the count of settings matching search criteria across all tenants
func (db *DB) GetAllSettingsCountWithSearch(ctx context.Context, search utils.SearchParams) (int, error) {
	return db.settings.GetAllSettingsCountWithSearch(ctx, search)
}

// GetSetting retrieves a specific setting by key and tenant
func (db *DB) GetSetting(ctx context.Context, tenantID int, key string) (*ports.Setting, error) {
	return db.settings.GetSetting(ctx, tenantID, key)
}

// CreateSetting creates a new setting
func (db *DB) CreateSetting(ctx context.Context, newSetting ports.NewSetting) (*ports.Setting, error) {
	return db.settings.CreateSetting(ctx, newSetting)
}

// UpdateSetting updates an existing setting
func (db *DB) UpdateSetting(ctx context.Context, tenantID int, key string, updates ports.SettingUpdates) (*ports.Setting, error) {
	return db.settings.UpdateSetting(ctx, tenantID, key, updates)
}

// DeleteSetting deletes a setting
func (db *DB) DeleteSetting(ctx context.Context, tenantID int, key string) error {
	return db.settings.DeleteSetting(ctx, tenantID, key)
}

// SettingExists checks if a setting exists
func (db *DB) SettingExists(ctx context.Context, tenantID int, key string) (bool, error) {
	return db.settings.SettingExists(ctx, tenantID, key)
}
