// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
package ui

import (
	"embed"
	"net/http"

	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

// Embed the built UI assets from the dist directory
//
//go:embed dist
var DistFS embed.FS

var DistDir = filesystem.New(filesystem.Config{
	Root:         http.FS(DistFS),
	PathPrefix:   "dist",
	NotFoundFile: "dist/index.html",
	Browse:       true, // TODO: False in production
})
