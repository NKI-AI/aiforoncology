// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

package domain

type User struct {
	ID                int    `json:"-"` // Internal database ID, not exposed in API
	TenantID          int    `json:"-"` // Do not expose tenant ID in the API
	TenantUID         string `json:"tenantUid,omitempty"`
	ShortUID          string `json:"userUid"`
	Email             string `json:"email"`
	FirstName         string `json:"firstName"`
	LastName          string `json:"lastName"`
	Password          string `json:"-"` // Do not expose password in the API
	MustResetPassword bool   `json:"mustResetPassword"`
	IsActive          bool   `json:"isActive"`
	EmailVerified     bool   `json:"emailVerified"`
	CreatedAt         string `json:"createdAt"`
	UpdatedAt         string `json:"updatedAt"`
}

type UserUpdates struct {
	Email             *string `json:"email,omitempty"`
	FirstName         *string `json:"firstName,omitempty"`
	LastName          *string `json:"lastName,omitempty"`
	MustResetPassword *bool   `json:"mustResetPassword,omitempty"`
	IsActive          *bool   `json:"isActive,omitempty"`
	EmailVerified     *bool   `json:"emailVerified,omitempty"`
}

type UsersResponse struct {
	Users      []User         `json:"users"`
	Pagination PaginationInfo `json:"pagination"`
}
