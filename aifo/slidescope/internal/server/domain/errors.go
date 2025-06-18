// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.

package domain

// ErrorResponse is a struct that represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}
