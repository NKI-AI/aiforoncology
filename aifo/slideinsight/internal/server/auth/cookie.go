// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package auth

import "strings"

// SameSiteFromConfig converts a configuration SameSite string into the format
// expected by Fiber's cookie implementation. Accepted values are "strict",
// "lax" and "none" (case insensitive). Any other value defaults to "Lax".
func SameSiteFromConfig(sameSite string) string {
	switch strings.ToLower(sameSite) {
	case "strict":
		return "Strict"
	case "none":
		return "None"
	default:
		return "Lax"
	}
}
