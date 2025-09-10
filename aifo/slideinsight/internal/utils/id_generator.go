// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package utils

import (
	"crypto/rand"
	"fmt"
)

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// base62EncodeFixed encodes a uint64 as a base62 string with a fixed length.
func base62EncodeFixed(n uint64, length int) string {
	if n == 0 {
		return fmt.Sprintf("%0*s", length, "0")
	}
	var encoded []byte
	for n > 0 {
		remainder := n % 62
		encoded = append([]byte{base62Chars[remainder]}, encoded...)
		n /= 62
	}
	// Prepend zeros if shorter than required length
	for len(encoded) < length {
		encoded = append([]byte{'0'}, encoded...)
	}
	return string(encoded)
}

// GenerateFixedShortUID generates a unique, random, fixed-length ID using base62 encoding.
// The resulting ID is 9 characters long and has good uniqueness properties.
func GenerateFixedShortUID() (string, error) {
	b := make([]byte, 6) // 48 bits
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	var n uint64
	for i := 0; i < 6; i++ {
		n = (n << 8) | uint64(b[i])
	}
	return base62EncodeFixed(n, 9), nil
}
