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

	"aifo.dev/aifo/slideinsight/internal/utils"
)

// SettingValueType represents the type of a setting value
type SettingValueType string

const (
	SettingValueTypeBoolean SettingValueType = "boolean"
	SettingValueTypeNumber  SettingValueType = "number"
	SettingValueTypeString  SettingValueType = "string"
	SettingValueTypeJSON    SettingValueType = "json"
)

// Setting represents a setting in the database
type Setting struct {
	ID        int              `json:"id"`
	TenantID  int              `json:"tenantId"` // 0 = global
	Key       string           `json:"key"`
	ValueType SettingValueType `json:"valueType"`
	Value     string           `json:"value"` // JSON stored as string
	CreatedAt time.Time        `json:"createdAt"`
	UpdatedAt time.Time        `json:"updatedAt"`
}

// NewSetting represents a new setting to be created
type NewSetting struct {
	TenantID  int              `json:"tenantId"` // 0 = global
	Key       string           `json:"key"`
	ValueType SettingValueType `json:"valueType"`
	Value     string           `json:"value"` // JSON stored as string
}

// SettingUpdates represents fields that can be updated for an existing setting
type SettingUpdates struct {
	ValueType *SettingValueType `json:"valueType,omitempty"`
	Value     *string           `json:"value,omitempty"`
}

// SettingsRepository defines the interface for setting-related database operations
type SettingsRepository interface {
	// GetSettings retrieves settings with search/filter and pagination support
	GetSettings(ctx context.Context, tenantID int, search utils.SearchParams, pagination utils.PaginationParams) ([]Setting, error)

	// GetAllSettings retrieves settings from all tenants with search/filter and pagination support
	GetAllSettings(ctx context.Context, search utils.SearchParams, pagination utils.PaginationParams) ([]Setting, error)

	// GetSettingsCount returns the total count of settings for a tenant
	GetSettingsCount(ctx context.Context, tenantID int) (int, error)

	// GetAllSettingsCount returns the total count of settings across all tenants
	GetAllSettingsCount(ctx context.Context) (int, error)

	// GetSettingsCountWithSearch returns the count of settings matching search criteria
	GetSettingsCountWithSearch(ctx context.Context, tenantID int, search utils.SearchParams) (int, error)

	// GetAllSettingsCountWithSearch returns the count of settings matching search criteria across all tenants
	GetAllSettingsCountWithSearch(ctx context.Context, search utils.SearchParams) (int, error)

	// GetSetting retrieves a specific setting by key and tenant
	GetSetting(ctx context.Context, tenantID int, key string) (*Setting, error)

	// CreateSetting creates a new setting
	CreateSetting(ctx context.Context, newSetting NewSetting) (*Setting, error)

	// UpdateSetting updates an existing setting
	UpdateSetting(ctx context.Context, tenantID int, key string, updates SettingUpdates) (*Setting, error)

	// DeleteSetting deletes a setting
	DeleteSetting(ctx context.Context, tenantID int, key string) error

	// SettingExists checks if a setting exists
	SettingExists(ctx context.Context, tenantID int, key string) (bool, error)
}
