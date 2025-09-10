// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

package domain

import "time"

// Setting represents a setting in the system
type Setting struct {
	ID        int       `json:"id"`
	TenantID  int       `json:"tenantId"`
	Key       string    `json:"key"`
	ValueType string    `json:"valueType"` // boolean, number, string, json
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// NewSetting represents a new setting to be created
type NewSetting struct {
	TenantID  int    `json:"tenantId"`
	Key       string `json:"key"`
	ValueType string `json:"valueType"`
	Value     string `json:"value"`
}

// SettingUpdates represents fields that can be updated for an existing setting
type SettingUpdates struct {
	ValueType *string `json:"valueType,omitempty"`
	Value     *string `json:"value,omitempty"`
}

// SettingsResponse represents the response for settings list
type SettingsResponse struct {
	Settings   []Setting      `json:"settings"`
	Pagination PaginationInfo `json:"pagination"`
}
