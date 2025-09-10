// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package utils

import (
	"encoding/json"
)

// ToJSON converts an interface{} to JSON bytes
func ToJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// ParseJSON parses JSON bytes into the provided interface{}
func ParseJSON(data string, v interface{}) error {
	return json.Unmarshal([]byte(data), v)
}
