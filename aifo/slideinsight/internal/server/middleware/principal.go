// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

package middleware

import (
	"context"
	"fmt"
)

// Principal holds the authenticated identity from the JWT
type Principal struct {
	Email     string
	UserUID   string
	TenantUID string
	TenantID  int
	// Fields for user switching functionality
	OriginalUserUID string // Set when an admin has switched to another user
	OriginalEmail   string // Original admin's email for audit trail
}

// ctxKey is an unexported type for context keys
type ctxKey string

const principalKey ctxKey = "principal"

// WithPrincipal returns a new Context that carries p
func WithPrincipal(ctx context.Context, p *Principal) context.Context {
	return context.WithValue(ctx, principalKey, p)
}

// FromContext extracts the Principal (or nil)
func FromContext(ctx context.Context) *Principal {
	if p, _ := ctx.Value(principalKey).(*Principal); p != nil {
		return p
	}
	return nil
}

// GetTenantIDFromContext extracts the tenant ID from the context
func GetTenantIDFromContext(ctx context.Context) (int, error) {
	principal := FromContext(ctx)
	if principal == nil {
		return 0, fmt.Errorf("no principal found in context")
	}
	if principal.TenantID == 0 {
		return 0, fmt.Errorf("no tenant ID found in principal")
	}
	return principal.TenantID, nil
}
